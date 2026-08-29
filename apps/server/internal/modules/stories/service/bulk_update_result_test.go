package stories

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestExecuteBulkStoryUpdatesReturnsOrderedMixedResult(t *testing.T) {
	storyIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	expectedError := errors.New("story is no longer editable")

	result := executeBulkStoryUpdates(
		context.Background(),
		storyIDs,
		func(_ context.Context, storyID uuid.UUID) error {
			if storyID == storyIDs[1] {
				return expectedError
			}
			return nil
		},
	)

	require.Equal(t, 3, result.TotalCount)
	require.Equal(t, 2, result.SucceededCount)
	require.Equal(t, 1, result.FailedCount)
	require.True(t, result.Partial)
	require.Equal(t, storyIDs, []uuid.UUID{
		result.Items[0].StoryID,
		result.Items[1].StoryID,
		result.Items[2].StoryID,
	})
	require.True(t, result.Items[0].Success)
	require.False(t, result.Items[1].Success)
	require.Equal(t, expectedError.Error(), result.Items[1].Error)
	require.True(t, result.Items[2].Success)
}

func TestExecuteBulkStoryUpdatesReportsEveryItemAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	storyIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	var updateCalls atomic.Int32

	result := executeBulkStoryUpdates(
		ctx,
		storyIDs,
		func(context.Context, uuid.UUID) error {
			updateCalls.Add(1)
			return nil
		},
	)

	require.Zero(t, updateCalls.Load())
	require.Equal(t, len(storyIDs), result.TotalCount)
	require.Zero(t, result.SucceededCount)
	require.Equal(t, len(storyIDs), result.FailedCount)
	require.False(t, result.Partial)
	for index, item := range result.Items {
		require.Equal(t, storyIDs[index], item.StoryID)
		require.False(t, item.Success)
		require.Equal(t, context.Canceled.Error(), item.Error)
	}
}
