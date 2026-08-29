//go:build integration

package objectivestatusrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	objectivestatusdomain "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestObjectiveStatusesEnforceTenantRolesAndConcurrentInvariants(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	workspaceA := insertObjectiveStatusWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertObjectiveStatusWorkspace(t, ctx, postgres.Pool, "b")
	adminA := insertObjectiveStatusUser(t, ctx, postgres.Pool, "admin-a", true)
	memberA := insertObjectiveStatusUser(t, ctx, postgres.Pool, "member-a", true)
	inactiveA := insertObjectiveStatusUser(t, ctx, postgres.Pool, "inactive-a", false)
	insertObjectiveStatusMember(t, ctx, postgres.Pool, workspaceA, adminA, "admin")
	insertObjectiveStatusMember(t, ctx, postgres.Pool, workspaceA, memberA, "member")
	insertObjectiveStatusMember(t, ctx, postgres.Pool, workspaceA, inactiveA, "admin")

	input := objectivestatusdomain.NewStatus{Name: "Unauthorized", Category: "started", Color: "#000000"}
	if _, err := repository.Create(ctx, memberA, workspaceA, input); !errors.Is(err, objectivestatusdomain.ErrNotFound) {
		t.Fatalf("member objective-status create error = %v, want ErrNotFound", err)
	}
	if _, err := repository.Create(ctx, adminA, workspaceB, input); !errors.Is(err, objectivestatusdomain.ErrNotFound) {
		t.Fatalf("cross-tenant objective-status create error = %v, want ErrNotFound", err)
	}
	if _, err := repository.Create(ctx, inactiveA, workspaceA, input); !errors.Is(err, objectivestatusdomain.ErrNotFound) {
		t.Fatalf("inactive objective-status create error = %v, want ErrNotFound", err)
	}

	const concurrentCreates = 8
	start := make(chan struct{})
	created := make(chan objectivestatusdomain.Status, concurrentCreates)
	errorsByCreate := make(chan error, concurrentCreates)
	var waitGroup sync.WaitGroup
	for index := range concurrentCreates {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			status, err := repository.Create(ctx, adminA, workspaceA, objectivestatusdomain.NewStatus{
				Name: fmt.Sprintf("Concurrent %d", index), Category: "started",
				IsDefault: true, Color: "#123456",
			})
			if err != nil {
				errorsByCreate <- err
				return
			}
			created <- status
		}()
	}
	close(start)
	waitGroup.Wait()
	close(created)
	close(errorsByCreate)
	for err := range errorsByCreate {
		t.Fatalf("concurrent objective-status create: %v", err)
	}

	orders := make(map[int]struct{}, concurrentCreates)
	var attachedStatus objectivestatusdomain.Status
	for status := range created {
		orders[status.OrderIndex] = struct{}{}
		attachedStatus = status
	}
	if len(orders) != concurrentCreates {
		t.Fatalf("unique objective-status order indices = %d, want %d", len(orders), concurrentCreates)
	}
	var storedDefaults int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM objective_statuses WHERE workspace_id = $1 AND is_default`, workspaceA).Scan(&storedDefaults); err != nil {
		t.Fatalf("count objective-status defaults: %v", err)
	}
	if storedDefaults != 1 {
		t.Fatalf("stored objective-status defaults = %d, want 1", storedDefaults)
	}

	memberStatuses, err := repository.ListForMember(ctx, memberA, workspaceA)
	if err != nil || len(memberStatuses) != concurrentCreates {
		t.Fatalf("member objective statuses = %d, %v, want %d", len(memberStatuses), err, concurrentCreates)
	}
	crossTenantStatuses, err := repository.ListForMember(ctx, memberA, workspaceB)
	if err != nil || len(crossTenantStatuses) != 0 {
		t.Fatalf("cross-tenant objective statuses = %#v, %v", crossTenantStatuses, err)
	}
	inactiveStatuses, err := repository.ListForMember(ctx, inactiveA, workspaceA)
	if err != nil || len(inactiveStatuses) != 0 {
		t.Fatalf("inactive objective statuses = %#v, %v", inactiveStatuses, err)
	}

	insertObjectiveWithStatus(t, ctx, postgres.Pool, workspaceA, attachedStatus.ID)
	if err := repository.Delete(ctx, adminA, workspaceA, attachedStatus.ID); !errors.Is(err, objectivestatusdomain.ErrStatusHasObjectives) {
		t.Fatalf("delete attached objective status error = %v, want ErrStatusHasObjectives", err)
	}

	last, err := repository.Create(ctx, adminA, workspaceA, objectivestatusdomain.NewStatus{
		Name: "Only paused", Category: "paused", Color: "#654321",
	})
	if err != nil {
		t.Fatalf("create last objective status fixture: %v", err)
	}
	if err := repository.Delete(ctx, adminA, workspaceA, last.ID); !errors.Is(err, objectivestatusdomain.ErrLastInCategory) {
		t.Fatalf("delete last objective status error = %v, want ErrLastInCategory", err)
	}

	first, err := repository.Create(ctx, adminA, workspaceA, objectivestatusdomain.NewStatus{
		Name: "Canceled one", Category: "cancelled", Color: "#111111",
	})
	if err != nil {
		t.Fatalf("create first objective-status delete-race fixture: %v", err)
	}
	second, err := repository.Create(ctx, adminA, workspaceA, objectivestatusdomain.NewStatus{
		Name: "Canceled two", Category: "cancelled", Color: "#222222",
	})
	if err != nil {
		t.Fatalf("create second objective-status delete-race fixture: %v", err)
	}
	deleteErrors := make(chan error, 2)
	deleteStart := make(chan struct{})
	for _, statusID := range []uuid.UUID{first.ID, second.ID} {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-deleteStart
			deleteErrors <- repository.Delete(ctx, adminA, workspaceA, statusID)
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
		case errors.Is(err, objectivestatusdomain.ErrLastInCategory):
			retained++
		default:
			t.Fatalf("concurrent objective-status delete error = %v", err)
		}
	}
	if deleted != 1 || retained != 1 {
		t.Fatalf("concurrent objective-status deletes: deleted=%d retained=%d, want 1/1", deleted, retained)
	}
}

func insertObjectiveStatusWorkspace(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Objective statuses "+suffix, "objective-statuses-"+suffix+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertObjectiveStatusUser(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name, is_active) VALUES ($1, $2, $3, $4, $5)`, id, suffix+"-"+uuid.NewString(), suffix+"-"+uuid.NewString()+"@example.test", "Objective status "+suffix, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertObjectiveStatusMember(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, role string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, CAST($3 AS user_role))`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertObjectiveWithStatus(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, statusID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO objectives (name, workspace_id, status_id, sequence_id) VALUES ($1, $2, $3, 1)`, "Attached objective "+uuid.NewString(), workspaceID, statusID); err != nil {
		t.Fatalf("insert objective: %v", err)
	}
}
