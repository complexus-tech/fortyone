package searchrepository

import (
	"context"
	"errors"
	"fmt"

	searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"
	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

// FindSimilarStories ranks stories visible to the active actor against a
// proposed title. The generated query owns tenant and membership enforcement.
func (r *repo) FindSimilarStories(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	title string,
	teamID *uuid.UUID,
	limit int,
) ([]searchdomain.CoreSimilarStory, error) {
	if err := r.ready(); err != nil {
		return nil, err
	}
	if limit < 1 {
		return nil, errors.New("similar-story limit must be positive")
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("validate similar-story limit: %w", err)
	}

	rows, err := r.queries.FindSimilarStories(ctx, searchsql.FindSimilarStoriesParams{
		ResultLimit: resultLimit,
		Title:       title,
		ActorID:     actorID,
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("find similar stories: %w", err)
	}

	stories := make([]searchdomain.CoreSimilarStory, 0, len(rows))
	for _, row := range rows {
		story, err := toCoreSimilarStory(row)
		if err != nil {
			return nil, err
		}
		stories = append(stories, story)
	}
	return stories, nil
}
