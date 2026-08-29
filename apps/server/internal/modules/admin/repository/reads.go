package adminrepository

import (
	"context"
	"errors"
	"fmt"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) GetAdminUser(
	ctx context.Context,
	actorID uuid.UUID,
) (admindomain.UserSummary, error) {
	var result admindomain.UserSummary
	err := repository.withActiveInternalAdmin(ctx, actorID, func(queries adminsql.Querier) error {
		row, err := queries.GetAdminUser(ctx, adminsql.GetAdminUserParams{UserID: actorID})
		if err != nil {
			return fmt.Errorf("get current admin: %w", err)
		}
		result, err = userFromGetRow(row)
		return err
	})
	return result, err
}

func (repository *Repository) GetDashboardSummary(
	ctx context.Context,
	query admindomain.DashboardSummaryQuery,
) (admindomain.DashboardSummary, error) {
	var result admindomain.DashboardSummary
	err := repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		row, err := queries.GetAdminDashboardSummary(ctx, adminsql.GetAdminDashboardSummaryParams{
			NowAt: query.Now,
		})
		if err != nil {
			return fmt.Errorf("get admin dashboard summary: %w", err)
		}
		result, err = dashboardFromRow(row)
		return err
	})
	return result, err
}

func (repository *Repository) ListWorkspaces(
	ctx context.Context,
	query admindomain.ListWorkspacesQuery,
) (admindomain.ListResult[admindomain.WorkspaceSummary], error) {
	if _, err := admindomain.ParseWorkspaceStatus(string(query.Status)); err != nil {
		return admindomain.ListResult[admindomain.WorkspaceSummary]{}, err
	}
	page, err := newSQLPage(query.Page)
	if err != nil {
		return admindomain.ListResult[admindomain.WorkspaceSummary]{}, err
	}

	var result admindomain.ListResult[admindomain.WorkspaceSummary]
	err = repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		rows, err := queries.ListAdminWorkspaces(ctx, adminsql.ListAdminWorkspacesParams{
			SearchText: query.Search, StatusFilter: string(query.Status), NowAt: query.Now,
			RowLimit: page.limit, RowOffset: page.offset,
		})
		if err != nil {
			return fmt.Errorf("list admin workspaces: %w", err)
		}
		items := make([]admindomain.WorkspaceSummary, 0, len(rows))
		for _, row := range rows {
			item, err := workspaceFromListRow(row)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		total := int64(0)
		if len(rows) > 0 {
			total = rows[0].TotalCount
		}
		pagination, err := paginationResult(query.Page, total)
		if err != nil {
			return err
		}
		result = admindomain.ListResult[admindomain.WorkspaceSummary]{
			Items: items, Pagination: pagination,
		}
		return nil
	})
	return result, err
}

func (repository *Repository) GetWorkspaceOverview(
	ctx context.Context,
	query admindomain.GetWorkspaceQuery,
) (admindomain.WorkspaceOverview, error) {
	var result admindomain.WorkspaceOverview
	err := repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		var err error
		result, err = getWorkspaceOverview(ctx, queries, query.WorkspaceID)
		return err
	})
	return result, err
}

func (repository *Repository) ListUsers(
	ctx context.Context,
	query admindomain.ListUsersQuery,
) (admindomain.ListResult[admindomain.UserSummary], error) {
	page, err := newSQLPage(query.Page)
	if err != nil {
		return admindomain.ListResult[admindomain.UserSummary]{}, err
	}

	var result admindomain.ListResult[admindomain.UserSummary]
	err = repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		rows, err := queries.ListAdminUsers(ctx, adminsql.ListAdminUsersParams{
			SearchText: query.Search, RowLimit: page.limit, RowOffset: page.offset,
		})
		if err != nil {
			return fmt.Errorf("list admin users: %w", err)
		}
		items := make([]admindomain.UserSummary, 0, len(rows))
		for _, row := range rows {
			item, err := userFromListRow(row)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		total := int64(0)
		if len(rows) > 0 {
			total = rows[0].TotalCount
		}
		pagination, err := paginationResult(query.Page, total)
		if err != nil {
			return err
		}
		result = admindomain.ListResult[admindomain.UserSummary]{
			Items: items, Pagination: pagination,
		}
		return nil
	})
	return result, err
}

func (repository *Repository) GetUserOverview(
	ctx context.Context,
	query admindomain.GetUserQuery,
) (admindomain.UserOverview, error) {
	var result admindomain.UserOverview
	err := repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		var err error
		result, err = getUserOverview(ctx, queries, query.UserID)
		return err
	})
	return result, err
}

func getWorkspaceOverview(
	ctx context.Context,
	queries adminsql.Querier,
	workspaceID uuid.UUID,
) (admindomain.WorkspaceOverview, error) {
	row, err := queries.GetAdminWorkspace(ctx, adminsql.GetAdminWorkspaceParams{WorkspaceID: workspaceID})
	if errors.Is(err, pgx.ErrNoRows) {
		return admindomain.WorkspaceOverview{}, admindomain.ErrNotFound
	}
	if err != nil {
		return admindomain.WorkspaceOverview{}, fmt.Errorf("get admin workspace: %w", err)
	}
	workspace, err := workspaceFromGetRow(row)
	if err != nil {
		return admindomain.WorkspaceOverview{}, err
	}
	members, err := queries.ListAdminWorkspaceMembers(ctx, adminsql.ListAdminWorkspaceMembersParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return admindomain.WorkspaceOverview{}, fmt.Errorf("list admin workspace members: %w", err)
	}
	mapped := make([]admindomain.WorkspaceMember, 0, len(members))
	for _, member := range members {
		mapped = append(mapped, admindomain.WorkspaceMember{
			UserID: member.UserID, Email: member.Email, FullName: stringValue(member.FullName),
			Role: member.Role, IsInternal: member.IsInternal, JoinedAt: member.JoinedAt,
		})
	}
	return admindomain.WorkspaceOverview{Workspace: workspace, Members: mapped}, nil
}

func getUserOverview(
	ctx context.Context,
	queries adminsql.Querier,
	userID uuid.UUID,
) (admindomain.UserOverview, error) {
	row, err := queries.GetAdminUser(ctx, adminsql.GetAdminUserParams{UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return admindomain.UserOverview{}, admindomain.ErrNotFound
	}
	if err != nil {
		return admindomain.UserOverview{}, fmt.Errorf("get admin user: %w", err)
	}
	user, err := userFromGetRow(row)
	if err != nil {
		return admindomain.UserOverview{}, err
	}
	memberships, err := queries.ListAdminUserMemberships(ctx, adminsql.ListAdminUserMembershipsParams{
		UserID: userID,
	})
	if err != nil {
		return admindomain.UserOverview{}, fmt.Errorf("list admin user memberships: %w", err)
	}
	mapped := make([]admindomain.UserMembership, 0, len(memberships))
	for _, membership := range memberships {
		mapped = append(mapped, admindomain.UserMembership{
			WorkspaceID: membership.WorkspaceID, WorkspaceName: membership.WorkspaceName,
			WorkspaceSlug: membership.WorkspaceSlug, Role: membership.Role, JoinedAt: membership.JoinedAt,
		})
	}
	return admindomain.UserOverview{User: user, Memberships: mapped}, nil
}
