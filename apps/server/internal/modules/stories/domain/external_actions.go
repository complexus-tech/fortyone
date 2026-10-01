package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExternalStoryDelivery controls whether a story created through a trusted
// external application port may leave FortyOne through the mutation-event
// publisher. It is deliberately transport-invisible: trusted adapters set it
// after decoding instead of accepting it from public callers.
type ExternalStoryDelivery uint8

const (
	ExternalStoryDeliveryDefault ExternalStoryDelivery = iota
	ExternalStoryDeliveryInternalOnly
)

// NewStory is the transport-neutral input accepted by legacy story-creation
// application ports. New mutation code should prefer CreateStoryCommand, which
// also carries the actor scope and durable event metadata.
type NewStory struct {
	Title                    string                `json:"title"`
	EstimateValue            *int16                `json:"estimateValue"`
	EstimatedDurationMinutes *int                  `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes *int                  `json:"minimumFocusBlockMinutes"`
	AutoSchedulingEnabled    bool                  `json:"autoSchedulingEnabled"`
	AutoSchedulingLocked     bool                  `json:"autoSchedulingLocked"`
	Description              *string               `json:"description"`
	DescriptionHTML          *string               `json:"descriptionHTML"`
	Parent                   *uuid.UUID            `json:"parentId"`
	Objective                *uuid.UUID            `json:"objectiveId"`
	Status                   *uuid.UUID            `json:"statusId"`
	Assignee                 *uuid.UUID            `json:"assigneeId"`
	BlockedBy                *uuid.UUID            `json:"blockedById"`
	Blocking                 *uuid.UUID            `json:"blockingId"`
	Related                  *uuid.UUID            `json:"relatedId"`
	Reporter                 *uuid.UUID            `json:"reporterId"`
	Priority                 string                `json:"priority"`
	Sprint                   *uuid.UUID            `json:"sprintId"`
	KeyResult                *uuid.UUID            `json:"keyResultId"`
	LabelIDs                 []uuid.UUID           `json:"labelIds"`
	CustomFieldValues        []CustomFieldValue    `json:"customFieldValues"`
	StartDate                *time.Time            `json:"startDate"`
	EndDate                  *time.Time            `json:"endDate"`
	Team                     uuid.UUID             `json:"teamId"`
	CreationKey              *string               `json:"-"`
	ExternalDelivery         ExternalStoryDelivery `json:"-"`
}

// CustomFieldValue is an exact, transport-neutral value supplied at creation.
// Its definition, type and live team scope are checked by the transaction
// participant before the story and its outbox event can commit.
type CustomFieldValue struct {
	FieldID uuid.UUID `json:"fieldId"`
	Value   *string   `json:"value" validate:"omitempty,max=4000"`
}

// NewComment is the transport-neutral input used by the story service's
// compatibility comment port. The comments module remains responsible for the
// authoritative mutation and authorization policy.
type NewComment struct {
	CreationID *uuid.UUID
	StoryID    uuid.UUID
	Parent     *uuid.UUID
	UserID     uuid.UUID
	Comment    string
	Mentions   []uuid.UUID
}
