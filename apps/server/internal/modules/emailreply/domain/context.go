package emailreplydomain

import (
	"time"

	"github.com/google/uuid"
)

// TargetKind identifies an entity that can be exposed to the email agent.
// These values are persistence contracts, not model-provided identifiers.
type TargetKind string

const (
	TargetObjective TargetKind = "objective"
	TargetKeyResult TargetKind = "key_result"
	TargetStory     TargetKind = "story"
	TargetFeedback  TargetKind = "feedback"
)

// ActionKind identifies the database state needed to authorize or reconcile a
// proposed email mutation.
type ActionKind string

const (
	ActionObjectiveUpdate ActionKind = "update_objective"
	ActionKeyResultUpdate ActionKind = "update_key_result"
	ActionStoryUpdate     ActionKind = "update_story"
	ActionFeedbackStatus  ActionKind = "update_feedback_status"
)

type ActorScope struct {
	WorkspaceSlug string
	Role          string
	TeamIDs       []uuid.UUID
}

// TargetSnapshot is the repository's typed projection of the four target
// tables. Only fields relevant to the selected Kind are populated.
type TargetSnapshot struct {
	Kind            TargetKind
	ID              uuid.UUID
	TeamID          uuid.UUID
	Name            string
	Health          string
	MeasurementType string
	CurrentValue    float64
	TargetValue     float64
	StartDate       *time.Time
	EndDate         *time.Time
	StatusName      string
	AssigneeName    string
	Status          string
	UpdatedAt       time.Time
}

type ChoiceKind string

const (
	ChoiceStoryStatus   ChoiceKind = "story_status"
	ChoiceStoryAssignee ChoiceKind = "story_assignee"
)

type Choice struct {
	Kind   ChoiceKind
	ID     uuid.UUID
	TeamID uuid.UUID
	Name   string
}

// ProposalState is a typed union of the persisted fields used for idempotent
// post-commit reconciliation. Only fields relevant to Kind are populated.
type ProposalState struct {
	Kind            ActionKind
	ObjectiveHealth string
	KeyResultValue  float64
	FeedbackStatus  string
	StoryStatusID   *uuid.UUID
	StoryAssigneeID *uuid.UUID
	StoryEndDate    *time.Time
}
