package objectives

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service errors
var (
	ErrNotFound   = errors.New("objective not found")
	ErrNameExists = errors.New("an objective with this name already exists")
)

// Repository provides access to the objectives storage.
type Repository interface {
	List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]CoreObjective, error)
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreObjective, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Create(ctx context.Context, objective CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (CoreObjective, []keyresults.CoreKeyResult, error)
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

// Get returns an objective by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreObjective, error) {
	s.log.Info(ctx, "business.core.objectives.Get")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Get")
	defer span.End()

	objective, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreObjective{}, err
	}

	span.AddEvent("objective retrieved.", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return objective, nil
}

// List returns a list of objectives.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]CoreObjective, error) {
	s.log.Info(ctx, "business.core.objectives.list")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.List")
	defer span.End()

	objectives, err := s.repo.List(ctx, workspaceId, userID, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("objectives retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(objectives)),
	))
	return objectives, nil
}

// Update updates an objective in the system
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	s.log.Info(ctx, "business.core.objectives.Update")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Update")
	defer span.End()

	if err := s.repo.Update(ctx, id, workspaceId, updates); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		span.RecordError(err)
		return err
	}

	span.AddEvent("objective updated", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// Delete removes an objective from the system
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.objectives.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		span.RecordError(err)
		return err
	}

	span.AddEvent("objective deleted", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// Create creates a new objective with optional key results
func (s *Service) Create(ctx context.Context, newObjective CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (CoreObjective, []keyresults.CoreKeyResult, error) {
	s.log.Info(ctx, "business.core.objectives.Create")
	ctx, span := web.AddSpan(ctx, "business.core.objectives.Create")
	defer span.End()

	createdObj, createdKRs, err := s.repo.Create(ctx, newObjective, workspaceID, keyResults)
	if err != nil {
		span.RecordError(err)
		return CoreObjective{}, nil, err
	}

	span.AddEvent("objective created.", trace.WithAttributes(
		attribute.String("objective.id", createdObj.ID.String()),
		attribute.Int("key_results.count", len(createdKRs)),
	))

	return createdObj, createdKRs, nil
}
