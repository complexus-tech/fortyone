package maya

import (
	"context"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

// These aliases keep Maya's use-case files expressed in the language of the
// consuming module while depending only on provider-owned domain contracts.
type (
	Story                  = storydomain.Story
	MemberWorkload         = reportdomain.CoreMemberWorkload
	WorkloadReportFilters  = reportdomain.ReportFilters
	ScheduleBlock          = calendar.CoreScheduleBlock
	ScheduleSegmentInput   = calendar.MayaScheduleSegmentInput
	ScheduleReconcileInput = calendar.MayaScheduleReconcileInput
	RunStatus              = mayadomain.RunStatus
	ActionStatus           = mayadomain.ActionStatus
	ActionType             = mayadomain.ActionType
	RunTrigger             = mayadomain.RunTrigger
	CoreRun                = mayadomain.CoreRun
	CoreAction             = mayadomain.CoreAction
	ActionPayload          = mayadomain.ActionPayload
	AssignStoryPayload     = mayadomain.AssignStoryPayload
	ScheduleBlockPayload   = mayadomain.ScheduleBlockPayload
	RiskPayload            = mayadomain.RiskPayload
	ScheduleStoryRef       = mayadomain.ScheduleStoryRef
	ScheduleRecoveryRef    = mayadomain.ScheduleRecoveryRef
	CreateRunInput         = mayadomain.CreateRunInput
)

const (
	MaximumEstimatedDurationMinutes = storydomain.MaximumEstimatedDurationMinutes
	ScheduleBlockSourceMaya         = calendar.ScheduleBlockSourceMaya
	AutoSchedulingStatusAtRisk      = storydomain.AutoSchedulingStatusAtRisk
	AutoSchedulingStatusCannotFit   = storydomain.AutoSchedulingStatusCannotFit
	AutoSchedulingStatusLocked      = storydomain.AutoSchedulingStatusLocked
	AutoSchedulingStatusNeedsOwner  = storydomain.AutoSchedulingStatusNeedsOwner
	AutoSchedulingStatusNeedsTime   = storydomain.AutoSchedulingStatusNeedsTime
	AutoSchedulingStatusOff         = storydomain.AutoSchedulingStatusOff
	AutoSchedulingStatusPlanning    = storydomain.AutoSchedulingStatusPlanning
	AutoSchedulingStatusScheduled   = storydomain.AutoSchedulingStatusScheduled
	RunStatusRunning                = mayadomain.RunStatusRunning
	RunStatusSucceeded              = mayadomain.RunStatusSucceeded
	RunStatusFailed                 = mayadomain.RunStatusFailed
	ActionStatusProposed            = mayadomain.ActionStatusProposed
	ActionStatusApplied             = mayadomain.ActionStatusApplied
	ActionStatusFailed              = mayadomain.ActionStatusFailed
	ActionTypeAssignStory           = mayadomain.ActionTypeAssignStory
	ActionTypeScheduleWorkBlock     = mayadomain.ActionTypeScheduleWorkBlock
	ActionTypeFlagScheduleRisk      = mayadomain.ActionTypeFlagScheduleRisk
	RunTriggerManual                = mayadomain.RunTriggerManual
	RunTriggerEvent                 = mayadomain.RunTriggerEvent
	ScheduleBlockOperationUpsert    = mayadomain.ScheduleBlockOperationUpsert
	ScheduleBlockOperationDelete    = mayadomain.ScheduleBlockOperationDelete
	ScheduleBlockOperationRetain    = mayadomain.ScheduleBlockOperationRetain
)

var (
	ErrStoryNotFound             = storydomain.ErrNotFound
	ErrAutoSchedulingOwnerLocked = storydomain.ErrAutoSchedulingOwnerLocked
	ErrStoryChanged              = storydomain.ErrStoryChanged
)

type PlanInput struct {
	Context                  context.Context
	AsOf                     time.Time
	WorkspaceID              uuid.UUID
	Story                    storydomain.Story
	DurationMinutes          int
	MinimumFocusBlockMinutes int
	WindowStart              time.Time
	WindowEnd                time.Time
	WorkingDays              []int
	Candidates               []CandidateSchedule
	AssignmentReason         string
}

type CandidateSchedule struct {
	Member               reportdomain.CoreMemberWorkload
	Timezone             string
	WorkingDays          []int
	WorkingStartMinute   int
	WorkingEndMinute     int
	BusyWindows          []calendar.CoreBusyWindow
	Blocks               []calendar.CoreScheduleBlock
	PreemptibleBlockIDs  []uuid.UUID
	PreferredStartMinute *int
}

type PlanResult struct {
	Summary           string
	SelectedUserID    *uuid.UUID
	Actions           []CoreAction
	Timezone          string
	PreemptedBlockIDs []uuid.UUID
	DurationMinutes   int
	ScheduledMinutes  int
	RemainingMinutes  int
}

type CandidateRecommendationInput struct {
	WorkspaceID     uuid.UUID
	Story           storydomain.Story
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
	GetWorkPlan(ctx context.Context, runID, workspaceID, triggeredBy uuid.UUID) (WorkPlan, error)
	MarkActionApplied(ctx context.Context, actionID uuid.UUID) error
	MarkActionFailed(ctx context.Context, actionID uuid.UUID, message string) error
}

type ScheduleRepository interface {
	ListScheduleStoryRefsForUser(ctx context.Context, userID uuid.UUID) ([]ScheduleStoryRef, error)
	ClaimScheduleRecoveryStoryRefs(ctx context.Context, limit int, retryBefore, interruptedRunBefore time.Time) ([]ScheduleRecoveryRef, error)
	CompleteInterruptedScheduleRun(ctx context.Context, runID uuid.UUID, message string) error
	ListMayaScheduleOwners(ctx context.Context, workspaceID, storyID uuid.UUID) ([]uuid.UUID, error)
	WorkspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error)
	StoryIsActiveForAutoScheduling(ctx context.Context, workspaceID, storyID uuid.UUID) (bool, error)
	StoryIsSchedulableForUser(ctx context.Context, workspaceID, storyID, userID uuid.UUID) (bool, error)
	StoryScheduleOwnershipIsRetainable(ctx context.Context, workspaceID, storyID, userID uuid.UUID) (bool, error)
	WithScheduleStoryLock(ctx context.Context, workspaceID, storyID uuid.UUID, reconcile func() error) error
}
