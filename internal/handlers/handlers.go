package handlers

import (
	"github.com/complexus-tech/projects-api/internal/handlers/activitiesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/chatsessionsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/commentsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/documentsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/epicsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/healthgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/invitationsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/keyresultsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/labelsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/linksgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/notificationsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/objectivesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/objectivestatusgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/reportsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/searchgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/sprintsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/ssegrp"
	"github.com/complexus-tech/projects-api/internal/handlers/statesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/storiesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/subscriptionsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/teamsettingsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/teamsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/usersgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/workspacesgrp"
	"github.com/complexus-tech/projects-api/internal/mux"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type handlers struct{}

func New() handlers {
	return handlers{}
}

func (handlers) BuildAllRoutes(app *web.App, cfg mux.Config) {

	healthgrp.Routes(healthgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	storiesgrp.Routes(storiesgrp.Config{
		DB:          cfg.DB,
		Log:         cfg.Log,
		SecretKey:   cfg.SecretKey,
		Publisher:   cfg.Publisher,
		AzureConfig: cfg.AzureConfig,
		Validate:    cfg.Validate,
		Cache:       cfg.Cache,
	}, app)

	objectivesgrp.Routes(objectivesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	objectivestatusgrp.Routes(objectivestatusgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	labelsgrp.Routes(labelsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	linksgrp.Routes(linksgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	sprintsgrp.Routes(sprintsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	epicsgrp.Routes(epicsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	documentsgrp.Routes(documentsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	statesgrp.Routes(statesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	teamsgrp.Routes(teamsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	usersgrp.Routes(usersgrp.Config{
		DB:            cfg.DB,
		Log:           cfg.Log,
		SecretKey:     cfg.SecretKey,
		GoogleService: cfg.GoogleService,
		Publisher:     cfg.Publisher,
		TasksService:  cfg.TasksService,
		AzureConfig:   cfg.AzureConfig,
		Cache:         cfg.Cache,
	}, app)

	workspacesgrp.Routes(workspacesgrp.Config{
		DB:            cfg.DB,
		Log:           cfg.Log,
		SecretKey:     cfg.SecretKey,
		Publisher:     cfg.Publisher,
		Cache:         cfg.Cache,
		StripeClient:  cfg.StripeClient,
		WebhookSecret: cfg.WebhookSecret,
		TasksService:  cfg.TasksService,
		SystemUserID:  cfg.SystemUserID,
		AzureConfig:   cfg.AzureConfig,
	}, app)

	commentsgrp.Routes(commentsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	activitiesgrp.Routes(activitiesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	reportsgrp.Routes(reportsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	keyresultsgrp.Routes(keyresultsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	notificationsgrp.Routes(notificationsgrp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		Redis:        cfg.Redis,
		TasksService: cfg.TasksService,
	}, app)

	invitationsgrp.Routes(invitationsgrp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		Publisher:    cfg.Publisher,
		StripeClient: cfg.StripeClient,
		StripeSecret: cfg.WebhookSecret,
		TasksService: cfg.TasksService,
		SystemUserID: cfg.SystemUserID,
		Cache:        cfg.Cache,
	}, app)

	searchgrp.Routes(searchgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

	subscriptionsgrp.Routes(subscriptionsgrp.Config{
		DB:            cfg.DB,
		Log:           cfg.Log,
		SecretKey:     cfg.SecretKey,
		StripeClient:  cfg.StripeClient,
		WebhookSecret: cfg.WebhookSecret,
		Publisher:     cfg.Publisher,
		TasksService:  cfg.TasksService,
		SystemUserID:  cfg.SystemUserID,
		Cache:         cfg.Cache,
	}, app)

	ssegrp.Routes(ssegrp.Config{
		Log:        cfg.Log,
		SecretKey:  cfg.SecretKey,
		SSEHub:     cfg.SSEHub,
		CorsOrigin: cfg.CorsOrigin,
	}, app)

	teamsettingsgrp.Routes(teamsettingsgrp.Config{
		DB:           cfg.DB,
		Log:          cfg.Log,
		SecretKey:    cfg.SecretKey,
		TasksService: cfg.TasksService,
		Cache:        cfg.Cache,
	}, app)

	chatsessionsgrp.Routes(chatsessionsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Cache:     cfg.Cache,
	}, app)

}
