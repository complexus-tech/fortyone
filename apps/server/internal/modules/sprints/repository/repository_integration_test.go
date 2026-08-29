//go:build integration

package sprintsrepository

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestSprintRepositoryTenantSecurityTransactionsAndAnalytics(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	fixture := newSprintIntegrationFixture(t, ctx)
	assertSprintPostgres18(t, ctx, fixture)

	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueSprintName("primary")))
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	if created.StartDate.Hour() != 0 || created.StartDate.Location() != time.UTC || created.ObjectiveID == nil {
		t.Fatalf("normalized created sprint = %#v", created)
	}
	if got := sprintRowCount(t, ctx, fixture.postgres.Pool, `
		SELECT COUNT(*) FROM audit_events
		WHERE entity_type = 'sprint' AND entity_id = $1 AND event_type = 'sprint.created'
	`, created.ID); got != 1 {
		t.Fatalf("create audit count = %d, want 1", got)
	}

	foreign, err := fixture.repo.Create(ctx, fixture.foreignCreateCommand(uniqueSprintName("foreign")))
	if err != nil {
		t.Fatalf("create foreign sprint: %v", err)
	}
	assertSprintCreateAuthorization(t, ctx, fixture)
	assertSprintReadAuthorization(t, ctx, fixture, created.ID, foreign.ID)

	for index, category := range []string{"completed", "started", "unstarted", "paused", "cancelled"} {
		assignee := fixture.assigneeA
		if index == 0 {
			assignee = fixture.actorA
		}
		fixture.insertStory(t, ctx, created.ID, assignee, category, index+1, false, false)
	}
	fixture.insertStory(t, ctx, created.ID, fixture.assigneeA, "completed", 101, true, false)
	fixture.insertStory(t, ctx, created.ID, fixture.assigneeA, "started", 102, false, true)

	visible, err := fixture.repo.List(ctx, sprintdomain.ListQuery{
		WorkspaceID: fixture.workspaceA,
		ActorID:     fixture.actorA,
		Filter: sprintdomain.ListFilter{
			Search: "PRIMARY", Limit: 10,
		},
	})
	if err != nil || len(visible) != 1 || visible[0].ID != created.ID {
		t.Fatalf("visible sprint list = %#v, error=%v", visible, err)
	}
	if visible[0].TotalStories != 5 || visible[0].CompletedStories != 1 ||
		visible[0].StartedStories != 1 || visible[0].UnstartedStories != 1 ||
		visible[0].CancelledStories != 1 {
		t.Fatalf("active story counts = %#v", visible[0])
	}

	running, err := fixture.repo.Running(
		ctx,
		fixture.workspaceA,
		fixture.actorA,
		time.Date(2026, time.August, 28, 20, 0, 0, 0, time.UTC),
	)
	if err != nil || len(running) != 1 || running[0].ID != created.ID {
		t.Fatalf("running sprints = %#v, error=%v", running, err)
	}

	now := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)
	analytics, err := fixture.repo.GetAnalytics(ctx, created.ID, fixture.workspaceA, fixture.actorA, now)
	if err != nil {
		t.Fatalf("get sprint analytics: %v", err)
	}
	if analytics.StoryBreakdown != (sprintdomain.StoryBreakdown{
		Total: 5, Completed: 1, InProgress: 1, Todo: 1, Blocked: 1, Cancelled: 1,
	}) {
		t.Fatalf("story breakdown = %#v", analytics.StoryBreakdown)
	}
	if analytics.Overview.CompletionPercentage != 20 || len(analytics.Burndown) != 12 {
		t.Fatalf("analytics overview/burndown = %#v / %d points", analytics.Overview, len(analytics.Burndown))
	}
	assertSprintAllocation(t, analytics.TeamAllocation, fixture.actorA, fixture.assigneeA)

	updated := testSprintUpdateAndCAS(t, ctx, fixture, created)
	testConcurrentSprintCAS(t, ctx, fixture)
	testSprintAuditRollback(t, ctx, fixture)
	testSprintRevocation(t, ctx, fixture)
	assertSprintQueryPlans(t, ctx, fixture, created.ID)

	if err := fixture.repo.Delete(ctx, sprintdomain.DeleteCommand{
		SprintID: updated.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorB,
	}); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("cross-tenant delete error = %v, want ErrNotFound", err)
	}
	if err := fixture.repo.Delete(ctx, sprintdomain.DeleteCommand{
		SprintID: updated.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
	}); err != nil {
		t.Fatalf("delete sprint: %v", err)
	}
	if got := sprintRowCount(t, ctx, fixture.postgres.Pool, `
		SELECT COUNT(*) FROM audit_events
		WHERE entity_type = 'sprint' AND entity_id = $1 AND event_type = 'sprint.deleted'
	`, updated.ID); got != 1 {
		t.Fatalf("delete audit count = %d, want 1", got)
	}
}

