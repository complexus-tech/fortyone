//go:build integration

package keyresultsrepository

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/google/uuid"
)

func TestKeyResultRepositoryTenantTransactionsAndConcurrency(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	fixture := newKeyResultIntegrationFixture(t, ctx)
	assertKeyResultPostgres18(t, ctx, fixture)

	command := fixture.createCommand(uniqueKeyResultName("primary"))
	secondary := command.KeyResults[0]
	secondary.Name = uniqueKeyResultName("secondary")
	secondary.Lead = nil
	secondary.Contributors = nil
	command.KeyResults = append(command.KeyResults, secondary)
	created, err := fixture.repo.CreateBatch(ctx, command)
	if err != nil {
		t.Fatalf("create key-result aggregate: %v", err)
	}
	if len(created) != 2 || created[0].SequenceID != 1 || created[1].SequenceID != 2 {
		t.Fatalf("created key-result sequences = %#v", created)
	}
	if got := keyResultRowCount(t, ctx, fixture.postgres.Pool,
		"SELECT COUNT(*) FROM key_result_contributors WHERE key_result_id = $1", created[0].ID,
	); got != 1 {
		t.Fatalf("deduplicated contributors = %d, want 1", got)
	}
	if got := keyResultRowCount(t, ctx, fixture.postgres.Pool,
		"SELECT COUNT(*) FROM okr_activities WHERE key_result_id = ANY($1)", []uuid.UUID{created[0].ID, created[1].ID},
	); got != 2 {
		t.Fatalf("atomic create activities = %d, want 2", got)
	}

	testKeyResultReadScopeAndPagination(t, ctx, fixture, created)
	testKeyResultInvalidAssigneeRollsBackBeforeSequence(t, ctx, fixture)
	testKeyResultCreateActivityFailureRollsBackAggregate(t, ctx, fixture)
	testKeyResultUpdateAndActivityRollback(t, ctx, fixture, created[0])
	testConcurrentKeyResultSequences(t, ctx, fixture)
	testConcurrentKeyResultCAS(t, ctx, fixture)
	testKeyResultDeleteActivityRollback(t, ctx, fixture)

	foreign, err := fixture.repo.CreateBatch(ctx, fixture.foreignCreateCommand(uniqueKeyResultName("foreign")))
	if err != nil || len(foreign) != 1 {
		t.Fatalf("create foreign key result = %#v, %v", foreign, err)
	}
	if err := fixture.repo.Delete(ctx, keyresultsdomain.DeleteCommand{
		Access:      keyresultsdomain.AccessScope{WorkspaceID: fixture.workspaceB, ActorID: fixture.actorB, AllTeams: true},
		KeyResultID: created[1].ID,
	}); !errors.Is(err, keyresultsdomain.ErrNotFound) {
		t.Fatalf("cross-tenant delete error = %v, want ErrNotFound", err)
	}
}

func testKeyResultReadScopeAndPagination(
	t *testing.T,
	ctx context.Context,
	fixture keyResultIntegrationFixture,
	created []keyresultsdomain.KeyResult,
) {
	t.Helper()
	got, err := fixture.repo.Get(ctx, keyresultsdomain.GetQuery{Access: fixture.accessA(), KeyResultID: created[0].ID})
	if err != nil || got.ID != created[0].ID || len(got.Contributors) != 1 {
		t.Fatalf("authorized get = %#v, %v", got, err)
	}

	deniedScopes := []keyresultsdomain.AccessScope{
		{WorkspaceID: fixture.workspaceB, ActorID: fixture.actorB, AllTeams: true},
		{WorkspaceID: fixture.workspaceA, ActorID: fixture.actorB, AllTeams: true},
		{WorkspaceID: fixture.workspaceA, ActorID: fixture.outsiderA, AllTeams: true},
		{WorkspaceID: fixture.workspaceA, ActorID: fixture.inactiveA, AllTeams: true},
		{WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA, TeamIDs: []uuid.UUID{fixture.teamB}},
	}
	for _, access := range deniedScopes {
		if _, err := fixture.repo.Get(ctx, keyresultsdomain.GetQuery{Access: access, KeyResultID: created[0].ID}); !errors.Is(err, keyresultsdomain.ErrNotFound) {
			t.Fatalf("denied access %#v error = %v, want ErrNotFound", access, err)
		}
	}

	query := keyresultsdomain.PaginatedListQuery{
		Access: fixture.accessA(),
		Filters: keyresultsdomain.Filters{
			MeasurementTypes: []string{"percentage"}, Page: 1, PageSize: 1,
			OrderBy: "name", OrderDirection: "asc",
		},
	}
	first, err := fixture.repo.ListPaginated(ctx, query)
	if err != nil {
		t.Fatalf("first stable page: %v", err)
	}
	secondRead, err := fixture.repo.ListPaginated(ctx, query)
	if err != nil {
		t.Fatalf("repeat stable page: %v", err)
	}
	if first.TotalCount != 2 || len(first.KeyResults) != 1 || !first.HasMore ||
		len(secondRead.KeyResults) != 1 || secondRead.KeyResults[0].ID != first.KeyResults[0].ID ||
		!strings.HasPrefix(first.KeyResults[0].Name, "primary-") {
		t.Fatalf("stable first page = %#v; repeat = %#v", first, secondRead)
	}
	query.Filters.Page = 2
	second, err := fixture.repo.ListPaginated(ctx, query)
	if err != nil || len(second.KeyResults) != 1 || second.HasMore || !strings.HasPrefix(second.KeyResults[0].Name, "secondary-") {
		t.Fatalf("stable second page = %#v, %v", second, err)
	}
}

