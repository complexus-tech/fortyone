package sprints

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the sprints storage.
type Repository interface {
	List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreSprint, error)
	Running(ctx context.Context, workspaceId uuid.UUID) ([]CoreSprint, error)
	Create(ctx context.Context, sprint CoreNewSprint) (CoreSprint, error)
	Update(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID, updates CoreUpdateSprint) (CoreSprint, error)
	Delete(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) error
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

// List returns a list of sprints.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreSprint, error) {
	s.log.Info(ctx, "business.core.sprints.list")
	ctx, span := web.AddSpan(ctx, "business.core.sprints.List")
	defer span.End()

	sprints, err := s.repo.List(ctx, workspaceId, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("sprints retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(sprints)),
	))
	return sprints, nil
}

// Running returns a list of running sprints.
func (s *Service) Running(ctx context.Context, workspaceId uuid.UUID) ([]CoreSprint, error) {
	s.log.Info(ctx, "business.core.sprints.running")
	ctx, span := web.AddSpan(ctx, "business.core.sprints.Running")
	defer span.End()

	sprints, err := s.repo.Running(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("running sprints retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(sprints)),
	))
	return sprints, nil
}

// Create creates a new sprint.
func (s *Service) Create(ctx context.Context, sprint CoreNewSprint) (CoreSprint, error) {
	s.log.Info(ctx, "business.core.sprints.create")
	ctx, span := web.AddSpan(ctx, "business.core.sprints.Create")
	defer span.End()

	result, err := s.repo.Create(ctx, sprint)
	if err != nil {
		span.RecordError(err)
		return CoreSprint{}, err
	}

	span.AddEvent("sprint created.", trace.WithAttributes(
		attribute.String("sprint.id", result.ID.String()),
		attribute.String("workspace.id", result.Workspace.String()),
	))
	return result, nil
}

// Update updates an existing sprint.
func (s *Service) Update(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID, updates CoreUpdateSprint) (CoreSprint, error) {
	s.log.Info(ctx, "business.core.sprints.update")
	ctx, span := web.AddSpan(ctx, "business.core.sprints.Update")
	defer span.End()

	result, err := s.repo.Update(ctx, sprintID, workspaceID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreSprint{}, err
	}

	span.AddEvent("sprint updated.", trace.WithAttributes(
		attribute.String("sprint.id", result.ID.String()),
		attribute.String("workspace.id", result.Workspace.String()),
	))
	return result, nil
}

// Delete removes a sprint.
func (s *Service) Delete(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.sprints.delete")
	ctx, span := web.AddSpan(ctx, "business.core.sprints.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, sprintID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("sprint deleted.", trace.WithAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))
	return nil
}
