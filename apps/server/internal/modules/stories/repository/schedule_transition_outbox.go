package storiesrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduleTransitionStoryState struct {
	Status string
	Reason *string
	Locked bool
}

type latestScheduleTransition struct {
	SemanticFingerprint string
	TransitionSequence  int64
}

// UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged serializes scheduler
// decisions on the story row. The state update and durable event are committed
// together so workers can never observe an event for a rolled-back decision.
func (r *repo) UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	stateUpdatedAt time.Time,
	locked *bool,
	outbox storydomain.ScheduleTransitionOutboxInput,
) (bool, *storydomain.ScheduleTransitionOutboxEvent, error) {
	if err := validateScheduleTransitionOutboxInput(storyID, workspaceID, outbox); err != nil {
		return false, nil, err
	}
	if err := r.mutationConfigured(); err != nil {
		return false, nil, err
	}

	var applied bool
	var claimed *storydomain.ScheduleTransitionOutboxEvent
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		row, err := queries.LockStoryForScheduleTransition(ctx, storyreadsql.LockStoryForScheduleTransitionParams{
			StoryID: storyID, WorkspaceID: workspaceID, ExpectedUpdatedAt: expectedUpdatedAt.UTC(),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("lock story for schedule transition: %w", err)
		}
		current := scheduleTransitionStoryState{
			Status: row.AutoSchedulingStatus, Reason: row.AutoSchedulingReason, Locked: row.AutoSchedulingLocked,
		}

		latest := latestScheduleTransition{}
		latestRow, latestErr := queries.GetLatestStoryScheduleTransition(ctx, storyreadsql.GetLatestStoryScheduleTransitionParams{
			StoryID: storyID, WorkspaceID: workspaceID,
		})
		hasLatest := latestErr == nil
		if hasLatest {
			latest = latestScheduleTransition{
				SemanticFingerprint: latestRow.SemanticFingerprint,
				TransitionSequence:  latestRow.TransitionSequence,
			}
		} else if !errors.Is(latestErr, pgx.ErrNoRows) {
			return fmt.Errorf("read latest story schedule transition fingerprint: %w", latestErr)
		}

		targetLocked := current.Locked
		if locked != nil {
			targetLocked = *locked
		}
		if isImmediateScheduleTransitionRetry(
			current, latest, hasLatest, status, reason, targetLocked, outbox.SemanticFingerprint,
		) {
			applied = true
			return nil
		}

		updated, err := queries.UpdateStoryScheduleTransitionState(ctx, storyreadsql.UpdateStoryScheduleTransitionStateParams{
			AutoSchedulingStatus: status, AutoSchedulingReason: reason,
			AutoSchedulingUpdatedAt: timePointer(stateUpdatedAt.UTC()),
			SetAutoSchedulingLocked: locked != nil, AutoSchedulingLocked: targetLocked,
			StoryID: storyID, WorkspaceID: workspaceID, ExpectedUpdatedAt: expectedUpdatedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("update story state for schedule transition: %w", err)
		}
		if updated != 1 {
			return nil
		}

		claimToken := uuid.Nil
		var claimTokenParam *uuid.UUID
		if outbox.ClaimImmediately {
			claimToken = uuid.New()
			claimTokenParam = &claimToken
		}
		if err := queries.InsertStoryScheduleTransition(ctx, storyreadsql.InsertStoryScheduleTransitionParams{
			EventID: outbox.EventID, ActorID: outbox.ActorID, StoryID: storyID, WorkspaceID: workspaceID,
			EventPayload: append([]byte(nil), outbox.EventPayload...), SemanticFingerprint: outbox.SemanticFingerprint,
			TransitionSequence: latest.TransitionSequence + 1, ClaimImmediately: outbox.ClaimImmediately,
			ClaimToken: claimTokenParam,
		}); err != nil {
			return fmt.Errorf("insert story schedule transition outbox: %w", err)
		}
		applied = true
		if outbox.ClaimImmediately {
			claimed = &storydomain.ScheduleTransitionOutboxEvent{
				EventID: outbox.EventID, StoryID: storyID, WorkspaceID: workspaceID, ActorID: outbox.ActorID,
				SemanticFingerprint: outbox.SemanticFingerprint, TransitionSequence: latest.TransitionSequence + 1,
				ClaimToken: claimToken, AttemptCount: 1,
				EventPayload: append(json.RawMessage(nil), outbox.EventPayload...),
			}
		}
		return nil
	})
	if err != nil {
		return false, nil, err
	}
	return applied, claimed, nil
}

