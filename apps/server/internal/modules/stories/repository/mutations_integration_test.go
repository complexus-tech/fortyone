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
)

func TestStoryMutationRepositoryLifecycleIsAtomicScopedAndCompareAndSwapSafe(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)

	scope := mutationScopeForFixture(t, fixture)
	preconditions, err := repository.PrepareStoryMutation(ctx, scope, fixture.teamID, nil)
	if err != nil {
		t.Fatalf("prepare authorized story mutation: %v", err)
	}
	if preconditions.EstimateScheme != "tshirt" {
		t.Fatalf("estimate scheme = %q, want tshirt", preconditions.EstimateScheme)
	}
	if _, err := repository.PrepareStoryMutation(ctx, scope, fixture.foreignTeamID, nil); !errors.Is(err, storydomain.ErrMutationForbidden) {
		t.Fatalf("cross-tenant team authorization error = %v, want forbidden", err)
	}

	storyID := uuid.New()
	createCommand := storyCreateMutationCommand(t, fixture, storyID, baseTime)
	created, err := repository.CreateStoryMutation(ctx, createCommand)
	if err != nil {
		t.Fatalf("create story mutation: %v", err)
	}
	if !created.Created || created.Story.ID != storyID || created.Story.SequenceID != 1 {
		t.Fatalf("created story = %#v", created)
	}
	assertMutationRowCount(t, ctx, postgres, "stories", "id", storyID, 1)
	assertMutationRowCount(t, ctx, postgres, "story_activities", "story_id", storyID, 1)
	assertMutationEventCount(t, ctx, postgres, storyID, "story.created", 1)

	t.Run("reference failure rolls back story activity sequence and event", func(t *testing.T) {
		invalidStoryID := uuid.New()
		command := storyCreateMutationCommand(t, fixture, invalidStoryID, baseTime.Add(time.Minute))
		command.Story.Status = &fixture.foreignStatusID
		_, err := repository.CreateStoryMutation(ctx, command)
		if !errors.Is(err, storydomain.ErrInvalidMutation) {
			t.Fatalf("cross-tenant status error = %v, want invalid mutation", err)
		}
		assertMutationRowCount(t, ctx, postgres, "stories", "id", invalidStoryID, 0)
		assertMutationRowCount(t, ctx, postgres, "story_activities", "story_id", invalidStoryID, 0)
		assertMutationEventCount(t, ctx, postgres, invalidStoryID, "story.created", 0)

		var sequence int
		if err := postgres.Pool.QueryRow(
			ctx,
			"SELECT current_sequence FROM team_story_sequences WHERE workspace_id = $1 AND team_id = $2",
			fixture.workspaceID,
			fixture.teamID,
		).Scan(&sequence); err != nil {
			t.Fatalf("read story sequence after rollback: %v", err)
		}
		if sequence != 1 {
			t.Fatalf("sequence advanced across rolled-back mutation: got %d want 1", sequence)
		}
	})

	t.Run("concurrent updates accept exactly one expected version", func(t *testing.T) {
		current, err := repository.GetStoryForMutation(ctx, scope, storyID)
		if err != nil {
			t.Fatalf("load story before concurrent update: %v", err)
		}
		commands := []storydomain.UpdateStoryCommand{
			storyUpdateMutationCommand(t, fixture, storyID, current.UpdatedAt, baseTime.Add(2*time.Minute), "Concurrent A"),
			storyUpdateMutationCommand(t, fixture, storyID, current.UpdatedAt, baseTime.Add(3*time.Minute), "Concurrent B"),
		}
		start := make(chan struct{})
		results := make(chan error, len(commands))
		var workers sync.WaitGroup
		for _, command := range commands {
			command := command
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				_, applyErr := repository.ApplyStoryMutation(ctx, command)
				results <- applyErr
			}()
		}
		close(start)
		workers.Wait()
		close(results)

		successes, conflicts := 0, 0
		for resultErr := range results {
			switch {
			case resultErr == nil:
				successes++
			case errors.Is(resultErr, storydomain.ErrMutationConflict):
				conflicts++
			default:
				t.Fatalf("concurrent update returned unexpected error: %v", resultErr)
			}
		}
		if successes != 1 || conflicts != 1 {
			t.Fatalf("concurrent outcomes: successes=%d conflicts=%d", successes, conflicts)
		}

		updated, err := repository.GetStoryForMutation(ctx, scope, storyID)
		if err != nil {
			t.Fatalf("load concurrently updated story: %v", err)
		}
		if updated.Title != "Concurrent A" && updated.Title != "Concurrent B" {
			t.Fatalf("unexpected final title %q", updated.Title)
		}
		assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 1)
		assertMutationActivityCount(t, ctx, postgres, storyID, "update", 1)
	})

	t.Run("outbox failure rolls the business mutation back", func(t *testing.T) {
		current, err := repository.GetStoryForMutation(ctx, scope, storyID)
		if err != nil {
			t.Fatalf("load story before atomicity check: %v", err)
		}
		command := storyUpdateMutationCommand(
			t, fixture, storyID, current.UpdatedAt, baseTime.Add(4*time.Minute), "Must roll back",
		)
		command.Event.ID = createCommand.Event.ID
		_, err = repository.ApplyStoryMutation(ctx, command)
		if !errors.Is(err, storydomain.ErrMutationConflict) {
			t.Fatalf("duplicate outbox identity error = %v, want conflict", err)
		}
		after, err := repository.GetStoryForMutation(ctx, scope, storyID)
		if err != nil {
			t.Fatalf("load story after failed outbox write: %v", err)
		}
		if after.Title != current.Title || !after.UpdatedAt.Equal(current.UpdatedAt) {
			t.Fatalf("story changed despite failed outbox write: before=%#v after=%#v", current, after)
		}
		assertMutationActivityCount(t, ctx, postgres, storyID, "update", 1)
	})

	t.Run("live membership is rechecked before every write", func(t *testing.T) {
		current, err := repository.GetStoryForMutation(ctx, scope, storyID)
		if err != nil {
			t.Fatalf("load story before revoking access: %v", err)
		}
		command := storyUpdateMutationCommand(
			t, fixture, storyID, current.UpdatedAt, baseTime.Add(5*time.Minute), "Forbidden update",
		)
		mustMutationExec(
			t, ctx, postgres.Pool,
			"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
			fixture.teamID, fixture.actorID,
		)
		if _, err := repository.GetStoryForMutation(ctx, scope, storyID); !errors.Is(err, storydomain.ErrNotFound) {
			t.Fatalf("revoked actor read error = %v, want not found", err)
		}
		if _, err := repository.ApplyStoryMutation(ctx, command); !errors.Is(err, storydomain.ErrMutationForbidden) {
			t.Fatalf("revoked actor update error = %v, want forbidden", err)
		}
		assertMutationEventCount(t, ctx, postgres, storyID, "story.updated", 1)
		mustMutationExec(
			t, ctx, postgres.Pool,
			"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)",
			fixture.teamID, fixture.actorID,
		)
	})

	t.Run("delete rechecks reporter policy and commits its event", func(t *testing.T) {
		deleteStoryID := uuid.New()
		createDeleteTarget := storyCreateMutationCommand(t, fixture, deleteStoryID, baseTime.Add(6*time.Minute))
		createdDeleteTarget, err := repository.CreateStoryMutation(ctx, createDeleteTarget)
		if err != nil {
			t.Fatalf("create delete target: %v", err)
		}
		if createdDeleteTarget.Story.SequenceID != 2 {
			t.Fatalf("sequence after rejected create = %d, want 2", createdDeleteTarget.Story.SequenceID)
		}
		command := storyDeleteMutationCommand(
			t, fixture, deleteStoryID, createdDeleteTarget.Story.UpdatedAt, baseTime.Add(7*time.Minute),
		)
		result, err := repository.DeleteStoryMutation(ctx, command)
		if err != nil {
			t.Fatalf("delete story mutation: %v", err)
		}
		if !result.Deleted || result.DeletedAt == nil || !result.DeletedAt.Equal(command.Event.OccurredAt) {
			t.Fatalf("delete result = %#v", result)
		}
		assertMutationEventCount(t, ctx, postgres, deleteStoryID, "story.deleted", 1)
		assertMutationActivityCount(t, ctx, postgres, deleteStoryID, "delete", 1)
		if _, err := repository.GetStoryForMutation(ctx, scope, deleteStoryID); !errors.Is(err, storydomain.ErrNotFound) {
			t.Fatalf("deleted story remains visible to mutation reads: %v", err)
		}
	})

	t.Run("service-account scope team restriction and revocation are live", func(t *testing.T) {
		serviceScope, credentialID := seedStoryMutationServiceAccount(
			t, ctx, postgres.Pool, fixture, baseTime.Add(-time.Hour),
		)
		if _, err := repository.PrepareStoryMutation(ctx, serviceScope, fixture.teamID, nil); err != nil {
			t.Fatalf("prepare service-account mutation: %v", err)
		}
		if _, err := repository.PrepareStoryMutation(
			ctx, serviceScope, fixture.otherTeamID, nil,
		); !errors.Is(err, storydomain.ErrMutationForbidden) {
			t.Fatalf("restricted service-account team error = %v, want forbidden", err)
		}

		storyID := uuid.New()
		occurredAt := baseTime.Add(8 * time.Minute)
		statusID := fixture.statusID
		command := storydomain.CreateStoryCommand{
			Scope: serviceScope,
			Story: storydomain.Story{
				ID: storyID, Title: "Service-account story", Status: &statusID,
				Priority: "High", AutoSchedulingStatus: "off",
				Team: fixture.teamID, Workspace: fixture.workspaceID,
				CreatedAt: occurredAt, UpdatedAt: occurredAt,
			},
			Event: storydomain.MutationEvent{
				ID: uuid.New(), WorkspaceID: fixture.workspaceID, StoryID: storyID,
				Type: storydomain.MutationEventStoryCreated, Actor: serviceScope.Actor,
				Payload: mustMutationJSON(t, map[string]any{"story_id": storyID}), OccurredAt: occurredAt,
			},
		}
		if _, err := repository.CreateStoryMutation(ctx, command); err != nil {
			t.Fatalf("create story as service account: %v", err)
		}
		var actorKind string
		var storedCredentialID uuid.UUID
		if err := postgres.Pool.QueryRow(
			ctx,
			"SELECT actor_kind, actor_credential_id FROM story_mutation_events WHERE event_id = $1",
			command.Event.ID,
		).Scan(&actorKind, &storedCredentialID); err != nil {
			t.Fatalf("read service-account story event: %v", err)
		}
		if actorKind != "service_account" || storedCredentialID != credentialID {
			t.Fatalf("service-account event attribution = kind %q credential %s", actorKind, storedCredentialID)
		}

		revokedAt := occurredAt.Add(time.Minute)
		mustMutationExec(
			t, ctx, postgres.Pool,
			"UPDATE api_credentials SET revoked_at = $1, revoked_reason = 'integration test' WHERE credential_id = $2",
			revokedAt, credentialID,
		)
		if _, err := repository.PrepareStoryMutation(
			ctx, serviceScope, fixture.teamID, nil,
		); !errors.Is(err, storydomain.ErrMutationForbidden) {
			t.Fatalf("revoked service-account mutation error = %v, want forbidden", err)
		}
	})
}

func assertMutationRowCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	table, idColumn string,
	id uuid.UUID,
	want int,
) {
	t.Helper()
	allowed := map[string]string{
		"stories:id":                "SELECT COUNT(*) FROM stories WHERE id = $1",
		"story_activities:story_id": "SELECT COUNT(*) FROM story_activities WHERE story_id = $1",
	}
	query, ok := allowed[table+":"+idColumn]
	if !ok {
		t.Fatalf("unsupported mutation count assertion %s.%s", table, idColumn)
	}
	var count int
	if err := postgres.Pool.QueryRow(ctx, query, id).Scan(&count); err != nil {
		t.Fatalf("count %s rows: %v", table, err)
	}
	if count != want {
		t.Fatalf("%s count for %s = %d, want %d", table, id, count, want)
	}
}

func assertMutationEventCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	eventType string,
	want int,
) {
	t.Helper()
	var count int
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM story_mutation_events WHERE story_id = $1 AND event_type = $2",
		storyID,
		eventType,
	).Scan(&count); err != nil {
		t.Fatalf("count story mutation events: %v", err)
	}
	if count != want {
		t.Fatalf("%s event count for %s = %d, want %d", eventType, storyID, count, want)
	}
}

func assertMutationActivityCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	storyID uuid.UUID,
	activityType string,
	want int,
) {
	t.Helper()
	var count int
	if err := postgres.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM story_activities WHERE story_id = $1 AND activity_type = $2",
		storyID,
		activityType,
	).Scan(&count); err != nil {
		t.Fatalf("count story mutation activities: %v", err)
	}
	if count != want {
		t.Fatalf("%s activity count for %s = %d, want %d", activityType, storyID, count, want)
	}
}
