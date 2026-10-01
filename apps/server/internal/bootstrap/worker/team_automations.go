package workerbootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/customfieldsadapter"
	"github.com/complexus-tech/projects-api/internal/bootstrap/workautomationsadapter"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	automationsrepository "github.com/complexus-tech/projects-api/internal/modules/workautomations/repository"
	automations "github.com/complexus-tech/projects-api/internal/modules/workautomations/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerTeamAutomationTask(mux *asynq.ServeMux, log *logger.Logger, pool *pgxpool.Pool, eventPublisher *publisher.Publisher) error {
	if mux == nil || pool == nil {
		return errors.New("team automation worker dependencies are required")
	}
	storyStore := storiesrepository.New(log, pool, storiesrepository.WithCustomFieldCreation(customfieldsadapter.CreationBinder))
	storyService := stories.New(log, storyStore, eventPublisher, nil)
	worker := automations.NewWorker(automationsrepository.New(pool), workautomationsadapter.New(storyService))
	mux.HandleFunc(tasks.TypeTeamAutomationDispatch, func(ctx context.Context, _ *asynq.Task) error {
		processed, err := worker.DispatchBatch(ctx)
		if err != nil {
			return fmt.Errorf("dispatch team automation batch: %w", err)
		}
		if processed > 0 {
			log.Info(ctx, "Processed team automation batch", "automations", processed)
		}
		return nil
	})
	return nil
}
