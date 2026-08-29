package eventconsumer

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestConsumerRejectsRunBeforeInitialization(t *testing.T) {
	t.Parallel()

	consumer := &Consumer{redis: redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})}
	t.Cleanup(func() { require.NoError(t, consumer.redis.Close()) })

	require.ErrorContains(t, consumer.Run(context.Background()), "must be initialized")
}

func TestRedisGroupErrorClassification(t *testing.T) {
	t.Parallel()

	require.True(t, isRedisGroupError(errors.New("BUSYGROUP Consumer Group name already exists"), "BUSYGROUP"))
	require.True(t, isRedisGroupError(errors.New("NOGROUP No such key"), "NOGROUP"))
	require.False(t, isRedisGroupError(errors.New("dial tcp: connection refused"), "NOGROUP"))
}

func TestConsumerRetryWaitHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()

	require.False(t, waitForConsumerRetry(ctx, time.Minute))
	require.Less(t, time.Since(started), 250*time.Millisecond)
}

func TestConsumerInitializationFailureIsFatal(t *testing.T) {
	t.Parallel()

	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	consumer := &Consumer{redis: redisClient}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := consumer.Initialize(ctx)
	require.ErrorContains(t, err, "create Redis stream consumer group")
	require.False(t, consumer.lifecycle.initialized.Load())
}

func TestConsumerRunJoinsLoopsOnCancellation(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	var output bytes.Buffer
	consumer := &Consumer{
		redis: redisClient,
		log:   logger.NewWithJSON(&output, slog.LevelDebug, "consumer-test"),
	}
	require.NoError(t, consumer.Initialize(context.Background()))

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	commandsBeforeRun := redisServer.CommandCount()
	go func() { result <- consumer.Run(ctx) }()
	require.Eventually(t, func() bool {
		return redisServer.CommandCount() > commandsBeforeRun
	}, 5*time.Second, 10*time.Millisecond, "consumer read loop did not start")
	cancel()

	joinDeadline := time.NewTimer(streamReadBlock + time.Second)
	defer joinDeadline.Stop()
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-joinDeadline.C:
		t.Fatal("consumer loops were not joined after cancellation")
	}
}
