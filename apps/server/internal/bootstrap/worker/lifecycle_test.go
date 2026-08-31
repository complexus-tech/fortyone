package workerbootstrap

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type fakeTaskServer struct {
	startErr error
	started  chan struct{}
	start    atomic.Bool
	stop     atomic.Bool
	shutdown atomic.Bool
}

func (s *fakeTaskServer) Start(asynq.Handler) error {
	s.start.Store(true)
	if s.started != nil {
		close(s.started)
	}
	return s.startErr
}

func (s *fakeTaskServer) Stop() {
	s.stop.Store(true)
}

func (s *fakeTaskServer) Shutdown() {
	s.shutdown.Store(true)
}

type fakeTaskScheduler struct {
	startErr error
	start    atomic.Bool
	shutdown atomic.Bool
}

func (s *fakeTaskScheduler) Start() error {
	s.start.Store(true)
	return s.startErr
}

func (s *fakeTaskScheduler) Shutdown() {
	s.shutdown.Store(true)
}

func TestWorkerRunSupervisesStartupAndGracefulShutdown(t *testing.T) {
	t.Parallel()

	server := &fakeTaskServer{started: make(chan struct{})}
	scheduler := &fakeTaskScheduler{}
	app := newLifecycleTestApp(server, scheduler)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- app.Run(ctx)
	}()

	<-server.started
	cancel()

	require.NoError(t, <-result)
	require.True(t, scheduler.start.Load())
	require.True(t, scheduler.shutdown.Load())
	require.True(t, server.start.Load())
	require.True(t, server.stop.Load())
	require.True(t, server.shutdown.Load())
	require.False(t, app.ready.Load())
}

func TestWorkerRunStopsSchedulerWhenTaskServerCannotStart(t *testing.T) {
	t.Parallel()

	startErr := errors.New("cannot reserve queues")
	server := &fakeTaskServer{startErr: startErr}
	scheduler := &fakeTaskScheduler{}
	app := newLifecycleTestApp(server, scheduler)

	err := app.Run(context.Background())

	require.ErrorContains(t, err, "Worker task server could not start")
	require.ErrorIs(t, err, startErr)
	require.True(t, scheduler.start.Load())
	require.True(t, scheduler.shutdown.Load())
	require.True(t, server.start.Load())
	require.False(t, server.stop.Load())
	require.False(t, server.shutdown.Load())
}

func TestWorkerRunDoesNotStartTaskServerWhenSchedulerCannotStart(t *testing.T) {
	t.Parallel()

	server := &fakeTaskServer{}
	startErr := errors.New("cannot register schedules")
	scheduler := &fakeTaskScheduler{startErr: startErr}
	app := newLifecycleTestApp(server, scheduler)

	err := app.Run(context.Background())

	require.ErrorContains(t, err, "Worker scheduler could not start")
	require.ErrorIs(t, err, startErr)
	require.True(t, scheduler.start.Load())
	require.False(t, scheduler.shutdown.Load())
	require.False(t, server.start.Load())
}

func TestWorkerRunRejectsNilContext(t *testing.T) {
	t.Parallel()

	app := newLifecycleTestApp(&fakeTaskServer{}, &fakeTaskScheduler{})
	var nilContext context.Context
	require.ErrorContains(t, app.Run(nilContext), "context is required")
}

func newLifecycleTestApp(server taskServer, scheduler taskScheduler) *App {
	var logs bytes.Buffer
	return &App{
		log:        logger.NewWithJSON(&logs, slog.LevelDebug, "worker-lifecycle-test"),
		server:     server,
		scheduler:  scheduler,
		taskMux:    asynq.NewServeMux(),
		pingRedis:  func(context.Context) error { return nil },
		httpConfig: validWorkerHTTPConfig(),
		ready:      &atomic.Bool{},
	}
}
