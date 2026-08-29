//go:build integration

package workspaceuow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWorkspaceUnitOfWorkCommitsAndRollsBackTheCompleteGraph(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	userID := insertWorkspaceUOWUser(t, ctx, postgres.Pool, "graph")
	manager := newWorkspaceUOWManager(t, postgres.Pool)

	rollbackErr := errors.New("force workspace graph rollback")
	var rolledBackGraph workspaceGraph
	err := manager.WithinTransaction(ctx, func(transaction workspaces.Transaction) error {
		var createErr error
		rolledBackGraph, createErr = createWorkspaceGraph(ctx, transaction, userID, "uow-rollback-"+uuid.NewString())
		if createErr != nil {
			return createErr
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("rollback error = %v, want sentinel", err)
	}
	assertWorkspaceGraph(t, ctx, postgres.Pool, rolledBackGraph, userID, false)

	var committedGraph workspaceGraph
	if err := manager.WithinTransaction(ctx, func(transaction workspaces.Transaction) error {
		var createErr error
		committedGraph, createErr = createWorkspaceGraph(ctx, transaction, userID, "uow-commit-"+uuid.NewString())
		return createErr
	}); err != nil {
		t.Fatalf("commit workspace graph: %v", err)
	}
	assertWorkspaceGraph(t, ctx, postgres.Pool, committedGraph, userID, true)
}

func TestWorkspaceUnitOfWorkRejectsRetainedScopeAndConcurrentDuplicateSlug(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	userID := insertWorkspaceUOWUser(t, ctx, postgres.Pool, "concurrency")
	manager := newWorkspaceUOWManager(t, postgres.Pool)

	var retained workspaces.Transaction
	if err := manager.WithinTransaction(ctx, func(transaction workspaces.Transaction) error {
		retained = transaction
		return nil
	}); err != nil {
		t.Fatalf("capture transaction scope: %v", err)
	}
	if _, err := retained.CreateWorkspace(ctx, workspaces.CoreWorkspace{Name: "Late", Slug: "late-" + uuid.NewString()}, userID); !errors.Is(err, ErrTransactionScopeClosed) {
		t.Fatalf("retained scope error = %v, want ErrTransactionScopeClosed", err)
	}

	slug := "uow-concurrent-" + uuid.NewString()
	start := make(chan struct{})
	errorsChannel := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			errorsChannel <- manager.WithinTransaction(ctx, func(transaction workspaces.Transaction) error {
				_, err := createWorkspaceGraph(ctx, transaction, userID, slug)
				return err
			})
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsChannel)
	successes, conflicts := 0, 0
	for createErr := range errorsChannel {
		switch {
		case createErr == nil:
			successes++
		case errors.Is(createErr, workspacedomain.ErrSlugTaken):
			conflicts++
		default:
			t.Fatalf("concurrent workspace creation error = %v", createErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent outcomes success/conflict = %d/%d, want 1/1", successes, conflicts)
	}
	assertSingleWorkspaceGraphBySlug(t, ctx, postgres.Pool, slug)
}

type workspaceGraph struct {
	workspaceID uuid.UUID
	teamID      uuid.UUID
	slug        string
}

func createWorkspaceGraph(ctx context.Context, transaction workspaces.Transaction, userID uuid.UUID, slug string) (workspaceGraph, error) {
	workspace, err := transaction.CreateWorkspace(ctx, workspaces.CoreWorkspace{
		Name: "Unit of work", Slug: slug, TeamSize: "2-10",
	}, userID)
	if err != nil {
		return workspaceGraph{}, err
	}
	graph := workspaceGraph{workspaceID: workspace.ID, slug: slug}
	if err := transaction.AddWorkspaceMember(ctx, workspace.ID, userID, "admin"); err != nil {
		return graph, err
	}
	team, err := transaction.CreateTeam(ctx, workspaces.DefaultTeam{
		Name: "Team 1", Code: "TM", Color: workspace.Color, Workspace: workspace.ID,
	})
	if err != nil {
		return graph, err
	}
	graph.teamID = team.ID
	if err := transaction.AddTeamMember(ctx, team.ID, userID, workspace.ID); err != nil {
		return graph, err
	}
	if err := transaction.UpdateLastUsedWorkspace(ctx, userID, workspace.ID); err != nil {
		return graph, err
	}
	if err := transaction.InitializeWorkspaceSettings(ctx, workspace.ID); err != nil {
		return graph, err
	}
	return graph, nil
}

func newWorkspaceUOWManager(t *testing.T, pool *pgxpool.Pool) *Manager {
	t.Helper()
	manager, err := New(
		pool,
		workspacesrepository.New(pool),
		teamsrepository.New(pool),
		usersrepository.New(pool),
	)
	if err != nil {
		t.Fatalf("new workspace unit of work: %v", err)
	}
	return manager
}

func insertWorkspaceUOWUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name) VALUES ($1, $2, $3, $4)`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "UOW "+label); err != nil {
		t.Fatalf("insert unit-of-work user: %v", err)
	}
	return id
}

func assertWorkspaceGraph(t *testing.T, ctx context.Context, pool *pgxpool.Pool, graph workspaceGraph, userID uuid.UUID, want bool) {
	t.Helper()
	var workspaceCount, membershipCount, teamCount, teamMembershipCount int
	var workspaceSettingsCount, objectiveStatusCount, storyStatusCount, teamSettingsCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM workspaces WHERE workspace_id = $1 AND slug = $4),
			(SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND user_id = $3),
			(SELECT COUNT(*) FROM teams WHERE team_id = $2 AND workspace_id = $1),
			(SELECT COUNT(*) FROM team_members WHERE team_id = $2 AND user_id = $3),
			(SELECT COUNT(*) FROM workspace_settings WHERE workspace_id = $1),
			(SELECT COUNT(*) FROM objective_statuses WHERE workspace_id = $1),
			(SELECT COUNT(*) FROM statuses WHERE workspace_id = $1 AND team_id = $2),
			(SELECT COUNT(*) FROM team_story_automation_settings WHERE workspace_id = $1 AND team_id = $2)
	`, graph.workspaceID, graph.teamID, userID, graph.slug).Scan(
		&workspaceCount, &membershipCount, &teamCount, &teamMembershipCount,
		&workspaceSettingsCount, &objectiveStatusCount, &storyStatusCount, &teamSettingsCount,
	); err != nil {
		t.Fatalf("read workspace graph: %v", err)
	}
	counts := []int{workspaceCount, membershipCount, teamCount, teamMembershipCount, workspaceSettingsCount, teamSettingsCount}
	if !want {
		counts = append(counts, objectiveStatusCount, storyStatusCount)
		for _, count := range counts {
			if count != 0 {
				t.Fatalf("rolled-back workspace graph counts = %#v plus objective/story %d/%d, want zeros", counts, objectiveStatusCount, storyStatusCount)
			}
		}
		assertWorkspaceUOWLastUsed(t, ctx, pool, userID, nil)
		return
	}
	for _, count := range counts {
		if count != 1 {
			t.Fatalf("committed workspace graph core counts = %#v, want ones", counts)
		}
	}
	if objectiveStatusCount != len(workspacedomain.DefaultObjectiveStatuses) || storyStatusCount == 0 {
		t.Fatalf("committed default counts objective/story = %d/%d", objectiveStatusCount, storyStatusCount)
	}
	assertWorkspaceUOWLastUsed(t, ctx, pool, userID, &graph.workspaceID)
}

