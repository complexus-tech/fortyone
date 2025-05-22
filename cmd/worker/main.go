package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"

	"github.com/hibiken/asynq"
	"github.com/josemukorivo/config"
)

var (
	service = "projects-worker"
)

type WorkerConfig struct {
	Redis struct {
		Host     string `default:"localhost" env:"APP_REDIS_HOST"`
		Port     string `default:"6379" env:"APP_REDIS_PORT"`
		Password string `default:"" env:"APP_REDIS_PASSWORD"`
		Name     int    `default:"0" env:"APP_REDIS_DB"`
	}
	Queues map[string]int `default:"critical:6,default:3,low:1,onboarding:5"`
	// MailerLite struct {
	//  APIKey  string `env:"APP_MAILERLITE_API_KEY"`
	//  GroupID string `env:"APP_MAILERLITE_ONBOARDING_GROUP_ID"`
	// }
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
		cfg.Queues = map[string]int{"critical": 6, "default": 3, "low": 1, "onboarding": 5}
	}
	rdbConn := asynq.RedisClientOpt{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Name,
	}
	srv := asynq.NewServer(
		rdbConn,
		asynq.Config{
			Concurrency: 10,
			Queues:      cfg.Queues,
		},
	)

	workerTaskService := taskhandlers.NewWorkerHandlers(log)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeUserOnboardingStart, workerTaskService.HandleUserOnboardingStart)
	log.Info(ctx, "Starting Asynq worker server...")
	if err := srv.Run(mux); err != nil {
		log.Error(ctx, "Asynq server Run() failed", "error", err)
		return fmt.Errorf("asynq server run error: %w", err)
	}
	return nil
}
