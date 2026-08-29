package maya

import (
	"context"
	"errors"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

var (
	ErrNotConfigured    = errors.New("maya agent is not configured")
	ErrPlanNotFound     = mayadomain.ErrPlanNotFound
	ErrMayaAccessDenied = mayadomain.ErrMayaAccessDenied
)

const defaultCandidateLimit = 15

type (
	User           = usersdomain.User
	UserListFilter = usersdomain.ListUsersFilter
)

type StoriesService interface {
	Get(ctx context.Context, storyID, workspaceID uuid.UUID) (storydomain.Story, error)
	UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error
	UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error
	UpdateExternalWithReasonIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error
	UpdateAutomationIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error
	UpdateAutomationStateIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, status string, reason *string, locked *bool, schedule *events.StoryScheduleTransition) error
}

type ReportsService interface {
	GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreWorkloadAnalysis, error)
}

type CalendarService interface {
	ListSchedule(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error)
	CreateScheduleBlock(ctx context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error)
}

type ScheduleCalendarService interface {
	ListSchedulingAvailability(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error)
	ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]calendar.CoreScheduleBlock, error)
	MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	ReconcileMayaScheduleBlocks(ctx context.Context, input calendar.MayaScheduleReconcileInput) (calendar.CoreScheduleReconcileResult, error)
	DispatchScheduleEventOutbox(ctx context.Context, userID uuid.UUID) error
}

type ScheduleFeedbackService interface {
	ListManualSchedulePreference(ctx context.Context, workspaceID, userID uuid.UUID) (calendar.CoreSchedulePreference, error)
}

type UsersService interface {
	List(ctx context.Context, workspaceID uuid.UUID, filter usersdomain.ListUsersFilter) ([]usersdomain.User, error)
}

type UserScheduleReader interface {
	GetUser(ctx context.Context, userID uuid.UUID) (usersdomain.User, error)
}

type WorkspaceSettingsService interface {
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspacedomain.WorkspaceSettings, error)
}
