package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type integrationStatusEventRepository struct {
	Repository

	story                   CoreSingleStory
	updates                 []map[string]any
	conditionalCalls        int
	conditionalUpdateResult bool
	expectedUpdatedAt       time.Time
	activities              []CoreActivity
}

func (r *integrationStatusEventRepository) Get(context.Context, uuid.UUID, uuid.UUID) (CoreSingleStory, error) {
	return r.story, nil
}

func (r *integrationStatusEventRepository) Update(_ context.Context, _, _ uuid.UUID, updates map[string]any) error {
	r.updates = append(r.updates, cloneStoryUpdates(updates))
	return nil
}

func (r *integrationStatusEventRepository) UpdateIfUnchanged(
	_ context.Context,
	_, _ uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) (bool, error) {
	r.conditionalCalls++
	r.expectedUpdatedAt = expectedUpdatedAt
	if r.conditionalUpdateResult {
		r.updates = append(r.updates, cloneStoryUpdates(updates))
	}
	return r.conditionalUpdateResult, nil
}

func (r *integrationStatusEventRepository) GetStatusCategory(context.Context, string) (string, error) {
	return "started", nil
}

func (r *integrationStatusEventRepository) RecordActivities(_ context.Context, activities []CoreActivity) ([]CoreActivity, error) {
	r.activities = append(r.activities, activities...)
	return activities, nil
}

func (r *integrationStatusEventRepository) GetNotificationAudience(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func cloneStoryUpdates(updates map[string]any) map[string]any {
	cloned := make(map[string]any, len(updates))
	for key, value := range updates {
		cloned[key] = value
	}
	return cloned
}

type recordingStoryEventPublisher struct {
	events []events.Event
}

func (p *recordingStoryEventPublisher) Publish(_ context.Context, event events.Event) error {
	p.events = append(p.events, event)
	return nil
}

func newIntegrationStatusEventService(repo *integrationStatusEventRepository, eventPublisher eventPublisher) *Service {
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "test"),
		repo,
		nil,
		nil,
	)
	service.publisher = eventPublisher
	return service
}

func TestUpdateExternalWithReasonPublishesActualStatusTransition(t *testing.T) {
	previousStatusID := uuid.New()
	nextStatusID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	repo := &integrationStatusEventRepository{
		story: CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Status:    &previousStatusID,
		},
	}
	eventPublisher := &recordingStoryEventPublisher{}
	service := newIntegrationStatusEventService(repo, eventPublisher)

	err := service.UpdateExternalWithReason(
		context.Background(),
		actorID,
		storyID,
		workspaceID,
		map[string]any{"status_id": nextStatusID},
		"GitHub moved the story",
	)
	if err != nil {
		t.Fatalf("update external story status: %v", err)
	}

	if len(repo.updates) != 1 {
		t.Fatalf("repository updates=%d, want 1", len(repo.updates))
	}
	assertPublishedStatusTransition(t, eventPublisher.events, actorID, storyID, workspaceID, previousStatusID, nextStatusID)
}

func TestUpdateExternalWithReasonSkipsEventForUnchangedStatus(t *testing.T) {
	statusID := uuid.New()
	repo := &integrationStatusEventRepository{
		story: CoreSingleStory{Status: &statusID},
	}
	eventPublisher := &recordingStoryEventPublisher{}
	service := newIntegrationStatusEventService(repo, eventPublisher)

	err := service.UpdateExternalWithReason(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		map[string]any{"status_id": statusID},
		"GitHub moved the story",
	)
	if err != nil {
		t.Fatalf("update unchanged external story status: %v", err)
	}

	if len(repo.updates) != 0 {
		t.Fatalf("repository updates=%d, want 0", len(repo.updates))
	}
	if len(eventPublisher.events) != 0 {
		t.Fatalf("published events=%d, want 0", len(eventPublisher.events))
	}
}

func TestUpdateExternalWithReasonKeepsNonStatusIntegrationUpdatesEventSilent(t *testing.T) {
	repo := &integrationStatusEventRepository{
		story: CoreSingleStory{Title: "Before"},
	}
	eventPublisher := &recordingStoryEventPublisher{}
	service := newIntegrationStatusEventService(repo, eventPublisher)

	err := service.UpdateExternalWithReason(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		map[string]any{"title": "After"},
		"GitHub changed the title",
	)
	if err != nil {
		t.Fatalf("update external story title: %v", err)
	}

	if len(repo.updates) != 1 {
		t.Fatalf("repository updates=%d, want 1", len(repo.updates))
	}
	if len(eventPublisher.events) != 0 {
		t.Fatalf("published events=%d, want 0", len(eventPublisher.events))
	}
}

