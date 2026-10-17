//go:build integration

package searchrepository

import (
	"context"
	"testing"
	"time"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestSearchStoriesMatchesExactDisplayIDWithinActorScope(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := seedSearchFixture(t, ctx, postgres.Pool)
	repository := New(postgres.Pool)
	mustSearchExec(t, ctx, postgres.Pool, "UPDATE teams SET code = 'PRO' WHERE team_id = $1", fixture.teamA)
	mustSearchExec(t, ctx, postgres.Pool, "UPDATE stories SET sequence_id = 98 WHERE id = $1", fixture.storyNewest)

	for _, query := range []string{"PRO-98", "pro-98", "  Pro-98  "} {
		t.Run(query, func(t *testing.T) {
			params := search.SearchParams{Query: query, SortBy: search.SortByRelevance, Page: 1, PageSize: 1}
			rows, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, params)
			if err != nil || len(rows) != 1 || total != 1 || rows[0].ID != fixture.storyNewest {
				t.Fatalf("display-ID lookup = %#v, total %d, error %v", rows, total, err)
			}
			// Empty pages use a separate count query with the same search predicate.
			params.Page = 2
			rows, total, err = repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, params)
			if err != nil || len(rows) != 0 || total != 1 {
				t.Fatalf("empty display-ID page = %#v, total %d, error %v", rows, total, err)
			}
		})
	}

	for _, test := range []struct {
		name      string
		query     string
		workspace uuid.UUID
		actor     uuid.UUID
		team      *uuid.UUID
	}{
		{name: "hidden team", query: "AH-1", workspace: fixture.workspaceA, actor: fixture.actor},
		{name: "foreign workspace", query: "PRO-98", workspace: fixture.workspaceB, actor: fixture.actor},
		{name: "inactive actor", query: "PRO-98", workspace: fixture.workspaceA, actor: fixture.inactiveUser},
		{name: "deleted story", query: "A2-1", workspace: fixture.workspaceA, actor: fixture.actor},
		{name: "team filter", query: "PRO-98", workspace: fixture.workspaceA, actor: fixture.actor, team: &fixture.teamA2},
		{name: "inexact sequence", query: "PRO-9", workspace: fixture.workspaceA, actor: fixture.actor},
	} {
		t.Run(test.name, func(t *testing.T) {
			rows, total, err := repository.SearchStories(ctx, test.workspace, test.actor, search.SearchParams{
				Query: test.query, TeamID: test.team, SortBy: search.SortByRelevance, Page: 1, PageSize: 20,
			})
			if err != nil || len(rows) != 0 || total != 0 {
				t.Fatalf("inaccessible or unmatched lookup = %#v, total %d, error %v", rows, total, err)
			}
		})
	}

	// A story that only mentions the ID in its title must not outrank the target.
	insertSearchStory(t, ctx, postgres.Pool, uuid.New(), fixture.teamA, fixture.workspaceA,
		fixture.statusA, fixture.actor, 99, "PRO-98", "medium", time.Now())
	rows, total, err := repository.SearchStories(ctx, fixture.workspaceA, fixture.actor, search.SearchParams{
		Query: "PRO-98", SortBy: search.SortByRelevance, Page: 1, PageSize: 1,
	})
	if err != nil || len(rows) != 1 || total != 2 || rows[0].ID != fixture.storyNewest {
		t.Fatalf("exact ID relevance = %#v, total %d, error %v", rows, total, err)
	}
}
