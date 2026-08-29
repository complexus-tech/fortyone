package admin

import (
	"context"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
)

func (service *Service) GetCurrentAdmin(ctx context.Context, actorID uuid.UUID) (UserSummary, error) {
	ctx, span := adminTracer.Start(ctx, "admin.GetCurrentAdmin")
	defer span.End()

	user, err := service.repo.GetAdminUser(ctx, actorID)
	if err != nil {
		return UserSummary{}, err
	}
	service.resolveUserAvatar(ctx, &user)
	return user, nil
}

func (service *Service) GetDashboardSummary(ctx context.Context, actorID uuid.UUID) (DashboardSummary, error) {
	ctx, span := adminTracer.Start(ctx, "admin.GetDashboardSummary")
	defer span.End()
	return service.repo.GetDashboardSummary(ctx, admindomain.DashboardSummaryQuery{
		ActorID: actorID,
		Now:     service.clock.Now().UTC(),
	})
}

func (service *Service) ListWorkspaces(ctx context.Context, actorID uuid.UUID, input ListWorkspacesInput) (ListResult[WorkspaceSummary], error) {
	ctx, span := adminTracer.Start(ctx, "admin.ListWorkspaces")
	defer span.End()

	query, err := listWorkspacesQuery(actorID, input)
	if err != nil {
		return ListResult[WorkspaceSummary]{}, err
	}
	query.Now = service.clock.Now().UTC()
	result, err := service.repo.ListWorkspaces(ctx, query)
	if err != nil {
		return ListResult[WorkspaceSummary]{}, err
	}
	for index := range result.Items {
		service.resolveWorkspaceLogo(ctx, &result.Items[index])
	}
	return result, nil
}

func (service *Service) GetWorkspaceOverview(ctx context.Context, actorID, workspaceID uuid.UUID) (WorkspaceOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.GetWorkspaceOverview")
	defer span.End()

	overview, err := service.repo.GetWorkspaceOverview(ctx, admindomain.GetWorkspaceQuery{
		ActorID: actorID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return WorkspaceOverview{}, err
	}
	service.resolveWorkspaceLogo(ctx, &overview.Workspace)
	return overview, nil
}

func (service *Service) ListUsers(ctx context.Context, actorID uuid.UUID, input ListUsersInput) (ListResult[UserSummary], error) {
	ctx, span := adminTracer.Start(ctx, "admin.ListUsers")
	defer span.End()

	result, err := service.repo.ListUsers(ctx, admindomain.ListUsersQuery{
		ActorID: actorID,
		Page:    normalizePage(input.Pagination),
		Search:  admindomain.NormalizeSearch(input.Query),
	})
	if err != nil {
		return ListResult[UserSummary]{}, err
	}
	for index := range result.Items {
		service.resolveUserAvatar(ctx, &result.Items[index])
	}
	return result, nil
}

func (service *Service) GetUserOverview(ctx context.Context, actorID, userID uuid.UUID) (UserOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.GetUserOverview")
	defer span.End()

	overview, err := service.repo.GetUserOverview(ctx, admindomain.GetUserQuery{ActorID: actorID, UserID: userID})
	if err != nil {
		return UserOverview{}, err
	}
	service.resolveUserAvatar(ctx, &overview.User)
	return overview, nil
}

func (service *Service) ListAuditLogs(ctx context.Context, actorID uuid.UUID, input ListAuditLogsInput) (ListResult[AuditLog], error) {
	ctx, span := adminTracer.Start(ctx, "admin.ListAuditLogs")
	defer span.End()

	query, err := listAuditLogsQuery(actorID, input)
	if err != nil {
		return ListResult[AuditLog]{}, err
	}
	return service.repo.ListAuditLogs(ctx, query)
}

func (service *Service) ListAdminNotes(ctx context.Context, actorID uuid.UUID, input ListAdminNotesInput) (ListResult[AdminNote], error) {
	ctx, span := adminTracer.Start(ctx, "admin.ListAdminNotes")
	defer span.End()

	query, err := listAdminNotesQuery(actorID, input)
	if err != nil {
		return ListResult[AdminNote]{}, err
	}
	return service.repo.ListAdminNotes(ctx, query)
}
