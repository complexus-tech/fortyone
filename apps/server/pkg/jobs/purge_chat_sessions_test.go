package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPurgeDeletedChatSessionsUsesOneUTCCutoffAcrossBoundedBatches(t *testing.T) {
	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	store := &chatSessionPurgerStub{results: []int64{maintenancePurgeBatchSize, 7}}

	err := purgeDeletedChatSessionsAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Equal(t, []int{maintenancePurgeBatchSize, maintenancePurgeBatchSize}, store.batchSizes)
	require.Equal(t, []time.Time{
		now.UTC().Add(-30 * 24 * time.Hour),
		now.UTC().Add(-30 * 24 * time.Hour),
	}, store.deletedBefore)
}

func TestDrainMaintenanceBatchesStopsOnShortBatch(t *testing.T) {
	results := []int64{maintenancePurgeBatchSize, 23}
	var calls int

	total, err := drainMaintenanceBatches(context.Background(), "test purge", func(_ context.Context, batchSize int) (int64, error) {
		require.Equal(t, maintenancePurgeBatchSize, batchSize)
		result := results[calls]
		calls++
		return result, nil
	})

	require.NoError(t, err)
	require.Equal(t, int64(maintenancePurgeBatchSize+23), total)
	require.Equal(t, 2, calls)
}

func TestDrainMaintenanceBatchesRejectsInvalidRowCounts(t *testing.T) {
	tests := map[string]int64{
		"negative":  -1,
		"oversized": int64(maintenancePurgeBatchSize + 1),
	}

	for name, result := range tests {
		t.Run(name, func(t *testing.T) {
			total, err := drainMaintenanceBatches(context.Background(), "test purge", func(context.Context, int) (int64, error) {
				return result, nil
			})

			require.Zero(t, total)
			require.ErrorIs(t, err, errInvalidMaintenancePurgeResult)
		})
	}
}

func TestDrainMaintenanceBatchesPreservesPartialProgressOnStoreError(t *testing.T) {
	storeErr := errors.New("store unavailable")
	var calls int

	total, err := drainMaintenanceBatches(context.Background(), "test purge", func(context.Context, int) (int64, error) {
		calls++
		if calls == 1 {
			return maintenancePurgeBatchSize, nil
		}
		return 0, storeErr
	})

	require.Equal(t, int64(maintenancePurgeBatchSize), total)
	require.ErrorIs(t, err, storeErr)
}

func TestDrainMaintenanceBatchesHonorsCancellationBeforeStoreCall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false

	total, err := drainMaintenanceBatches(ctx, "test purge", func(context.Context, int) (int64, error) {
		called = true
		return 0, nil
	})

	require.Zero(t, total)
	require.False(t, called)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDrainMaintenanceBatchesReportsPersistentBacklog(t *testing.T) {
	var calls int

	total, err := drainMaintenanceBatches(context.Background(), "test purge", func(context.Context, int) (int64, error) {
		calls++
		return maintenancePurgeBatchSize, nil
	})

	require.Equal(t, maintenancePurgeMaxBatches, calls)
	require.Equal(t, int64(maintenancePurgeBatchSize*maintenancePurgeMaxBatches), total)
	require.ErrorIs(t, err, errMaintenanceBacklogRemaining)
}

type chatSessionPurgerStub struct {
	results       []int64
	deletedBefore []time.Time
	batchSizes    []int
}

func (store *chatSessionPurgerStub) PurgeDeletedChatSessions(
	_ context.Context,
	deletedBefore time.Time,
	batchSize int,
) (int64, error) {
	store.deletedBefore = append(store.deletedBefore, deletedBefore)
	store.batchSizes = append(store.batchSizes, batchSize)
	if len(store.results) == 0 {
		return 0, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}
