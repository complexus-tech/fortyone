//go:build integration

package statesrepository

import (
	"errors"
	"testing"

	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
)

func TestWIPLimitCountsActiveWorkAndRequiresTeamWriteAccess(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace := insertStateWorkspace(t, ctx, postgres.Pool, "wip")
	actor := insertStateUser(t, ctx, postgres.Pool, "wip-member", true)
	guest := insertStateUser(t, ctx, postgres.Pool, "wip-guest", true)
	unjoined := insertStateUser(t, ctx, postgres.Pool, "wip-unjoined", true)
	insertStateWorkspaceMember(t, ctx, postgres.Pool, workspace, actor)
	insertStateWorkspaceMember(t, ctx, postgres.Pool, workspace, guest)
	insertStateWorkspaceMember(t, ctx, postgres.Pool, workspace, unjoined)
	if _, err := postgres.Pool.Exec(ctx, `UPDATE workspace_members SET role = 'guest' WHERE workspace_id = $1 AND user_id = $2`, workspace, guest); err != nil {
		t.Fatal(err)
	}
	team := insertStateTeam(t, ctx, postgres.Pool, workspace, "WIP")
	insertStateTeamMember(t, ctx, postgres.Pool, team, actor)
	insertStateTeamMember(t, ctx, postgres.Pool, team, guest)
	repository := New(postgres.Pool)
	state, err := repository.Create(ctx, actor, workspace, statesdomain.NewState{Name: "In progress", Category: "started", Team: team, Color: "#123456"})
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if _, err = repository.Update(ctx, actor, workspace, state.ID, statesdomain.UpdateState{WIPLimit: &limit}); err != nil {
		t.Fatal(err)
	}
	if _, err = repository.Update(ctx, guest, workspace, state.ID, statesdomain.UpdateState{WIPLimit: &limit}); !errors.Is(err, statesdomain.ErrNotFound) {
		t.Fatalf("guest WIP change = %v", err)
	}
	if _, err = repository.Update(ctx, unjoined, workspace, state.ID, statesdomain.UpdateState{WIPLimit: &limit}); !errors.Is(err, statesdomain.ErrNotFound) {
		t.Fatalf("unjoined WIP change = %v", err)
	}
	for _, invalid := range []int{-1, 10001} {
		if _, err = repository.Update(ctx, actor, workspace, state.ID, statesdomain.UpdateState{WIPLimit: &invalid}); !errors.Is(err, statesdomain.ErrInvalidWIPLimit) {
			t.Fatalf("invalid WIP %d = %v", invalid, err)
		}
	}
	for range 2 {
		insertStateStory(t, ctx, postgres.Pool, workspace, team, state.ID)
	}
	if _, err = postgres.Pool.Exec(ctx, `INSERT INTO stories (team_id, workspace_id, status_id, title, archived_at, is_draft) VALUES ($1,$2,$3,'Archived',NOW(),FALSE), ($1,$2,$3,'Draft',NULL,TRUE)`, team, workspace, state.ID); err != nil {
		t.Fatal(err)
	}
	states, err := repository.TeamListForMember(ctx, workspace, team, actor)
	if err != nil || len(states) != 1 || states[0].WIPLimit == nil || *states[0].WIPLimit != 1 || states[0].ActiveCount != 2 {
		t.Fatalf("WIP snapshot = %#v, %v", states, err)
	}
	clear := 0
	if _, err = repository.Update(ctx, actor, workspace, state.ID, statesdomain.UpdateState{WIPLimit: &clear}); err != nil {
		t.Fatal(err)
	}
	states, err = repository.TeamListForMember(ctx, workspace, team, actor)
	if err != nil || len(states) != 1 || states[0].WIPLimit != nil || states[0].ActiveCount != 2 {
		t.Fatalf("cleared WIP = %#v, %v", states, err)
	}
}