func (r *repo) ClaimScheduleTransitionOutboxEvents(
	ctx context.Context,
	limit int,
	staleAfter time.Duration,
) ([]storydomain.ScheduleTransitionOutboxEvent, error) {
	if limit <= 0 {
		return []storydomain.ScheduleTransitionOutboxEvent{}, nil
	}
	if err := r.mutationConfigured(); err != nil {
		return nil, err
	}
	batchSize, err := int32BatchSize(limit)
	if err != nil {
		return nil, err
	}
	if staleAfter < 0 {
		staleAfter = 0
	}
	rows, err := r.reads.ClaimStoryScheduleTransitions(ctx, storyreadsql.ClaimStoryScheduleTransitionsParams{
		StaleAfterSeconds: staleAfter.Seconds(), BatchSize: batchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("claim story schedule transition outbox: %w", err)
	}
	result := make([]storydomain.ScheduleTransitionOutboxEvent, 0, len(rows))
	for _, row := range rows {
		if row.ClaimToken == nil || *row.ClaimToken == uuid.Nil {
			return nil, errors.New("claimed story schedule transition is missing its claim token")
		}
		result = append(result, storydomain.ScheduleTransitionOutboxEvent{
			EventID: row.ScheduleTransitionEventID, StoryID: row.StoryID, WorkspaceID: row.WorkspaceID,
			ActorID: row.ActorID, SemanticFingerprint: row.SemanticFingerprint,
			TransitionSequence: row.TransitionSequence, ClaimToken: *row.ClaimToken,
			AttemptCount: int(row.AttemptCount), EventPayload: append(json.RawMessage(nil), row.EventPayload...),
		})
	}
	return result, nil
}

func (r *repo) CompleteScheduleTransitionOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	rows, err := r.reads.CompleteStoryScheduleTransition(ctx, storyreadsql.CompleteStoryScheduleTransitionParams{
		EventID: eventID, ClaimToken: &claimToken,
	})
	return claimGuardedScheduleTransitionResult(rows, err)
}

func (r *repo) RetryScheduleTransitionOutboxEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	failure string,
	retryAt time.Time,
) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	retryAt = retryAt.UTC()
	rows, err := r.reads.RetryStoryScheduleTransition(ctx, storyreadsql.RetryStoryScheduleTransitionParams{
		RetryAt: &retryAt, LastError: normalizedScheduleTransitionOutboxFailure(failure),
		EventID: eventID, ClaimToken: &claimToken,
	})
	return claimGuardedScheduleTransitionResult(rows, err)
}

func (r *repo) FailScheduleTransitionOutboxEvent(
	ctx context.Context,
	eventID, claimToken uuid.UUID,
	failure string,
) error {
	if err := r.mutationConfigured(); err != nil {
		return err
	}
	rows, err := r.reads.FailStoryScheduleTransition(ctx, storyreadsql.FailStoryScheduleTransitionParams{
		LastError: normalizedScheduleTransitionOutboxFailure(failure), EventID: eventID, ClaimToken: &claimToken,
	})
	return claimGuardedScheduleTransitionResult(rows, err)
}

func (r *repo) DeleteCompletedScheduleTransitionOutboxEvents(
	ctx context.Context,
	completedBefore time.Time,
	limit int,
) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	if err := r.mutationConfigured(); err != nil {
		return 0, err
	}
	batchSize, err := int32BatchSize(limit)
	if err != nil {
		return 0, err
	}
	completedBefore = completedBefore.UTC()
	rows, err := r.reads.DeleteCompletedStoryScheduleTransitions(ctx, storyreadsql.DeleteCompletedStoryScheduleTransitionsParams{
		CompletedBefore: &completedBefore, BatchSize: batchSize,
	})
	if err != nil {
		return 0, fmt.Errorf("delete completed story schedule transition outbox: %w", err)
	}
	return int(rows), nil
}

func validateScheduleTransitionOutboxInput(
	storyID, workspaceID uuid.UUID,
	input storydomain.ScheduleTransitionOutboxInput,
) error {
	if input.EventID == uuid.Nil || input.ActorID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil ||
		input.StoryID != storyID || input.WorkspaceID != workspaceID ||
		strings.TrimSpace(input.SemanticFingerprint) == "" || len(input.EventPayload) == 0 || !json.Valid(input.EventPayload) {
		return errors.New("story schedule transition outbox input is incomplete")
	}
	return nil
}

func claimGuardedScheduleTransitionResult(rows int64, err error) error {
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
	current scheduleTransitionStoryState,
	latest latestScheduleTransition,
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

func normalizedScheduleTransitionOutboxFailure(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "story schedule transition publication failed"
	}
	return value
}

func int32BatchSize(value int) (int32, error) {
	if value <= 0 {
		return 0, errors.New("story schedule transition batch size is out of range")
	}
	converted, err := safecast.Int32(value)
	if err != nil {
		return 0, errors.New("story schedule transition batch size is out of range")
	}
	return converted, nil
}

func timePointer(value time.Time) *time.Time {
	return &value
}
