package messagingrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	storyMutationCancellationStatusCancelled        = "cancelled"
	storyMutationCancellationStatusAlreadyCancelled = "already_cancelled"
)

type storyMutationConfirmationRow struct {
	Status      messaging.StoryMutationConfirmationStatus
	Result      []byte
	LastError   *string
	HasProposal bool
	ExpiresAt   time.Time
}

func (repository *Repository) RegisterStoryMutationConfirmation(
	ctx context.Context,
	input messaging.StoryMutationConfirmationStateInput,
) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	if err := validateStoryMutationConfirmationStateInput(input); err != nil {
		return err
	}
	if err := repository.queries.InsertStoryMutationConfirmation(ctx, messagingsql.InsertStoryMutationConfirmationParams{
		ConfirmationID: input.ConfirmationID, WorkspaceID: input.WorkspaceID,
		UserID: input.UserID, TeamID: input.TeamID, Operation: string(input.Operation),
		TokenHash: input.TokenHash, Proposal: string(input.Proposal), ExpiresAt: input.ExpiresAt.UTC(),
	}); err != nil {
		return fmt.Errorf("register messaging story mutation confirmation: %w", err)
	}
	return nil
}

func (repository *Repository) LoadStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
) (messaging.StoryMutationConfirmationRecord, error) {
	if !repository.configured() {
		return messaging.StoryMutationConfirmationRecord{}, errors.New("messaging repository is not configured")
	}
	if err := validateStoryMutationConfirmationBinding(binding); err != nil {
		return messaging.StoryMutationConfirmationRecord{}, err
	}
	row, err := repository.queries.LoadStoryMutationConfirmation(ctx, storyMutationBindingParams(binding))
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.StoryMutationConfirmationRecord{}, messaging.ErrInvalidConfirmation
	}
	if err != nil {
		return messaging.StoryMutationConfirmationRecord{}, fmt.Errorf("load story mutation confirmation: %w", err)
	}
	record := messaging.StoryMutationConfirmationRecord{
		TeamID: row.TeamID, Operation: messaging.StoryMutationOperation(row.Operation),
		Status:    messaging.StoryMutationConfirmationStatus(row.Status),
		Proposal:  append(json.RawMessage(nil), row.Proposal...),
		LastError: strings.TrimSpace(valueOrEmptyString(row.LastError)),
	}
	if len(row.Result) != 0 && string(row.Result) != "null" {
		var result messaging.StoryMutationResult
		if err := json.Unmarshal(row.Result, &result); err != nil {
			return messaging.StoryMutationConfirmationRecord{}, fmt.Errorf("decode loaded story mutation result: %w", err)
		}
		record.Result = &result
	}
	return record, nil
}

func (repository *Repository) ApplyStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
	apply func(context.Context) (messaging.StoryMutationResult, error),
) (messaging.StoryMutationResult, bool, error) {
	if !repository.configured() {
		return messaging.StoryMutationResult{}, false, errors.New("messaging repository is not configured")
	}
	if apply == nil {
		return messaging.StoryMutationResult{}, false, errors.New("story mutation apply callback is required")
	}
	if err := validateStoryMutationConfirmationBinding(binding); err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	duplicate, err := repository.consumeStoryMutationConfirmation(
		ctx, binding, now.UTC(), messaging.StoryMutationConfirmationApplied,
	)
	if err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	result, persisted, err := repository.applyStoryMutationResult(ctx, binding, apply)
	if err != nil {
		return result, false, err
	}
	return result, duplicate || persisted, nil
}

func (repository *Repository) CancelStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
) (messaging.StoryMutationCancellationResult, error) {
	if !repository.configured() {
		return messaging.StoryMutationCancellationResult{}, errors.New("messaging repository is not configured")
	}
	if err := validateStoryMutationConfirmationBinding(binding); err != nil {
		return messaging.StoryMutationCancellationResult{}, err
	}
	duplicate, err := repository.consumeStoryMutationConfirmation(
		ctx, binding, now.UTC(), messaging.StoryMutationConfirmationCancelled,
	)
	if err != nil {
		return messaging.StoryMutationCancellationResult{}, err
	}
	status := storyMutationCancellationStatusCancelled
	if duplicate {
		status = storyMutationCancellationStatusAlreadyCancelled
	}
	return messaging.StoryMutationCancellationResult{Status: status}, nil
}

