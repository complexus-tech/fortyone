package messagingdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidConfirmation   = errors.New("invalid story mutation confirmation")
	ErrExpiredConfirmation   = errors.New("story mutation confirmation expired")
	ErrCancelledConfirmation = errors.New("story mutation confirmation was cancelled")
	ErrAppliedConfirmation   = errors.New("story mutation confirmation was already applied")
)

type StoryMutationOperation string

const (
	StoryMutationCreate      StoryMutationOperation = "create_story"
	StoryMutationCreateBatch StoryMutationOperation = "create_stories"
	StoryMutationUpdate      StoryMutationOperation = "update_story"
	StoryMutationComment     StoryMutationOperation = "add_story_comment"
	StoryMutationRelation    StoryMutationOperation = "add_story_relationship"
)

type StoryMutationResult struct {
	Status                   string                    `json:"status"`
	Operation                StoryMutationOperation    `json:"operation"`
	StoryID                  uuid.UUID                 `json:"story_id"`
	Reference                string                    `json:"reference"`
	TeamID                   uuid.UUID                 `json:"team_id"`
	Title                    string                    `json:"title"`
	Priority                 string                    `json:"priority"`
	AssigneeID               *uuid.UUID                `json:"assignee_id,omitempty"`
	EstimatedDurationMinutes *int                      `json:"estimated_duration_minutes,omitempty"`
	MinimumFocusBlockMinutes *int                      `json:"minimum_focus_block_minutes,omitempty"`
	AutoSchedulingEnabled    bool                      `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool                      `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string                    `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string                   `json:"auto_scheduling_reason,omitempty"`
	AutoSchedulingUpdatedAt  *time.Time                `json:"auto_scheduling_updated_at,omitempty"`
	CommentID                *uuid.UUID                `json:"comment_id,omitempty"`
	AssociationID            *uuid.UUID                `json:"association_id,omitempty"`
	Items                    []StoryMutationItemResult `json:"items,omitempty"`
}

type StoryMutationItemResult struct {
	Index                    int        `json:"index"`
	Status                   string     `json:"status"`
	StoryID                  uuid.UUID  `json:"story_id"`
	Reference                string     `json:"reference"`
	TeamID                   uuid.UUID  `json:"team_id"`
	Title                    string     `json:"title"`
	Priority                 string     `json:"priority"`
	AssigneeID               *uuid.UUID `json:"assignee_id,omitempty"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes,omitempty"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes,omitempty"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool       `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string     `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string    `json:"auto_scheduling_reason,omitempty"`
	AutoSchedulingUpdatedAt  *time.Time `json:"auto_scheduling_updated_at,omitempty"`
}

type StoryMutationCancellationResult struct {
	Status string `json:"status"`
}

type StoryMutationConfirmationStatus string

const (
	StoryMutationConfirmationPending   StoryMutationConfirmationStatus = "pending"
	StoryMutationConfirmationApplied   StoryMutationConfirmationStatus = "applied"
	StoryMutationConfirmationCancelled StoryMutationConfirmationStatus = "cancelled"
	StoryMutationConfirmationExpired   StoryMutationConfirmationStatus = "expired"
)

type StoryMutationConfirmationStateInput struct {
	ConfirmationID uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	TeamID         uuid.UUID
	Operation      StoryMutationOperation
	TokenHash      []byte
	Proposal       json.RawMessage
	ExpiresAt      time.Time
}

type StoryMutationConfirmationRecord struct {
	TeamID    uuid.UUID
	Operation StoryMutationOperation
	Status    StoryMutationConfirmationStatus
	Proposal  json.RawMessage
	Result    *StoryMutationResult
	LastError string
}

type StoryMutationConfirmationBinding struct {
	ConfirmationID uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	TokenHash      []byte
}
