//go:build integration && chatload

package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

const loadBrowserOrigin = "http://127.0.0.1:4177"

// Preserve actionable server errors on failure without writing every request
// log or growing the report without bound. The writer is shared by replicas.
type loadErrorLog struct {
	mu   sync.Mutex
	data []byte
}

func (l *loadErrorLog) Write(value []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	remaining := 128*1024 - len(l.data)
	if remaining > 0 {
		l.data = append(l.data, value[:min(len(value), remaining)]...)
	}
	return len(value), nil
}

func (l *loadErrorLog) snapshot() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return string(l.data)
}

type loadPerson struct {
	ID    uuid.UUID
	Token string
	Team  int
}
type loadSent struct {
	At     time.Time
	Room   uuid.UUID
	Steady bool
}
type loadSamples struct {
	Values []float64
	Errors int
}
type loadMetrics struct {
	mu               sync.Mutex
	Requests         map[string]*loadSamples
	Arrivals         map[uuid.UUID]map[uuid.UUID]time.Time
	Sent             map[uuid.UUID]loadSent
	Errors           []string
	Active           atomic.Int64
	Peak             atomic.Int64
	PeakRedisClients atomic.Int64
	PeakRedisMemory  atomic.Int64
	PeakQueueDepth   atomic.Int64
	PeakQueueAgeMS   atomic.Int64
}

func (m *loadMetrics) record(kind string, elapsed time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Requests[kind] == nil {
		m.Requests[kind] = &loadSamples{}
	}
	m.Requests[kind].Values = append(m.Requests[kind].Values, float64(elapsed.Microseconds())/1000)
	if err != nil {
		m.Requests[kind].Errors++
	}
}
func (m *loadMetrics) operationError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Errors) < 100 {
		m.Errors = append(m.Errors, err.Error())
	}
}
func (m *loadMetrics) errorSnapshot() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.Errors...)
}
func (m *loadMetrics) workerError(err error) { m.operationError(fmt.Errorf("outbox: %w", err)) }
func (m *loadMetrics) createdCount() int     { m.mu.Lock(); defer m.mu.Unlock(); return len(m.Sent) }
func (m *loadMetrics) probesComplete(ids []uuid.UUID, people []loadPerson) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for team, id := range ids {
		for _, person := range people {
			if person.Team == team && m.Arrivals[id][person.ID].IsZero() {
				return false
			}
		}
	}
	return true
}
func (m *loadMetrics) report(users int, hold, elapsed time.Duration, heap, goroutines uint64) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	requests := map[string]any{}
	for operation, samples := range m.Requests {
		requests[operation] = map[string]any{
			"count":  len(samples.Values),
			"errors": samples.Errors,
			"p50_ms": loadPercentile(samples.Values, .5),
			"p95_ms": loadPercentile(samples.Values, .95),
			"p99_ms": loadPercentile(samples.Values, .99),
			"max_ms": loadPercentile(samples.Values, 1),
		}
	}
	lags := []float64{}
	steadyLags := []float64{}
	deliveries := 0
	for id, sent := range m.Sent {
		for _, at := range m.Arrivals[id] {
			deliveries++
			lag := float64(at.Sub(sent.At).Microseconds()) / 1000
			lags = append(lags, lag)
			if sent.Steady {
				steadyLags = append(steadyLags, lag)
			}
		}
	}
	return map[string]any{
		"run_at":                             time.Now().UTC().Format(time.RFC3339),
		"go_version":                         runtime.Version(),
		"cpus":                               runtime.NumCPU(),
		"distinct_users":                     users,
		"active_at_end":                      m.Active.Load(),
		"peak_sse_connections":               m.Peak.Load(),
		"hold_seconds":                       hold.Seconds(),
		"total_seconds":                      elapsed.Seconds(),
		"requests":                           requests,
		"message_creates":                    len(m.Sent),
		"observed_created_deliveries":        deliveries,
		"steady_observed_created_deliveries": len(steadyLags),
		"steady_delivery_lag_ms": map[string]any{
			"p50": loadPercentile(steadyLags, .5),
			"p95": loadPercentile(steadyLags, .95),
			"p99": loadPercentile(steadyLags, .99),
			"max": loadPercentile(steadyLags, 1),
		},
		"delivery_lag_ms": map[string]any{
			"p50": loadPercentile(lags, .5),
			"p95": loadPercentile(lags, .95),
			"p99": loadPercentile(lags, .99),
			"max": loadPercentile(lags, 1),
		},
		"all_user_probe_deliveries":  users,
		"peak_redis_clients":         m.PeakRedisClients.Load(),
		"peak_redis_memory_bytes":    m.PeakRedisMemory.Load(),
		"peak_pending_outbox_events": m.PeakQueueDepth.Load(),
		"peak_oldest_outbox_age_ms":  m.PeakQueueAgeMS.Load(),
		"peak_go_heap_bytes":         heap,
		"peak_goroutines":            goroutines,
		"errors":                     append([]string(nil), m.Errors...),
	}
}
func loadPercentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copyValues := append([]float64(nil), values...)
	sort.Float64s(copyValues)
	index := int(float64(len(copyValues)-1) * p)
	return copyValues[index]
}
func loadInteger(t *testing.T, key string, fallback int) int {
	t.Helper()
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	require.NoError(t, err)
	require.Positive(t, value)
	return value
}
func loadInfoInteger(info, key string) int64 {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, key+":") {
			value, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, key+":")), 10, 64)
			return value
		}
	}
	return 0
}

