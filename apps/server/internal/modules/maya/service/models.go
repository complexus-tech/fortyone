package maya

import (
	"context"
	"encoding/json"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
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
	AssigneeID uuid.UUID `json:"assigneeId"`
}

type ScheduleBlockPayload struct {
	UserID  uuid.UUID `json:"userId"`
	Title   string    `json:"title"`
	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`
}

type RiskPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PlanInput struct {
	Context          context.Context
	WorkspaceID      uuid.UUID
	Story            stories.CoreSingleStory
	DurationMinutes  int
	WindowStart      time.Time
	WindowEnd        time.Time
	Candidates       []CandidateSchedule
	AssignmentReason string
}

type CandidateSchedule struct {
	Member      reports.CoreMemberWorkload
	BusyWindows []calendar.CoreBusyWindow
	Blocks      []calendar.CoreScheduleBlock
}

type PlanResult struct {
	Summary        string
	SelectedUserID *uuid.UUID
	Actions        []CoreAction
}

type CandidateRecommendationInput struct {
	WorkspaceID     uuid.UUID
	Story           stories.CoreSingleStory
	DurationMinutes int
	WindowStart     time.Time
	WindowEnd       time.Time
	Candidates      []CandidateRecommendation
}

type CandidateRecommendation struct {
	UserID                uuid.UUID
	FullName              string
	Username              string
	TeamAIRoleTitle       string
	TeamAIRoleDescription string
	OpenStories           int
	EstimateTotal         int
	HasAvailableSlot      bool
	SlotStart             time.Time
	SlotEnd               time.Time
	LastStoryActivityAt   *time.Time
	DaysSinceLastActivity *int
	RecentlyActive        bool
}

type CandidateRecommendationResult struct {
	UserID uuid.UUID
	Reason string
}

type CandidateAdvisor interface {
	RecommendCandidate(ctx context.Context, input CandidateRecommendationInput) (CandidateRecommendationResult, error)
}

type BatchAssignmentStory struct {
	ID              uuid.UUID
	Title           string
	Description     string
	Priority        string
	EstimateValue   *int16
	EstimateLabel   *string
	DurationMinutes int
}

type BatchAssignmentRecommendationInput struct {
	WorkspaceID uuid.UUID
	Stories     []BatchAssignmentStory
	Candidates  []CandidateRecommendation
}

type BatchAssignmentRecommendation struct {
	StoryID    uuid.UUID
	AssigneeID uuid.UUID
	Reason     string
}

type BatchAssignmentRecommendationResult struct {
	Assignments []BatchAssignmentRecommendation
}

type BatchAssignmentAdvisor interface {
	RecommendAssignments(ctx context.Context, input BatchAssignmentRecommendationInput) (BatchAssignmentRecommendationResult, error)
}

type Repository interface {
	CreateRun(ctx context.Context, input CreateRunInput) (CoreRun, error)
	CompleteRun(ctx context.Context, runID uuid.UUID, status RunStatus, summary string, message *string) (CoreRun, error)
	CreateActions(ctx context.Context, actions []CoreAction) ([]CoreAction, error)
	MarkActionApplied(ctx context.Context, actionID uuid.UUID) error
	MarkActionFailed(ctx context.Context, actionID uuid.UUID, message string) error
}

type CreateRunInput struct {
	WorkspaceID uuid.UUID
	StoryID     uuid.UUID
	TriggeredBy uuid.UUID
	Trigger     RunTrigger
	Context     json.RawMessage
}
