package states

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

// Repository provides access to the states storage.
type Repository interface {
	Create(ctx context.Context, workspaceId uuid.UUID, state CoreNewState) (CoreState, error)
	Update(ctx context.Context, workspaceId, stateId uuid.UUID, state CoreUpdateState) (CoreState, error)
	Delete(ctx context.Context, workspaceId, stateId uuid.UUID) error
	List(ctx context.Context, workspaceId uuid.UUID) ([]CoreState, error)
	TeamList(ctx context.Context, workspaceId uuid.UUID, teamId uuid.UUID) ([]CoreState, error)
	Get(ctx context.Context, workspaceId, stateId uuid.UUID) (CoreState, error)
	CountStoriesWithStatus(ctx context.Context, statusID uuid.UUID) (int, error)
	CountStatusesInCategory(ctx context.Context, teamID uuid.UUID, category string) (int, error)
}

// Service errors
var (
	ErrNotFound         = errors.New("status not found")
	ErrStatusHasStories = errors.New("cannot delete status with attached stories")
	ErrLastInCategory   = errors.New("cannot delete the last status in a category")
)

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

// Create creates a new state.
func (s *Service) Create(ctx context.Context, workspaceId uuid.UUID, ns CoreNewState) (CoreState, error) {
	s.log.Info(ctx, "business.core.states.Create")
	ctx, span := web.AddSpan(ctx, "business.core.states.Create")
	defer span.End()

	state, err := s.repo.Create(ctx, workspaceId, ns)
	if err != nil {
		span.RecordError(err)
		return CoreState{}, err
	}

	span.AddEvent("state created.", trace.WithAttributes(
		attribute.String("state.name", state.Name),
	))
	return state, nil
}

// Update updates an existing state.
func (s *Service) Update(ctx context.Context, workspaceId, stateId uuid.UUID, us CoreUpdateState) (CoreState, error) {
	s.log.Info(ctx, "business.core.states.Update")
	ctx, span := web.AddSpan(ctx, "business.core.states.Update")
	defer span.End()

	state, err := s.repo.Update(ctx, workspaceId, stateId, us)
	if err != nil {
		span.RecordError(err)
		return CoreState{}, err
	}

	span.AddEvent("state updated.", trace.WithAttributes(
		attribute.String("state.id", stateId.String()),
	))
	return state, nil
}

// Delete deletes a state.
func (s *Service) Delete(ctx context.Context, workspaceId, stateId uuid.UUID) error {
	s.log.Info(ctx, "business.core.states.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.states.Delete")
	defer span.End()

	// 1. Get status details first to get team and category
	status, err := s.repo.Get(ctx, workspaceId, stateId)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get status: %w", err)
	}

	// 2. Check for stories using this status
	storiesCount, err := s.repo.CountStoriesWithStatus(ctx, stateId)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check stories: %w", err)
	}
	if storiesCount > 0 {
		return ErrStatusHasStories
	}

	// 3. Check if it's the last in category for this team
	categoryCount, err := s.repo.CountStatusesInCategory(ctx, status.Team, status.Category)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check category count: %w", err)
	}
	if categoryCount <= 1 {
		return ErrLastInCategory
	}

	// 4. Proceed with deletion
	if err := s.repo.Delete(ctx, workspaceId, stateId); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("status deleted.", trace.WithAttributes(
		attribute.String("status.id", stateId.String()),
	))
	return nil
}

// List returns a list of states.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID) ([]CoreState, error) {
	s.log.Info(ctx, "business.core.states.List")
	ctx, span := web.AddSpan(ctx, "business.core.states.List")
	defer span.End()

	states, err := s.repo.List(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("states retrieved.", trace.WithAttributes(
		attribute.Int("states.count", len(states)),
	))
	return states, nil
}

// TeamList returns a list of states for a team.
func (s *Service) TeamList(ctx context.Context, workspaceId uuid.UUID, teamId uuid.UUID) ([]CoreState, error) {
	s.log.Info(ctx, "business.core.states.TeamList")
	ctx, span := web.AddSpan(ctx, "business.core.states.TeamList")
	defer span.End()

	states, err := s.repo.TeamList(ctx, workspaceId, teamId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("states retrieved.", trace.WithAttributes(
		attribute.Int("states.count", len(states)),
	))
	return states, nil
}

// Get returns a state by ID.
func (s *Service) Get(ctx context.Context, workspaceId uuid.UUID, stateId uuid.UUID) (CoreState, error) {
	s.log.Info(ctx, "business.core.states.Get")
	ctx, span := web.AddSpan(ctx, "business.core.states.Get")
	defer span.End()

	state, err := s.repo.Get(ctx, workspaceId, stateId)
	if err != nil {
		span.RecordError(err)
		return CoreState{}, err
	}

	span.AddEvent("state retrieved.", trace.WithAttributes(
		attribute.String("state.id", stateId.String()),
		attribute.String("state.name", state.Name),
	))
	return state, nil
}
