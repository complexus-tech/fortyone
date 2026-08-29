//go:build integration

package searchrepository

import (
	"context"
	"strconv"
	"testing"
	"time"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type searchFixture struct {
	workspaceA   uuid.UUID
	workspaceB   uuid.UUID
	teamA        uuid.UUID
	teamA2       uuid.UUID
	teamHidden   uuid.UUID
	teamB        uuid.UUID
	statusA      uuid.UUID
	statusA2     uuid.UUID
	objectiveA   uuid.UUID
	objectiveA2  uuid.UUID
	actor        uuid.UUID
	inactiveUser uuid.UUID
	labelA       uuid.UUID
	storyNewest  uuid.UUID
}

func TestSearchRepositoryEnforcesTenancyFiltersAndStablePagination(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	assertSearchPostgres18(t, ctx, postgres.Pool)
	fixture := seedSearchFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)

	params := search.SearchParams{
		Query:    "integration",
		SortBy:   search.SortByCreated,
		Page:     1,
		PageSize: 2,
	}
	firstPage, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, params)
	if err != nil {
		t.Fatalf("search first story page: %v", err)
	}
	if len(firstPage) != 2 || total != 3 || firstPage[0].ID != fixture.storyNewest {
		t.Fatalf("first story page = %#v, total %d", firstPage, total)
	}
	for _, story := range firstPage {
		if story.Workspace != fixture.workspaceA || story.Team == fixture.teamHidden || story.Team == fixture.teamB {
			t.Fatalf("story escaped actor scope: %#v", story)
		}
	}

	params.Page = 2
	secondPage, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, params)
	if err != nil {
		t.Fatalf("search second story page: %v", err)
	}
	if len(secondPage) != 1 || total != 3 {
		t.Fatalf("second story page = %#v, total %d", secondPage, total)
	}
	params.Page = 10
	emptyPage, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, params)
	if err != nil || len(emptyPage) != 0 || total != 3 {
		t.Fatalf("empty story page = %#v, total %d, error %v", emptyPage, total, err)
	}

	priority := "high"
	filtered, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, search.SearchParams{
		TeamID:   &fixture.teamA,
		LabelID:  &fixture.labelA,
		Priority: &priority,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("search filtered stories: %v", err)
	}
	if len(filtered) != 1 || total != 1 || len(filtered[0].Labels) != 1 || filtered[0].Labels[0] != fixture.labelA {
		t.Fatalf("filtered stories = %#v, total %d", filtered, total)
	}

	assertNoSearchResults(t, ctx, repository, fixture.workspaceB, fixture.actor)
	assertNoSearchResults(t, ctx, repository, fixture.workspaceA, fixture.inactiveUser)

	objectives, total, err := repository.SearchObjectives(ctx, fixture.workspaceA, fixture.actor, search.SearchParams{
		Query:    "objective",
		SortBy:   search.SortByCreated,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("search objectives: %v", err)
	}
	if len(objectives) != 2 || total != 2 {
		t.Fatalf("objectives = %#v, total %d", objectives, total)
	}
	for _, objective := range objectives {
		if objective.Workspace != fixture.workspaceA || objective.Team == fixture.teamHidden || objective.Team == fixture.teamB {
			t.Fatalf("objective escaped actor scope: %#v", objective)
		}
	}

	similar, err := repository.FindSimilarStories(
		ctx,
		fixture.workspaceA,
		fixture.actor,
		"Build Slack integration",
		nil,
		5,
	)
	if err != nil {
		t.Fatalf("find similar stories: %v", err)
	}
	if len(similar) == 0 || similar[0].Confidence != 1 {
		t.Fatalf("similar stories = %#v", similar)
	}
	for _, story := range similar {
		if story.Team == fixture.teamHidden || story.Team == fixture.teamB {
			t.Fatalf("similar story escaped actor scope: %#v", story)
		}
	}
}

