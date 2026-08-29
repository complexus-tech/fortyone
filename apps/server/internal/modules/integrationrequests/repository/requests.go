package integrationrequestsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) UpsertPending(ctx context.Context, input integrationrequestdomain.UpsertRequestInput) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("encode integration request metadata: %w", err)
	}
	sourceNumber, err := optionalInt32(input.SourceNumber)
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("convert integration request source number: %w", err)
	}
	estimatedDuration, err := optionalInt32(input.EstimatedDurationMinutes)
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("convert estimated duration: %w", err)
	}
	minimumFocusBlock, err := optionalInt32(input.MinimumFocusBlockMinutes)
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("convert minimum focus block: %w", err)
	}
	row, err := r.queries.UpsertPendingIntegrationRequest(ctx, integrationrequestssql.UpsertPendingIntegrationRequestParams{
		WorkspaceID: input.WorkspaceID, TeamID: input.TeamID, Provider: input.Provider,
		SourceType: input.SourceType, SourceExternalID: input.SourceExternalID, SourceNumber: sourceNumber,
		SourceURL: input.SourceURL, Title: input.Title, Description: input.Description,
		StatusID: input.StatusID, Priority: input.Priority, AssigneeID: input.AssigneeID,
		EstimateUnit: input.EstimateValue, EstimatedDurationMinutes: estimatedDuration,
		MinimumFocusBlockMinutes: minimumFocusBlock, ObjectiveID: input.ObjectiveID,
		KeyResultID: input.KeyResultID, SprintID: input.SprintID, StartDate: input.StartDate,
		EndDate: input.EndDate, LabelIds: cleanUUIDs(input.LabelIDs), Metadata: metadata,
		CreatedByUserID: input.CreatedByUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return r.GetByExternal(ctx, input.WorkspaceID, input.Provider, input.SourceType, input.SourceExternalID)
	}
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("upsert pending integration request: %w", err)
	}
	return integrationRequestFromSQL(row)
}

func (r *Repo) AuthorizeTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	_, err := r.queries.AuthorizeIntegrationRequestTeam(ctx, integrationrequestssql.AuthorizeIntegrationRequestTeamParams{
		WorkspaceID: workspaceID, TeamID: teamID, ActorID: userID,
	})
	if err != nil {
		return mapNotFound("authorize integration request team", err)
	}
	return nil
}

func (r *Repo) ListByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter integrationrequestdomain.ListRequestsFilter) ([]integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	params, err := listParams(workspaceID, teamID, userID, filter)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListIntegrationRequestsByTeam(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list integration requests by team: %w", err)
	}
	result := make([]integrationrequestdomain.IntegrationRequest, 0, len(rows))
	for _, row := range rows {
		request, mapErr := integrationRequestFromSQL(row)
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, request)
	}
	return result, nil
}

func (r *Repo) CountByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter integrationrequestdomain.ListRequestsFilter) (int, error) {
	if err := r.configured(); err != nil {
		return 0, err
	}
	list, err := listParams(workspaceID, teamID, userID, filter)
	if err != nil {
		return 0, err
	}
	count, err := r.queries.CountIntegrationRequestsByTeam(ctx, integrationrequestssql.CountIntegrationRequestsByTeamParams{
		WorkspaceID: list.WorkspaceID, TeamID: list.TeamID, RequestStatus: list.RequestStatus,
		ActorID: list.ActorID, HasSearch: list.HasSearch, SearchPattern: list.SearchPattern,
		HasProvider: list.HasProvider, Provider: list.Provider, HasPriority: list.HasPriority,
		Priority: list.Priority, HasAssignee: list.HasAssignee, AssigneeID: list.AssigneeID,
		HasCreatedAfter: list.HasCreatedAfter, CreatedAfter: list.CreatedAfter,
		HasCreatedBefore: list.HasCreatedBefore, CreatedBefore: list.CreatedBefore,
	})
	if err != nil {
		return 0, fmt.Errorf("count integration requests by team: %w", err)
	}
	converted, err := safecast.Int64(count)
	if err != nil {
		return 0, fmt.Errorf("convert integration request count: %w", err)
	}
	return converted, nil
}