func testKeyResultInvalidAssigneeRollsBackBeforeSequence(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
	t.Helper()
	before := keyResultSequence(t, ctx, fixture)
	invalid := fixture.createCommand(uniqueKeyResultName("invalid-assignee"))
	invalid.KeyResults[0].Lead = &fixture.outsiderA
	if _, err := fixture.repo.CreateBatch(ctx, invalid); !errors.Is(err, keyresultsdomain.ErrInvalidReference) {
		t.Fatalf("invalid assignee error = %v, want ErrInvalidReference", err)
	}
	if after := keyResultSequence(t, ctx, fixture); after != before {
		t.Fatalf("invalid assignee consumed sequence: before=%d after=%d", before, after)
	}
}

func testKeyResultCreateActivityFailureRollsBackAggregate(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
	t.Helper()
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_key_result_create_activity() RETURNS trigger AS $$
		BEGIN
			IF NEW.activity_type = 'create' AND NEW.update_type = 'key_result' THEN
				RAISE EXCEPTION 'forced key result create activity failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_key_result_create_activity
		BEFORE INSERT ON okr_activities
		FOR EACH ROW EXECUTE FUNCTION reject_key_result_create_activity()
	`)
	beforeRows := keyResultRowCount(t, ctx, fixture.postgres.Pool, "SELECT COUNT(*) FROM key_results WHERE team_id = $1", fixture.teamA)
	beforeSequence := keyResultSequence(t, ctx, fixture)
	if _, err := fixture.repo.CreateBatch(ctx, fixture.createCommand(uniqueKeyResultName("rollback-create"))); err == nil {
		t.Fatal("create error = nil, want forced activity failure")
	}
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP TRIGGER reject_key_result_create_activity ON okr_activities")
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP FUNCTION reject_key_result_create_activity()")
	if afterRows := keyResultRowCount(t, ctx, fixture.postgres.Pool, "SELECT COUNT(*) FROM key_results WHERE team_id = $1", fixture.teamA); afterRows != beforeRows {
		t.Fatalf("failed create persisted rows: before=%d after=%d", beforeRows, afterRows)
	}
	if afterSequence := keyResultSequence(t, ctx, fixture); afterSequence != beforeSequence {
		t.Fatalf("failed create consumed sequence: before=%d after=%d", beforeSequence, afterSequence)
	}
}

func testKeyResultUpdateAndActivityRollback(
	t *testing.T,
	ctx context.Context,
	fixture keyResultIntegrationFixture,
	keyResult keyresultsdomain.KeyResult,
) {
	t.Helper()
	expected := keyResult.UpdatedAt
	updated, err := fixture.repo.Update(ctx, keyresultsdomain.UpdateCommand{
		Access: fixture.accessA(), KeyResultID: keyResult.ID, ExpectedUpdatedAt: &expected, Comment: "reviewed",
		Patch: keyresultsdomain.Patch{
			CurrentValue: keyresultsdomain.SetField(50.0),
			Contributors: keyresultsdomain.SetField([]uuid.UUID{}),
		},
	})
	if err != nil {
		t.Fatalf("typed key-result update: %v", err)
	}
	if updated.After.CurrentValue != 50 || len(updated.After.Contributors) != 0 || !updated.After.UpdatedAt.After(expected) ||
		strings.Join(updated.ChangedFields, ",") != "current_value,contributors" {
		t.Fatalf("typed update result = %#v", updated)
	}
	if got := keyResultRowCount(t, ctx, fixture.postgres.Pool,
		"SELECT COUNT(*) FROM okr_activities WHERE key_result_id = $1", keyResult.ID,
	); got != 3 {
		t.Fatalf("create plus update activities = %d, want 3", got)
	}

	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_key_result_update_activity() RETURNS trigger AS $$
		BEGIN
			IF NEW.activity_type = 'update' AND NEW.update_type = 'key_result' THEN
				RAISE EXCEPTION 'forced key result update activity failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_key_result_update_activity
		BEFORE INSERT ON okr_activities
		FOR EACH ROW EXECUTE FUNCTION reject_key_result_update_activity()
	`)
	failedName := uniqueKeyResultName("must-not-persist")
	if _, err := fixture.repo.Update(ctx, keyresultsdomain.UpdateCommand{
		Access: fixture.accessA(), KeyResultID: keyResult.ID,
		Patch: keyresultsdomain.Patch{Name: keyresultsdomain.SetField(failedName)},
	}); err == nil {
		t.Fatal("update error = nil, want activity failure")
	}
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP TRIGGER reject_key_result_update_activity ON okr_activities")
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP FUNCTION reject_key_result_update_activity()")
	persisted, err := fixture.repo.Get(ctx, keyresultsdomain.GetQuery{Access: fixture.accessA(), KeyResultID: keyResult.ID})
	if err != nil || persisted.Name == failedName || persisted.Name != keyResult.Name {
		t.Fatalf("rolled-back key result = %#v, %v", persisted, err)
	}
}

