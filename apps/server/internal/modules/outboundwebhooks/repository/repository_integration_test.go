//go:build integration

package outboundwebhooksrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryLifecycleIsTenantBoundIdempotentAndSerialPerEndpoint(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newOutboundWebhookFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	endpointID, createAuditID := uuid.New(), uuid.New()
	endpoint, err := repository.CreateEndpoint(ctx, outboundwebhooksdomain.CreateEndpoint{
		ID: endpointID, AuditID: createAuditID, WorkspaceID: fixture.workspaceA,
		OwnerPrincipalID: fixture.principalA, Actor: fixture.actorA,
		Name: "Production events", URL: "https://hooks.example.com/receive",
		Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated},
		CreatedAt:     createdAt, RequestID: "create-request",
	}, strings.Repeat("e", 64))
	if err != nil {
		t.Fatalf("create endpoint: %v", err)
	}
	if endpoint.WorkspaceID != fixture.workspaceA || endpoint.OwnerPrincipalID != fixture.principalA || endpoint.SecretGeneration != 1 {
		t.Fatalf("created endpoint = %+v", endpoint)
	}
	if _, err := repository.GetEndpoint(ctx, fixture.workspaceB, endpointID); !errors.Is(err, outboundwebhooksdomain.ErrEndpointNotFound) {
		t.Fatalf("cross-tenant GetEndpoint() error = %v", err)
	}
	listed, err := repository.ListEndpoints(ctx, fixture.workspaceA, nil, 2)
	if err != nil || len(listed) != 1 || len(listed[0].Subscriptions) != 1 || listed[0].Subscriptions[0] != outboundwebhooksdomain.EventStoryCreated {
		t.Fatalf("ListEndpoints(workspace A) = %+v, %v", listed, err)
	}
	listed, err = repository.ListEndpoints(ctx, fixture.workspaceB, nil, 2)
	if err != nil || len(listed) != 0 {
		t.Fatalf("ListEndpoints(workspace B) = %+v, %v", listed, err)
	}

	newGeneration, err := repository.RotateEndpointSecret(
		ctx, fixture.actorA, fixture.workspaceA, endpointID, uuid.New(), 1,
		strings.Repeat("r", 64), createdAt.Add(25*time.Hour), createdAt.Add(time.Hour), "rotate-request",
	)
	if err != nil || newGeneration != 2 {
		t.Fatalf("RotateEndpointSecret() = %d, %v", newGeneration, err)
	}
	if _, err := repository.RotateEndpointSecret(
		ctx, fixture.actorA, fixture.workspaceA, endpointID, uuid.New(), 1,
		strings.Repeat("s", 64), createdAt.Add(26*time.Hour), createdAt.Add(2*time.Hour), "stale-rotate",
	); !errors.Is(err, outboundwebhooksdomain.ErrEndpointConflict) {
		t.Fatalf("stale RotateEndpointSecret() error = %v", err)
	}

	firstEvent := fixture.event(uuid.New(), outboundwebhooksdomain.EventStoryCreated, createdAt.Add(3*time.Hour), json.RawMessage(`{"story_id":"one","position":1}`))
	firstBody := envelopeBody(t, firstEvent)
	deliveries, err := repository.PublishEvent(ctx, firstEvent, firstBody)
	if err != nil || len(deliveries) != 1 {
		t.Fatalf("first PublishEvent() deliveries=%d error=%v", len(deliveries), err)
	}
	// PostgreSQL jsonb owns semantic payload equality. Whitespace and object-key
	// order differences must remain an idempotent replay of the same event ID.
	semanticallyEqual := firstEvent
	semanticallyEqual.Payload = json.RawMessage(`{ "position": 1, "story_id": "one" }`)
	deliveries, err = repository.PublishEvent(ctx, semanticallyEqual, envelopeBody(t, semanticallyEqual))
	if err != nil || len(deliveries) != 0 {
		t.Fatalf("idempotent PublishEvent() deliveries=%d error=%v", len(deliveries), err)
	}
	mismatched := firstEvent
	mismatched.Payload = json.RawMessage(`{"story_id":"different","position":1}`)
	if _, err := repository.PublishEvent(ctx, mismatched, envelopeBody(t, mismatched)); !errors.Is(err, outboundwebhooksdomain.ErrEndpointConflict) || !errors.Is(err, outboundwebhooksdomain.ErrInvalidPayload) {
		t.Fatalf("mismatched duplicate PublishEvent() error = %v", err)
	}

	secondEvent := fixture.event(uuid.New(), outboundwebhooksdomain.EventStoryCreated, createdAt.Add(4*time.Hour), json.RawMessage(`{"story_id":"two"}`))
	if deliveries, err = repository.PublishEvent(ctx, secondEvent, envelopeBody(t, secondEvent)); err != nil || len(deliveries) != 1 {
		t.Fatalf("second PublishEvent() deliveries=%d error=%v", len(deliveries), err)
	}

	claimAt := createdAt.Add(5 * time.Hour)
	firstClaim, err := repository.ClaimNextDelivery(ctx, uuid.New(), claimAt, claimAt.Add(30*time.Second))
	if err != nil {
		t.Fatalf("claim first delivery: %v", err)
	}
	if firstClaim.SecretGeneration != 2 || firstClaim.PreviousSecretGeneration == nil || *firstClaim.PreviousSecretGeneration != 1 {
		t.Fatalf("claimed rotation state = %+v", firstClaim)
	}
	if _, err := repository.ClaimNextDelivery(ctx, uuid.New(), claimAt, claimAt.Add(30*time.Second)); !errors.Is(err, outboundwebhooksdomain.ErrDeliveryNotFound) {
		t.Fatalf("parallel claim for same endpoint error = %v, want no work", err)
	}
	status := 204
	if err := repository.CompleteAttempt(ctx, outboundwebhooksdomain.DeliveryAttempt{
		ID: uuid.New(), DeliveryID: firstClaim.ID, LeaseToken: firstClaim.LeaseToken,
		AttemptNumber: firstClaim.AttemptNumber, Outcome: outboundwebhooksdomain.AttemptSucceeded,
		HTTPStatus: &status, Duration: 10 * time.Millisecond, StartedAt: claimAt,
		FinishedAt: claimAt.Add(10 * time.Millisecond), DisableAfterFailures: 20,
	}, fixture.workspaceA, endpointID); err != nil {
		t.Fatalf("complete first delivery: %v", err)
	}

	secondClaimAt := claimAt.Add(time.Minute)
	secondClaim, err := repository.ClaimNextDelivery(ctx, uuid.New(), secondClaimAt, secondClaimAt.Add(30*time.Second))
	if err != nil || secondClaim.ID == firstClaim.ID {
		t.Fatalf("claim second delivery = %+v, %v", secondClaim, err)
	}
	nextAttemptAt := secondClaimAt.Add(time.Minute)
	if err := repository.CompleteAttempt(ctx, outboundwebhooksdomain.DeliveryAttempt{
		ID: uuid.New(), DeliveryID: secondClaim.ID, LeaseToken: secondClaim.LeaseToken,
		AttemptNumber: secondClaim.AttemptNumber, Outcome: outboundwebhooksdomain.AttemptRetryScheduled,
		ErrorCode: "network_error", Duration: time.Second, StartedAt: secondClaimAt,
		FinishedAt: secondClaimAt.Add(time.Second), NextAttemptAt: &nextAttemptAt,
		CountEndpointFailure: true, DisableAfterFailures: 20,
	}, fixture.workspaceA, endpointID); err != nil {
		t.Fatalf("schedule second delivery retry: %v", err)
	}
	endpoint, err = repository.GetEndpoint(ctx, fixture.workspaceA, endpointID)
	if err != nil || endpoint.ConsecutiveFailures != 1 {
		t.Fatalf("endpoint after retry = %+v, %v", endpoint, err)
	}

	assertImmutableOutboundRow(t, ctx, postgres.Pool,
		"UPDATE outbound_webhook_audit_events SET result = 'failed' WHERE audit_event_id = $1", createAuditID)
	assertImmutableOutboundRow(t, ctx, postgres.Pool,
		"DELETE FROM outbound_webhook_delivery_attempts WHERE delivery_id = $1", firstClaim.ID)
}

