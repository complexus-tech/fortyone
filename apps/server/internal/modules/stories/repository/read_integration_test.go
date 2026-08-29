//go:build integration

package storiesrepository

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type storyReadFixture struct {
	workspaceA  uuid.UUID
	workspaceB  uuid.UUID
	teamA       uuid.UUID
	teamHidden  uuid.UUID
	teamB       uuid.UUID
	actor       uuid.UUID
	inactive    uuid.UUID
	visible     uuid.UUID
	visibleTwo  uuid.UUID
	visibleSub  uuid.UUID
	hidden      uuid.UUID
	crossTenant uuid.UUID
	foreignTeam uuid.UUID
	foreignUser uuid.UUID
}

func TestStoryReadRepositoryEnforcesActorTenantAndTeamVisibility(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	repository := New(nil, postgres.Pool)
	scope := unrestrictedStoryReadScope(fixture.actor, fixture.workspaceA)

	story, err := repository.GetVisibleStory(ctx, scope, fixture.visible)
	if err != nil {
		t.Fatalf("get visible story: %v", err)
	}
	if story.ID != fixture.visible || len(story.SubStories) != 1 || story.SubStories[0].ID != fixture.visibleSub {
		t.Fatalf("visible story projection = %#v", story)
	}
	if len(story.Associations) != 1 || story.Associations[0].Story.ID != fixture.visibleTwo {
		t.Fatalf("visible associations = %#v", story.Associations)
	}

	assertStoryNotVisible(t, ctx, repository, scope, fixture.hidden)
	assertStoryNotVisible(t, ctx, repository, scope, fixture.crossTenant)
	assertStoryNotVisible(t, ctx, repository, unrestrictedStoryReadScope(fixture.inactive, fixture.workspaceA), fixture.visible)

	byReference, err := repository.QueryVisibleStoryByRef(ctx, scope, "a", 1)
	if err != nil || byReference.ID != fixture.visible {
		t.Fatalf("story by reference = %#v, error %v", byReference, err)
	}
	if _, err := repository.QueryVisibleStoryByRef(ctx, scope, "hidden", 1); !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("hidden reference error = %v, want not found", err)
	}

	myStories, err := repository.ListMyVisibleStories(ctx, scope)
	if err != nil {
		t.Fatalf("list actor stories: %v", err)
	}
	if len(myStories) != 2 {
		t.Fatalf("actor stories = %#v, want two visible parent stories", myStories)
	}
	for _, item := range myStories {
		if item.Workspace != fixture.workspaceA || item.Team != fixture.teamA {
			t.Fatalf("actor story escaped scope: %#v", item)
		}
	}
}

func TestStoryCategoryReadHasStableBoundedPagesAndRestrictions(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	repository := New(nil, postgres.Pool)
	scope := unrestrictedStoryReadScope(fixture.actor, fixture.workspaceA)

	first, hasMore, err := repository.ListVisibleStoriesByCategory(ctx, scope, fixture.teamA, "started", 1, 1, false)
	if err != nil || len(first) != 1 || !hasMore {
		t.Fatalf("first category page = %#v, hasMore %v, error %v", first, hasMore, err)
	}
	second, hasMore, err := repository.ListVisibleStoriesByCategory(ctx, scope, fixture.teamA, "started", 2, 1, false)
	if err != nil || len(second) != 1 || hasMore {
		t.Fatalf("second category page = %#v, hasMore %v, error %v", second, hasMore, err)
	}
	if first[0].ID == second[0].ID {
		t.Fatalf("stable pages repeated story %s", first[0].ID)
	}
	repeated, _, err := repository.ListVisibleStoriesByCategory(ctx, scope, fixture.teamA, "started", 1, 1, false)
	if err != nil || repeated[0].ID != first[0].ID {
		t.Fatalf("repeated first page = %#v, error %v", repeated, err)
	}

	restricted := scope
	restricted.UnrestrictedTeamAccess = false
	restricted.AllowedTeamIDs = []uuid.UUID{fixture.teamHidden}
	items, hasMore, err := repository.ListVisibleStoriesByCategory(ctx, restricted, fixture.teamA, "started", 1, 20, false)
	if err != nil || len(items) != 0 || hasMore {
		t.Fatalf("credential-restricted category = %#v, hasMore %v, error %v", items, hasMore, err)
	}
	items, _, err = repository.ListVisibleStoriesByCategory(
		ctx,
		unrestrictedStoryReadScope(fixture.inactive, fixture.workspaceA),
		fixture.teamA,
		"started",
		1,
		20,
		false,
	)
	if err != nil || len(items) != 0 {
		t.Fatalf("inactive actor category = %#v, error %v", items, err)
	}
}

