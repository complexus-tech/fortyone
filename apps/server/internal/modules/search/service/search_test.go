package search

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	similarStories  []CoreSimilarStory
	findCalls       int
	title           string
	teamID          *uuid.UUID
	limit           int
	storyCalls      int
	objectiveCalls  int
	storyParams     SearchParams
	objectiveParams SearchParams
}

func (r *repositoryStub) SearchStories(_ context.Context, _ uuid.UUID, _ uuid.UUID, params SearchParams) ([]CoreSearchStory, int, error) {
	r.storyCalls++
	r.storyParams = params
	return nil, 0, nil
}

func (r *repositoryStub) SearchObjectives(_ context.Context, _ uuid.UUID, _ uuid.UUID, params SearchParams) ([]CoreSearchObjective, int, error) {
	r.objectiveCalls++
	r.objectiveParams = params
	return nil, 0, nil
}

func (r *repositoryStub) FindSimilarStories(_ context.Context, _ uuid.UUID, _ uuid.UUID, title string, teamID *uuid.UUID, limit int) ([]CoreSimilarStory, error) {
	r.findCalls++
	r.title = title
	r.teamID = teamID
	r.limit = limit
	return append([]CoreSimilarStory(nil), r.similarStories...), nil
}

func newTestService(repo Repository) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "search-test"), repo)
}

func TestFindSimilarStoriesNormalizesInputAndCapsLimit(t *testing.T) {
	teamID := uuid.New()
	expected := []CoreSimilarStory{{
		ID:         uuid.New(),
		SequenceID: 42,
		Title:      "Add a Linear integration",
		Team:       teamID,
		Confidence: 0.91,
	}}
	repo := &repositoryStub{similarStories: expected}
	service := newTestService(repo)

	stories, err := service.FindSimilarStories(
		context.Background(),
		uuid.New(),
		uuid.New(),
		"  Add Linear integration  ",
		&teamID,
		100,
	)

	require.NoError(t, err)
	require.Equal(t, expected, stories)
	require.Equal(t, "Add Linear integration", repo.title)
	require.Equal(t, &teamID, repo.teamID)
	require.Equal(t, maxSimilarStoriesLimit, repo.limit)
}

func TestFindSimilarStoriesSkipsTitlesBelowTheSuggestionThreshold(t *testing.T) {
	repo := &repositoryStub{}
	service := newTestService(repo)

	stories, err := service.FindSimilarStories(
		context.Background(),
		uuid.New(),
		uuid.New(),
		"create",
		nil,
		0,
	)

	require.NoError(t, err)
	require.Empty(t, stories)
	require.Zero(t, repo.findCalls)
}

func TestSearchNormalizesDefaultsBeforeCallingRepositories(t *testing.T) {
	priority := "  high  "
	repo := &repositoryStub{}
	service := newTestService(repo)

	_, err := service.Search(context.Background(), uuid.New(), uuid.New(), SearchParams{
		Query:    "  roadmap  ",
		Priority: &priority,
	})

	require.NoError(t, err)
	require.Equal(t, 1, repo.storyCalls)
	require.Equal(t, 1, repo.objectiveCalls)
	require.Equal(t, "roadmap", repo.storyParams.Query)
	require.Equal(t, SearchTypeAll, repo.storyParams.Type)
	require.Equal(t, SortByRelevance, repo.storyParams.SortBy)
	require.Equal(t, defaultSearchPage, repo.storyParams.Page)
	require.Equal(t, defaultSearchPageSize, repo.storyParams.PageSize)
	require.Equal(t, "high", *repo.storyParams.Priority)
	require.Equal(t, repo.storyParams, repo.objectiveParams)
}

func TestSearchRejectsUnboundedOrUnsupportedParameters(t *testing.T) {
	tests := map[string]SearchParams{
		"query too long":      {Query: strings.Repeat("a", maxSearchQueryRunes+1)},
		"page too large":      {Page: maxSearchPage + 1},
		"page size too large": {PageSize: maxSearchPageSize + 1},
		"unknown type":        {Type: SearchType("documents")},
		"unknown sort":        {SortBy: SortOption("random")},
	}

	for name, params := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &repositoryStub{}
			_, err := newTestService(repo).Search(context.Background(), uuid.New(), uuid.New(), params)
			require.ErrorIs(t, err, ErrInvalidSearchParams)
			require.Zero(t, repo.storyCalls)
			require.Zero(t, repo.objectiveCalls)
		})
	}
}

func TestFindSimilarStoriesRejectsOversizedTitles(t *testing.T) {
	repo := &repositoryStub{}
	_, err := newTestService(repo).FindSimilarStories(
		context.Background(),
		uuid.New(),
		uuid.New(),
		strings.Repeat("a", maxSimilarityTitleRunes+1),
		nil,
		defaultSimilarStoriesLimit,
	)

	require.ErrorIs(t, err, ErrInvalidSearchParams)
	require.Zero(t, repo.findCalls)
}
