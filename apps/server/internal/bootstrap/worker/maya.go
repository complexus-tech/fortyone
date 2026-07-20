package workerbootstrap

import (
	"strings"

	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teamsettingsrepository "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository"
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func buildMayaService(log *logger.Logger, db *sqlx.DB, cfg Config, mayaActorID uuid.UUID) *maya.Service {
	mentionsRepo := mentionsrepository.New(log, db)
	storiesService := stories.New(log, storiesrepository.New(log, db), mentionsRepo, nil, nil)
	reportsService := reports.New(log, reportsrepository.New(log, db))
	calendarService := calendar.New(log, calendarrepository.New(log, db), calendar.Config{
		SecretKey:  cfg.Auth.SecretKey,
		WebsiteURL: cfg.Website.URL,
		Providers:  map[calendar.Provider]calendar.CalendarProvider{},
	})
	usersService := users.New(log, usersrepository.New(log, db), nil)

	planner := maya.NewPlanner()
	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		aiClient := maya.NewOpenAICompatibleClient(maya.OpenAICompatibleConfig{
			APIKey: strings.TrimSpace(cfg.AIAPIKey),
		})
		planner = maya.NewPlannerWithAdvisor(maya.NewOpenAIAdvisor(aiClient))
	}

	return maya.New(maya.Dependencies{
		Repository:   mayarepository.New(log, db),
		Stories:      storiesService,
		Reports:      reportsService,
		Calendar:     calendarService,
		Users:        usersService,
		TeamSettings: teamsettings.New(log, teamsettingsrepository.New(log, db), nil),
		Planner:      planner,
		MayaActorID:  mayaActorID,
	})
}
