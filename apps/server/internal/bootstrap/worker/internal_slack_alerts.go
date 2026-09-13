package workerbootstrap

import (
	"context"
	"fmt"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const internalSlackAlertTask = "slack:internal-alert:dispatch"

func registerInternalSlackAlerts(mux *asynq.ServeMux, scheduler scheduleRegistrar, cfg slack.InternalAlertConfig, pool *pgxpool.Pool, vault *credentialvault.Vault, log *logger.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	repository := slackrepository.New(pool)
	dispatcher, err := slack.NewInternalAlertDispatcher(cfg, repository, repository, vault, log)
	if err != nil {
		return err
	}
	mux.HandleFunc(internalSlackAlertTask, func(ctx context.Context, _ *asynq.Task) error {
		return dispatcher.Dispatch(ctx)
	})
	_, err = scheduler.Register("@every 5s", asynq.NewTask(internalSlackAlertTask, nil),
		asynq.Queue("notifications"), asynq.MaxRetry(0), asynq.Timeout(35*time.Second), asynq.Unique(4*time.Second))
	if err != nil {
		return fmt.Errorf("register internal Slack alert dispatch: %w", err)
	}
	return nil
}
