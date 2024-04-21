package states

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the states storage.
type Repository interface {
	List(ctx context.Context) ([]CoreState, error)
}

// Service provides story-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// List returns a list of states.
func (s *Service) List(ctx context.Context) ([]CoreState, error) {
	s.log.Info(ctx, "business.core.states.list")
	ctx, span := web.AddSpan(ctx, "business.core.states.List")
	defer span.End()

	states, err := s.repo.List(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("states retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(states)),
	))
	return states, nil
}
