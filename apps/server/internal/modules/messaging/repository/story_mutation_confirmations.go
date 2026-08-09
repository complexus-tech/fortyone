package messagingrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	storyMutationCancellationStatusCancelled        = "cancelled"
	storyMutationCancellationStatusAlreadyCancelled = "already_cancelled"
)

type storyMutationConfirmationRow struct {
	Status    messaging.StoryMutationConfirmationStatus `db:"status"`
	Result    []byte                                    `db:"result"`
	ExpiresAt time.Time                                 `db:"expires_at"`
}

// RegisterStoryMutationConfirmation persists a proposal before its signed
// token is shown by any provider. The token itself is never stored.
func (r *Repository) RegisterStoryMutationConfirmation(
	ctx context.Context,
	input messaging.StoryMutationConfirmationStateInput,
) error {
	if r == nil || r.db == nil {
		return errors.New("messaging repository is not configured")
	}
	if err := validateStoryMutationConfirmationStateInput(input); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO messaging_story_mutation_confirmations (
			confirmation_id, workspace_id, user_id, team_id, operation,
			token_hash, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
	`, input.ConfirmationID, input.WorkspaceID, input.UserID, input.TeamID,
		input.Operation, input.TokenHash, input.ExpiresAt.UTC())
	if err != nil {
		return fmt.Errorf("register messaging story mutation confirmation: %w", err)
	}
	return nil
}

// ApplyStoryMutationConfirmation atomically consumes pending consent before a
// write, then serializes and persists the result. The initial state transition
// commits before apply runs, so cancellation can never win after confirmation
// has been accepted. A failed apply remains retryable only by Confirm.
func (r *Repository) ApplyStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
	apply func(context.Context) (messaging.StoryMutationResult, error),
) (messaging.StoryMutationResult, bool, error) {
	if r == nil || r.db == nil {
		return messaging.StoryMutationResult{}, false, errors.New("messaging repository is not configured")
	}
	if apply == nil {
		return messaging.StoryMutationResult{}, false, errors.New("story mutation apply callback is required")
	}
	if err := validateStoryMutationConfirmationBinding(binding); err != nil {
		return messaging.StoryMutationResult{}, false, err
	}

	duplicate, err := r.consumeStoryMutationConfirmation(ctx, binding, now.UTC(), messaging.StoryMutationConfirmationApplied)
	if err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	result, persisted, err := r.applyStoryMutationResult(ctx, binding, apply)
	if err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	return result, duplicate || persisted, nil
}

// CancelStoryMutationConfirmation atomically consumes pending consent without
// invoking a write. Repeating Cancel is idempotent; an applied confirmation is
// never relabelled as cancelled.
func (r *Repository) CancelStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
) (messaging.StoryMutationCancellationResult, error) {
	if r == nil || r.db == nil {
		return messaging.StoryMutationCancellationResult{}, errors.New("messaging repository is not configured")
	}
	if err := validateStoryMutationConfirmationBinding(binding); err != nil {
		return messaging.StoryMutationCancellationResult{}, err
	}

	duplicate, err := r.consumeStoryMutationConfirmation(ctx, binding, now.UTC(), messaging.StoryMutationConfirmationCancelled)
	if err != nil {
		return messaging.StoryMutationCancellationResult{}, err
	}
	status := storyMutationCancellationStatusCancelled
	if duplicate {
		status = storyMutationCancellationStatusAlreadyCancelled
	}
	return messaging.StoryMutationCancellationResult{Status: status}, nil
}

func (r *Repository) consumeStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
	desired messaging.StoryMutationConfirmationStatus,
) (duplicate bool, err error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin story mutation confirmation transition: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row, err := selectStoryMutationConfirmationForUpdate(ctx, tx, binding)
	if err != nil {
		return false, err
	}
	if row.Status == messaging.StoryMutationConfirmationPending && !now.Before(row.ExpiresAt.UTC()) {
		if err := setStoryMutationConfirmationStatus(ctx, tx, binding.ConfirmationID, messaging.StoryMutationConfirmationExpired, now); err != nil {
			return false, err
		}
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit expired story mutation confirmation: %w", err)
		}
		return false, messaging.ErrExpiredConfirmation
	}

	switch row.Status {
	case messaging.StoryMutationConfirmationPending:
		if err := setStoryMutationConfirmationStatus(ctx, tx, binding.ConfirmationID, desired, now); err != nil {
			return false, err
		}
	case desired:
		duplicate = true
	case messaging.StoryMutationConfirmationApplied:
		return false, messaging.ErrAppliedConfirmation
	case messaging.StoryMutationConfirmationCancelled:
		return false, messaging.ErrCancelledConfirmation
	case messaging.StoryMutationConfirmationExpired:
		return false, messaging.ErrExpiredConfirmation
	default:
		return false, fmt.Errorf("%w: unsupported confirmation status %q", messaging.ErrInvalidConfirmation, row.Status)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit story mutation confirmation transition: %w", err)
	}
	return duplicate, nil
}

func (r *Repository) applyStoryMutationResult(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	apply func(context.Context) (messaging.StoryMutationResult, error),
) (result messaging.StoryMutationResult, persisted bool, err error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.StoryMutationResult{}, false, fmt.Errorf("begin story mutation application: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row, err := selectStoryMutationConfirmationForUpdate(ctx, tx, binding)
	if err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	if row.Status != messaging.StoryMutationConfirmationApplied {
		return messaging.StoryMutationResult{}, false, storyMutationTerminalError(row.Status)
	}
	if len(row.Result) != 0 && string(row.Result) != "null" {
		if err := json.Unmarshal(row.Result, &result); err != nil {
			return messaging.StoryMutationResult{}, false, fmt.Errorf("decode persisted story mutation result: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return messaging.StoryMutationResult{}, false, fmt.Errorf("commit story mutation result read: %w", err)
		}
		return result, true, nil
	}

	result, applyErr := apply(ctx)
	if applyErr != nil {
		_, persistErr := tx.ExecContext(ctx, `
			UPDATE messaging_story_mutation_confirmations
			SET last_error = $2, updated_at = NOW()
			WHERE confirmation_id = $1 AND status = 'applied'
		`, binding.ConfirmationID, truncateStoryMutationError(applyErr.Error()))
		if persistErr == nil {
			persistErr = tx.Commit()
		}
		if persistErr != nil {
			return messaging.StoryMutationResult{}, false, errors.Join(
				applyErr,
				fmt.Errorf("persist story mutation apply failure: %w", persistErr),
			)
		}
		return messaging.StoryMutationResult{}, false, applyErr
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return messaging.StoryMutationResult{}, false, fmt.Errorf("encode story mutation result: %w", err)
	}
	resultUpdate, err := tx.ExecContext(ctx, `
		UPDATE messaging_story_mutation_confirmations
		SET result = CAST($2 AS jsonb), last_error = NULL, updated_at = NOW()
		WHERE confirmation_id = $1 AND status = 'applied'
	`, binding.ConfirmationID, payload)
	if err != nil {
		return messaging.StoryMutationResult{}, false, fmt.Errorf("persist story mutation result: %w", err)
	}
	if err := requireAffectedRow(resultUpdate, "persist story mutation result"); err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return messaging.StoryMutationResult{}, false, fmt.Errorf("commit story mutation result: %w", err)
	}
	return result, false, nil
}

func selectStoryMutationConfirmationForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	binding messaging.StoryMutationConfirmationBinding,
) (storyMutationConfirmationRow, error) {
	var row storyMutationConfirmationRow
	err := tx.GetContext(ctx, &row, `
		SELECT status, COALESCE(result, CAST('null' AS jsonb)) AS result, expires_at
		FROM messaging_story_mutation_confirmations
		WHERE confirmation_id = $1
		  AND workspace_id = $2
		  AND user_id = $3
		  AND token_hash = $4
		FOR UPDATE
	`, binding.ConfirmationID, binding.WorkspaceID, binding.UserID, binding.TokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return storyMutationConfirmationRow{}, messaging.ErrInvalidConfirmation
	}
	if err != nil {
		return storyMutationConfirmationRow{}, fmt.Errorf("lock story mutation confirmation: %w", err)
	}
	return row, nil
}

func setStoryMutationConfirmationStatus(
	ctx context.Context,
	tx *sqlx.Tx,
	confirmationID uuid.UUID,
	status messaging.StoryMutationConfirmationStatus,
	now time.Time,
) error {
	var appliedAt, cancelledAt, expiredAt any
	switch status {
	case messaging.StoryMutationConfirmationApplied:
		appliedAt = now
	case messaging.StoryMutationConfirmationCancelled:
		cancelledAt = now
	case messaging.StoryMutationConfirmationExpired:
		expiredAt = now
	default:
		return fmt.Errorf("%w: unsupported terminal status %q", messaging.ErrInvalidConfirmation, status)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE messaging_story_mutation_confirmations
		SET status = $2,
		    applied_at = $3,
		    cancelled_at = $4,
		    expired_at = $5,
		    updated_at = NOW()
		WHERE confirmation_id = $1 AND status = 'pending'
	`, confirmationID, status, appliedAt, cancelledAt, expiredAt)
	if err != nil {
		return fmt.Errorf("transition story mutation confirmation: %w", err)
	}
	return requireAffectedRow(result, "transition story mutation confirmation")
}