func testConcurrentKeyResultSequences(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
	t.Helper()
	const writers = 6
	sequences := make(chan int, writers)
	errorsChannel := make(chan error, writers)
	var wait sync.WaitGroup
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			created, err := fixture.repo.CreateBatch(ctx, fixture.createCommand(uniqueKeyResultName("concurrent")))
			if err != nil {
				errorsChannel <- err
				return
			}
			sequences <- created[0].SequenceID
		}()
	}
	wait.Wait()
	close(sequences)
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatalf("concurrent create: %v", err)
	}
	unique := make(map[int]struct{}, writers)
	for sequence := range sequences {
		unique[sequence] = struct{}{}
	}
	if len(unique) != writers {
		t.Fatalf("unique concurrent sequences = %d, want %d: %v", len(unique), writers, unique)
	}
}

func testConcurrentKeyResultCAS(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.CreateBatch(ctx, fixture.createCommand(uniqueKeyResultName("cas")))
	if err != nil {
		t.Fatalf("create CAS key result: %v", err)
	}
	expected := created[0].UpdatedAt
	errorsChannel := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := fixture.repo.Update(ctx, keyresultsdomain.UpdateCommand{
				Access: fixture.accessA(), KeyResultID: created[0].ID, ExpectedUpdatedAt: &expected,
				Patch: keyresultsdomain.Patch{Name: keyresultsdomain.SetField(uniqueKeyResultName("cas-winner"))},
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
		case errors.Is(err, keyresultsdomain.ErrVersionConflict):
			conflicts++
		default:
			t.Fatalf("concurrent CAS error = %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS outcomes success/conflict = %d/%d, want 1/1", successes, conflicts)
	}
}

func testKeyResultDeleteActivityRollback(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.CreateBatch(ctx, fixture.createCommand(uniqueKeyResultName("delete-rollback")))
	if err != nil {
		t.Fatalf("create delete rollback key result: %v", err)
	}
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_key_result_delete_activity() RETURNS trigger AS $$
		BEGIN
			IF NEW.activity_type = 'delete' AND NEW.update_type = 'key_result' THEN
				RAISE EXCEPTION 'forced key result delete activity failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_key_result_delete_activity
		BEFORE INSERT ON okr_activities
		FOR EACH ROW EXECUTE FUNCTION reject_key_result_delete_activity()
	`)
	if err := fixture.repo.Delete(ctx, keyresultsdomain.DeleteCommand{
		Access: fixture.accessA(), KeyResultID: created[0].ID,
	}); err == nil {
		t.Fatal("delete error = nil, want activity failure")
	}
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP TRIGGER reject_key_result_delete_activity ON okr_activities")
	mustKeyResultExec(t, ctx, fixture.postgres.Pool, "DROP FUNCTION reject_key_result_delete_activity()")
	if _, err := fixture.repo.Get(ctx, keyresultsdomain.GetQuery{Access: fixture.accessA(), KeyResultID: created[0].ID}); err != nil {
		t.Fatalf("failed delete did not roll back aggregate: %v", err)
	}
	if err := fixture.repo.Delete(ctx, keyresultsdomain.DeleteCommand{Access: fixture.accessA(), KeyResultID: created[0].ID}); err != nil {
		t.Fatalf("delete after removing trigger: %v", err)
	}
	if _, err := fixture.repo.Get(ctx, keyresultsdomain.GetQuery{Access: fixture.accessA(), KeyResultID: created[0].ID}); !errors.Is(err, keyresultsdomain.ErrNotFound) {
		t.Fatalf("deleted key result get error = %v, want ErrNotFound", err)
	}
	if got := keyResultRowCount(t, ctx, fixture.postgres.Pool, `
		SELECT COUNT(*) FROM okr_activities
		WHERE objective_id = $1 AND activity_type = 'delete' AND update_type = 'key_result' AND current_value = $2
	`, fixture.objectiveA, created[0].Name); got != 1 {
		t.Fatalf("delete activity count = %d, want 1", got)
	}
}

func keyResultSequence(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) int {
	t.Helper()
	var sequence int
	if err := fixture.postgres.Pool.QueryRow(ctx, `
		SELECT current_sequence
		FROM team_key_result_sequences
		WHERE workspace_id = $1 AND team_id = $2
	`, fixture.workspaceA, fixture.teamA).Scan(&sequence); err != nil {
		t.Fatalf("read key-result sequence: %v", err)
	}
	return sequence
}

func assertKeyResultPostgres18(t *testing.T, ctx context.Context, fixture keyResultIntegrationFixture) {
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
