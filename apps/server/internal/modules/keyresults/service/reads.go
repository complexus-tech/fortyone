package keyresults

import (
	"context"
	"fmt"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) Get(ctx context.Context, id, workspaceID uuid.UUID) (CoreKeyResult, error) {
	return service.GetForActor(ctx, id, workspaceID, uuid.Nil)
}

func (service *Service) GetForActor(
	ctx context.Context,
	id, workspaceID, actorID uuid.UUID,
) (CoreKeyResult, error) {
	access, err := service.accessFor(ctx, workspaceID, actorID, platformauth.ScopeObjectivesRead)
	if err != nil {
		return CoreKeyResult{}, err
	}
	query := keyresultsdomain.GetQuery{Access: access, KeyResultID: id}
	if err := query.Validate(); err != nil {
		return CoreKeyResult{}, err
	}
	result, err := service.repo.Get(ctx, query)
	if err != nil {
		return CoreKeyResult{}, err
	}
	return result, nil
}

func (service *Service) List(
	ctx context.Context,
	objectiveID, workspaceID uuid.UUID,
) ([]CoreKeyResult, error) {
	access, err := service.accessFor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesRead)
	if err != nil {
		return nil, err
	}
	query := keyresultsdomain.ObjectiveListQuery{
		Access: access, ObjectiveID: objectiveID,
	}
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return service.repo.List(ctx, query)
}

func (service *Service) ListPaginated(
	ctx context.Context,
	filters CoreKeyResultFilters,
) (CoreKeyResultListResponse, error) {
	access, err := service.accessFor(
		ctx, filters.WorkspaceID, filters.CurrentUserID, platformauth.ScopeObjectivesRead,
	)
	if err != nil {
		return CoreKeyResultListResponse{}, err
	}
	query, err := (keyresultsdomain.PaginatedListQuery{
		Access: access, Filters: filters,
	}).Normalize()
	if err != nil {
		return CoreKeyResultListResponse{}, err
	}
	response, err := service.repo.ListPaginated(ctx, query)
	if err != nil {
		return CoreKeyResultListResponse{}, fmt.Errorf("list key results: %w", err)
	}
	results := make([]CoreKeyResultWithObjective, 0, len(response.KeyResults))
	for _, value := range response.KeyResults {
		results = append(results, CoreKeyResultWithObjective{
			CoreKeyResult: value.KeyResult, ObjectiveName: value.ObjectiveName,
			ObjectiveID: value.ObjectiveID, TeamID: value.TeamID, TeamName: value.TeamName,
			TeamCode: value.TeamCode, WorkspaceID: value.WorkspaceID,
		})
	}
	return CoreKeyResultListResponse{
		KeyResults: results, TotalCount: response.TotalCount, Page: response.Page,
		PageSize: response.PageSize, HasMore: response.HasMore,
	}, nil
}
