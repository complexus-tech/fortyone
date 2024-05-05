package stories

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Set of error variables for stories service.
var (
	ErrNotFound = errors.New("story not found")
)

// Repository provides access to the story storage.
type Repository interface {
	MyStories(ctx context.Context) ([]CoreStoryList, error)
	Get(ctx context.Context, id string) (CoreSingleStory, error)
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

// Get returns the story with the specified ID.
func (s *Service) Get(ctx context.Context, id string) (CoreSingleStory, error) {
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
