package storiesrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var _ stories.ScheduleTransitionOutboxRepository = (*repo)(nil)

const (
	lockStoryForScheduleTransitionQuery = `
		SELECT auto_scheduling_status, auto_scheduling_reason, auto_scheduling_locked
		FROM stories
		WHERE id = $1
			AND workspace_id = $2
			AND updated_at = $3
		FOR UPDATE
	`
	latestStoryScheduleTransitionFingerprintQuery = `
		SELECT semantic_fingerprint, transition_sequence
		FROM story_schedule_transition_outbox
		WHERE story_id = $1 AND workspace_id = $2
		ORDER BY transition_sequence DESC, schedule_transition_event_id DESC
		LIMIT 1
	`
	updateStoryAutoSchedulingStateForTransitionQuery = `
		UPDATE stories
		SET auto_scheduling_status = $4,
			auto_scheduling_reason = $5,
			auto_scheduling_updated_at = $6,
			auto_scheduling_locked = COALESCE($7, auto_scheduling_locked)
		WHERE id = $1
			AND workspace_id = $2
			AND updated_at = $3
	`
	insertStoryScheduleTransitionOutboxQuery = `
		INSERT INTO story_schedule_transition_outbox (
			schedule_transition_event_id,
			actor_id,
			story_id,
			workspace_id,
			event_type,
			event_payload,
			semantic_fingerprint,
			transition_sequence,
			status,
			attempt_count,
			next_attempt_at,
			claim_token,
			claimed_at
		) VALUES (
			$1, $2, $3, $4, 'story.updated', CAST($5 AS jsonb), $6, $7,
			CASE WHEN $8 THEN 'processing' ELSE 'pending' END,
			CASE WHEN $8 THEN 1 ELSE 0 END,
			CASE WHEN $8 THEN NULL ELSE CURRENT_TIMESTAMP END,
			CASE WHEN $8 THEN $9 ELSE NULL END,
			CASE WHEN $8 THEN CURRENT_TIMESTAMP ELSE NULL END
		)
	`
	claimStoryScheduleTransitionOutboxQuery = `
		WITH candidates AS (
			SELECT schedule_transition_event_id
			FROM story_schedule_transition_outbox
			WHERE (status IN ('pending', 'retrying') AND next_attempt_at <= CURRENT_TIMESTAMP)
				OR (status = 'processing' AND claimed_at <= CURRENT_TIMESTAMP - CAST($2 AS interval))
			ORDER BY COALESCE(next_attempt_at, claimed_at), created_at, schedule_transition_event_id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		), claimed AS (
			UPDATE story_schedule_transition_outbox outbox
			SET status = 'processing',
				attempt_count = outbox.attempt_count + 1,
				next_attempt_at = NULL,
				claim_token = gen_random_uuid(),
				claimed_at = CURRENT_TIMESTAMP,
				completed_at = NULL,
				last_error = NULL,
				updated_at = CURRENT_TIMESTAMP
			FROM candidates
			WHERE outbox.schedule_transition_event_id = candidates.schedule_transition_event_id
			RETURNING outbox.*
		)
		SELECT schedule_transition_event_id, actor_id, story_id, workspace_id,
			semantic_fingerprint, transition_sequence, claim_token, attempt_count, event_payload
		FROM claimed
		ORDER BY created_at, schedule_transition_event_id
	`
	completeStoryScheduleTransitionOutboxQuery = `
		UPDATE story_schedule_transition_outbox
		SET status = 'completed',
			next_attempt_at = NULL,
			claim_token = NULL,
			claimed_at = NULL,
			completed_at = CURRENT_TIMESTAMP,
			last_error = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE schedule_transition_event_id = $1
			AND claim_token = $2
			AND status = 'processing'
	`
	retryStoryScheduleTransitionOutboxQuery = `
		UPDATE story_schedule_transition_outbox
		SET status = 'retrying',
			next_attempt_at = $3,
			claim_token = NULL,
			claimed_at = NULL,
			completed_at = NULL,
			last_error = LEFT($4, 4000),
			updated_at = CURRENT_TIMESTAMP
		WHERE schedule_transition_event_id = $1
			AND claim_token = $2
			AND status = 'processing'
	`
	failStoryScheduleTransitionOutboxQuery = `
		UPDATE story_schedule_transition_outbox
		SET status = 'failed',
			next_attempt_at = NULL,
			claim_token = NULL,
			claimed_at = NULL,
			completed_at = NULL,
			last_error = LEFT($3, 4000),
			updated_at = CURRENT_TIMESTAMP
		WHERE schedule_transition_event_id = $1
			AND claim_token = $2
			AND status = 'processing'
	`
	deleteCompletedStoryScheduleTransitionOutboxQuery = `
		WITH expired AS (
			SELECT schedule_transition_event_id
			FROM story_schedule_transition_outbox
			WHERE status = 'completed' AND completed_at < $1
			ORDER BY completed_at, schedule_transition_event_id
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		)
		DELETE FROM story_schedule_transition_outbox outbox
		USING expired
		WHERE outbox.schedule_transition_event_id = expired.schedule_transition_event_id
	`
)

