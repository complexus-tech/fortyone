package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/complexus-tech/projects-api/internal/handlers"
	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/josemukorivo/config"
)

var (
	service = "projects-api"
	version = "0.0.1"
	environ = "development"
)

type Config struct {
	Web struct {
		APIHost         string        `default:"localhost:8000" env:"APP_API_HOST"`
		ReadTimeout     time.Duration `default:"5s" env:"APP_API_READ_TIMEOUT"`
		WriteTimeout    time.Duration `default:"10s" env:"APP_API_WRITE_TIMEOUT"`
		IdleTimeout     time.Duration `default:"120s" env:"APP_API_IDLE_TIMEOUT"`
		ShutdownTimeout time.Duration `default:"30s" env:"APP_API_SHUTDOWN_TIMEOUT"`
		DebugHost       string        `default:"localhost:9000" env:"APP_API_DEBUG_HOST"`
	}
	DB struct {
		Host         string `default:"localhost"`
		Port         string `default:"5432"`
		User         string `default:"postgres"`
		Password     string `default:"password"`
		Name         string `default:"complexus"`
		MaxIdleConns int    `default:"2" env:"APP_DB_MAX_IDLE_CONNS"`
		MaxOpenConns int    `default:"3" env:"APP_DB_MAX_OPEN_CONNS"`
		DisableTLS   bool   `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
}

func main() {

	log := logger.NewWithText(os.Stderr, slog.LevelInfo, service)
	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, fmt.Sprintf("error shutting down: %s", err))
		os.Exit(1)
	}

}

// run starts the HTTP server, tracing and listens for OS signals to gracefully shutdown.
func run(ctx context.Context, log *logger.Logger) error {
	var cfg Config
	err := config.Parse("app", &cfg)
	if err != nil {
		return fmt.Errorf("error parsing config: %s", err)
	}

	db, err := database.Open(database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})

	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging db: %w", err)
	}

	log.Info(ctx, fmt.Sprintf("connected to db %s.", cfg.DB.Name))

	defer func() {
		log.Info(ctx, "closing the database connection")
		db.Close()
	}()

	shutdown := make(chan os.Signal, 1)
	// Start Tracing
	t := tracing.New(service, version, environ)
	traceProvider, err := t.StartTracing()
	if err != nil {
		return fmt.Errorf("error starting tracing: %w", err)
	}
	// Graceful shutdown of tracing if server is stopped
	defer traceProvider.Shutdown(ctx)
	tracer := traceProvider.Tracer(service)

	server := http.Server{
		Addr: cfg.Web.APIHost,
		Handler: handlers.API(handlers.Config{
			DB:       db,
			Shutdown: shutdown,
			Log:      log,
			Tracer:   tracer,
		}),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info(ctx, "started http server on :8000")
		serverErrors <- server.ListenAndServe()
	}()

	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("error: listening and serving: %s", err)

	case <-shutdown:
		log.Info(ctx, "starting shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Info(ctx, "error: graceful shutdown: %s", err)
			if err := server.Close(); err != nil {
				log.Info(ctx, "error: could not stop http server: %s", err)
			}
		}
	}
	return nil
}
