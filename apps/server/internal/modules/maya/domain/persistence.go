package mayadomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrActionNotProposed        = errors.New("maya action is no longer proposed")
	ErrInvalidPlanInput         = errors.New("invalid maya plan input")
	ErrPersistenceNotConfigured = errors.New("maya persistence is not configured")
	ErrPlanNotFound             = errors.New("maya plan not found")
)

type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
)

type ActionStatus string

const (
	ActionStatusProposed ActionStatus = "proposed"
	ActionStatusApplied  ActionStatus = "applied"
	ActionStatusFailed   ActionStatus = "failed"
)

type ActionType string

const (
	ActionTypeAssignStory       ActionType = "assign_story"
	ActionTypeScheduleWorkBlock ActionType = "schedule_work_block"
	ActionTypeFlagScheduleRisk  ActionType = "flag_schedule_risk"
)

type RunTrigger string

const (
	RunTriggerManual RunTrigger = "manual"
	RunTriggerEvent  RunTrigger = "event"
)

type CoreRun struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspaceId"`
	StoryID     uuid.UUID       `json:"storyId"`
	TriggeredBy uuid.UUID       `json:"triggeredBy"`
	Trigger     RunTrigger      `json:"trigger"`
	Status      RunStatus       `json:"status"`
	Summary     string          `json:"summary"`
	Context     json.RawMessage `json:"context,omitempty"`
	Error       *string         `json:"error,omitempty"`
	StartedAt   time.Time       `json:"startedAt"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type CoreAction struct {
	ID          uuid.UUID       `json:"id"`
	RunID       uuid.UUID       `json:"runId"`
	WorkspaceID uuid.UUID       `json:"workspaceId"`
	StoryID     uuid.UUID       `json:"storyId"`
	Type        ActionType      `json:"type"`
	Status      ActionStatus    `json:"status"`
	Reason      string          `json:"reason"`
	Payload     ActionPayload   `json:"payload"`
	PayloadJSON json.RawMessage `json:"-"`
	Error       *string         `json:"error,omitempty"`
	AppliedAt   *time.Time      `json:"appliedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type ActionPayload struct {
	AssignStory   *AssignStoryPayload   `json:"assignStory,omitempty"`
	ScheduleBlock *ScheduleBlockPayload `json:"scheduleBlock,omitempty"`
	Risk          *RiskPayload          `json:"risk,omitempty"`
}

type AssignStoryPayload struct {
	AssigneeID        uuid.UUID `json:"assigneeId"`
	ExpectedUpdatedAt time.Time `json:"expectedUpdatedAt"`
}

type ScheduleBlockPayload struct {
	UserID                 uuid.UUID   `json:"userId"`
	SegmentIndex           int         `json:"segmentIndex"`
	Operation              string      `json:"operation,omitempty"`
	Title                  string      `json:"title"`
	StartAt                time.Time   `json:"startAt"`
	EndAt                  time.Time   `json:"endAt"`
	PlannedStartAt         time.Time   `json:"plannedStartAt"`
	PlannedEndAt           time.Time   `json:"plannedEndAt"`
	ExpectedStoryUpdatedAt time.Time   `json:"expectedStoryUpdatedAt"`
	PreemptBlockIDs        []uuid.UUID `json:"preemptBlockIds,omitempty"`
}

const (
	ScheduleBlockOperationUpsert = "upsert"
	ScheduleBlockOperationDelete = "delete"
	ScheduleBlockOperationRetain = "retain"
)

type RiskPayload struct {
	Code             string `json:"code"`
	Message          string `json:"message"`
	RequiredMinutes  int    `json:"requiredMinutes,omitempty"`
	ScheduledMinutes int    `json:"scheduledMinutes,omitempty"`
	RemainingMinutes int    `json:"remainingMinutes,omitempty"`
}

type WorkPlan struct {
	Run     CoreRun      `json:"run"`
	Actions []CoreAction `json:"actions"`
}

type ScheduleStoryRef struct {
	WorkspaceID uuid.UUID `db:"workspace_id"`
	StoryID     uuid.UUID `db:"story_id"`
}

type ScheduleRecoveryRef struct {
	ScheduleStoryRef
	InterruptedRunID *uuid.UUID `db:"interrupted_run_id"`
}

type CreateRunInput struct {
	WorkspaceID uuid.UUID
	StoryID     uuid.UUID
	TriggeredBy uuid.UUID
	Trigger     RunTrigger
	Context     json.RawMessage
}

// AssignmentCandidateStory is the persistence-neutral projection consumed by
// Maya's background assignment and schedule-recovery workers.
type AssignmentCandidateStory struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
	AssigneeID  uuid.UUID
}

// WorkFocusMember identifies one team membership whose inferred work focus can
// be derived without replacing manually supplied role context.
type WorkFocusMember struct {
	WorkspaceID           uuid.UUID
	TeamID                uuid.UUID
	UserID                uuid.UUID
	ManualRoleTitle       string
	ManualRoleDescription string
}

type WorkFocusEvidence struct {
	Title       string
	Description string
	Labels      []string
}

type WorkFocusInferenceResult struct {
	ShouldInfer     bool
	RoleTitle       string
	RoleDescription string
	StoryCount      int
	Confidence      float64
}
