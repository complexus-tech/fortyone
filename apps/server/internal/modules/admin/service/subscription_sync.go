package admin

import (
	"context"
	"errors"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
)

func (service *Service) RequestWorkspaceSubscriptionSync(
	ctx context.Context,
	actorID, workspaceID uuid.UUID,
	input RequestWorkspaceSubscriptionSyncInput,
) (WorkspaceOverview, error) {
	ctx, span := adminTracer.Start(ctx, "admin.RequestWorkspaceSubscriptionSync")
	defer span.End()

	reason, err := admindomain.RequireReason(input.Reason)
	if err != nil {
		return WorkspaceOverview{}, err
	}
	attempt, before, err := service.repo.BeginSubscriptionSync(ctx, admindomain.BeginSubscriptionSyncCommand{
		ActorID: actorID, WorkspaceID: workspaceID, Reason: reason,
	})
	if err != nil {
		return WorkspaceOverview{}, err
	}

	if service.subscriptionSyncer == nil {
		_, finishErr := service.repo.FinishSubscriptionSync(ctx, admindomain.FinishSubscriptionSyncCommand{
			Attempt: attempt, Outcome: admindomain.SubscriptionSyncFailed,
		})
		return before, errors.Join(ErrIntegrationUnavailable, finishErr)
	}

	syncErr := service.subscriptionSyncer.SyncSubscription(ctx, workspaceID)
	outcome := admindomain.SubscriptionSyncSucceeded
	if syncErr != nil {
		outcome = admindomain.SubscriptionSyncFailed
	}
	after, finishErr := service.repo.FinishSubscriptionSync(ctx, admindomain.FinishSubscriptionSyncCommand{
		Attempt: attempt, Outcome: outcome,
	})
	if syncErr != nil || finishErr != nil {
		return before, errors.Join(syncErr, finishErr)
	}
	service.resolveWorkspaceLogo(ctx, &after.Workspace)
	return after, nil
}
