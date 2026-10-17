//go:build integration && chatload

package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/bootstrap/chatadapter"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chathttp "github.com/complexus-tech/projects-api/internal/modules/chat/http"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	ssorepository "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/repository"
	ssoservice "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/service"
	notificationrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notificationservice "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspacesservice "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	securityrepository "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/repository"
	securityservice "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/sse"
	ssehttp "github.com/complexus-tech/projects-api/internal/sse/http"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// This test mounts the production routes, middleware, domain services,
// repositories, session revalidation, outbox dispatcher and Redis/SSE hubs.
// Each API replica owns its own HTTP server, connection pool and hub. They
// deliberately share one process so heap numbers include the load generator.
// No email, push, AI, transcription or remote object-store workers run here.
func TestMessagesConcurrentLoad(t *testing.T) {
	for _, name := range []string{"TEST_DATABASE_URL", "TEST_REDIS_URL"} {
		endpoint, err := url.Parse(os.Getenv(name))
		require.NoError(t, err)
		require.True(t, net.ParseIP(endpoint.Hostname()).IsLoopback(), "%s must be loopback", name)
	}
	users := loadInteger(t, "CHAT_LOAD_USERS", 1000)
	hold := time.Duration(loadInteger(t, "CHAT_LOAD_HOLD_SECONDS", 300)) * time.Second
	ramp := time.Duration(loadInteger(t, "CHAT_LOAD_RAMP_SECONDS", 30)) * time.Second
	pg := testkit.NewPostgres(t)
	redisOptions, err := redis.ParseURL(os.Getenv("TEST_REDIS_URL"))
	require.NoError(t, err)
	redisOptions.PoolSize = 64
	rdb := redis.NewClient(redisOptions)
	t.Cleanup(func() { _ = rdb.Close() })
	require.NoError(t, rdb.Ping(t.Context()).Err())
	// run-chat-load.py starts a private empty Redis process. Refuse shared state:
	// Asynq uses fixed key names and Pub/Sub channels are not database-scoped.
	count, err := rdb.DBSize(t.Context()).Result()
	require.NoError(t, err)
	require.Zero(t, count, "use the disposable runner, never an existing Redis instance")
	errorLog := &loadErrorLog{}
	log := logger.NewWithText(errorLog, slog.LevelError, "chat-load")
	metrics := &loadMetrics{Requests: map[string]*loadSamples{}, Arrivals: map[uuid.UUID]map[uuid.UUID]time.Time{}, Sent: map[uuid.UUID]loadSent{}}
	workspace := uuid.New()
	_, err = pg.Pool.Exec(t.Context(), `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Messages load test',$2)`, workspace, workspace.String())
	require.NoError(t, err)
	const teamCount = 10
	teams, rooms, roots := make([]uuid.UUID, teamCount), make([]uuid.UUID, teamCount), make([]uuid.UUID, teamCount)
	people := make([]loadPerson, users)
	sessions := cache.New(rdb, log)
	for i := range people {
		id := uuid.New()
		token := uuid.NewString() + uuid.NewString()
		_, err = pg.Pool.Exec(t.Context(), `INSERT INTO users(user_id,username,email,full_name) VALUES($1,$2,$3,$4)`, id, id.String(), id.String()+"@example.test", fmt.Sprintf("Load user %d", i))
		require.NoError(t, err)
		_, err = pg.Pool.Exec(t.Context(), `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, workspace, id)
		require.NoError(t, err)
		session, err := auth.NewBrowserSession(id, 1)
		require.NoError(t, err)
		session.ExpiresAt = time.Now().Add(time.Hour)
		require.NoError(t, sessions.Set(t.Context(), cache.AuthSessionCacheKey(token), session, time.Hour))
		people[i] = loadPerson{ID: id, Token: token, Team: i % teamCount}
	}
	seedRepo := chatrepository.New(log, pg.Pool)
	seedService := chatservice.New(seedRepo, chatservice.WithWorkspaceSettings(chatrollout.New(workspacesrepository.New(pg.Pool))))
	for team := range teamCount {
		teams[team] = uuid.New()
		_, err = pg.Pool.Exec(t.Context(), `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,$3,$4,'#555555')`, teams[team], workspace, fmt.Sprintf("Load team %d", team), fmt.Sprintf("L%d", team))
		require.NoError(t, err)
		for i := range people {
			if people[i].Team == team {
				_, err = pg.Pool.Exec(t.Context(), `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, teams[team], people[i].ID)
				require.NoError(t, err)
			}
		}
		require.NoError(t, pg.Pool.QueryRow(t.Context(), `SELECT id FROM chat_conversations WHERE team_id=$1`, teams[team]).Scan(&rooms[team]))
		actor, err := auth.NewHumanActor(people[team].ID).WithWorkspace(workspace)
		require.NoError(t, err)
		for n := range 60 {
			message, err := seedService.CreateMessage(t.Context(), chatdomain.CreateMessageCommand{Scope: chatdomain.ActorScope{WorkspaceID: workspace, Actor: actor}, ConversationID: rooms[team], CreationID: uuid.New(), Text: fmt.Sprintf("History fixture %d", n)})
			require.NoError(t, err)
			roots[team] = message.ID
		}
	}
	// Fixtures predate this run and have no connected recipients. Start metrics
	// from an empty outbox; this database is disposable and owned by testkit.
	_, err = pg.Pool.Exec(t.Context(), `DELETE FROM chat_outbox`)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	taskQueue, err := tasks.New(rdb, log)
	require.NoError(t, err)
	wakeup := chatservice.WithAfterCommit(func(ctx context.Context) {
		wake, stop := context.WithTimeout(ctx, 2*time.Second)
		defer stop()
		if err := taskQueue.EnqueueChatDelivery(wake); err != nil {
			metrics.operationError(fmt.Errorf("enqueue chat delivery: %w", err))
		}
	})
	servers := make([]*httptest.Server, 2)
	origins, err := web.NewOriginPolicy(loadBrowserOrigin)
	require.NoError(t, err)
	var hubs sync.WaitGroup
	for replica := range servers {
		poolConfig, err := pgxpool.ParseConfig(pg.DatabaseURL)
		require.NoError(t, err)
		poolConfig.MaxConns = 24
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		require.NoError(t, err)
		t.Cleanup(pool.Close)
		resolver := mid.NewBrowserSessionResolver(sessions, usersrepository.New(pool))
		workspaceAccess := workspaceResolver{service: workspacesservice.New(log, workspacesrepository.New(pool), nil, workspacesservice.Dependencies{}), security: securityservice.New(securityrepository.New(pool)), sso: ssoservice.New(ssorepository.New(pool), nil, nil, "")}
		hub := sse.NewHub(log, rdb)
		hubs.Add(1)
		go func() { defer hubs.Done(); _ = hub.Run(ctx) }()
		app := web.New(nil, noop.NewTracerProvider().Tracer("chat-load"), mid.Logger(log), mid.RequireTrustedBrowserOrigin(origins))
		app.SetOriginPolicy(origins)
		app.SetLogger(log)
		app.StrictSlash(false)
		chathttp.Routes(chathttp.Config{Log: log, BrowserSessions: resolver, WorkspaceResolver: workspaceAccess, Service: chatservice.New(chatrepository.New(log, pool), chatservice.WithWorkspaceSettings(chatrollout.New(workspacesrepository.New(pool))), wakeup)}, app)
		ssehttp.Routes(ssehttp.Config{Log: log, BrowserSessions: resolver, WorkspaceResolver: workspaceAccess, SSEHub: hub, Origins: origins}, app)
		servers[replica] = httptest.NewServer(app)
	}
	defer func() {
		cancel()
		for _, server := range servers {
			server.CloseClientConnections()
			server.Close()
		}
		hubs.Wait()
	}()
	notificationService := notificationservice.New(log, notificationrepository.New(pg.Pool), rdb, loadNotificationQueue{Service: taskQueue})
	faults := &loadPublisher{publisher: publisher.New(rdb, log)}
	dispatcher := chatadapter.Dispatcher{Store: seedRepo, Updates: faults, Notifications: notificationService}
	var paused atomic.Bool
	var dispatchGate sync.RWMutex
	var dispatching atomic.Int64
	taskMux := asynq.NewServeMux()
	taskMux.HandleFunc(tasks.TypeChatDelivery, func(ctx context.Context, _ *asynq.Task) error {
		dispatchGate.RLock()
		defer dispatchGate.RUnlock()
		if paused.Load() {
			return nil
		}
		dispatching.Add(1)
		defer dispatching.Add(-1)
		_, err := dispatcher.Dispatch(ctx, 50)
		if err != nil && ctx.Err() == nil && !errors.Is(err, errLoadPublishFailure) {
			metrics.workerError(err)
		}
		return err
	})
	for range 2 {
		worker := asynq.NewServer(asynq.RedisClientOpt{Addr: redisOptions.Addr, DB: redisOptions.DB}, asynq.Config{Concurrency: 2, Queues: map[string]int{"default": 1}, ShutdownTimeout: 5 * time.Second, LogLevel: asynq.ErrorLevel})
		require.NoError(t, worker.Start(taskMux))
		defer worker.Shutdown()
	}
	scheduler := asynq.NewScheduler(asynq.RedisClientOpt{Addr: redisOptions.Addr, DB: redisOptions.DB}, &asynq.SchedulerOpts{LogLevel: asynq.ErrorLevel})
	_, err = scheduler.Register("@every 5s", asynq.NewTask(tasks.TypeChatDelivery, nil), asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(30*time.Second), asynq.Unique(time.Second))
	require.NoError(t, err)
	require.NoError(t, scheduler.Start())
	defer scheduler.Shutdown()
	transport := &http.Transport{MaxIdleConns: 200, MaxIdleConnsPerHost: 100, MaxConnsPerHost: 0, ResponseHeaderTimeout: 15 * time.Second, DisableCompression: true}
	client := &http.Client{Transport: transport}
	defer transport.CloseIdleConnections()
	runner := &loadRunner{t: t, ctx: ctx, client: client, servers: servers, workspace: workspace, metrics: metrics, people: people, rooms: rooms, roots: roots}
	runner.reconcile = newLoadReconciliation(t, runner, users)
	defer runner.reconcile.stop()
	streams := make([]*loadStream, users)
	start := time.Now()
	reportWritten := false
	var connect sync.WaitGroup
	for i := range people {
		connect.Add(1)
		go func(index int) {
			defer connect.Done()
			// Split every team across both replicas, rather than correlating a
			// user's team (index%10) with a fixed server parity.
			streams[index] = runner.connect(index, (index/teamCount)%2)
		}(i)
		time.Sleep(ramp / time.Duration(users))
	}
	connect.Wait()
	defer func() {
		for _, stream := range streams {
			if stream != nil {
				stream.close()
			}
		}
	}()
	defer func() {
		if reportWritten {
			return
		}
		partial := metrics.report(users, hold, time.Since(start), 0, 0)
		partial["passed"] = false
		partial["application_errors"] = errorLog.snapshot()
		partial["all_user_probe_deliveries"] = 0
		partial["incomplete"] = "Test stopped before final probes/budgets; zero peaks mean unavailable, not zero memory."
		writeLoadReport(t, partial)
	}()
	require.Equal(t, int64(users), metrics.Active.Load(), "every distinct user must have a live authenticated SSE stream")
	// Independently count the real subscriptions that registration acknowledged
	// before sending; no synthetic delivery bypass.
	require.Eventually(t, func() bool {
		counts, err := rdb.PubSubNumSub(ctx, "workspace-updates:"+workspace.String()).Result()
		return err == nil && counts["workspace-updates:"+workspace.String()] == int64(users)
	}, 30*time.Second, 100*time.Millisecond)
	t.Logf("Connected %d distinct authenticated users on two replicas; holding mixed traffic for %s", users, hold)
	require.NoError(t, runner.validateCursor(0))
	var peakHeap, peakGoroutines uint64
	holdStart := time.Now()
	runner.steady.Store(true)
	runner.reconcile.start()
	monitorCtx, stopMonitor := context.WithCancel(ctx)
	monitorDone := make(chan struct{})
	go func() { defer close(monitorDone); metrics.monitor(monitorCtx, pg.Pool, rdb) }()
	defer func() { stopMonitor(); <-monitorDone }()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var requests sync.WaitGroup
	limit := make(chan struct{}, 100)
	tick := 0
	reconnected := false
	for time.Since(holdStart) < hold {
		<-ticker.C
		tick++
		if tick%600 == 0 {
			t.Logf("Held %.0fs: %d SSE clients active, %d messages created", time.Since(holdStart).Seconds(), metrics.Active.Load(), metrics.createdCount())
		}
		var memory runtime.MemStats
		if tick%20 == 0 {
			runtime.ReadMemStats(&memory)
		}
		if memory.HeapAlloc > peakHeap {
			peakHeap = memory.HeapAlloc
		}
		if uint64(runtime.NumGoroutine()) > peakGoroutines {
			peakGoroutines = uint64(runtime.NumGoroutine())
		}
		// 20 requests/second read history. 5 messages/second and one reaction/
		// second are additional independent operations, distributed over users.
		index := (tick - 1) % users
		jobs := []func() error{func() error { return runner.read(index) }}
		if tick%4 == 0 {
			author := (tick/4 - 1) % users
			jobs = append(jobs, func() error { return runner.send(author, nil) })
		}
		if tick%20 == 0 {
			author := (tick/20 - 1) % users
			jobs = append(jobs, func() error { return runner.react(author) })
		}
		if tick%100 == 0 {
			author := (tick/100 - 1) % users
			root := roots[people[author].Team]
			jobs = append(jobs, func() error { return runner.send(author, &root) })
		}
		for _, job := range jobs {
			select {
			case limit <- struct{}{}:
				requests.Add(1)
				go func(operation func() error) {
					defer requests.Done()
					defer func() { <-limit }()
					if err := operation(); err != nil {
						metrics.operationError(err)
					}
				}(job)
			default:
				metrics.operationError(fmt.Errorf("load generator concurrency limit reached"))
			}
		}
		if !reconnected && time.Since(holdStart) > hold/2 {
			reconnected = true
			// Disconnect ten percent, then reconnect to the other API instance. The
			// others stay connected; the endpoint returns authoritative history.
			for i := 0; i < users; i += 10 {
				streams[i].close()
			}
			for i := 0; i < users; i += 10 {
				old := i
				nextReplica := 1 - streams[i].replica
				requests.Add(1)
				go func() {
					defer requests.Done()
					streams[old] = runner.connect(old, nextReplica)
					if err := runner.read(old); err != nil {
						metrics.operationError(err)
					}
				}()
			}
		}
	}
	requests.Wait()
	runner.reconcile.stop()
	runner.steady.Store(false)
	stopMonitor()
	<-monitorDone
	require.Equal(t, int64(users), metrics.Active.Load(), "all streams remain connected after held traffic/reconnect")
	// Drain normal traffic before the intentional outage so fault windows do
	// not contaminate held-traffic delivery percentiles.
	require.Eventually(t, func() bool {
		var pending int
		err := pg.Pool.QueryRow(ctx, `SELECT count(*) FROM chat_outbox WHERE delivered_at IS NULL`).Scan(&pending)
		return err == nil && pending == 0
	}, 30*time.Second, 100*time.Millisecond)
	// Verify outbox retry across a real temporary worker outage. The API commits
	// while delivery is stopped; after resume the original event is delivered.
	dispatchGate.Lock()
	paused.Store(true)
	dispatchGate.Unlock()
	require.Eventually(t, func() bool { return dispatching.Load() == 0 }, 20*time.Second, 20*time.Millisecond)
	before := metrics.createdCount()
	require.NoError(t, runner.send(0, nil))
	after := metrics.createdCount()
	require.Equal(t, before+1, after)
	var pendingDuringPause int
	require.NoError(t, pg.Pool.QueryRow(ctx, `SELECT count(*) FROM chat_outbox WHERE delivered_at IS NULL`).Scan(&pendingDuringPause))
	require.Positive(t, pendingDuringPause, "API commit remains durable while workers are stopped")
	faults.fail.Store(true)
	time.Sleep(time.Second)
	recoveryStarted := time.Now()
	paused.Store(false)
	require.Eventually(t, func() bool {
		var pending int
		err := pg.Pool.QueryRow(ctx, `SELECT count(*) FROM chat_outbox WHERE delivered_at IS NULL`).Scan(&pending)
		return err == nil && pending == 0
	}, 30*time.Second, 100*time.Millisecond)
	recoveryDuration := time.Since(recoveryStarted)
	require.Equal(t, int64(1), faults.failures.Load(), "one real dispatcher publish failed and was retried")
	var retried int
	require.NoError(t, pg.Pool.QueryRow(ctx, `SELECT count(*) FROM chat_outbox WHERE attempts > 1 AND delivered_at IS NOT NULL`).Scan(&retried))
	require.Positive(t, retried, "failed event was durably retried")
	time.Sleep(time.Second)
	// Messages created during intentional reconnect are recovered by history;
	// live fanout completeness is checked after a settled reconnect with probes.
	probeIDs := make([]uuid.UUID, 0, teamCount)
	for team := range teamCount {
		id, err := runner.create(team, nil)
		require.NoError(t, err)
		probeIDs = append(probeIDs, id)
	}
	require.Eventually(t, func() bool { return metrics.probesComplete(probeIDs, people) }, 30*time.Second, 100*time.Millisecond, "cross-replica delivery to every authorized team member")
	require.NoError(t, runner.validateIdempotency(0))
	require.NoError(t, runner.validateMutations(0))
	require.NoError(t, runner.validateCatchup(0, streams))
	require.Eventually(t, func() bool {
		var pending int
		err := pg.Pool.QueryRow(ctx, `SELECT count(*) FROM chat_outbox WHERE delivered_at IS NULL`).Scan(&pending)
		return err == nil && pending == 0
	}, 30*time.Second, 100*time.Millisecond, "all post-load mutation events drain")
	require.NoError(t, runner.validateWorkspaceFeature(pg.Pool), "both API replicas observe live disable/re-enable without restart")
	require.Equal(t, int64(users), metrics.Active.Load(), "workspace feature toggling preserves the shared platform SSE streams")
	report := metrics.report(users, hold, time.Since(start), peakHeap, peakGoroutines)
	report["live_workspace_feature_toggle"] = "both replicas: disabled capabilities/read/create, then re-enabled capabilities/read; no restart"
	report["application_errors"] = errorLog.snapshot()
	clientsInfo, err := rdb.Info(ctx, "clients").Result()
	require.NoError(t, err)
	report["redis_clients"] = loadInfoInteger(clientsInfo, "connected_clients")
	memoryInfo, err := rdb.Info(ctx, "memory").Result()
	require.NoError(t, err)
	report["redis_memory_bytes"] = loadInfoInteger(memoryInfo, "used_memory")
	report["target_history_requests_per_second"] = 20
	report["held_history_requests"] = tick
	report["held_top_level_messages"] = tick / 4
	report["held_thread_replies"] = tick / 100
	report["held_reactions"] = tick / 20
	report["active_chat_users"] = len(runner.reconcile.hints)
	report["reactive_sse_hints"] = runner.reconcile.received.Load()
	report["reactive_refresh_cycles"] = runner.reconcile.cycles.Load()
	report["reactive_http_reads"] = runner.reconcile.cycles.Load() * 4
	report["held_reactive_http_reads"] = runner.reconcile.heldCycles.Load() * 4
	report["reactive_http_reads_per_second"] = float64(runner.reconcile.heldCycles.Load()*4) / hold.Seconds()
	report["synchronized_chat_start"] = runner.reconcile.synchronized
	report["achieved_history_requests_per_second"] = float64(tick) / hold.Seconds()
	report["outbox_recovery_ms"] = float64(recoveryDuration.Milliseconds())
	report["retried_outbox_events"] = retried
	report["pending_during_worker_pause"] = pendingDuringPause
	report["scenario"] = "10 teams, 2 real HTTP API/hub replicas in one process, 20 history reads/s + 5 message creates/s + 1 reaction/s + 0.2 replies/s; 10% cross-replica reconnect; real Asynq wakeups and 5s sweep, durable outbox pause/retry/recovery"
	report["limits"] = "Loopback laptop; no TLS/proxy, browser rendering, object storage, AI, voice, email/push delivery or network partition. API replicas share process/Redis client; process memory includes generator. Not a production capacity guarantee."
	defer func() {
		report["passed"] = !t.Failed()
		writeLoadReport(t, report)
		reportWritten = true
	}()
	require.Empty(t, metrics.errorSnapshot(), "no unexpected HTTP, stream, worker or load-generator errors")
	require.GreaterOrEqual(t, float64(tick)/hold.Seconds(), 19.0, "achieve at least 95% of the offered 20 history requests/second")
	delivery := report["steady_delivery_lag_ms"].(map[string]any)
	require.Less(t, delivery["p95"].(float64), 2000.0, "steady SSE delivery p95 budget (ms)")
	require.Less(t, delivery["p99"].(float64), 5000.0, "steady SSE delivery p99 budget (ms)")
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	for operation, samples := range metrics.Requests {
		require.Less(t, loadPercentile(samples.Values, .95), 1000.0, "%s p95 budget (ms)", operation)
		require.Less(t, loadPercentile(samples.Values, .99), 2000.0, "%s p99 budget (ms)", operation)
	}
}
