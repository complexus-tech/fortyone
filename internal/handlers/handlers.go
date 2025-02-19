package handlers

import (
	"github.com/complexus-tech/projects-api/internal/handlers/activitiesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/commentsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/documentsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/epicsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/healthgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/keyresultsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/labelsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/linksgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/notificationsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/objectivesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/objectivestatusgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/reportsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/sprintsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/statesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/storiesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/teamsgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/usersgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/workspacesgrp"
	"github.com/complexus-tech/projects-api/internal/mux"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type handlers struct{}

// New returns a new handlers instance which implements the mux.RouteAdder interface.
func New() handlers {
	return handlers{}
}

// BuildAllRoutes implements the mux.RouteAdder interface. It builds all the routes for the application.
func (handlers) BuildAllRoutes(app *web.App, cfg mux.Config) {

	// register the heath routes
	healthgrp.Routes(healthgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	// register the stories routes
	storiesgrp.Routes(storiesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Publisher: cfg.Publisher,
	}, app)

	// register the objectives routes
	objectivesgrp.Routes(objectivesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the objective statuses routes
	objectivestatusgrp.Routes(objectivestatusgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the labels routes
	labelsgrp.Routes(labelsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the links routes
	linksgrp.Routes(linksgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the sprints routes
	sprintsgrp.Routes(sprintsgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
		// SecretKey: cfg.SecretKey,
	}, app)

	// register epics routes
	epicsgrp.Routes(epicsgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
		// SecretKey: cfg.SecretKey,
	}, app)

	// register the documents routes
	documentsgrp.Routes(documentsgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
		// SecretKey: cfg.SecretKey,
	}, app)

	// register the states routes
	statesgrp.Routes(statesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the teams routes
	teamsgrp.Routes(teamsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the users routes
	usersgrp.Routes(usersgrp.Config{
		DB:            cfg.DB,
		Log:           cfg.Log,
		SecretKey:     cfg.SecretKey,
		GoogleService: cfg.GoogleService,
	}, app)

	// register the workspaces routes
	workspacesgrp.Routes(workspacesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
		Publisher: cfg.Publisher,
	}, app)

	// register the comments routes
	commentsgrp.Routes(commentsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the activities routes
	activitiesgrp.Routes(activitiesgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the reports routes
	reportsgrp.Routes(reportsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the key results routes
	keyresultsgrp.Routes(keyresultsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

	// register the notifications routes
	notificationsgrp.Routes(notificationsgrp.Config{
		DB:        cfg.DB,
		Log:       cfg.Log,
		SecretKey: cfg.SecretKey,
	}, app)

}
