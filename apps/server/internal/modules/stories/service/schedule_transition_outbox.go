package stories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const (
	scheduleTransitionOutboxBatchSize      = 50
	scheduleTransitionOutboxStaleAfter     = 10 * time.Minute
	scheduleTransitionOutboxInitialDelay   = 30 * time.Second
	scheduleTransitionOutboxMaxDelay       = 15 * time.Minute
	scheduleTransitionOutboxRetention      = 30 * 24 * time.Hour
	scheduleTransitionOutboxRetentionBatch = 500
	scheduleTransitionOutboxSchemaVersion  = 1
)

// CoreScheduleTransitionOutboxInput is the immutable delivery snapshot written
// in the same transaction as its scheduler-owned story state.
type CoreScheduleTransitionOutboxInput struct {
	EventID             uuid.UUID
	StoryID             uuid.UUID
	WorkspaceID         uuid.UUID
	ActorID             uuid.UUID
	SemanticFingerprint string
	EventPayload        json.RawMessage
	ClaimImmediately    bool
}

// CoreScheduleTransitionOutboxEvent is one claimed durable StoryUpdated event.
// The raw envelope is retained so retries publish the original timestamp and
// audience snapshot instead of reconstructing mutable state.
type CoreScheduleTransitionOutboxEvent struct {
	EventID             uuid.UUID
	StoryID             uuid.UUID
	WorkspaceID         uuid.UUID
	ActorID             uuid.UUID
	SemanticFingerprint string
	TransitionSequence  int64
	ClaimToken          uuid.UUID
	AttemptCount        int
	EventPayload        json.RawMessage
}

type scheduleTransitionOutboxWriter interface {
	UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx context.Context,
		storyID, workspaceID uuid.UUID,
		expectedUpdatedAt time.Time,
		status string,
		reason *string,
		stateUpdatedAt time.Time,
		locked *bool,
		outbox CoreScheduleTransitionOutboxInput,
	) (bool, *CoreScheduleTransitionOutboxEvent, error)
	CompleteScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID) error
	RetryScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error
	FailScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string) error
}

// ScheduleTransitionOutboxRepository is the worker-facing lifecycle contract.
// It is deliberately separate from Repository so HTTP-facing test doubles do
// not inherit background-delivery concerns.
type ScheduleTransitionOutboxRepository interface {
	ClaimScheduleTransitionOutboxEvents(context.Context, int, time.Duration) ([]CoreScheduleTransitionOutboxEvent, error)
	CompleteScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID) error
	RetryScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string, time.Time) error
	FailScheduleTransitionOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string) error
	DeleteCompletedScheduleTransitionOutboxEvents(context.Context, time.Time, int) (int, error)
}

// ScheduleTransitionOutboxDispatcher publishes durable scheduler decisions and
// owns their retry lifecycle. Publisher failures retry indefinitely with a
// capped delay; only malformed immutable payloads are terminal.
type ScheduleTransitionOutboxDispatcher struct {
	repo      ScheduleTransitionOutboxRepository
	publisher eventPublisher
	log       *logger.Logger
	now       func() time.Time
}

func NewScheduleTransitionOutboxDispatcher(
	log *logger.Logger,
	repo ScheduleTransitionOutboxRepository,
	publisher eventPublisher,
) *ScheduleTransitionOutboxDispatcher {
	return &ScheduleTransitionOutboxDispatcher{
		repo:      repo,
		publisher: publisher,
		log:       log,
		now:       time.Now,
	}
}

