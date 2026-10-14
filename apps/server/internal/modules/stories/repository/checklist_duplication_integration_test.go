//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestStoryDuplicationRecordsChecklistAdditionsAtomically(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	createdAt := time.Date(2026, time.October, 14, 8, 0, 0, 0, time.UTC)
	sourceID := createSecondaryMutationStory(t, ctx, repository, fixture, createdAt)
	sourceItemID := uuid.New()
	mustMutationExec(t, ctx, postgres.Pool,
		"INSERT INTO story_checklist_items (id, story_id, title, completed, position, created_at, updated_at) VALUES ($1, $2, 'Review release', true, 3, $3, $3)",
		sourceItemID, sourceID, createdAt,
	)
	scope := mutationScopeForFixture(t, fixture)
	source, err := repository.GetStoryForMutation(ctx, scope, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	targetID := uuid.New()
	command := duplicateStoryCommand(t, scope, source, targetID, createdAt.Add(time.Minute))
	if _, err := repository.DuplicateStoryMutation(ctx, command); err != nil {
		t.Fatal(err)
	}
	var itemID, actorID uuid.UUID
	var completed bool
	var activityAt time.Time
	var oldValue, newValue json.RawMessage
	if err := postgres.Pool.QueryRow(ctx,
		`SELECT item.id, item.completed, activity.user_id, activity.created_at, activity.old_value, activity.new_value
		 FROM story_checklist_items AS item
		 JOIN story_activities AS activity ON activity.story_id = item.story_id
		 WHERE item.story_id = $1 AND activity.field_changed = 'checklist_item'
		   AND activity.reason = 'checklist_item_added' AND activity.activity_type = 'update'`,
		targetID,
	).Scan(&itemID, &completed, &actorID, &activityAt, &oldValue, &newValue); err != nil {
		t.Fatal(err)
	}
	if itemID == sourceItemID || completed || actorID != fixture.actorID || !activityAt.Equal(command.OccurredAt) {
		t.Fatalf("copied item identity, completion, or attribution incorrect: %v %v %v %v", itemID, completed, actorID, activityAt)
	}
	assertChecklistActivityJSON(t, oldValue, nil)
	assertChecklistActivityJSON(t, newValue, checklistExpectedSnapshot(itemID, "Review release", false, 3))
	assertMutationActivityCount(t, ctx, postgres, targetID, "create", 1)
	assertMutationActivityCount(t, ctx, postgres, targetID, "update", 1)
	assertMutationActivityCount(t, ctx, postgres, sourceID, "update", 0)

	rollbackTargetID := uuid.New()
	// The isolated test database rejects just this target's checklist activity,
	// after the duplicate story, copied item, and creation activity are written.
	mustMutationExec(t, ctx, postgres.Pool, fmt.Sprintf(
		"ALTER TABLE story_activities ADD CONSTRAINT reject_copied_checklist_activity CHECK (story_id <> '%s' OR field_changed <> 'checklist_item')",
		rollbackTargetID,
	))
	rollback := duplicateStoryCommand(t, scope, source, rollbackTargetID, createdAt.Add(2*time.Minute))
	if _, err := repository.DuplicateStoryMutation(ctx, rollback); err == nil {
		t.Fatal("expected copied checklist activity rejection")
	}
	assertMutationRowCount(t, ctx, postgres, "stories", "id", rollbackTargetID, 0)
	assertMutationRowCount(t, ctx, postgres, "story_checklist_items", "story_id", rollbackTargetID, 0)
	assertMutationRowCount(t, ctx, postgres, "story_activities", "story_id", rollbackTargetID, 0)
	assertMutationEventIDAbsent(t, ctx, postgres.Pool, rollback.Event.ID)
	assertMutationRowCount(t, ctx, postgres, "story_checklist_items", "story_id", sourceID, 1)
	assertMutationActivityCount(t, ctx, postgres, sourceID, "update", 0)
}
