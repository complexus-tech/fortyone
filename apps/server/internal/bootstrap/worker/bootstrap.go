package workerbootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/jmoiron/sqlx"
)

type App struct {
	log       *logger.Logger
	db        *sqlx.DB
	server    *asynq.Server
	scheduler *asynq.Scheduler
	taskMux   *asynq.ServeMux
	redisOpt  asynq.RedisClientOpt
}

func New(ctx context.Context, log *logger.Logger) (App, error) {
	log.Info(ctx, "Starting worker run function")

	cfg, err := loadConfig()
	if err != nil {
		return App{}, fmt.Errorf("error parsing worker configuration: %w", err)
	}

	db, err := openDB(cfg)
	if err != nil {
		return App{}, err
	}
	log.Info(ctx, "database connection established")

	redisOpt := redisClientOpt(cfg)
	scheduler := asynq.NewScheduler(redisOpt, nil)
	if err := registerSchedules(scheduler); err != nil {
		_ = db.Close()
		return App{}, err
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues:      cfg.Queues,
		},
	)

	brevoService, err := brevo.NewService(brevo.Config{
		APIKey: cfg.Brevo.APIKey,
	}, log)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("error initializing brevo service: %w", err)
	}

	mailerService, err := mailer.NewService(mailer.Config{
		Host:        cfg.Email.Host,
		Port:        cfg.Email.Port,
		Username:    cfg.Email.Username,
		Password:    cfg.Email.Password,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
		Environment: cfg.Email.Environment,
		BaseDir:     cfg.Email.BaseDir,
	}, log)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("error initializing mailer service: %w", err)
	}

	systemUserID, err := uuid.Parse(cfg.System.UserID)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("invalid system user ID: %w", err)
	}

	taskMux := buildTaskMux(log, db, brevoService, mailerService, systemUserID)

	return App{
		log:       log,
		db:        db,
		server:    server,
		scheduler: scheduler,
		taskMux:   taskMux,
		redisOpt:  redisOpt,
	}, nil
}

func (a App) Run(ctx context.Context) error {
	defer func() {
		a.log.Info(ctx, "closing database connection")
		a.db.Close()
	}()

	monitor := asynqmon.New(asynqmon.Options{
		RootPath:     "/",
		RedisConnOpt: a.redisOpt,
	})
	http.Handle(monitor.RootPath()+"/", monitor)

	go func() {
		a.log.Info(ctx, "Starting Asynqmon monitoring server...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			a.log.Error(ctx, "Failed to start HTTP server", "error", err)
		}
	}()

	go func() {
		a.log.Info(ctx, "Starting cleanup scheduler...")
		if err := a.scheduler.Run(); err != nil {
			a.log.Error(ctx, "Failed to start scheduler", "error", err)
		}
	}()

	defer func() {
		a.log.Info(ctx, "shutting down scheduler")
		a.scheduler.Shutdown()
	}()

	a.log.Info(ctx, "Starting Asynq worker server...")
	if err := a.server.Run(a.taskMux); err != nil {
		a.log.Error(ctx, "Asynq server Run() failed", "error", err)
		return fmt.Errorf("asynq server run error: %w", err)
	}

	return nil
}
