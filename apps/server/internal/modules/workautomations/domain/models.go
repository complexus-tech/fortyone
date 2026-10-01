package workautomationsdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid automation")
	ErrNotFound     = errors.New("automation not found")
)

type Automation struct {
	ID            uuid.UUID       `json:"id"`
	TeamID        uuid.UUID       `json:"teamId"`
	OwnerID       uuid.UUID       `json:"ownerId"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Configuration json.RawMessage `json:"configuration"`
	Paused        bool            `json:"paused"`
	NextRunAt     *time.Time      `json:"nextRunAt"`
	LastRunAt     *time.Time      `json:"lastRunAt"`
	LastError     string          `json:"lastError"`
	CreatedAt     time.Time       `json:"createdAt"`
	CanEdit       bool            `json:"canEdit"`
}

type Input struct {
	TeamID        uuid.UUID       `json:"teamId" validate:"required"`
	Kind          string          `json:"kind" validate:"required,oneof=rule recurrence"`
	Name          string          `json:"name" validate:"required,max=100"`
	Configuration json.RawMessage `json:"configuration" validate:"required"`
}

type Update struct {
	Paused *bool `json:"paused" validate:"required"`
}

type RuleConfiguration struct {
	Version    int        `json:"version"`
	Trigger    string     `json:"trigger"`
	Conditions Conditions `json:"conditions"`
	Actions    Actions    `json:"actions"`
}

type Conditions struct {
	StatusIDs   []uuid.UUID `json:"statusIds,omitempty"`
	Priorities  []string    `json:"priorities,omitempty"`
	AssigneeIDs []uuid.UUID `json:"assigneeIds,omitempty"`
	Unassigned  bool        `json:"unassigned,omitempty"`
}

type Actions struct {
	StatusID      *uuid.UUID `json:"statusId,omitempty"`
	Priority      *string    `json:"priority,omitempty"`
	AssigneeID    *uuid.UUID `json:"assigneeId,omitempty"`
	ClearAssignee bool       `json:"clearAssignee,omitempty"`
}

type Schedule struct {
	Frequency string `json:"frequency"`
	Timezone  string `json:"timezone"`
	StartsOn  string `json:"startsOn"`
	LocalTime string `json:"localTime"`
	Weekday   int    `json:"weekday"`
	MonthDay  int    `json:"monthDay"`
}

type RecurrenceConfiguration struct {
	Version  int      `json:"version"`
	Schedule Schedule `json:"schedule"`
	Draft    Draft    `json:"draft"`
}

// Draft is a snapshot of supported task metadata. It has no task identity,
// deadlines, attachments or arbitrary fields to copy into future occurrences.
type Draft struct {
	Title                    string             `json:"title"`
	Description              string             `json:"description"`
	DescriptionHTML          string             `json:"descriptionHTML"`
	Priority                 string             `json:"priority"`
	StatusID                 *uuid.UUID         `json:"statusId,omitempty"`
	AssigneeID               *uuid.UUID         `json:"assigneeId,omitempty"`
	LabelIDs                 []uuid.UUID        `json:"labelIds,omitempty"`
	EstimateValue            *int16             `json:"estimateValue,omitempty"`
	EstimatedDurationMinutes *int               `json:"estimatedDurationMinutes,omitempty"`
	MinimumFocusBlockMinutes *int               `json:"minimumFocusBlockMinutes,omitempty"`
	Checklist                []string           `json:"checklist,omitempty"`
	CustomFieldValues        []CustomFieldValue `json:"customFieldValues,omitempty"`
}

type CustomFieldValue struct {
	FieldID uuid.UUID `json:"fieldId"`
	Value   *string   `json:"value"`
}

type Claimed struct {
	Automation
	WorkspaceID uuid.UUID
	LeaseToken  uuid.UUID
	DueAt       *time.Time
	EventAt     time.Time
	EventID     uuid.UUID
}

type Event struct {
	ID        uuid.UUID
	StoryID   uuid.UUID
	Kind      string
	CreatedAt time.Time
}

type Run struct {
	ID           uuid.UUID  `json:"id"`
	AutomationID uuid.UUID  `json:"automationId"`
	Occurrence   string     `json:"occurrence"`
	Status       string     `json:"status"`
	StoryID      *uuid.UUID `json:"storyId"`
	Error        string     `json:"error"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
}

type Story struct {
	ID         uuid.UUID
	TeamID     uuid.UUID
	StatusID   *uuid.UUID
	Priority   string
	AssigneeID *uuid.UUID
	UpdatedAt  time.Time
}
