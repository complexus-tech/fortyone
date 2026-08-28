package workerbootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/hibiken/asynq"
)

type taskServer interface {
	Start(asynq.Handler) error
	Stop()
	Shutdown()
}

type taskScheduler interface {
	Start() error
	Shutdown()
}

func (a *App) Run(ctx context.Context) (runErr error) {
	if ctx == nil {
		return errors.New("worker context is required")
	}
	if err := a.validateRuntime(); err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, a.closeResources())
	}()

	handler, closeQueueMonitor, err := a.newHTTPHandler()
	if err != nil {
		return fmt.Errorf("initialize worker HTTP handler: %w", err)
	}
	defer func() {
		runErr = errors.Join(runErr, closeMonitor(closeQueueMonitor))
	}()

	listener, err := net.Listen("tcp", a.httpConfig.Host)
	if err != nil {
		return fmt.Errorf("listen for worker health traffic on %s: %w", a.httpConfig.Host, err)
	}
	listenerTransferred := false
	defer func() {
		if !listenerTransferred {
			runErr = errors.Join(runErr, listener.Close())
		}
	}()

	if err := ctx.Err(); err != nil {
		return nil
	}

	a.log.Info(ctx, "Starting worker scheduler")
	if err := a.scheduler.Start(); err != nil {
		return fmt.Errorf("start worker scheduler: %w", err)
	}
	schedulerStarted := true
	defer func() {
		if schedulerStarted {
			a.scheduler.Shutdown()
		}
	}()

	a.log.Info(ctx, "Starting Asynq worker server")
	if err := a.server.Start(a.taskMux); err != nil {
		return fmt.Errorf("start Asynq worker server: %w", err)
	}
	serverStarted := true
	defer func() {
		if serverStarted {
			a.server.Shutdown()
		}
	}()

	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: a.httpConfig.ReadHeaderTimeout,
		ReadTimeout:       a.httpConfig.ReadTimeout,
		WriteTimeout:      a.httpConfig.WriteTimeout,
		IdleTimeout:       a.httpConfig.IdleTimeout,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	httpResult := make(chan error, 1)
	go func() {
		err := httpServer.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		httpResult <- err
	}()
	listenerTransferred = true
	a.ready.Store(true)
	a.log.Info(ctx, "Worker is ready", "health_address", listener.Addr().String())

	var (
		serveErr           error
		httpResultReceived bool
	)
	select {
	case <-ctx.Done():
	case serveErr = <-httpResult:
		httpResultReceived = true
		if serveErr != nil {
			runErr = fmt.Errorf("serve worker health traffic: %w", serveErr)
		}
	}

	a.ready.Store(false)
	a.log.Info(context.Background(), "Worker is draining")

	a.server.Stop()
	a.scheduler.Shutdown()
	schedulerStarted = false

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.httpConfig.ShutdownTimeout)
	defer cancel()
	httpShutdownResult := make(chan error, 1)
	go func() {
		httpShutdownResult <- httpServer.Shutdown(shutdownCtx)
	}()
	a.server.Shutdown()
	serverStarted = false
	shutdownErr := <-httpShutdownResult
	if shutdownErr != nil {
		if closeErr := httpServer.Close(); closeErr != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("force close worker HTTP server: %w", closeErr))
		}
	}
	if !httpResultReceived {
		serveErr = <-httpResult
	}
	if serveErr != nil {
		serveErr = fmt.Errorf("serve worker health traffic: %w", serveErr)
	}
	if shutdownErr != nil {
		shutdownErr = fmt.Errorf("shut down worker HTTP server: %w", shutdownErr)
	}

	return errors.Join(runErr, serveErr, shutdownErr)
}

func (a *App) validateRuntime() error {
	if a == nil {
		return errors.New("worker application is required")
	}
	if a.log == nil {
		return errors.New("worker logger is required")
	}
	if a.server == nil {
		return errors.New("worker task server is required")
	}
	if a.scheduler == nil {
		return errors.New("worker scheduler is required")
	}
	if a.taskMux == nil {
		return errors.New("worker task handler is required")
	}
	if a.ready == nil {
		return errors.New("worker readiness state is required")
	}
	if a.pingRedis == nil {
		return errors.New("worker Redis health check is required")
	}
	if err := validateHTTPConfig(a.httpConfig, a.monitorConfig); err != nil {
		return fmt.Errorf("validate worker runtime: %w", err)
	}
	return nil
}

func (a *App) closeResources() error {
	var closeErrors []error
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close worker Redis client: %w", err))
		}
	}
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close worker database: %w", err))
		}
	}
	return errors.Join(closeErrors...)
}
