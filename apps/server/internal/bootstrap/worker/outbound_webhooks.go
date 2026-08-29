package workerbootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	outboundwebhooksrepository "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const outboundWebhookDispatchBatchSize = 16

type outboundWebhookDispatcher interface {
	DispatchOne(context.Context) (bool, error)
	RecoverExpiredLeases(context.Context) (int64, error)
}

type storyMutationEventDispatcher interface {
	DispatchBatch(context.Context) (int, error)
}

func buildOutboundWebhookDispatcher(
	pool *pgxpool.Pool,
	vault *credentialvault.Vault,
) (*outboundwebhooksservice.Dispatcher, error) {
	if pool == nil || vault == nil {
		return nil, errors.New("outbound webhook worker dependencies are required")
	}
	secrets, err := outboundwebhooksservice.NewSecretManager(vault)
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook secrets: %w", err)
	}
	httpClient, err := safehttp.New(safehttp.Config{
		Timeout:               10 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		MaxResponseBytes:      64 << 10,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook HTTP client: %w", err)
	}
	dispatcher, err := outboundwebhooksservice.NewDispatcher(
		outboundwebhooksrepository.New(pool), secrets, httpClient,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook dispatcher: %w", err)
	}
	return dispatcher, nil
}

func registerOutboundWebhookTask(
	mux *asynq.ServeMux,
	log *logger.Logger,
	storyEvents storyMutationEventDispatcher,
	dispatcher outboundWebhookDispatcher,
) error {
	if mux == nil || storyEvents == nil || dispatcher == nil {
		return errors.New("outbound webhook task dependencies are required")
	}
	mux.HandleFunc(tasks.TypeOutboundWebhookDispatch, outboundWebhookTaskHandler(log, storyEvents, dispatcher))
	return nil
}

func outboundWebhookTaskHandler(
	log *logger.Logger,
	storyEvents storyMutationEventDispatcher,
	dispatcher outboundWebhookDispatcher,
) asynq.HandlerFunc {
	return func(ctx context.Context, _ *asynq.Task) error {
		transferred, err := storyEvents.DispatchBatch(ctx)
		if err != nil {
			return fmt.Errorf("dispatch story mutation event intents: %w", err)
		}
		recovered, err := dispatcher.RecoverExpiredLeases(ctx)
		if err != nil {
			return fmt.Errorf("recover outbound webhook delivery leases: %w", err)
		}
		dispatched := 0
		for dispatched < outboundWebhookDispatchBatchSize {
			worked, dispatchErr := dispatcher.DispatchOne(ctx)
			if dispatchErr != nil {
				return fmt.Errorf("dispatch outbound webhook batch after %d deliveries: %w", dispatched, dispatchErr)
			}
			if !worked {
				break
			}
			dispatched++
		}
		if log != nil && (transferred > 0 || recovered > 0 || dispatched > 0) {
			log.Info(ctx, "Processed outbound webhook delivery batch", "story_events", transferred, "recovered_leases", recovered, "dispatched", dispatched)
		}
		return nil
	}
}
