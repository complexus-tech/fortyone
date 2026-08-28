package taskhandlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	platformidempotency "github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type idempotencyPurgerStub struct {
	results []int64
	err     error
	calls   []int
}

func (stub *idempotencyPurgerStub) PurgeExpired(_ context.Context, batchSize int) (int64, error) {
	stub.calls = append(stub.calls, batchSize)
	if stub.err != nil {
		return 0, stub.err
	}
	if len(stub.results) == 0 {
		return 0, nil
	}
	result := stub.results[0]
	stub.results = stub.results[1:]
	return result, nil
}

func TestIdempotencyCleanupDrainsUntilShortBatch(t *testing.T) {
	purger := &idempotencyPurgerStub{results: []int64{
		platformidempotency.MaxPurgeBatchSize,
		platformidempotency.MaxPurgeBatchSize,
		17,
	}}
	handler := NewIdempotencyCleanupHandler(testTaskLogger(), purger)

	require.NoError(t, handler.Handle(context.Background(), asynq.NewTask("test", nil)))
	require.Equal(t, []int{
		platformidempotency.MaxPurgeBatchSize,
		platformidempotency.MaxPurgeBatchSize,
		platformidempotency.MaxPurgeBatchSize,
	}, purger.calls)
}

func TestIdempotencyCleanupPropagatesPurgeFailure(t *testing.T) {
	sentinel := errors.New("database unavailable")
	handler := NewIdempotencyCleanupHandler(testTaskLogger(), &idempotencyPurgerStub{err: sentinel})

	err := handler.Handle(context.Background(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)
}

func TestIdempotencyCleanupRejectsInvalidStoreResult(t *testing.T) {
	handler := NewIdempotencyCleanupHandler(testTaskLogger(), &idempotencyPurgerStub{
		results: []int64{platformidempotency.MaxPurgeBatchSize + 1},
	})

	err := handler.Handle(context.Background(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, errInvalidIdempotencyPurgeResult)
}

func TestIdempotencyCleanupRequiresDependencies(t *testing.T) {
	err := NewIdempotencyCleanupHandler(nil, nil).Handle(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "dependencies are required")
}

func testTaskLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelError, "idempotency-cleanup-test")
}