func storyMutationTerminalError(status messaging.StoryMutationConfirmationStatus) error {
	switch status {
	case messaging.StoryMutationConfirmationCancelled:
		return messaging.ErrCancelledConfirmation
	case messaging.StoryMutationConfirmationExpired:
		return messaging.ErrExpiredConfirmation
	case messaging.StoryMutationConfirmationApplied:
		return nil
	default:
		return fmt.Errorf("%w: unsupported confirmation status %q", messaging.ErrInvalidConfirmation, status)
	}
}

func validateStoryMutationConfirmationStateInput(input messaging.StoryMutationConfirmationStateInput) error {
	if input.ConfirmationID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.TeamID == uuid.Nil {
		return fmt.Errorf("%w: confirmation, workspace, user, and team are required", messaging.ErrInvalidConfirmation)
	}
	if input.Operation != messaging.StoryMutationCreate && input.Operation != messaging.StoryMutationUpdate {
		return fmt.Errorf("%w: unsupported operation %q", messaging.ErrInvalidConfirmation, input.Operation)
	}
	if len(input.TokenHash) != sha256DigestSize {
		return fmt.Errorf("%w: token hash must contain %d bytes", messaging.ErrInvalidConfirmation, sha256DigestSize)
	}
	if input.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: expiration is required", messaging.ErrInvalidConfirmation)
	}
	return nil
}

func validateStoryMutationConfirmationBinding(binding messaging.StoryMutationConfirmationBinding) error {
	if binding.ConfirmationID == uuid.Nil || binding.WorkspaceID == uuid.Nil || binding.UserID == uuid.Nil {
		return fmt.Errorf("%w: confirmation, workspace, and user are required", messaging.ErrInvalidConfirmation)
	}
	if len(binding.TokenHash) != sha256DigestSize {
		return fmt.Errorf("%w: token hash must contain %d bytes", messaging.ErrInvalidConfirmation, sha256DigestSize)
	}
	return nil
}

const sha256DigestSize = 32

func truncateStoryMutationError(message string) string {
	message = strings.TrimSpace(message)
	const maximumRunes = 2_000
	runes := []rune(message)
	if len(runes) <= maximumRunes {
		return message
	}
	return string(runes[:maximumRunes])
}
