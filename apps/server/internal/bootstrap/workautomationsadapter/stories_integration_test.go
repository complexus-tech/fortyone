//go:build integration

package workautomationsadapter

import (
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/bootstrap/customfieldsadapter"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	automation "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func TestRecurringDraftUsesCanonicalIdempotentMutationAndInternalOutbox(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	pool := postgres.Pool
	workspace, team, owner := uuid.New(), uuid.New(), uuid.New()
	suffix := uuid.NewString()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(t.Context(), query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Recurrence',$2)`, workspace, "recurrence-"+suffix)
	exec(`INSERT INTO users(user_id,username,email,full_name) VALUES($1,$2,$3,'Owner')`, owner, owner.String(), owner.String()+"@example.com")
	exec(`INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, workspace, owner)
	exec(`INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Team','REC','#000000')`, team, workspace)
	exec(`INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, owner)
	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	backend := stories.New(log, storiesrepository.New(log, pool, storiesrepository.WithCustomFieldCreation(customfieldsadapter.CreationBinder)), nil, nil)
	adapter := New(backend)
	draft := automation.Draft{Title: "Weekly review", Description: "Review", DescriptionHTML: `<p>Review</p><ul data-type="taskList"><li data-checked="true">Completed once</li></ul>`, Priority: "High", Checklist: []string{"Check <logs>"}}
	key := "team-automation:" + uuid.NewString() + ":2026-10-01T07:00:00Z"
	first, err := adapter.Create(t.Context(), owner, workspace, team, draft, key)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := adapter.Create(t.Context(), owner, workspace, team, draft, key)
	if err != nil || retry != first {
		t.Fatalf("recovery duplicated task: %s %s %v", first, retry, err)
	}
	var count int
	var internal bool
	if err := pool.QueryRow(t.Context(), `SELECT count(*) FROM stories WHERE external_creation_key=$1`, key).Scan(&count); err != nil || count != 1 {
		t.Fatal("duplicate recurrence task persisted")
	}
	if err := pool.QueryRow(t.Context(), `SELECT bool_and(payload->>'_delivery'='internal_only') FROM story_mutation_events WHERE story_id=$1`, first).Scan(&internal); err != nil || !internal {
		t.Fatalf("missing internal transaction outbox: %v %v", internal, err)
	}
	snapshot, err := adapter.Get(t.Context(), owner, workspace, first)
	if err != nil {
		t.Fatal(err)
	}
	priority := "Urgent"
	if err := adapter.Update(t.Context(), owner, workspace, snapshot, automation.Actions{Priority: &priority}, "Team rule: Triage"); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(t.Context(), `SELECT bool_and(payload->>'_delivery'='internal_only') FROM story_mutation_events WHERE story_id=$1`, first).Scan(&internal); err != nil || !internal {
		t.Fatal("rule output escaped internal event delivery")
	}
	exec(`DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, team, owner)
	if _, err := adapter.Create(t.Context(), owner, workspace, team, draft, key+"-revoked"); err == nil {
		t.Fatal("revoked owner created a recurrence task")
	}
	if err := adapter.Update(t.Context(), owner, workspace, snapshot, automation.Actions{Priority: &priority}, "Team rule: Triage"); err == nil {
		t.Fatal("revoked owner updated a task")
	}
}
