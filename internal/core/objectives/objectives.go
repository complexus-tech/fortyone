package objectives

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the objectives storage.
type Repository interface {
	List(ctx context.Context) ([]CoreObjective, error)
}

// Service provides issue-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new issues service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// List returns a list of objectives.
func (s *Service) List(ctx context.Context) ([]CoreObjective, error) {
	s.log.Info(ctx, "business.core.objectives.list")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.List")
	defer span.End()

	objectives, err := s.repo.List(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("objectives retrieved.", trace.WithAttributes(
		attribute.Int("issue.count", len(objectives)),
	))
	return objectives, nil
}
