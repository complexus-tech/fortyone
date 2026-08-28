package messagingrepository

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/stretchr/testify/require"
)

func TestMessagingMaintenanceQueriesAreBoundedOrderedAndApplicationClocked(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/maintenance.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	queryNames := []string{
		"-- name: purgeexpiredmessagingnonces :execrows",
		"-- name: expiremessagingstorymutationconfirmations :execrows",
		"-- name: purgeoldmessagingoutbounddeliveries :execrows",
		"-- name: purgeoldmessaginginboundevents :execrows",
		"-- name: purgecompletedslackuninstalls :execrows",
		"-- name: purgeoldmessagingmessages :execrows",
		"-- name: purgeexpiredmessagingemailreplytokens :execrows",
		"-- name: purgeemptymessagingconversations :execrows",
	}
	lastPosition := -1
	for _, name := range queryNames {
		position := strings.Index(query, name)
		if position < 0 {
			t.Errorf("messaging maintenance query is missing %q", name)
			continue
		}
		if position <= lastPosition {
			t.Errorf("messaging maintenance query %q is out of retention order", name)
		}
		lastPosition = position
	}

	require.Equal(t, 8, strings.Count(query, "with candidates as materialized"))
	require.Equal(t, 8, strings.Count(query, "limit cast(sqlc.arg(batch_size) as integer)"))
	require.Equal(t, 8, strings.Count(query, "skip locked"))
	require.NotContains(t, query, "now()")
	require.NotContains(t, query, "current_timestamp")

	for _, contract := range []string{
		"nonce.expires_at < sqlc.arg(expired_before)",
		"confirmation.expires_at <= sqlc.arg(expired_at)",
		"confirmation.status = 'applied' and confirmation.operation = 'create_stories' and confirmation.proposal is not null",
		"set status = 'expired', proposal = null, applied_at = null, expired_at = sqlc.arg(expired_at)",
		"delivery.created_at < sqlc.arg(created_before)",
		"event.received_at < sqlc.arg(received_before)",
		"uninstall.status = 'completed' and uninstall.completed_at < sqlc.arg(completed_before)",
		"message.created_at < sqlc.arg(created_before)",
		"token.expires_at < sqlc.arg(retained_before) or (token.revoked_at is not null and token.revoked_at < sqlc.arg(retained_before))",
		"conversation.updated_at < sqlc.arg(updated_before) and not exists",
		"message.conversation_id = conversation.id",
	} {
		require.Contains(t, query, contract)
	}
}

func TestPurgeMessagingDataBatchUsesOneTransactionAndPreservesOperationOrder(t *testing.T) {
	t.Parallel()

	results := messagingdomain.RetentionPurgeResult{
		NoncesDeleted:                   1,
		ConfirmationsRedacted:           2,
		OutboundDeliveriesDeleted:       3,
		InboundEventsDeleted:            4,
		CompletedSlackUninstallsDeleted: 5,
		MessagesDeleted:                 6,
		ReplyTokensDeleted:              7,
		ConversationsDeleted:            8,
	}
	outerQueries := &messagingMaintenanceQueries{}
	txQueries := &messagingMaintenanceQueries{results: results}
	repository := &Repository{queries: outerQueries}
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(messagingsql.Querier) error,
	) error {
		transactionOutcome = "begun"
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}
	cutoffs := messagingRetentionTestCutoffs()

	result, err := repository.PurgeMessagingDataBatch(t.Context(), cutoffs, 500)

	require.NoError(t, err)
	require.Equal(t, results, result)
	require.Equal(t, "committed", transactionOutcome)
	require.Empty(t, outerQueries.calls)
	require.Equal(t, []string{
		"expired_nonces",
		"expired_confirmations",
		"old_outbound_deliveries",
		"old_inbound_events",
		"completed_slack_uninstalls",
		"old_messages",
		"expired_reply_tokens",
		"empty_conversations",
	}, txQueries.calls)
	txQueries.requireParams(t, utcMessagingRetentionCutoffs(cutoffs), 500)
}

