package stories

import (
	"context"
	"database/sql"
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
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error)
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Create(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, error)
	GetNextSequenceID(ctx context.Context, teamId uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error)
	MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error)
	List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error)
	GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]CoreStoryList, error)
	// Add this method later for labels
	// GetLabels(ctx context.Context, storyId uuid.UUID, workspaceId uuid.UUID) ([]Label, error)
	RecordActivity(ctx context.Context, storyID uuid.UUID, activityType string, description string, userID uuid.UUID) error
	GetLastActivity(ctx context.Context, storyID uuid.UUID, activityType string) (string, uuid.UUID, error)
}

// Add this new type
type CoreSingleStoryWithSubs struct {
	CoreSingleStory
	SubStories []CoreStoryList `json:"subStories"`
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
func (s *Service) Create(ctx context.Context, ns CoreNewStory, workspaceId uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.create")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Create")
	defer span.End()

	story := toCoreSingleStory(ns, workspaceId)

	cs, err := s.repo.Create(ctx, &story)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	span.AddEvent("story created.", trace.WithAttributes(
		attribute.String("story.title", cs.Title),
	))
	return cs, nil
}

// MyStories returns a list of stories.
func (s *Service) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.list")
	ctx, span := web.AddSpan(ctx, "business.core.stories.List")
	defer span.End()

	stories, err := s.repo.MyStories(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// List returns a list of stories for a workspace with additional filters.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.List")
	ctx, span := web.AddSpan(ctx, "business.core.stories.List")
	defer span.End()

	stories, err := s.repo.List(ctx, workspaceId, filters)
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
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.Get")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Get")
	defer span.End()

	story, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	subStories, err := s.repo.GetSubStories(ctx, id, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	story.SubStories = subStories

	// Fetch labels later when implemented
	// labels, err := s.repo.GetLabels(ctx, id, workspaceId)
	// if err != nil {
	//     span.RecordError(err)
	//     return CoreSingleStory{}, err
	// }
	// story.Labels = labels

	return story, nil
}

// Delete deletes the story with the specified ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Update updates the story with the specified ID.
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Update")
	defer span.End()

	for field, value := range updates {
		activityType := fmt.Sprintf("%s_updated", field)
		description := fmt.Sprintf("%s updated to %v", field, value)

		lastDescription, lastUserID, err := s.repo.GetLastActivity(ctx, id, activityType)
		if err != nil && err != sql.ErrNoRows {
			span.RecordError(err)
			return err
		}

		if lastDescription != description || lastUserID != userID {
			if err := s.repo.RecordActivity(ctx, id, activityType, description, userID); err != nil {
				span.RecordError(err)
				return err
			}
		}
	}

	if err := s.repo.Update(ctx, id, workspaceId, updates); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// BulkDelete deletes the stories with the specified IDs.
func (s *Service) BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkDelete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkDelete")
	defer span.End()

	if err := s.repo.BulkDelete(ctx, ids, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Restore restores the story with the specified ID.
func (s *Service) Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Restore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Restore")
	defer span.End()

	if err := s.repo.Restore(ctx, id, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (s *Service) BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkRestore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkRestore")
	defer span.End()

	if err := s.repo.BulkRestore(ctx, ids, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}