func TestRepositoryStopsFanoutAndDispatchWhenOwnerLosesAccess(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	fixture := newOutboundWebhookFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	now := time.Unix(1_700_100_000, 0).UTC()
	endpointID := uuid.New()
	_, err := repository.CreateEndpoint(ctx, outboundwebhooksdomain.CreateEndpoint{
		ID: endpointID, AuditID: uuid.New(), WorkspaceID: fixture.workspaceA,
		OwnerPrincipalID: fixture.principalA, Actor: fixture.actorA,
		Name: "Revocation test", URL: "https://hooks.example.com/receive",
		Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated}, CreatedAt: now,
	}, strings.Repeat("e", 64))
	if err != nil {
		t.Fatalf("create endpoint: %v", err)
	}
	beforeRevocation := fixture.event(uuid.New(), outboundwebhooksdomain.EventStoryCreated, now.Add(time.Minute), json.RawMessage(`{"story_id":"before"}`))
	if deliveries, err := repository.PublishEvent(ctx, beforeRevocation, envelopeBody(t, beforeRevocation)); err != nil || len(deliveries) != 1 {
		t.Fatalf("publish before revocation deliveries=%d error=%v", len(deliveries), err)
	}
	disabledAt := now.Add(2 * time.Minute)
	if _, err := postgres.Pool.Exec(ctx, `
		UPDATE principals
		SET status = 'disabled', disabled_at = $2, disabled_reason = 'membership revoked', updated_at = $2
		WHERE principal_id = $1
	`, fixture.principalA, disabledAt); err != nil {
		t.Fatalf("disable principal: %v", err)
	}
	if _, err := repository.ClaimNextDelivery(ctx, uuid.New(), disabledAt, disabledAt.Add(30*time.Second)); !errors.Is(err, outboundwebhooksdomain.ErrDeliveryNotFound) {
		t.Fatalf("claim after principal disable error = %v", err)
	}
	afterRevocation := fixture.event(uuid.New(), outboundwebhooksdomain.EventStoryCreated, now.Add(3*time.Minute), json.RawMessage(`{"story_id":"after"}`))
	if deliveries, err := repository.PublishEvent(ctx, afterRevocation, envelopeBody(t, afterRevocation)); err != nil || len(deliveries) != 0 {
		t.Fatalf("publish after revocation deliveries=%d error=%v", len(deliveries), err)
	}
}

