package adminrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) UpdateWorkspaceTrial(
	ctx context.Context,
	command admindomain.UpdateWorkspaceTrialCommand,
) (admindomain.WorkspaceOverview, error) {
	reason, err := admindomain.RequireReason(command.Reason)
	if err != nil {
		return admindomain.WorkspaceOverview{}, err
	}
	command.Reason = reason
	var result admindomain.WorkspaceOverview
	err = repository.withActiveInternalAdmin(ctx, command.ActorID, func(queries adminsql.Querier) error {
		current, err := lockWorkspace(ctx, queries, command.WorkspaceID)
		if err != nil {
			return err
		}
		now := command.Now.UTC()
		trialEndsOn := command.TrialEndsOn.UTC()
		if !trialEndsOn.After(now) {
			return admindomain.ErrInvalidTrialEndsOn
		}
		if current.TrialEndsOn != nil && current.TrialEndsOn.After(now) &&
			!trialEndsOn.After(current.TrialEndsOn.UTC()) {
			return admindomain.ErrInvalidTrialEndsOn
		}

		updated, err := queries.UpdateAdminWorkspaceTrial(ctx, adminsql.UpdateAdminWorkspaceTrialParams{
			NewTrialEndsOn: trialEndsOn, ChangedAt: nextTimestamp(now, current.UpdatedAt),
			WorkspaceID: command.WorkspaceID, ExpectedUpdatedAt: current.UpdatedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrConflict
		}
		if err != nil {
			return fmt.Errorf("update admin workspace trial: %w", err)
		}
		targetID := command.WorkspaceID
		if _, err := insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetWorkspace,
			TargetID: &targetID, WorkspaceID: &targetID,
			Action: admindomain.AuditWorkspaceTrialUpdated, FieldName: "trial_ends_on",
			OldValue: current.TrialEndsOn, NewValue: updated.TrialEndsOn, Reason: command.Reason,
			Metadata: map[string]any{"workspace_name": current.Name, "workspace_slug": current.Slug},
		}); err != nil {
			return err
		}
		result, err = getWorkspaceOverview(ctx, queries, command.WorkspaceID)
		return err
	})
	return result, err
}

func (repository *Repository) SetWorkspaceDeleted(
	ctx context.Context,
	command admindomain.SetWorkspaceDeletedCommand,
) (admindomain.WorkspaceOverview, error) {
	reason, err := admindomain.RequireReason(command.Reason)
	if err != nil {
		return admindomain.WorkspaceOverview{}, err
	}
	command.Reason = reason
	var result admindomain.WorkspaceOverview
	err = repository.withActiveInternalAdmin(ctx, command.ActorID, func(queries adminsql.Querier) error {
		current, err := lockWorkspace(ctx, queries, command.WorkspaceID)
		if err != nil {
			return err
		}
		updated, err := queries.SetAdminWorkspaceDeleted(ctx, adminsql.SetAdminWorkspaceDeletedParams{
			DeleteRequested: command.Deleted, ChangedAt: nextTimestamp(command.Now.UTC(), current.UpdatedAt),
			ActorUserID: command.ActorID, WorkspaceID: command.WorkspaceID,
			ExpectedUpdatedAt: current.UpdatedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrConflict
		}
		if err != nil {
			return fmt.Errorf("set admin workspace deleted state: %w", err)
		}

		action := admindomain.AuditWorkspaceRestored
		if command.Deleted {
			action = admindomain.AuditWorkspaceDeleted
		}
		targetID := command.WorkspaceID
		if _, err := insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetWorkspace,
			TargetID: &targetID, WorkspaceID: &targetID, Action: action,
			FieldName: "deleted_at", OldValue: current.DeletedAt, NewValue: updated.DeletedAt,
			Reason:   command.Reason,
			Metadata: map[string]any{"workspace_name": current.Name, "workspace_slug": current.Slug},
		}); err != nil {
			return err
		}
		result, err = getWorkspaceOverview(ctx, queries, command.WorkspaceID)
		return err
	})
	return result, err
}

func (repository *Repository) UpdateUserState(
	ctx context.Context,
	command admindomain.UpdateUserStateCommand,
) (admindomain.UserOverview, error) {
	reason, err := admindomain.RequireReason(command.Reason)
	if err != nil {
		return admindomain.UserOverview{}, err
	}
	command.Reason = reason
	active, activeSet := command.Patch.IsActive.Value()
	internal, internalSet := command.Patch.IsInternal.Value()
	if (!activeSet && !internalSet) || (activeSet && active == nil) || (internalSet && internal == nil) {
		return admindomain.UserOverview{}, admindomain.ErrInvalidAction
	}

	var result admindomain.UserOverview
	err = repository.withinTransaction(ctx, func(queries adminsql.Querier) error {
		participants, err := queries.LockAdminUserMutationParticipants(
			ctx,
			adminsql.LockAdminUserMutationParticipantsParams{
				ActorID: command.ActorID, TargetUserID: command.UserID,
			},
		)
		if err != nil {
			return fmt.Errorf("lock admin user mutation participants: %w", err)
		}
		if err := validateUserMutationParticipants(participants, command.ActorID, command.UserID); err != nil {
			return err
		}

		current, err := getUserOverview(ctx, queries, command.UserID)
		if err != nil {
			return err
		}
		params := adminsql.UpdateAdminUserStateParams{
			IsActiveSet: activeSet, NewIsActive: current.User.IsActive,
			IsInternalSet: internalSet, NewIsInternal: current.User.IsInternal,
			ChangedAt: nextTimestamp(command.Now.UTC(), current.User.UpdatedAt),
			UserID:    command.UserID, ExpectedUpdatedAt: current.User.UpdatedAt,
		}
		if active != nil {
			params.NewIsActive = *active
		}
		if internal != nil {
			params.NewIsInternal = *internal
		}
		if _, err := queries.UpdateAdminUserState(ctx, params); errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrConflict
		} else if err != nil {
			return fmt.Errorf("update admin user state: %w", err)
		}
		if err := insertUserStateAudits(ctx, queries, command, current.User, params); err != nil {
			return err
		}
		result, err = getUserOverview(ctx, queries, command.UserID)
		return err
	})
	return result, mapDatabaseError(err)
}