func (r *Repo) GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	row, err := r.queries.GetIntegrationRequestForUser(ctx, integrationrequestssql.GetIntegrationRequestForUserParams{
		WorkspaceID: workspaceID, RequestID: requestID, ActorID: userID,
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, mapNotFound("get integration request for user", err)
	}
	return integrationRequestFromSQL(row)
}

// Get is reserved for trusted provider callbacks without a FortyOne actor.
// User-facing reads must use GetForUser.
func (r *Repo) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	row, err := r.queries.GetIntegrationRequest(ctx, integrationrequestssql.GetIntegrationRequestParams{
		WorkspaceID: workspaceID, RequestID: requestID,
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, mapNotFound("get trusted integration request", err)
	}
	return integrationRequestFromSQL(row)
}

func (r *Repo) GetByExternal(ctx context.Context, workspaceID uuid.UUID, provider, sourceType, sourceExternalID string) (integrationrequestdomain.IntegrationRequest, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.IntegrationRequest{}, err
	}
	row, err := r.queries.GetIntegrationRequestByExternal(ctx, integrationrequestssql.GetIntegrationRequestByExternalParams{
		WorkspaceID: workspaceID, Provider: provider, SourceType: sourceType, SourceExternalID: sourceExternalID,
	})
	if err != nil {
		return integrationrequestdomain.IntegrationRequest{}, mapNotFound("get integration request by external key", err)
	}
	return integrationRequestFromSQL(row)
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(category)
	statusID, err := r.queries.FindFirstIntegrationRequestStatusByCategory(ctx, integrationrequestssql.FindFirstIntegrationRequestStatusByCategoryParams{
		TeamID: teamID, Category: &trimmed,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find first integration request status by category: %w", err)
	}
	return &statusID, nil
}

func listParams(workspaceID, teamID, actorID uuid.UUID, filter integrationrequestdomain.ListRequestsFilter) (integrationrequestssql.ListIntegrationRequestsByTeamParams, error) {
	status := strings.TrimSpace(filter.Status)
	if status == "" {
		status = integrationrequestdomain.StatusPending
	}
	search := strings.TrimSpace(filter.Search)
	params := integrationrequestssql.ListIntegrationRequestsByTeamParams{
		WorkspaceID: workspaceID, TeamID: teamID, RequestStatus: status, ActorID: actorID,
		HasSearch: search != "", SearchPattern: "%" + search + "%",
		HasProvider: filter.Provider != "", Provider: filter.Provider,
		HasPriority: filter.Priority != "", Priority: filter.Priority,
		HasAssignee: filter.AssigneeID != nil, AssigneeID: filter.AssigneeID,
		HasCreatedAfter: filter.CreatedAfter != nil, CreatedAfter: filter.CreatedAfter,
		HasCreatedBefore: filter.CreatedBefore != nil, CreatedBefore: filter.CreatedBefore,
		Paginated: filter.PageSize > 0,
	}
	if !params.Paginated {
		return params, nil
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	rowLimit, err := safecast.Int32(filter.PageSize)
	if err != nil {
		return integrationrequestssql.ListIntegrationRequestsByTeamParams{}, fmt.Errorf("convert integration request page size: %w", err)
	}
	offset, err := safecast.Int32((page - 1) * filter.PageSize)
	if err != nil {
		return integrationrequestssql.ListIntegrationRequestsByTeamParams{}, fmt.Errorf("convert integration request page offset: %w", err)
	}
	params.RowLimit = rowLimit
	params.RowOffset = offset
	return params, nil
}

func optionalInt32(value *int) (*int32, error) {
	if value == nil {
		return nil, nil
	}
	converted, err := safecast.Int32(*value)
	if err != nil {
		return nil, err
	}
	return &converted, nil
}

func cleanUUIDs(values []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			result = append(result, value)
		}
	}
	return result
}