type loadStream struct {
	cancel   context.CancelFunc
	body     io.ReadCloser
	done     chan struct{}
	expected atomic.Bool
	replica  int
}

func (s *loadStream) close() { s.expected.Store(true); s.cancel(); _ = s.body.Close(); <-s.done }

type loadRunner struct {
	t            *testing.T
	ctx          context.Context
	client       *http.Client
	servers      []*httptest.Server
	workspace    uuid.UUID
	metrics      *loadMetrics
	people       []loadPerson
	rooms, roots []uuid.UUID
	steady       atomic.Bool
	reconcile    *loadReconciliation
}

func (r *loadRunner) path(suffix string) string {
	return "/workspaces/" + r.workspace.String() + "/chat/" + suffix
}
func (r *loadRunner) request(index int, method, path, kind string, body any, result any) error {
	return r.requestStatus(index, method, path, kind, body, result, 0)
}

// A nonzero expectedStatus is used only for deliberate feature-denial probes.
// Ordinary load calls still fail on every HTTP error and are never retried.
func (r *loadRunner) requestStatus(index int, method, path, kind string, body any, result any, expectedStatus int) (err error) {
	start := time.Now()
	defer func() { r.metrics.record(kind, time.Since(start), err) }()
	var reader io.Reader
	if body != nil {
		raw, e := json.Marshal(body)
		if e != nil {
			return e
		}
		reader = bytes.NewReader(raw)
	}
	ctx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
	defer cancel()
	request, e := http.NewRequestWithContext(ctx, method, r.servers[index%2].URL+path, reader)
	if e != nil {
		return e
	}
	request.AddCookie(&http.Cookie{Name: "fortyone_session", Value: r.people[index].Token})
	request.Header.Set("Origin", loadBrowserOrigin)
	request.Header.Set("Content-Type", "application/json")
	response, e := r.client.Do(request)
	if e != nil {
		return e
	}
	defer response.Body.Close()
	raw, e := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if e != nil {
		return e
	}
	if (expectedStatus == 0 && response.StatusCode >= 400) || (expectedStatus != 0 && response.StatusCode != expectedStatus) {
		return fmt.Errorf("%s %s returned %d: %s", method, kind, response.StatusCode, raw)
	}
	if result == nil {
		return nil
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if e = json.Unmarshal(raw, &envelope); e != nil {
		return e
	}
	return json.Unmarshal(envelope.Data, result)
}
func (r *loadRunner) connect(index, replica int) *loadStream {
	start := time.Now()
	ctx, cancel := context.WithCancel(r.ctx)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, r.servers[replica].URL+"/workspaces/"+r.workspace.String()+"/notifications/subscribe", nil)
	if err != nil {
		cancel()
		r.metrics.operationError(err)
		return nil
	}
	request.AddCookie(&http.Cookie{Name: "fortyone_session", Value: r.people[index].Token})
	request.Header.Set("Origin", loadBrowserOrigin)
	response, err := r.client.Do(request)
	if err != nil {
		cancel()
		r.metrics.operationError(err)
		return nil
	}
	if response.StatusCode != 200 {
		cancel()
		response.Body.Close()
		r.metrics.operationError(fmt.Errorf("SSE connect status %d", response.StatusCode))
		return nil
	}
	r.metrics.record("sse_connect", time.Since(start), nil)
	stream := &loadStream{cancel: cancel, body: response.Body, done: make(chan struct{}), replica: replica}
	active := r.metrics.Active.Add(1)
	for current := r.metrics.Peak.Load(); active > current && !r.metrics.Peak.CompareAndSwap(current, active); current = r.metrics.Peak.Load() {
	}
	go func() {
		defer close(stream.done)
		defer r.metrics.Active.Add(-1)
		scanner := bufio.NewScanner(response.Body)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var event publisher.ChatUpdate
			if json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event) != nil || event.Type != publisher.ChatUpdatedType {
				continue
			}
			if event.UserID != r.people[index].ID || event.WorkspaceID != r.workspace {
				r.metrics.operationError(fmt.Errorf("SSE event escaped recipient scope for user %d", index))
				continue
			}
			if r.reconcile != nil && event.ConversationID == r.rooms[r.people[index].Team] {
				r.reconcile.hint(index)
			}
			if event.Change != "message.created" || event.MessageID == nil {
				continue
			}
			r.metrics.mu.Lock()
			if r.metrics.Arrivals[*event.MessageID] == nil {
				r.metrics.Arrivals[*event.MessageID] = map[uuid.UUID]time.Time{}
			}
			if r.metrics.Arrivals[*event.MessageID][r.people[index].ID].IsZero() {
				r.metrics.Arrivals[*event.MessageID][r.people[index].ID] = time.Now()
			}
			r.metrics.mu.Unlock()
		}
		if !stream.expected.Load() && r.ctx.Err() == nil {
			r.metrics.operationError(fmt.Errorf("unexpected SSE close for user %d: %v", index, scanner.Err()))
		}
	}()
	if r.reconcile != nil {
		r.reconcile.connected(index)
	}
	return stream
}
func (r *loadRunner) read(index int) error {
	var page chatdomain.MessagePage
	return r.request(index, http.MethodGet, r.path("conversations/"+r.rooms[r.people[index].Team].String()+"/messages?limit=30"), "history", nil, &page)
}
func (r *loadRunner) create(index int, root *uuid.UUID) (uuid.UUID, error) {
	var message chatdomain.Message
	err := r.request(index, http.MethodPost, r.path("conversations/"+r.rooms[r.people[index].Team].String()+"/messages"), "create", map[string]any{"creationId": uuid.New(), "text": "Concurrent load message " + uuid.NewString(), "threadRootId": root}, &message)
	if err != nil {
		return uuid.Nil, err
	}
	r.metrics.mu.Lock()
	r.metrics.Sent[message.ID] = loadSent{At: message.CreatedAt, Room: message.ConversationID, Steady: r.steady.Load()}
	r.metrics.mu.Unlock()
	return message.ID, nil
}
func (r *loadRunner) send(index int, root *uuid.UUID) error {
	_, err := r.create(index, root)
	return err
}
func (r *loadRunner) react(index int) error {
	return r.request(index, http.MethodPut, r.path("messages/"+r.roots[r.people[index].Team].String()+"/reactions/"+url.PathEscape("✅")), "reaction", nil, nil)
}
func (r *loadRunner) validateCursor(index int) error {
	var first, second chatdomain.MessagePage
	path := r.path("conversations/" + r.rooms[r.people[index].Team].String() + "/messages?limit=20")
	if err := r.request(index, http.MethodGet, path, "cursor", nil, &first); err != nil {
		return err
	}
	if first.OlderCursor == nil || !first.HasOlder {
		return fmt.Errorf("history did not return an older cursor")
	}
	if err := r.request(index, http.MethodGet, path+"&before="+url.QueryEscape(*first.OlderCursor), "cursor", nil, &second); err != nil {
		return err
	}
	seen := map[uuid.UUID]bool{}
	for _, message := range first.Items {
		seen[message.ID] = true
	}
	for _, message := range second.Items {
		if seen[message.ID] {
			return fmt.Errorf("cursor pages overlapped")
		}
	}
	return nil
}
func (r *loadRunner) validateIdempotency(index int) error {
	input := map[string]any{"creationId": uuid.New(), "text": "Idempotency probe"}
	var first, second chatdomain.Message
	path := r.path("conversations/" + r.rooms[r.people[index].Team].String() + "/messages")
	if err := r.request(index, http.MethodPost, path, "idempotent_create", input, &first); err != nil {
		return err
	}
	if err := r.request(index, http.MethodPost, path, "idempotent_replay", input, &second); err != nil {
		return err
	}
	if first.ID != second.ID {
		return fmt.Errorf("idempotency replay created another message")
	}
	return nil
}

