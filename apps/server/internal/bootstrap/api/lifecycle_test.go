package api

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/stretchr/testify/require"
)

type fakeAPIComponent struct {
	initializeErr error
	runErr        error
	fail          chan struct{}
	started       chan struct{}
	initialize    atomic.Int32
	run           atomic.Int32
	startOnce     sync.Once
}

func (c *fakeAPIComponent) Initialize(context.Context) error {
	c.initialize.Add(1)
	return c.initializeErr
}

func (c *fakeAPIComponent) Run(ctx context.Context) error {
	c.run.Add(1)
	c.startOnce.Do(func() {
		if c.started != nil {
			close(c.started)
		}
	})
	if c.fail == nil {
		<-ctx.Done()
		return nil
	}
	select {
	case <-ctx.Done():
		return nil
	case <-c.fail:
		return c.runErr
	}
}

type fakeTelemetry struct {
	shutdown atomic.Int32
	err      error
}

type drainErrorAPIComponent struct {
	started chan struct{}
	err     error
}

func (c *drainErrorAPIComponent) Initialize(context.Context) error { return nil }

func (c *drainErrorAPIComponent) Run(ctx context.Context) error {
	close(c.started)
	<-ctx.Done()
	return c.err
}

func (t *fakeTelemetry) Shutdown(context.Context) error {
	t.shutdown.Add(1)
	return t.err
}

func TestProcessSupervisesHTTPAndBackgroundComponents(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	telemetry := &fakeTelemetry{}
	consumer := &fakeAPIComponent{started: make(chan struct{})}
	hub := &fakeAPIComponent{started: make(chan struct{})}
	requestStarted := make(chan struct{})
	requestStopped := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
		close(requestStopped)
		w.WriteHeader(http.StatusNoContent)
	})

	var closeCount atomic.Int32
	process := newLifecycleTestProcess(t, readiness, telemetry, handler,
		[]Component{
			{Name: "redis_stream_consumer", Runtime: consumer},
			{Name: "sse_hub", Runtime: hub},
		},
		[]Resource{{Name: "dependencies", Close: func() error {
			closeCount.Add(1)
			return nil
		}}},
	)
	listener := newTestListener(t)
	process.listen = func(string, string) (net.Listener, error) { return listener, nil }

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- process.Run(ctx) }()

	require.Eventually(t, func() bool {
		return readiness.Phase() == platformhealth.PhaseReady
	}, time.Second, time.Millisecond)
	<-consumer.started
	<-hub.started

	requestResult := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if response != nil {
			_ = response.Body.Close()
		}
		requestResult <- err
	}()
	<-requestStarted
	cancel()

	require.NoError(t, <-result)
	require.NoError(t, <-requestResult)
	<-requestStopped
	require.Equal(t, platformhealth.PhaseDraining, readiness.Phase())
	require.Equal(t, int32(1), telemetry.shutdown.Load())
	require.Equal(t, int32(1), closeCount.Load())
	require.Equal(t, int32(1), consumer.initialize.Load())
	require.Equal(t, int32(1), consumer.run.Load())
	require.Equal(t, int32(1), hub.initialize.Load())
	require.Equal(t, int32(1), hub.run.Load())
}

func TestProcessPropagatesRuntimeComponentFailure(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	telemetry := &fakeTelemetry{}
	fail := make(chan struct{})
	consumer := &fakeAPIComponent{
		started: make(chan struct{}),
		fail:    fail,
		runErr:  errors.New("consumer group disappeared"),
	}
	hub := &fakeAPIComponent{}
	process := newLifecycleTestProcess(t, readiness, telemetry, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		[]Component{
			{Name: "redis_stream_consumer", Runtime: consumer},
			{Name: "sse_hub", Runtime: hub},
		},
		nil,
	)
	listener := newTestListener(t)
	process.listen = func(string, string) (net.Listener, error) { return listener, nil }

	result := make(chan error, 1)
	go func() { result <- process.Run(context.Background()) }()
	<-consumer.started
	close(fail)

	err := <-result
	require.ErrorContains(t, err, "redis_stream_consumer failed")
	require.ErrorContains(t, err, "consumer group disappeared")
	require.Equal(t, platformhealth.PhaseFailed, readiness.Phase())
	require.Equal(t, int32(1), telemetry.shutdown.Load())
}

func TestProcessTreatsComponentInitializationFailureAsFatal(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	telemetry := &fakeTelemetry{}
	component := &fakeAPIComponent{initializeErr: errors.New("NOGROUP setup failed")}
	var closeCount atomic.Int32
	process := newLifecycleTestProcess(t, readiness, telemetry, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		[]Component{{Name: "redis_stream_consumer", Runtime: component}},
		[]Resource{{Name: "database", Close: func() error {
			closeCount.Add(1)
			return nil
		}}},
	)

	err := process.Run(context.Background())
	require.ErrorContains(t, err, "initialize API component redis_stream_consumer")
	require.Equal(t, platformhealth.PhaseFailed, readiness.Phase())
	require.Equal(t, int32(1), telemetry.shutdown.Load())
	require.Equal(t, int32(1), closeCount.Load())
	require.Equal(t, int32(0), component.run.Load())

	require.ErrorContains(t, process.Run(context.Background()), "only run once")
	require.Equal(t, int32(1), telemetry.shutdown.Load())
	require.Equal(t, int32(1), closeCount.Load())
}