func assertSprintCreateAuthorization(t *testing.T, ctx context.Context, fixture sprintIntegrationFixture) {
	t.Helper()
	for name, actorID := range map[string]uuid.UUID{
		"team outsider":   fixture.outsiderA,
		"inactive member": fixture.inactiveA,
		"guest member":    fixture.guestA,
		"foreign actor":   fixture.actorB,
	} {
		command := fixture.createCommand(uniqueSprintName(name))
		command.ActorID = actorID
		if _, err := fixture.repo.Create(ctx, command); !errors.Is(err, sprintdomain.ErrForbidden) {
			t.Fatalf("%s create error = %v, want ErrForbidden", name, err)
		}
	}
	for name, objectiveID := range map[string]uuid.UUID{
		"other team":      fixture.objectiveOtherA,
		"other workspace": fixture.objectiveB,
	} {
		command := fixture.createCommand(uniqueSprintName("invalid-objective-" + name))
		command.Sprint.ObjectiveID = &objectiveID
		if _, err := fixture.repo.Create(ctx, command); !errors.Is(err, sprintdomain.ErrInvalidReference) {
			t.Fatalf("%s objective create error = %v, want ErrInvalidReference", name, err)
		}
	}
}

func assertSprintReadAuthorization(
	t *testing.T,
	ctx context.Context,
	fixture sprintIntegrationFixture,
	sprintID, foreignSprintID uuid.UUID,
) {
	t.Helper()
	foreignVisible, err := fixture.repo.List(ctx, sprintdomain.ListQuery{
		WorkspaceID: fixture.workspaceA,
		ActorID:     fixture.actorB,
	})
	if err != nil || len(foreignVisible) != 0 {
		t.Fatalf("cross-tenant list = %#v, error=%v", foreignVisible, err)
	}
	for name, actorID := range map[string]uuid.UUID{
		"team outsider":   fixture.outsiderA,
		"inactive member": fixture.inactiveA,
		"guest member":    fixture.guestA,
	} {
		items, err := fixture.repo.List(ctx, sprintdomain.ListQuery{
			WorkspaceID: fixture.workspaceA,
			ActorID:     actorID,
		})
		if err != nil || len(items) != 0 {
			t.Fatalf("%s list = %#v, error=%v", name, items, err)
		}
		if _, err := fixture.repo.GetByID(ctx, sprintID, fixture.workspaceA, actorID); !errors.Is(err, sprintdomain.ErrNotFound) {
			t.Fatalf("%s direct read error = %v, want ErrNotFound", name, err)
		}
	}
	if _, err := fixture.repo.GetByID(ctx, foreignSprintID, fixture.workspaceA, fixture.actorA); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("foreign sprint direct read error = %v, want ErrNotFound", err)
	}
	if _, err := fixture.repo.GetAnalytics(
		ctx,
		sprintID,
		fixture.workspaceA,
		fixture.outsiderA,
		time.Now(),
	); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("team outsider analytics error = %v, want ErrNotFound", err)
	}
}

func assertSprintAllocation(
	t *testing.T,
	allocation []sprintdomain.TeamMemberAllocation,
	actorID, assigneeID uuid.UUID,
) {
	t.Helper()
	if len(allocation) != 3 {
		// The third active, non-guest member is the revocation test actor and has
		// no assigned stories. Inactive and guest membership rows must be absent.
		t.Fatalf("team allocation count = %d, want 3: %#v", len(allocation), allocation)
	}
	counts := make(map[uuid.UUID]sprintdomain.TeamMemberAllocation, len(allocation))
	for _, member := range allocation {
		counts[member.MemberID] = member
	}
	if counts[actorID].Assigned != 1 || counts[actorID].Completed != 1 {
		t.Fatalf("actor allocation = %#v", counts[actorID])
	}
	if counts[assigneeID].Assigned != 4 || counts[assigneeID].Completed != 0 {
		t.Fatalf("assignee allocation = %#v", counts[assigneeID])
	}
}

