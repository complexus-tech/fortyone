package eventconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var pendingRecoveryEpoch = time.Date(2026, time.September, 4, 10, 0, 0, 0, time.UTC)

func TestPendingRecoveryReachesLaterEventsAndRetriesEarlierFailures(t *testing.T) {
	t.Parallel()

	consumer, server, mail := newPendingRecoveryConsumer(t)
	queued := make([]events.Event, pendingClaimBatchSize+1)
	for index := range pendingClaimBatchSize {
		queued[index] = events.Event{Type: "unsupported-event"}
	}
	queued[pendingClaimBatchSize] = recoveryVerificationEvent()
	ids := seedPendingRecoveryEvents(t, consumer, queued)
	server.SetTime(pendingRecoveryEpoch.Add(pendingClaimTimeout + time.Second))

	cursor, err := consumer.claimPendingBatch(context.Background(), "recovery", pendingClaimScanStart)
	require.NoError(t, err)
	require.Empty(t, mail.sent, "a recovery tick must stay within its page budget")
	pending := readRecoveryPending(t, consumer)
	require.Len(t, pending, len(queued), "failed events must remain pending")
	require.Equal(t, int64(2), pending[0].RetryCount)
	require.Equal(t, int64(1), pending[len(pending)-1].RetryCount)

	cursor, err = consumer.claimPendingBatch(context.Background(), "recovery", cursor)
	require.NoError(t, err)
	require.Equal(t, []string{"recipient@example.test"}, mail.sent)
	pending = readRecoveryPending(t, consumer)
	require.Len(t, pending, pendingClaimBatchSize)
	for _, message := range pending {
		require.NotEqual(t, ids[len(ids)-1], message.ID, "the later event must be acknowledged")
		require.Equal(t, int64(2), message.RetryCount, "the cursor boundary must not be retried twice")
	}

	server.SetTime(pendingRecoveryEpoch.Add(2 * (pendingClaimTimeout + time.Second)))
	_, err = consumer.claimPendingBatch(context.Background(), "recovery", cursor)
	require.NoError(t, err)
	for _, message := range readRecoveryPending(t, consumer) {
		require.Equal(t, int64(3), message.RetryCount, "a completed scan must revisit failed events")
	}
	require.Len(t, mail.sent, 1, "acknowledged events must not be redelivered")
}

func TestPendingRecoverySkipsRecentlyDeliveredEvents(t *testing.T) {
	t.Parallel()

	consumer, server, mail := newPendingRecoveryConsumer(t)
	ids := seedPendingRecoveryEvents(t, consumer, []events.Event{
		recoveryVerificationEvent(), recoveryVerificationEvent(),
	})
	server.SetTime(pendingRecoveryEpoch.Add(pendingClaimTimeout + time.Second))
	_, err := consumer.redis.XClaim(context.Background(), &redis.XClaimArgs{
		Stream: eventStreamKey, Group: eventConsumerGroup, Consumer: "active",
		MinIdle: 0, Messages: []string{ids[0]},
	}).Result()
	require.NoError(t, err)

	cursor, err := consumer.claimPendingBatch(context.Background(), "recovery", pendingClaimScanStart)
	require.NoError(t, err)
	require.Len(t, mail.sent, 1)
	pending := readRecoveryPending(t, consumer)
	require.Len(t, pending, 1)
	require.Equal(t, ids[0], pending[0].ID)
	require.Equal(t, "active", pending[0].Consumer)
	require.Equal(t, int64(2), pending[0].RetryCount)

	server.SetTime(pendingRecoveryEpoch.Add(2 * (pendingClaimTimeout + time.Second)))
	_, err = consumer.claimPendingBatch(context.Background(), "recovery", cursor)
	require.NoError(t, err)
	require.Len(t, mail.sent, 2)
	require.Empty(t, readRecoveryPending(t, consumer))
}

