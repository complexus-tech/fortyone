package storiesrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

// Keep the item identity and content in the activity itself: checklist items and
// their converted stories can be deleted after this history is written.
type checklistActivitySnapshot struct {
	ID               uuid.UUID  `json:"id"`
	Title            string     `json:"title"`
	Completed        bool       `json:"completed"`
	Position         int        `json:"position"`
	ConvertedStoryID *uuid.UUID `json:"convertedStoryId,omitempty"`
	ConvertedAt      *time.Time `json:"convertedAt,omitempty"`
}

func checklistActivitySnapshots(items []storydomain.ChecklistItem) []checklistActivitySnapshot {
	snapshots := make([]checklistActivitySnapshot, 0, len(items))
	for _, item := range items {
		snapshot := checklistActivitySnapshot{
			ID: item.ID, Title: item.Title, Completed: item.Completed,
			Position: item.Position, ConvertedAt: item.ConvertedAt,
		}
		if item.ConvertedStory != nil {
			id := item.ConvertedStory.ID
			snapshot.ConvertedStoryID = &id
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

func insertChecklistActivities(
	ctx context.Context,
	queries checklistQueries,
	command storydomain.ChecklistCommand,
	before, after []storydomain.ChecklistItem,
) error {
	if command.Scope.ActivityUser == nil {
		return nil
	}
	activity := storydomain.MutationActivity{
		StoryID: command.StoryID, UserID: *command.Scope.ActivityUser,
		WorkspaceID: command.Scope.WorkspaceID, CreatedAt: command.Event.OccurredAt,
	}

	oldItems, newItems := checklistActivitySnapshots(before), checklistActivitySnapshots(after)
	if command.Action == storydomain.ChecklistReorder {
		return insertChecklistActivity(ctx, queries, activity, "checklist", "checklist_reordered", "", oldItems, newItems)
	}

	newItemsByID := make(map[uuid.UUID]checklistActivitySnapshot, len(newItems))
	for _, item := range newItems {
		newItemsByID[item.ID] = item
	}
	oldItemsByID := make(map[uuid.UUID]checklistActivitySnapshot, len(oldItems))
	for _, oldItem := range oldItems {
		oldItemsByID[oldItem.ID] = oldItem
		newItem, exists := newItemsByID[oldItem.ID]
		if !exists {
			if err := insertChecklistActivity(ctx, queries, activity, "checklist_item", "checklist_item_removed", oldItem.Title, oldItem, nil); err != nil {
				return err
			}
			continue
		}
		if oldItem.ConvertedAt == nil && newItem.ConvertedAt != nil {
			// Conversion also completes the item; one entry captures both changes.
			if err := insertChecklistActivity(ctx, queries, activity, "checklist_item", "checklist_item_converted", newItem.Title, oldItem, newItem); err != nil {
				return err
			}
			continue
		}
		if oldItem.Title != newItem.Title {
			if err := insertChecklistActivity(ctx, queries, activity, "checklist_item", "checklist_item_updated", newItem.Title, oldItem, newItem); err != nil {
				return err
			}
		}
		if oldItem.Completed != newItem.Completed {
			reason := "checklist_item_completed"
			if !newItem.Completed {
				reason = "checklist_item_reopened"
			}
			if err := insertChecklistActivity(ctx, queries, activity, "checklist_item", reason, newItem.Title, oldItem, newItem); err != nil {
				return err
			}
		}
	}
	for _, item := range newItems {
		if _, exists := oldItemsByID[item.ID]; exists {
			continue
		}
		if err := insertChecklistActivity(ctx, queries, activity, "checklist_item", "checklist_item_added", item.Title, nil, item); err != nil {
			return err
		}
	}
	return nil
}

func insertCopiedChecklistActivities(ctx context.Context, queries checklistQueries, command storydomain.DuplicateStoryCommand) error {
	items, err := listChecklist(ctx, queries, command.Scope.WorkspaceID, command.TargetStoryID)
	if err != nil {
		return err
	}
	for _, item := range checklistActivitySnapshots(items) {
		if err := insertChecklistActivity(ctx, queries, command.Activity, "checklist_item", "checklist_item_added", item.Title, nil, item); err != nil {
			return err
		}
	}
	return nil
}

func insertChecklistActivity(
	ctx context.Context,
	queries checklistQueries,
	activity storydomain.MutationActivity,
	field, reason, title string,
	before, after any,
) error {
	oldValue, err := json.Marshal(before)
	if err != nil {
		return fmt.Errorf("encode previous checklist activity value: %w", err)
	}
	newValue, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("encode current checklist activity value: %w", err)
	}
	activity.ID = uuid.New()
	activity.Type = "update"
	activity.Field = field
	activity.CurrentValue = title
	activity.Reason = &reason
	activity.OldValue = oldValue
	activity.NewValue = newValue
	// Every checklist change is retained, even within the usual compaction window.
	return upsertMutationActivity(ctx, queries, activity, false)
}
