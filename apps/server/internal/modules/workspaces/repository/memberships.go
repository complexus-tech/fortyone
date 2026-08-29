package workspacesrepository

import (
	"context"
	"fmt"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
)

func (r *repo) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	return addMember(ctx, r.queries, workspaceID, userID, role)
}

func addMember(
	ctx context.Context,
	queries workspacesql.Querier,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role string,
) error {
	typedRole := workspacesql.UserRole(role)
	if !typedRole.Valid() {
		return fmt.Errorf("invalid workspace role %q", role)
	}
	rows, err := queries.AddWorkspaceMember(ctx, workspacesql.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        typedRole,
	})
	if err != nil {
		if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
			return workspacedomain.ErrAlreadyWorkspaceMember
		}
		return fmt.Errorf("add workspace member: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("add workspace member: affected %d rows, want 1", rows)
	}
	return nil
}

func (r *repo) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return r.withinTransaction(ctx, func(queries workspacesql.Querier) error {
		rows, err := queries.RemoveWorkspaceMember(ctx, workspacesql.RemoveWorkspaceMemberParams{
			WorkspaceID: workspaceID,
			UserID:      userID,
		})
		if err != nil {
			return fmt.Errorf("remove workspace member: %w", err)
		}
		if rows == 0 {
			return workspacedomain.ErrMemberNotFound
		}
		if err := queries.RemoveWorkspaceTeamMemberships(ctx, workspacesql.RemoveWorkspaceTeamMembershipsParams{
			UserID:      userID,
			WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("remove workspace team memberships: %w", err)
		}
		return nil
	})
}

func (r *repo) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	typedRole := workspacesql.UserRole(role)
	if !typedRole.Valid() {
		return fmt.Errorf("invalid workspace role %q", role)
	}
	rows, err := r.queries.UpdateWorkspaceMemberRole(ctx, workspacesql.UpdateWorkspaceMemberRoleParams{
		Role:        typedRole,
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("update workspace member role: %w", err)
	}
	if rows == 0 {
		return workspacedomain.ErrMemberNotFound
	}
	return nil
}

func (r *repo) GetWorkspaceAdminEmails(ctx context.Context, workspaceID, actorID uuid.UUID) ([]string, error) {
	emails, err := r.queries.ListWorkspaceAdminEmails(ctx, workspacesql.ListWorkspaceAdminEmailsParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace admin emails: %w", err)
	}
	return emails, nil
}

func (r *repo) ResolveCurrentMembership(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (workspacedomain.CurrentMembership, error) {
	row, err := r.queries.ResolveCurrentWorkspaceMembership(
		ctx,
		workspacesql.ResolveCurrentWorkspaceMembershipParams{Slug: slug, UserID: userID},
	)
	if err != nil {
		return workspacedomain.CurrentMembership{}, mapWorkspaceNotFound("resolve current workspace membership", err)
	}
	return workspacedomain.CurrentMembership{
		WorkspaceID: row.WorkspaceID,
		Name:        row.Name,
		Slug:        row.Slug,
		Role:        row.UserRole,
	}, nil
}

func (r *repo) RecordAccess(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if err := r.queries.RecordWorkspaceAccess(ctx, workspacesql.RecordWorkspaceAccessParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		return fmt.Errorf("record workspace access: %w", err)
	}
	return nil
}
