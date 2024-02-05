package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Handler is the signature used by all application handlers in this service.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type app struct {
	mux         *http.ServeMux
	mw          []Middleware
	shutdown    chan os.Signal
	tracer      trace.Tracer
	strictSlash bool
}

// New creates an application struct that will handle all requests to the application.
func New(shutdown chan os.Signal, tracer trace.Tracer, mw ...Middleware) *app {
	// mux := mux.NewRouter()
	mux := http.NewServeMux()
	return &app{
		mux:         mux,
		mw:          mw,
		shutdown:    shutdown,
		tracer:      tracer,
		strictSlash: true,
	}
}

// ServeHTTP implements the http.Handler interface so that App can be used as a Mux.
// It then calls the ServeHTTP method on the embedded Mux.
func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !a.strictSlash {
		path := a.stripSlash(r.URL.Path)
		r.URL.Path = path
	}
	a.mux.ServeHTTP(w, r)
}

// StripSlash will remove the trailing slash from the URL.
func (a *app) StrictSlash(strictSlash bool) {
	a.strictSlash = strictSlash
}

// StripSlash will remove the trailing slash from the URL.
func (a *app) stripSlash(path string) string {
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

// Handle will will apply middleware to the handler and then add it to the mux router.
func (a *app) Handle(method string, pattern string, handler Handler, mw ...Middleware) {
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
			a.Shutdown()

			return
		}

	}

	// Add this handler to the mux router.
	a.mux.HandleFunc(fmt.Sprintf("%s %s", method, pattern), h)
}

// Get is a shortcut for app.Handle(http.MethodGet, path, handler, mw...)
func (a *app) Get(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodGet, path, handler, mw...)
}

// Post is a shortcut for app.Handle(http.MethodPost, path, handler, mw...)
func (a *app) Post(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPost, path, handler, mw...)
}

// Put is a shortcut for app.Handle(http.MethodPut, path, handler, mw...)
func (a *app) Put(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPut, path, handler, mw...)
}

// Delete is a shortcut for app.Handle(http.MethodDelete, path, handler, mw...)
func (a *app) Delete(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodDelete, path, handler, mw...)
}

// Patch is a shortcut for app.Handle(http.MethodPatch, path, handler, mw...)
func (a *app) Patch(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPatch, path, handler, mw...)
}

// Shutdown will gracefully shutdown the application.
func (a *app) Shutdown() {
	a.shutdown <- syscall.SIGTERM
}

// startSpan will start a span for the request and add it to the context.
func (a *app) startSpan(w http.ResponseWriter, r *http.Request) (context.Context, trace.Span) {
	ctx := r.Context()

	span := trace.SpanFromContext(ctx)

	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "pkg.web.handle")
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.endpoint", r.RequestURI))
	}
	// Inject the span context into the request header.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
	return ctx, span
}
