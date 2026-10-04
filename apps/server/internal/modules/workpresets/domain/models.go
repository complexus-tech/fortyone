package workpresetsdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("work preset not found")
	ErrInvalidInput = errors.New("invalid work preset")
)

type Kind string

const (
	View     Kind = "view"
	Template Kind = "template"
)

type Visibility string

const (
	Personal Visibility = "personal"
	Team     Visibility = "team"
)

type Preset struct {
	ID            uuid.UUID       `json:"id"`
	TeamID        uuid.UUID       `json:"teamId"`
	OwnerID       uuid.UUID       `json:"ownerId"`
	Kind          Kind            `json:"kind"`
	Visibility    Visibility      `json:"visibility"`
	Name          string          `json:"name"`
	Configuration json.RawMessage `json:"configuration"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	CanEdit       bool            `json:"canEdit"`
}

type Input struct {
	TeamID        uuid.UUID       `json:"teamId" validate:"required"`
	Kind          Kind            `json:"kind" validate:"required,oneof=view template"`
	Visibility    Visibility      `json:"visibility" validate:"required,oneof=personal team"`
	Name          string          `json:"name" validate:"required,max=100"`
	Configuration json.RawMessage `json:"configuration" validate:"required"`
}

type Update struct {
	Name          string          `json:"name" validate:"required,max=100"`
	Configuration json.RawMessage `json:"configuration,omitempty"`
}

type List struct {
	TeamID uuid.UUID
	Kind   Kind
	Limit  int
	Before *Position
}

type Position struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
}

type ViewConfiguration struct {
	Version     int          `json:"version"`
	Description *string      `json:"description,omitempty"`
	Icon        *string      `json:"icon,omitempty"`
	Scope       *MyWorkScope `json:"scope,omitempty"`
	Layout      string       `json:"layout"`
	Filters     Filters      `json:"filters"`
	ViewOptions ViewOptions  `json:"viewOptions"`
}

// My Work is evaluated for the reader, independently of the preset's owning team.
type MyWorkScope struct {
	Kind          string  `json:"kind"`
	Tab           string  `json:"tab"`
	Category      *string `json:"category,omitempty"`
	Overdue       *bool   `json:"overdue,omitempty"`
	CreatedAfter  *string `json:"createdAfter,omitempty"`
	CreatedBefore *string `json:"createdBefore,omitempty"`
}

type Filters struct {
	StatusIDs       []uuid.UUID       `json:"statusIds"`
	AssigneeIDs     []uuid.UUID       `json:"assigneeIds"`
	ReporterIDs     []uuid.UUID       `json:"reporterIds"`
	Priorities      []string          `json:"priorities"`
	TeamIDs         []uuid.UUID       `json:"teamIds"`
	SprintIDs       []uuid.UUID       `json:"sprintIds"`
	LabelIDs        []uuid.UUID       `json:"labelIds"`
	EstimateValues  []float64         `json:"estimateValues"`
	ParentID        *uuid.UUID        `json:"parentId"`
	ObjectiveID     *uuid.UUID        `json:"objectiveId"`
	EpicID          *uuid.UUID        `json:"epicId"`
	KeyResultID     *uuid.UUID        `json:"keyResultId"`
	ContentContains *string           `json:"contentContains,omitempty"`
	StartDate       *string           `json:"startDate,omitempty"`
	EndDate         *string           `json:"endDate,omitempty"`
	HasNoAssignee   *bool             `json:"hasNoAssignee"`
	HasBlockedBy    *bool             `json:"hasBlockedBy"`
	AssignedToMe    bool              `json:"assignedToMe"`
	CreatedByMe     bool              `json:"createdByMe"`
	CompletedAfter  *string           `json:"completedAfter,omitempty"`
	CompletedBefore *string           `json:"completedBefore,omitempty"`
	IsCompleted     *bool             `json:"isCompleted,omitempty"`
	IsNotCompleted  *bool             `json:"isNotCompleted,omitempty"`
	Operators       map[string]string `json:"operators,omitempty"`
}

type ViewOptions struct {
	DisplayColumnsVersion  int                 `json:"displayColumnsVersion,omitempty"`
	GroupBy                string              `json:"groupBy"`
	OrderBy                string              `json:"orderBy"`
	OrderDirection         string              `json:"orderDirection"`
	ShowEmptyGroups        bool                `json:"showEmptyGroups"`
	ShowSubStories         bool                `json:"showSubStories"`
	DisplayColumns         []string            `json:"displayColumns"`
	HiddenKanbanGroups     map[string][]string `json:"hiddenKanbanGroups,omitempty"`
	SelectedCustomFieldIDs []uuid.UUID         `json:"selectedCustomFieldIds,omitempty"`
}

type CustomFieldValue struct {
	FieldID uuid.UUID `json:"fieldId"`
	Value   *string   `json:"value"`
}

type TemplateConfiguration struct {
	Version                  int                `json:"version"`
	Title                    string             `json:"title"`
	Description              string             `json:"description"`
	DescriptionHTML          string             `json:"descriptionHTML"`
	Priority                 string             `json:"priority"`
	StatusID                 *uuid.UUID         `json:"statusId,omitempty"`
	AssigneeID               *uuid.UUID         `json:"assigneeId,omitempty"`
	LabelIDs                 []uuid.UUID        `json:"labelIds,omitempty"`
	EstimateLabel            *string            `json:"estimateLabel,omitempty"`
	EstimateValue            *float64           `json:"estimateValue,omitempty"`
	EstimatedDurationMinutes *int               `json:"estimatedDurationMinutes,omitempty"`
	MinimumFocusBlockMinutes *int               `json:"minimumFocusBlockMinutes,omitempty"`
	Checklist                []string           `json:"checklist"`
	CustomFieldValues        []CustomFieldValue `json:"customFieldValues,omitempty"`
}
