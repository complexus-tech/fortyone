//go:build integration

package labelsrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	labelsdomain "github.com/complexus-tech/projects-api/internal/modules/labels/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLabelsEnforceTenantMembershipAndStablePagination(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	workspaceA := insertLabelWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertLabelWorkspace(t, ctx, postgres.Pool, "b")
	actorA := insertLabelUser(t, ctx, postgres.Pool, "actor-a", true)
	actorB := insertLabelUser(t, ctx, postgres.Pool, "actor-b", true)
	inactiveA := insertLabelUser(t, ctx, postgres.Pool, "inactive-a", false)
	insertLabelMember(t, ctx, postgres.Pool, workspaceA, actorA)
	insertLabelMember(t, ctx, postgres.Pool, workspaceB, actorB)
	insertLabelMember(t, ctx, postgres.Pool, workspaceA, inactiveA)
	teamA := insertLabelTeam(t, ctx, postgres.Pool, workspaceA, "A")
	teamA2 := insertLabelTeam(t, ctx, postgres.Pool, workspaceA, "A2")
	teamB := insertLabelTeam(t, ctx, postgres.Pool, workspaceB, "B")

	if _, err := repository.CreateLabel(ctx, actorA, labelsdomain.NewLabel{
		Name: "Cross-team", TeamID: &teamB, WorkspaceID: workspaceA, Color: "#000000",
	}); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("cross-tenant team label create error = %v, want ErrNotFound", err)
	}
	if _, err := repository.CreateLabel(ctx, actorB, labelsdomain.NewLabel{
		Name: "Cross-actor", TeamID: &teamA, WorkspaceID: workspaceA, Color: "#000000",
	}); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("cross-tenant actor label create error = %v, want ErrNotFound", err)
	}
	if _, err := repository.CreateLabel(ctx, inactiveA, labelsdomain.NewLabel{
		Name: "Inactive", WorkspaceID: workspaceA, Color: "#000000",
	}); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("inactive actor label create error = %v, want ErrNotFound", err)
	}

	global := createLabel(t, ctx, repository, actorA, labelsdomain.NewLabel{
		Name: "Global security", WorkspaceID: workspaceA, Color: "#111111",
	})
	teamScoped := createLabel(t, ctx, repository, actorA, labelsdomain.NewLabel{
		Name: "Team security", TeamID: &teamA, WorkspaceID: workspaceA, Color: "#222222",
	})
	createLabel(t, ctx, repository, actorA, labelsdomain.NewLabel{
		Name: "Other team", TeamID: &teamA2, WorkspaceID: workspaceA, Color: "#333333",
	})

	teamLabels, err := repository.GetLabels(ctx, actorA, workspaceA, labelsdomain.Filters{TeamID: &teamA})
	if err != nil || len(teamLabels) != 2 {
		t.Fatalf("team-filtered labels = %#v, %v, want global and team-scoped", teamLabels, err)
	}
	searchLabels, err := repository.GetLabels(ctx, actorA, workspaceA, labelsdomain.Filters{Search: "security"})
	if err != nil || len(searchLabels) != 2 {
		t.Fatalf("searched labels = %#v, %v, want two", searchLabels, err)
	}
	for _, test := range []struct {
		name        string
		actorID     uuid.UUID
		workspaceID uuid.UUID
	}{
		{name: "other tenant", actorID: actorB, workspaceID: workspaceB},
		{name: "inactive account", actorID: inactiveA, workspaceID: workspaceA},
	} {
		t.Run(test.name, func(t *testing.T) {
			labels, listErr := repository.GetLabels(ctx, test.actorID, test.workspaceID, labelsdomain.Filters{})
			if listErr != nil || len(labels) != 0 {
				t.Fatalf("labels = %#v, %v, want empty", labels, listErr)
			}
		})
	}
	if _, err := repository.GetLabel(ctx, actorB, global.ID, workspaceB); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("cross-tenant label get error = %v, want ErrNotFound", err)
	}
	if _, err := repository.UpdateLabel(ctx, actorB, global.ID, workspaceA, "Changed", "#ffffff"); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("cross-actor label update error = %v, want ErrNotFound", err)
	}
	if err := repository.DeleteLabel(ctx, actorB, global.ID, workspaceA); !errors.Is(err, labelsdomain.ErrNotFound) {
		t.Fatalf("cross-actor label delete error = %v, want ErrNotFound", err)
	}

	const concurrentLabels = 8
	var waitGroup sync.WaitGroup
	createErrors := make(chan error, concurrentLabels)
	for range concurrentLabels {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, createErr := repository.CreateLabel(ctx, actorA, labelsdomain.NewLabel{
				Name: "Concurrent " + uuid.NewString(), WorkspaceID: workspaceA, Color: "#444444",
			})
			createErrors <- createErr
		}()
	}
	waitGroup.Wait()
	close(createErrors)
	for err := range createErrors {
		if err != nil {
			t.Fatalf("concurrent label create: %v", err)
		}
	}

	if _, err := postgres.Pool.Exec(ctx, `UPDATE labels SET created_at = TIMESTAMPTZ '2026-01-01 00:00:00+00' WHERE workspace_id = $1`, workspaceA); err != nil {
		t.Fatalf("align label timestamps: %v", err)
	}
	pageSize := 6
	firstPage, err := repository.GetLabels(ctx, actorA, workspaceA, labelsdomain.Filters{Limit: &pageSize})
	if err != nil || len(firstPage) != pageSize {
		t.Fatalf("first label page = %d, %v, want %d", len(firstPage), err, pageSize)
	}
	secondPage, err := repository.GetLabels(ctx, actorA, workspaceA, labelsdomain.Filters{Limit: &pageSize, Offset: pageSize})
	if err != nil || len(secondPage) != concurrentLabels+3-pageSize {
		t.Fatalf("second label page = %d, %v", len(secondPage), err)
	}
	seen := make(map[uuid.UUID]struct{}, concurrentLabels+3)
	for _, label := range append(firstPage, secondPage...) {
		if _, duplicate := seen[label.ID]; duplicate {
			t.Fatalf("label %s appeared in both deterministic pages", label.ID)
		}
		seen[label.ID] = struct{}{}
	}
	if len(seen) != concurrentLabels+3 {
		t.Fatalf("unique paged labels = %d, want %d", len(seen), concurrentLabels+3)
	}

	if err := repository.DeleteLabel(ctx, actorA, teamScoped.ID, workspaceA); err != nil {
		t.Fatalf("delete member label: %v", err)
	}
}

func createLabel(t testing.TB, ctx context.Context, repository *Repository, actorID uuid.UUID, input labelsdomain.NewLabel) labelsdomain.Label {
	t.Helper()
	label, err := repository.CreateLabel(ctx, actorID, input)
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	return label
}

func insertLabelWorkspace(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Labels "+suffix, "labels-"+suffix+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertLabelUser(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name, is_active) VALUES ($1, $2, $3, $4, $5)`, id, suffix+"-"+uuid.NewString(), suffix+"-"+uuid.NewString()+"@example.test", "Label "+suffix, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertLabelMember(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertLabelTeam(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')`, id, "Label Team "+suffix, workspaceID, "L"+uuid.NewString()[:7]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}