func assertSingleWorkspaceGraphBySlug(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug string) {
	t.Helper()
	var workspaceCount, membershipCount, teamCount, teamMembershipCount, settingsCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM workspaces WHERE slug = $1),
			(SELECT COUNT(*) FROM workspace_members AS member INNER JOIN workspaces AS workspace USING (workspace_id) WHERE workspace.slug = $1),
			(SELECT COUNT(*) FROM teams AS team INNER JOIN workspaces AS workspace USING (workspace_id) WHERE workspace.slug = $1),
			(SELECT COUNT(*) FROM team_members AS member INNER JOIN teams AS team USING (team_id) INNER JOIN workspaces AS workspace ON workspace.workspace_id = team.workspace_id WHERE workspace.slug = $1),
			(SELECT COUNT(*) FROM workspace_settings AS settings INNER JOIN workspaces AS workspace USING (workspace_id) WHERE workspace.slug = $1)
	`, slug).Scan(&workspaceCount, &membershipCount, &teamCount, &teamMembershipCount, &settingsCount); err != nil {
		t.Fatalf("read concurrent workspace graph: %v", err)
	}
	if workspaceCount != 1 || membershipCount != 1 || teamCount != 1 || teamMembershipCount != 1 || settingsCount != 1 {
		t.Fatalf("concurrent graph counts = %d/%d/%d/%d/%d, want ones", workspaceCount, membershipCount, teamCount, teamMembershipCount, settingsCount)
	}
}

func assertWorkspaceUOWLastUsed(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, want *uuid.UUID) {
	t.Helper()
	var got *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT last_used_workspace_id FROM users WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatalf("read last-used workspace: %v", err)
	}
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil || *want != *got {
		t.Fatalf("last-used workspace = %v, want %v", got, want)
	}
}
