package handlers

import (
	"github.com/complexus-tech/projects-api/internal/handlers/healthgrp"
	"github.com/complexus-tech/projects-api/internal/handlers/issuesgrp"
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

	healthgrp.Routes(healthgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

	issuesgrp.Routes(issuesgrp.Config{
		DB:  cfg.DB,
		Log: cfg.Log,
	}, app)

}