func TestPurgeMessagingDataBatchReturnsNoRolledBackCountsAndStopsOnFailure(t *testing.T) {
	t.Parallel()

	databaseErr := errors.New("database unavailable")
	txQueries := &messagingMaintenanceQueries{
		results: messagingdomain.RetentionPurgeResult{
			NoncesDeleted:             1,
			ConfirmationsRedacted:     2,
			OutboundDeliveriesDeleted: 3,
		},
		failAt:  "old_inbound_events",
		failErr: databaseErr,
	}
	repository := &Repository{queries: &messagingMaintenanceQueries{}}
	transactionOutcome := ""
	repository.runTransaction = func(
		ctx context.Context,
		operation func(messagingsql.Querier) error,
	) error {
		err := operation(txQueries)
		if err != nil {
			transactionOutcome = "rolled_back"
			return err
		}
		transactionOutcome = "committed"
		return nil
	}

	result, err := repository.PurgeMessagingDataBatch(t.Context(), messagingRetentionTestCutoffs(), 500)

	require.ErrorIs(t, err, databaseErr)
	require.ErrorContains(t, err, "purge old messaging inbound events")
	require.Equal(t, messagingdomain.RetentionPurgeResult{}, result)
	require.Equal(t, "rolled_back", transactionOutcome)
	require.Equal(t, []string{
		"expired_nonces",
		"expired_confirmations",
		"old_outbound_deliveries",
		"old_inbound_events",
	}, txQueries.calls)
}

func TestPurgeMessagingDataBatchRejectsInvalidInputsBeforeTransaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cutoffs   messagingdomain.RetentionCutoffs
		batchSize int
		wantError error
	}{
		{
			name:      "missing cutoffs",
			batchSize: 500,
		},
		{
			name: "cutoff after expiry clock",
			cutoffs: func() messagingdomain.RetentionCutoffs {
				cutoffs := messagingRetentionTestCutoffs()
				cutoffs.ProviderDataBefore = cutoffs.ConfirmationsExpiredAt.Add(time.Second)
				return cutoffs
			}(),
			batchSize: 500,
		},
		{
			name:      "batch overflow",
			cutoffs:   messagingRetentionTestCutoffs(),
			batchSize: int(math.MaxInt32) + 1,
			wantError: safecast.ErrOutOfRange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			queries := &messagingMaintenanceQueries{}
			repository := &Repository{queries: queries}
			transactionCalled := false
			repository.runTransaction = func(context.Context, func(messagingsql.Querier) error) error {
				transactionCalled = true
				return nil
			}

			_, err := repository.PurgeMessagingDataBatch(t.Context(), test.cutoffs, test.batchSize)

			require.Error(t, err)
			if test.wantError != nil {
				require.ErrorIs(t, err, test.wantError)
			}
			require.False(t, transactionCalled)
			require.Empty(t, queries.calls)
		})
	}
}

type messagingMaintenanceQueries struct {
	messagingsql.Querier
	results messagingdomain.RetentionPurgeResult
	failAt  string
	failErr error
	calls   []string

	nonceParams        messagingsql.PurgeExpiredMessagingNoncesParams
	confirmationParams messagingsql.ExpireMessagingStoryMutationConfirmationsParams
	outboundParams     messagingsql.PurgeOldMessagingOutboundDeliveriesParams
	inboundParams      messagingsql.PurgeOldMessagingInboundEventsParams
	uninstallParams    messagingsql.PurgeCompletedSlackUninstallsParams
	messageParams      messagingsql.PurgeOldMessagingMessagesParams
	replyTokenParams   messagingsql.PurgeExpiredMessagingEmailReplyTokensParams
	conversationParams messagingsql.PurgeEmptyMessagingConversationsParams
}

func (queries *messagingMaintenanceQueries) result(name string, rows int64) (int64, error) {
	queries.calls = append(queries.calls, name)
	if queries.failAt == name {
		return 0, queries.failErr
	}
	return rows, nil
}

func (queries *messagingMaintenanceQueries) PurgeExpiredMessagingNonces(
	_ context.Context,
	params messagingsql.PurgeExpiredMessagingNoncesParams,
) (int64, error) {
	queries.nonceParams = params
	return queries.result("expired_nonces", queries.results.NoncesDeleted)
}

func (queries *messagingMaintenanceQueries) ExpireMessagingStoryMutationConfirmations(
	_ context.Context,
	params messagingsql.ExpireMessagingStoryMutationConfirmationsParams,
) (int64, error) {
	queries.confirmationParams = params
	return queries.result("expired_confirmations", queries.results.ConfirmationsRedacted)
}

