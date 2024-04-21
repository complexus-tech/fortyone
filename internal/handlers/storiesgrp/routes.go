package storiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/stories/storiesrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB))

	h := New(storiesService)

	app.Get("/stories/{id}", h.Get)
	app.Get("/my-stories", h.MyStories)

}
