//go:build integration

package idempotency

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

type integrationClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *integrationClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *integrationClock) Advance(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(duration)
}

type integrationFixture struct {
	database *testkit.Postgres
	service  *Service
	clock    *integrationClock
	config   Config
	scope    Scope
	key      Key
	rawKey   string
}

func newIntegrationFixture(t *testing.T, operationName string) integrationFixture {
	t.Helper()

	database := testkit.NewPostgres(t)
	var serverVersion int
	if err := database.Pool.QueryRow(t.Context(), "SELECT CAST(current_setting('server_version_num') AS integer)").Scan(&serverVersion); err != nil {
		t.Fatalf("read PostgreSQL server version: %v", err)
	}
	if serverVersion < 180000 {
		t.Fatalf("idempotency integration contract requires PostgreSQL 18 or later, got server_version_num %d", serverVersion)
	}

	config := Config{
		LeaseDuration:     MinLeaseDuration,
		RetentionDuration: MinRetentionDuration,
	}
	service, err := New(database.Pool, config)
	if err != nil {
		t.Fatalf("create idempotency service: %v", err)
	}
	clock := &integrationClock{now: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)}
	service.now = clock.Now
	service.newReceiptID = uuid.New

	operation, err := ParseOperation(operationName)
	if err != nil {
		t.Fatalf("parse operation: %v", err)
	}
	scope, err := NewScope(auth.NewHumanActor(uuid.New()), MethodPost, operation)
	if err != nil {
		t.Fatalf("create receipt scope: %v", err)
	}
	rawKey := "integration-receipt-" + uuid.NewString()
	key, err := ParseKey(rawKey)
	if err != nil {
		t.Fatalf("parse receipt key: %v", err)
	}

	return integrationFixture{
		database: database,
		service:  service,
		clock:    clock,
		config:   config,
		scope:    scope,
		key:      key,
		rawKey:   rawKey,
	}
}

func TestReceiptSameBodyCompletesAndReplays(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.same_body")
	requestBody := []byte(`{"title":"same request"}`)
	first, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, requestBody)
	if err != nil {
		t.Fatalf("begin first receipt: %v", err)
	}
	if first.State != BeginStateNew || first.Lease.Generation != 1 || first.Reclaimed {
		t.Fatalf("first begin = %#v, want a new generation-one lease", first)
	}

	response, err := NewResponse(201, []byte(`{"id":"story-1"}`), "application/json")
	if err != nil {
		t.Fatalf("create replay response: %v", err)
	}
	if err := fixture.service.Complete(t.Context(), first.Lease, response); err != nil {
		t.Fatalf("complete receipt: %v", err)
	}

	replay, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, requestBody)
	if err != nil {
		t.Fatalf("begin completed receipt: %v", err)
	}
	if replay.State != BeginStateCompleted || replay.Replay.StatusCode() != 201 ||
		replay.Replay.ContentType() != "application/json" ||
		!bytes.Equal(replay.Replay.Body(), response.Body()) {
		t.Fatalf("completed replay = %#v, want exact safe response", replay)
	}

	var keyDigest, requestHash []byte
	if err := fixture.database.Pool.QueryRow(t.Context(), `
		SELECT key_digest, request_hash
		FROM public.api_idempotency_receipts
		WHERE receipt_id = $1
	`, first.Lease.ReceiptID).Scan(&keyDigest, &requestHash); err != nil {
		t.Fatalf("read receipt digests: %v", err)
	}
	if len(keyDigest) != 32 || len(requestHash) != 32 ||
		bytes.Equal(keyDigest, []byte(fixture.rawKey)) || bytes.Equal(requestHash, requestBody) {
		t.Fatalf("receipt persisted unexpected raw material: key digest=%d request hash=%d", len(keyDigest), len(requestHash))
	}
}

func TestReceiptDifferentBodyConflicts(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.conflict")
	if result, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("first")); err != nil || result.State != BeginStateNew {
		t.Fatalf("first begin = (%#v, %v), want new", result, err)
	}

	result, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("different"))
	if err != nil {
		t.Fatalf("conflicting begin error = %v", err)
	}
	if result.State != BeginStateConflict {
		t.Fatalf("conflicting begin state = %q, want %q", result.State, BeginStateConflict)
	}
}

func TestConcurrentReceiptBeginHasOneWinner(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.concurrent")
	start := make(chan struct{})
	results := make(chan BeginResult, 2)
	errorsChannel := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			result, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("same"))
			results <- result
			errorsChannel <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent begin error = %v", err)
		}
	}
	counts := map[BeginState]int{}
	for result := range results {
		counts[result.State]++
	}
	if counts[BeginStateNew] != 1 || counts[BeginStateInProgress] != 1 {
		t.Fatalf("concurrent begin counts = %#v, want one new and one in-progress", counts)
	}
}

