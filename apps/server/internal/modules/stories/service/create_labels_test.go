package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type atomicCreateRepo struct {
	Repository

	createdStory      *CoreSingleStory
	idempotentStory   *CoreSingleStory
	idempotentCreated bool
	createCalls       int
	updateLabelsCalls int
	activities        []CoreActivity
}

func (r *atomicCreateRepo) GetTeamEstimateScheme(context.Context, uuid.UUID, uuid.UUID) (string, error) {
	return DefaultEstimateScheme, nil
}

func (r *atomicCreateRepo) Create(_ context.Context, story *CoreSingleStory) (CoreSingleStory, error) {
	r.createCalls++
	copy := *story
	copy.Labels = append([]uuid.UUID(nil), story.Labels...)
	r.createdStory = &copy

	created := copy
	created.ID = uuid.New()
	created.SequenceID = 1
	created.Labels = uniqueUUIDs(created.Labels)
	return created, nil
}

func (r *atomicCreateRepo) CreateIdempotent(_ context.Context, story *CoreSingleStory) (CoreSingleStory, bool, error) {
	copy := *story
	r.createdStory = &copy
	if r.idempotentStory != nil {
		return *r.idempotentStory, r.idempotentCreated, nil
	}
	copy.ID = uuid.New()
	return copy, true, nil
}

func (r *atomicCreateRepo) UpdateLabels(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error {
	r.updateLabelsCalls++
	return nil
}

func TestCreateWithExistingIdempotencyKeySkipsDuplicateSideEffects(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	creationKey := "slack:view:V123"
	existing := &CoreSingleStory{
		ID:         uuid.New(),
		SequenceID: 42,
		Title:      "Existing story",
		Team:       teamID,
		Workspace:  workspaceID,
	}
	repo := &atomicCreateRepo{idempotentStory: existing}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)

	created, err := service.createWithOptions(context.Background(), CoreNewStory{
		Title:       "Existing story",
		Team:        teamID,
		Reporter:    &actorID,
		CreationKey: &creationKey,
	}, workspaceID, actorID, createOptions{})
	if err != nil {
		t.Fatalf("create idempotent story: %v", err)
	}
	if created.ID != existing.ID {
		t.Fatalf("created story ID = %s, want existing %s", created.ID, existing.ID)
	}
	if repo.createCalls != 0 {
		t.Fatalf("ordinary Create calls = %d, want 0", repo.createCalls)
	}
	if len(repo.activities) != 0 {
		t.Fatalf("duplicate create activities = %v, want none", repo.activities)
	}
}

func (r *atomicCreateRepo) RecordActivities(_ context.Context, activities []CoreActivity) ([]CoreActivity, error) {
	r.activities = append(r.activities, activities...)
	return activities, nil
}

func TestCreatePassesLabelsThroughSingleRepositoryCreate(t *testing.T) {
	repo := &atomicCreateRepo{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	firstLabelID := uuid.New()
	secondLabelID := uuid.New()
	requestedLabels := []uuid.UUID{firstLabelID, secondLabelID, firstLabelID}

	created, err := service.createWithOptions(context.Background(), CoreNewStory{
		Title:    "Create from Slack",
		Team:     teamID,
		Reporter: &actorID,
		LabelIDs: requestedLabels,
	}, workspaceID, actorID, createOptions{})
	if err != nil {
		t.Fatalf("create story: %v", err)
	}
	if repo.createdStory == nil {
		t.Fatal("repository Create was not called")
	}
	if !reflect.DeepEqual(repo.createdStory.Labels, requestedLabels) {
		t.Fatalf("repository labels = %v, want %v", repo.createdStory.Labels, requestedLabels)
	}
	if repo.updateLabelsCalls != 0 {
		t.Fatalf("UpdateLabels calls = %d, want 0", repo.updateLabelsCalls)
	}
	wantCreatedLabels := []uuid.UUID{firstLabelID, secondLabelID}
	if !reflect.DeepEqual(created.Labels, wantCreatedLabels) {
		t.Fatalf("created labels = %v, want %v", created.Labels, wantCreatedLabels)
	}
	if len(repo.activities) != 1 || repo.activities[0].Type != "create" {
		t.Fatalf("create activities = %v, want one create activity", repo.activities)
	}
}

func TestCreateValidatesMayaAssignmentBeforePersistence(t *testing.T) {
	repo := &atomicCreateRepo{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	mayaID := uuid.New()
	wantErr := errors.New("Maya is unavailable for this workspace")
	service.ConfigureMayaAssignment(mayaID, func(context.Context, MayaAssignmentInput) error {
		return wantErr
	})
	service.ConfigureAutoSchedulingEligibility(func(context.Context, uuid.UUID) (bool, error) {
		return true, nil
	})
	durationMinutes := 60
	deliveryDate := time.Now().UTC().Add(24 * time.Hour)

	_, err := service.createWithOptions(context.Background(), CoreNewStory{
		Title:                    "Create for Maya",
		Team:                     teamID,
		Reporter:                 &actorID,
		Assignee:                 &mayaID,
		AutoSchedulingEnabled:    true,
		EstimatedDurationMinutes: &durationMinutes,
		EndDate:                  &deliveryDate,
	}, workspaceID, actorID, createOptions{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("create error = %v, want %v", err, wantErr)
	}
	if repo.createCalls != 0 || repo.createdStory != nil {
		t.Fatalf("story was persisted before Maya validation: calls=%d story=%v", repo.createCalls, repo.createdStory)
	}
}
