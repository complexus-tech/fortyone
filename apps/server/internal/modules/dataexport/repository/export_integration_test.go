//go:build integration

package dataexportrepository

import (
	"encoding/json"
	"errors"
	"testing"

	exportdomain "github.com/complexus-tech/projects-api/internal/modules/dataexport/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestExportPreservesExactWorkGraphAndFencesPrivateTeams(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	actor, member, workspace, team, privateTeam := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	story, child, hidden, draft, deleted, amount := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec := func(query string, arguments ...any) {
		t.Helper()
		if _, err := postgres.Pool.Exec(ctx, query, arguments...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users (user_id,username,email,full_name,is_active) VALUES ($1,$2,$3,'Export admin',TRUE),($4,$5,$6,'Member',TRUE)`, actor, actor.String(), actor.String()+"@example.test", member, member.String(), member.String()+"@example.test")
	exec(`INSERT INTO workspaces (workspace_id,name,slug) VALUES ($1,'Export', $2)`, workspace, workspace.String())
	exec(`INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'admin'),($1,$3,'member')`, workspace, actor, member)
	exec(`INSERT INTO teams (team_id,workspace_id,name,code,color,is_private) VALUES ($1,$3,'Visible','EXP','#123456',FALSE),($2,$3,'Private','HID','#123456',TRUE)`, team, privateTeam, workspace)
	exec(`INSERT INTO team_members (team_id,user_id) VALUES ($1,$2),($1,$3)`, team, actor, member)
	exec(`INSERT INTO stories (id,workspace_id,team_id,title,description,description_html,estimate_unit,archived_at) VALUES ($1,$2,$3,'=Customer deal','Full description','<p>Full description</p>',8,NOW())`, story, workspace, team)
	exec(`INSERT INTO stories (id,workspace_id,team_id,title,parent_id,blocked_by_id) VALUES ($1,$2,$3,'Child',$4,$4),($5,$2,$6,'Private task',NULL,NULL)`, child, workspace, team, story, hidden, privateTeam)
	exec(`INSERT INTO stories (id,workspace_id,team_id,title,is_draft,deleted_at) VALUES ($1,$3,$4,'Draft',TRUE,NULL),($2,$3,$4,'Deleted',FALSE,NOW())`, draft, deleted, workspace, team)
	exec(`INSERT INTO custom_fields (id,workspace_id,team_id,name,field_type,currency,icon) VALUES ($1,$2,$3,'Deal amount','money','USD','automation')`, amount, workspace, team)
	exec(`INSERT INTO story_custom_field_values (story_id,field_id,workspace_id,team_id,field_type,numeric_value) VALUES ($1,$2,$3,$4,'money',CAST('9007199254740993.01' AS numeric))`, story, amount, workspace, team)
	exec(`INSERT INTO story_comments (story_id,commenter_id,content) VALUES ($1,$2,'<p>Historical note</p>')`, story, actor)
	exec(`INSERT INTO story_links (story_id,url,title) VALUES ($1,'https://example.test/brief','Brief')`, story)
	repository := New(postgres.Pool)
	scope := exportdomain.Scope{ActorID: actor, WorkspaceID: workspace}
	snapshot, err := repository.Snapshot(ctx, scope)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.TaskCount != 2 {
		t.Fatalf("tasks = %d, want 2", snapshot.TaskCount)
	}
	var definitions []struct {
		Icon *string `json:"icon"`
	}
	if err := json.Unmarshal(snapshot.Envelope.CustomFields, &definitions); err != nil || len(definitions) != 1 || definitions[0].Icon == nil || *definitions[0].Icon != "automation" {
		t.Fatal("custom field icon was omitted from backup", err)
	}
	var graph struct {
		Teams []struct {
			SourceID uuid.UUID `json:"sourceId"`
		} `json:"teams"`
		Tasks []struct {
			SourceID     uuid.UUID `json:"sourceId"`
			Associations []struct {
				Type   string    `json:"type"`
				Target uuid.UUID `json:"targetSourceId"`
			} `json:"associations"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(snapshot.Envelope.Analysis, &graph); err != nil {
		t.Fatal(err)
	}
	if len(graph.Teams) != 1 || graph.Teams[0].SourceID != team || len(graph.Tasks) != 2 {
		t.Fatalf("private/draft/deleted data leaked: %#v", graph)
	}
	var childFound bool
	for _, task := range graph.Tasks {
		if task.SourceID == child {
			childFound = len(task.Associations) == 1 && task.Associations[0].Type == "blocked_by" && task.Associations[0].Target == story
		}
	}
	if !childFound {
		t.Fatal("legacy blocker relationship was not retained")
	}
	var data []struct {
		SourceID    uuid.UUID         `json:"sourceId"`
		Estimate    int               `json:"estimateValue"`
		Description string            `json:"descriptionHTML"`
		Comments    []json.RawMessage `json:"comments"`
		Values      []struct {
			Value string `json:"value"`
		} `json:"customFieldValues"`
	}
	if err := json.Unmarshal(snapshot.Envelope.TaskData, &data); err != nil {
		t.Fatal(err)
	}
	var preserved bool
	for _, task := range data {
		if task.SourceID == story {
			preserved = task.Estimate == 8 && task.Description == "<p>Full description</p>" && len(task.Comments) == 1 && len(task.Values) == 1 && task.Values[0].Value == "9007199254740993.01"
		}
	}
	if !preserved {
		t.Fatalf("canonical task data changed: %s", snapshot.Envelope.TaskData)
	}
	scope.ActorID = member
	if _, err := repository.Snapshot(ctx, scope); !errors.Is(err, exportdomain.ErrForbidden) {
		t.Fatalf("member export = %v", err)
	}
	scope.ActorID, scope.TeamID = actor, &privateTeam
	if _, err := repository.Snapshot(ctx, scope); !errors.Is(err, exportdomain.ErrForbidden) {
		t.Fatalf("private-team export = %v", err)
	}
}
