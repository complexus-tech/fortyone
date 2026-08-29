package taskhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestStoryRetentionHandlersDelegateBothRetentionPhases(t *testing.T) {
	store := &storyRetentionStoreStub{}
	handlers := NewStoryRetentionHandlers(StoryRetentionHandlerDependencies{
		Log:     testTaskLogger(),
		Store:   store,
		Objects: retainedObjectStoreStub{},
	})

	require.NoError(t, handlers.HandleDeletedStories(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleAttachmentObjectDeletions(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.purgeStoryCalls)
	require.Equal(t, 1, store.claimObjectCalls)
	require.Equal(t, 1, store.purgeCompletedCalls)
}

func TestStoryRetentionHandlersPreserveStoreFailures(t *testing.T) {
	sentinel := errors.New("story retention store unavailable")
	store := &storyRetentionStoreStub{purgeStoryErr: sentinel}
	handlers := NewStoryRetentionHandlers(StoryRetentionHandlerDependencies{
		Log:     testTaskLogger(),
		Store:   store,
		Objects: retainedObjectStoreStub{},
	})

	err := handlers.HandleDeletedStories(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)

	store.purgeStoryErr = nil
	store.claimObjectErr = sentinel
	err = handlers.HandleAttachmentObjectDeletions(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)
}

func TestStoryRetentionHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *StoryRetentionHandlers
	err := handlers.HandleDeletedStories(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewStoryRetentionHandlers(StoryRetentionHandlerDependencies{Log: testTaskLogger()})
	err = handlers.HandleDeletedStories(t.Context(), nil)
	require.ErrorContains(t, err, "deleted story retention store is required")
}

type storyRetentionStoreStub struct {
	purgeStoryCalls     int
	claimObjectCalls    int
	purgeCompletedCalls int
	purgeStoryErr       error
	claimObjectErr      error
}

func (store *storyRetentionStoreStub) PurgeDeletedStoriesBatch(
	context.Context,
	storydomain.StoryRetentionBatch,
) (storydomain.StoryRetentionResult, error) {
	store.purgeStoryCalls++
	return storydomain.StoryRetentionResult{}, store.purgeStoryErr
}

func (store *storyRetentionStoreStub) ClaimAttachmentObjectDeletions(
	context.Context,
	storydomain.AttachmentObjectDeletionClaimBatch,
) ([]storydomain.AttachmentObjectDeletion, error) {
	store.claimObjectCalls++
	return nil, store.claimObjectErr
}

func (*storyRetentionStoreStub) CompleteAttachmentObjectDeletion(
	context.Context,
	storydomain.AttachmentObjectDeletionCompletion,
) (bool, error) {
	return true, nil
}

func (*storyRetentionStoreStub) FailAttachmentObjectDeletion(
	context.Context,
	storydomain.AttachmentObjectDeletionFailure,
) (bool, error) {
	return true, nil
}

func (store *storyRetentionStoreStub) PurgeCompletedAttachmentObjectDeletions(
	context.Context,
	time.Time,
	int,
) (int64, error) {
	store.purgeCompletedCalls++
	return 0, nil
}

type retainedObjectStoreStub struct{}

func (retainedObjectStoreStub) RetainedObjectStorage() (string, string, error) {
	return "aws", "attachments", nil
}

func (retainedObjectStoreStub) DeleteRetainedObject(context.Context, string, string, string) error {
	return nil
}
