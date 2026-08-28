package admin

import (
	"context"
	"errors"
	"testing"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionSyncRecordsIntentThenSuccess(t *testing.T) {
	actorID, workspaceID := uuid.New(), uuid.New()
	attemptID := uuid.New()
	repository := &adminTestRepository{
		workspace:   admindomain.WorkspaceOverview{Workspace: admindomain.WorkspaceSummary{ID: workspaceID}},
		syncAttempt: admindomain.SubscriptionSyncAttempt{AuditID: attemptID, ActorID: actorID, WorkspaceID: workspaceID},
	}
	syncer := &adminTestSubscriptionSyncer{}
	service := New(repository, WithSubscriptionSyncer(syncer))

	_, err := service.RequestWorkspaceSubscriptionSync(context.Background(), actorID, workspaceID, RequestWorkspaceSubscriptionSyncInput{Reason: " billing support "})

	require.NoError(t, err)
	require.Equal(t, workspaceID, syncer.workspaceID)
	require.Equal(t, "billing support", repository.beginSyncCommand.Reason)
	require.Len(t, repository.finishSyncCommands, 1)
	require.Equal(t, admindomain.SubscriptionSyncSucceeded, repository.finishSyncCommands[0].Outcome)
	require.Equal(t, attemptID, repository.finishSyncCommands[0].Attempt.AuditID)
}

func TestSubscriptionSyncRecordsFailureAndReturnsProviderError(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	repository := &adminTestRepository{syncAttempt: admindomain.SubscriptionSyncAttempt{AuditID: uuid.New()}}
	service := New(repository, WithSubscriptionSyncer(&adminTestSubscriptionSyncer{err: providerErr}))

	_, err := service.RequestWorkspaceSubscriptionSync(context.Background(), uuid.New(), uuid.New(), RequestWorkspaceSubscriptionSyncInput{Reason: "retry"})

	require.ErrorIs(t, err, providerErr)
	require.Len(t, repository.finishSyncCommands, 1)
	require.Equal(t, admindomain.SubscriptionSyncFailed, repository.finishSyncCommands[0].Outcome)
}

func TestSubscriptionSyncUnavailableIsAuditedAsFailure(t *testing.T) {
	repository := &adminTestRepository{syncAttempt: admindomain.SubscriptionSyncAttempt{AuditID: uuid.New()}}
	service := New(repository)

	_, err := service.RequestWorkspaceSubscriptionSync(context.Background(), uuid.New(), uuid.New(), RequestWorkspaceSubscriptionSyncInput{Reason: "retry"})

	require.ErrorIs(t, err, ErrIntegrationUnavailable)
	require.Len(t, repository.finishSyncCommands, 1)
	require.Equal(t, admindomain.SubscriptionSyncFailed, repository.finishSyncCommands[0].Outcome)
}
