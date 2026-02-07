package api

import (
	activitieshttp "github.com/complexus-tech/projects-api/internal/modules/activities/http"
	chatsessionshttp "github.com/complexus-tech/projects-api/internal/modules/chatsessions/http"
	commentshttp "github.com/complexus-tech/projects-api/internal/modules/comments/http"
	documentshttp "github.com/complexus-tech/projects-api/internal/modules/documents/http"
	epicshttp "github.com/complexus-tech/projects-api/internal/modules/epics/http"
	healthhttp "github.com/complexus-tech/projects-api/internal/modules/health/http"
	invitationshttp "github.com/complexus-tech/projects-api/internal/modules/invitations/http"
	keyresultshttp "github.com/complexus-tech/projects-api/internal/modules/keyresults/http"
	labelshttp "github.com/complexus-tech/projects-api/internal/modules/labels/http"
	linkshttp "github.com/complexus-tech/projects-api/internal/modules/links/http"
	notificationshttp "github.com/complexus-tech/projects-api/internal/modules/notifications/http"
	objectiveshttp "github.com/complexus-tech/projects-api/internal/modules/objectives/http"
	objectivestatushttp "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/http"
	reportshttp "github.com/complexus-tech/projects-api/internal/modules/reports/http"
	searchhttp "github.com/complexus-tech/projects-api/internal/modules/search/http"
	sprintshttp "github.com/complexus-tech/projects-api/internal/modules/sprints/http"
	stateshttp "github.com/complexus-tech/projects-api/internal/modules/states/http"
	storieshttp "github.com/complexus-tech/projects-api/internal/modules/stories/http"
	subscriptionshttp "github.com/complexus-tech/projects-api/internal/modules/subscriptions/http"
	teamshttp "github.com/complexus-tech/projects-api/internal/modules/teams/http"
	teamsettingshttp "github.com/complexus-tech/projects-api/internal/modules/teamsettings/http"
	usershttp "github.com/complexus-tech/projects-api/internal/modules/users/http"
	workspaceshttp "github.com/complexus-tech/projects-api/internal/modules/workspaces/http"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	ssehttp "github.com/complexus-tech/projects-api/internal/sse/http"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type routes struct{}

func New() routes {
	return routes{}
}

func (routes) BuildAllRoutes(app *web.App, cfg mux.Config) {
	svcs := buildServices(cfg)

	healthhttp.Routes(healthhttp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	storieshttp.Routes(storieshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Publisher:      cfg.Publisher,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Validate:       cfg.Validate,
		Cache:          cfg.Cache,
		Stories:        svcs.stories,
		Comments:       svcs.comments,
		Links:          svcs.links,
		Attachments:    svcs.attachments,
	}, app)

	objectiveshttp.Routes(objectiveshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Cache:          cfg.Cache,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Objectives:     svcs.objectives,
		KeyResults:     svcs.keyResults,
		OKRActivities:  svcs.okrActivities,
		Attachments:    svcs.attachments,
	}, app)

	objectivestatushttp.Routes(objectivestatushttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.objectiveStats,
	}, app)

	labelshttp.Routes(labelshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.labels,
	}, app)

	linkshttp.Routes(linkshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.links,
	}, app)

	sprintshttp.Routes(sprintshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Cache:          cfg.Cache,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Sprints:        svcs.sprints,
		Attachments:    svcs.attachments,
	}, app)

	epicshttp.Routes(epicshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.epics,
	}, app)

	documentshttp.Routes(documentshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.documents,
	}, app)

	stateshttp.Routes(stateshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.states,
	}, app)

	teamshttp.Routes(teamshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.teams,
	}, app)

	usershttp.Routes(usershttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		GoogleService:  cfg.GoogleService,
		Publisher:      cfg.Publisher,
		TasksService:   cfg.TasksService,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Cache:          cfg.Cache,
		Users:          svcs.users,
		Attachments:    svcs.attachments,
	}, app)

	workspaceshttp.Routes(workspaceshttp.Config{
		DB:              cfg.DB,
		Log:             cfg.Log,
		SecretKey:       cfg.SecretKey,
		Publisher:       cfg.Publisher,
		Cache:           cfg.Cache,
		StripeClient:    cfg.StripeClient,
		WebhookSecret:   cfg.WebhookSecret,
		TasksService:    cfg.TasksService,
		SystemUserID:    cfg.SystemUserID,
		StorageConfig:   cfg.StorageConfig,
		StorageService:  cfg.StorageService,
		Workspaces:      svcs.workspaces,
		Teams:           svcs.teams,
		Stories:         svcs.stories,
		Statuses:        svcs.states,
		Users:           svcs.users,
		ObjectiveStatus: svcs.objectiveStats,
		Subscriptions:   svcs.subscriptions,
		Attachments:     svcs.attachments,
	}, app)

	commentshttp.Routes(commentshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.comments,
	}, app)

	activitieshttp.Routes(activitieshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Cache:          cfg.Cache,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Activities:     svcs.activities,
		Attachments:    svcs.attachments,
	}, app)

	reportshttp.Routes(reportshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Cache:          cfg.Cache,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		Reports:        svcs.reports,
		Attachments:    svcs.attachments,
	}, app)

	keyresultshttp.Routes(keyresultshttp.Config{
		DB:             cfg.DB,
		Log:            cfg.Log,
		SecretKey:      cfg.SecretKey,
		Cache:          cfg.Cache,
		StorageConfig:  cfg.StorageConfig,
		StorageService: cfg.StorageService,
		KeyResults:     svcs.keyResults,
		OKRActivities:  svcs.okrActivities,
		Attachments:    svcs.attachments,
	}, app)

	notificationshttp.Routes(notificationshttp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		Redis:        cfg.Redis,
		TasksService: cfg.TasksService,
		Cache:        cfg.Cache,
		Service:      svcs.notifications,
	}, app)

	invitationshttp.Routes(invitationshttp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		Publisher:    cfg.Publisher,
		StripeClient: cfg.StripeClient,
		StripeSecret: cfg.WebhookSecret,
		TasksService: cfg.TasksService,
		SystemUserID: cfg.SystemUserID,
		Cache:        cfg.Cache,
		Invitations:  svcs.invitations,
		UsersService: svcs.users,
	}, app)

	searchhttp.Routes(searchhttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.search,
	}, app)

	subscriptionshttp.Routes(subscriptionshttp.Config{
		DB:            cfg.DB,
		Log:           cfg.Log,
		SecretKey:     cfg.SecretKey,
		StripeClient:  cfg.StripeClient,
		WebhookSecret: cfg.WebhookSecret,
		Publisher:     cfg.Publisher,
		TasksService:  cfg.TasksService,
		SystemUserID:  cfg.SystemUserID,
		Cache:         cfg.Cache,
		Subscriptions: svcs.subscriptions,
		Users:         svcs.users,
		Workspaces:    svcs.workspaces,
	}, app)

	ssehttp.Routes(ssehttp.Config{
		Log:        cfg.Log,
		DB:         cfg.DB,
		SecretKey:  cfg.SecretKey,
		SSEHub:     cfg.SSEHub,
		CorsOrigin: cfg.CorsOrigin,
		Cache:      cfg.Cache,
	}, app)

	teamsettingshttp.Routes(teamsettingshttp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		TasksService: cfg.TasksService,
		Cache:        cfg.Cache,
		Service:      svcs.teamSettings,
	}, app)

	chatsessionshttp.Routes(chatsessionshttp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
		Service:   svcs.chatSessions,
	}, app)

}
