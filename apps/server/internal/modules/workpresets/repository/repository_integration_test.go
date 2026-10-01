//go:build integration

package workpresetsrepository

import (
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type presetFixture struct {
	pool                                                                 *pgxpool.Pool
	repo                                                                 *Repository
	workspace, otherWorkspace, team, owner, peer, guest, admin, outsider uuid.UUID
}

func newPresetFixture(t *testing.T) presetFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := presetFixture{pool: postgres.Pool, repo: New(postgres.Pool), workspace: uuid.New(), otherWorkspace: uuid.New(), team: uuid.New(), owner: uuid.New(), peer: uuid.New(), guest: uuid.New(), admin: uuid.New(), outsider: uuid.New()}
	suffix := uuid.NewString()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := f.pool.Exec(t.Context(), query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO workspaces (workspace_id,name,slug) VALUES ($1,'Preset workspace',$2),($3,'Other workspace',$4)`, f.workspace, "presets-"+suffix, f.otherWorkspace, "presets-other-"+suffix)
	exec(`INSERT INTO teams (team_id,workspace_id,name,code,color,is_private) VALUES ($1,$2,'Private team','PRE','#000000',TRUE)`, f.team, f.workspace)
	for _, actor := range []struct {
		id     uuid.UUID
		role   string
		joined bool
	}{{f.owner, "member", true}, {f.peer, "member", true}, {f.guest, "guest", true}, {f.admin, "admin", false}, {f.outsider, "member", false}} {
		exec(`INSERT INTO users (user_id,username,email,full_name) VALUES ($1,$2,$3,'Preset actor')`, actor.id, actor.id.String(), actor.id.String()+"@example.com")
		exec(`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,CAST($3 AS user_role))`, f.workspace, actor.id, actor.role)
		if actor.joined {
			exec(`INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)`, f.team, actor.id)
		}
	}
	exec(`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'admin')`, f.otherWorkspace, f.owner)
	return f
}

func (f presetFixture) create(t *testing.T, visibility domain.Visibility) domain.Preset {
	t.Helper()
	item, err := f.repo.Create(t.Context(), f.owner, f.workspace, domain.Input{TeamID: f.team, Kind: domain.Template, Visibility: visibility, Name: "Reusable", Configuration: []byte(`{"version":1,"title":"Review","description":"Steps","descriptionHTML":"<p>Steps</p>","priority":"High","checklist":[]}`)})
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func TestPresetsHonorPersonalTeamAndCurrentActorAuthority(t *testing.T) {
	f := newPresetFixture(t)
	personal, shared := f.create(t, domain.Personal), f.create(t, domain.Team)
	list := func(actor uuid.UUID) []domain.Preset {
		t.Helper()
		items, err := f.repo.List(t.Context(), actor, f.workspace, domain.List{TeamID: f.team, Kind: domain.Template, Limit: 101})
		if err != nil {
			t.Fatal(err)
		}
		return items
	}
	if len(list(f.owner)) != 2 {
		t.Fatal("owner cannot read own presets")
	}
	for _, actor := range []uuid.UUID{f.peer, f.guest, f.admin} {
		items := list(actor)
		if len(items) != 1 || items[0].ID != shared.ID {
			t.Fatal("personal preset leaked or shared preset missing")
		}
	}
	if list(f.guest)[0].CanEdit {
		t.Fatal("guest received edit capability")
	}
	if len(list(f.outsider)) != 0 {
		t.Fatal("private team preset leaked to nonmember")
	}
	for _, actor := range []uuid.UUID{f.peer, f.guest, f.outsider} {
		if _, err := f.repo.Update(t.Context(), actor, f.workspace, shared.ID, domain.Update{Name: "Denied"}); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("unauthorized rename: %v", err)
		}
		if err := f.repo.Archive(t.Context(), actor, f.workspace, shared.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("unauthorized archive: %v", err)
		}
	}
	if _, err := f.repo.Update(t.Context(), f.admin, f.workspace, shared.ID, domain.Update{Name: "Shared review"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.repo.Update(t.Context(), f.admin, f.workspace, personal.ID, domain.Update{Name: "Denied"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("admin exposed a personal preset")
	}
	if _, err := f.repo.Create(t.Context(), f.owner, f.otherWorkspace, domain.Input{TeamID: f.team, Kind: domain.View, Visibility: domain.Team, Name: "Wrong tenant", Configuration: []byte(`{}`)}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("cross tenant create: %v", err)
	}
	if err := f.repo.Archive(t.Context(), f.owner, f.workspace, shared.ID); err != nil {
		t.Fatal(err)
	}
	if len(list(f.owner)) != 1 {
		t.Fatal("archive failed to remove preset from active list")
	}
	var archived bool
	if err := f.pool.QueryRow(t.Context(), `SELECT archived_at IS NOT NULL FROM work_presets WHERE id=$1`, shared.ID).Scan(&archived); err != nil || !archived {
		t.Fatal("archive destroyed or failed to retain the preset")
	}
	if _, err := f.pool.Exec(t.Context(), `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.owner); err != nil {
		t.Fatal(err)
	}
	if len(list(f.owner)) != 0 {
		t.Fatal("removed team member retained access")
	}
	if _, err := f.repo.Update(t.Context(), f.owner, f.workspace, personal.ID, domain.Update{Name: "Denied"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("revoked mutation: %v", err)
	}
	if _, err := f.pool.Exec(t.Context(), `UPDATE users SET is_active=FALSE WHERE user_id=$1`, f.admin); err != nil {
		t.Fatal(err)
	}
	if len(list(f.admin)) != 0 {
		t.Fatal("inactive actor retained access")
	}
}

func TestPresetKeysetPaginationDoesNotRepeatItems(t *testing.T) {
	f := newPresetFixture(t)
	for range 3 {
		f.create(t, domain.Team)
	}
	first, err := f.repo.List(t.Context(), f.owner, f.workspace, domain.List{TeamID: f.team, Kind: domain.Template, Limit: 2})
	if err != nil || len(first) != 2 {
		t.Fatalf("first page: %v %v", first, err)
	}
	last := first[1]
	second, err := f.repo.List(t.Context(), f.owner, f.workspace, domain.List{TeamID: f.team, Kind: domain.Template, Limit: 2, Before: &domain.Position{CreatedAt: last.CreatedAt, ID: last.ID}})
	if err != nil || len(second) != 1 || second[0].ID == first[0].ID || second[0].ID == last.ID {
		t.Fatalf("second page: %v %v", second, err)
	}
}
