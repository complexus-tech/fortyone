package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	workerbootstrap "github.com/complexus-tech/projects-api/internal/bootstrap/worker"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

var (
	service = "projects-worker"
	version = "development"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, service)
	log.Info(ctx, "Starting worker process", "version", version)

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
