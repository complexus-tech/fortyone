package workerbootstrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestWorkerLivenessDoesNotDependOnReadiness(t *testing.T) {
	t.Parallel()

	ready := &atomic.Bool{}
	handler, err := newWorkerHTTPHandler(ready, func(context.Context) error {
		return errors.New("Redis is unavailable")
	}, MonitorConfig{}, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"alive"}`, response.Body.String())
}

func TestWorkerReadinessReflectsLifecycleAndRedis(t *testing.T) {
	t.Parallel()

	ready := &atomic.Bool{}
	redisErr := errors.New("Redis is unavailable")
	handler, err := newWorkerHTTPHandler(ready, func(context.Context) error {
		return redisErr
	}, MonitorConfig{}, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.JSONEq(t, `{"status":"not_ready"}`, response.Body.String())

	ready.Store(true)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.JSONEq(t, `{"status":"dependency_unavailable"}`, response.Body.String())
}

func TestWorkerReadinessSucceedsOnlyWhenLifecycleAndRedisAreHealthy(t *testing.T) {
	t.Parallel()

	ready := &atomic.Bool{}
	ready.Store(true)
	handler, err := newWorkerHTTPHandler(ready, func(context.Context) error { return nil }, MonitorConfig{}, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"ready"}`, response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}

func TestWorkerMonitorCanBeDisabled(t *testing.T) {
	t.Parallel()

	handler, err := newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, MonitorConfig{}, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, workerMonitorPath, nil))

	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestWorkerMonitorEnablesQueueManagement(t *testing.T) {
	t.Parallel()
	app := App{
		ready:         &atomic.Bool{},
		pingRedis:     func(context.Context) error { return nil },
		monitorConfig: MonitorConfig{Enabled: true},
		redisOpt:      asynq.RedisClientOpt{Addr: "127.0.0.1:1"},
	}
	handler, close, err := app.newHTTPHandler()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, close()) })

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, page.Code)
	require.Contains(t, page.Body.String(), `window.FLAG_READ_ONLY="false"`)

	// Invalid JSON is rejected before Redis access. A read-only monitor would
	// reject the request with 403 before the mutation handler can validate it.
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/queues/test/pending_tasks:batch_delete", strings.NewReader("invalid json")))
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestWorkerMonitorRoutesThroughLoadBalancer(t *testing.T) {
	t.Parallel()

	config := MonitorConfig{
		Enabled: true,
	}
	monitor := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	})
	handler, err := newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, config, monitor)
	require.NoError(t, err)

	for _, path := range []string{"/", "/static/js/main.js", "/api/queues"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "https://worker.fortyone.app"+path, nil)
			request.RemoteAddr = "10.0.2.15:43100"
			request.Header.Set("X-Forwarded-For", "203.0.113.10")
			request.Header.Set("X-Amzn-Oidc-Identity", "operator")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, path, response.Body.String())
			require.Empty(t, response.Header().Get("WWW-Authenticate"))
			require.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
		})
	}
}

func TestWorkerHealthChecksRemainAvailableWithMonitorEnabled(t *testing.T) {
	t.Parallel()

	ready := &atomic.Bool{}
	ready.Store(true)
	monitor := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("health check reached the monitor")
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler, err := newWorkerHTTPHandler(ready, func(context.Context) error { return nil }, MonitorConfig{
		Enabled: true,
	}, monitor)
	require.NoError(t, err)

	for _, path := range []string{"/health/live", "/health/ready"} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			request := httptest.NewRequest(method, path, nil)
			request.RemoteAddr = "10.0.2.15:43100"
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, "%s %s", method, path)
		}
	}
}

func TestWorkerHTTPHandlerRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	_, err := newWorkerHTTPHandler(nil, func(context.Context) error { return nil }, MonitorConfig{}, nil)
	require.ErrorContains(t, err, "readiness state")

	_, err = newWorkerHTTPHandler(&atomic.Bool{}, nil, MonitorConfig{}, nil)
	require.ErrorContains(t, err, "Redis health check")

	_, err = newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, MonitorConfig{Enabled: true}, nil)
	require.ErrorContains(t, err, "monitor handler")
}
