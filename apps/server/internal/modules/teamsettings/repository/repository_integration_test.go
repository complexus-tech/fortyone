//go:build integration

package teamsettingsrepository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestRepositoryEnforcesTenantMembershipAndTypedPatchContracts(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	assertPostgres18(t, ctx, postgres)
	fixture := newTeamSettingsFixture(t, ctx, postgres)
	repository := New(postgres.Pool)

	isMember, err := repository.IsActiveTeamMember(ctx, fixture.teamA, fixture.workspaceA, fixture.activeUserA)
	if err != nil || !isMember {
		t.Fatalf("active membership = %t, error %v", isMember, err)
	}
	for name, membership := range map[string]struct {
		teamID, workspaceID, actorID uuid.UUID
	}{
		"second tenant": {fixture.teamA, fixture.workspaceB, fixture.activeUserA},
		"other team":    {fixture.teamB, fixture.workspaceB, fixture.activeUserA},
		"inactive user": {fixture.teamA, fixture.workspaceA, fixture.inactiveUserA},
	} {
		isMember, err = repository.IsActiveTeamMember(ctx, membership.teamID, membership.workspaceID, membership.actorID)
		if err != nil || isMember {
			t.Fatalf("%s membership = %t, error %v", name, isMember, err)
		}
	}

	if _, err := repository.GetSprintSettings(ctx, fixture.teamA, fixture.workspaceB); !errors.Is(err, teamsettings.ErrTeamSettingsNotFound) {
		t.Fatalf("cross-tenant sprint read error = %v", err)
	}
	assertSettingsRowCount(t, ctx, postgres, "team_sprint_settings", fixture.teamA, fixture.workspaceB, 0)

	if _, err := repository.GetSprintSettings(ctx, fixture.teamA, fixture.workspaceA); err != nil {
		t.Fatalf("get sprint settings: %v", err)
	}
	updated, err := repository.UpdateSprintSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamSprintSettings{
			UpcomingSprintsCount:         teamsettings.PatchField[int]{Value: 0, Present: true},
			WorkingDays:                  teamsettings.PatchField[[]int]{Value: []int{1, 2, 3, 4, 5}, Present: true},
			MoveIncompleteStoriesEnabled: teamsettings.PatchField[bool]{Value: false, Present: true},
		},
		teamsettings.SystemAuditActor(),
	)
	if err != nil {
		t.Fatalf("typed sprint patch: %v", err)
	}
	if updated.UpcomingSprintsCount != 0 || updated.MoveIncompleteStoriesEnabled || len(updated.WorkingDays) != 5 {
		t.Fatalf("typed sprint patch result = %#v", updated)
	}

	if _, err := repository.GetStoryAutomationSettings(ctx, fixture.teamA, fixture.workspaceA); err != nil {
		t.Fatalf("get story automation settings: %v", err)
	}
	if _, err := repository.GetStoryAutomationSettings(ctx, fixture.teamA, fixture.workspaceB); !errors.Is(err, teamsettings.ErrTeamSettingsNotFound) {
		t.Fatalf("cross-tenant story automation read error = %v", err)
	}
	if _, err := repository.UpdateEstimationSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamEstimationSettings{
			Scheme: teamsettings.PatchField[string]{Value: "points", Present: true},
		},
		teamsettings.SystemAuditActor(),
	); err != nil {
		t.Fatalf("update estimation settings: %v", err)
	}
	if _, err := repository.GetEstimationSettings(ctx, fixture.teamA, fixture.workspaceB); !errors.Is(err, teamsettings.ErrTeamSettingsNotFound) {
		t.Fatalf("cross-tenant estimation read error = %v", err)
	}
	assertAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA, 2)
	if _, err := repository.UpdateEstimationSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamEstimationSettings{
			Scheme: teamsettings.PatchField[string]{Value: "hours", Present: true},
		},
		teamsettings.SystemAuditActor(),
	); !errors.Is(err, teamsettings.ErrInvalidEstimateScheme) {
		t.Fatalf("database constraint error = %v, want ErrInvalidEstimateScheme", err)
	}
	storedEstimation, err := repository.GetEstimationSettings(ctx, fixture.teamA, fixture.workspaceA)
	if err != nil {
		t.Fatalf("reload estimation settings: %v", err)
	}
	if storedEstimation.Scheme != "points" {
		t.Fatalf("estimation scheme after constraint rollback = %q, want points", storedEstimation.Scheme)
	}
	assertAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA, 2)
}

