//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestStoryMutationOutboxUsesDisjointClaimsFencingAndImmutablePayloads(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)

	eventIDs := make([]uuid.UUID, 4)
	for index := range eventIDs {
		eventIDs[index] = uuid.New()
		insertMutationOutboxEvent(
			t, ctx, postgres, fixture, eventIDs[index], uuid.New(),
			storydomain.MutationEventStoryUpdated, baseTime.Add(time.Duration(index)*time.Second),
		)
	}

	claimNow := baseTime.Add(time.Minute)
	start := make(chan struct{})
	claims := make(chan []storydomain.MutationEvent, 2)
	errorsFound := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			events, err := repository.ClaimStoryMutationEvents(
				ctx, 2, claimNow, claimNow.Add(-2*time.Minute),
			)
			if err != nil {
				errorsFound <- err
				return
			}
			claims <- events
		}()
	}
	close(start)
	workers.Wait()
	close(claims)
	close(errorsFound)
	for claimErr := range errorsFound {
		t.Fatalf("claim mutation events concurrently: %v", claimErr)
	}

	claimedByID := make(map[uuid.UUID]storydomain.MutationEvent, len(eventIDs))
	for batch := range claims {
		if len(batch) != 2 {
			t.Fatalf("concurrent claim batch size = %d, want 2", len(batch))
		}
		for _, event := range batch {
			if _, duplicate := claimedByID[event.ID]; duplicate {
				t.Fatalf("event %s was leased to two workers", event.ID)
			}
			if event.ClaimToken == uuid.Nil || event.AttemptCount != 1 {
				t.Fatalf("invalid first claim: %#v", event)
			}
			claimedByID[event.ID] = event
		}
	}
	if len(claimedByID) != len(eventIDs) {
		t.Fatalf("claimed %d distinct events, want %d", len(claimedByID), len(eventIDs))
	}

	retryTarget := claimedByID[eventIDs[0]]
	if err := repository.CompleteStoryMutationEvent(
		ctx, retryTarget.ID, uuid.New(), claimNow.Add(time.Minute),
	); !errors.Is(err, storydomain.ErrMutationEventNotFound) {
		t.Fatalf("wrong completion fence error = %v, want event not found", err)
	}
	nextAttemptAt := claimNow.Add(5 * time.Minute)
	if err := repository.RetryStoryMutationEvent(
		ctx, retryTarget.ID, retryTarget.ClaimToken, nextAttemptAt, claimNow.Add(time.Minute), "temporary failure",
	); err != nil {
		t.Fatalf("schedule mutation event retry: %v", err)
	}
	for eventID, event := range claimedByID {
		if eventID == retryTarget.ID {
			continue
		}
		if err := repository.CompleteStoryMutationEvent(
			ctx, event.ID, event.ClaimToken, claimNow.Add(time.Minute),
		); err != nil {
			t.Fatalf("complete claimed mutation event %s: %v", event.ID, err)
		}
	}

	early, err := repository.ClaimStoryMutationEvents(
		ctx, 10, claimNow.Add(2*time.Minute), claimNow.Add(-time.Minute),
	)
	if err != nil {
		t.Fatalf("claim before retry is due: %v", err)
	}
	if len(early) != 0 {
		t.Fatalf("retry was visible before next_attempt_at: %#v", early)
	}
	retried, err := repository.ClaimStoryMutationEvents(
		ctx, 10, nextAttemptAt.Add(time.Second), nextAttemptAt.Add(-time.Minute),
	)
	if err != nil {
		t.Fatalf("claim due retry: %v", err)
	}
	if len(retried) != 1 || retried[0].ID != retryTarget.ID || retried[0].AttemptCount != 2 ||
		retried[0].ClaimToken == retryTarget.ClaimToken {
		t.Fatalf("retried claim = %#v", retried)
	}
	if err := repository.CompleteStoryMutationEvent(
		ctx, retried[0].ID, retried[0].ClaimToken, nextAttemptAt.Add(time.Minute),
	); err != nil {
		t.Fatalf("complete retried mutation event: %v", err)
	}

	staleEventID := uuid.New()
	insertMutationOutboxEvent(
		t, ctx, postgres, fixture, staleEventID, uuid.New(),
		storydomain.MutationEventStoryDeleted, nextAttemptAt.Add(2*time.Minute),
	)
	firstLeaseAt := nextAttemptAt.Add(3 * time.Minute)
	firstLease, err := repository.ClaimStoryMutationEvents(
		ctx, 1, firstLeaseAt, firstLeaseAt.Add(-2*time.Minute),
	)
	if err != nil || len(firstLease) != 1 || firstLease[0].ID != staleEventID {
		t.Fatalf("claim stale-event fixture = %#v, error %v", firstLease, err)
	}
	reclaimedAt := firstLeaseAt.Add(3 * time.Minute)
	reclaimed, err := repository.ClaimStoryMutationEvents(
		ctx, 1, reclaimedAt, firstLeaseAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("reclaim stale event: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0].ID != staleEventID ||
		reclaimed[0].ClaimToken == firstLease[0].ClaimToken || reclaimed[0].AttemptCount != 2 {
		t.Fatalf("stale reclaim = %#v", reclaimed)
	}
	if err := repository.CompleteStoryMutationEvent(
		ctx, reclaimed[0].ID, reclaimed[0].ClaimToken, reclaimedAt.Add(time.Minute),
	); err != nil {
		t.Fatalf("complete reclaimed mutation event: %v", err)
	}

	_, err = postgres.Pool.Exec(
		ctx,
		"UPDATE story_mutation_events SET payload = $1 WHERE event_id = $2",
		[]byte(`{"tampered":true}`),
		eventIDs[1],
	)
	var postgresErr *pgconn.PgError
	if !errors.As(err, &postgresErr) || postgresErr.Code != "55000" {
		t.Fatalf("payload identity mutation error = %v, want SQLSTATE 55000", err)
	}
}

func insertMutationOutboxEvent(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	fixture storyMutationFixture,
	eventID, storyID uuid.UUID,
	eventType storydomain.MutationEventType,
	occurredAt time.Time,
) {
	t.Helper()
	mustMutationExec(
		t,
		ctx,
		postgres.Pool,
		`INSERT INTO story_mutation_events (
			event_id, workspace_id, story_id, event_type, actor_kind, actor_id,
			payload, occurred_at, status, attempt_count, next_attempt_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'human_user', $5, $6, $7, 'pending', 0, $7, $7, $7)`,
		eventID,
		fixture.workspaceID,
		storyID,
		string(eventType),
		fixture.actorID,
		mustMutationJSON(t, map[string]any{"story_id": storyID}),
		occurredAt,
	)
}
