package searchrepository

import (
	"context"
	"fmt"

	searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"
	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/google/uuid"
)

// SearchStories returns one typed, page-bounded result set. COUNT(*) OVER()
// keeps rows and their pagination metadata on the same database snapshot.
func (r *repo) SearchStories(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	params searchdomain.SearchParams,
) ([]searchdomain.CoreSearchStory, int, error) {
	if err := r.ready(); err != nil {
		return nil, 0, err
	}
	pageOffset, pageLimit, err := databasePage(params.Page, params.PageSize)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.SearchStories(ctx, searchsql.SearchStoriesParams{
		ActorID:     actorID,
		WorkspaceID: workspaceID,
		QueryText:   params.Query,
		TeamID:      params.TeamID,
		AssigneeID:  params.AssigneeID,
		StatusID:    params.StatusID,
		Priority:    params.Priority,
		LabelID:     params.LabelID,
		SortBy:      string(params.SortBy),
		PageOffset:  pageOffset,
		PageLimit:   pageLimit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search stories: %w", err)
	}
	if len(rows) == 0 {
		total, err := r.queries.CountSearchStories(ctx, searchsql.CountSearchStoriesParams{
			ActorID:     actorID,
			WorkspaceID: workspaceID,
			QueryText:   params.Query,
			TeamID:      params.TeamID,
			AssigneeID:  params.AssigneeID,
			StatusID:    params.StatusID,
			Priority:    params.Priority,
			LabelID:     params.LabelID,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("count empty story search page: %w", err)
		}
		return []searchdomain.CoreSearchStory{}, boundedCount(total), nil
	}

	stories := make([]searchdomain.CoreSearchStory, 0, len(rows))
	for _, row := range rows {
		story, err := toCoreSearchStory(row)
		if err != nil {
			return nil, 0, err
		}
		stories = append(stories, story)
	}
	return stories, boundedCount(rows[0].TotalCount), nil
}