func (repository *Repository) RequestSessionRevocation(
	ctx context.Context,
	command admindomain.RequestSessionRevocationCommand,
) (admindomain.UserOverview, error) {
	reason, err := admindomain.RequireReason(command.Reason)
	if err != nil {
		return admindomain.UserOverview{}, err
	}
	command.Reason = reason
	var result admindomain.UserOverview
	err = repository.withActiveInternalAdmin(ctx, command.ActorID, func(queries adminsql.Querier) error {
		if _, err := queries.LockUserTarget(ctx, adminsql.LockUserTargetParams{UserID: command.UserID}); errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrNotFound
		} else if err != nil {
			return fmt.Errorf("lock session revocation target: %w", err)
		}
		var err error
		result, err = getUserOverview(ctx, queries, command.UserID)
		if err != nil {
			return err
		}
		if _, err := queries.RevokeAdminUserBrowserSessions(ctx, adminsql.RevokeAdminUserBrowserSessionsParams{
			UserID: command.UserID,
		}); err != nil {
			return fmt.Errorf("revoke admin user browser sessions: %w", err)
		}
		targetID := command.UserID
		_, err = insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetUser, TargetID: &targetID,
			Action: admindomain.AuditUserSessionRevocationRequested, FieldName: "auth_session",
			OldValue: "active", NewValue: "revocation_requested", Reason: command.Reason,
			Metadata: map[string]any{"user_email": result.User.Email, "user_name": result.User.FullName},
		})
		return err
	})
	return result, err
}

func (repository *Repository) CreateAdminNote(
	ctx context.Context,
	command admindomain.CreateAdminNoteCommand,
) (admindomain.AdminNote, error) {
	command.Body = strings.TrimSpace(command.Body)
	if command.Body == "" {
		return admindomain.AdminNote{}, admindomain.ErrInvalidNote
	}

	var result admindomain.AdminNote
	err := repository.withActiveInternalAdmin(ctx, command.ActorID, func(queries adminsql.Querier) error {
		if err := lockNoteTarget(ctx, queries, command); err != nil {
			return err
		}
		created, err := queries.CreateAdminNote(ctx, adminsql.CreateAdminNoteParams{
			TargetType: string(command.TargetType), TargetID: command.TargetID,
			WorkspaceID: command.WorkspaceID, Body: command.Body, CreatedByUserID: command.ActorID,
		})
		if err != nil {
			return fmt.Errorf("create admin note: %w", err)
		}
		creator, err := queries.GetAdminUser(ctx, adminsql.GetAdminUserParams{UserID: command.ActorID})
		if err != nil {
			return fmt.Errorf("load admin note creator: %w", err)
		}
		result = admindomain.AdminNote{
			ID: created.ID, TargetType: created.TargetType, TargetID: created.TargetID,
			WorkspaceID: created.WorkspaceID, Body: created.Body,
			CreatedByUserID: created.CreatedByUserID, CreatedByName: stringValue(creator.FullName),
			CreatedByEmail: creator.Email, CreatedAt: created.CreatedAt,
		}
		targetID := command.TargetID
		_, err = insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.ActorID, TargetType: command.TargetType, TargetID: &targetID,
			WorkspaceID: command.WorkspaceID, Action: admindomain.AuditAdminNoteCreated,
			FieldName: "note", NewValue: command.Body, Reason: "Admin note added",
			Metadata: map[string]any{"note_id": created.ID},
		})
		return err
	})
	return result, err
}

func lockWorkspace(
	ctx context.Context,
	queries adminsql.Querier,
	workspaceID uuid.UUID,
) (adminsql.LockAdminWorkspaceMutationTargetRow, error) {
	row, err := queries.LockAdminWorkspaceMutationTarget(ctx, adminsql.LockAdminWorkspaceMutationTargetParams{
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return adminsql.LockAdminWorkspaceMutationTargetRow{}, admindomain.ErrNotFound
	}
	if err != nil {
		return adminsql.LockAdminWorkspaceMutationTargetRow{}, fmt.Errorf("lock admin workspace: %w", err)
	}
	return row, nil
}

func nextTimestamp(requested, current time.Time) time.Time {
	if requested.After(current) {
		return requested
	}
	return current.Add(time.Microsecond)
}
