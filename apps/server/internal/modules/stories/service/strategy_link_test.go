package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type strategyLinkRepo struct {
	Repository
	story      CoreSingleStory
	keyResult  CoreKeyResultReference
	updates    map[string]any
	activities []CoreActivity
}

func (r *strategyLinkRepo) Get(context.Context, uuid.UUID, uuid.UUID) (CoreSingleStory, error) {
	return r.story, nil
}

func (r *strategyLinkRepo) ResolveKeyResult(context.Context, uuid.UUID, uuid.UUID) (CoreKeyResultReference, error) {
	return r.keyResult, nil
}

func (r *strategyLinkRepo) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, updates map[string]any) error {
	r.updates = updates
	return nil
}

func (r *strategyLinkRepo) RecordActivities(_ context.Context, activities []CoreActivity) ([]CoreActivity, error) {
	r.activities = append(r.activities, activities...)
	return activities, nil
}

func newStrategyLinkService(repo Repository) *Service {
	return New(
		logger.NewWithText(io.Discard, slog.LevelError, "test"),
		repo,
		nil,
		nil,
	)
}

func TestKeyResultUpdateAlignsItsObjective(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	keyResultID := uuid.New()
	objectiveID := uuid.New()
	repo := &strategyLinkRepo{
		story:     CoreSingleStory{ID: storyID, Workspace: workspaceID},
		keyResult: CoreKeyResultReference{ObjectiveID: objectiveID, Name: "Grow enterprise adoption"},
	}

	err := newStrategyLinkService(repo).updateWithOptions(
		context.Background(),
		storyID,
		workspaceID,
		uuid.New(),
		map[string]any{"key_result_id": &keyResultID},
		updateOptions{},
	)
	if err != nil {
		t.Fatalf("expected strategy link update to succeed, got %v", err)
	}

	updatedObjectiveID, ok := repo.updates["objective_id"].(*uuid.UUID)
	if !ok || updatedObjectiveID == nil || *updatedObjectiveID != objectiveID {
		t.Fatalf("expected objective %s, got %#v", objectiveID, repo.updates["objective_id"])
	}

	var keyResultActivity *CoreActivity
	for i := range repo.activities {
		if repo.activities[i].Field == "key_result_id" {
			keyResultActivity = &repo.activities[i]
			break
		}
	}
	if keyResultActivity == nil {
		t.Fatal("expected a key result activity to be recorded")
	}
	if keyResultActivity.CurrentValue != "Grow enterprise adoption" {
		t.Fatalf("expected key result name in activity, got %q", keyResultActivity.CurrentValue)
	}
}

func TestMatchingObjectiveAndKeyResultUpdateSucceeds(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	keyResultID := uuid.New()
	objectiveID := uuid.New()
	repo := &strategyLinkRepo{
		story:     CoreSingleStory{ID: storyID, Workspace: workspaceID},
		keyResult: CoreKeyResultReference{ObjectiveID: objectiveID, Name: "Grow enterprise adoption"},
	}

	err := newStrategyLinkService(repo).updateWithOptions(
		context.Background(),
		storyID,
		workspaceID,
		uuid.New(),
		map[string]any{
			"objective_id":  &objectiveID,
			"key_result_id": &keyResultID,
		},
		updateOptions{},
	)
	if err != nil {
		t.Fatalf("expected matching strategy links to succeed, got %v", err)
	}

	updatedObjectiveID, ok := repo.updates["objective_id"].(*uuid.UUID)
	if !ok || updatedObjectiveID == nil || *updatedObjectiveID != objectiveID {
		t.Fatalf("expected objective %s, got %#v", objectiveID, repo.updates["objective_id"])
	}
}

func TestMismatchedObjectiveAndKeyResultUpdateIsRejected(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	keyResultID := uuid.New()
	objectiveID := uuid.New()
	differentObjectiveID := uuid.New()
	repo := &strategyLinkRepo{
		story:     CoreSingleStory{ID: storyID, Workspace: workspaceID},
		keyResult: CoreKeyResultReference{ObjectiveID: objectiveID, Name: "Grow enterprise adoption"},
	}

	err := newStrategyLinkService(repo).updateWithOptions(
		context.Background(),
		storyID,
		workspaceID,
		uuid.New(),
		map[string]any{
			"objective_id":  &differentObjectiveID,
			"key_result_id": &keyResultID,
		},
		updateOptions{},
	)
	if !errors.Is(err, ErrObjectiveKeyResultMismatch) {
		t.Fatalf("expected objective/key-result mismatch, got %v", err)
	}
	if repo.updates != nil {
		t.Fatalf("expected mismatched links not to be persisted, got %#v", repo.updates)
	}
}

func TestMismatchedObjectiveAndKeyResultCreateIsRejected(t *testing.T) {
	workspaceID := uuid.New()
	keyResultID := uuid.New()
	objectiveID := uuid.New()
	differentObjectiveID := uuid.New()
	reporterID := uuid.New()
	repo := &strategyLinkRepo{
		keyResult: CoreKeyResultReference{ObjectiveID: objectiveID, Name: "Grow enterprise adoption"},
	}

	_, err := newStrategyLinkService(repo).createWithOptions(
		context.Background(),
		CoreNewStory{
			Objective: &differentObjectiveID,
			KeyResult: &keyResultID,
			Reporter:  &reporterID,
		},
		workspaceID,
		reporterID,
		createOptions{},
	)
	if !errors.Is(err, ErrObjectiveKeyResultMismatch) {
		t.Fatalf("expected objective/key-result mismatch, got %v", err)
	}
}

func TestObjectiveChangeClearsIncompatibleKeyResult(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	currentObjectiveID := uuid.New()
	keyResultID := uuid.New()
	nextObjectiveID := uuid.New()
	repo := &strategyLinkRepo{
		story: CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Objective: &currentObjectiveID,
			KeyResult: &keyResultID,
		},
	}

	err := newStrategyLinkService(repo).updateWithOptions(
		context.Background(),
		storyID,
		workspaceID,
		uuid.New(),
		map[string]any{"objective_id": &nextObjectiveID},
		updateOptions{},
	)
	if err != nil {
		t.Fatalf("expected objective update to succeed, got %v", err)
	}

	keyResultUpdate, exists := repo.updates["key_result_id"]
	if !exists || keyResultUpdate != nil {
		t.Fatalf("expected key result to be cleared, got %#v", keyResultUpdate)
	}
}