func TestFilteredGroupedAndCountReadsShareVisibilityBoundary(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	assertStoryReadPostgres18(t, ctx, postgres.Pool)
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	repository := New(nil, postgres.Pool)
	scope := unrestrictedStoryReadScope(fixture.actor, fixture.workspaceA)

	items, err := repository.ListVisibleStories(ctx, scope, storydomain.StoryFilters{})
	if err != nil {
		t.Fatalf("list visible stories: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("visible story list = %#v, want two top-level stories", items)
	}
	for _, item := range items {
		if item.Workspace != fixture.workspaceA || item.Team != fixture.teamA {
			t.Fatalf("filtered story escaped visibility: %#v", item)
		}
	}

	count, err := repository.CountVisibleStories(ctx, scope)
	if err != nil || count != 3 {
		t.Fatalf("visible count = %d, error %v, want three including child", count, err)
	}
	count, err = repository.CountVisibleStories(ctx, unrestrictedStoryReadScope(fixture.inactive, fixture.workspaceA))
	if err != nil || count != 0 {
		t.Fatalf("inactive actor count = %d, error %v", count, err)
	}

	query := storydomain.StoryQuery{
		Filters:         storydomain.StoryFilters{Categories: []string{"started"}},
		GroupBy:         storydomain.StoryGroupStatus,
		OrderBy:         storydomain.StoryOrderCreated,
		OrderDirection:  storydomain.SortDescending,
		StoriesPerGroup: 1,
		Page:            1,
		PageSize:        1,
	}
	groups, err := repository.ListVisibleGroupedStories(ctx, scope, query)
	if err != nil {
		t.Fatalf("list visible groups: %v", err)
	}
	if len(groups) != 1 || groups[0].TotalCount != 2 || groups[0].LoadedCount != 1 || !groups[0].HasMore {
		t.Fatalf("visible groups = %#v", groups)
	}
	second, hasMore, err := repository.ListVisibleGroupStories(ctx, scope, groups[0].Key, storydomain.StoryQuery{
		Filters:         query.Filters,
		GroupBy:         query.GroupBy,
		OrderBy:         query.OrderBy,
		OrderDirection:  query.OrderDirection,
		Page:            2,
		PageSize:        1,
		StoriesPerGroup: 1,
	})
	if err != nil || len(second) != 1 || hasMore || second[0].ID == groups[0].Stories[0].ID {
		t.Fatalf("second group page = %#v, hasMore %v, error %v", second, hasMore, err)
	}
	assigneeQuery := query
	assigneeQuery.GroupBy = storydomain.StoryGroupAssignee
	assigneeGroups, err := repository.ListVisibleGroupedStories(ctx, scope, assigneeQuery)
	if err != nil {
		t.Fatalf("list visible assignee groups: %v", err)
	}
	for _, group := range assigneeGroups {
		if group.Key == fixture.foreignUser.String() {
			t.Fatalf("assignee catalog leaked non-workspace member %s", fixture.foreignUser)
		}
	}

	restricted := scope
	restricted.UnrestrictedTeamAccess = false
	restricted.AllowedTeamIDs = []uuid.UUID{fixture.teamHidden}
	items, err = repository.ListVisibleStories(ctx, restricted, storydomain.StoryFilters{})
	if err != nil || len(items) != 0 {
		t.Fatalf("credential-restricted list = %#v, error %v", items, err)
	}
	groups, err = repository.ListVisibleGroupedStories(ctx, restricted, query)
	if err != nil || len(groups) != 0 {
		t.Fatalf("credential-restricted groups = %#v, error %v", groups, err)
	}
}

func unrestrictedStoryReadScope(actorID, workspaceID uuid.UUID) storydomain.ReadScope {
	return storydomain.ReadScope{
		ActorID:                actorID,
		WorkspaceID:            workspaceID,
		UnrestrictedTeamAccess: true,
		AllowedTeamIDs:         []uuid.UUID{},
	}
}

func assertStoryNotVisible(t *testing.T, ctx context.Context, repository *repo, scope storydomain.ReadScope, storyID uuid.UUID) {
	t.Helper()
	_, err := repository.GetVisibleStory(ctx, scope, storyID)
	if !errors.Is(err, storydomain.ErrNotFound) {
		t.Fatalf("story %s error = %v, want not found", storyID, err)
	}
}

func assertStoryReadPostgres18(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var rawVersion string
	if err := pool.QueryRow(ctx, "SHOW server_version_num").Scan(&rawVersion); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(rawVersion)
	if err != nil || version < 180000 || version >= 190000 {
		t.Fatalf("story read integration tests require PostgreSQL 18, got %q", rawVersion)
	}
}

func seedStoryReadFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) storyReadFixture {
	t.Helper()
	fixture := storyReadFixture{
		workspaceA: uuid.New(), workspaceB: uuid.New(), teamA: uuid.New(), teamHidden: uuid.New(), teamB: uuid.New(),
		actor: uuid.New(), inactive: uuid.New(), visible: uuid.New(), visibleTwo: uuid.New(),
		visibleSub: uuid.New(), hidden: uuid.New(), crossTenant: uuid.New(), foreignTeam: uuid.New(), foreignUser: uuid.New(),
	}
	insertStoryReadUser(t, ctx, pool, fixture.actor, true)
	insertStoryReadUser(t, ctx, pool, fixture.inactive, false)
	insertStoryReadUser(t, ctx, pool, fixture.foreignUser, true)
	insertStoryReadWorkspace(t, ctx, pool, fixture.workspaceA, fixture.actor, "a")
	insertStoryReadWorkspace(t, ctx, pool, fixture.workspaceB, fixture.foreignUser, "b")
	for _, userID := range []uuid.UUID{fixture.actor, fixture.inactive} {
		mustStoryReadExec(t, ctx, pool, "INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')", fixture.workspaceA, userID)
	}
	mustStoryReadExec(t, ctx, pool, "INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')", fixture.workspaceB, fixture.foreignUser)
	insertStoryReadTeam(t, ctx, pool, fixture.teamA, fixture.workspaceA, "A")
	insertStoryReadTeam(t, ctx, pool, fixture.teamHidden, fixture.workspaceA, "HIDDEN")
	insertStoryReadTeam(t, ctx, pool, fixture.teamB, fixture.workspaceB, "B")
	for _, userID := range []uuid.UUID{fixture.actor, fixture.inactive} {
		mustStoryReadExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamA, userID)
	}
	// Deliberately malformed memberships and story rows prove that reads enforce
	// the workspace/team relationship instead of trusting single-column FKs.
	mustStoryReadExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamA, fixture.foreignUser)
	mustStoryReadExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamB, fixture.foreignUser)
	mustStoryReadExec(t, ctx, pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamB, fixture.actor)
	statusA := insertStoryReadStatus(t, ctx, pool, fixture.workspaceA, fixture.teamA)
	statusHidden := insertStoryReadStatus(t, ctx, pool, fixture.workspaceA, fixture.teamHidden)
	statusB := insertStoryReadStatus(t, ctx, pool, fixture.workspaceB, fixture.teamB)
	createdAt := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	insertStoryReadStory(t, ctx, pool, fixture.visible, fixture.workspaceA, fixture.teamA, statusA, fixture.actor, 1, nil, createdAt)
	insertStoryReadStory(t, ctx, pool, fixture.visibleTwo, fixture.workspaceA, fixture.teamA, statusA, fixture.actor, 2, nil, createdAt)
	insertStoryReadStory(t, ctx, pool, fixture.visibleSub, fixture.workspaceA, fixture.teamA, statusA, fixture.actor, 3, &fixture.visible, createdAt)
	insertStoryReadStory(t, ctx, pool, fixture.hidden, fixture.workspaceA, fixture.teamHidden, statusHidden, fixture.actor, 1, nil, createdAt)
	insertStoryReadStory(t, ctx, pool, fixture.crossTenant, fixture.workspaceB, fixture.teamB, statusB, fixture.foreignUser, 1, nil, createdAt)
	insertStoryReadStory(t, ctx, pool, fixture.foreignTeam, fixture.workspaceA, fixture.teamB, statusB, fixture.actor, 2, nil, createdAt)
	mustStoryReadExec(t, ctx, pool, `INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id) VALUES ($1, $2, 'related', $3)`, fixture.visible, fixture.visibleTwo, fixture.workspaceA)
	mustStoryReadExec(t, ctx, pool, `INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id) VALUES ($1, $2, 'related', $3)`, fixture.visible, fixture.hidden, fixture.workspaceA)
	mustStoryReadExec(t, ctx, pool, `INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id) VALUES ($1, $2, 'related', $3)`, fixture.visible, fixture.foreignTeam, fixture.workspaceA)
	return fixture
}

func insertStoryReadUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustStoryReadExec(t, ctx, pool, `INSERT INTO users (user_id, username, email, is_active) VALUES ($1, $2, $3, $4)`, id, "story-"+suffix, "story-"+suffix+"@example.com", active)
}

func insertStoryReadWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, creator uuid.UUID, label string) {
	t.Helper()
	mustStoryReadExec(t, ctx, pool, `INSERT INTO workspaces (workspace_id, name, slug, created_by) VALUES ($1, $2, $3, $4)`, id, "Story "+label, "story-"+label+"-"+uuid.NewString(), creator)
}

func insertStoryReadTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustStoryReadExec(t, ctx, pool, `INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')`, id, "Story "+code, workspaceID, code)
}

func insertStoryReadStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, teamID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	mustStoryReadExec(t, ctx, pool, `INSERT INTO statuses (status_id, name, category, workspace_id, team_id) VALUES ($1, 'Started', 'started', $2, $3)`, id, workspaceID, teamID)
	return id
}

func insertStoryReadStory(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID, teamID, statusID, reporterID uuid.UUID, sequence int, parentID *uuid.UUID, createdAt time.Time) {
	t.Helper()
	mustStoryReadExec(t, ctx, pool, `INSERT INTO stories (id, workspace_id, team_id, status_id, reporter_id, sequence_id, title, priority, parent_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, 'High', $8, $9, $9)`, id, workspaceID, teamID, statusID, reporterID, sequence, "Story "+id.String(), parentID, createdAt)
}

func mustStoryReadExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, statement string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, args...); err != nil {
		t.Fatalf("seed story read fixture: %v", err)
	}
}