func assertNoSearchResults(
	t *testing.T,
	ctx context.Context,
	repository *repo,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) {
	t.Helper()
	params := search.SearchParams{SortBy: search.SortByRelevance, Page: 1, PageSize: 20}
	stories, storyTotal, err := repository.SearchStories(ctx, workspaceID, actorID, params)
	if err != nil || len(stories) != 0 || storyTotal != 0 {
		t.Fatalf("scoped story search = %#v, total %d, error %v", stories, storyTotal, err)
	}
	objectives, objectiveTotal, err := repository.SearchObjectives(ctx, workspaceID, actorID, params)
	if err != nil || len(objectives) != 0 || objectiveTotal != 0 {
		t.Fatalf("scoped objective search = %#v, total %d, error %v", objectives, objectiveTotal, err)
	}
}

func assertSearchPostgres18(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var rawVersion string
	if err := pool.QueryRow(ctx, "SHOW server_version_num").Scan(&rawVersion); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(rawVersion)
	if err != nil {
		t.Fatalf("parse PostgreSQL version %q: %v", rawVersion, err)
	}
	if version < 180000 || version >= 190000 {
		t.Fatalf("search integration tests require PostgreSQL 18, got %d", version)
	}
}

func seedSearchFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) searchFixture {
	t.Helper()
	fixture := searchFixture{
		workspaceA:   uuid.New(),
		workspaceB:   uuid.New(),
		teamA:        uuid.New(),
		teamA2:       uuid.New(),
		teamHidden:   uuid.New(),
		teamB:        uuid.New(),
		statusA:      uuid.New(),
		statusA2:     uuid.New(),
		objectiveA:   uuid.New(),
		objectiveA2:  uuid.New(),
		actor:        uuid.New(),
		inactiveUser: uuid.New(),
		labelA:       uuid.New(),
		storyNewest:  uuid.New(),
	}
	insertSearchUser(t, ctx, pool, fixture.actor, "actor", true)
	insertSearchUser(t, ctx, pool, fixture.inactiveUser, "inactive", false)
	insertSearchWorkspace(t, ctx, pool, fixture.workspaceA, "a", fixture.actor)
	insertSearchWorkspace(t, ctx, pool, fixture.workspaceB, "b", fixture.actor)
	for _, userID := range []uuid.UUID{fixture.actor, fixture.inactiveUser} {
		mustSearchExec(t, ctx, pool, `
			INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')
		`, fixture.workspaceA, userID)
	}
	insertSearchTeam(t, ctx, pool, fixture.teamA, fixture.workspaceA, "A")
	insertSearchTeam(t, ctx, pool, fixture.teamA2, fixture.workspaceA, "A2")
	insertSearchTeam(t, ctx, pool, fixture.teamHidden, fixture.workspaceA, "AH")
	insertSearchTeam(t, ctx, pool, fixture.teamB, fixture.workspaceB, "B")
	for _, teamID := range []uuid.UUID{fixture.teamA, fixture.teamA2, fixture.teamB} {
		mustSearchExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", teamID, fixture.actor)
	}
	mustSearchExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamA, fixture.inactiveUser)

	insertSearchStatus(t, ctx, pool, fixture.statusA, fixture.workspaceA, fixture.teamA, "Started")
	insertSearchStatus(t, ctx, pool, fixture.statusA2, fixture.workspaceA, fixture.teamA2, "Backlog")
	objectiveStatusA := uuid.New()
	objectiveStatusB := uuid.New()
	insertSearchObjectiveStatus(t, ctx, pool, objectiveStatusA, fixture.workspaceA)
	insertSearchObjectiveStatus(t, ctx, pool, objectiveStatusB, fixture.workspaceB)
	mustSearchExec(t, ctx, pool, `
		INSERT INTO labels (label_id, name, team_id, workspace_id) VALUES ($1, 'Integration', $2, $3)
	`, fixture.labelA, fixture.teamA, fixture.workspaceA)

	base := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	storyOld := uuid.New()
	storyTeamTwo := uuid.New()
	insertSearchStory(t, ctx, pool, storyOld, fixture.teamA, fixture.workspaceA, fixture.statusA, fixture.actor, 1, "Build Slack integration", "high", base)
	insertSearchStory(t, ctx, pool, fixture.storyNewest, fixture.teamA, fixture.workspaceA, fixture.statusA, fixture.actor, 2, "Refine Slack integration", "medium", base.Add(2*time.Hour))
	insertSearchStory(t, ctx, pool, storyTeamTwo, fixture.teamA2, fixture.workspaceA, fixture.statusA2, fixture.actor, 1, "Build GitLab integration", "high", base.Add(time.Hour))
	insertSearchStory(t, ctx, pool, uuid.New(), fixture.teamHidden, fixture.workspaceA, uuid.Nil, fixture.actor, 1, "Hidden integration", "high", base.Add(3*time.Hour))
	insertSearchStory(t, ctx, pool, uuid.New(), fixture.teamB, fixture.workspaceB, uuid.Nil, fixture.actor, 1, "Secret Slack integration", "high", base.Add(4*time.Hour))
	mustSearchExec(t, ctx, pool, "UPDATE stories SET deleted_at = NOW() WHERE id = $1", storyTeamTwo)
	insertSearchStory(t, ctx, pool, uuid.New(), fixture.teamA2, fixture.workspaceA, fixture.statusA2, fixture.actor, 2, "Ship GitHub integration", "low", base.Add(time.Hour))
	mustSearchExec(t, ctx, pool, "INSERT INTO story_labels (story_id, label_id) VALUES ($1, $2)", storyOld, fixture.labelA)

	insertSearchObjective(t, ctx, pool, fixture.objectiveA, fixture.teamA, fixture.workspaceA, objectiveStatusA, "First integration objective", base)
	insertSearchObjective(t, ctx, pool, fixture.objectiveA2, fixture.teamA2, fixture.workspaceA, objectiveStatusA, "Second integration objective", base.Add(time.Hour))
	insertSearchObjective(t, ctx, pool, uuid.New(), fixture.teamHidden, fixture.workspaceA, objectiveStatusA, "Hidden objective", base.Add(2*time.Hour))
	insertSearchObjective(t, ctx, pool, uuid.New(), fixture.teamB, fixture.workspaceB, objectiveStatusB, "Secret objective", base.Add(3*time.Hour))
	return fixture
}

func insertSearchUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+suffix, label+suffix+"@example.com", "Search "+label, active)
}

func insertSearchWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, creator uuid.UUID) {
	t.Helper()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by) VALUES ($1, $2, $3, $4)
	`, id, "Search "+label, "search-"+label+"-"+uuid.NewString(), creator)
}

func insertSearchTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Search "+code, workspaceID, code)
}

func insertSearchStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID, teamID uuid.UUID, name string) {
	t.Helper()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO statuses (status_id, name, category, workspace_id, team_id) VALUES ($1, $2, 'started', $3, $4)
	`, id, name, workspaceID, teamID)
}

func insertSearchObjectiveStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID) {
	t.Helper()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO objective_statuses (status_id, name, category, workspace_id) VALUES ($1, 'Active', 'started', $2)
	`, id, workspaceID)
}

func insertSearchStory(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, teamID, workspaceID, statusID, assigneeID uuid.UUID, sequence int, title, priority string, createdAt time.Time) {
	t.Helper()
	var nullableStatus any = statusID
	if statusID == uuid.Nil {
		nullableStatus = nil
	}
	mustSearchExec(t, ctx, pool, `
		INSERT INTO stories (
			id, sequence_id, team_id, title, status_id, assignee_id, reporter_id,
			priority, workspace_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $8, $9, $9)
	`, id, sequence, teamID, title, nullableStatus, assigneeID, priority, workspaceID, createdAt)
}

func insertSearchObjective(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, teamID, workspaceID, statusID uuid.UUID, name string, createdAt time.Time) {
	t.Helper()
	mustSearchExec(t, ctx, pool, `
		INSERT INTO objectives (
			objective_id, sequence_id, name, team_id, workspace_id, status_id, priority, health, created_at, updated_at
		) VALUES ($1, 1, $2, $3, $4, $5, 'high', 'On Track', $6, $6)
	`, id, name, teamID, workspaceID, statusID, createdAt)
}

func mustSearchExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, statement string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed search fixture: %v", err)
	}
}
