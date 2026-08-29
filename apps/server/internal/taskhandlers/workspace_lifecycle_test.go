package taskhandlers

import (
	"context"
	"errors"
	"testing"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceLifecycleHandlersDelegateToTypedStore(t *testing.T) {
	store := &workspaceLifecycleStoreStub{}
	handlers := newTestWorkspaceLifecycleHandlers(store)

	require.NoError(t, handlers.HandleWorkspaceInactivityWarning(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleWorkspaceDeletion(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleWorkspaceCleanup(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.warningListCalls)
	require.Equal(t, 1, store.deletionCalls)
	require.Equal(t, 1, store.cleanupCalls)
}

func TestWorkspaceLifecycleHandlersPreserveStoreFailures(t *testing.T) {
	sentinel := errors.New("workspace lifecycle store unavailable")
	tests := []struct {
		name      string
		configure func(*workspaceLifecycleStoreStub)
		handle    func(*WorkspaceLifecycleHandlers) error
	}{
		{
			name:      "warning",
			configure: func(store *workspaceLifecycleStoreStub) { store.warningListErr = sentinel },
			handle: func(handlers *WorkspaceLifecycleHandlers) error {
				return handlers.HandleWorkspaceInactivityWarning(context.Background(), asynq.NewTask("test", nil))
			},
		},
		{
			name:      "deletion",
			configure: func(store *workspaceLifecycleStoreStub) { store.deletionErr = sentinel },
			handle: func(handlers *WorkspaceLifecycleHandlers) error {
				return handlers.HandleWorkspaceDeletion(context.Background(), asynq.NewTask("test", nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &workspaceLifecycleStoreStub{}
			test.configure(store)
			err := test.handle(newTestWorkspaceLifecycleHandlers(store))
			require.ErrorIs(t, err, sentinel)
		})
	}
}

func TestWorkspaceLifecycleHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *WorkspaceLifecycleHandlers
	err := handlers.HandleWorkspaceDeletion(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewWorkspaceLifecycleHandlers(WorkspaceLifecycleHandlerDependencies{
		Log:    testTaskLogger(),
		Mailer: guidanceMailerStub{},
	})
	err = handlers.HandleWorkspaceDeletion(t.Context(), nil)
	require.ErrorContains(t, err, "inactive workspace deletion store is required")
}

func newTestWorkspaceLifecycleHandlers(
	store *workspaceLifecycleStoreStub,
) *WorkspaceLifecycleHandlers {
	return NewWorkspaceLifecycleHandlers(WorkspaceLifecycleHandlerDependencies{
		Log:    testTaskLogger(),
		Store:  store,
		Mailer: guidanceMailerStub{},
	})
}

type workspaceLifecycleStoreStub struct {
	warningListCalls int
	deletionCalls    int
	cleanupCalls     int
	warningListErr   error
	deletionErr      error
}

func (store *workspaceLifecycleStoreStub) ListWorkspaceInactivityWarningCandidates(
	context.Context,
	workspacedomain.InactivityWarningQuery,
) ([]workspacedomain.InactivityWarningCandidate, error) {
	store.warningListCalls++
	return nil, store.warningListErr
}

func (*workspaceLifecycleStoreStub) RecordWorkspaceInactivityWarning(
	context.Context,
	workspacedomain.InactivityWarningReceipt,
) error {
	return nil
}

func (store *workspaceLifecycleStoreStub) DeleteInactiveWorkspacesBatch(
	context.Context,
	workspacedomain.InactivityDeletionBatch,
) (workspacedomain.InactivityDeletionResult, error) {
	store.deletionCalls++
	return workspacedomain.InactivityDeletionResult{}, store.deletionErr
}

func (store *workspaceLifecycleStoreStub) PurgeSoftDeletedWorkspacesBatch(
	context.Context,
	workspacedomain.DeletedWorkspacePurgeBatch,
) (workspacedomain.DeletedWorkspacePurgeResult, error) {
	store.cleanupCalls++
	return workspacedomain.DeletedWorkspacePurgeResult{}, nil
}
