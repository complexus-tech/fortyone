package calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestManualResizeEnqueuesStoryScheduleReconciliationAfterPersistence(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	actorID := uuid.New()
	blockID := uuid.New()
	storyID := uuid.New()
	startAt := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	repo := &fakeRepo{manualRescheduleResult: ManualScheduleBlockResult{
		Block:                    CoreScheduleBlock{ID: blockID, WorkspaceID: workspaceID, UserID: userID},
		StoryScheduleReconcileID: &storyID,
	}}
	tasks := &fakeCalendarTasks{}
	updates := &fakeCalendarUpdates{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks, Updates: updates})

	block, err := service.ManuallyRescheduleScheduleBlock(context.Background(), ManualScheduleBlockInput{
		WorkspaceID:      workspaceID,
		UserID:           userID,
		ActorID:          actorID,
		BlockID:          blockID,
		StartAt:          startAt,
		EndAt:            startAt.Add(90 * time.Minute),
		Change:           ManualScheduleBlockChangeResize,
		ClientMutationID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("ManuallyRescheduleScheduleBlock returned error: %v", err)
	}
	if block.ID != blockID {
		t.Fatalf("returned block ID = %s, want %s", block.ID, blockID)
	}
	if len(repo.manualRescheduleInputs) != 1 {
		t.Fatalf("manual repository calls = %d, want 1", len(repo.manualRescheduleInputs))
	}
	if len(tasks.storyWorkspaceIDs) != 1 || tasks.storyWorkspaceIDs[0] != workspaceID || len(tasks.storyIDs) != 1 || tasks.storyIDs[0] != storyID {
		t.Fatalf("story schedule reconciliation scope = (%v, %v), want (%s, %s)", tasks.storyWorkspaceIDs, tasks.storyIDs, workspaceID, storyID)
	}
	if updates.calls != 1 {
		t.Fatalf("calendar invalidation calls = %d, want 1", updates.calls)
	}
}

func TestManualRescheduleOnlyEnqueuesStoryReconciliationForAppliedEstimateChange(t *testing.T) {
	t.Parallel()

	for _, change := range []ManualScheduleBlockChange{ManualScheduleBlockChangeMove, ManualScheduleBlockChangeResize} {
		change := change
		t.Run(string(change), func(t *testing.T) {
			t.Parallel()
			workspaceID := uuid.New()
			userID := uuid.New()
			blockID := uuid.New()
			startAt := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
			repo := &fakeRepo{manualRescheduleResult: ManualScheduleBlockResult{
				Block: CoreScheduleBlock{ID: blockID, WorkspaceID: workspaceID, UserID: userID},
			}}
			tasks := &fakeCalendarTasks{storyScheduleErr: errors.New("task queue unavailable")}
			service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks})

			_, err := service.ManuallyRescheduleScheduleBlock(context.Background(), ManualScheduleBlockInput{
				WorkspaceID:      workspaceID,
				UserID:           userID,
				ActorID:          uuid.New(),
				BlockID:          blockID,
				StartAt:          startAt,
				EndAt:            startAt.Add(time.Hour),
				Change:           change,
				ClientMutationID: uuid.New(),
			})
			if err != nil {
				t.Fatalf("ManuallyRescheduleScheduleBlock returned error: %v", err)
			}
			if len(tasks.storyIDs) != 0 {
				t.Fatalf("story reconciliation was enqueued without an applied estimate change: %v", tasks.storyIDs)
			}
		})
	}
}

func TestManualResizeDoesNotFailCommittedChangeWhenStoryReconciliationEnqueueFails(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	storyID := uuid.New()
	startAt := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	repo := &fakeRepo{manualRescheduleResult: ManualScheduleBlockResult{
		Block:                    CoreScheduleBlock{ID: blockID, WorkspaceID: workspaceID, UserID: userID},
		StoryScheduleReconcileID: &storyID,
	}}
	tasks := &fakeCalendarTasks{storyScheduleErr: errors.New("task queue unavailable")}
	service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks})

	_, err := service.ManuallyRescheduleScheduleBlock(context.Background(), ManualScheduleBlockInput{
		WorkspaceID:      workspaceID,
		UserID:           userID,
		ActorID:          uuid.New(),
		BlockID:          blockID,
		StartAt:          startAt,
		EndAt:            startAt.Add(time.Hour),
		Change:           ManualScheduleBlockChangeResize,
		ClientMutationID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("committed resize must remain successful after task enqueue failure: %v", err)
	}
	if len(tasks.storyIDs) != 1 || tasks.storyIDs[0] != storyID {
		t.Fatalf("story reconciliation attempts = %v, want [%s]", tasks.storyIDs, storyID)
	}
}

func TestCreateScheduleBlockValidatesRangeAndType(t *testing.T) {
	t.Parallel()

	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret"})
	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startAt.Add(-time.Hour) }

	if _, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		BlockType:   ScheduleBlockTypeWork,
		Title:       "Invalid",
		StartAt:     startAt,
		EndAt:       startAt,
		IsLocked:    true,
		Source:      ScheduleBlockSourceUser,
	}); err == nil {
		t.Fatal("expected invalid range error")
	}

	block, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		BlockType:   ScheduleBlockTypeFocus,
		Title:       "Deep work",
		StartAt:     startAt,
		EndAt:       startAt.Add(90 * time.Minute),
		IsLocked:    true,
		Source:      ScheduleBlockSourceUser,
	})
	if err != nil {
		t.Fatalf("CreateScheduleBlock returned error: %v", err)
	}
	if block.Title != "Deep work" || block.BlockType != ScheduleBlockTypeFocus || !block.IsLocked {
		t.Fatalf("unexpected schedule block: %#v", block)
	}
}

func TestCreateScheduleBlockRejectsTimesOutsideSyncedCoverage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret"})
	service.now = func() time.Time { return now }
	workspaceID := uuid.New()
	userID := uuid.New()

	for _, startAt := range []time.Time{
		now.Add(defaultSyncLookback).Add(-time.Minute),
		now.Add(defaultSyncLookahead).Add(-30 * time.Minute),
	} {
		_, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
			WorkspaceID: workspaceID,
			UserID:      userID,
			BlockType:   ScheduleBlockTypeFocus,
			Title:       "Outside coverage",
			StartAt:     startAt,
			EndAt:       startAt.Add(time.Hour),
		})
		if !errors.Is(err, ErrInvalidScheduleRange) {
			t.Fatalf("expected invalid range for %s, got %v", startAt, err)
		}
	}
}

func TestCreateScheduleBlockRejectsStoryOutsideWorkspace(t *testing.T) {
	t.Parallel()

	allowed := false
	storyID := uuid.New()
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &fakeRepo{storyAllowed: &allowed}
	service := New(nil, repo, Config{SecretKey: "test-secret"})
	_, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     &storyID,
		BlockType:   ScheduleBlockTypeWork,
		Title:       "Cross-workspace story",
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, ErrInvalidScheduleBlock) {
		t.Fatalf("expected invalid schedule block, got %v", err)
	}
	if repo.storyUserID != userID {
		t.Fatalf("story authorization checked the wrong user: got %s want %s", repo.storyUserID, userID)
	}
}
