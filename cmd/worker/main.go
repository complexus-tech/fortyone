package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailerlite"
	"github.com/complexus-tech/projects-api/pkg/tasks"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/josemukorivo/config"
)

var (
	service = "projects-worker"
)

type WorkerConfig struct {
	DB struct {
		Host         string `default:"localhost" env:"APP_DB_HOST"`
		Port         string `default:"5432" env:"APP_DB_PORT"`
		User         string `default:"postgres" env:"APP_DB_USER"`
		Password     string `default:"password" env:"APP_DB_PASSWORD"`
		Name         string `default:"complexus" env:"APP_DB_NAME"`
		MaxIdleConns int    `default:"25" env:"APP_DB_MAX_IDLE_CONNS"`
		MaxOpenConns int    `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
		DisableTLS   bool   `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
	Redis struct {
		Host     string `default:"localhost" env:"APP_REDIS_HOST"`
		Port     string `default:"6379" env:"APP_REDIS_PORT"`
		Password string `default:"" env:"APP_REDIS_PASSWORD"`
		Name     int    `default:"0" env:"APP_REDIS_DB"`
	}
	Queues     map[string]int `default:"critical:6,default:3,low:1,onboarding:5,cleanup:2,notifications:4"`
	MailerLite struct {
		APIKey            string `env:"APP_MAILERLITE_API_KEY"`
		OnboardingGroupID string `env:"APP_MAILERLITE_ONBOARDING_GROUP_ID" default:"155224971194402532"`
	}
}

func main() {
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, service)
	ctx := context.Background()
	if err := run(ctx, log); err != nil {
		log.Error(ctx, "Worker process ended with error", "error", err)
		os.Exit(1)
	}
	log.Info(ctx, "Worker process shut down successfully")
}

func run(ctx context.Context, log *logger.Logger) error {
	log.Info(ctx, "Starting worker run function")
	var cfg WorkerConfig
	if err := config.Parse("app", &cfg); err != nil {
		return fmt.Errorf("error parsing worker configuration: %w", err)
	}
	if cfg.Queues == nil {
		cfg.Queues = map[string]int{"critical": 6, "default": 3, "low": 1, "onboarding": 5, "cleanup": 2, "notifications": 4}
	}

	// Initialize database connection
	dbConfig := database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}

	db, err := database.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer func() {
		log.Info(ctx, "closing database connection")
		db.Close()
	}()

	log.Info(ctx, "database connection established")

	rdbConn := asynq.RedisClientOpt{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Name,
	}

	// Set up scheduler for periodic cleanup tasks
	scheduler := asynq.NewScheduler(rdbConn, nil)

	_, err = scheduler.Register(
		"@daily",
		asynq.NewTask(tasks.TypeDeleteStories, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register delete stories task: %w", err)
	}

	_, err = scheduler.Register(
		"@weekly",
		asynq.NewTask(tasks.TypeTokenCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register token cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"@weekly",
		asynq.NewTask(tasks.TypeWebhookCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register webhook cleanup task: %w", err)
	}

	srv := asynq.NewServer(
		rdbConn,
		asynq.Config{
			Concurrency: 10,
			Queues:      cfg.Queues,
		},
	)

	// Initialize MailerLite service
	mailerLiteService, err := mailerlite.NewService(log, mailerlite.Config{
		APIKey: cfg.MailerLite.APIKey,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MailerLite service: %w", err)
	}

	if cfg.MailerLite.OnboardingGroupID == "" {
		return fmt.Errorf("mailerlite onboarding group ID is not configured")
	}

	// Set up task handlers
	workerTaskService := taskhandlers.NewWorkerHandlers(log, mailerLiteService, cfg.MailerLite.OnboardingGroupID)
	cleanupHandlers := taskhandlers.NewCleanupHandlers(log, db)

	mux := asynq.NewServeMux()
	// Register existing handlers
	mux.HandleFunc(tasks.TypeUserOnboardingStart, workerTaskService.HandleUserOnboardingStart)
	mux.HandleFunc(tasks.TypeSubscriberUpdate, workerTaskService.HandleSubscriberUpdate)
	mux.HandleFunc(tasks.TypeNotificationEmail, workerTaskService.HandleNotificationEmail)
	// Register cleanup handlers
	mux.HandleFunc(tasks.TypeTokenCleanup, cleanupHandlers.HandleTokenCleanup)
	mux.HandleFunc(tasks.TypeDeleteStories, cleanupHandlers.HandleDeleteStories)
	mux.HandleFunc(tasks.TypeWebhookCleanup, cleanupHandlers.HandleWebhookCleanup)

	h := asynqmon.New(asynqmon.Options{
		RootPath:     "/",
		RedisConnOpt: rdbConn,
	})
	http.Handle(h.RootPath()+"/", h)

	go func() {
		log.Info(ctx, "Starting Asynqmon monitoring server...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Error(ctx, "Failed to start HTTP server", "error", err)
		}
	}()

	// Start scheduler
	go func() {
		log.Info(ctx, "Starting cleanup scheduler...")
		if err := scheduler.Run(); err != nil {
			log.Error(ctx, "Failed to start scheduler", "error", err)
		}
	}()

	// Graceful shutdown for scheduler
	defer func() {
		log.Info(ctx, "shutting down scheduler")
		scheduler.Shutdown()
	}()

	log.Info(ctx, "Starting Asynq worker server...")
	if err := srv.Run(mux); err != nil {
		log.Error(ctx, "Asynq server Run() failed", "error", err)
		return fmt.Errorf("asynq server run error: %w", err)
	}
	return nil
}
