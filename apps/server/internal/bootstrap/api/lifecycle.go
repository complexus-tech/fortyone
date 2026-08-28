package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const defaultMaxHeaderBytes = 1 << 20

// LifecycleComponent is a long-lived API capability owned by Process. An
// implementation must return promptly from Run after ctx is cancelled.
type LifecycleComponent interface {
	Initialize(context.Context) error
	Run(context.Context) error
}

// TelemetryProvider is the bounded shutdown contract used by OpenTelemetry.
type TelemetryProvider interface {
	Shutdown(context.Context) error
}

// Component assigns an operator-safe name to a supervised runtime capability.
type Component struct {
	Name    string
	Runtime LifecycleComponent
}

// Resource is process-owned infrastructure closed after all request and
// background work has stopped. Resources are closed in the declared order.
type Resource struct {
	Name  string
	Close func() error
}

// ProcessConfig controls the externally visible HTTP server and shutdown
// budgets. Long-lived SSE responses clear their own write deadline.
type ProcessConfig struct {
	Address                  string
	ReadHeaderTimeout        time.Duration
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	IdleTimeout              time.Duration
	ShutdownTimeout          time.Duration
	TelemetryShutdownTimeout time.Duration
}

// ProcessOptions contains every dependency whose lifecycle the API process
// owns. Construction validates the complete graph before any goroutine starts.
type ProcessOptions struct {
	Config     ProcessConfig
	Handler    http.Handler
	Log        *logger.Logger
	Readiness  *platformhealth.Readiness
	Telemetry  TelemetryProvider
	Components []Component
	Resources  []Resource
}

// Process supervises the HTTP server, SSE hub, Redis stream consumer, tracing,
// and process-owned dependencies under one root context.
type Process struct {
	options ProcessOptions
	listen  func(network, address string) (net.Listener, error)

	runStarted atomic.Bool
	closeOnce  sync.Once
	closeErr   error
}

// NewProcess validates and constructs an API process supervisor.
func NewProcess(options ProcessOptions) (*Process, error) {
	process := &Process{
		options: options,
		listen:  net.Listen,
	}
	if err := process.validate(); err != nil {
		return nil, err
	}
	return process, nil
}

// Run starts every component, exposes readiness, and blocks until cancellation
// or a terminal runtime error. It may be called exactly once.
func (p *Process) Run(ctx context.Context) (runErr error) {
	if p == nil {
		return errors.New("API process is required")
	}
	if ctx == nil {
		return errors.New("API process context is required")
	}
	if !p.runStarted.CompareAndSwap(false, true) {
		return errors.New("API process can only run once")
	}
	defer func() {
		runErr = errors.Join(runErr, p.closeOwnedResources())
	}()

	for _, component := range p.options.Components {
		if err := component.Runtime.Initialize(ctx); err != nil {
			if ctx.Err() != nil {
				p.options.Readiness.MarkDraining()
				return nil
			}
			p.options.Readiness.MarkFailed()
			return fmt.Errorf("initialize API component %s: %w", component.Name, err)
		}
	}

	listener, err := p.listen("tcp", p.options.Config.Address)
	if err != nil {
		p.options.Readiness.MarkFailed()
		return fmt.Errorf("listen for API traffic on %s: %w", p.options.Config.Address, err)
	}
	listenerTransferred := false
	defer func() {
		if !listenerTransferred {
			runErr = errors.Join(runErr, listener.Close())
		}
	}()

	if ctx.Err() != nil {
		p.options.Readiness.MarkDraining()
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	httpServer := &http.Server{
		Handler:           p.options.Handler,
		ReadHeaderTimeout: p.options.Config.ReadHeaderTimeout,
		ReadTimeout:       p.options.Config.ReadTimeout,
		WriteTimeout:      p.options.Config.WriteTimeout,
		IdleTimeout:       p.options.Config.IdleTimeout,
		MaxHeaderBytes:    defaultMaxHeaderBytes,
		BaseContext: func(net.Listener) context.Context {
			return runCtx
		},
	}

	resultCount := len(p.options.Components) + 1
	results := make(chan lifecycleResult, resultCount)
	for _, component := range p.options.Components {
		component := component
		go func() {
			results <- lifecycleResult{name: component.Name, err: component.Runtime.Run(runCtx)}
		}()
	}
	go func() {
		err := httpServer.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		results <- lifecycleResult{name: "http", err: err}
	}()
	listenerTransferred = true
	p.options.Readiness.MarkReady()
	p.options.Log.Info(ctx, "API is ready", "address", listener.Addr().String())

	var (
		firstResult *lifecycleResult
		terminalErr error
	)
	select {
	case <-ctx.Done():
		p.options.Readiness.MarkDraining()
		p.options.Log.Info(context.Background(), "API is draining")
	case result := <-results:
		firstResult = &result
		if ctx.Err() != nil {
			p.options.Readiness.MarkDraining()
			p.options.Log.Info(context.Background(), "API is draining")
			if result.err != nil && !errors.Is(result.err, context.Canceled) && !errors.Is(result.err, context.DeadlineExceeded) {
				terminalErr = fmt.Errorf("stop API component %s: %w", result.name, result.err)
			}
		} else {
			p.options.Readiness.MarkFailed()
			terminalErr = unexpectedLifecycleExit(result)
			p.options.Log.Error(context.Background(), "API component stopped", "component", result.name, "error", terminalErr)
		}
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), p.options.Config.ShutdownTimeout)
	defer shutdownCancel()
	shutdownErr := p.shutdownHTTP(shutdownCtx, httpServer)
	remaining := resultCount
	if firstResult != nil {
		remaining--
	}
	componentErr := p.waitForComponents(shutdownCtx, results, remaining)
	return errors.Join(terminalErr, shutdownErr, componentErr)
}

