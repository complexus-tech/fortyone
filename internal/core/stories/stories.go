package stories

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

// Set of error variables for stories service.
var (
	ErrNotFound = errors.New("story not found")
)

// Repository provides access to the story storage.
type Repository interface {
	Get(ctx context.Context, id uuid.UUID) (CoreSingleStory, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	BulkRestore(ctx context.Context, ids []uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, updates map[string]any) error
	Create(ctx context.Context, story *CoreSingleStory) error
	GetNextSequenceID(ctx context.Context, teamId uuid.UUID) (int, error)
	MyStories(ctx context.Context) ([]CoreStoryList, error)
	TeamStories(ctx context.Context, teamId uuid.UUID) ([]CoreStoryList, error)
	ObjectiveStories(ctx context.Context, objectiveId uuid.UUID) ([]CoreStoryList, error)
	EpicStories(ctx context.Context, epicId uuid.UUID) ([]CoreStoryList, error)
	SprintStories(ctx context.Context, sprintId uuid.UUID) ([]CoreStoryList, error)
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

// Create creates a new story.
func (s *Service) Create(ctx context.Context, ns CoreNewStory) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.create")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Create")
	defer span.End()

	story := toCoreSingleStory(ns)

	sequenceID, err := s.repo.GetNextSequenceID(ctx, story.Team)

	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, fmt.Errorf("getting next sequence ID: %w", err)
	}
	story.SequenceID = sequenceID
	s.log.Info(ctx, "business.core.stories.create", "sequenceID", sequenceID)

	if err := s.repo.Create(ctx, &story); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	span.AddEvent("story created.", trace.WithAttributes(
		attribute.String("story.title", story.Title),
	))
	return story, nil
}

// MyStories returns a list of stories.
func (s *Service) MyStories(ctx context.Context) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.list")
	ctx, span := web.AddSpan(ctx, "business.core.stories.List")
	defer span.End()

	stories, err := s.repo.MyStories(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// TeamStories returns a list of stories for a team.
func (s *Service) TeamStories(ctx context.Context, teamId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.TeamStories")
	ctx, span := web.AddSpan(ctx, "business.core.stories.TeamStories")
	defer span.End()

	stories, err := s.repo.TeamStories(ctx, teamId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// ObjectiveStories returns a list of stories for an objective.
func (s *Service) ObjectiveStories(ctx context.Context, objectiveId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.ObjectiveStories")
	ctx, span := web.AddSpan(ctx, "business.core.stories.ObjectiveStories")
	defer span.End()

	stories, err := s.repo.ObjectiveStories(ctx, objectiveId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// EpicStories returns a list of stories for an epic.
func (s *Service) EpicStories(ctx context.Context, epicId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.EpicStories")
	ctx, span := web.AddSpan(ctx, "business.core.stories.EpicStories")
	defer span.End()

	stories, err := s.repo.EpicStories(ctx, epicId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// SprintStories returns a list of stories for a sprint.
func (s *Service) SprintStories(ctx context.Context, sprintId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.SprintStories")
	ctx, span := web.AddSpan(ctx, "business.core.stories.SprintStories")
	defer span.End()

	stories, err := s.repo.SprintStories(ctx, sprintId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// Get returns the story with the specified ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.Get")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Get")
	defer span.End()

	story, err := s.repo.Get(ctx, id)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	return story, nil
}

// Delete deletes the story with the specified ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Update updates the story with the specified ID.
func (s *Service) Update(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	s.log.Info(ctx, "business.core.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Update")
	defer span.End()

	if err := s.repo.Update(ctx, id, updates); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// BulkDelete deletes the stories with the specified IDs.
func (s *Service) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkDelete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkDelete")
	defer span.End()

	if err := s.repo.BulkDelete(ctx, ids); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Restore restores the story with the specified ID.
func (s *Service) Restore(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Restore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Restore")
	defer span.End()

	if err := s.repo.Restore(ctx, id); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (s *Service) BulkRestore(ctx context.Context, ids []uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkRestore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkRestore")
	defer span.End()

	if err := s.repo.BulkRestore(ctx, ids); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
