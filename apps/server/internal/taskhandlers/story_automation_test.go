package taskhandlers

import (
	"context"
	"errors"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestStoryAutomationHandlersDelegateEveryPolicy(t *testing.T) {
	store := &storyAutomationHandlerStoreStub{}
	handlers := NewStoryAutomationHandlers(StoryAutomationHandlerDependencies{
		Log: testTaskLogger(), Store: store, SystemUserID: uuid.New(),
	})

	require.NoError(t, handlers.HandleStoryAutoArchive(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleStoryAutoClose(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleSprintStoryMigration(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.archiveCalls)
	require.Equal(t, 1, store.closeCalls)
	require.Equal(t, 1, store.migrationCalls)
}

func TestStoryAutomationHandlersPropagateStoreFailures(t *testing.T) {
	sentinel := errors.New("story automation store unavailable")
	store := &storyAutomationHandlerStoreStub{err: sentinel}
	handlers := NewStoryAutomationHandlers(StoryAutomationHandlerDependencies{
		Log: testTaskLogger(), Store: store, SystemUserID: uuid.New(),
	})

	require.ErrorIs(t, handlers.HandleStoryAutoArchive(t.Context(), asynq.NewTask("test", nil)), sentinel)
	require.ErrorIs(t, handlers.HandleStoryAutoClose(t.Context(), asynq.NewTask("test", nil)), sentinel)
	require.ErrorIs(t, handlers.HandleSprintStoryMigration(t.Context(), asynq.NewTask("test", nil)), sentinel)
}

func TestStoryAutomationHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *StoryAutomationHandlers
	require.ErrorContains(t, handlers.HandleStoryAutoArchive(t.Context(), nil), "dependencies are required")

	handlers = NewStoryAutomationHandlers(StoryAutomationHandlerDependencies{Log: testTaskLogger()})
	require.ErrorContains(t, handlers.HandleStoryAutoArchive(t.Context(), nil), "store is required")

	handlers = NewStoryAutomationHandlers(StoryAutomationHandlerDependencies{
		Log: testTaskLogger(), Store: &storyAutomationHandlerStoreStub{},
	})
	require.ErrorContains(t, handlers.HandleStoryAutoClose(t.Context(), nil), "system user is required")
}

type storyAutomationHandlerStoreStub struct {
	archiveCalls   int
	closeCalls     int
	migrationCalls int
	err            error
}

func (store *storyAutomationHandlerStoreStub) ArchiveEligibleStoriesBatch(
	context.Context,
	storydomain.StoryAutoArchiveBatch,
) (storydomain.StoryAutoArchiveResult, error) {
	store.archiveCalls++
	return storydomain.StoryAutoArchiveResult{}, store.err
}

func (store *storyAutomationHandlerStoreStub) CloseEligibleStoriesBatch(
	context.Context,
	storydomain.StoryAutoCloseBatch,
) (storydomain.StoryAutoCloseResult, error) {
	store.closeCalls++
	return storydomain.StoryAutoCloseResult{}, store.err
}

func (store *storyAutomationHandlerStoreStub) MigrateEligibleSprintStoriesBatch(
	context.Context,
	storydomain.SprintStoryMigrationBatch,
) (storydomain.SprintStoryMigrationResult, error) {
	store.migrationCalls++
	return storydomain.SprintStoryMigrationResult{}, store.err
}