func (s *Service) dispatchImmediateScheduleTransition(
	ctx context.Context,
	repository scheduleTransitionOutboxWriter,
	item CoreScheduleTransitionOutboxEvent,
) {
	event, err := parseScheduleTransitionOutboxEvent(item)
	if err != nil {
		lifecycleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if failErr := repository.FailScheduleTransitionOutboxEvent(
			lifecycleCtx,
			item.EventID,
			item.ClaimToken,
			err.Error(),
		); failErr != nil && s.log != nil {
			s.log.Error(ctx, "failed to terminate malformed story schedule event", "error", failErr, "event_id", item.EventID)
		}
		return
	}
	publishCtx, cancelPublish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	publishErr := s.publisher.Publish(publishCtx, event)
	cancelPublish()
	lifecycleCtx, cancelLifecycle := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancelLifecycle()
	if publishErr != nil {
		retryAt := time.Now().UTC().Add(scheduleTransitionOutboxRetryDelay(item.AttemptCount))
		if releaseErr := repository.RetryScheduleTransitionOutboxEvent(
			lifecycleCtx,
			item.EventID,
			item.ClaimToken,
			publishErr.Error(),
			retryAt,
		); releaseErr != nil && s.log != nil {
			s.log.Error(ctx, "failed to release story schedule event after publish failure", "error", releaseErr, "event_id", item.EventID)
		} else if s.log != nil {
			s.log.Error(ctx, "failed to publish Maya auto-scheduling state", "error", publishErr, "story_id", item.StoryID)
		}
		return
	}
	if err := repository.CompleteScheduleTransitionOutboxEvent(lifecycleCtx, item.EventID, item.ClaimToken); err != nil && s.log != nil {
		// An uncertain completion is recovered from the stale processing claim.
		s.log.Error(ctx, "failed to complete published story schedule event", "error", err, "event_id", item.EventID)
	}
}

// DispatchReadyScheduleTransitionOutbox publishes one bounded batch and then
// removes a bounded batch of old completed rows.
func (d *ScheduleTransitionOutboxDispatcher) DispatchReadyScheduleTransitionOutbox(ctx context.Context) (int, error) {
	if d == nil || d.repo == nil || d.publisher == nil {
		return 0, errors.New("story schedule transition outbox is unavailable")
	}
	claimed, err := d.repo.ClaimScheduleTransitionOutboxEvents(
		ctx,
		scheduleTransitionOutboxBatchSize,
		scheduleTransitionOutboxStaleAfter,
	)
	if err != nil {
		return 0, fmt.Errorf("claim story schedule transition outbox: %w", err)
	}

	delivered := 0
	var lifecycleErrors error
	for _, item := range claimed {
		event, parseErr := parseScheduleTransitionOutboxEvent(item)
		if parseErr != nil {
			failCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			if err := d.repo.FailScheduleTransitionOutboxEvent(
				failCtx,
				item.EventID,
				item.ClaimToken,
				parseErr.Error(),
			); err != nil {
				lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("fail malformed story schedule event %s: %w", item.EventID, err))
			}
			cancel()
			continue
		}

		if err := d.publisher.Publish(ctx, event); err != nil {
			retryAt := d.now().UTC().Add(scheduleTransitionOutboxRetryDelay(item.AttemptCount))
			releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			releaseErr := d.repo.RetryScheduleTransitionOutboxEvent(
				releaseCtx,
				item.EventID,
				item.ClaimToken,
				err.Error(),
				retryAt,
			)
			cancel()
			if releaseErr != nil {
				lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("release story schedule event %s after publish failure: %w", item.EventID, releaseErr))
			} else if d.log != nil {
				d.log.Error(ctx, "failed to publish durable story schedule transition", "error", err, "event_id", item.EventID)
			}
			continue
		}

		if err := d.repo.CompleteScheduleTransitionOutboxEvent(ctx, item.EventID, item.ClaimToken); err != nil {
			// Completion is an uncertain write. Leave the claim untouched so
			// stale-claim recovery replays the exact, downstream-idempotent event.
			lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("complete story schedule event %s: %w", item.EventID, err))
			continue
		}
		delivered++
	}

	cutoff := d.now().UTC().Add(-scheduleTransitionOutboxRetention)
	retentionCtx, cancelRetention := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	if _, err := d.repo.DeleteCompletedScheduleTransitionOutboxEvents(
		retentionCtx,
		cutoff,
		scheduleTransitionOutboxRetentionBatch,
	); err != nil {
		lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("delete completed story schedule events: %w", err))
	}
	cancelRetention()
	return delivered, lifecycleErrors
}

