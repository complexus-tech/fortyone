package storiesrepository

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestChecklistActivitiesRetainActorTimeAndItemChanges(t *testing.T) {
	for _, test := range []struct {
		name               string
		action             storydomain.ChecklistAction
		title              string
		completed          *bool
		initiallyCompleted bool
		reasons            []string
	}{
		{name: "add", action: storydomain.ChecklistAdd, title: "Review release", reasons: []string{"checklist_item_added"}},
		{name: "rename", action: storydomain.ChecklistUpdate, title: "Review release", reasons: []string{"checklist_item_updated"}},
		{name: "complete", action: storydomain.ChecklistUpdate, completed: checklistBool(true), reasons: []string{"checklist_item_completed"}},
		{name: "reopen", action: storydomain.ChecklistUpdate, initiallyCompleted: true, completed: checklistBool(false), reasons: []string{"checklist_item_reopened"}},
		{name: "rename and complete", action: storydomain.ChecklistUpdate, title: "Review release", completed: checklistBool(true), reasons: []string{"checklist_item_updated", "checklist_item_completed"}},
		{name: "rename and reopen", action: storydomain.ChecklistUpdate, title: "Review release", initiallyCompleted: true, completed: checklistBool(false), reasons: []string{"checklist_item_updated", "checklist_item_reopened"}},
		{name: "remove", action: storydomain.ChecklistDelete, reasons: []string{"checklist_item_removed"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, f, command := newChecklistFixture(t)
			f.items[0].Completed = test.initiallyCompleted
			command.Action = test.action
			command.Completed = test.completed
			if test.title != "" {
				command.Title = &test.title
			}
			oldValue := checklistExpectedSnapshot(command.ItemID, "Ship it", test.initiallyCompleted, 0)
			newValue := checklistExpectedSnapshot(command.ItemID, "Ship it", test.initiallyCompleted, 0)
			if test.title != "" {
				newValue["title"] = test.title
			}
			if test.completed != nil {
				newValue["completed"] = *test.completed
			}
			if test.action == storydomain.ChecklistAdd {
				command.ItemID = uuid.New()
				oldValue = nil
				newValue = checklistExpectedSnapshot(command.ItemID, test.title, false, 1)
			}
			if test.action == storydomain.ChecklistDelete {
				newValue = nil
			}
			result, err := r.ApplyChecklistMutation(context.Background(), command)
			if err != nil || !result.Changed {
				t.Fatalf("mutation failed: %+v %v", result, err)
			}
			if len(f.activities) != len(test.reasons) {
				t.Fatalf("got %d activities, want %d", len(f.activities), len(test.reasons))
			}
			for i, activity := range f.activities {
				assertChecklistActivityIdentity(t, activity, command, "checklist_item", test.reasons[i])
				assertChecklistActivityJSON(t, activity.OldValue, oldValue)
				assertChecklistActivityJSON(t, activity.NewValue, newValue)
				wantTitle := "Ship it"
				if test.title != "" {
					wantTitle = test.title
				}
				if activity.CurrentValue != wantTitle {
					t.Fatalf("activity lost item title: %+v", activity)
				}
			}
		})
	}
}

func TestChecklistClearRetainsEachRemovedItem(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	secondID := uuid.New()
	f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: secondID, StoryID: f.storyID, Title: "Already done", Completed: true, Position: 7})
	command.Action = storydomain.ChecklistClear
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || len(result.Items) != 0 || len(f.activities) != 2 {
		t.Fatalf("clear lost removal history: %+v %+v %v", result, f.activities, err)
	}
	want := []map[string]any{
		checklistExpectedSnapshot(command.ItemID, "Ship it", false, 0),
		checklistExpectedSnapshot(secondID, "Already done", true, 7),
	}
	for i, activity := range f.activities {
		assertChecklistActivityIdentity(t, activity, command, "checklist_item", "checklist_item_removed")
		assertChecklistActivityJSON(t, activity.OldValue, want[i])
		assertChecklistActivityJSON(t, activity.NewValue, nil)
		if activity.CurrentValue != want[i]["title"] {
			t.Fatalf("removed item title lost: %+v", activity)
		}
	}
}

func TestChecklistReorderRetainsPreviousAndCurrentOrder(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	secondID := uuid.New()
	f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: secondID, StoryID: f.storyID, Title: "Review", Position: 7})
	command.Action = storydomain.ChecklistReorder
	command.ItemIDs = []uuid.UUID{secondID, command.ItemID}
	if _, err := r.ApplyChecklistMutation(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if len(f.activities) != 1 {
		t.Fatalf("got %d reorder activities", len(f.activities))
	}
	activity := f.activities[0]
	assertChecklistActivityIdentity(t, activity, command, "checklist", "checklist_reordered")
	assertChecklistActivityJSON(t, activity.OldValue, []map[string]any{
		checklistExpectedSnapshot(command.ItemID, "Ship it", false, 0),
		checklistExpectedSnapshot(secondID, "Review", false, 7),
	})
	assertChecklistActivityJSON(t, activity.NewValue, []map[string]any{
		checklistExpectedSnapshot(secondID, "Review", false, 0),
		checklistExpectedSnapshot(command.ItemID, "Ship it", false, 1),
	})
	if activity.CurrentValue != "" {
		t.Fatalf("unexpected reorder title: %q", activity.CurrentValue)
	}
}

