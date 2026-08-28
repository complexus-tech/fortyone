package workspacesrepository

import (
	"context"
	"errors"
	"fmt"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) List(ctx context.Context, userID uuid.UUID) ([]workspacedomain.Workspace, error) {
	rows, err := r.queries.ListWorkspacesForUser(ctx, workspacesql.ListWorkspacesForUserParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("list workspaces for user: %w", err)
	}

	result := make([]workspacedomain.Workspace, len(rows))
	for index, row := range rows {
		result[index] = workspaceFromMembershipRecord(membershipWorkspaceRecord{
			workspaceRecord: workspaceRecordFromListRow(row),
			isActive:        row.IsActive,
			userRole:        row.UserRole,
		})
	}
	return result, nil
}

func (r *repo) Get(ctx context.Context, workspaceID, userID uuid.UUID) (workspacedomain.Workspace, error) {
	row, err := r.queries.GetWorkspaceForMember(ctx, workspacesql.GetWorkspaceForMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return workspacedomain.Workspace{}, mapWorkspaceNotFound("get workspace for member", err)
	}
	return workspaceFromMembershipRecord(membershipWorkspaceRecord{
		workspaceRecord: workspaceRecordFromMemberRow(row),
		isActive:        row.IsActive,
		userRole:        row.UserRole,
	}), nil
}

func (r *repo) GetBySlug(ctx context.Context, slug string, userID uuid.UUID) (workspacedomain.Workspace, error) {
	row, err := r.queries.GetWorkspaceForMemberBySlug(ctx, workspacesql.GetWorkspaceForMemberBySlugParams{
		UserID: userID,
		Slug:   slug,
	})
	if err != nil {
		return workspacedomain.Workspace{}, mapWorkspaceNotFound("get workspace for member by slug", err)
	}
	return workspaceFromMembershipRecord(membershipWorkspaceRecord{
		workspaceRecord: workspaceRecordFromMemberSlugRow(row),
		isActive:        row.IsActive,
		userRole:        row.UserRole,
	}), nil
}

func (r *repo) GetByID(ctx context.Context, workspaceID uuid.UUID) (workspacedomain.Workspace, error) {
	row, err := r.queries.GetWorkspaceByID(ctx, workspacesql.GetWorkspaceByIDParams{WorkspaceID: workspaceID})
	if err != nil {
		return workspacedomain.Workspace{}, mapWorkspaceNotFound("get workspace by id", err)
	}
	return workspaceFromRecord(workspaceRecordFromIDRow(row)), nil
}

func (r *repo) GetPublicBySlug(ctx context.Context, slug string) (workspacedomain.Workspace, error) {
	row, err := r.queries.GetPublicWorkspaceBySlug(ctx, workspacesql.GetPublicWorkspaceBySlugParams{Slug: slug})
	if err != nil {
		return workspacedomain.Workspace{}, mapWorkspaceNotFound("get public workspace by slug", err)
	}
	return workspaceFromRecord(workspaceRecordFromPublicRow(row)), nil
}

func (r *repo) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	exists, err := r.queries.WorkspaceSlugExists(ctx, workspacesql.WorkspaceSlugExistsParams{Slug: slug})
	if err != nil {
		return false, fmt.Errorf("check workspace slug availability: %w", err)
	}
	return !exists, nil
}

func mapWorkspaceNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return workspacedomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
