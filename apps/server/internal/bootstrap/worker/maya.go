package workerbootstrap

import (
	"strings"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildMayaService(
	log *logger.Logger,
	pool *pgxpool.Pool,
	mayaRepository *mayarepository.Repo,
	cfg Config,
	calendarService *calendar.Service,
	mayaActorID uuid.UUID,
	eventPublisher *publisher.Publisher,
) *maya.Service {
	storiesService := stories.New(log, storiesrepository.New(log, pool), eventPublisher, nil)
	storiesService.ConfigureCommentCreator(buildStoryCommentCreator(log, pool))
	storiesService.ConfigureMayaActor(mayaActorID)
	storiesService.ConfigureAutoSchedulingEligibility(mayaRepository.WorkspaceCanUseMaya)
	reportsService := reports.New(log, reportsrepository.New(log, pool))
	usersService := users.New(log, usersrepository.New(pool), nil)

	planner := maya.NewPlanner()
	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		aiClient := maya.NewOpenAICompatibleClient(maya.OpenAICompatibleConfig{
			APIKey: strings.TrimSpace(cfg.AIAPIKey),
			Model:  strings.TrimSpace(cfg.AIModel),
		})
		planner = maya.NewPlannerWithAdvisor(maya.NewOpenAIAdvisor(aiClient))
	}

	return maya.New(maya.Dependencies{
		Repository:        mayaRepository,
		Realtime:          mayaRepository,
		Stories:           storiesService,
		Reports:           reportsService,
		Calendar:          calendarService,
		Users:             usersService,
		WorkspaceSettings: workspacesrepository.New(pool),
		Planner:           planner,
		MayaActorID:       mayaActorID,
	})
}