func (repository *Repository) consumeStoryMutationConfirmation(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	now time.Time,
	desired messaging.StoryMutationConfirmationStatus,
) (bool, error) {
	var (
		duplicate bool
		resultErr error
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		row, err := lockStoryMutationConfirmation(ctx, queries, binding)
		if err != nil {
			return err
		}
		if (row.Status == messaging.StoryMutationConfirmationPending ||
			(row.Status == messaging.StoryMutationConfirmationApplied && row.HasProposal)) &&
			!now.Before(row.ExpiresAt.UTC()) {
			if err := transitionStoryMutationConfirmation(ctx, queries, binding.ConfirmationID, row.Status, messaging.StoryMutationConfirmationExpired, now); err != nil {
				return err
			}
			resultErr = messaging.ErrExpiredConfirmation
			return nil
		}
		switch row.Status {
		case messaging.StoryMutationConfirmationPending:
			return transitionStoryMutationConfirmation(ctx, queries, binding.ConfirmationID, row.Status, desired, now)
		case desired:
			duplicate = true
			return nil
		case messaging.StoryMutationConfirmationApplied:
			return messaging.ErrAppliedConfirmation
		case messaging.StoryMutationConfirmationCancelled:
			return messaging.ErrCancelledConfirmation
		case messaging.StoryMutationConfirmationExpired:
			return messaging.ErrExpiredConfirmation
		default:
			return fmt.Errorf("%w: unsupported confirmation status %q", messaging.ErrInvalidConfirmation, row.Status)
		}
	})
	if err != nil {
		return false, fmt.Errorf("transition story mutation confirmation: %w", err)
	}
	return duplicate, resultErr
}

func (repository *Repository) applyStoryMutationResult(
	ctx context.Context,
	binding messaging.StoryMutationConfirmationBinding,
	apply func(context.Context) (messaging.StoryMutationResult, error),
) (messaging.StoryMutationResult, bool, error) {
	var (
		result    messaging.StoryMutationResult
		persisted bool
		applyErr  error
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		row, err := lockStoryMutationConfirmation(ctx, queries, binding)
		if err != nil {
			return err
		}
		if row.Status != messaging.StoryMutationConfirmationApplied {
			return storyMutationTerminalError(row.Status)
		}
		var persistedResult *messaging.StoryMutationResult
		if len(row.Result) != 0 && string(row.Result) != "null" {
			if err := json.Unmarshal(row.Result, &result); err != nil {
				return fmt.Errorf("decode persisted story mutation result: %w", err)
			}
			stored := result
			persistedResult = &stored
		}
		if persistedResult != nil && row.LastError == nil {
			persisted = true
			return nil
		}

		result, applyErr = apply(ctx)
		if applyErr != nil {
			result = preferredStoryMutationProgress(persistedResult, result)
			resultPayload := ""
			if result.Operation == messaging.StoryMutationCreateBatch {
				encoded, encodeErr := json.Marshal(result)
				if encodeErr != nil {
					return errors.Join(applyErr, fmt.Errorf("encode partial story mutation result: %w", encodeErr))
				}
				resultPayload = string(encoded)
			}
			lastError := truncateStoryMutationError(applyErr.Error())
			if err := queries.RecordStoryMutationApplyFailure(ctx, messagingsql.RecordStoryMutationApplyFailureParams{
				Result: resultPayload, LastError: &lastError, ConfirmationID: binding.ConfirmationID,
			}); err != nil {
				return errors.Join(applyErr, fmt.Errorf("persist story mutation apply failure: %w", err))
			}
			return nil
		}

		payload, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("encode story mutation result: %w", err)
		}
		affected, err := queries.CompleteStoryMutationApply(ctx, messagingsql.CompleteStoryMutationApplyParams{
			Result: payload, ConfirmationID: binding.ConfirmationID,
		})
		if err != nil {
			return fmt.Errorf("persist story mutation result: %w", err)
		}
		return requireAffectedRows(affected, "persist story mutation result")
	})
	if err != nil {
		return messaging.StoryMutationResult{}, false, err
	}
	if applyErr != nil {
		return result, false, applyErr
	}
	return result, persisted, nil
}

