package workerbootstrap

import (
	"strings"

	"github.com/complexus-tech/projects-api/internal/bootstrap/mayaadapter"
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
	storyStore := storiesrepository.New(log, pool)
	storiesService := stories.New(log, storyStore, eventPublisher, nil)
	storiesService.ConfigureCommentCreator(buildStoryCommentCreator(log, pool))
	storiesService.ConfigureMayaActor(mayaActorID)
	storiesService.ConfigureAutoSchedulingEligibility(mayaRepository.WorkspaceCanUseMaya)
	reportsRepo := reportsrepository.New(log, pool)
	reportsService := reports.New(log, reportsRepo)
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
		Stories:           mayaadapter.New(storiesService, storyStore, mayaActorID),
		Reports:           mayaadapter.NewReports(reportsService, reportsRepo, mayaActorID),
		Calendar:          calendarService,
		Users:             usersService,
		WorkspaceSettings: workspacesrepository.New(pool),
		Planner:           planner,
		MayaActorID:       mayaActorID,
	})
}