func TestPendingRecoveryCancellationStopsBeforeClaimingAnotherEvent(t *testing.T) {
	t.Parallel()

	consumer, server, mail := newPendingRecoveryConsumer(t)
	ids := seedPendingRecoveryEvents(t, consumer, []events.Event{
		recoveryVerificationEvent(), recoveryVerificationEvent(),
	})
	server.SetTime(pendingRecoveryEpoch.Add(pendingClaimTimeout + time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mail.afterSend = cancel

	_, err := consumer.claimPendingBatch(ctx, "recovery", pendingClaimScanStart)
	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, mail.sent, 1)
	pending := readRecoveryPending(t, consumer)
	var laterMessage *redis.XPendingExt
	for index := range pending {
		if pending[index].ID == ids[1] {
			laterMessage = &pending[index]
		}
	}
	require.NotNil(t, laterMessage)
	require.Equal(t, "original", laterMessage.Consumer)
	require.Equal(t, int64(1), laterMessage.RetryCount)

	commandsBefore := server.CommandCount()
	_, err = consumer.claimPendingBatch(ctx, "recovery", pendingClaimScanStart)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, commandsBefore, server.CommandCount(), "an already canceled batch must not touch Redis")
}

func TestPendingRecoveryPreservesCursorWhenRedisIsUnavailable(t *testing.T) {
	t.Parallel()

	consumer, server, _ := newPendingRecoveryConsumer(t)
	server.SetError("ERR temporarily unavailable")
	cursor, err := consumer.claimPendingBatch(context.Background(), "recovery", "100-0")
	require.Error(t, err)
	require.Equal(t, "100-0", cursor)
}

type pendingRecoveryMailer struct {
	sent      []string
	afterSend func()
}

func (*pendingRecoveryMailer) Send(context.Context, mailer.Email) error {
	return errors.New("unexpected untemplated email")
}

func (m *pendingRecoveryMailer) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	m.sent = append(m.sent, email.To...)
	if m.afterSend != nil {
		m.afterSend()
	}
	return nil
}

func newPendingRecoveryConsumer(t *testing.T) (*Consumer, *miniredis.Miniredis, *pendingRecoveryMailer) {
	t.Helper()
	server := miniredis.RunT(t)
	server.SetTime(pendingRecoveryEpoch)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	mail := &pendingRecoveryMailer{}
	consumer := &Consumer{
		redis: client, log: logger.NewWithText(io.Discard, slog.LevelDebug, "recovery-test"),
		mailerService: mail, websiteURL: "https://example.test",
	}
	require.NoError(t, consumer.Initialize(context.Background()))
	return consumer, server, mail
}

func recoveryVerificationEvent() events.Event {
	return events.Event{
		Type: events.EmailVerification,
		Payload: events.EmailVerificationPayload{
			Email: "recipient@example.test", Token: "test-verification-code",
		},
		Timestamp: pendingRecoveryEpoch,
	}
}

func seedPendingRecoveryEvents(t *testing.T, consumer *Consumer, queued []events.Event) []string {
	t.Helper()
	ctx := context.Background()
	ids := make([]string, len(queued))
	for index, event := range queued {
		payload, err := json.Marshal(event)
		require.NoError(t, err)
		ids[index], err = consumer.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: eventStreamKey,
			Values: map[string]any{"type": string(event.Type), "payload": string(payload)},
		}).Result()
		require.NoError(t, err)
	}
	streams, err := consumer.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: eventConsumerGroup, Consumer: "original", Streams: []string{eventStreamKey, ">"},
		Count: int64(len(queued)), Block: -1,
	}).Result()
	require.NoError(t, err)
	require.Len(t, streams, 1)
	require.Len(t, streams[0].Messages, len(queued))
	return ids
}

func readRecoveryPending(t *testing.T, consumer *Consumer) []redis.XPendingExt {
	t.Helper()
	pending, err := consumer.redis.XPendingExt(context.Background(), &redis.XPendingExtArgs{
		Stream: eventStreamKey, Group: eventConsumerGroup, Start: "-", End: "+", Count: 100,
	}).Result()
	require.NoError(t, err)
	return pending
}
