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
	service                = "projects-worker"
	version                = "development"
	workerBootstrapFailure = logger.MustDefineError(
		"worker.bootstrap.failed",
		"Worker failed during startup",
	)
	workerRuntimeFailure = logger.MustDefineError(
		"worker.runtime.failed",
		"Worker runtime stopped unexpectedly",
	)
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, service)
	log.Info(ctx, "Starting worker process", "version", version)

	app, err := workerbootstrap.New(ctx, log)
	if err != nil {
		logWorkerFailure(ctx, log, "bootstrap", workerBootstrapFailure.WrapIfUnclassified(err))
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		logWorkerFailure(ctx, log, "runtime", workerRuntimeFailure.WrapIfUnclassified(err))
		os.Exit(1)
	}

	log.Info(ctx, "Worker process shut down successfully")
}

func logWorkerFailure(ctx context.Context, log *logger.Logger, phase string, err error) {
	log.Error(
		ctx,
		"Worker process ended with error",
		"version", version,
		"phase", phase,
		"error", err,
	)
}
