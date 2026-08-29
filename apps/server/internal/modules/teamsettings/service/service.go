package teamsettings

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

// Repository is the service-owned storage port. Generated sqlc types remain in
// the adapter and cannot leak through this interface.
type Repository interface {
	IsActiveTeamMember(ctx context.Context, teamID, workspaceID, actorID uuid.UUID) (bool, error)
	GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamSprintSettings, error)
	UpdateSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamSprintSettings, actor AuditActor) (CoreTeamSprintSettings, error)
	ReconcileSprintSchedule(ctx context.Context, settings CoreTeamSprintSettings, actor AuditActor) (int, error)
	GetStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamStoryAutomationSettings, error)
	UpdateStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamStoryAutomationSettings, actor AuditActor) (CoreTeamStoryAutomationSettings, error)
	GetEstimationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (CoreTeamEstimationSettings, error)
	UpdateEstimationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates CoreUpdateTeamEstimationSettings, actor AuditActor) (CoreTeamEstimationSettings, error)
}

// AutomationScheduler is a post-commit wake-up boundary. Its jobs scan durable
// settings state and also run periodically, so a dispatch failure must not turn
// a committed settings change into an apparent rollback.
type AutomationScheduler interface {
	ScheduleSprintCreation() error
	ScheduleStoryAutoClose() error
	ScheduleStoryAutoArchive() error
}

type Service struct {
	repo      Repository
	log       *logger.Logger
	scheduler AutomationScheduler
}

func New(log *logger.Logger, repo Repository, scheduler AutomationScheduler) *Service {
	return &Service{repo: repo, log: log, scheduler: scheduler}
}
