package workerbootstrap

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/hibiken/asynqmon"
)

const (
	workerMonitorPath     = "/admin/queues"
	readinessCheckTimeout = 2 * time.Second
)

type redisPingFunc func(context.Context) error

type monitorCloser func() error

type healthResponse struct {
	Status string `json:"status"`
}

func (a *App) newHTTPHandler() (http.Handler, monitorCloser, error) {
	var (
		monitor http.Handler
		close   monitorCloser
	)
	if a.monitorConfig.Enabled {
		queueMonitor := asynqmon.New(asynqmon.Options{
			RootPath:     workerMonitorPath,
			RedisConnOpt: a.redisOpt,
			ReadOnly:     true,
		})
		monitor = queueMonitor
		close = queueMonitor.Close
	}

	handler, err := newWorkerHTTPHandler(a.ready, a.pingRedis, a.monitorConfig, monitor)
	if err != nil {
		if close != nil {
			_ = close()
		}
		return nil, nil, err
	}
	return handler, close, nil
}

func newWorkerHTTPHandler(
	ready *atomic.Bool,
	pingRedis redisPingFunc,
	monitorConfig MonitorConfig,
	monitor http.Handler,
) (http.Handler, error) {
	if ready == nil {
		return nil, errors.New("worker readiness state is required")
	}
	if pingRedis == nil {
		return nil, errors.New("worker Redis health check is required")
	}
	if monitorConfig.Enabled && monitor == nil {
		return nil, errors.New("worker queue monitor handler is required when monitoring is enabled")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", livenessHandler)
	mux.HandleFunc("HEAD /health/live", livenessHandler)
	mux.HandleFunc("GET /health/ready", readinessHandler(ready, pingRedis))
	mux.HandleFunc("HEAD /health/ready", readinessHandler(ready, pingRedis))

	if monitorConfig.Enabled {
		// The queue console is an operator-only surface. Keep its network
		// boundary independent from the health endpoints on this listener: an
		// operator must reach it through a local tunnel/exec session and then
		// authenticate. Forwarded headers are intentionally ignored because a
		// direct client can spoof them when no trusted-proxy boundary is present.
		protectedMonitor := loopbackOnly(basicAuth(monitorConfig, monitor))
		mux.Handle(workerMonitorPath, protectedMonitor)
		mux.Handle(workerMonitorPath+"/", protectedMonitor)
	}

	return securityHeaders(mux), nil
}

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	writeHealthResponse(w, http.StatusOK, "alive")
}

func readinessHandler(ready *atomic.Bool, pingRedis redisPingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			writeHealthResponse(w, http.StatusServiceUnavailable, "not_ready")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readinessCheckTimeout)
		defer cancel()
		if err := pingRedis(ctx); err != nil {
			writeHealthResponse(w, http.StatusServiceUnavailable, "dependency_unavailable")
			return
		}

		writeHealthResponse(w, http.StatusOK, "ready")
	}
}

func writeHealthResponse(w http.ResponseWriter, status int, state string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: state})
}

func basicAuth(config MonitorConfig, next http.Handler) http.Handler {
	expectedUsername := sha256.Sum256([]byte(config.Username))
	expectedPassword := sha256.Sum256([]byte(config.Password))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		actualUsername := sha256.Sum256([]byte(username))
		actualPassword := sha256.Sum256([]byte(password))
		usernameMatches := subtle.ConstantTimeCompare(actualUsername[:], expectedUsername[:])
		passwordMatches := subtle.ConstantTimeCompare(actualPassword[:], expectedPassword[:])
		if !ok || usernameMatches != 1 || passwordMatches != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="FortyOne worker monitor", charset="UTF-8"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ip := net.ParseIP(strings.TrimSpace(host))
		if ip == nil || !ip.IsLoopback() {
			// Return Not Found instead of advertising an operator console to
			// network peers that are outside its trust boundary.
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

func closeMonitor(close monitorCloser) error {
	if close == nil {
		return nil
	}
	if err := close(); err != nil {
		return fmt.Errorf("close worker queue monitor: %w", err)
	}
	return nil
}
