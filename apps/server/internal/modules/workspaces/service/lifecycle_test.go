package workspaces

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceDeletionRequiresSuccessfulCancellation(t *testing.T) {
	t.Parallel()

	providerErr := errors.New("billing provider unavailable")
	databaseErr := errors.New("workspace database unavailable")
	deletedAt := time.Now()
	tests := []struct {
		name             string
		lookupErr        error
		deletedAt        *time.Time
		cancellationErr  error
		deleteErr        error
		wantErr          error
		wantOperations   []string
		wantNotification bool
	}{
		{
			name:            "provider failure leaves workspace available",
			cancellationErr: providerErr, wantErr: providerErr,
			wantOperations: []string{"get", "cancel"},
		},
		{
			name:           "successful cancellation precedes deletion",
			wantOperations: []string{"get", "cancel", "delete", "get"}, wantNotification: true,
		},
		{
			name:      "deletion failure after cancellation is surfaced",
			deleteErr: databaseErr, wantErr: databaseErr,
			wantOperations: []string{"get", "cancel", "delete"},
		},
		{
			name:      "missing workspace does not reach billing",
			lookupErr: ErrNotFound, wantErr: ErrNotFound,
			wantOperations: []string{"get"},
		},
		{
			name:      "workspace lookup failure does not reach billing",
			lookupErr: databaseErr, wantErr: databaseErr,
			wantOperations: []string{"get"},
		},
		{
			name:      "already deleted workspace does not reach billing",
			deletedAt: &deletedAt, wantErr: ErrNotFound,
			wantOperations: []string{"get"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspaceID, actorID := uuid.New(), uuid.New()
			var operations []string
			repository := &lifecycleRepositoryStub{
				workspace: CoreWorkspace{ID: workspaceID, Name: "Workspace", Slug: "workspace", DeletedAt: test.deletedAt},
				actorID:   actorID, lookupErr: test.lookupErr, deleteErr: test.deleteErr, operations: &operations,
			}
			publisher := &lifecyclePublisherStub{}
			service := &Service{
				repo: repository, log: logger.NewWithText(io.Discard, slog.LevelError, "workspace-lifecycle-test"),
				subscriptions: lifecycleSubscriptionStub{operations: &operations, err: test.cancellationErr},
				publisher:     publisher, users: lifecycleUserStub{},
			}

			err := service.Delete(t.Context(), workspaceID, actorID)

			require.ErrorIs(t, err, test.wantErr)
			require.Equal(t, test.wantOperations, operations)
			if test.cancellationErr != nil {
				require.ErrorContains(t, err, "cancel workspace subscription before deletion")
			}
			if test.wantNotification {
				require.NotNil(t, repository.workspace.DeletedAt)
				require.Len(t, publisher.events, 2)
				require.Equal(t, events.WorkspaceDeletionScheduledConfirmation, publisher.events[0].Type)
				require.Equal(t, events.WorkspaceDeletionScheduledNotification, publisher.events[1].Type)
			} else {
				require.Equal(t, test.deletedAt, repository.workspace.DeletedAt)
				require.Empty(t, publisher.events)
			}
		})
	}
}

func TestWorkspaceDeletionCanBeRetriedAfterBillingFailure(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	var operations []string
	repository := &lifecycleRepositoryStub{
		workspace: CoreWorkspace{ID: workspaceID}, actorID: actorID, operations: &operations,
	}
	service := &Service{
		repo: repository, log: logger.NewWithText(io.Discard, slog.LevelError, "workspace-lifecycle-test"),
		subscriptions: lifecycleSubscriptionStub{operations: &operations, err: errors.New("temporary billing failure")},
		publisher:     &lifecyclePublisherStub{}, users: lifecycleUserStub{},
	}
	require.Error(t, service.Delete(t.Context(), workspaceID, actorID))
	require.Nil(t, repository.workspace.DeletedAt)

	service.subscriptions = lifecycleSubscriptionStub{operations: &operations}
	require.NoError(t, service.Delete(t.Context(), workspaceID, actorID))
	require.NotNil(t, repository.workspace.DeletedAt)
}

func TestWorkspaceRestoreKeepsExistingBillingBehavior(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	deletedAt := time.Now()
	var operations []string
	repository := &lifecycleRepositoryStub{
		workspace: CoreWorkspace{ID: workspaceID, DeletedAt: &deletedAt}, actorID: actorID, operations: &operations,
	}
	publisher := &lifecyclePublisherStub{}
	service := &Service{
		repo: repository, log: logger.NewWithText(io.Discard, slog.LevelError, "workspace-lifecycle-test"),
		publisher: publisher, users: lifecycleUserStub{},
	}

	require.NoError(t, service.Restore(t.Context(), workspaceID, actorID))
	require.Nil(t, repository.workspace.DeletedAt)
	require.Equal(t, []string{"restore", "get"}, operations)
	require.Len(t, publisher.events, 2)
	require.Equal(t, events.WorkspaceRestoredConfirmation, publisher.events[0].Type)
	require.Equal(t, events.WorkspaceRestoredNotification, publisher.events[1].Type)
}

type lifecycleRepositoryStub struct {
	Repository
	workspace  CoreWorkspace
	actorID    uuid.UUID
	lookupErr  error
	deleteErr  error
	operations *[]string
}

func (stub *lifecycleRepositoryStub) Get(_ context.Context, workspaceID, actorID uuid.UUID) (CoreWorkspace, error) {
	*stub.operations = append(*stub.operations, "get")
	if workspaceID != stub.workspace.ID || actorID != stub.actorID {
		return CoreWorkspace{}, ErrNotFound
	}
	return stub.workspace, stub.lookupErr
}

func (stub *lifecycleRepositoryStub) Delete(_ context.Context, workspaceID, actorID uuid.UUID) error {
	*stub.operations = append(*stub.operations, "delete")
	if workspaceID != stub.workspace.ID || actorID != stub.actorID {
		return ErrNotFound
	}
	if stub.deleteErr == nil {
		now := time.Now()
		stub.workspace.DeletedAt = &now
	}
	return stub.deleteErr
}

func (stub *lifecycleRepositoryStub) Restore(_ context.Context, workspaceID, actorID uuid.UUID) error {
	*stub.operations = append(*stub.operations, "restore")
	if workspaceID != stub.workspace.ID || actorID != stub.actorID {
		return ErrNotFound
	}
	stub.workspace.DeletedAt = nil
	return nil
}

func (*lifecycleRepositoryStub) GetWorkspaceAdminEmails(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return nil, nil
}

type lifecycleSubscriptionStub struct {
	SubscriptionManager
	operations *[]string
	err        error
}

func (stub lifecycleSubscriptionStub) CancelWorkspaceSubscription(context.Context, uuid.UUID) error {
	*stub.operations = append(*stub.operations, "cancel")
	return stub.err
}

type lifecycleUserStub struct{ UserDirectory }

func (lifecycleUserStub) GetWorkspaceUser(context.Context, uuid.UUID) (WorkspaceUser, error) {
	return WorkspaceUser{Email: "admin@example.com", FullName: "Workspace Admin"}, nil
}

type lifecyclePublisherStub struct{ events []events.Event }

func (stub *lifecyclePublisherStub) Publish(_ context.Context, event events.Event) error {
	stub.events = append(stub.events, event)
	return nil
}