func (queries *messagingMaintenanceQueries) PurgeOldMessagingOutboundDeliveries(
	_ context.Context,
	params messagingsql.PurgeOldMessagingOutboundDeliveriesParams,
) (int64, error) {
	queries.outboundParams = params
	return queries.result("old_outbound_deliveries", queries.results.OutboundDeliveriesDeleted)
}

func (queries *messagingMaintenanceQueries) PurgeOldMessagingInboundEvents(
	_ context.Context,
	params messagingsql.PurgeOldMessagingInboundEventsParams,
) (int64, error) {
	queries.inboundParams = params
	return queries.result("old_inbound_events", queries.results.InboundEventsDeleted)
}

func (queries *messagingMaintenanceQueries) PurgeCompletedSlackUninstalls(
	_ context.Context,
	params messagingsql.PurgeCompletedSlackUninstallsParams,
) (int64, error) {
	queries.uninstallParams = params
	return queries.result("completed_slack_uninstalls", queries.results.CompletedSlackUninstallsDeleted)
}

func (queries *messagingMaintenanceQueries) PurgeOldMessagingMessages(
	_ context.Context,
	params messagingsql.PurgeOldMessagingMessagesParams,
) (int64, error) {
	queries.messageParams = params
	return queries.result("old_messages", queries.results.MessagesDeleted)
}

func (queries *messagingMaintenanceQueries) PurgeExpiredMessagingEmailReplyTokens(
	_ context.Context,
	params messagingsql.PurgeExpiredMessagingEmailReplyTokensParams,
) (int64, error) {
	queries.replyTokenParams = params
	return queries.result("expired_reply_tokens", queries.results.ReplyTokensDeleted)
}

func (queries *messagingMaintenanceQueries) PurgeEmptyMessagingConversations(
	_ context.Context,
	params messagingsql.PurgeEmptyMessagingConversationsParams,
) (int64, error) {
	queries.conversationParams = params
	return queries.result("empty_conversations", queries.results.ConversationsDeleted)
}

func (queries *messagingMaintenanceQueries) requireParams(
	t *testing.T,
	cutoffs messagingdomain.RetentionCutoffs,
	batchSize int32,
) {
	t.Helper()
	require.Equal(t, cutoffs.ExpiredNoncesBefore, queries.nonceParams.ExpiredBefore)
	require.Equal(t, batchSize, queries.nonceParams.BatchSize)
	require.NotNil(t, queries.confirmationParams.ExpiredAt)
	require.Equal(t, cutoffs.ConfirmationsExpiredAt, *queries.confirmationParams.ExpiredAt)
	require.Equal(t, batchSize, queries.confirmationParams.BatchSize)
	require.Equal(t, cutoffs.ProviderDataBefore, queries.outboundParams.CreatedBefore)
	require.Equal(t, batchSize, queries.outboundParams.BatchSize)
	require.Equal(t, cutoffs.ProviderDataBefore, queries.inboundParams.ReceivedBefore)
	require.Equal(t, batchSize, queries.inboundParams.BatchSize)
	require.NotNil(t, queries.uninstallParams.CompletedBefore)
	require.Equal(t, cutoffs.ProviderDataBefore, *queries.uninstallParams.CompletedBefore)
	require.Equal(t, batchSize, queries.uninstallParams.BatchSize)
	require.Equal(t, cutoffs.ProviderDataBefore, queries.messageParams.CreatedBefore)
	require.Equal(t, batchSize, queries.messageParams.BatchSize)
	require.Equal(t, cutoffs.ReplyTokensBefore, queries.replyTokenParams.RetainedBefore)
	require.Equal(t, batchSize, queries.replyTokenParams.BatchSize)
	require.Equal(t, cutoffs.ProviderDataBefore, queries.conversationParams.UpdatedBefore)
	require.Equal(t, batchSize, queries.conversationParams.BatchSize)
}

func messagingRetentionTestCutoffs() messagingdomain.RetentionCutoffs {
	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	return messagingdomain.RetentionCutoffs{
		ExpiredNoncesBefore:    now.Add(-24 * time.Hour),
		ConfirmationsExpiredAt: now,
		ProviderDataBefore:     now.Add(-30 * 24 * time.Hour),
		ReplyTokensBefore:      now.Add(-24 * time.Hour),
	}
}
