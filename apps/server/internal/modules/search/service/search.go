package search

import (
	"context"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/pkg/logger"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the search storage.
type Repository interface {
	SearchStories(ctx context.Context, workspaceID uuid.UUID, userId uuid.UUID, params SearchParams) ([]CoreSearchStory, int, error)
	SearchObjectives(ctx context.Context, workspaceID uuid.UUID, userId uuid.UUID, params SearchParams) ([]CoreSearchObjective, int, error)
	FindSimilarStories(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, title string, teamID *uuid.UUID, limit int) ([]CoreSimilarStory, error)
}

const (
	defaultSimilarStoriesLimit       = 3
	maxSimilarStoriesLimit           = 5
	minimumSimilarityTitleCharacters = 10
)

// Service provides search-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new search service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Search searches for content based on the provided parameters.
func (s *Service) Search(ctx context.Context, workspaceID uuid.UUID, userId uuid.UUID, params SearchParams) (CoreSearchResult, error) {
	s.log.Info(ctx, "business.core.search.Search")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.search.Search")
	defer span.End()

	if workspaceID == uuid.Nil || userId == uuid.Nil {
		return CoreSearchResult{}, invalidSearchParams("workspace and actor IDs are required")
	}

	var err error
	params, err = normalizeSearchParams(params)
	if err != nil {
		span.RecordError(err)
		return CoreSearchResult{}, err
	}
	span.SetAttributes(
		attribute.String("workspace.id", workspaceID.String()),
		attribute.String("actor.id", userId.String()),
	)

	span.AddEvent("search initialized", trace.WithAttributes(
		attribute.Bool("search.has_query", params.Query != ""),
		attribute.Int("search.query_runes", utf8.RuneCountInString(params.Query)),
		attribute.String("search.type", string(params.Type)),
	))

	var storiesResult []CoreSearchStory
	var objectivesResult []CoreSearchObjective
	var totalStories, totalObjectives int

	// Search stories if requested
	if params.Type == SearchTypeAll || params.Type == SearchTypeStories {
		storiesResult, totalStories, err = s.repo.SearchStories(ctx, workspaceID, userId, params)
		if err != nil {
			span.RecordError(err)
			return CoreSearchResult{}, err
		}
	}

	// Search objectives if requested
	if params.Type == SearchTypeAll || params.Type == SearchTypeObjectives {
		objectivesResult, totalObjectives, err = s.repo.SearchObjectives(ctx, workspaceID, userId, params)
		if err != nil {
			span.RecordError(err)
			return CoreSearchResult{}, err
		}
	}

	span.AddEvent("search completed", trace.WithAttributes(
		attribute.Int("search.stories.count", len(storiesResult)),
		attribute.Int("search.objectives.count", len(objectivesResult)),
	))

	return CoreSearchResult{
		Stories:         storiesResult,
		Objectives:      objectivesResult,
		TotalStories:    totalStories,
		TotalObjectives: totalObjectives,
	}, nil
}

// FindSimilarStories returns stories whose titles are close to a proposed title.
func (s *Service) FindSimilarStories(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, title string, teamID *uuid.UUID, limit int) ([]CoreSimilarStory, error) {
	s.log.Info(ctx, "business.core.search.FindSimilarStories")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.search.FindSimilarStories")
	defer span.End()

	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil, invalidSearchParams("workspace and actor IDs are required")
	}
	if teamID != nil && *teamID == uuid.Nil {
		return nil, invalidSearchParams("team ID is required")
	}

	var err error
	title, err = normalizeSimilarityTitle(title)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.SetAttributes(
		attribute.String("workspace.id", workspaceID.String()),
		attribute.String("actor.id", userID.String()),
		attribute.Int("search.query_runes", utf8.RuneCountInString(title)),
	)
	if len([]rune(title)) < minimumSimilarityTitleCharacters {
		return []CoreSimilarStory{}, nil
	}
	if limit <= 0 {
		limit = defaultSimilarStoriesLimit
	}
	if limit > maxSimilarStoriesLimit {
		limit = maxSimilarStoriesLimit
	}

	stories, err := s.repo.FindSimilarStories(ctx, workspaceID, userID, title, teamID, limit)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("similar stories found", trace.WithAttributes(
		attribute.Int("search.similar_stories.count", len(stories)),
	))
	return stories, nil
}
