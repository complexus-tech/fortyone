package searchrepository

import (
	"context"
	"fmt"

	searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"
	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/google/uuid"
)

// SearchObjectives returns objectives visible to the active workspace and team
// member. All optional filters remain generated, typed SQL parameters.
func (r *repo) SearchObjectives(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	params searchdomain.SearchParams,
) ([]searchdomain.CoreSearchObjective, int, error) {
	if err := r.ready(); err != nil {
		return nil, 0, err
	}
	pageOffset, pageLimit, err := databasePage(params.Page, params.PageSize)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.SearchObjectives(ctx, searchsql.SearchObjectivesParams{
		ActorID:     actorID,
		WorkspaceID: workspaceID,
		QueryText:   params.Query,
		TeamID:      params.TeamID,
		StatusID:    params.StatusID,
		SortBy:      string(params.SortBy),
		PageOffset:  pageOffset,
		PageLimit:   pageLimit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search objectives: %w", err)
	}
	if len(rows) == 0 {
		total, err := r.queries.CountSearchObjectives(ctx, searchsql.CountSearchObjectivesParams{
			ActorID:     actorID,
			WorkspaceID: workspaceID,
			QueryText:   params.Query,
			TeamID:      params.TeamID,
			StatusID:    params.StatusID,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("count empty objective search page: %w", err)
		}
		return []searchdomain.CoreSearchObjective{}, boundedCount(total), nil
	}

	objectives := make([]searchdomain.CoreSearchObjective, 0, len(rows))
	for _, row := range rows {
		objective, err := toCoreSearchObjective(row)
		if err != nil {
			return nil, 0, err
		}
		objectives = append(objectives, objective)
	}
	return objectives, boundedCount(rows[0].TotalCount), nil
}
