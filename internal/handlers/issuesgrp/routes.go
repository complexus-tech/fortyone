package issuesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/complexus-tech/projects-api/internal/core/issues/issuesrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {
	issuesService := issues.New(cfg.Log, issuesrepo.New(cfg.Log, cfg.DB))

	h := New(issuesService)

	app.Get("/issues/{id}", h.Get)
	app.Get("/my-issues", h.MyIssues)

}
