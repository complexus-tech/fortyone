package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	"github.com/stretchr/testify/require"
)

func TestPurgeMessagingDataUsesOneUTCCutoffSetAcrossBoundedBatches(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	store := &messagingDataPurgerStub{results: []messagingdomain.RetentionPurgeResult{
		{NoncesDeleted: maintenancePurgeBatchSize, MessagesDeleted: 2},
		{NoncesDeleted: 3, ConversationsDeleted: 4},
	}}

	err := purgeMessagingDataAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Equal(t, []int{maintenancePurgeBatchSize, maintenancePurgeBatchSize}, store.batchSizes)
	require.Equal(t, []messagingdomain.RetentionCutoffs{
		messagingRetentionCutoffsAt(now.UTC()),
		messagingRetentionCutoffsAt(now.UTC()),
	}, store.cutoffs)
}

func TestDrainMessagingRetentionBatchesAggregatesCommittedResults(t *testing.T) {
	t.Parallel()

	store := &messagingDataPurgerStub{results: []messagingdomain.RetentionPurgeResult{
		{InboundEventsDeleted: maintenancePurgeBatchSize, MessagesDeleted: 8},
		{InboundEventsDeleted: 2, ReplyTokensDeleted: 4},
	}}
	var committed []messagingdomain.RetentionPurgeResult

	result, err := drainMessagingRetentionBatches(
		t.Context(),
		store,
		messagingRetentionCutoffsAt(time.Now().UTC()),
		func(_ int, result messagingdomain.RetentionPurgeResult) {
			committed = append(committed, result)
		},
	)

	require.NoError(t, err)
	require.Equal(t, messagingdomain.RetentionPurgeResult{
		InboundEventsDeleted: maintenancePurgeBatchSize + 2,
		MessagesDeleted:      8,
		ReplyTokensDeleted:   4,
	}, result)
	require.Len(t, committed, 2)
	require.Equal(t, int64(maintenancePurgeBatchSize+14), result.TotalAffected())
}

func TestDrainMessagingRetentionBatchesRejectsInvalidPerKindCount(t *testing.T) {
	t.Parallel()

	store := &messagingDataPurgerStub{results: []messagingdomain.RetentionPurgeResult{{
		OutboundDeliveriesDeleted: maintenancePurgeBatchSize + 1,
	}}}
	committed := false

	result, err := drainMessagingRetentionBatches(
		t.Context(),
		store,
		messagingRetentionCutoffsAt(time.Now().UTC()),
		func(int, messagingdomain.RetentionPurgeResult) { committed = true },
	)

	require.Equal(t, messagingdomain.RetentionPurgeResult{}, result)
	require.ErrorIs(t, err, errInvalidMaintenancePurgeResult)
	require.False(t, committed)
}

func TestDrainMessagingRetentionBatchesReportsPartialCommittedProgressOnStoreError(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("store unavailable")
	store := &messagingDataPurgerStub{
		results: []messagingdomain.RetentionPurgeResult{{NoncesDeleted: maintenancePurgeBatchSize}},
		errors:  []error{nil, storeErr},
	}

	result, err := drainMessagingRetentionBatches(
		t.Context(),
		store,
		messagingRetentionCutoffsAt(time.Now().UTC()),
		nil,
	)

	require.Equal(t, int64(maintenancePurgeBatchSize), result.NoncesDeleted)
	require.ErrorIs(t, err, storeErr)
	require.Len(t, store.cutoffs, 2)
	require.Equal(t, store.cutoffs[0], store.cutoffs[1])
}

func TestDrainMessagingRetentionBatchesReportsPersistentBacklog(t *testing.T) {
	t.Parallel()

	results := make([]messagingdomain.RetentionPurgeResult, maintenancePurgeMaxBatches)
	for index := range results {
		results[index].MessagesDeleted = maintenancePurgeBatchSize
	}
	store := &messagingDataPurgerStub{results: results}

	result, err := drainMessagingRetentionBatches(
		t.Context(),
		store,
		messagingRetentionCutoffsAt(time.Now().UTC()),
		nil,
	)

	require.ErrorIs(t, err, errMaintenanceBacklogRemaining)
	require.Len(t, store.batchSizes, maintenancePurgeMaxBatches)
	require.Equal(t, int64(maintenancePurgeBatchSize*maintenancePurgeMaxBatches), result.MessagesDeleted)
}

type messagingDataPurgerStub struct {
	results    []messagingdomain.RetentionPurgeResult
	errors     []error
	cutoffs    []messagingdomain.RetentionCutoffs
	batchSizes []int
}

func (store *messagingDataPurgerStub) PurgeMessagingDataBatch(
	_ context.Context,
	cutoffs messagingdomain.RetentionCutoffs,
	batchSize int,
) (messagingdomain.RetentionPurgeResult, error) {
	store.cutoffs = append(store.cutoffs, cutoffs)
	store.batchSizes = append(store.batchSizes, batchSize)
	var result messagingdomain.RetentionPurgeResult
	if len(store.results) > 0 {
		result = store.results[0]
		store.results = store.results[1:]
	}
	var err error
	if len(store.errors) > 0 {
		err = store.errors[0]
		store.errors = store.errors[1:]
	}
	return result, err
}

func messagingRetentionCutoffsAt(now time.Time) messagingdomain.RetentionCutoffs {
	return messagingdomain.RetentionCutoffs{
		ExpiredNoncesBefore:    now.Add(-messagingAliasRetention),
		ConfirmationsExpiredAt: now,
		ProviderDataBefore:     now.Add(-messagingProviderRetention),
		ReplyTokensBefore:      now.Add(-messagingAliasRetention),
	}
}