func testSprintUpdateAndCAS(
	t *testing.T,
	ctx context.Context,
	fixture sprintIntegrationFixture,
	created sprintdomain.Sprint,
) sprintdomain.Sprint {
	t.Helper()
	expected := created.UpdatedAt
	newStart := time.Date(2026, time.August, 25, 22, 0, 0, 0, time.UTC)
	newEnd := time.Date(2026, time.September, 7, 22, 0, 0, 0, time.UTC)
	updated, err := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		ExpectedUpdatedAt: &expected,
		Patch: sprintdomain.Patch{
			Name:        platformpatch.Set(uniqueSprintName("updated")),
			Goal:        platformpatch.Clear[string](),
			ObjectiveID: platformpatch.Set(fixture.objectiveA2),
			StartDate:   platformpatch.Set(newStart),
			EndDate:     platformpatch.Set(newEnd),
		},
	})
	if err != nil {
		t.Fatalf("compare-and-swap update: %v", err)
	}
	if updated.Goal != nil || updated.ObjectiveID == nil || *updated.ObjectiveID != fixture.objectiveA2 ||
		updated.StartDate.Hour() != 0 || updated.ScheduleManagedByAutomation {
		t.Fatalf("updated sprint = %#v", updated)
	}
	if _, err := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		ExpectedUpdatedAt: &expected,
		Patch:             sprintdomain.Patch{Name: platformpatch.Set(uniqueSprintName("stale"))},
	}); !errors.Is(err, sprintdomain.ErrVersionConflict) {
		t.Fatalf("stale update error = %v, want ErrVersionConflict", err)
	}
	if _, err := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		Patch: sprintdomain.Patch{ObjectiveID: platformpatch.Set(fixture.objectiveOtherA)},
	}); !errors.Is(err, sprintdomain.ErrInvalidReference) {
		t.Fatalf("cross-team objective update error = %v, want ErrInvalidReference", err)
	}
	tooLate := updated.EndDate.AddDate(0, 0, 1)
	if _, err := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		Patch: sprintdomain.Patch{StartDate: platformpatch.Set(tooLate)},
	}); !errors.Is(err, sprintdomain.ErrInvalid) {
		t.Fatalf("crossed dates update error = %v, want ErrInvalid", err)
	}
	return updated
}