func TestChecklistConversionRetainsSourceAndCreatedStoryActivity(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || result.CreatedStory == nil || len(f.activities) != 2 {
		t.Fatalf("conversion activities missing: %+v %+v %v", result, f.activities, err)
	}
	createdActivity, sourceActivity := f.activities[0], f.activities[1]
	if createdActivity.ActivityType != "create" || createdActivity.FieldChanged != "story" || createdActivity.StoryID != result.CreatedStory.ID {
		t.Fatalf("created story activity changed: %+v", createdActivity)
	}
	assertChecklistActivityIdentity(t, sourceActivity, command, "checklist_item", "checklist_item_converted")
	assertChecklistActivityJSON(t, sourceActivity.OldValue, checklistExpectedSnapshot(command.ItemID, "Ship it", false, 0))
	wantNew := checklistExpectedSnapshot(command.ItemID, "Ship it", true, 0)
	wantNew["convertedStoryId"] = result.CreatedStory.ID.String()
	wantNew["convertedAt"] = command.Event.OccurredAt.Format(time.RFC3339Nano)
	assertChecklistActivityJSON(t, sourceActivity.NewValue, wantNew)
	if _, err := r.ApplyChecklistMutation(context.Background(), command); err != nil || len(f.activities) != 2 {
		t.Fatalf("repeat conversion duplicated history: %+v %v", f.activities, err)
	}
}

func TestChecklistNoOpChangesDoNotWriteActivities(t *testing.T) {
	for _, action := range []storydomain.ChecklistAction{storydomain.ChecklistUpdate, storydomain.ChecklistClear, storydomain.ChecklistReorder} {
		t.Run(string(action), func(t *testing.T) {
			r, f, command := newChecklistFixture(t)
			command.Action = action
			command.Title = &f.items[0].Title
			command.Completed = checklistBool(false)
			command.ItemIDs = []uuid.UUID{command.ItemID}
			if action == storydomain.ChecklistClear {
				f.items = nil
			}
			result, err := r.ApplyChecklistMutation(context.Background(), command)
			if err != nil || result.Changed || len(f.activities) != 0 || f.events != 0 || f.touched != 0 {
				t.Fatalf("no-op wrote side effects: %+v %+v %v", result, f, err)
			}
		})
	}
}

func TestChecklistActivityFailureRollsBackItemChanges(t *testing.T) {
	for _, action := range []storydomain.ChecklistAction{storydomain.ChecklistAdd, storydomain.ChecklistUpdate, storydomain.ChecklistDelete, storydomain.ChecklistClear, storydomain.ChecklistReorder} {
		t.Run(string(action), func(t *testing.T) {
			r, f, command := newChecklistFixture(t)
			secondID := uuid.New()
			f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: secondID, StoryID: f.storyID, Title: "Review", Position: 1})
			before := slices.Clone(f.items)
			f.fail = "activity"
			command.Action = action
			title := "Changed"
			command.Title = &title
			command.ItemIDs = []uuid.UUID{secondID, command.ItemID}
			if action == storydomain.ChecklistAdd {
				command.ItemID = uuid.New()
			}
			if result, err := r.ApplyChecklistMutation(context.Background(), command); err == nil || result.Changed {
				t.Fatalf("expected activity failure: %+v %v", result, err)
			}
			if !reflect.DeepEqual(f.items, before) || len(f.activities) != 0 || f.events != 0 || f.touched != 0 {
				t.Fatalf("activity failure committed changes: %+v", f)
			}
		})
	}
}

