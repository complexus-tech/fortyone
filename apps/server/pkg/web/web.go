package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Handler is the signature used by all application handlers in this service.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the main application handler that manages routing, middleware, and shutdown.
// It wraps the standard http.ServeMux with additional functionality for middleware
// composition, graceful shutdown, and OpenTelemetry tracing.
type App struct {
	mux         *http.ServeMux
	mw          []Middleware
	shutdown    chan os.Signal
	tracer      oteltrace.Tracer
	strictSlash bool
	origins     OriginPolicy
	log         *logger.Logger
}

// New creates an application struct that will handle all requests to the application.
func New(shutdown chan os.Signal, tracer oteltrace.Tracer, mw ...Middleware) *App {
	mux := http.NewServeMux()
	return &App{
		mux:         mux,
		mw:          mw,
		shutdown:    shutdown,
		tracer:      tracer,
		strictSlash: true,
	}
}

// ServeHTTP implements the http.Handler interface so that App can be used as a Mux.
// It then calls the ServeHTTP method on the embedded Mux.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// extract this into a configurable middleware
	if origin := a.origins.AllowedOrigin(r); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Retry-After, RateLimit-Policy, RateLimit")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Idempotency-Key, Mcp-Protocol-Version, Mcp-Session-Id, Last-Event-ID")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !a.strictSlash {
		path := a.stripSlash(r.URL.Path)
		r.URL.Path = path
	}
	a.mux.ServeHTTP(w, r)
}

// SetOriginPolicy configures the credentialed CORS allowlist before the
// application starts serving requests.
func (a *App) SetOriginPolicy(policy OriginPolicy) {
	a.origins = policy
}

// SetLogger configures the request-bound structured logger used for unexpected
// handler failures. The raw error is deliberately not logged here: this is the
// final transport boundary and errors can contain SQL, credentials, or provider
// payloads. Lower layers may log a safe, classified cause with domain context.
func (a *App) SetLogger(log *logger.Logger) {
	a.log = log
}

// StripSlash will remove the trailing slash from the URL.
func (a *App) StrictSlash(strictSlash bool) {
	a.strictSlash = strictSlash
}

// StripSlash will remove the trailing slash from the URL.
func (a *App) stripSlash(path string) string {
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

// Handle will will apply middleware to the handler and then add it to the mux router.
func (a *App) Handle(method string, pattern string, handler Handler, mw ...Middleware) {
	// First handler to execute is the one passed in.
	handler = wrapMiddleware(mw, handler)
	// Then wrap with the application level middleware.
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		ctx, span := a.startSpan(w, r)
		defer span.End()
		requestID := uuid.NewString()
		w.Header().Set("X-Request-ID", requestID)
		ctx = logger.WithRequestID(ctx, requestID)

		v := &Values{
			TraceID:   span.SpanContext().TraceID().String(),
			RequestID: requestID,
			Tracer:    a.tracer,
			Now:       time.Now(),
		}

		ctx = SetValues(ctx, v)

		if err := handler(ctx, w, r); err != nil {
			span.SetStatus(codes.Error, "request handler failed")
			if a.log != nil {
				a.log.Error(ctx, "request handler failed", "status_code", http.StatusInternalServerError)
			}
			_ = RespondError(ctx, w, err, http.StatusInternalServerError)
			return
		}

	}

	// Add this handler to the mux router.
	routePattern := pattern
	if method != "" {
		routePattern = fmt.Sprintf("%s %s", method, pattern)
	}
	a.mux.HandleFunc(routePattern, h)
}

// NotFound registers a JSON fallback for requests that do not match a more
// specific route. It should be registered after all application routes.
func (a *App) NotFound(handler Handler) {
	a.Handle("", "/", handler)
}

// Get is a shortcut for app.Handle(http.MethodGet, path, handler, mw...)
func (a *App) Get(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodGet, path, handler, mw...)
}

// Post is a shortcut for app.Handle(http.MethodPost, path, handler, mw...)
func (a *App) Post(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPost, path, handler, mw...)
}

// Put is a shortcut for app.Handle(http.MethodPut, path, handler, mw...)
func (a *App) Put(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPut, path, handler, mw...)
}

// Delete is a shortcut for app.Handle(http.MethodDelete, path, handler, mw...)
func (a *App) Delete(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodDelete, path, handler, mw...)
}

// Patch is a shortcut for app.Handle(http.MethodPatch, path, handler, mw...)
func (a *App) Patch(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPatch, path, handler, mw...)
}

// Shutdown will gracefully shutdown the application.
func (a *App) Shutdown() {
	a.shutdown <- syscall.SIGTERM
}

// startSpan will start a span for the request and add it to the context.
func (a *App) startSpan(w http.ResponseWriter, r *http.Request) (context.Context, oteltrace.Span) {
	ctx := r.Context()

	endpoint := strings.TrimSpace(r.Pattern)
	if endpoint == "" {
		endpoint = "unmatched"
	}
	ctx, span := apptracing.StartHTTPSpan(ctx, a.tracer, "pkg.web.handle", r.Method, endpoint)
	apptracing.InjectTraceHeaders(ctx, w.Header())
	return ctx, span
}