var errLoadPublishFailure = errors.New("injected one-shot Redis publish failure")

type loadPublisher struct {
	publisher *publisher.Publisher
	fail      atomic.Bool
	failures  atomic.Int64
}

func (p *loadPublisher) PublishChatUpdated(ctx context.Context, update publisher.ChatUpdate) error {
	if p.fail.CompareAndSwap(true, false) {
		p.failures.Add(1)
		return errLoadPublishFailure
	}
	return p.publisher.PublishChatUpdated(ctx, update)
}
func (r *loadRunner) validateMutations(index int) error {
	id, err := r.create(index, nil)
	if err != nil {
		return err
	}
	var edited chatdomain.Message
	path := r.path("messages/" + id.String())
	if err = r.request(index, http.MethodPatch, path, "edit", map[string]any{"version": 1, "text": "Edited during load verification"}, &edited); err != nil {
		return err
	}
	if edited.Version != 2 || edited.Text != "Edited during load verification" {
		return fmt.Errorf("edit did not produce expected version")
	}
	var loaded chatdomain.Message
	if err = r.request(index, http.MethodGet, path, "read_message", nil, &loaded); err != nil {
		return err
	}
	if loaded.Version != edited.Version {
		return fmt.Errorf("read did not see edited version")
	}
	if err = r.request(index, http.MethodPost, r.path("conversations/"+r.rooms[r.people[index].Team].String()+"/read"), "mark_read", map[string]any{"throughMessageId": id}, nil); err != nil {
		return err
	}
	if err = r.request(index, http.MethodGet, r.path("messages/"+r.roots[r.people[index].Team].String()+"/replies?limit=30"), "thread", nil, &chatdomain.MessagePage{}); err != nil {
		return err
	}
	return r.request(index, http.MethodDelete, path+"?version=2", "delete", nil, nil)
}
func (r *loadRunner) validateCatchup(index int, streams []*loadStream) error {
	var before, after chatdomain.MessagePage
	path := r.path("conversations/" + r.rooms[r.people[index].Team].String() + "/messages?limit=30")
	if err := r.request(index, http.MethodGet, path, "catchup_cursor", nil, &before); err != nil {
		return err
	}
	if before.NewerCursor == nil {
		return fmt.Errorf("no catchup cursor")
	}
	nextReplica := 1 - streams[index].replica
	streams[index].close()
	missed, err := r.create(index, nil)
	if err != nil {
		return err
	}
	// Let the real outbox publish while this browser is disconnected. SSE is a
	// hint, so recovery must work from authoritative cursor history, not replay.
	time.Sleep(2 * time.Second)
	streams[index] = r.connect(index, nextReplica)
	if err = r.request(index, http.MethodGet, path+"&after="+url.QueryEscape(*before.NewerCursor), "catchup_history", nil, &after); err != nil {
		return err
	}
	for _, message := range after.Items {
		if message.ID == missed {
			return nil
		}
	}
	return fmt.Errorf("reconnected cursor history missed committed message")
}

