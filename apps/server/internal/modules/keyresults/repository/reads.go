package keyresultsrepository

import (
	"context"
	"fmt"
	"math"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresultssql "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

func (repository *Repository) Get(
	ctx context.Context,
	query keyresultsdomain.GetQuery,
) (keyresultsdomain.KeyResult, error) {
	if err := query.Validate(); err != nil {
		return keyresultsdomain.KeyResult{}, err
	}
	if err := repository.configured(); err != nil {
		return keyresultsdomain.KeyResult{}, err
	}
	row, err := repository.queries.GetKeyResult(ctx, keyresultssql.GetKeyResultParams{
		ActorID: query.Access.ActorID, KeyResultID: query.KeyResultID,
		WorkspaceID: uuidPointer(query.Access.WorkspaceID), AllTeams: query.Access.AllTeams,
		AllowedTeamIds: query.Access.TeamIDs,
	})
	if err != nil {
		return keyresultsdomain.KeyResult{}, fmt.Errorf("get key result: %w", mapDatabaseError(err))
	}
	return keyResultFromGetRow(row), nil
}

func (repository *Repository) List(
	ctx context.Context,
	query keyresultsdomain.ObjectiveListQuery,
) ([]keyresultsdomain.KeyResult, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListObjectiveKeyResults(ctx, keyresultssql.ListObjectiveKeyResultsParams{
		ActorID: query.Access.ActorID, ObjectiveID: query.ObjectiveID,
		WorkspaceID: uuidPointer(query.Access.WorkspaceID), AllTeams: query.Access.AllTeams,
		AllowedTeamIds: query.Access.TeamIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("list objective key results: %w", mapDatabaseError(err))
	}
	results := make([]keyresultsdomain.KeyResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, keyResultFromObjectiveListRow(row))
	}
	return results, nil
}

func (repository *Repository) ListPaginated(
	ctx context.Context,
	query keyresultsdomain.PaginatedListQuery,
) (keyresultsdomain.ListResponse, error) {
	normalized, err := query.Normalize()
	if err != nil {
		return keyresultsdomain.ListResponse{}, err
	}
	if err := repository.configured(); err != nil {
		return keyresultsdomain.ListResponse{}, err
	}
	filters := normalized.Filters
	offset, err := checkedOffset(filters.Page, filters.PageSize)
	if err != nil {
		return keyresultsdomain.ListResponse{}, err
	}
	limit, err := safecast.Int32(filters.PageSize)
	if err != nil {
		return keyresultsdomain.ListResponse{}, fmt.Errorf("%w: page size: %v", keyresultsdomain.ErrInvalid, err)
	}
	params := keyresultssql.CountKeyResultsParams{
		ActorID: normalized.Access.ActorID, WorkspaceID: uuidPointer(normalized.Access.WorkspaceID),
		AllTeams: normalized.Access.AllTeams, AllowedTeamIds: normalized.Access.TeamIDs,
		FilterTeamIds: filters.TeamIDs, ObjectiveIds: filters.ObjectiveIDs,
		MeasurementTypes: filters.MeasurementTypes, LeadIds: filters.LeadIDs,
		CreatedByIds: filters.CreatedBy, CreatedAfter: filters.CreatedAfter,
		CreatedBefore: filters.CreatedBefore, UpdatedAfter: filters.UpdatedAfter,
		UpdatedBefore: filters.UpdatedBefore, EndDateAfter: filters.EndDateAfter,
		EndDateBefore: filters.EndDateBefore,
	}
	totalCount, err := repository.queries.CountKeyResults(ctx, params)
	if err != nil {
		return keyresultsdomain.ListResponse{}, fmt.Errorf("count key results: %w", mapDatabaseError(err))
	}
	total, err := safecast.Int64(totalCount)
	if err != nil {
		return keyresultsdomain.ListResponse{}, fmt.Errorf("%w: result count: %v", keyresultsdomain.ErrInvalid, err)
	}
	rows, err := repository.queries.ListKeyResults(ctx, keyresultssql.ListKeyResultsParams{
		ActorID: params.ActorID, WorkspaceID: params.WorkspaceID, AllTeams: params.AllTeams,
		AllowedTeamIds: params.AllowedTeamIds, FilterTeamIds: params.FilterTeamIds,
		ObjectiveIds: params.ObjectiveIds, MeasurementTypes: params.MeasurementTypes,
		LeadIds: params.LeadIds, CreatedByIds: params.CreatedByIds,
		CreatedAfter: params.CreatedAfter, CreatedBefore: params.CreatedBefore,
		UpdatedAfter: params.UpdatedAfter, UpdatedBefore: params.UpdatedBefore,
		EndDateAfter: params.EndDateAfter, EndDateBefore: params.EndDateBefore,
		SortKey: normalized.SortKey(), ResultOffset: offset, ResultLimit: limit,
	})
	if err != nil {
		return keyresultsdomain.ListResponse{}, fmt.Errorf("list key results: %w", mapDatabaseError(err))
	}
	results := make([]keyresultsdomain.KeyResultWithObjective, 0, len(rows))
	for _, row := range rows {
		results = append(results, keyResultWithObjectiveFromRow(row))
	}
	consumed := int64(offset) + int64(len(results))
	return keyresultsdomain.ListResponse{
		KeyResults: results, TotalCount: total, Page: filters.Page,
		PageSize: filters.PageSize, HasMore: consumed < totalCount,
	}, nil
}

func checkedOffset(page, pageSize int) (int32, error) {
	if page < 1 || pageSize < 1 || page-1 > math.MaxInt/pageSize {
		return 0, fmt.Errorf("%w: pagination offset is outside the supported range", keyresultsdomain.ErrInvalid)
	}
	offset := (page - 1) * pageSize
	result, err := safecast.Int32(offset)
	if err != nil {
		return 0, fmt.Errorf("%w: pagination offset: %v", keyresultsdomain.ErrInvalid, err)
	}
	return result, nil
}
