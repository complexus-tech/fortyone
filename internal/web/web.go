package web

import (
	"context"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Handler is the signature used by all application handlers in this service.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type application struct {
	mux         *mux.Router
	mw          []Middleware
	shutdown    chan os.Signal
	tracer      trace.Tracer
	strictSlash bool
}

func NewApp(shutdown chan os.Signal, tracer trace.Tracer, mw ...Middleware) *application {
	mux := mux.NewRouter()
	return &application{
		mux:         mux,
		mw:          mw,
		shutdown:    shutdown,
		tracer:      tracer,
		strictSlash: true,
	}
}

// ServeHTTP implements the http.Handler interface so that App can be used as a Mux.
// It then calls the ServeHTTP method on the embedded Mux.
func (a *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !a.strictSlash {
		path := a.stripSlash(r.URL.Path)
		r.URL.Path = path
	}
	a.mux.ServeHTTP(w, r)
}

func (a *application) StrictSlash(strictSlash bool) {
	a.strictSlash = strictSlash
}

func (a *application) stripSlash(path string) string {
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

// Handle will will apply middleware to the handler and then add it to the mux router.
func (a *application) Handle(method string, pattern string, handler Handler, mw ...Middleware) {
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
	if strings.ToLower(method) == "all" {
		a.mux.HandleFunc(pattern, h)
	} else {
		a.mux.HandleFunc(pattern, h).Methods(method)
	}
}

// Get is a shortcut for app.Handle(http.MethodGet, path, handler, mw...)
func (a *application) Get(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodGet, path, handler, mw...)
}

// Post is a shortcut for app.Handle(http.MethodPost, path, handler, mw...)
func (a *application) Post(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPost, path, handler, mw...)
}

// Put is a shortcut for app.Handle(http.MethodPut, path, handler, mw...)
func (a *application) Put(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPut, path, handler, mw...)
}

// Delete is a shortcut for app.Handle(http.MethodDelete, path, handler, mw...)
func (a *application) Delete(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodDelete, path, handler, mw...)
}

// Patch is a shortcut for app.Handle(http.MethodPatch, path, handler, mw...)
func (a *application) Patch(path string, handler Handler, mw ...Middleware) {
	a.Handle(http.MethodPatch, path, handler, mw...)
}

// Shutdown will gracefully shutdown the application.
func (a *application) Shutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *application) startSpan(w http.ResponseWriter, r *http.Request) (context.Context, trace.Span) {
	ctx := r.Context()

	span := trace.SpanFromContext(ctx)

	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "internal.web.handle")
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.endpoint", r.RequestURI))
	}
	// Inject the span context into the request header.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
	return ctx, span
}
