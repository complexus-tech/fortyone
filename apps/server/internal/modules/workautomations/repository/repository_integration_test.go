//go:build integration

package workautomationsrepository

import (
	"errors"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type automationFixture struct {
	pool                                                                 *pgxpool.Pool
	repo                                                                 *Repository
	workspace, otherWorkspace, team, owner, peer, guest, admin, outsider uuid.UUID
}

func newAutomationFixture(t *testing.T) automationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := automationFixture{pool: postgres.Pool, repo: New(postgres.Pool), workspace: uuid.New(), otherWorkspace: uuid.New(), team: uuid.New(), owner: uuid.New(), peer: uuid.New(), guest: uuid.New(), admin: uuid.New(), outsider: uuid.New()}
	suffix := uuid.NewString()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := f.pool.Exec(t.Context(), query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO workspaces (workspace_id,name,slug) VALUES ($1,'Automation workspace',$2),($3,'Other workspace',$4)`, f.workspace, "automations-"+suffix, f.otherWorkspace, "automations-other-"+suffix)
	exec(`INSERT INTO teams (team_id,workspace_id,name,code,color,is_private) VALUES ($1,$2,'Private team','PRE','#000000',TRUE)`, f.team, f.workspace)
	for _, actor := range []struct {
		id     uuid.UUID
		role   string
		joined bool
	}{{f.owner, "member", true}, {f.peer, "member", true}, {f.guest, "guest", true}, {f.admin, "admin", false}, {f.outsider, "member", false}} {
		exec(`INSERT INTO users (user_id,username,email,full_name) VALUES ($1,$2,$3,'Automation actor')`, actor.id, actor.id.String(), actor.id.String()+"@example.com")
		exec(`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,CAST($3 AS user_role))`, f.workspace, actor.id, actor.role)
		if actor.joined {
			exec(`INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)`, f.team, actor.id)
		}
	}
	exec(`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'admin')`, f.otherWorkspace, f.owner)
	return f
}

func (f automationFixture) create(t *testing.T) domain.Automation {
	t.Helper()
	item, err := f.repo.Create(t.Context(), f.owner, f.workspace, domain.Input{TeamID: f.team, Kind: "rule", Name: "Triage", Configuration: []byte(`{"version":1,"trigger":"story.created","conditions":{},"actions":{"priority":"High"}}`)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return item
}
func TestAutomationPermissionsPauseAndArchiveUseCurrentAuthority(t *testing.T) {
	f := newAutomationFixture(t)
	item := f.create(t)
	for _, actor := range []uuid.UUID{f.owner, f.peer, f.guest, f.admin} {
		items, err := f.repo.List(t.Context(), actor, f.workspace, f.team)
		if err != nil || len(items) != 1 {
			t.Fatalf("authorized read failed: %v %+v", err, items)
		}
		if (actor == f.peer || actor == f.guest) && items[0].CanEdit {
			t.Fatal("nonowner received edit authority")
		}
	}
	items, err := f.repo.List(t.Context(), f.outsider, f.workspace, f.team)
	if err != nil || len(items) != 0 {
		t.Fatal("private team automation leaked")
	}
	for _, actor := range []uuid.UUID{f.peer, f.guest, f.outsider} {
		if _, err := f.repo.Pause(t.Context(), actor, f.workspace, item.ID, true, nil); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("unauthorized pause: %v", err)
		}
		if err := f.repo.Archive(t.Context(), actor, f.workspace, item.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("unauthorized archive: %v", err)
		}
	}
	if _, err := f.repo.Create(t.Context(), f.guest, f.workspace, domain.Input{TeamID: f.team, Kind: "rule", Name: "Denied", Configuration: []byte(`{}`)}, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("guest create: %v", err)
	}
	if _, err := f.repo.Create(t.Context(), f.owner, f.otherWorkspace, domain.Input{TeamID: f.team, Kind: "rule", Name: "Denied", Configuration: []byte(`{}`)}, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cross-tenant create: %v", err)
	}
	claim, found, err := f.repo.Claim(t.Context())
	if err != nil || !found {
		t.Fatal("active automation not claimed")
	}
	allowed, err := f.repo.Authorized(t.Context(), claim)
	if err != nil || !allowed {
		t.Fatal("owner should be authorized")
	}
	if _, err := f.repo.Pause(t.Context(), f.admin, f.workspace, item.ID, true, nil); err != nil {
		t.Fatal(err)
	}
	allowed, err = f.repo.Authorized(t.Context(), claim)
	if err != nil || allowed {
		t.Fatal("paused automation retained live claim")
	}
	if _, found, err := f.repo.Claim(t.Context()); err != nil || found {
		t.Fatal("paused automation was claimed")
	}
	if _, err := f.repo.Pause(t.Context(), f.owner, f.workspace, item.ID, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.Archive(t.Context(), f.admin, f.workspace, item.ID); err != nil {
		t.Fatal(err)
	}
	var archived bool
	if err := f.pool.QueryRow(t.Context(), `SELECT archived_at IS NOT NULL FROM team_automations WHERE id=$1`, item.ID).Scan(&archived); err != nil || !archived {
		t.Fatal("archive did not preserve automation")
	}
}

func TestAutomationClaimsRecoverWithoutReplayingCompletedOrMissingLateEvents(t *testing.T) {
	f := newAutomationFixture(t)
	item := f.create(t)
	story := uuid.New()
	if _, err := f.pool.Exec(t.Context(), `INSERT INTO stories(id,workspace_id,team_id,title,reporter_id,priority) VALUES($1,$2,$3,'Task',$4,'Low')`, story, f.workspace, f.team, f.owner); err != nil {
		t.Fatal(err)
	}
	insertEvent := func(at time.Time) uuid.UUID {
		t.Helper()
		id := uuid.New()
		if _, err := f.pool.Exec(t.Context(), `INSERT INTO outbound_webhook_events(event_id,workspace_id,event_type,subject_type,subject_id,actor_kind,actor_id,payload,occurred_at,created_at) VALUES($1,$2,'story.created','story',$3,'human_user',$4,'{}',$5,$5)`, id, f.workspace, story, f.owner, at); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first := insertEvent(item.CreatedAt.Add(2 * time.Second))
	claim, found, err := f.repo.Claim(t.Context())
	if err != nil || !found {
		t.Fatal("missing claim")
	}
	if _, found, err := New(f.pool).Claim(t.Context()); err != nil || found {
		t.Fatal("concurrent worker obtained active lease")
	}
	events, err := f.repo.Events(t.Context(), claim)
	if err != nil || len(events) != 1 || events[0].ID != first {
		t.Fatalf("missing event: %v %+v", err, events)
	}
	run, found, err := f.repo.ClaimRun(t.Context(), claim, "event:"+first.String())
	if err != nil || !found {
		t.Fatalf("missing durable run: %v found=%v",err,found)
	}
	if _, found, err := f.repo.ClaimRun(t.Context(), claim, "event:"+first.String()); err != nil || found {
		t.Fatal("same run claimed twice by one worker")
	}
	if err := f.repo.CompleteRun(t.Context(), claim, run, "succeeded", &story, ""); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.Advance(t.Context(), claim, events[0], ""); err != nil {
		t.Fatal(err)
	}
	late := insertEvent(item.CreatedAt.Add(time.Second))
	events, err = f.repo.Events(t.Context(), claim)
	if err != nil || len(events) != 1 || events[0].ID != late {
		t.Fatalf("late committing event skipped or completed replayed: %v %+v", err, events)
	}
	running, found, err := f.repo.ClaimRun(t.Context(), claim, "event:"+late.String())
	if err != nil || !found {
		t.Fatal("missing late event run")
	}
	if _, err := f.pool.Exec(t.Context(), `UPDATE team_automations SET lease_until=now()-interval '1 minute' WHERE id=$1`, item.ID); err != nil {
		t.Fatal(err)
	}
	recovered, found, err := New(f.pool).Claim(t.Context())
	if err != nil || !found || recovered.LeaseToken == claim.LeaseToken {
		t.Fatal("expired lease did not recover")
	}
	retry, found, err := f.repo.ClaimRun(t.Context(), recovered, "event:"+late.String())
	if err != nil || !found || retry != running {
		t.Fatal("recovery did not preserve durable run identity")
	}
	if err := f.repo.CompleteRun(t.Context(), claim, running, "succeeded", &story, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("stale worker completed recovered run")
	}
	if err := f.repo.CompleteRun(t.Context(), recovered, running, "succeeded", &story, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.pool.Exec(t.Context(), `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.owner); err != nil {
		t.Fatal(err)
	}
	allowed, err := f.repo.Authorized(t.Context(), recovered)
	if err != nil || allowed {
		t.Fatal("revoked owner retained execution authority")
	}
}
