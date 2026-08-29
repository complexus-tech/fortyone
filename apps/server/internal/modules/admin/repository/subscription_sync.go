package adminrepository

import (
	"context"
	"fmt"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
)

func (repository *Repository) BeginSubscriptionSync(
	ctx context.Context,
	command admindomain.BeginSubscriptionSyncCommand,
) (admindomain.SubscriptionSyncAttempt, admindomain.WorkspaceOverview, error) {
	reason, err := admindomain.RequireReason(command.Reason)
	if err != nil {
		return admindomain.SubscriptionSyncAttempt{}, admindomain.WorkspaceOverview{}, err
	}
	command.Reason = reason

	var attempt admindomain.SubscriptionSyncAttempt
	var overview admindomain.WorkspaceOverview
	err = repository.withActiveInternalAdmin(ctx, command.ActorID, func(queries adminsql.Querier) error {
		if _, err := lockWorkspace(ctx, queries, command.WorkspaceID); err != nil {
			return err
		}
		var err error
		overview, err = getWorkspaceOverview(ctx, queries, command.WorkspaceID)
		if err != nil {
			return err
		}
		targetID := command.WorkspaceID
		inserted, err := insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.ActorID, TargetType: admindomain.TargetSubscription,
			TargetID: &targetID, WorkspaceID: &targetID,
			Action:    admindomain.AuditSubscriptionSyncRequested,
			FieldName: "subscription_status", OldValue: pointerValue(overview.Workspace.SubscriptionStatus),
			NewValue: "sync_requested", Reason: command.Reason,
			Metadata: subscriptionAuditMetadata(overview.Workspace, map[string]any{"phase": "intent"}),
		})
		if err != nil {
			return err
		}
		attempt = admindomain.SubscriptionSyncAttempt{
			AuditID: inserted.ID, ActorID: command.ActorID, WorkspaceID: command.WorkspaceID,
			Reason: command.Reason, BeforeStatus: overview.Workspace.SubscriptionStatus,
		}
		return nil
	})
	return attempt, overview, err
}

func (repository *Repository) FinishSubscriptionSync(
	ctx context.Context,
	command admindomain.FinishSubscriptionSyncCommand,
) (admindomain.WorkspaceOverview, error) {
	if command.Outcome != admindomain.SubscriptionSyncSucceeded &&
		command.Outcome != admindomain.SubscriptionSyncFailed {
		return admindomain.WorkspaceOverview{}, admindomain.ErrInvalidAction
	}

	var overview admindomain.WorkspaceOverview
	err := repository.withActiveInternalAdmin(ctx, command.Attempt.ActorID, func(queries adminsql.Querier) error {
		if _, err := lockWorkspace(ctx, queries, command.Attempt.WorkspaceID); err != nil {
			return err
		}
		var err error
		overview, err = getWorkspaceOverview(ctx, queries, command.Attempt.WorkspaceID)
		if err != nil {
			return err
		}
		action := admindomain.AuditSubscriptionSynced
		if command.Outcome == admindomain.SubscriptionSyncFailed {
			action = admindomain.AuditSubscriptionSyncFailed
		}
		targetID := command.Attempt.WorkspaceID
		_, err = insertAuditLog(ctx, queries, auditEntry{
			ActorID: command.Attempt.ActorID, TargetType: admindomain.TargetSubscription,
			TargetID: &targetID, WorkspaceID: &targetID, Action: action,
			FieldName: "subscription_status", OldValue: pointerValue(command.Attempt.BeforeStatus),
			NewValue: pointerValue(overview.Workspace.SubscriptionStatus), Reason: command.Attempt.Reason,
			Metadata: subscriptionAuditMetadata(overview.Workspace, map[string]any{
				"phase": "result", "outcome": string(command.Outcome),
				"request_audit_id": command.Attempt.AuditID,
			}),
		})
		return err
	})
	if err != nil {
		return admindomain.WorkspaceOverview{}, fmt.Errorf("finish subscription sync audit: %w", err)
	}
	return overview, nil
}

func subscriptionAuditMetadata(
	workspace admindomain.WorkspaceSummary,
	extra map[string]any,
) map[string]any {
	metadata := map[string]any{
		"workspace_name": workspace.Name, "workspace_slug": workspace.Slug,
		"stripe_customer_id":      pointerValue(workspace.StripeCustomerID),
		"stripe_subscription_id":  pointerValue(workspace.StripeSubscriptionID),
		"subscription_tier":       pointerValue(workspace.SubscriptionTier),
		"subscription_seat_count": workspace.SubscriptionSeats,
	}
	for key, value := range extra {
		metadata[key] = value
	}
	return metadata
}

func pointerValue[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