func TestStaleReceiptTakeoverFencesPreviousCompletion(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.takeover")
	first, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("same"))
	if err != nil {
		t.Fatalf("begin first receipt: %v", err)
	}
	fixture.clock.Advance(fixture.config.LeaseDuration)

	takenOver, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("same"))
	if err != nil {
		t.Fatalf("take over stale receipt: %v", err)
	}
	if takenOver.State != BeginStateNew || !takenOver.Reclaimed ||
		takenOver.Lease.ReceiptID != first.Lease.ReceiptID ||
		takenOver.Lease.Generation != first.Lease.Generation+1 {
		t.Fatalf("takeover = %#v, want reclaimed next generation", takenOver)
	}

	response, err := NewResponse(200, []byte("done"), "text/plain; charset=utf-8")
	if err != nil {
		t.Fatalf("create response: %v", err)
	}
	if err := fixture.service.Complete(t.Context(), first.Lease, response); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale completion error = %v, want ErrLeaseLost", err)
	}
	if err := fixture.service.Complete(t.Context(), takenOver.Lease, response); err != nil {
		t.Fatalf("complete current lease: %v", err)
	}
}

func TestReceiptExpiryStartsFreshLifecycleAndPurges(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.expiry")
	first, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("first"))
	if err != nil {
		t.Fatalf("begin first receipt: %v", err)
	}
	fixture.clock.Advance(fixture.config.RetentionDuration)

	restarted, err := fixture.service.Begin(t.Context(), fixture.scope, fixture.key, []byte("different after expiry"))
	if err != nil {
		t.Fatalf("restart expired receipt: %v", err)
	}
	if restarted.State != BeginStateNew || restarted.Reclaimed || restarted.Lease.Generation != 1 ||
		restarted.Lease.ReceiptID == first.Lease.ReceiptID {
		t.Fatalf("expired restart = %#v, want a fresh receipt lifecycle", restarted)
	}

	response, err := NewResponse(204, nil, "application/json")
	if err != nil {
		t.Fatalf("create response: %v", err)
	}
	if err := fixture.service.Complete(t.Context(), first.Lease, response); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("expired receipt completion error = %v, want ErrLeaseLost", err)
	}
	if err := fixture.service.Complete(t.Context(), restarted.Lease, response); err != nil {
		t.Fatalf("complete restarted receipt: %v", err)
	}
	fixture.clock.Advance(fixture.config.RetentionDuration)
	deleted, err := fixture.service.PurgeExpired(t.Context(), 10)
	if err != nil {
		t.Fatalf("purge expired receipts: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("purged receipts = %d, want 1", deleted)
	}
}

func TestReceiptScopeSeparatesOptionalWorkspaceAndPrincipal(t *testing.T) {
	fixture := newIntegrationFixture(t, "test.receipts.scope")
	ctx := t.Context()
	principalID := uuid.New()
	workspaceA, workspaceB := uuid.New(), uuid.New()
	suffix := uuid.NewString()
	if _, err := fixture.database.Pool.Exec(ctx, `
		INSERT INTO public.users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Idempotency Actor')
	`, principalID, "idempotency-"+suffix, "idempotency-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert scope actor: %v", err)
	}
	if _, err := fixture.database.Pool.Exec(ctx, `
		INSERT INTO public.workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Scope A', $2, $3), ($4, 'Scope B', $5, $3)
	`, workspaceA, "scope-a-"+suffix, principalID, workspaceB, "scope-b-"+suffix); err != nil {
		t.Fatalf("insert scope workspaces: %v", err)
	}

	operation, err := ParseOperation("test.receipts.workspace_scope")
	if err != nil {
		t.Fatalf("parse scope operation: %v", err)
	}
	baseActor := auth.NewHumanActor(principalID)
	actors := []auth.Actor{baseActor}
	for _, workspaceID := range []uuid.UUID{workspaceA, workspaceB} {
		actor, withWorkspaceErr := baseActor.WithWorkspace(workspaceID)
		if withWorkspaceErr != nil {
			t.Fatalf("scope actor to workspace: %v", withWorkspaceErr)
		}
		actors = append(actors, actor)
	}
	for index, actor := range actors {
		scope, scopeErr := NewScope(actor, MethodPost, operation)
		if scopeErr != nil {
			t.Fatalf("create scope %d: %v", index, scopeErr)
		}
		result, beginErr := fixture.service.Begin(ctx, scope, fixture.key, []byte("same"))
		if beginErr != nil || result.State != BeginStateNew {
			t.Fatalf("scope %d begin = (%#v, %v), want independent new receipt", index, result, beginErr)
		}
	}

	otherScope, err := NewScope(auth.NewHumanActor(uuid.New()), MethodPost, operation)
	if err != nil {
		t.Fatalf("create other principal scope: %v", err)
	}
	result, err := fixture.service.Begin(ctx, otherScope, fixture.key, []byte("same"))
	if err != nil || result.State != BeginStateNew {
		t.Fatalf("other principal begin = (%#v, %v), want independent new receipt", result, err)
	}

	var rows int
	if err := fixture.database.Pool.QueryRow(ctx, `
		SELECT count(*)
		FROM public.api_idempotency_receipts
		WHERE route_operation = $1
	`, "test.receipts.workspace_scope").Scan(&rows); err != nil {
		t.Fatalf("count scoped receipts: %v", err)
	}
	if rows != 4 {
		t.Fatalf("scoped receipt rows = %d, want 4", rows)
	}
}