func lockStoryMutationConfirmation(
	ctx context.Context,
	queries messagingsql.Querier,
	binding messaging.StoryMutationConfirmationBinding,
) (storyMutationConfirmationRow, error) {
	row, err := queries.LockStoryMutationConfirmation(ctx, messagingsql.LockStoryMutationConfirmationParams(storyMutationBindingParams(binding)))
	if errors.Is(err, pgx.ErrNoRows) {
		return storyMutationConfirmationRow{}, messaging.ErrInvalidConfirmation
	}
	if err != nil {
		return storyMutationConfirmationRow{}, fmt.Errorf("lock story mutation confirmation: %w", err)
	}
	return storyMutationConfirmationRow{
		Status: messaging.StoryMutationConfirmationStatus(row.Status), Result: row.Result,
		LastError: row.LastError, HasProposal: row.HasProposal, ExpiresAt: row.ExpiresAt,
	}, nil
}

func transitionStoryMutationConfirmation(
	ctx context.Context,
	queries messagingsql.Querier,
	confirmationID uuid.UUID,
	currentStatus, status messaging.StoryMutationConfirmationStatus,
	now time.Time,
) error {
	switch status {
	case messaging.StoryMutationConfirmationApplied,
		messaging.StoryMutationConfirmationCancelled,
		messaging.StoryMutationConfirmationExpired:
	default:
		return fmt.Errorf("%w: unsupported terminal status %q", messaging.ErrInvalidConfirmation, status)
	}
	affected, err := queries.TransitionStoryMutationConfirmation(ctx, messagingsql.TransitionStoryMutationConfirmationParams{
		Status: string(status), Now: now, ConfirmationID: confirmationID, CurrentStatus: string(currentStatus),
	})
	if err != nil {
		return fmt.Errorf("transition story mutation confirmation: %w", err)
	}
	return requireAffectedRows(affected, "transition story mutation confirmation")
}

func storyMutationBindingParams(binding messaging.StoryMutationConfirmationBinding) messagingsql.LoadStoryMutationConfirmationParams {
	return messagingsql.LoadStoryMutationConfirmationParams{
		ConfirmationID: binding.ConfirmationID, WorkspaceID: binding.WorkspaceID,
		UserID: binding.UserID, TokenHash: binding.TokenHash,
	}
}

func preferredStoryMutationProgress(
	persisted *messaging.StoryMutationResult,
	current messaging.StoryMutationResult,
) messaging.StoryMutationResult {
	if persisted == nil || len(current.Items) >= len(persisted.Items) {
		return current
	}
	return *persisted
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
	switch input.Operation {
	case messaging.StoryMutationCreate,
		messaging.StoryMutationCreateBatch,
		messaging.StoryMutationUpdate,
		messaging.StoryMutationComment,
		messaging.StoryMutationRelation:
	default:
		return fmt.Errorf("%w: unsupported operation %q", messaging.ErrInvalidConfirmation, input.Operation)
	}
	if input.Operation == messaging.StoryMutationCreateBatch {
		if len(input.Proposal) == 0 || len(input.Proposal) > maximumStoryMutationProposalBytes || !json.Valid(input.Proposal) {
			return fmt.Errorf("%w: batch proposal must contain valid bounded JSON", messaging.ErrInvalidConfirmation)
		}
	} else if len(input.Proposal) != 0 {
		return fmt.Errorf("%w: server-side proposal is only supported for batch creation", messaging.ErrInvalidConfirmation)
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

const maximumStoryMutationProposalBytes = 64 << 10

func truncateStoryMutationError(message string) string {
	message = strings.TrimSpace(message)
	const maximumRunes = 2_000
	runes := []rune(message)
	if len(runes) <= maximumRunes {
		return message
	}
	return string(runes[:maximumRunes])
}
