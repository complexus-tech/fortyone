//go:build integration

package activitiesrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	activitiesdomain "github.com/complexus-tech/projects-api/internal/modules/activities/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestActivitiesEnforceTenantMembershipAndConcurrentAppends(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	workspaceA := insertActivityWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertActivityWorkspace(t, ctx, postgres.Pool, "b")
	actorA := insertActivityUser(t, ctx, postgres.Pool, "actor-a", true)
	actorB := insertActivityUser(t, ctx, postgres.Pool, "actor-b", true)
	inactiveA := insertActivityUser(t, ctx, postgres.Pool, "inactive-a", false)
	insertActivityMember(t, ctx, postgres.Pool, workspaceA, actorA)
	insertActivityMember(t, ctx, postgres.Pool, workspaceB, actorB)
	insertActivityMember(t, ctx, postgres.Pool, workspaceA, inactiveA)
	teamA := insertActivityTeam(t, ctx, postgres.Pool, workspaceA, "A")
	teamB := insertActivityTeam(t, ctx, postgres.Pool, workspaceB, "B")
	storyA := insertActivityStory(t, ctx, postgres.Pool, workspaceA, teamA, false)
	deletedStoryA := insertActivityStory(t, ctx, postgres.Pool, workspaceA, teamA, true)
	storyB := insertActivityStory(t, ctx, postgres.Pool, workspaceB, teamB, false)

	valid := activitiesdomain.NewActivity{
		StoryID: storyA, UserID: actorA, Type: "updated", Field: "priority",
		CurrentValue: "high", WorkspaceID: workspaceA,
	}
	if err := repository.Create(ctx, valid); err != nil {
		t.Fatalf("create valid activity: %v", err)
	}
	for _, test := range []struct {
		name  string
		input activitiesdomain.NewActivity
	}{
		{name: "workspace mismatch", input: activitiesdomain.NewActivity{StoryID: storyB, UserID: actorA, Type: "updated", Field: "priority", CurrentValue: "low", WorkspaceID: workspaceA}},
		{name: "actor outside workspace", input: activitiesdomain.NewActivity{StoryID: storyA, UserID: actorB, Type: "updated", Field: "priority", CurrentValue: "low", WorkspaceID: workspaceA}},
		{name: "inactive actor", input: activitiesdomain.NewActivity{StoryID: storyA, UserID: inactiveA, Type: "updated", Field: "priority", CurrentValue: "low", WorkspaceID: workspaceA}},
		{name: "deleted story", input: activitiesdomain.NewActivity{StoryID: deletedStoryA, UserID: actorA, Type: "updated", Field: "priority", CurrentValue: "low", WorkspaceID: workspaceA}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := repository.Create(ctx, test.input); !errors.Is(err, activitiesdomain.ErrScopeMismatch) {
				t.Fatalf("Create() error = %v, want ErrScopeMismatch", err)
			}
		})
	}

	const concurrentAppends = 8
	var waitGroup sync.WaitGroup
	appendErrors := make(chan error, concurrentAppends)
	for range concurrentAppends {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			entry := valid
			entry.CurrentValue = uuid.NewString()
			appendErrors <- repository.Create(ctx, entry)
		}()
	}
	waitGroup.Wait()
	close(appendErrors)
	for err := range appendErrors {
		if err != nil {
			t.Fatalf("concurrent activity append: %v", err)
		}
	}

	now := time.Now().UTC()
	activities, err := repository.GetActivities(ctx, actorA, concurrentAppends+1, workspaceA, activitiesdomain.Filters{
		StartDate: now.Add(-time.Hour), EndDate: now.Add(time.Hour),
	})
	if err != nil || len(activities) != concurrentAppends+1 {
		t.Fatalf("member activities = %d, %v, want %d", len(activities), err, concurrentAppends+1)
	}
	seen := make(map[uuid.UUID]struct{}, len(activities))
	for _, activity := range activities {
		if activity.WorkspaceID != workspaceA || activity.User.ID != actorA || !activity.User.IsActive {
			t.Fatalf("activity escaped member scope: %#v", activity)
		}
		seen[activity.ID] = struct{}{}
	}
	if len(seen) != concurrentAppends+1 {
		t.Fatalf("unique activity IDs = %d, want %d", len(seen), concurrentAppends+1)
	}

	crossTenant, err := repository.GetActivities(ctx, actorA, 100, workspaceB, activitiesdomain.Filters{
		StartDate: now.Add(-time.Hour), EndDate: now.Add(time.Hour),
	})
	if err != nil || len(crossTenant) != 0 {
		t.Fatalf("cross-tenant activities = %#v, %v, want empty", crossTenant, err)
	}
	inactive, err := repository.GetActivities(ctx, inactiveA, 100, workspaceA, activitiesdomain.Filters{
		StartDate: now.Add(-time.Hour), EndDate: now.Add(time.Hour),
	})
	if err != nil || len(inactive) != 0 {
		t.Fatalf("inactive activities = %#v, %v, want empty", inactive, err)
	}
}

func insertActivityWorkspace(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Activities "+suffix, "activities-"+suffix+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertActivityUser(t testing.TB, ctx context.Context, pool *pgxpool.Pool, suffix string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name, is_active) VALUES ($1, $2, $3, $4, $5)`, id, suffix+"-"+uuid.NewString(), suffix+"-"+uuid.NewString()+"@example.test", "Activity "+suffix, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertActivityMember(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertActivityTeam(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, suffix string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')`, id, "Activity Team "+suffix, workspaceID, "A"+uuid.NewString()[:7]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertActivityStory(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID, teamID uuid.UUID, deleted bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	deletedAt := any(nil)
	if deleted {
		deletedAt = time.Now().UTC()
	}
	if _, err := pool.Exec(ctx, `INSERT INTO stories (id, team_id, title, workspace_id, deleted_at) VALUES ($1, $2, $3, $4, $5)`, id, teamID, "Activity story "+uuid.NewString(), workspaceID, deletedAt); err != nil {
		t.Fatalf("insert story: %v", err)
	}
	return id
}
