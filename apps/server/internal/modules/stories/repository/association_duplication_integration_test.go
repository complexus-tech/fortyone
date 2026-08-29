//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStoryDuplicationIsTenantFencedAtomicConcurrentAndPlanBacked(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 75*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	sourceID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)
	scope := mutationScopeForFixture(t, fixture)
	source, err := repository.GetStoryForMutation(ctx, scope, sourceID)
	if err != nil {
		t.Fatalf("load duplication source: %v", err)
	}
	assertStoryDuplicationSourcePlan(t, ctx, postgres.Pool, sourceID, fixture.workspaceID)

	targetID := uuid.New()
	command := duplicateStoryCommand(t, scope, source, targetID, baseTime.Add(time.Minute))
	result, err := repository.DuplicateStoryMutation(ctx, command)
	if err != nil {
		t.Fatalf("duplicate authorized story: %v", err)
	}
	if result.Story.ID != targetID || result.Story.Title != "Copy of "+source.Title ||
		result.Story.Reporter == nil || *result.Story.Reporter != fixture.actorID {
		t.Fatalf("duplicated story = %#v", result.Story)
	}
	assertMutationEventCount(t, ctx, postgres, targetID, "story.created", 1)
	assertMutationActivityCount(t, ctx, postgres, targetID, "create", 1)

	t.Run("outbox conflict rolls back story media sequence and activity", func(t *testing.T) {
		rollbackTargetID := uuid.New()
		rollbackCommand := duplicateStoryCommand(
			t, scope, source, rollbackTargetID, baseTime.Add(2*time.Minute),
		)
		if err := postgres.Pool.QueryRow(
			ctx,
			"SELECT event_id FROM story_mutation_events WHERE story_id = $1 AND event_type = 'story.created'",
			sourceID,
		).Scan(&rollbackCommand.Event.ID); err != nil {
			t.Fatalf("load colliding event id: %v", err)
		}
		if _, err := repository.DuplicateStoryMutation(ctx, rollbackCommand); !errors.Is(err, storydomain.ErrMutationConflict) {
			t.Fatalf("duplicate rollback error = %v, want mutation conflict", err)
		}
		assertMutationRowCount(t, ctx, postgres, "stories", "id", rollbackTargetID, 0)
		assertMutationRowCount(t, ctx, postgres, "story_activities", "story_id", rollbackTargetID, 0)
		var currentSequence int
		if err := postgres.Pool.QueryRow(
			ctx,
			"SELECT current_sequence FROM team_story_sequences WHERE workspace_id = $1 AND team_id = $2",
			fixture.workspaceID,
			fixture.teamID,
		).Scan(&currentSequence); err != nil {
			t.Fatalf("read sequence after duplicate rollback: %v", err)
		}
		if currentSequence != result.Story.SequenceID {
			t.Fatalf("sequence advanced across rollback: got %d want %d", currentSequence, result.Story.SequenceID)
		}
	})

	t.Run("cross tenant source is hidden and leaves no target", func(t *testing.T) {
		foreignFixture := foreignSecondaryMutationFixture(fixture)
		foreignSourceID := createSecondaryMutationStory(
			t, ctx, repository, foreignFixture, baseTime.Add(3*time.Minute),
		)
		foreignSource, err := repository.GetStoryForMutation(
			ctx, mutationScopeForFixture(t, foreignFixture), foreignSourceID,
		)
		if err != nil {
			t.Fatalf("load foreign duplication source fixture: %v", err)
		}
		crossTenantTargetID := uuid.New()
		crossTenant := duplicateStoryCommand(
			t, scope, foreignSource, crossTenantTargetID, baseTime.Add(4*time.Minute),
		)
		if _, err := repository.DuplicateStoryMutation(ctx, crossTenant); !errors.Is(err, storydomain.ErrNotFound) {
			t.Fatalf("cross-tenant duplicate error = %v, want not found", err)
		}
		assertMutationRowCount(t, ctx, postgres, "stories", "id", crossTenantTargetID, 0)
		assertMutationEventIDAbsent(t, ctx, postgres.Pool, crossTenant.Event.ID)
	})

	t.Run("concurrent duplicate commands allocate unique sequences", func(t *testing.T) {
		commands := []storydomain.DuplicateStoryCommand{
			duplicateStoryCommand(t, scope, source, uuid.New(), baseTime.Add(5*time.Minute)),
			duplicateStoryCommand(t, scope, source, uuid.New(), baseTime.Add(6*time.Minute)),
		}
		start := make(chan struct{})
		outcomes := make(chan storydomain.DuplicateStoryResult, len(commands))
		errorsChannel := make(chan error, len(commands))
		var workers sync.WaitGroup
		for _, candidate := range commands {
			candidate := candidate
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				outcome, duplicateErr := repository.DuplicateStoryMutation(ctx, candidate)
				outcomes <- outcome
				errorsChannel <- duplicateErr
			}()
		}
		close(start)
		workers.Wait()
		close(outcomes)
		close(errorsChannel)
		for duplicateErr := range errorsChannel {
			if duplicateErr != nil {
				t.Fatalf("concurrent duplicate error: %v", duplicateErr)
			}
		}
		sequences := make(map[int]struct{}, len(commands))
		for outcome := range outcomes {
			sequences[outcome.Story.SequenceID] = struct{}{}
		}
		if len(sequences) != len(commands) {
			t.Fatalf("concurrent duplicate sequences = %v, want unique", sequences)
		}
	})

	t.Run("source CAS and live membership are rechecked", func(t *testing.T) {
		staleTargetID := uuid.New()
		stale := duplicateStoryCommand(t, scope, source, staleTargetID, baseTime.Add(7*time.Minute))
		current, err := repository.GetStoryForMutation(ctx, scope, sourceID)
		if err != nil {
			t.Fatalf("load source before CAS update: %v", err)
		}
		if _, err := repository.ApplyStoryMutation(
			ctx,
			storyUpdateMutationCommand(
				t, fixture, sourceID, current.UpdatedAt, baseTime.Add(8*time.Minute), "Changed source",
			),
		); err != nil {
			t.Fatalf("change source before stale duplicate: %v", err)
		}
		if _, err := repository.DuplicateStoryMutation(ctx, stale); !errors.Is(err, storydomain.ErrMutationConflict) {
			t.Fatalf("stale duplicate error = %v, want conflict", err)
		}
		assertMutationRowCount(t, ctx, postgres, "stories", "id", staleTargetID, 0)

		current, err = repository.GetStoryForMutation(ctx, scope, sourceID)
		if err != nil {
			t.Fatalf("reload source before membership revoke: %v", err)
		}
		revokedTargetID := uuid.New()
		revoked := duplicateStoryCommand(t, scope, current, revokedTargetID, baseTime.Add(9*time.Minute))
		mustMutationExec(
			t, ctx, postgres.Pool,
			"DELETE FROM team_members WHERE team_id = $1 AND user_id = $2",
			fixture.teamID, fixture.actorID,
		)
		if _, err := repository.DuplicateStoryMutation(ctx, revoked); !errors.Is(err, storydomain.ErrMutationForbidden) {
			t.Fatalf("revoked duplicate error = %v, want forbidden", err)
		}
		assertMutationRowCount(t, ctx, postgres, "stories", "id", revokedTargetID, 0)
	})
}