func TestScheduleConflictRollsBackSettingsSprintAndAudit(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newTeamSettingsFixture(t, ctx, postgres)
	repository := New(postgres.Pool)

	if _, err := repository.GetSprintSettings(ctx, fixture.teamA, fixture.workspaceA); err != nil {
		t.Fatalf("initialize sprint settings: %v", err)
	}
	if _, err := repository.UpdateSprintSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamSprintSettings{
			AutoCreateSprints: teamsettings.PatchField[bool]{Value: true, Present: true},
		},
		teamsettings.SystemAuditActor(),
	); err != nil {
		t.Fatalf("enable sprint automation: %v", err)
	}

	var scheduleDate time.Time
	if err := postgres.Pool.QueryRow(ctx, "SELECT CURRENT_DATE").Scan(&scheduleDate); err != nil {
		t.Fatalf("get database date: %v", err)
	}
	nextStart := nextSprintStartAfter(scheduleDate, "Monday")
	managedID := insertTeamSettingsSprint(
		t, ctx, postgres, fixture.teamA, fixture.workspaceA,
		"Managed", nextStart.AddDate(0, 0, 28), nextStart.AddDate(0, 0, 41), true,
	)
	customID := insertTeamSettingsSprint(
		t, ctx, postgres, fixture.teamA, fixture.workspaceA,
		"Custom", nextStart, nextStart.AddDate(0, 0, 2), false,
	)
	beforeStart, beforeEnd := readSprintDates(t, ctx, postgres, managedID)
	beforeAudits := readAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA)

	_, err := repository.UpdateSprintSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamSprintSettings{
			SprintDurationWeeks: teamsettings.PatchField[int]{Value: 1, Present: true},
		},
		teamsettings.SystemAuditActor(),
	)
	if !errors.Is(err, teamsettings.ErrSprintScheduleConflict) {
		t.Fatalf("schedule conflict error = %v", err)
	}

	settings, err := repository.GetSprintSettings(ctx, fixture.teamA, fixture.workspaceA)
	if err != nil {
		t.Fatalf("reload sprint settings: %v", err)
	}
	if settings.SprintDurationWeeks != 2 {
		t.Fatalf("duration after rollback = %d, want 2", settings.SprintDurationWeeks)
	}
	afterStart, afterEnd := readSprintDates(t, ctx, postgres, managedID)
	if !afterStart.Equal(beforeStart) || !afterEnd.Equal(beforeEnd) {
		t.Fatalf("managed sprint changed despite rollback: %v..%v -> %v..%v", beforeStart, beforeEnd, afterStart, afterEnd)
	}
	if afterAudits := readAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA); afterAudits != beforeAudits {
		t.Fatalf("audit count after rollback = %d, want %d", afterAudits, beforeAudits)
	}

	if _, err := postgres.Pool.Exec(ctx, "DELETE FROM sprints WHERE sprint_id = $1", customID); err != nil {
		t.Fatalf("remove conflicting custom sprint: %v", err)
	}
	updated, err := repository.UpdateSprintSettings(
		ctx,
		fixture.teamA,
		fixture.workspaceA,
		teamsettings.CoreUpdateTeamSprintSettings{
			SprintDurationWeeks: teamsettings.PatchField[int]{Value: 1, Present: true},
		},
		teamsettings.SystemAuditActor(),
	)
	if err != nil {
		t.Fatalf("reschedule managed sprint: %v", err)
	}
	if updated.SprintDurationWeeks != 1 {
		t.Fatalf("updated duration = %d, want 1", updated.SprintDurationWeeks)
	}
	afterStart, afterEnd = readSprintDates(t, ctx, postgres, managedID)
	if !afterStart.Equal(nextStart) || !afterEnd.Equal(nextStart.AddDate(0, 0, 6)) {
		t.Fatalf("managed sprint dates = %v..%v, want %v..%v", afterStart, afterEnd, nextStart, nextStart.AddDate(0, 0, 6))
	}
	assertAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA, beforeAudits+2)
}

func TestConcurrentTypedUpdatesReturnRetryableDomainConflict(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newTeamSettingsFixture(t, ctx, postgres)
	baseQueries := teamsettingssql.New(postgres.Pool)
	repository := newWithQueries(baseQueries)
	repository.runTransaction = synchronizedSerializableRunner(t, postgres, baseQueries, 2)

	if _, err := New(postgres.Pool).GetEstimationSettings(ctx, fixture.teamA, fixture.workspaceA); err != nil {
		t.Fatalf("initialize estimation settings: %v", err)
	}

	updates := []string{"points", "tshirt"}
	errorsByUpdate := make(chan error, len(updates))
	var waitGroup sync.WaitGroup
	for _, scheme := range updates {
		scheme := scheme
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := repository.UpdateEstimationSettings(
				ctx,
				fixture.teamA,
				fixture.workspaceA,
				teamsettings.CoreUpdateTeamEstimationSettings{
					Scheme: teamsettings.PatchField[string]{Value: scheme, Present: true},
				},
				teamsettings.SystemAuditActor(),
			)
			errorsByUpdate <- err
		}()
	}
	waitGroup.Wait()
	close(errorsByUpdate)

	var successes, conflicts int
	for err := range errorsByUpdate {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, teamsettings.ErrConcurrentUpdate):
			conflicts++
		default:
			t.Fatalf("concurrent update returned unexpected error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent outcomes = %d success/%d conflict, want 1/1", successes, conflicts)
	}
	finalSettings, err := New(postgres.Pool).GetEstimationSettings(ctx, fixture.teamA, fixture.workspaceA)
	if err != nil {
		t.Fatalf("read final estimation settings: %v", err)
	}
	if finalSettings.Scheme != "points" && finalSettings.Scheme != "tshirt" {
		t.Fatalf("final estimation scheme = %q", finalSettings.Scheme)
	}
	assertAuditEventCount(t, ctx, postgres, fixture.workspaceA, fixture.teamA, 1)
}

func synchronizedSerializableRunner(
	t *testing.T,
	postgres *testkit.Postgres,
	baseQueries *teamsettingssql.Queries,
	participants int32,
) func(context.Context, func(teamsettingssql.Querier) error) error {
	t.Helper()
	ready := make(chan struct{})
	var arrived atomic.Int32
	return func(ctx context.Context, operation func(teamsettingssql.Querier) error) error {
		tx, err := postgres.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: pgx.ReadWrite})
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var snapshot string
		if err := tx.QueryRow(ctx, "SELECT txid_current_snapshot()").Scan(&snapshot); err != nil {
			return err
		}
		if arrived.Add(1) == participants {
			close(ready)
		}
		select {
		case <-ready:
		case <-ctx.Done():
			return ctx.Err()
		}
		if err := operation(baseQueries.WithTx(tx)); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
}