func TestChecklistRapidChangesDoNotCompactHistory(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	command.Action = storydomain.ChecklistUpdate
	command.Completed = checklistBool(true)
	if _, err := r.ApplyChecklistMutation(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	command.Completed = checklistBool(false)
	command.Event.ID = uuid.New()
	command.Event.OccurredAt = command.Event.OccurredAt.Add(time.Second)
	if _, err := r.ApplyChecklistMutation(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if len(f.activities) != 2 || f.activities[0].ActivityID == f.activities[1].ActivityID {
		t.Fatalf("rapid changes lost distinct history: %+v", f.activities)
	}
	for _, activity := range f.activities {
		if activity.Compact {
			t.Fatal("checklist activities must never compact")
		}
	}
	if *f.activities[0].Reason != "checklist_item_completed" || *f.activities[1].Reason != "checklist_item_reopened" {
		t.Fatalf("completion history lost: %+v", f.activities)
	}
}

func TestChecklistClearRollsBackEarlierActivitiesIfLaterActivityFails(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	f.items = append(f.items, storyreadsql.ListStoryChecklistRow{ID: uuid.New(), StoryID: f.storyID, Title: "Review", Position: 1})
	before := slices.Clone(f.items)
	f.fail = "activity_after_first"
	command.Action = storydomain.ChecklistClear
	if _, err := r.ApplyChecklistMutation(context.Background(), command); err == nil {
		t.Fatal("expected second removal activity to fail")
	}
	if !reflect.DeepEqual(f.items, before) || len(f.activities) != 0 || f.events != 0 || f.touched != 0 {
		t.Fatalf("partial clear or audit survived rollback: %+v", f)
	}
}

func TestChecklistTransactionRetryWritesHistoryOnlyOnce(t *testing.T) {
	r, f, command := newChecklistFixture(t)
	command.Action = storydomain.ChecklistUpdate
	command.Completed = checklistBool(true)
	transaction := r.runChecklistTransaction
	attempts := 0
	r.runChecklistTransaction = func(ctx context.Context, operation func(checklistQueries) error) error {
		return transaction(ctx, func(queries checklistQueries) error {
			if err := operation(queries); err != nil {
				return err
			}
			attempts++
			if attempts == 1 {
				return &pgconn.PgError{Code: "40001", Message: "serialization failure"}
			}
			return nil
		})
	}
	result, err := r.ApplyChecklistMutation(context.Background(), command)
	if err != nil || !result.Changed || attempts != 2 || len(f.activities) != 1 || f.events != 1 || f.touched != 1 {
		t.Fatalf("retry lost or duplicated audit: attempts=%d result=%+v fixture=%+v error=%v", attempts, result, f, err)
	}
	assertChecklistActivityJSON(t, f.activities[0].OldValue, checklistExpectedSnapshot(command.ItemID, "Ship it", false, 0))
	assertChecklistActivityJSON(t, f.activities[0].NewValue, checklistExpectedSnapshot(command.ItemID, "Ship it", true, 0))
}

func TestCopiedChecklistActivitiesUseTargetIdentityAndCreationTime(t *testing.T) {
	_, f, command := newChecklistFixture(t)
	duplicate := storydomain.DuplicateStoryCommand{
		Scope: command.Scope, SourceStoryID: uuid.New(), TargetStoryID: command.StoryID,
		OccurredAt: command.Event.OccurredAt,
		Activity: storydomain.MutationActivity{
			ID: uuid.New(), StoryID: command.StoryID, UserID: *command.Scope.ActivityUser,
			Type: "create", Field: "story", CurrentValue: "Copy of source",
			WorkspaceID: command.Scope.WorkspaceID, CreatedAt: command.Event.OccurredAt,
		},
	}
	if err := insertCopiedChecklistActivities(context.Background(), f, duplicate); err != nil {
		t.Fatal(err)
	}
	if len(f.activities) != 1 {
		t.Fatalf("got %d copied item activities", len(f.activities))
	}
	activity := f.activities[0]
	assertChecklistActivityIdentity(t, activity, command, "checklist_item", "checklist_item_added")
	assertChecklistActivityJSON(t, activity.OldValue, nil)
	assertChecklistActivityJSON(t, activity.NewValue, checklistExpectedSnapshot(command.ItemID, "Ship it", false, 0))
	if activity.ActivityID == duplicate.Activity.ID || activity.StoryID == duplicate.SourceStoryID {
		t.Fatalf("copied item activity reused source or creation identity: %+v", activity)
	}
}

func checklistExpectedSnapshot(id uuid.UUID, title string, completed bool, position int) map[string]any {
	return map[string]any{"id": id.String(), "title": title, "completed": completed, "position": float64(position)}
}

func checklistBool(value bool) *bool { return &value }

func assertChecklistActivityIdentity(t *testing.T, activity storyreadsql.UpsertStoryMutationActivityParams, command storydomain.ChecklistCommand, field, reason string) {
	t.Helper()
	if activity.ActivityID == uuid.Nil || activity.StoryID != command.StoryID || activity.UserID != command.Scope.Actor.PrincipalID || activity.WorkspaceID == nil || *activity.WorkspaceID != command.Scope.WorkspaceID || !activity.CreatedAt.Equal(command.Event.OccurredAt) || activity.ActivityType != "update" || activity.FieldChanged != field || activity.Reason == nil || *activity.Reason != reason || activity.Compact {
		t.Fatalf("activity metadata mismatch: %+v", activity)
	}
}

func assertChecklistActivityJSON(t *testing.T, actual json.RawMessage, expected any) {
	t.Helper()
	var decoded, want any
	if err := json.Unmarshal(actual, &decoded); err != nil {
		t.Fatalf("invalid activity JSON: %s: %v", actual, err)
	}
	encoded, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("activity snapshot = %s, want %s", actual, encoded)
	}
}
