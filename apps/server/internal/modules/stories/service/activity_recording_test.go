package stories

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type activityRecordingRepo struct {
	Repository

	activities []CoreActivity
	removed    CoreStoryAssociation
}

func (r *activityRecordingRepo) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID, labels []uuid.UUID) error {
	return nil
}

func (r *activityRecordingRepo) AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	return CoreStoryAssociation{
		ID:          uuid.New(),
		FromStoryID: fromID,
		ToStoryID:   toID,
		Type:        associationType,
		Story: CoreStoryList{
			ID:         toID,
			SequenceID: 62,
			Title:      "Related story",
		},
	}, nil
}

func (r *activityRecordingRepo) UpdateAssociation(ctx context.Context, associationID, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	return CoreStoryAssociation{
		ID:          associationID,
		FromStoryID: fromID,
		ToStoryID:   toID,
		Type:        associationType,
		Story: CoreStoryList{
			ID:         toID,
			SequenceID: 63,
			Title:      "Updated related story",
		},
	}, nil
}

func (r *activityRecordingRepo) RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	if r.removed.ID != uuid.Nil {
		return r.removed, nil
	}
	return CoreStoryAssociation{
		ID:          associationID,
		FromStoryID: uuid.New(),
		ToStoryID:   uuid.New(),
		Type:        "related",
	}, nil
}

func (r *activityRecordingRepo) RecordActivities(ctx context.Context, activities []CoreActivity) ([]CoreActivity, error) {
	r.activities = append(r.activities, activities...)
	return activities, nil
}

func newActivityRecordingService(repo *activityRecordingRepo) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)
}

func TestUpdateLabelsRecordsActivity(t *testing.T) {
	repo := &activityRecordingRepo{}
	service := newActivityRecordingService(repo)
	actorID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	labelID := uuid.New()

	ctx := auth.SetUserID(context.Background(), actorID)
	if err := service.UpdateLabels(ctx, storyID, workspaceID, []uuid.UUID{labelID}); err != nil {
		t.Fatalf("expected labels to update, got error: %v", err)
	}

	if len(repo.activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(repo.activities))
	}
	activity := repo.activities[0]
	if activity.StoryID != storyID {
		t.Fatalf("expected activity story %s, got %s", storyID, activity.StoryID)
	}
	if activity.UserID != actorID {
		t.Fatalf("expected activity user %s, got %s", actorID, activity.UserID)
	}
	if activity.WorkspaceID != workspaceID {
		t.Fatalf("expected activity workspace %s, got %s", workspaceID, activity.WorkspaceID)
	}
	if activity.Type != "update" {
		t.Fatalf("expected update activity, got %q", activity.Type)
	}
	if activity.Field != "labels" {
		t.Fatalf("expected labels field, got %q", activity.Field)
	}
	if activity.CurrentValue == "" {
		t.Fatal("expected current value to describe label change")
	}
}

func TestAddAssociationRecordsActivityForBothStories(t *testing.T) {
	repo := &activityRecordingRepo{}
	service := newActivityRecordingService(repo)
	actorID := uuid.New()
	fromStoryID := uuid.New()
	toStoryID := uuid.New()
	workspaceID := uuid.New()

	ctx := auth.SetUserID(context.Background(), actorID)
	if _, err := service.AddAssociation(ctx, fromStoryID, toStoryID, "blocking", workspaceID); err != nil {
		t.Fatalf("expected association to be added, got error: %v", err)
	}

	if len(repo.activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(repo.activities))
	}

	activityByStoryID := map[uuid.UUID]CoreActivity{}
	for _, activity := range repo.activities {
		activityByStoryID[activity.StoryID] = activity
	}

	fromActivity, ok := activityByStoryID[fromStoryID]
	if !ok {
		t.Fatalf("expected activity for source story %s", fromStoryID)
	}
	if fromActivity.Field != "blocking_id" {
		t.Fatalf("expected blocking_id field, got %q", fromActivity.Field)
	}

	toActivity, ok := activityByStoryID[toStoryID]
	if !ok {
		t.Fatalf("expected activity for target story %s", toStoryID)
	}
	if toActivity.Field != "blocked_by_id" {
		t.Fatalf("expected blocked_by_id field, got %q", toActivity.Field)
	}
}

func TestUpdateAssociationRecordsActivityForBothStories(t *testing.T) {
	repo := &activityRecordingRepo{}
	service := newActivityRecordingService(repo)
	actorID := uuid.New()
	associationID := uuid.New()
	fromStoryID := uuid.New()
	toStoryID := uuid.New()
	workspaceID := uuid.New()

	ctx := auth.SetUserID(context.Background(), actorID)
	if _, err := service.UpdateAssociation(ctx, associationID, fromStoryID, toStoryID, "duplicate", workspaceID); err != nil {
		t.Fatalf("expected association to update, got error: %v", err)
	}

	if len(repo.activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(repo.activities))
	}

	activityByStoryID := map[uuid.UUID]CoreActivity{}
	for _, activity := range repo.activities {
		activityByStoryID[activity.StoryID] = activity
	}

	fromActivity := activityByStoryID[fromStoryID]
	if fromActivity.Field != "duplicate_id" {
		t.Fatalf("expected duplicate_id field, got %q", fromActivity.Field)
	}

	toActivity := activityByStoryID[toStoryID]
	if toActivity.Field != "duplicated_by_id" {
		t.Fatalf("expected duplicated_by_id field, got %q", toActivity.Field)
	}
}

func TestRemoveAssociationRecordsActivityForBothStories(t *testing.T) {
	associationID := uuid.New()
	fromStoryID := uuid.New()
	toStoryID := uuid.New()
	repo := &activityRecordingRepo{
		removed: CoreStoryAssociation{
			ID:          associationID,
			FromStoryID: fromStoryID,
			ToStoryID:   toStoryID,
			Type:        "related",
		},
	}
	service := newActivityRecordingService(repo)
	actorID := uuid.New()
	workspaceID := uuid.New()

	ctx := auth.SetUserID(context.Background(), actorID)
	if err := service.RemoveAssociation(ctx, associationID, workspaceID); err != nil {
		t.Fatalf("expected association to remove, got error: %v", err)
	}

	if len(repo.activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(repo.activities))
	}

	activityByStoryID := map[uuid.UUID]CoreActivity{}
	for _, activity := range repo.activities {
		activityByStoryID[activity.StoryID] = activity
	}

	if _, ok := activityByStoryID[fromStoryID]; !ok {
		t.Fatalf("expected activity for source story %s", fromStoryID)
	}
	if _, ok := activityByStoryID[toStoryID]; !ok {
		t.Fatalf("expected activity for target story %s", toStoryID)
	}
}