func buildScheduleTransitionOutboxInput(
	event events.Event,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	locked *bool,
	schedule *events.StoryScheduleTransition,
	claimImmediately bool,
) (CoreScheduleTransitionOutboxInput, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return CoreScheduleTransitionOutboxInput{}, fmt.Errorf("encode story schedule transition event: %w", err)
	}
	storyPayload, ok := event.Payload.(events.StoryUpdatedPayload)
	if !ok || schedule == nil {
		return CoreScheduleTransitionOutboxInput{}, errors.New("story schedule transition payload is required")
	}
	fingerprintPayload := struct {
		SchemaVersion     int                             `json:"schemaVersion"`
		StoryID           uuid.UUID                       `json:"storyId"`
		WorkspaceID       uuid.UUID                       `json:"workspaceId"`
		ActorID           uuid.UUID                       `json:"actorId"`
		ExpectedUpdatedAt time.Time                       `json:"expectedUpdatedAt"`
		Status            string                          `json:"status"`
		Reason            *string                         `json:"reason"`
		Locked            *bool                           `json:"locked"`
		Schedule          *events.StoryScheduleTransition `json:"schedule"`
	}{
		SchemaVersion:     scheduleTransitionOutboxSchemaVersion,
		StoryID:           storyPayload.StoryID,
		WorkspaceID:       storyPayload.WorkspaceID,
		ActorID:           event.ActorID,
		ExpectedUpdatedAt: expectedUpdatedAt.UTC(),
		Status:            status,
		Reason:            reason,
		Locked:            locked,
		Schedule:          schedule,
	}
	semanticPayload, err := json.Marshal(fingerprintPayload)
	if err != nil {
		return CoreScheduleTransitionOutboxInput{}, fmt.Errorf("encode story schedule transition fingerprint: %w", err)
	}
	digest := sha256.Sum256(semanticPayload)
	return CoreScheduleTransitionOutboxInput{
		EventID:             uuid.New(),
		StoryID:             storyPayload.StoryID,
		WorkspaceID:         storyPayload.WorkspaceID,
		ActorID:             event.ActorID,
		SemanticFingerprint: hex.EncodeToString(digest[:]),
		EventPayload:        payload,
		ClaimImmediately:    claimImmediately,
	}, nil
}

func parseScheduleTransitionOutboxEvent(item CoreScheduleTransitionOutboxEvent) (events.Event, error) {
	var envelope struct {
		Type      events.EventType `json:"type"`
		Payload   json.RawMessage  `json:"payload"`
		Timestamp time.Time        `json:"timestamp"`
		ActorID   uuid.UUID        `json:"actor_id"`
	}
	if err := json.Unmarshal(item.EventPayload, &envelope); err != nil {
		return events.Event{}, fmt.Errorf("decode immutable story schedule event: %w", err)
	}
	var payload events.StoryUpdatedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return events.Event{}, fmt.Errorf("decode immutable story schedule payload: %w", err)
	}
	if item.EventID == uuid.Nil || item.ClaimToken == uuid.Nil || item.AttemptCount <= 0 || item.TransitionSequence <= 0 ||
		item.StoryID == uuid.Nil || item.WorkspaceID == uuid.Nil || item.ActorID == uuid.Nil ||
		envelope.Type != events.StoryUpdated || envelope.ActorID != item.ActorID || envelope.Timestamp.IsZero() ||
		payload.StoryID != item.StoryID || payload.WorkspaceID != item.WorkspaceID || payload.Schedule == nil ||
		strings.TrimSpace(item.SemanticFingerprint) == "" {
		return events.Event{}, errors.New("immutable story schedule event envelope is incomplete or inconsistent")
	}
	return events.Event{
		Type:      envelope.Type,
		Payload:   payload,
		Timestamp: envelope.Timestamp.UTC(),
		ActorID:   envelope.ActorID,
	}, nil
}

func scheduleTransitionOutboxRetryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return scheduleTransitionOutboxInitialDelay
	}
	delay := scheduleTransitionOutboxInitialDelay
	for index := 1; index < attempt && delay < scheduleTransitionOutboxMaxDelay; index++ {
		delay *= 2
		if delay >= scheduleTransitionOutboxMaxDelay {
			return scheduleTransitionOutboxMaxDelay
		}
	}
	return delay
}