func testConcurrentSprintCAS(t *testing.T, ctx context.Context, fixture sprintIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueSprintName("concurrent-cas")))
	if err != nil {
		t.Fatalf("create concurrent CAS sprint: %v", err)
	}
	expected := created.UpdatedAt
	errorsChannel := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, updateErr := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
				SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
				ExpectedUpdatedAt: &expected,
				Patch:             sprintdomain.Patch{Name: platformpatch.Set(uniqueSprintName("cas-winner"))},
			})
			errorsChannel <- updateErr
		}()
	}
	wait.Wait()
	close(errorsChannel)
	successes, conflicts := 0, 0
	for updateErr := range errorsChannel {
		switch {
		case updateErr == nil:
			successes++
		case errors.Is(updateErr, sprintdomain.ErrVersionConflict):
			conflicts++
		default:
			t.Fatalf("concurrent CAS error = %v", updateErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS outcomes success/conflict = %d/%d, want 1/1", successes, conflicts)
	}
}

func testSprintAuditRollback(t *testing.T, ctx context.Context, fixture sprintIntegrationFixture) {
	t.Helper()
	created, err := fixture.repo.Create(ctx, fixture.createCommand(uniqueSprintName("rollback")))
	if err != nil {
		t.Fatalf("create rollback sprint: %v", err)
	}
	mustSprintExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_sprint_update_audit() RETURNS trigger AS $$
		BEGIN
			IF NEW.entity_type = 'sprint' AND NEW.event_type = 'sprint.updated' THEN
				RAISE EXCEPTION 'forced sprint audit failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`)
	mustSprintExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_sprint_update_audit
		BEFORE INSERT ON audit_events
		FOR EACH ROW EXECUTE FUNCTION reject_sprint_update_audit()
	`)
	newName := uniqueSprintName("must-not-persist")
	_, err = fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA,
		Patch: sprintdomain.Patch{Name: platformpatch.Set(newName)},
	})
	if err == nil {
		t.Fatal("update error = nil, want audit failure")
	}
	mustSprintExec(t, ctx, fixture.postgres.Pool, "DROP TRIGGER reject_sprint_update_audit ON audit_events")
	mustSprintExec(t, ctx, fixture.postgres.Pool, "DROP FUNCTION reject_sprint_update_audit()")
	var persistedName string
	if err := fixture.postgres.Pool.QueryRow(ctx, `
		SELECT name FROM sprints WHERE sprint_id = $1
	`, created.ID).Scan(&persistedName); err != nil {
		t.Fatalf("read rolled-back sprint: %v", err)
	}
	if persistedName != created.Name {
		t.Fatalf("rolled-back sprint name = %q, want %q", persistedName, created.Name)
	}
}

func testSprintRevocation(t *testing.T, ctx context.Context, fixture sprintIntegrationFixture) {
	t.Helper()
	command := fixture.createCommand(uniqueSprintName("revoked"))
	command.ActorID = fixture.revocableA
	created, err := fixture.repo.Create(ctx, command)
	if err != nil {
		t.Fatalf("create revocation sprint: %v", err)
	}
	mustSprintExec(t, ctx, fixture.postgres.Pool, `
		DELETE FROM team_members WHERE team_id = $1 AND user_id = $2
	`, fixture.teamA, fixture.revocableA)
	items, err := fixture.repo.List(ctx, sprintdomain.ListQuery{
		WorkspaceID: fixture.workspaceA,
		ActorID:     fixture.revocableA,
	})
	if err != nil || len(items) != 0 {
		t.Fatalf("revoked list = %#v, error=%v", items, err)
	}
	if _, err := fixture.repo.GetByID(
		ctx,
		created.ID,
		fixture.workspaceA,
		fixture.revocableA,
	); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("revoked direct read error = %v, want ErrNotFound", err)
	}
	if _, err := fixture.repo.GetAnalytics(
		ctx,
		created.ID,
		fixture.workspaceA,
		fixture.revocableA,
		time.Now(),
	); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("revoked analytics error = %v, want ErrNotFound", err)
	}
	if _, err := fixture.repo.Update(ctx, sprintdomain.UpdateCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.revocableA,
		Patch: sprintdomain.Patch{Name: platformpatch.Set(uniqueSprintName("revoked-update"))},
	}); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("revoked update error = %v, want ErrNotFound", err)
	}
	if err := fixture.repo.Delete(ctx, sprintdomain.DeleteCommand{
		SprintID: created.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.revocableA,
	}); !errors.Is(err, sprintdomain.ErrNotFound) {
		t.Fatalf("revoked delete error = %v, want ErrNotFound", err)
	}
}

func assertSprintPostgres18(t *testing.T, ctx context.Context, fixture sprintIntegrationFixture) {
	t.Helper()
	var raw string
	if err := fixture.postgres.Pool.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	versionNumber, err := strconv.Atoi(raw)
	if err != nil || versionNumber < 180000 || versionNumber >= 190000 {
		t.Fatalf("PostgreSQL version = %q, want 18.x", raw)
	}
}

func assertSprintQueryPlans(
	t *testing.T,
	ctx context.Context,
	fixture sprintIntegrationFixture,
	sprintID uuid.UUID,
) {
	t.Helper()
	connection, err := fixture.postgres.Pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire query-plan connection: %v", err)
	}
	defer connection.Release()
	tx, err := connection.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatalf("begin query-plan transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		t.Fatalf("disable sequential scans for plan assertion: %v", err)
	}
	assertPlanUsesIndex(t, ctx, tx, "idx_sprints_workspace_end_id", `
		EXPLAIN (COSTS OFF)
		SELECT sprint_id
		FROM sprints
		WHERE workspace_id = $1
		ORDER BY end_date DESC, sprint_id DESC
		LIMIT 50
	`, fixture.workspaceA)
	assertPlanUsesIndex(t, ctx, tx, "idx_stories_workspace_sprint_status_active", `
		EXPLAIN (COSTS OFF)
		SELECT status_id, COUNT(*)
		FROM stories
		WHERE workspace_id = $1
		  AND sprint_id = $2
		  AND deleted_at IS NULL
		  AND archived_at IS NULL
		GROUP BY status_id
	`, fixture.workspaceA, sprintID)
}

func assertPlanUsesIndex(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	indexName, query string,
	arguments ...any,
) {
	t.Helper()
	rows, err := tx.Query(ctx, query, arguments...)
	if err != nil {
		t.Fatalf("explain query for %s: %v", indexName, err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan query plan for %s: %v", indexName, err)
		}
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read query plan for %s: %v", indexName, err)
	}
	if !strings.Contains(plan.String(), indexName) {
		t.Fatalf("query plan did not use %s:\n%s", indexName, plan.String())
	}
}