func TestStoryAssociationMutationUsesCASAndRollsBackEventsAtomically(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryMutationFixture(t, ctx, postgres.Pool)
	repository := NewMutationRepository(nil, postgres.Pool)
	baseTime := time.Date(2026, time.August, 28, 19, 0, 0, 0, time.UTC)
	scope := mutationScopeForFixture(t, fixture)
	fromID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime)
	toID := createSecondaryMutationStory(t, ctx, repository, fixture, baseTime.Add(time.Minute))

	association := storydomain.AssociationSnapshot{
		ID: uuid.New(), FromStoryID: fromID, ToStoryID: toID, Type: "related",
	}
	add := associationMutationCommand(
		t, scope, storydomain.AssociationMutationAdd, association, nil, baseTime.Add(2*time.Minute),
	)
	created, err := repository.ApplyStoryAssociationMutation(ctx, add)
	if err != nil {
		t.Fatalf("add story association: %v", err)
	}
	if created.FromStoryTitle == "" || created.ToStoryTitle == "" {
		t.Fatalf("association titles were not resolved: %#v", created)
	}
	assertMutationEventCount(t, ctx, postgres, fromID, "story.updated", 1)
	assertMutationEventCount(t, ctx, postgres, toID, "story.updated", 1)

	t.Run("outbox conflict rolls back association", func(t *testing.T) {
		candidate := storydomain.AssociationSnapshot{
			ID: uuid.New(), FromStoryID: fromID, ToStoryID: toID, Type: "blocking",
		}
		command := associationMutationCommand(
			t, scope, storydomain.AssociationMutationAdd, candidate, nil, baseTime.Add(3*time.Minute),
		)
		if err := postgres.Pool.QueryRow(
			ctx,
			"SELECT event_id FROM story_mutation_events WHERE story_id = $1 AND event_type = 'story.created'",
			fromID,
		).Scan(&command.Events[0].ID); err != nil {
			t.Fatalf("load association event collision: %v", err)
		}
		if _, err := repository.ApplyStoryAssociationMutation(ctx, command); !errors.Is(err, storydomain.ErrMutationConflict) {
			t.Fatalf("association rollback error = %v, want conflict", err)
		}
		assertStoryAssociationAbsent(t, ctx, postgres.Pool, candidate.ID)
		assertMutationEventIDAbsent(t, ctx, postgres.Pool, command.Events[1].ID)
	})

	t.Run("cross tenant target leaves no association", func(t *testing.T) {
		foreignFixture := foreignSecondaryMutationFixture(fixture)
		foreignID := createSecondaryMutationStory(
			t, ctx, repository, foreignFixture, baseTime.Add(4*time.Minute),
		)
		candidate := storydomain.AssociationSnapshot{
			ID: uuid.New(), FromStoryID: fromID, ToStoryID: foreignID, Type: "related",
		}
		command := associationMutationCommand(
			t, scope, storydomain.AssociationMutationAdd, candidate, nil, baseTime.Add(5*time.Minute),
		)
		if _, err := repository.ApplyStoryAssociationMutation(ctx, command); !errors.Is(err, storydomain.ErrNotFound) {
			t.Fatalf("cross-tenant association error = %v, want not found", err)
		}
		assertStoryAssociationAbsent(t, ctx, postgres.Pool, candidate.ID)
		for _, event := range command.Events {
			assertMutationEventIDAbsent(t, ctx, postgres.Pool, event.ID)
		}
	})

	t.Run("concurrent CAS update accepts exactly one", func(t *testing.T) {
		expected, err := repository.PrepareStoryAssociationMutation(ctx, scope, association.ID)
		if err != nil {
			t.Fatalf("prepare association update: %v", err)
		}
		commands := []storydomain.AssociationMutationCommand{
			associationMutationCommand(t, scope, storydomain.AssociationMutationUpdate,
				storydomain.AssociationSnapshot{ID: association.ID, FromStoryID: fromID, ToStoryID: toID, Type: "blocking"},
				&expected, baseTime.Add(6*time.Minute)),
			associationMutationCommand(t, scope, storydomain.AssociationMutationUpdate,
				storydomain.AssociationSnapshot{ID: association.ID, FromStoryID: fromID, ToStoryID: toID, Type: "duplicate"},
				&expected, baseTime.Add(7*time.Minute)),
		}
		start := make(chan struct{})
		outcomes := make(chan error, len(commands))
		var workers sync.WaitGroup
		for _, command := range commands {
			command := command
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				_, applyErr := repository.ApplyStoryAssociationMutation(ctx, command)
				outcomes <- applyErr
			}()
		}
		close(start)
		workers.Wait()
		close(outcomes)
		successes, conflicts := 0, 0
		for outcome := range outcomes {
			switch {
			case outcome == nil:
				successes++
			case errors.Is(outcome, storydomain.ErrMutationConflict):
				conflicts++
			default:
				t.Fatalf("unexpected association CAS result: %v", outcome)
			}
		}
		if successes != 1 || conflicts != 1 {
			t.Fatalf("association CAS outcomes successes=%d conflicts=%d", successes, conflicts)
		}
	})
}

