package admin

import (
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/google/uuid"
)

func normalizePage(input PaginationInput) pagination.OffsetParams {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	return pagination.OffsetParams{Page: page, PageSize: limit}
}

func listWorkspacesQuery(actorID uuid.UUID, input ListWorkspacesInput) (admindomain.ListWorkspacesQuery, error) {
	status, err := admindomain.ParseWorkspaceStatus(input.Status)
	if err != nil {
		return admindomain.ListWorkspacesQuery{}, err
	}
	return admindomain.ListWorkspacesQuery{
		ActorID: actorID, Page: normalizePage(input.Pagination),
		Search: admindomain.NormalizeSearch(input.Query), Status: status,
	}, nil
}

func listWorkspaceIntegrationsQuery(
	actorID uuid.UUID,
	input ListWorkspaceIntegrationsInput,
) (admindomain.ListWorkspaceIntegrationsQuery, error) {
	provider, err := admindomain.ParseIntegrationProvider(input.Provider)
	if err != nil {
		return admindomain.ListWorkspaceIntegrationsQuery{}, err
	}
	status, err := admindomain.ParseIntegrationStatus(input.Status)
	if err != nil {
		return admindomain.ListWorkspaceIntegrationsQuery{}, err
	}
	return admindomain.ListWorkspaceIntegrationsQuery{
		ActorID:  actorID,
		Page:     normalizePage(input.Pagination),
		Search:   admindomain.NormalizeSearch(input.Query),
		Provider: provider,
		Status:   status,
	}, nil
}

func listAuditLogsQuery(actorID uuid.UUID, input ListAuditLogsInput) (admindomain.ListAuditLogsQuery, error) {
	targetType, err := admindomain.ParseTargetType(input.TargetType)
	if err != nil {
		return admindomain.ListAuditLogsQuery{}, err
	}
	action, err := admindomain.ParseAuditAction(input.Action)
	if err != nil {
		return admindomain.ListAuditLogsQuery{}, err
	}
	return admindomain.ListAuditLogsQuery{
		ActorID: actorID, Page: normalizePage(input.Pagination), WorkspaceID: input.WorkspaceID,
		TargetType: targetType, Search: admindomain.NormalizeSearch(input.Query),
		Action: action, ActorSearch: admindomain.NormalizeSearch(input.ActorQuery),
		From: utcTime(input.From), To: utcTime(input.To),
	}, nil
}

func listAdminNotesQuery(actorID uuid.UUID, input ListAdminNotesInput) (admindomain.ListAdminNotesQuery, error) {
	targetType, err := admindomain.ParseTargetType(input.TargetType)
	if err != nil {
		return admindomain.ListAdminNotesQuery{}, err
	}
	if targetType != admindomain.TargetAny && targetType != admindomain.TargetWorkspace && targetType != admindomain.TargetUser {
		return admindomain.ListAdminNotesQuery{}, admindomain.ErrInvalidFilter
	}
	return admindomain.ListAdminNotesQuery{
		ActorID: actorID, Page: normalizePage(input.Pagination), TargetType: targetType,
		TargetID: input.TargetID, WorkspaceID: input.WorkspaceID,
	}, nil
}

func utcTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
