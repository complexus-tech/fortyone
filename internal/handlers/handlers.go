package handlers

import (
	"github.com/complexus-tech/projects-api/internal/handlers/healthgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/issuesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/objectivesgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/sprintsgrp"
	"github.com/complexus-tech/projects-api/internal/mux"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type handlers struct{}

// BuildRoutes returns a new handlers instance which
// implements the mux.RouteAdder interface.
func BuildRoutes() handlers {
	return handlers{}
}

// BuildAllRoutes implements the mux.RouteAdder interface.
// It builds all the routes for the application.
func (handlers) BuildAllRoutes(app *web.App, cfg mux.Config) {

	// register the heath routes
	healthgrp.Routes(healthgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	// register the issues routes
	issuesgrp.Routes(issuesgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	// register the objectives routes
	objectivesgrp.Routes(objectivesgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	// register the sprints routes
	sprintsgrp.Routes(sprintsgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

}
