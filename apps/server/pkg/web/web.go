package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
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
	if origin := AllowedOrigin(r); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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

		v := &Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now(),
		}

		ctx = SetValues(ctx, v)

		if err := handler(ctx, w, r); err != nil {
			log.Print(err)
			Respond(ctx, w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	// Add this handler to the mux router.
	a.mux.HandleFunc(fmt.Sprintf("%s %s", method, pattern), h)
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

	ctx, span := apptracing.StartHTTPSpan(ctx, a.tracer, "pkg.web.handle", r.Method, r.RequestURI)
	apptracing.InjectTraceHeaders(ctx, w.Header())
	return ctx, span
}
