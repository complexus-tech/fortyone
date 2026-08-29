package stories

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type scheduleTransitionOutboxRepoStub struct {
	Repository

	story              CoreSingleStory
	audience           []uuid.UUID
	stored             CoreScheduleTransitionOutboxEvent
	status             string
	completeFailures   int
	retryCalls         int
	failedCalls        int
	retentionCalls     int
	standardStateCalls int
}

func (r *scheduleTransitionOutboxRepoStub) Get(_ context.Context, _, _ uuid.UUID) (CoreSingleStory, error) {
	return r.story, nil
}

func (r *scheduleTransitionOutboxRepoStub) GetNotificationAudience(_ context.Context, _, _ uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.audience...), nil
}

func (r *scheduleTransitionOutboxRepoStub) UpdateAutoSchedulingStateIfUnchanged(
	_ context.Context,
	_, _ uuid.UUID,
	_ time.Time,
	_ string,
	_ *string,
	_ time.Time,
	_ *bool,
) (bool, error) {
	r.standardStateCalls++
	return true, nil
}

func (r *scheduleTransitionOutboxRepoStub) UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
	_ context.Context,
	storyID, workspaceID uuid.UUID,
	_ time.Time,
	status string,
	reason *string,
	_ time.Time,
	locked *bool,
	outbox CoreScheduleTransitionOutboxInput,
) (bool, *CoreScheduleTransitionOutboxEvent, error) {
	r.story.AutoSchedulingStatus = status
	r.story.AutoSchedulingReason = reason
	if locked != nil {
		r.story.AutoSchedulingLocked = *locked
	}
	r.stored = CoreScheduleTransitionOutboxEvent{
		EventID:             outbox.EventID,
		StoryID:             storyID,
		WorkspaceID:         workspaceID,
		ActorID:             outbox.ActorID,
		SemanticFingerprint: outbox.SemanticFingerprint,
		TransitionSequence:  1,
		AttemptCount:        0,
		EventPayload:        append(json.RawMessage(nil), outbox.EventPayload...),
	}
	if !outbox.ClaimImmediately {
		r.status = "pending"
		return true, nil, nil
	}
	r.status = "processing"
	r.stored.AttemptCount = 1
	r.stored.ClaimToken = uuid.New()
	claim := r.stored
	return true, &claim, nil
}

func (r *scheduleTransitionOutboxRepoStub) ClaimScheduleTransitionOutboxEvents(_ context.Context, _ int, _ time.Duration) ([]CoreScheduleTransitionOutboxEvent, error) {
	if r.status != "retrying" && r.status != "processing" && r.status != "pending" {
		return []CoreScheduleTransitionOutboxEvent{}, nil
	}
	r.status = "processing"
	r.stored.AttemptCount++
	r.stored.ClaimToken = uuid.New()
	return []CoreScheduleTransitionOutboxEvent{r.stored}, nil
}

func (r *scheduleTransitionOutboxRepoStub) CompleteScheduleTransitionOutboxEvent(_ context.Context, eventID, claimToken uuid.UUID) error {
	if eventID != r.stored.EventID || claimToken != r.stored.ClaimToken {
		return errors.New("unexpected outbox claim")
	}
	if r.completeFailures > 0 {
		r.completeFailures--
		return errors.New("completion outcome is unknown")
	}
	r.status = "completed"
	return nil
}

func (r *scheduleTransitionOutboxRepoStub) RetryScheduleTransitionOutboxEvent(
	_ context.Context,
	eventID, claimToken uuid.UUID,
	_ string,
	_ time.Time,
) error {
	if eventID != r.stored.EventID || claimToken != r.stored.ClaimToken {
		return errors.New("unexpected outbox claim")
	}
	r.retryCalls++
	r.status = "retrying"
	return nil
}

func (r *scheduleTransitionOutboxRepoStub) FailScheduleTransitionOutboxEvent(_ context.Context, _, _ uuid.UUID, _ string) error {
	r.failedCalls++
	r.status = "failed"
	return nil
}

func (r *scheduleTransitionOutboxRepoStub) DeleteCompletedScheduleTransitionOutboxEvents(_ context.Context, _ time.Time, _ int) (int, error) {
	r.retentionCalls++
	return 0, nil
}

type scheduleTransitionPublisherStub struct {
	failures int
	events   []events.Event
}

func (p *scheduleTransitionPublisherStub) Publish(_ context.Context, event events.Event) error {
	p.events = append(p.events, event)
	if p.failures > 0 {
		p.failures--
		return errors.New("Redis is unavailable")
	}
	return nil
}

func TestScheduleTransitionPublishFailureIsRecoveredFromDurableOutbox(t *testing.T) {
	fixture := newScheduleTransitionOutboxFixture()
	fixture.publisher.failures = 1

	err := fixture.service.UpdateAutomationStateIfUnchanged(
		context.Background(),
		fixture.actorID,
		fixture.storyID,
		fixture.workspaceID,
		fixture.expectedUpdatedAt,
		AutoSchedulingStatusScheduled,
		stringPointer("Scheduled around the assignee's meetings."),
		nil,
		fixture.transition,
	)
	require.NoError(t, err)
	require.Equal(t, "retrying", fixture.repository.status)
	require.Equal(t, 1, fixture.repository.retryCalls)
	require.Equal(t, 1, len(fixture.publisher.events))
	require.Zero(t, fixture.repository.standardStateCalls, "schedule transitions must use the atomic state-plus-outbox write")

	dispatcher := NewScheduleTransitionOutboxDispatcher(fixture.log, fixture.repository, fixture.publisher)
	delivered, err := dispatcher.DispatchReadyScheduleTransitionOutbox(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, delivered)
	require.Equal(t, "completed", fixture.repository.status)
	require.Equal(t, 2, len(fixture.publisher.events))
	require.Equal(t, fixture.publisher.events[0], fixture.publisher.events[1], "retry must publish the exact immutable event")
	require.Equal(t, 1, fixture.repository.retentionCalls)
}

