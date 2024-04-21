package teams

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the teams storage.
type Repository interface {
	List(ctx context.Context) ([]CoreTeam, error)
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

// List returns a list of teams.
func (s *Service) List(ctx context.Context) ([]CoreTeam, error) {
	s.log.Info(ctx, "business.core.teams.list")
	ctx, span := web.AddSpan(ctx, "business.core.teams.List")
	defer span.End()

	teams, err := s.repo.List(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("teams retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(teams)),
	))
	return teams, nil
}
