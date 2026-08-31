package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

type sensitiveHandlerError struct {
	secret string
}

func (e sensitiveHandlerError) Error() string {
	return "provider failed with bearer " + e.secret
}

func TestNotFoundUsesRegisteredFallback(t *testing.T) {
	t.Parallel()

	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
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

func TestAppCorrelatesErrorEnvelopeAndResponseHeader(t *testing.T) {
	t.Parallel()

	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.Get("/failure", func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return RespondError(ctx, w, errors.New("denied"), http.StatusForbidden)
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/failure", nil))

	requestID := recorder.Header().Get("X-Request-ID")
	require.NotEmpty(t, requestID)
	var response Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, requestID, response.Error.RequestID)
}

func TestAppDoesNotLogUnexpectedHandlerErrorContents(t *testing.T) {
	t.Parallel()

	const secret = "fortyone-sensitive-provider-token"
	var logs bytes.Buffer
	log := logger.NewWithJSON(&logs, slog.LevelDebug, "web-test")
	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.SetLogger(log)
	app.Get("/failure", func(context.Context, http.ResponseWriter, *http.Request) error {
		return sensitiveHandlerError{secret: secret}
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/failure", nil))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.NotContains(t, recorder.Body.String(), secret)
	require.NotContains(t, logs.String(), secret)
	require.Contains(t, logs.String(), `"msg":"request handler failed"`)
	require.Contains(t, logs.String(), recorder.Header().Get("X-Request-ID"))
}

func TestAppRecoversHandlerPanicWithoutLeakingItsValue(t *testing.T) {
	t.Parallel()

	const secret = "fortyone-sensitive-panic-value"
	var logs bytes.Buffer
	log := logger.NewWithJSON(&logs, slog.LevelDebug, "web-test")
	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.SetLogger(log)
	app.Get("/panic", func(context.Context, http.ResponseWriter, *http.Request) error {
		panic(secret)
	})
	app.Get("/healthy", func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return Respond(ctx, w, "ok", http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	requestID := recorder.Header().Get("X-Request-ID")
	require.NotEmpty(t, requestID)
	var response Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.Error)
	require.Equal(t, "internal_error", response.Error.Code)
	require.Equal(t, "internal server error", response.Error.Message)
	require.Equal(t, requestID, response.Error.RequestID)
	require.NotContains(t, recorder.Body.String(), secret)
	require.NotContains(t, logs.String(), secret)
	require.Contains(t, logs.String(), `"msg":"request handler panicked"`)
	require.Contains(t, logs.String(), requestID)

	healthyRecorder := httptest.NewRecorder()
	app.ServeHTTP(healthyRecorder, httptest.NewRequest(http.MethodGet, "/healthy", nil))
	require.Equal(t, http.StatusOK, healthyRecorder.Code)
	require.Contains(t, healthyRecorder.Body.String(), `"ok"`)
}

func TestAppDoesNotAppendErrorAfterPanickingHandlerCommitsResponse(t *testing.T) {
	t.Parallel()

	const secret = "fortyone-sensitive-partial-response-panic"
	var logs bytes.Buffer
	log := logger.NewWithJSON(&logs, slog.LevelDebug, "web-test")
	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.SetLogger(log)
	app.Get("/partial", func(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("partial response"))
		panic(secret)
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/partial", nil))

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "partial response", recorder.Body.String())
	require.NotContains(t, logs.String(), secret)
	require.Contains(t, logs.String(), `"msg":"request handler panicked"`)
	require.Contains(t, logs.String(), `"response_started":true`)
}

func TestAppDoesNotAppendErrorAfterHandlerReturnsErrorFromCommittedResponse(t *testing.T) {
	t.Parallel()

	const secret = "fortyone-sensitive-partial-response-error"
	var logs bytes.Buffer
	log := logger.NewWithJSON(&logs, slog.LevelDebug, "web-test")
	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.SetLogger(log)
	app.Get("/partial-error", func(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("partial response"))
		return sensitiveHandlerError{secret: secret}
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/partial-error", nil))

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, "partial response", recorder.Body.String())
	require.NotContains(t, logs.String(), secret)
	require.Contains(t, logs.String(), `"msg":"request handler failed"`)
	require.Contains(t, logs.String(), `"response_started":true`)
}

func TestAppPanicBoundaryPreservesStreamingSupport(t *testing.T) {
	t.Parallel()

	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.Get("/stream", func(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
		flusher, ok := w.(http.Flusher)
		if !ok {
			return errors.New("wrapped response writer does not support flushing")
		}
		_, err := w.Write([]byte("data: ready\n\n"))
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/stream", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "data: ready\n\n", recorder.Body.String())
	require.True(t, recorder.Flushed)
}

func TestAppAppliesCredentialedSubdomainOriginPolicy(t *testing.T) {
	t.Parallel()

	policy, err := NewOriginPolicy("https://*.fortyone.app")
	require.NoError(t, err)
	app := New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("test"))
	app.SetOriginPolicy(policy)
	app.Get("/ok", func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return Respond(ctx, w, map[string]bool{"ok": true}, http.StatusOK)
	})

	for _, test := range []struct {
		origin string
		want   string
	}{
		{origin: "https://complexus.fortyone.app", want: "https://complexus.fortyone.app"},
		{origin: "https://fortyone.app.attacker.example"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/ok", nil)
		request.Header.Set("Origin", test.origin)
		app.ServeHTTP(recorder, request)
		require.Equal(t, test.want, recorder.Header().Get("Access-Control-Allow-Origin"))
		if test.want != "" {
			require.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
			require.Contains(t, recorder.Header().Get("Access-Control-Expose-Headers"), "X-Request-ID")
			require.Contains(t, recorder.Header().Get("Access-Control-Expose-Headers"), "RateLimit-Policy")
			require.Contains(t, recorder.Header().Get("Access-Control-Expose-Headers"), "RateLimit")
			require.Contains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "Idempotency-Key")
		}
	}
}
