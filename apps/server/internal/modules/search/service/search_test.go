package search

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	similarStories []CoreSimilarStory
	findCalls      int
	title          string
	teamID         *uuid.UUID
	limit          int
}

func (r *repositoryStub) SearchStories(context.Context, uuid.UUID, uuid.UUID, SearchParams) ([]CoreSearchStory, int, error) {
	return nil, 0, nil
}

func (r *repositoryStub) SearchObjectives(context.Context, uuid.UUID, uuid.UUID, SearchParams) ([]CoreSearchObjective, int, error) {
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