type outboundWebhookFixture struct {
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	userA      uuid.UUID
	principalA uuid.UUID
	actorA     platformauth.Actor
}

func newOutboundWebhookFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) outboundWebhookFixture {
	t.Helper()
	workspaceA := insertOutboundWebhookWorkspace(t, ctx, pool, "a")
	workspaceB := insertOutboundWebhookWorkspace(t, ctx, pool, "b")
	userA := insertOutboundWebhookUser(t, ctx, pool, "admin-a")
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST('admin' AS user_role))
	`, workspaceA, userA); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
	principalA := uuid.New()
	now := time.Unix(1_699_999_000, 0).UTC()
	if _, err := pool.Exec(ctx, `
		INSERT INTO principals (
			principal_id, workspace_id, kind, name, subject_user_id, status,
			created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, 'human_user', 'Admin A', $3, 'active', $3, $4, $4)
	`, principalA, workspaceA, userA, now); err != nil {
		t.Fatalf("insert human principal: %v", err)
	}
	actor, err := platformauth.NewHumanActor(userA).WithWorkspace(workspaceA)
	if err != nil {
		t.Fatalf("create human actor: %v", err)
	}
	return outboundWebhookFixture{
		workspaceA: workspaceA, workspaceB: workspaceB, userA: userA, principalA: principalA, actorA: actor,
	}
}

func (fixture outboundWebhookFixture) event(
	id uuid.UUID,
	eventType outboundwebhooksdomain.EventType,
	occurredAt time.Time,
	payload json.RawMessage,
) outboundwebhooksdomain.Event {
	return outboundwebhooksdomain.Event{
		ID: id, WorkspaceID: fixture.workspaceA, Type: eventType,
		PayloadVersion: outboundwebhooksdomain.PayloadVersion, SubjectType: eventType.SubjectType(),
		SubjectID: uuid.New(), Actor: fixture.actorA, Payload: payload,
		OccurredAt: occurredAt.UTC(), CreatedAt: occurredAt.Add(time.Second).UTC(),
	}
}

func envelopeBody(t *testing.T, event outboundwebhooksdomain.Event) []byte {
	t.Helper()
	body, err := outboundwebhooksdomain.NewEnvelope(event)
	if err != nil {
		t.Fatalf("build webhook envelope: %v", err)
	}
	return body
}

func insertOutboundWebhookWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, id, "Outbound "+label, "outbound-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertOutboundWebhookUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, TRUE, FALSE)
	`, id, label+"-"+id.String(), fmt.Sprintf("%s-%s@example.com", label, id), "Outbound "+label); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func assertImmutableOutboundRow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, id uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(ctx, query, id)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "55000" {
		t.Fatalf("immutable mutation error = %v, want SQLSTATE 55000", err)
	}
}
