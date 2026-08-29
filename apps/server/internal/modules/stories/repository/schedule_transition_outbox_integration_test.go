//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func TestStoryScheduleTransitionOutboxIsAtomicSequencedAndConcurrentlyClaimed(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 16, 0, 0, 0, time.UTC)
	storyID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)
	expectedUpdatedAt := secondaryStoryUpdatedAt(t, ctx, postgres, storyID)

	first := scheduleTransitionInput(t, fixture, storyID, "planning-v1", baseTime.Add(time.Minute))
	applied, claimed, err := repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx, storyID, fixture.workspaceID, expectedUpdatedAt,
		"planning", nil, baseTime.Add(time.Minute), nil, first,
	)
	if err != nil || !applied || claimed != nil {
		t.Fatalf("persist first schedule transition: applied=%v claimed=%#v error=%v", applied, claimed, err)
	}

	reason := "Capacity found"
	second := scheduleTransitionInput(t, fixture, storyID, "scheduled-v1", baseTime.Add(2*time.Minute))
	applied, claimed, err = repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx, storyID, fixture.workspaceID, expectedUpdatedAt,
		"scheduled", &reason, baseTime.Add(2*time.Minute), nil, second,
	)
	if err != nil || !applied || claimed != nil {
		t.Fatalf("persist second schedule transition: applied=%v claimed=%#v error=%v", applied, claimed, err)
	}
	assertScheduleTransitionState(t, ctx, postgres, storyID, "scheduled", &reason)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 2)

	// An immediate semantic retry acknowledges the already-durable latest
	// transition without creating another externally observable event.
	applied, claimed, err = repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx, storyID, fixture.workspaceID, expectedUpdatedAt,
		"scheduled", &reason, baseTime.Add(3*time.Minute), nil, second,
	)
	if err != nil || !applied || claimed != nil {
		t.Fatalf("replay latest schedule transition: applied=%v claimed=%#v error=%v", applied, claimed, err)
	}
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 2)

	rollbackReason := "Must roll back"
	conflicting := scheduleTransitionInput(t, fixture, storyID, "at-risk-v1", baseTime.Add(4*time.Minute))
	conflicting.EventID = first.EventID
	if _, _, err := repository.UpdateAutoSchedulingStateAndClaimTransitionIfUnchanged(
		ctx, storyID, fixture.workspaceID, expectedUpdatedAt,
		"at_risk", &rollbackReason, baseTime.Add(4*time.Minute), nil, conflicting,
	); err == nil {
		t.Fatal("duplicate schedule event identity unexpectedly committed")
	}
	assertScheduleTransitionState(t, ctx, postgres, storyID, "scheduled", &reason)
	assertScheduleTransitionCount(t, ctx, postgres, storyID, 2)

	start := make(chan struct{})
	claims := make(chan []stories.CoreScheduleTransitionOutboxEvent, 2)
	errorsFound := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			batch, claimErr := repository.ClaimScheduleTransitionOutboxEvents(ctx, 1, time.Hour)
			if claimErr != nil {
				errorsFound <- claimErr
				return
			}
			claims <- batch
		}()
	}
	close(start)
	workers.Wait()
	close(claims)
	close(errorsFound)
	for claimErr := range errorsFound {
		t.Fatalf("claim schedule transition concurrently: %v", claimErr)
	}

	claimedByID := make(map[uuid.UUID]stories.CoreScheduleTransitionOutboxEvent, 2)
	for batch := range claims {
		if len(batch) != 1 {
			t.Fatalf("schedule claim batch size = %d, want 1", len(batch))
		}
		item := batch[0]
		if _, duplicate := claimedByID[item.EventID]; duplicate {
			t.Fatalf("schedule event %s was claimed twice", item.EventID)
		}
		if item.ClaimToken == uuid.Nil || item.AttemptCount != 1 {
			t.Fatalf("invalid schedule transition claim: %#v", item)
		}
		claimedByID[item.EventID] = item
	}
	if len(claimedByID) != 2 {
		t.Fatalf("claimed %d distinct schedule events, want 2", len(claimedByID))
	}
	for _, item := range claimedByID {
		if err := repository.CompleteScheduleTransitionOutboxEvent(ctx, item.EventID, uuid.New()); err == nil {
			t.Fatalf("wrong claim token completed schedule event %s", item.EventID)
		}
		if err := repository.CompleteScheduleTransitionOutboxEvent(ctx, item.EventID, item.ClaimToken); err != nil {
			t.Fatalf("complete schedule event %s: %v", item.EventID, err)
		}
	}
	var completed, maximumSequence int
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*), MAX(transition_sequence) FROM story_schedule_transition_outbox WHERE story_id = $1 AND status = 'completed'",
		storyID,
	).Scan(&completed, &maximumSequence); err != nil {
		t.Fatalf("read completed schedule transitions: %v", err)
	}
	if completed != 2 || maximumSequence != 2 {
		t.Fatalf("completed schedule transitions=%d max sequence=%d, want 2 and 2", completed, maximumSequence)
	}
}

func scheduleTransitionInput(
	t *testing.T,
	fixture storyMutationFixture,
	storyID uuid.UUID,
	fingerprint string,
	occurredAt time.Time,
) stories.CoreScheduleTransitionOutboxInput {
	t.Helper()
	envelope := events.Event{
		Type: events.StoryUpdated,
		Payload: events.StoryUpdatedPayload{
			StoryID: storyID, WorkspaceID: fixture.workspaceID,
			Updates: map[string]any{"auto_scheduling_status": fingerprint},
		},
		Timestamp: occurredAt,
		ActorID:   fixture.actorID,
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("encode schedule transition fixture: %v", err)
	}
	return stories.CoreScheduleTransitionOutboxInput{
		EventID: uuid.New(), StoryID: storyID, WorkspaceID: fixture.workspaceID,
		ActorID: fixture.actorID, SemanticFingerprint: fingerprint, EventPayload: payload,
	}
}

func secondaryStoryUpdatedAt(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
) time.Time {
	t.Helper()
	var updatedAt time.Time
	if err := postgres.Pool.QueryRow(ctx, "SELECT updated_at FROM stories WHERE id = $1", storyID).Scan(&updatedAt); err != nil {
		t.Fatalf("read story updated_at: %v", err)
	}
	return updatedAt
}

func assertScheduleTransitionState(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	wantStatus string,
	wantReason *string,
) {
	t.Helper()
	var status string
	var reason *string
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT auto_scheduling_status, auto_scheduling_reason FROM stories WHERE id = $1",
		storyID,
	).Scan(&status, &reason); err != nil {
		t.Fatalf("read schedule transition state: %v", err)
	}
	if status != wantStatus || !equalScheduleTransitionReason(reason, wantReason) {
		t.Fatalf("schedule state status=%q reason=%v, want status=%q reason=%v", status, reason, wantStatus, wantReason)
	}
}

func assertScheduleTransitionCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	want int,
) {
	t.Helper()
	var count int
	if err := postgres.Pool.QueryRow(
		ctx, "SELECT COUNT(*) FROM story_schedule_transition_outbox WHERE story_id = $1", storyID,
	).Scan(&count); err != nil {
		t.Fatalf("count schedule transitions: %v", err)
	}
	if count != want {
		t.Fatalf("schedule transition count = %d, want %d", count, want)
	}
}