func TestScheduleTransitionCompletionUncertaintyReplaysIdempotentEvent(t *testing.T) {
	fixture := newScheduleTransitionOutboxFixture()
	fixture.repository.completeFailures = 1

	err := fixture.service.UpdateAutomationStateIfUnchanged(
		context.Background(),
		fixture.actorID,
		fixture.storyID,
		fixture.workspaceID,
		fixture.expectedUpdatedAt,
		AutoSchedulingStatusScheduled,
		stringPointer("Scheduled around the assignee's meetings."),
		nil,
		fixture.transition,
	)
	require.NoError(t, err)
	require.Equal(t, "processing", fixture.repository.status)
	require.Equal(t, 1, len(fixture.publisher.events))

	dispatcher := NewScheduleTransitionOutboxDispatcher(fixture.log, fixture.repository, fixture.publisher)
	delivered, err := dispatcher.DispatchReadyScheduleTransitionOutbox(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, delivered)
	require.Equal(t, "completed", fixture.repository.status)
	require.Equal(t, 2, len(fixture.publisher.events))

	first, err := json.Marshal(fixture.publisher.events[0])
	require.NoError(t, err)
	second, err := json.Marshal(fixture.publisher.events[1])
	require.NoError(t, err)
	require.JSONEq(t, string(first), string(second), "post-publish recovery must preserve timestamp and downstream dedupe identity")
}

func TestScheduleTransitionOutboxPermanentlyFailsOnlyMalformedPayload(t *testing.T) {
	fixture := newScheduleTransitionOutboxFixture()
	fixture.repository.status = "retrying"
	fixture.repository.stored = CoreScheduleTransitionOutboxEvent{
		EventID:             uuid.New(),
		StoryID:             fixture.storyID,
		WorkspaceID:         fixture.workspaceID,
		ActorID:             fixture.actorID,
		SemanticFingerprint: "malformed",
		TransitionSequence:  1,
		AttemptCount:        1,
		EventPayload:        json.RawMessage(`{"type":"story.updated"}`),
	}

	dispatcher := NewScheduleTransitionOutboxDispatcher(fixture.log, fixture.repository, fixture.publisher)
	delivered, err := dispatcher.DispatchReadyScheduleTransitionOutbox(context.Background())
	require.NoError(t, err)
	require.Zero(t, delivered)
	require.Equal(t, 1, fixture.repository.failedCalls)
	require.Equal(t, "failed", fixture.repository.status)
	require.Empty(t, fixture.publisher.events)
}

func TestScheduleTransitionOutboxRetryDelayIsCappedWithoutAttemptLimit(t *testing.T) {
	require.Equal(t, scheduleTransitionOutboxInitialDelay, scheduleTransitionOutboxRetryDelay(1))
	require.Equal(t, time.Minute, scheduleTransitionOutboxRetryDelay(2))
	require.Equal(t, scheduleTransitionOutboxMaxDelay, scheduleTransitionOutboxRetryDelay(1000))
}

type scheduleTransitionOutboxFixture struct {
	service           *Service
	repository        *scheduleTransitionOutboxRepoStub
	publisher         *scheduleTransitionPublisherStub
	log               *logger.Logger
	actorID           uuid.UUID
	storyID           uuid.UUID
	workspaceID       uuid.UUID
	expectedUpdatedAt time.Time
	transition        *events.StoryScheduleTransition
}

func newScheduleTransitionOutboxFixture() scheduleTransitionOutboxFixture {
	actorID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	assigneeID := uuid.New()
	expectedUpdatedAt := time.Date(2026, time.August, 15, 9, 30, 0, 0, time.UTC)
	startAt := expectedUpdatedAt.Add(2 * time.Hour)
	endAt := startAt.Add(90 * time.Minute)
	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	repository := &scheduleTransitionOutboxRepoStub{
		story: CoreSingleStory{
			ID:                    storyID,
			Workspace:             workspaceID,
			Assignee:              &assigneeID,
			AutoSchedulingEnabled: true,
			AutoSchedulingStatus:  AutoSchedulingStatusPlanning,
			UpdatedAt:             expectedUpdatedAt,
		},
		audience: []uuid.UUID{assigneeID},
	}
	publisher := &scheduleTransitionPublisherStub{}
	return scheduleTransitionOutboxFixture{
		service:           &Service{repo: repository, publisher: publisher, log: log},
		repository:        repository,
		publisher:         publisher,
		log:               log,
		actorID:           actorID,
		storyID:           storyID,
		workspaceID:       workspaceID,
		expectedUpdatedAt: expectedUpdatedAt,
		transition: &events.StoryScheduleTransition{
			Kind:      events.StoryScheduleTransitionFirstSchedule,
			UserID:    assigneeID,
			State:     events.StoryScheduleStateScheduled,
			StartAt:   &startAt,
			EndAt:     &endAt,
			Timezone:  "Africa/Harare",
			LocalDate: "2026-08-15",
		},
	}
}
