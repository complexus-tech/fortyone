package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/complexus-tech/projects-api/internal/handlers"
	"github.com/complexus-tech/projects-api/internal/mux"
	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/josemukorivo/config"
	"github.com/redis/go-redis/v9"
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
	Cache struct {
		Host     string `default:"localhost"`
		Port     string `default:"6379"`
		Password string `default:""`
		Name     int    `default:"0"`
	}
}

func main() {

	var logLevel slog.Level

	switch environ {
	case "development":
		logLevel = slog.LevelDebug
	case "production":
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelInfo
	}

	log := logger.NewWithText(os.Stderr, logLevel, service)
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
		return fmt.Errorf("error pinging database: %w", err)
	}

	log.Info(ctx, fmt.Sprintf("connected to database `%s`", cfg.DB.Name))

	defer func() {
		log.Info(ctx, "closing the database connection")
		db.Close()
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Cache.Host, cfg.Cache.Port),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.Name,
	})

	// Close the redis connection when the main function returns
	defer rdb.Close()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("error pinging redis: %w", err)
	}
	log.Info(ctx, fmt.Sprintf("connected to redis database `%d`", cfg.Cache.Name))

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

	muxCfg := mux.Config{
		DB:       db,
		Shutdown: shutdown,
		Log:      log,
		Tracer:   tracer,
	}
	apiMux := mux.New(muxCfg, handlers.BuildRoutes())

	server := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
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
