package storiesrepository

import (
	"encoding/json"
	"strings"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestBuildSubStoriesJSONExprUsesNestedTeamSummary(t *testing.T) {
	expression := (&repo{}).buildSubStoriesJSONExpr("parent_story", true)

	for _, expected := range []string{
		"'team', CASE",
		"'id', sub_team.team_id",
		"'name', sub_team.name",
		"'code', sub_team.code",
		"sub.workspace_id = parent_story.workspace_id",
	} {
		if !strings.Contains(expression, expected) {
			t.Fatalf("expected sub-story expression to contain %q, got %q", expected, expression)
		}
	}

	for _, legacyField := range []string{"'team_code'", "'team_name'"} {
		if strings.Contains(expression, legacyField) {
			t.Fatalf("expected sub-story expression not to contain legacy field %q", legacyField)
		}
	}
}

func TestCoreStoryListUnmarshalsNestedTeamSummary(t *testing.T) {
	teamID := uuid.New()
	payload := []byte(`{"team_id":"` + teamID.String() + `","team":{"id":"` + teamID.String() + `","name":"Web","code":"W"}}`)

	var story stories.CoreStoryList
	if err := json.Unmarshal(payload, &story); err != nil {
		t.Fatalf("unmarshal sub-story payload: %v", err)
	}

	if story.TeamSummary == nil {
		t.Fatal("expected nested team summary to be preserved")
	}
	if story.TeamSummary.ID != teamID || story.TeamSummary.Name != "Web" || story.TeamSummary.Code != "W" {
		t.Fatalf("unexpected team summary: %+v", story.TeamSummary)
	}
}

func TestMapToStoryListReadsNestedTeamSummary(t *testing.T) {
	teamID := uuid.New()
	story := (&repo{}).mapToStoryList(map[string]any{
		"team_id": teamID.String(),
		"team": map[string]any{
			"id":   teamID.String(),
			"name": "Web",
			"code": "W",
		},
	})

	if story.TeamSummary == nil {
		t.Fatal("expected nested team summary to be mapped")
	}
	if story.TeamSummary.ID != teamID || story.TeamSummary.Name != "Web" || story.TeamSummary.Code != "W" {
		t.Fatalf("unexpected team summary: %+v", story.TeamSummary)
	}
}