func TestUpdateExternalIfUnchangedPublishesOnlyAfterConditionalStatusUpdate(t *testing.T) {
	previousStatusID := uuid.New()
	nextStatusID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	expectedUpdatedAt := time.Date(2026, time.August, 13, 9, 30, 0, 0, time.FixedZone("CAT", 2*60*60))
	repo := &integrationStatusEventRepository{
		story: CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Status:    &previousStatusID,
		},
		conditionalUpdateResult: true,
	}
	eventPublisher := &recordingStoryEventPublisher{}
	service := newIntegrationStatusEventService(repo, eventPublisher)

	err := service.UpdateExternalIfUnchanged(
		context.Background(),
		actorID,
		storyID,
		workspaceID,
		expectedUpdatedAt,
		map[string]any{"status_id": nextStatusID},
	)
	if err != nil {
		t.Fatalf("conditionally update external story status: %v", err)
	}

	if repo.conditionalCalls != 1 {
		t.Fatalf("conditional update calls=%d, want 1", repo.conditionalCalls)
	}
	if !repo.expectedUpdatedAt.Equal(expectedUpdatedAt.UTC()) || repo.expectedUpdatedAt.Location() != time.UTC {
		t.Fatalf("expected update time=%v, want UTC %v", repo.expectedUpdatedAt, expectedUpdatedAt.UTC())
	}
	assertPublishedStatusTransition(t, eventPublisher.events, actorID, storyID, workspaceID, previousStatusID, nextStatusID)
}

func TestUpdateExternalIfUnchangedSkipsEventWhenCompareAndSwapFails(t *testing.T) {
	previousStatusID := uuid.New()
	repo := &integrationStatusEventRepository{
		story: CoreSingleStory{Status: &previousStatusID},
	}
	eventPublisher := &recordingStoryEventPublisher{}
	service := newIntegrationStatusEventService(repo, eventPublisher)

	err := service.UpdateExternalIfUnchanged(
		context.Background(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
		time.Now(),
		map[string]any{"status_id": uuid.New()},
	)
	if !errors.Is(err, ErrStoryChanged) {
		t.Fatalf("error=%v, want %v", err, ErrStoryChanged)
	}
	if len(eventPublisher.events) != 0 {
		t.Fatalf("published events=%d, want 0", len(eventPublisher.events))
	}
}

func assertPublishedStatusTransition(
	t *testing.T,
	publishedEvents []events.Event,
	actorID, storyID, workspaceID, previousStatusID, nextStatusID uuid.UUID,
) {
	t.Helper()
	if len(publishedEvents) != 1 {
		t.Fatalf("published events=%d, want 1", len(publishedEvents))
	}
	event := publishedEvents[0]
	if event.Type != events.StoryUpdated {
		t.Fatalf("event type=%q, want %q", event.Type, events.StoryUpdated)
	}
	if event.ActorID != actorID {
		t.Fatalf("event actor=%s, want %s", event.ActorID, actorID)
	}
	if event.Timestamp.IsZero() {
		t.Fatal("event timestamp is zero")
	}
	payload, ok := event.Payload.(events.StoryUpdatedPayload)
	if !ok {
		t.Fatalf("event payload type=%T, want events.StoryUpdatedPayload", event.Payload)
	}
	if payload.StoryID != storyID || payload.WorkspaceID != workspaceID {
		t.Fatalf("payload story/workspace=%s/%s, want %s/%s", payload.StoryID, payload.WorkspaceID, storyID, workspaceID)
	}
	if payload.PreviousStatusID == nil || *payload.PreviousStatusID != previousStatusID {
		t.Fatalf("previous status=%v, want %s", payload.PreviousStatusID, previousStatusID)
	}
	statusID, ok := payload.Updates["status_id"].(uuid.UUID)
	if !ok || statusID != nextStatusID {
		t.Fatalf("next status=%v, want %s", payload.Updates["status_id"], nextStatusID)
	}
}
