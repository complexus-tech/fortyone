//go:build integration

package nativepreviewrepository

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestPreviewReadEnforcesLiveResourcePermissions(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repository := New(postgres.Pool)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := postgres.Pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	workspace, otherWorkspace, team, hiddenTeam, actorID, ownerID, guestID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actorID, ownerID, guestID} {
		exec(`INSERT INTO users(user_id,username,email,is_active) VALUES($1,$2,$3,TRUE)`, id, id.String(), id.String()+"@example.com")
	}
	for _, id := range []uuid.UUID{workspace, otherWorkspace} {
		exec(`INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Preview test',$2)`, id, "preview-"+id.String())
	}
	for _, id := range []uuid.UUID{actorID, ownerID} {
		exec(`INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, workspace, id)
	}
	exec(`INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'guest')`, workspace, guestID)
	for i, id := range []uuid.UUID{team, hiddenTeam} {
		exec(`INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,$3,$3,'#123456')`, id, workspace, []string{"PRV", "HID"}[i])
	}
	for _, id := range []uuid.UUID{actorID, guestID} {
		exec(`INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, id)
	}
	storyID, hiddenStoryID, objectiveID := uuid.New(), uuid.New(), uuid.New()
	exec(`INSERT INTO stories(id,workspace_id,team_id,reporter_id,title,sequence_id) VALUES($1,$2,$3,$4,'Visible preview',1),($5,$2,$6,$4,'Hidden preview',1)`, storyID, workspace, team, ownerID, hiddenStoryID, hiddenTeam)
	exec(`INSERT INTO objectives(objective_id,workspace_id,team_id,name,created_by,sequence_id) VALUES($1,$2,$3,'Preview objective',$4,1)`, objectiveID, workspace, team, ownerID)
	documents := map[string]uuid.UUID{}
	for _, visibility := range []string{"workspace", "restricted", "private"} {
		id := uuid.New()
		documents[visibility] = id
		exec(`INSERT INTO documents(document_id,workspace_id,title,content_text,visibility,created_by,updated_by) VALUES($1,$2,'Preview document',$3,$4,$5,$5)`, id, workspace, strings.Repeat("text ", 200), visibility, ownerID)
	}
	for _, visibility := range []string{"restricted", "private"} {
		exec(`INSERT INTO document_members(document_id,user_id,role) VALUES($1,$2,'viewer')`, documents[visibility], actorID)
	}
	actor := auth.NewHumanActor(actorID)
	guest := auth.NewHumanActor(guestID)
	owner := auth.NewHumanActor(ownerID)
	story := domain.Target{Kind: "story", ID: storyID}
	objective := domain.Target{Kind: "objective", ID: objectiveID, TeamID: team}
	read := func(name string, who auth.Actor, tenant uuid.UUID, target domain.Target, allowed bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			for _, payload := range []bool{false, true} {
				snapshot, err := repository.Read(ctx, who, tenant, target, payload)
				if !allowed {
					if !errors.Is(err, domain.ErrNotFound) {
						t.Fatalf("payload=%v err=%v", payload, err)
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				if snapshot.ID == uuid.Nil || snapshot.Version == "" {
					t.Fatalf("invalid snapshot=%+v", snapshot)
				}
				if payload {
					var value map[string]any
					if err := json.Unmarshal(snapshot.Value, &value); err != nil {
						t.Fatal(err)
					}
					if _, ok := value["contentHtml"]; ok {
						t.Fatal("preview includes document body")
					}
					if target.Kind == "document" && len(value["contentText"].(string)) > 400 {
						t.Fatal("description not bounded")
					}
				} else if len(snapshot.Value) > 0 {
					t.Fatal("authorization read returned content")
				}
			}
		})
	}
	read("visible story", actor, workspace, story, true)
	read("reference", actor, workspace, domain.Target{Kind: "story", TeamCode: "prv", Sequence: 1}, true)
	read("hidden team", actor, workspace, domain.Target{Kind: "story", ID: hiddenStoryID}, false)
	read("foreign workspace", actor, otherWorkspace, story, false)
	read("objective member", actor, workspace, objective, true)
	read("objective guest", guest, workspace, objective, false)
	read("objective wrong team", actor, workspace, domain.Target{Kind: "objective", ID: objectiveID, TeamID: hiddenTeam}, false)
	restricted, _ := auth.RestrictedTeamAccess(hiddenTeam)
	scopedActor := actor
	scopedActor.TeamAccess = restricted
	read("credential team scope", scopedActor, workspace, story, false)
	read("workspace document guest", guest, workspace, domain.Target{Kind: "document", ID: documents["workspace"]}, true)
	read("restricted document grant", actor, workspace, domain.Target{Kind: "document", ID: documents["restricted"]}, true)
	read("private stale grant", actor, workspace, domain.Target{Kind: "document", ID: documents["private"]}, false)
	read("private owner", owner, workspace, domain.Target{Kind: "document", ID: documents["private"]}, true)
	exec(`DELETE FROM document_members WHERE document_id=$1 AND user_id=$2`, documents["restricted"], actorID)
	read("revoked document grant", actor, workspace, domain.Target{Kind: "document", ID: documents["restricted"]}, false)
	exec(`DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, team, actorID)
	read("revoked team membership", actor, workspace, story, false)
	exec(`INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, actorID)
	before, err := repository.Read(ctx, actor, workspace, story, false)
	if err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE stories SET title='Changed title',updated_at=$2 WHERE id=$1`, storyID, time.Now().Add(time.Second))
	after, err := repository.Read(ctx, actor, workspace, story, true)
	if err != nil || after.Version == before.Version || !strings.Contains(string(after.Value), "Changed title") {
		t.Fatalf("revision did not change: %v %s", err, after.Value)
	}
	documentTarget := domain.Target{Kind: "document", ID: documents["workspace"]}
	documentBefore, err := repository.Read(ctx, actor, workspace, documentTarget, false)
	if err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE documents SET title='Changed document title',content_text='Changed document body',updated_at=$2 WHERE document_id=$1`, documentTarget.ID, time.Now().Add(2*time.Second))
	documentAfter, err := repository.Read(ctx, actor, workspace, documentTarget, true)
	if err != nil || documentAfter.Version == documentBefore.Version || !strings.Contains(string(documentAfter.Value), "Changed document title") || !strings.Contains(string(documentAfter.Value), "Changed document body") {
		t.Fatalf("document revision did not change: %v %s", err, documentAfter.Value)
	}
	exec(`UPDATE stories SET deleted_at=NOW() WHERE id=$1`, storyID)
	read("deleted story", actor, workspace, story, false)
	exec(`UPDATE users SET is_active=FALSE WHERE user_id=$1`, actorID)
	read("inactive user document", actor, workspace, domain.Target{Kind: "document", ID: documents["workspace"]}, false)
	exec(`UPDATE users SET is_active=TRUE WHERE user_id=$1`, actorID)
	exec(`DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, workspace, actorID)
	read("revoked workspace membership", actor, workspace, domain.Target{Kind: "document", ID: documents["workspace"]}, false)
}