func duplicateStoryCommand(
	t *testing.T,
	scope storydomain.MutationScope,
	source storydomain.Story,
	targetID uuid.UUID,
	occurredAt time.Time,
) storydomain.DuplicateStoryCommand {
	t.Helper()
	reporterID := *scope.ActivityUser
	return storydomain.DuplicateStoryCommand{
		Scope: scope, SourceStoryID: source.ID, TargetStoryID: targetID,
		ExpectedSourceUpdatedAt: source.UpdatedAt, ReporterID: reporterID, OccurredAt: occurredAt,
		Event: storydomain.MutationEvent{
			ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: targetID,
			Type: storydomain.MutationEventStoryCreated, Actor: scope.Actor,
			Payload: mustMutationJSON(t, map[string]any{
				"storyId": targetID, "workspaceId": scope.WorkspaceID,
			}),
			OccurredAt: occurredAt,
		},
		Activity: storydomain.MutationActivity{
			ID: uuid.New(), StoryID: targetID, UserID: reporterID,
			Type: "create", Field: "story", CurrentValue: "Copy of " + source.Title,
			OldValue: mustMutationJSON(t, nil), NewValue: mustMutationJSON(t, targetID),
			WorkspaceID: scope.WorkspaceID, CreatedAt: occurredAt,
		},
	}
}

