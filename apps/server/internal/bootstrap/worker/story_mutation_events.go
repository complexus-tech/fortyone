package workerbootstrap

import (
	"errors"
	"fmt"

	providers "github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	outboundwebhooksrepository "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildStoryMutationEventDispatcher(
	log *logger.Logger,
	pool *pgxpool.Pool,
) (*stories.StoryMutationEventDispatcher, error) {
	if pool == nil {
		return nil, errors.New("story mutation event worker dependencies are required")
	}
	outboundPublisher, err := outboundwebhooksservice.NewPublisher(outboundwebhooksrepository.New(pool))
	if err != nil {
		return nil, fmt.Errorf("initialize outbound story event publisher: %w", err)
	}
	adapter, err := providers.NewOutboundStoryMutationPublisher(outboundPublisher)
	if err != nil {
		return nil, fmt.Errorf("initialize outbound story event adapter: %w", err)
	}
	dispatcher, err := stories.NewStoryMutationEventDispatcher(
		storiesrepository.NewMutationRepository(log, pool),
		adapter,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize story mutation event dispatcher: %w", err)
	}
	return dispatcher, nil
}
