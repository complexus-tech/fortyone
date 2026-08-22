package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestNotFoundUsesRegisteredFallback(t *testing.T) {
	t.Parallel()

	app := New(make(chan os.Signal, 1), trace.NewNoopTracerProvider().Tracer("test"))
	app.Get("/known", func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return Respond(ctx, w, "ok", http.StatusOK)
	})
	app.NotFound(func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return RespondError(ctx, w, errors.New("API route not found"), http.StatusNotFound)
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), `"code":"not_found"`)
}
