package workerbootstrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

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

func TestWorkerMonitorIsDisabledByDefault(t *testing.T) {
	t.Parallel()

	handler, err := newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, MonitorConfig{}, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, workerMonitorPath, nil))

	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestWorkerMonitorRequiresBasicAuthentication(t *testing.T) {
	t.Parallel()

	monitor := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	config := MonitorConfig{
		Enabled:  true,
		Username: "operator",
		Password: "a-long-monitor-password-used-only-in-tests",
	}
	handler, err := newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, config, monitor)
	require.NoError(t, err)

	for name, credentials := range map[string][2]string{
		"missing":        {"", ""},
		"wrong username": {"intruder", config.Password},
		"wrong password": {config.Username, "incorrect"},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, workerMonitorPath, nil)
			request.RemoteAddr = "127.0.0.1:43100"
			username, password := credentials[0], credentials[1]
			if username != "" || password != "" {
				request.SetBasicAuth(username, password)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusUnauthorized, response.Code)
			require.NotEmpty(t, response.Header().Get("WWW-Authenticate"))
		})
	}

	for _, remoteAddress := range []string{"127.0.0.1:43100", "[::1]:43100"} {
		request := httptest.NewRequest(http.MethodGet, workerMonitorPath, nil)
		request.RemoteAddr = remoteAddress
		request.Header.Set("X-Forwarded-For", "203.0.113.10")
		request.Header.Set("X-Real-IP", "203.0.113.10")
		request.SetBasicAuth(config.Username, config.Password)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusNoContent, response.Code, remoteAddress)
		require.Equal(t, "DENY", response.Header().Get("X-Frame-Options"), remoteAddress)
	}
}

func TestWorkerMonitorRejectsNonLoopbackPeersBeforeAuthentication(t *testing.T) {
	t.Parallel()

	monitorCalled := false
	monitor := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		monitorCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	config := MonitorConfig{
		Enabled:  true,
		Username: "operator",
		Password: "a-long-monitor-password-used-only-in-tests",
	}
	handler, err := newWorkerHTTPHandler(&atomic.Bool{}, func(context.Context) error { return nil }, config, monitor)
	require.NoError(t, err)

	for _, remoteAddress := range []string{
		"203.0.113.10:43100",
		"10.0.2.15:43100",
		"127.0.0.1",
		"::1",
		"malformed-address",
	} {
		request := httptest.NewRequest(http.MethodGet, workerMonitorPath, nil)
		request.RemoteAddr = remoteAddress
		request.Header.Set("X-Forwarded-For", "127.0.0.1")
		request.Header.Set("X-Real-IP", "127.0.0.1")
		request.SetBasicAuth(config.Username, config.Password)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusNotFound, response.Code, remoteAddress)
		require.Empty(t, response.Header().Get("WWW-Authenticate"), remoteAddress)
	}
	require.False(t, monitorCalled)
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
