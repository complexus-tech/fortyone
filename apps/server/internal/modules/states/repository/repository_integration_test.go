//go:build integration

package statesrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStatesEnforceTenantMembershipAndConcurrentInvariants(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	workspaceA := insertStateWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertStateWorkspace(t, ctx, postgres.Pool, "b")
	actorA := insertStateUser(t, ctx, postgres.Pool, "actor-a", true)
	inactiveA := insertStateUser(t, ctx, postgres.Pool, "inactive-a", false)
	insertStateWorkspaceMember(t, ctx, postgres.Pool, workspaceA, actorA)
	insertStateWorkspaceMember(t, ctx, postgres.Pool, workspaceA, inactiveA)
	teamA := insertStateTeam(t, ctx, postgres.Pool, workspaceA, "A")
	teamB := insertStateTeam(t, ctx, postgres.Pool, workspaceB, "B")
	insertStateTeamMember(t, ctx, postgres.Pool, teamA, actorA)
	insertStateTeamMember(t, ctx, postgres.Pool, teamA, inactiveA)

	if _, err := repository.Create(ctx, actorA, workspaceA, statesdomain.NewState{
		Name: "Cross tenant", Category: "started", Team: teamB, Color: "#000000",
	}); !errors.Is(err, statesdomain.ErrNotFound) {
		t.Fatalf("cross-tenant state create error = %v, want ErrNotFound", err)
	}
	if _, err := repository.Create(ctx, inactiveA, workspaceA, statesdomain.NewState{
		Name: "Inactive", Category: "started", Team: teamA, Color: "#000000",
	}); !errors.Is(err, statesdomain.ErrNotFound) {
		t.Fatalf("inactive state create error = %v, want ErrNotFound", err)
	}

	const concurrentCreates = 8
	start := make(chan struct{})
	created := make(chan statesdomain.State, concurrentCreates)
	errorsByCreate := make(chan error, concurrentCreates)
	var waitGroup sync.WaitGroup
	for index := range concurrentCreates {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			state, err := repository.Create(ctx, actorA, workspaceA, statesdomain.NewState{
				Name: fmt.Sprintf("Concurrent %d", index), Category: "started", Team: teamA,
				IsDefault: true, Color: "#123456",
			})
			if err != nil {
				errorsByCreate <- err
				return
			}
			created <- state
		}()
	}
	close(start)
	waitGroup.Wait()
	close(created)
	close(errorsByCreate)
	for err := range errorsByCreate {
		t.Fatalf("concurrent state create: %v", err)
	}

	orders := make(map[int]struct{}, concurrentCreates)
	var attachedState statesdomain.State
	for state := range created {
		orders[state.OrderIndex] = struct{}{}
		attachedState = state
	}
	if len(orders) != concurrentCreates {
		t.Fatalf("unique order indices = %d, want %d", len(orders), concurrentCreates)
	}
	var storedDefaults int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM statuses WHERE team_id = $1 AND is_default`, teamA).Scan(&storedDefaults); err != nil {
		t.Fatalf("count default states: %v", err)
	}
	if storedDefaults != 1 {
		t.Fatalf("stored default states = %d, want 1", storedDefaults)
	}

	memberStates, err := repository.List(ctx, workspaceA, actorA)
	if err != nil || len(memberStates) != concurrentCreates {
		t.Fatalf("member states = %d, %v, want %d", len(memberStates), err, concurrentCreates)
	}
	crossTenantStates, err := repository.List(ctx, workspaceB, actorA)
	if err != nil || len(crossTenantStates) != 0 {
		t.Fatalf("cross-tenant states = %#v, %v, want empty", crossTenantStates, err)
	}
	inactiveStates, err := repository.TeamListForMember(ctx, workspaceA, teamA, inactiveA)
	if err != nil || len(inactiveStates) != 0 {
		t.Fatalf("inactive member states = %#v, %v, want empty", inactiveStates, err)
	}

	insertStateStory(t, ctx, postgres.Pool, workspaceA, teamA, attachedState.ID)
	if err := repository.Delete(ctx, actorA, workspaceA, attachedState.ID); !errors.Is(err, statesdomain.ErrStatusHasStories) {
		t.Fatalf("delete attached state error = %v, want ErrStatusHasStories", err)
	}

	last, err := repository.Create(ctx, actorA, workspaceA, statesdomain.NewState{
		Name: "Only paused", Category: "paused", Team: teamA, Color: "#654321",
	})
	if err != nil {
		t.Fatalf("create last state fixture: %v", err)
	}
	if err := repository.Delete(ctx, actorA, workspaceA, last.ID); !errors.Is(err, statesdomain.ErrLastInCategory) {
		t.Fatalf("delete last category state error = %v, want ErrLastInCategory", err)
	}

	first, err := repository.Create(ctx, actorA, workspaceA, statesdomain.NewState{
		Name: "Canceled one", Category: "cancelled", Team: teamA, Color: "#111111",
	})
	if err != nil {
		t.Fatalf("create first delete-race state: %v", err)
	}
	second, err := repository.Create(ctx, actorA, workspaceA, statesdomain.NewState{
		Name: "Canceled two", Category: "cancelled", Team: teamA, Color: "#222222",
	})
	if err != nil {
		t.Fatalf("create second delete-race state: %v", err)
	}
	deleteErrors := make(chan error, 2)
	deleteStart := make(chan struct{})
	for _, stateID := range []uuid.UUID{first.ID, second.ID} {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-deleteStart
			deleteErrors <- repository.Delete(ctx, actorA, workspaceA, stateID)
		}()
	}
	close(deleteStart)
	waitGroup.Wait()
	close(deleteErrors)
	deleted, retained := 0, 0
	for err := range deleteErrors {
		switch {
		case err == nil:
			deleted++
		case errors.Is(err, statesdomain.ErrLastInCategory):
			retained++
		default:
			t.Fatalf("concurrent delete error = %v", err)
		}
	}
	if deleted != 1 || retained != 1 {
		t.Fatalf("concurrent deletes: deleted=%d retained=%d, want 1/1", deleted, retained)
	}
}

func insertStateWorkspace(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "States "+suffix, "states-"+suffix+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertStateUser(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name, is_active) VALUES ($1, $2, $3, $4, $5)`, id, suffix+"-"+uuid.NewString(), suffix+"-"+uuid.NewString()+"@example.test", "State "+suffix, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertStateWorkspaceMember(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertStateTeam(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')`, id, "Team "+suffix, workspaceID, suffix+uuid.NewString()[:6]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertStateTeamMember(t testing.TB, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func insertStateStory(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, teamID, statusID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO stories (team_id, title, workspace_id, status_id) VALUES ($1, $2, $3, $4)`, teamID, "Attached story", workspaceID, statusID); err != nil {
		t.Fatalf("insert story: %v", err)
	}
}
