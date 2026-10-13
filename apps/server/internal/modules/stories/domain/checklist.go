package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const MaxChecklistItems = 200
const MaxChecklistTitleLength = 255

type ChecklistStory struct {
	ID         uuid.UUID  `json:"id"`
	SequenceID int        `json:"sequenceId"`
	Title      string     `json:"title"`
	TeamID     uuid.UUID  `json:"teamId"`
	StatusID   *uuid.UUID `json:"statusId"`
}

// Checklist items are lightweight work within a story, never child stories.
// ConvertedAt preserves conversion history even after the linked story is deleted.
type ChecklistItem struct {
	ID             uuid.UUID       `json:"id"`
	StoryID        uuid.UUID       `json:"storyId"`
	Title          string          `json:"title"`
	Completed      bool            `json:"completed"`
	Position       int             `json:"position"`
	ConvertedStory *ChecklistStory `json:"convertedStory"`
	ConvertedAt    *time.Time      `json:"convertedAt"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type ChecklistAction string

const (
	ChecklistAdd     ChecklistAction = "add"
	ChecklistUpdate  ChecklistAction = "update"
	ChecklistDelete  ChecklistAction = "delete"
	ChecklistClear   ChecklistAction = "clear"
	ChecklistReorder ChecklistAction = "reorder"
	ChecklistConvert ChecklistAction = "convert"
)

type ChecklistCommand struct {
	Scope     MutationScope
	StoryID   uuid.UUID
	ItemID    uuid.UUID
	Action    ChecklistAction
	Title     *string
	Completed *bool
	ItemIDs   []uuid.UUID
	Event     MutationEvent
}

func (c ChecklistCommand) Validate() error {
	if err := c.Scope.Validate(); err != nil {
		return err
	}
	if c.StoryID == uuid.Nil {
		return fmt.Errorf("%w: story is required", ErrInvalidMutation)
	}
	if err := c.Event.Validate(); err != nil {
		return err
	}
	if c.Event.StoryID != c.StoryID || c.Event.WorkspaceID != c.Scope.WorkspaceID || c.Event.Type != MutationEventStoryUpdated {
		return ErrInvalidMutation
	}
	if c.Title != nil && (strings.TrimSpace(*c.Title) == "" || utf8.RuneCountInString(*c.Title) > MaxChecklistTitleLength) {
		return fmt.Errorf("%w: checklist titles must contain 1 to %d characters", ErrInvalidMutation, MaxChecklistTitleLength)
	}
	switch c.Action {
	case ChecklistAdd:
		if c.Title == nil || c.ItemID == uuid.Nil {
			return ErrInvalidMutation
		}
	case ChecklistUpdate:
		if c.ItemID == uuid.Nil || (c.Title == nil && c.Completed == nil) {
			return ErrInvalidMutation
		}
	case ChecklistDelete, ChecklistConvert:
		if c.ItemID == uuid.Nil {
			return ErrInvalidMutation
		}
	case ChecklistClear:
	case ChecklistReorder:
		if len(c.ItemIDs) > MaxChecklistItems {
			return ErrInvalidMutation
		}
		seen := make(map[uuid.UUID]struct{}, len(c.ItemIDs))
		for _, id := range c.ItemIDs {
			if id == uuid.Nil {
				return ErrInvalidMutation
			}
			if _, exists := seen[id]; exists {
				return fmt.Errorf("%w: checklist order contains duplicate items", ErrInvalidMutation)
			}
			seen[id] = struct{}{}
		}
	default:
		return ErrInvalidMutation
	}
	return nil
}

// ValidateChecklistOrder rejects stale or partial lists rather than silently
// discarding concurrent additions or accepting items from another story.
func ValidateChecklistOrder(items []ChecklistItem, ids []uuid.UUID) error {
	if len(items) != len(ids) {
		return ErrMutationConflict
	}
	current := make(map[uuid.UUID]bool, len(items))
	for _, item := range items {
		current[item.ID] = true
	}
	for _, id := range ids {
		if !current[id] {
			return ErrMutationConflict
		}
		delete(current, id)
	}
	return nil
}

type ChecklistMutationResult struct {
	Items        []ChecklistItem
	Changed      bool
	CreatedStory *ChecklistStory
}