func TestProcessTreatsListenerBindFailureAsFatal(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	telemetry := &fakeTelemetry{}
	component := &fakeAPIComponent{}
	process := newLifecycleTestProcess(
		t,
		readiness,
		telemetry,
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		[]Component{{Name: "background", Runtime: component}},
		nil,
	)
	process.listen = func(string, string) (net.Listener, error) {
		return nil, errors.New("address already in use")
	}

	err := process.Run(context.Background())
	require.ErrorContains(t, err, "listen for API traffic")
	require.ErrorContains(t, err, "address already in use")
	require.Equal(t, platformhealth.PhaseFailed, readiness.Phase())
	require.Equal(t, int32(1), component.initialize.Load())
	require.Equal(t, int32(0), component.run.Load())
	require.Equal(t, int32(1), telemetry.shutdown.Load())
}

func TestProcessCancellationIsNeverClassifiedAsComponentFailure(t *testing.T) {
	for iteration := 0; iteration < 25; iteration++ {
		readiness := newAPIReadiness(t)
		component := &fakeAPIComponent{started: make(chan struct{})}
		process := newLifecycleTestProcess(
			t,
			readiness,
			&fakeTelemetry{},
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
			[]Component{{Name: "background", Runtime: component}},
			nil,
		)
		listener := newTestListener(t)
		process.listen = func(string, string) (net.Listener, error) { return listener, nil }
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() { result <- process.Run(ctx) }()
		<-component.started
		cancel()

		require.NoError(t, <-result, "iteration %d", iteration)
		require.Equal(t, platformhealth.PhaseDraining, readiness.Phase(), "iteration %d", iteration)
	}
}

func TestProcessReportsNonCancellationErrorDuringDrain(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	component := &drainErrorAPIComponent{
		started: make(chan struct{}),
		err:     errors.New("flush interrupted"),
	}
	process := newLifecycleTestProcess(
		t,
		readiness,
		&fakeTelemetry{},
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		[]Component{{Name: "background", Runtime: component}},
		nil,
	)
	listener := newTestListener(t)
	process.listen = func(string, string) (net.Listener, error) { return listener, nil }
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- process.Run(ctx) }()
	<-component.started

	cancel()

	err := <-result
	require.ErrorContains(t, err, "flush interrupted")
	require.Equal(t, platformhealth.PhaseDraining, readiness.Phase())
}

func TestProcessPropagatesTelemetryAndResourceCloseErrors(t *testing.T) {
	t.Parallel()

	readiness := newAPIReadiness(t)
	telemetry := &fakeTelemetry{err: errors.New("exporter flush failed")}
	process := newLifecycleTestProcess(t, readiness, telemetry, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), nil,
		[]Resource{{Name: "redis", Close: func() error { return errors.New("close failed") }}},
	)
	listener := newTestListener(t)
	process.listen = func(string, string) (net.Listener, error) { return listener, nil }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := process.Run(ctx)
	require.ErrorContains(t, err, "flush API telemetry")
	require.ErrorContains(t, err, "close API resource redis")
	require.Equal(t, int32(1), telemetry.shutdown.Load())
}

func TestProcessRunRejectsNilReceiver(t *testing.T) {
	t.Parallel()

	var process *Process
	require.ErrorContains(t, process.Run(context.Background()), "API process is required")
}

func TestProcessWaitForComponentsDrainsBufferedResultsAfterDeadline(t *testing.T) {
	t.Parallel()

	process := &Process{}
	results := make(chan lifecycleResult, 2)
	results <- lifecycleResult{name: "consumer"}
	results <- lifecycleResult{name: "sse", err: errors.New("subscription stopped")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := process.waitForComponents(ctx, results, 2)
	require.ErrorContains(t, err, "stop API component sse")
	require.ErrorContains(t, err, "subscription stopped")
	require.NotErrorIs(t, err, context.Canceled)
}

func newLifecycleTestProcess(
	t *testing.T,
	readiness *platformhealth.Readiness,
	telemetry TelemetryProvider,
	handler http.Handler,
	components []Component,
	resources []Resource,
) *Process {
	t.Helper()

	process, err := NewProcess(ProcessOptions{
		Config: ProcessConfig{
			Address:                  "127.0.0.1:0",
			ReadHeaderTimeout:        time.Second,
			ReadTimeout:              time.Second,
			WriteTimeout:             time.Second,
			IdleTimeout:              time.Second,
			ShutdownTimeout:          time.Second,
			TelemetryShutdownTimeout: time.Second,
		},
		Handler:    handler,
		Log:        logger.NewWithJSON(io.Discard, slog.LevelDebug, "api-lifecycle-test"),
		Readiness:  readiness,
		Telemetry:  telemetry,
		Components: components,
		Resources:  resources,
	})
	require.NoError(t, err)
	return process
}

func newAPIReadiness(t *testing.T) *platformhealth.Readiness {
	t.Helper()
	readiness, err := platformhealth.NewReadiness(time.Second)
	require.NoError(t, err)
	return readiness
}

func newTestListener(t *testing.T) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return listener
}
