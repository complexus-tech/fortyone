package main

import (
	"context"
	"log/slog"
	"os"

	workerbootstrap "github.com/complexus-tech/projects-api/internal/bootstrap/worker"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

var (
	service = "projects-worker"
)

func main() {
	ctx := context.Background()
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, service)

	app, err := workerbootstrap.New(ctx, log)
	if err != nil {
		log.Error(ctx, "Worker process ended with error", "error", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		log.Error(ctx, "Worker process ended with error", "error", err)
		os.Exit(1)
	}

	log.Info(ctx, "Worker process shut down successfully")
}
