package objectivestatus

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the objective statuses storage.
type Repository interface {
	Create(ctx context.Context, workspaceId uuid.UUID, status CoreNewObjectiveStatus) (CoreObjectiveStatus, error)
	Update(ctx context.Context, workspaceId, statusId uuid.UUID, status CoreUpdateObjectiveStatus) (CoreObjectiveStatus, error)
	Delete(ctx context.Context, workspaceId, statusId uuid.UUID) error
	List(ctx context.Context, workspaceId uuid.UUID) ([]CoreObjectiveStatus, error)
	Get(ctx context.Context, workspaceId, statusId uuid.UUID) (CoreObjectiveStatus, error)
	CountObjectivesWithStatus(ctx context.Context, statusID uuid.UUID, workspaceID uuid.UUID) (int, error)
	CountStatusesInCategory(ctx context.Context, workspaceID uuid.UUID, category string) (int, error)
}

// Service errors
var (
	ErrNotFound            = errors.New("objective status not found")
	ErrStatusHasObjectives = errors.New("cannot delete status with attached objectives")
	ErrLastInCategory      = errors.New("cannot delete the last status in a category")
)

// Service provides objective status operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new objective status service instance.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Create creates a new objective status.
func (s *Service) Create(ctx context.Context, workspaceId uuid.UUID, ns CoreNewObjectiveStatus) (CoreObjectiveStatus, error) {
	s.log.Info(ctx, "business.core.objectivestatus.Create")
	ctx, span := web.AddSpan(ctx, "business.core.objectivestatus.Create")
	defer span.End()

	status, err := s.repo.Create(ctx, workspaceId, ns)
	if err != nil {
		span.RecordError(err)
		return CoreObjectiveStatus{}, err
	}

	span.AddEvent("objective status created.", trace.WithAttributes(
		attribute.String("status.name", status.Name),
	))
	return status, nil
}

// Update updates an existing objective status.
func (s *Service) Update(ctx context.Context, workspaceId, statusId uuid.UUID, us CoreUpdateObjectiveStatus) (CoreObjectiveStatus, error) {
	s.log.Info(ctx, "business.core.objectivestatus.Update")
	ctx, span := web.AddSpan(ctx, "business.core.objectivestatus.Update")
	defer span.End()

	status, err := s.repo.Update(ctx, workspaceId, statusId, us)
	if err != nil {
		span.RecordError(err)
		return CoreObjectiveStatus{}, err
	}

	span.AddEvent("objective status updated.", trace.WithAttributes(
		attribute.String("status.id", statusId.String()),
	))
	return status, nil
}

// Delete deletes an objective status.
func (s *Service) Delete(ctx context.Context, workspaceId, statusId uuid.UUID) error {
	s.log.Info(ctx, "business.core.objectivestatus.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.objectivestatus.Delete")
	defer span.End()

	// 1. Get status details first to get category
	status, err := s.repo.Get(ctx, workspaceId, statusId)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get status: %w", err)
	}

	// 2. Check for objectives using this status
	objectivesCount, err := s.repo.CountObjectivesWithStatus(ctx, statusId, workspaceId)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check objectives: %w", err)
	}
	if objectivesCount > 0 {
		return ErrStatusHasObjectives
	}

	// 3. Check if it's the last in category for this workspace
	categoryCount, err := s.repo.CountStatusesInCategory(ctx, workspaceId, status.Category)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check category count: %w", err)
	}
	if categoryCount <= 1 {
		return ErrLastInCategory
	}

	// 4. Proceed with deletion
	if err := s.repo.Delete(ctx, workspaceId, statusId); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("objective status deleted.", trace.WithAttributes(
		attribute.String("status.id", statusId.String()),
	))
	return nil
}

// List returns a list of objective statuses.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID) ([]CoreObjectiveStatus, error) {
	s.log.Info(ctx, "business.core.objectivestatus.List")
	ctx, span := web.AddSpan(ctx, "business.core.objectivestatus.List")
	defer span.End()

	statuses, err := s.repo.List(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("objective statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
	))
	return statuses, nil
}