type lifecycleResult struct {
	name string
	err  error
}

func (p *Process) shutdownHTTP(ctx context.Context, server *http.Server) error {
	if err := server.Shutdown(ctx); err != nil {
		closeErr := server.Close()
		return errors.Join(fmt.Errorf("gracefully shut down API HTTP server: %w", err), closeErr)
	}
	return nil
}

func (p *Process) waitForComponents(ctx context.Context, results <-chan lifecycleResult, remaining int) error {
	if remaining == 0 {
		return nil
	}

	var componentErrors []error
	recordResult := func(result lifecycleResult) {
		if result.err != nil && !errors.Is(result.err, context.Canceled) && !errors.Is(result.err, context.DeadlineExceeded) {
			componentErrors = append(componentErrors, fmt.Errorf("stop API component %s: %w", result.name, result.err))
		}
	}
	for remaining > 0 {
		// Prefer results which are already available over a simultaneously expired
		// shutdown deadline. This avoids reporting a false timeout after every
		// component has already stopped and published its terminal result.
		select {
		case result := <-results:
			remaining--
			recordResult(result)
			continue
		default:
		}

		select {
		case result := <-results:
			remaining--
			recordResult(result)
		case <-ctx.Done():
			for remaining > 0 {
				select {
				case result := <-results:
					remaining--
					recordResult(result)
				default:
					componentErrors = append(componentErrors, fmt.Errorf("wait for %d API components: %w", remaining, ctx.Err()))
					return errors.Join(componentErrors...)
				}
			}
		}
	}
	return errors.Join(componentErrors...)
}

func unexpectedLifecycleExit(result lifecycleResult) error {
	if result.err == nil {
		return fmt.Errorf("API component %s stopped unexpectedly", result.name)
	}
	return fmt.Errorf("API component %s failed: %w", result.name, result.err)
}

func (p *Process) closeOwnedResources() error {
	p.closeOnce.Do(func() {
		var closeErrors []error
		telemetryCtx, cancel := context.WithTimeout(context.Background(), p.options.Config.TelemetryShutdownTimeout)
		if err := p.options.Telemetry.Shutdown(telemetryCtx); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("flush API telemetry: %w", err))
		}
		cancel()

		for _, resource := range p.options.Resources {
			if err := resource.Close(); err != nil {
				closeErrors = append(closeErrors, fmt.Errorf("close API resource %s: %w", resource.Name, err))
			}
		}
		p.closeErr = errors.Join(closeErrors...)
	})
	return p.closeErr
}

func (p *Process) validate() error {
	if p == nil {
		return errors.New("API process is required")
	}
	if strings.TrimSpace(p.options.Config.Address) == "" {
		return errors.New("API HTTP address is required")
	}
	if p.options.Config.ReadHeaderTimeout <= 0 {
		return errors.New("API read-header timeout must be positive")
	}
	if p.options.Config.ReadTimeout <= 0 {
		return errors.New("API read timeout must be positive")
	}
	if p.options.Config.WriteTimeout <= 0 {
		return errors.New("API write timeout must be positive")
	}
	if p.options.Config.IdleTimeout <= 0 {
		return errors.New("API idle timeout must be positive")
	}
	if p.options.Config.ShutdownTimeout <= 0 {
		return errors.New("API shutdown timeout must be positive")
	}
	if p.options.Config.TelemetryShutdownTimeout <= 0 {
		return errors.New("API telemetry shutdown timeout must be positive")
	}
	if p.options.Handler == nil {
		return errors.New("API HTTP handler is required")
	}
	if p.options.Log == nil {
		return errors.New("API logger is required")
	}
	if p.options.Readiness == nil {
		return errors.New("API readiness state is required")
	}
	if p.options.Telemetry == nil {
		return errors.New("API telemetry provider is required")
	}

	componentNames := make(map[string]struct{}, len(p.options.Components))
	for _, component := range p.options.Components {
		name := strings.TrimSpace(component.Name)
		if name == "" {
			return errors.New("API component name is required")
		}
		if component.Runtime == nil {
			return fmt.Errorf("API component %s has no runtime", name)
		}
		if _, exists := componentNames[name]; exists {
			return fmt.Errorf("API component %s is registered more than once", name)
		}
		componentNames[name] = struct{}{}
	}

	resourceNames := make(map[string]struct{}, len(p.options.Resources))
	for _, resource := range p.options.Resources {
		name := strings.TrimSpace(resource.Name)
		if name == "" {
			return errors.New("API resource name is required")
		}
		if resource.Close == nil {
			return fmt.Errorf("API resource %s has no close function", name)
		}
		if _, exists := resourceNames[name]; exists {
			return fmt.Errorf("API resource %s is registered more than once", name)
		}
		resourceNames[name] = struct{}{}
	}
	return nil
}
