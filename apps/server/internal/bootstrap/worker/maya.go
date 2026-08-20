package workerbootstrap

import (
	"strings"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
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
	"github.com/jmoiron/sqlx"
)

func buildMayaService(
	log *logger.Logger,
	db *sqlx.DB,
	cfg Config,
	calendarService *calendar.Service,
	mayaActorID uuid.UUID,
	eventPublisher *publisher.Publisher,
) *maya.Service {
	mentionsRepo := mentionsrepository.New(log, db)
	storiesService := stories.New(log, storiesrepository.New(log, db), mentionsRepo, eventPublisher, nil)
	reportsService := reports.New(log, reportsrepository.New(log, db))
	usersService := users.New(log, usersrepository.New(log, db), nil)

	planner := maya.NewPlanner()
	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		aiClient := maya.NewOpenAICompatibleClient(maya.OpenAICompatibleConfig{
			APIKey: strings.TrimSpace(cfg.AIAPIKey),
			Model:  strings.TrimSpace(cfg.AIModel),
		})
		planner = maya.NewPlannerWithAdvisor(maya.NewOpenAIAdvisor(aiClient))
	}

	return maya.New(maya.Dependencies{
		Repository:        mayarepository.New(log, db),
		Stories:           storiesService,
		Reports:           reportsService,
		Calendar:          calendarService,
		Users:             usersService,
		WorkspaceSettings: workspacesrepository.New(log, db),
		Planner:           planner,
		MayaActorID:       mayaActorID,
	})
}
