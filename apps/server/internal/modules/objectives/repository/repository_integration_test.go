//go:build integration

package objectivesrepository

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
)

func TestObjectiveRepositoryTenantTransactionsAndConcurrency(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	fixture := newObjectiveIntegrationFixture(t, ctx)
	assertObjectivePostgres18(t, ctx, fixture)

	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueObjectiveName("primary")))
	if err != nil {
		t.Fatalf("create objective aggregate: %v", err)
	}
	if created.Objective.SequenceID != 1 || len(created.KeyResults) != 2 ||
		created.KeyResults[0].SequenceID != 1 || created.KeyResults[1].SequenceID != 2 {
		t.Fatalf("created aggregate sequences = objective %d, key results %#v", created.Objective.SequenceID, created.KeyResults)
	}
	if got := objectiveRowCount(t, ctx, fixture.postgres.Pool,
		"SELECT COUNT(*) FROM okr_activities WHERE objective_id = $1", created.Objective.ID); got != 3 {
		t.Fatalf("activity count = %d, want 3", got)
	}
	if got := objectiveRowCount(t, ctx, fixture.postgres.Pool,
		"SELECT COUNT(*) FROM key_result_contributors WHERE key_result_id = $1", created.KeyResults[0].ID); got != 1 {
		t.Fatalf("deduplicated contributor count = %d, want 1", got)
	}

	foreign, err := fixture.repo.Create(ctx, fixture.foreignCreateCommand(uniqueObjectiveName("foreign")))
	if err != nil {
		t.Fatalf("create foreign objective: %v", err)
	}
	visible, err := fixture.repo.List(ctx, objectivesdomain.ListQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
	})
	if err != nil || len(visible) != 1 || visible[0].ID != created.Objective.ID {
		t.Fatalf("tenant-visible objectives = %#v, error=%v", visible, err)
	}
	foreignVisible, err := fixture.repo.List(ctx, objectivesdomain.ListQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.actorB,
	})
	if err != nil || len(foreignVisible) != 0 {
		t.Fatalf("cross-tenant list = %#v, error=%v", foreignVisible, err)
	}
	if _, err := fixture.repo.Get(ctx, objectivesdomain.GetQuery{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceB, ActorID: fixture.actorB,
	}); !errors.Is(err, objectivesdomain.ErrNotFound) {
		t.Fatalf("cross-tenant get error = %v, want ErrNotFound", err)
	}
	if _, err := fixture.repo.GetAnalytics(ctx, objectivesdomain.AnalyticsQuery{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.outsiderA,
	}, time.Now()); !errors.Is(err, objectivesdomain.ErrNotFound) {
		t.Fatalf("team outsider analytics error = %v, want ErrNotFound", err)
	}
	if inactive, err := fixture.repo.List(ctx, objectivesdomain.ListQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.inactiveA,
	}); err != nil || len(inactive) != 0 {
		t.Fatalf("inactive actor list = %#v, error=%v", inactive, err)
	}
	if guest, err := fixture.repo.List(ctx, objectivesdomain.ListQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.guestA,
	}); err != nil || len(guest) != 0 {
		t.Fatalf("guest actor list = %#v, error=%v", guest, err)
	}

	invalid := fixture.createCommand(uniqueObjectiveName("invalid-status"))
	invalid.Objective.Status = fixture.statusB
	if _, err := fixture.repo.Create(ctx, invalid); !errors.Is(err, objectivesdomain.ErrInvalidReference) {
		t.Fatalf("foreign status create error = %v, want ErrInvalidReference", err)
	}
	unauthorized := fixture.createCommand(uniqueObjectiveName("unauthorized"))
	unauthorized.Objective.CreatedBy = fixture.outsiderA
	if _, err := fixture.repo.Create(ctx, unauthorized); !errors.Is(err, objectivesdomain.ErrForbidden) {
		t.Fatalf("team outsider create error = %v, want ErrForbidden", err)
	}

	expected := created.Objective.UpdatedAt
	updated, err := fixture.repo.Update(ctx, objectivesdomain.UpdateCommand{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		ExpectedUpdatedAt: &expected, Comment: "reviewed",
		Patch: objectivesdomain.ObjectivePatch{
			Name:        objectivesdomain.SetField(uniqueObjectiveName("updated")),
			Description: objectivesdomain.ClearField[string](),
		},
	})
	if err != nil {
		t.Fatalf("compare-and-swap update: %v", err)
	}
	if !updated.UpdatedAt.After(expected) {
		t.Fatalf("updated_at = %v, want after %v", updated.UpdatedAt, expected)
	}
	if _, err := fixture.repo.Update(ctx, objectivesdomain.UpdateCommand{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		ExpectedUpdatedAt: &expected,
		Patch:             objectivesdomain.ObjectivePatch{Name: objectivesdomain.SetField(uniqueObjectiveName("stale"))},
	}); !errors.Is(err, objectivesdomain.ErrVersionConflict) {
		t.Fatalf("stale update error = %v, want ErrVersionConflict", err)
	}

	testObjectiveUpdateRollback(t, ctx, fixture)
	testConcurrentObjectiveSequences(t, ctx, fixture)
	testConcurrentObjectiveCAS(t, ctx, fixture)
	testObjectiveStrategyAuthorization(t, ctx, fixture, created.Objective.ID, foreign.Objective.ID)

	if err := fixture.repo.Delete(ctx, objectivesdomain.DeleteCommand{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorB,
	}); !errors.Is(err, objectivesdomain.ErrNotFound) {
		t.Fatalf("cross-tenant delete error = %v, want ErrNotFound", err)
	}
}