type dbScheduleTransitionStoryState struct {
	Status string  `db:"auto_scheduling_status"`
	Reason *string `db:"auto_scheduling_reason"`
	Locked bool    `db:"auto_scheduling_locked"`
}

type dbScheduleTransitionOutboxClaim struct {
	EventID             uuid.UUID       `db:"schedule_transition_event_id"`
	ActorID             uuid.UUID       `db:"actor_id"`
	StoryID             uuid.UUID       `db:"story_id"`
	WorkspaceID         uuid.UUID       `db:"workspace_id"`
	SemanticFingerprint string          `db:"semantic_fingerprint"`
	TransitionSequence  int64           `db:"transition_sequence"`
	ClaimToken          uuid.UUID       `db:"claim_token"`
	AttemptCount        int             `db:"attempt_count"`
	EventPayload        json.RawMessage `db:"event_payload"`
}

type dbLatestScheduleTransition struct {
	SemanticFingerprint string `db:"semantic_fingerprint"`
	TransitionSequence  int64  `db:"transition_sequence"`
}

// UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged serializes scheduler
// decisions on the story row. The latest semantic fingerprint is used only to
// recognize an immediate retry; it is intentionally not globally unique so an
// identical transition may recur after an intervening decision.
func (r *repo) UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	stateUpdatedAt time.Time,
	locked *bool,
	outbox stories.CoreScheduleTransitionOutboxInput,
) (bool, *stories.CoreScheduleTransitionOutboxEvent, error) {
	if err := validateScheduleTransitionOutboxInput(storyID, workspaceID, outbox); err != nil {
		return false, nil, err
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, nil, fmt.Errorf("begin story schedule transition transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var current dbScheduleTransitionStoryState
	if err := tx.GetContext(
		ctx,
		&current,
		lockStoryForScheduleTransitionQuery,
		storyID,
		workspaceID,
		expectedUpdatedAt.UTC(),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("lock story for schedule transition: %w", err)
	}

	var latest dbLatestScheduleTransition
	latestErr := tx.GetContext(
		ctx,
		&latest,
		latestStoryScheduleTransitionFingerprintQuery,
		storyID,
		workspaceID,
	)
	if latestErr != nil && !errors.Is(latestErr, sql.ErrNoRows) {
		return false, nil, fmt.Errorf("read latest story schedule transition fingerprint: %w", latestErr)
	}
	targetLocked := current.Locked
	if locked != nil {
		targetLocked = *locked
	}
	if isImmediateScheduleTransitionRetry(
		current,
		latest,
		latestErr == nil,
		status,
		reason,
		targetLocked,
		outbox.SemanticFingerprint,
	) {
		if err := tx.Commit(); err != nil {
			return false, nil, fmt.Errorf("commit idempotent story schedule transition: %w", err)
		}
		return true, nil, nil
	}

	result, err := tx.ExecContext(
		ctx,
		updateStoryAutoSchedulingStateForTransitionQuery,
		storyID,
		workspaceID,
		expectedUpdatedAt.UTC(),
		status,
		reason,
		stateUpdatedAt.UTC(),
		locked,
	)
	if err != nil {
		return false, nil, fmt.Errorf("update story state for schedule transition: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, nil, fmt.Errorf("read story schedule transition update result: %w", err)
	}
	if rows != 1 {
		return false, nil, nil
	}

	claimToken := uuid.Nil
	if outbox.ClaimImmediately {
		claimToken = uuid.New()
	}
	if _, err := tx.ExecContext(
		ctx,
		insertStoryScheduleTransitionOutboxQuery,
		outbox.EventID,
		outbox.ActorID,
		storyID,
		workspaceID,
		[]byte(outbox.EventPayload),
		outbox.SemanticFingerprint,
		latest.TransitionSequence+1,
		outbox.ClaimImmediately,
		claimToken,
	); err != nil {
		return false, nil, fmt.Errorf("insert story schedule transition outbox: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, nil, fmt.Errorf("commit story schedule transition: %w", err)
	}
	if !outbox.ClaimImmediately {
		return true, nil, nil
	}
	return true, &stories.CoreScheduleTransitionOutboxEvent{
		EventID:             outbox.EventID,
		StoryID:             storyID,
		WorkspaceID:         workspaceID,
		ActorID:             outbox.ActorID,
		SemanticFingerprint: outbox.SemanticFingerprint,
		TransitionSequence:  latest.TransitionSequence + 1,
		ClaimToken:          claimToken,
		AttemptCount:        1,
		EventPayload:        append(json.RawMessage(nil), outbox.EventPayload...),
	}, nil
}

func (r *repo) ClaimScheduleTransitionOutboxEvents(
	ctx context.Context,
	limit int,
	staleAfter time.Duration,
) ([]stories.CoreScheduleTransitionOutboxEvent, error) {
	if limit <= 0 {
		return []stories.CoreScheduleTransitionOutboxEvent{}, nil
	}
	rows := make([]dbScheduleTransitionOutboxClaim, 0)
	if err := r.db.SelectContext(
		ctx,
		&rows,
		claimStoryScheduleTransitionOutboxQuery,
		limit,
		scheduleTransitionIntervalLiteral(staleAfter),
	); err != nil {
		return nil, fmt.Errorf("claim story schedule transition outbox: %w", err)
	}
	result := make([]stories.CoreScheduleTransitionOutboxEvent, len(rows))
	for index, row := range rows {
		result[index] = stories.CoreScheduleTransitionOutboxEvent{
			EventID:             row.EventID,
			StoryID:             row.StoryID,
			WorkspaceID:         row.WorkspaceID,
			ActorID:             row.ActorID,
			SemanticFingerprint: row.SemanticFingerprint,
			TransitionSequence:  row.TransitionSequence,
			ClaimToken:          row.ClaimToken,
			AttemptCount:        row.AttemptCount,
			EventPayload:        append(json.RawMessage(nil), row.EventPayload...),
		}
	}
	return result, nil
}

func (r *repo) CompleteScheduleTransitionOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	return execClaimGuardedScheduleTransitionOutbox(
		ctx,
		r.db,
		completeStoryScheduleTransitionOutboxQuery,
		eventID,
		claimToken,
	)
}

func (r *repo) RetryScheduleTransitionOutboxEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	failure string,
	retryAt time.Time,
) error {
	failure = normalizedScheduleTransitionOutboxFailure(failure)
	return execClaimGuardedScheduleTransitionOutbox(
		ctx,
		r.db,
		retryStoryScheduleTransitionOutboxQuery,
		eventID,
		claimToken,
		retryAt.UTC(),
		failure,
	)
}

func (r *repo) FailScheduleTransitionOutboxEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	failure string,
) error {
	return execClaimGuardedScheduleTransitionOutbox(
		ctx,
		r.db,
		failStoryScheduleTransitionOutboxQuery,
		eventID,
		claimToken,
		normalizedScheduleTransitionOutboxFailure(failure),
	)
}

func (r *repo) DeleteCompletedScheduleTransitionOutboxEvents(
	ctx context.Context,
	completedBefore time.Time,
	limit int,
) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(
		ctx,
		deleteCompletedStoryScheduleTransitionOutboxQuery,
		completedBefore.UTC(),
		limit,
	)
	if err != nil {
		return 0, fmt.Errorf("delete completed story schedule transition outbox: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted story schedule transition outbox count: %w", err)
	}
	return int(rows), nil
}

