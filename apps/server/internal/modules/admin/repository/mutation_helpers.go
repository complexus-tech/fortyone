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

func validateUserMutationParticipants(
	participants []adminsql.LockAdminUserMutationParticipantsRow,
	actorID, targetID uuid.UUID,
) error {
	var actor *adminsql.LockAdminUserMutationParticipantsRow
	targetFound := false
	for index := range participants {
		participant := &participants[index]
		if participant.UserID == actorID {
			actor = participant
		}
		if participant.UserID == targetID {
			targetFound = true
		}
	}
	if actor == nil || !actor.IsActive || !actor.IsInternal {
		return admindomain.ErrForbidden
	}
	if actorID == targetID {
		return admindomain.ErrSelfMutation
	}
	if !targetFound {
		return admindomain.ErrNotFound
	}
	return nil
}

func insertUserStateAudits(
	ctx context.Context,
	queries adminsql.Querier,
	command admindomain.UpdateUserStateCommand,
	current admindomain.UserSummary,
	updated adminsql.UpdateAdminUserStateParams,
) error {
	targetID := current.ID
	metadata := map[string]any{"user_email": current.Email, "user_name": current.FullName}
	entries := make([]auditEntry, 0, 3)
	if updated.IsActiveSet && updated.NewIsActive != current.IsActive {
		action := admindomain.AuditUserActivated
		if !updated.NewIsActive {
			action = admindomain.AuditUserDeactivated
		}
		entries = append(entries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetUser, TargetID: &targetID,
			Action: action, FieldName: "is_active", OldValue: current.IsActive,
			NewValue: updated.NewIsActive, Reason: command.Reason, Metadata: metadata,
		})
	}
	nextPolicy := current.LoginReactivationPolicy
	if updated.IsActiveSet {
		nextPolicy = admindomain.LoginReactivationAdminOnly
		if updated.NewIsActive {
			nextPolicy = admindomain.LoginReactivationVerifiedSignIn
		}
	}
	if nextPolicy != current.LoginReactivationPolicy {
		entries = append(entries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetUser, TargetID: &targetID,
			Action:    admindomain.AuditUserReactivationPolicyChanged,
			FieldName: "login_reactivation_policy", OldValue: current.LoginReactivationPolicy,
			NewValue: nextPolicy, Reason: command.Reason, Metadata: metadata,
		})
	}
	if updated.IsInternalSet && updated.NewIsInternal != current.IsInternal {
		action := admindomain.AuditUserInternalGranted
		if !updated.NewIsInternal {
			action = admindomain.AuditUserInternalRevoked
		}
		entries = append(entries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetUser, TargetID: &targetID,
			Action: action, FieldName: "is_internal", OldValue: current.IsInternal,
			NewValue: updated.NewIsInternal, Reason: command.Reason, Metadata: metadata,
		})
	}
	if len(entries) == 0 {
		entries = append(entries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetUser, TargetID: &targetID,
			Action: admindomain.AuditUserStateReviewed, FieldName: "state",
			Reason: command.Reason, Metadata: metadata,
		})
	}
	for _, entry := range entries {
		if _, err := insertAuditLog(ctx, queries, entry); err != nil {
			return err
		}
	}
	return nil
}

func lockNoteTarget(
	ctx context.Context,
	queries adminsql.Querier,
	command admindomain.CreateAdminNoteCommand,
) error {
	var err error
	switch command.TargetType {
	case admindomain.TargetWorkspace:
		if command.WorkspaceID == nil || *command.WorkspaceID != command.TargetID {
			return admindomain.ErrInvalidAction
		}
		_, err = queries.LockAdminNoteWorkspaceTarget(ctx, adminsql.LockAdminNoteWorkspaceTargetParams{
			WorkspaceID: command.TargetID,
		})
	case admindomain.TargetUser:
		if command.WorkspaceID == nil {
			_, err = queries.LockAdminNoteUserTarget(ctx, adminsql.LockAdminNoteUserTargetParams{
				UserID: command.TargetID,
			})
		} else {
			_, err = queries.LockAdminNoteUserWorkspaceTarget(
				ctx,
				adminsql.LockAdminNoteUserWorkspaceTargetParams{
					WorkspaceID: *command.WorkspaceID, UserID: command.TargetID,
				},
			)
		}
	default:
		return admindomain.ErrInvalidAction
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return admindomain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock admin note target: %w", err)
	}
	return nil
}