func testObjectiveUpdateRollback(t *testing.T, ctx context.Context, fixture objectiveIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueObjectiveName("rollback")))
	if err != nil {
		t.Fatalf("create rollback objective: %v", err)
	}
	mustObjectiveExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_objective_update_activity() RETURNS trigger AS $$
		BEGIN
			IF NEW.activity_type = 'update' THEN
				RAISE EXCEPTION 'forced objective activity failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	mustObjectiveExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_objective_update_activity
		BEFORE INSERT ON okr_activities
		FOR EACH ROW EXECUTE FUNCTION reject_objective_update_activity()
	`)
	newName := uniqueObjectiveName("must-not-persist")
	_, err = fixture.repo.Update(ctx, objectivesdomain.UpdateCommand{
		ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		Patch: objectivesdomain.ObjectivePatch{Name: objectivesdomain.SetField(newName)},
	})
	if err == nil {
		t.Fatal("update error = nil, want activity failure")
	}
	mustObjectiveExec(t, ctx, fixture.postgres.Pool, "DROP TRIGGER reject_objective_update_activity ON okr_activities")
	mustObjectiveExec(t, ctx, fixture.postgres.Pool, "DROP FUNCTION reject_objective_update_activity()")
	var persistedName string
	if err := fixture.postgres.Pool.QueryRow(ctx,
		"SELECT name FROM objectives WHERE objective_id = $1", created.Objective.ID,
	).Scan(&persistedName); err != nil {
		t.Fatalf("read rolled-back objective: %v", err)
	}
	if persistedName != created.Objective.Name {
		t.Fatalf("rolled-back objective name = %q, want %q", persistedName, created.Objective.Name)
	}
}

func testConcurrentObjectiveSequences(t *testing.T, ctx context.Context, fixture objectiveIntegrationFixture) {
	t.Helper()
	const writers = 6
	results := make(chan objectivesdomain.CreateResult, writers)
	errorsChannel := make(chan error, writers)
	var wait sync.WaitGroup
	for index := 0; index < writers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			command := fixture.createCommand(uniqueObjectiveName("concurrent"))
			command.KeyResults = nil
			result, err := fixture.repo.Create(ctx, command)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- result
		}()
	}
	wait.Wait()
	close(results)
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatalf("concurrent create: %v", err)
	}
	sequences := make(map[int]struct{}, writers)
	for result := range results {
		sequences[result.Objective.SequenceID] = struct{}{}
	}
	if len(sequences) != writers {
		t.Fatalf("unique objective sequences = %d, want %d: %v", len(sequences), writers, sequences)
	}
}

func testConcurrentObjectiveCAS(t *testing.T, ctx context.Context, fixture objectiveIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueObjectiveName("cas")))
	if err != nil {
		t.Fatalf("create CAS objective: %v", err)
	}
	expected := created.Objective.UpdatedAt
	errorsChannel := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := fixture.repo.Update(ctx, objectivesdomain.UpdateCommand{
				ObjectiveID: created.Objective.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
				ExpectedUpdatedAt: &expected,
				Patch:             objectivesdomain.ObjectivePatch{Name: objectivesdomain.SetField(uniqueObjectiveName("cas-winner"))},
			})
			errorsChannel <- err
		}()
	}
	wait.Wait()
	close(errorsChannel)
	successes, conflicts := 0, 0
	for err := range errorsChannel {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, objectivesdomain.ErrVersionConflict):
			conflicts++
		default:
			t.Fatalf("concurrent CAS error = %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS outcomes success/conflict = %d/%d, want 1/1", successes, conflicts)
	}
}

func testObjectiveStrategyAuthorization(
	t *testing.T,
	ctx context.Context,
	fixture objectiveIntegrationFixture,
	objectiveID, foreignObjectiveID uuid.UUID,
) {
	t.Helper()
	query := objectivesdomain.StrategyQuery{WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA}
	pillar, err := fixture.repo.CreateStrategicPillar(ctx, query, objectivesdomain.NewStrategicPillar{
		Name: uniqueObjectiveName("pillar"), OrderIndex: 1,
	})
	if err != nil {
		t.Fatalf("create strategic pillar: %v", err)
	}
	if err := fixture.repo.AlignObjective(ctx, query, objectiveID, &pillar.ID); err != nil {
		t.Fatalf("align objective: %v", err)
	}
	strategy, err := fixture.repo.GetStrategyMap(ctx, query)
	if err != nil || len(strategy.Pillars) != 1 || len(strategy.Pillars[0].ObjectiveIDs) != 1 {
		t.Fatalf("strategy map = %#v, error=%v", strategy, err)
	}
	if err := fixture.repo.AlignObjective(ctx, query, foreignObjectiveID, &pillar.ID); !errors.Is(err, objectivesdomain.ErrNotFound) {
		t.Fatalf("cross-tenant alignment error = %v, want ErrNotFound", err)
	}
	if _, err := fixture.repo.GetStrategyMap(ctx, objectivesdomain.StrategyQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.outsiderA,
	}); err != nil {
		// Workspace members may read the strategy shell, while objective
		// alignments remain filtered to their team memberships.
		t.Fatalf("workspace strategy read for team outsider: %v", err)
	}
	if _, err := fixture.repo.GetStrategyMap(ctx, objectivesdomain.StrategyQuery{
		WorkspaceID: fixture.workspaceA, ActorID: fixture.guestA,
	}); !errors.Is(err, objectivesdomain.ErrForbidden) {
		t.Fatalf("guest strategy read error = %v, want ErrForbidden", err)
	}
}

func assertObjectivePostgres18(t *testing.T, ctx context.Context, fixture objectiveIntegrationFixture) {
	t.Helper()
	var raw string
	if err := fixture.postgres.Pool.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 180000 || version >= 190000 {
		t.Fatalf("PostgreSQL version = %q, want 18.x", raw)
	}
}