func associationMutationCommand(
	t *testing.T,
	scope storydomain.MutationScope,
	action storydomain.AssociationMutationAction,
	association storydomain.AssociationSnapshot,
	expected *storydomain.AssociationSnapshot,
	occurredAt time.Time,
) storydomain.AssociationMutationCommand {
	t.Helper()
	storyIDs := []uuid.UUID{association.FromStoryID, association.ToStoryID}
	if expected != nil {
		storyIDs = append(storyIDs, expected.FromStoryID, expected.ToStoryID)
	}
	seen := make(map[uuid.UUID]struct{}, len(storyIDs))
	events := make([]storydomain.MutationEvent, 0, len(storyIDs))
	activities := make([]storydomain.MutationActivity, 0, len(storyIDs))
	for _, storyID := range storyIDs {
		if _, exists := seen[storyID]; exists {
			continue
		}
		seen[storyID] = struct{}{}
		events = append(events, storydomain.MutationEvent{
			ID: uuid.New(), WorkspaceID: scope.WorkspaceID, StoryID: storyID,
			Type: storydomain.MutationEventStoryUpdated, Actor: scope.Actor,
			Payload: mustMutationJSON(t, map[string]any{
				"storyId": storyID, "associationId": association.ID, "action": action,
			}),
			OccurredAt: occurredAt,
		})
		activities = append(activities, storydomain.MutationActivity{
			ID: uuid.New(), StoryID: storyID, UserID: *scope.ActivityUser,
			Type: "update", Field: "association", CurrentValue: string(action),
			OldValue: mustMutationJSON(t, nil), NewValue: mustMutationJSON(t, association.ID),
			WorkspaceID: scope.WorkspaceID, CreatedAt: occurredAt,
		})
	}
	return storydomain.AssociationMutationCommand{
		Scope: scope, Action: action, Association: association, Expected: expected,
		OccurredAt: occurredAt, Events: events, Activities: activities,
	}
}

func assertMutationEventIDAbsent(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	eventID uuid.UUID,
) {
	t.Helper()
	var count int
	if err := pool.QueryRow(
		ctx, "SELECT COUNT(*) FROM story_mutation_events WHERE event_id = $1", eventID,
	).Scan(&count); err != nil {
		t.Fatalf("count mutation event id: %v", err)
	}
	if count != 0 {
		t.Fatalf("mutation event %s exists after rollback", eventID)
	}
}

func assertStoryAssociationAbsent(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	associationID uuid.UUID,
) {
	t.Helper()
	var count int
	if err := pool.QueryRow(
		ctx, "SELECT COUNT(*) FROM story_associations WHERE id = $1", associationID,
	).Scan(&count); err != nil {
		t.Fatalf("count story association: %v", err)
	}
	if count != 0 {
		t.Fatalf("story association %s exists after rollback", associationID)
	}
}

func assertStoryDuplicationSourcePlan(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	storyID, workspaceID uuid.UUID,
) {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin duplication plan transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		t.Fatalf("disable sequential scan for duplication plan: %v", err)
	}
	rows, err := tx.Query(
		ctx,
		`EXPLAIN (COSTS OFF)
		 SELECT story.id
		 FROM public.stories AS story
		 WHERE story.id = $1 AND story.workspace_id = $2 AND story.deleted_at IS NULL`,
		storyID,
		workspaceID,
	)
	if err != nil {
		t.Fatalf("explain duplication source lookup: %v", err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan duplication source plan: %v", err)
		}
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read duplication source plan: %v", err)
	}
	if !strings.Contains(plan.String(), "stories_id_workspace_unique") {
		t.Fatalf("duplication source lookup does not use stories_id_workspace_unique:\n%s", plan.String())
	}
}
