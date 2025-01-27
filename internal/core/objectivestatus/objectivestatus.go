package objectivestatus

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CoreObjectiveStatus struct {
	ID         uuid.UUID
	Name       string
	Color      string
	Category   string
	OrderIndex int
	Team       uuid.UUID
	Workspace  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Repository provides access to the objective statuses storage.
type Repository interface {
	List(ctx context.Context, workspaceId uuid.UUID) ([]CoreObjectiveStatus, error)
}

// Service provides objective status-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new objective statuses service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// List returns a list of objective statuses.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID) ([]CoreObjectiveStatus, error) {
	s.log.Info(ctx, "business.core.objectivestatus.list")
	ctx, span := web.AddSpan(ctx, "business.core.objectivestatus.List")
	defer span.End()

	states, err := s.repo.List(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("objective statuses retrieved.", trace.WithAttributes(
		attribute.Int("objective_status.count", len(states)),
	))
	return states, nil
}