// Notification persistence and enqueueing remain real. Outbound jobs are
// routed to an unconsumed queue: this load test must not send email or push.
type loadNotificationQueue struct{ *tasks.Service }

func (q loadNotificationQueue) EnqueueNotificationEmailDigest(payload tasks.NotificationEmailDigestPayload, options ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.Service.EnqueueNotificationEmailDigest(payload, append(options, asynq.Queue("load-external-disabled"))...)
}
func (q loadNotificationQueue) EnqueueNotificationPush(payload tasks.NotificationPushPayload, options ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.Service.EnqueueNotificationPush(payload, append(options, asynq.Queue("load-external-disabled"))...)
}

func writeLoadReport(t *testing.T, report map[string]any) {
	t.Helper()
	encoded, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)
	path := os.Getenv("CHAT_LOAD_REPORT")
	require.NotEmpty(t, path)
	require.NoError(t, os.WriteFile(path, append(encoded, '\n'), 0600))
	t.Logf("CHAT_LOAD_REPORT %s\n%s", path, encoded)
}

func (m *loadMetrics) monitor(ctx context.Context, pool *pgxpool.Pool, client *redis.Client) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sample, cancel := context.WithTimeout(ctx, time.Second)
			var depth int64
			var age float64
			err := pool.QueryRow(sample, `SELECT count(*), COALESCE(CAST(EXTRACT(EPOCH FROM now()-min(created_at)) AS double precision),0) FROM chat_outbox WHERE delivered_at IS NULL`).Scan(&depth, &age)
			if err == nil {
				loadAtomicMax(&m.PeakQueueDepth, depth)
				loadAtomicMax(&m.PeakQueueAgeMS, int64(age*1000))
			} else if ctx.Err() == nil {
				m.operationError(fmt.Errorf("sample outbox: %w", err))
			}
			info, err := client.Info(sample, "clients", "memory").Result()
			if err == nil {
				loadAtomicMax(&m.PeakRedisClients, loadInfoInteger(info, "connected_clients"))
				loadAtomicMax(&m.PeakRedisMemory, loadInfoInteger(info, "used_memory"))
			} else if ctx.Err() == nil {
				m.operationError(fmt.Errorf("sample Redis: %w", err))
			}
			cancel()
		}
	}
}
func loadAtomicMax(target *atomic.Int64, value int64) {
	for current := target.Load(); value > current && !target.CompareAndSwap(current, value); current = target.Load() {
	}
}