func validateScheduleTransitionOutboxInput(
	storyID, workspaceID uuid.UUID,
	input stories.CoreScheduleTransitionOutboxInput,
) error {
	if input.EventID == uuid.Nil || input.ActorID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil ||
		input.StoryID != storyID || input.WorkspaceID != workspaceID ||
		strings.TrimSpace(input.SemanticFingerprint) == "" || len(input.EventPayload) == 0 {
		return errors.New("story schedule transition outbox input is incomplete")
	}
	return nil
}

func execClaimGuardedScheduleTransitionOutbox(
	ctx context.Context,
	executor sqlx.ExtContext,
	query string,
	args ...any,
) error {
	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("story schedule transition outbox claim no longer belongs to this worker")
	}
	return nil
}

func equalScheduleTransitionReason(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func isImmediateScheduleTransitionRetry(
	current dbScheduleTransitionStoryState,
	latest dbLatestScheduleTransition,
	hasLatest bool,
	status string,
	reason *string,
	locked bool,
	fingerprint string,
) bool {
	return hasLatest &&
		latest.SemanticFingerprint == fingerprint &&
		current.Status == status &&
		equalScheduleTransitionReason(current.Reason, reason) &&
		current.Locked == locked
}

func scheduleTransitionIntervalLiteral(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	return fmt.Sprintf("%f seconds", value.Seconds())
}

func normalizedScheduleTransitionOutboxFailure(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "story schedule transition publication failed"
	}
	return value
}
