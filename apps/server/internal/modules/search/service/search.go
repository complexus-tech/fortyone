package search

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the search storage.
type Repository interface {
	SearchStories(ctx context.Context, workspaceID uuid.UUID, userId uuid.UUID, params SearchParams) ([]CoreSearchStory, int, error)
	SearchObjectives(ctx context.Context, workspaceID uuid.UUID, userId uuid.UUID, params SearchParams) ([]CoreSearchObjective, int, error)
}

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
	ctx, span := web.AddSpan(ctx, "business.core.search.Search")
	defer span.End()

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.Type == "" {
		params.Type = SearchTypeAll
	}

	span.AddEvent("search initialized", trace.WithAttributes(
		attribute.String("search.query", params.Query),
		attribute.String("search.type", string(params.Type)),
	))

	var storiesResult []CoreSearchStory
	var objectivesResult []CoreSearchObjective
	var totalStories, totalObjectives int
	var err error

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
