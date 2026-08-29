package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func TestFortyOneToolExecutorCatalogIsSmallReadOnlyAndStrict(t *testing.T) {
	t.Parallel()

	executor := newToolExecutorForTest(t, &teamsServiceStub{}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})
	definitions := executor.Definitions()
	if len(definitions) != 4 {
		t.Fatalf("expected four tools, got %d", len(definitions))
	}
	if err := validateToolDefinitions(definitions); err != nil {
		t.Fatalf("catalog is not strict: %v", err)
	}
	expected := map[string]bool{
		toolListTeams:      false,
		toolListMyTasks:    false,
		toolSearchWork:     false,
		toolListObjectives: false,
	}
	for _, definition := range definitions {
		if _, ok := expected[definition.Name]; !ok {
			t.Fatalf("unexpected tool %q", definition.Name)
		}
		expected[definition.Name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("tool %q is missing", name)
		}
	}

	definitions[0].Parameters["additionalProperties"] = true
	if executor.Definitions()[0].Parameters["additionalProperties"] != false {
		t.Fatal("Definitions must return a defensive copy")
	}
}

func TestFortyOneToolDefinitionsSerializeEmptyRequiredAsArray(t *testing.T) {
	t.Parallel()

	executor := newToolExecutorForTest(t, &teamsServiceStub{}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})
	definitions := cloneToolDefinitions(executor.Definitions())
	if err := validateToolDefinitions(definitions); err != nil {
		t.Fatalf("validateToolDefinitions() error = %v", err)
	}

	var listTeams ToolDefinition
	for _, definition := range definitions {
		if definition.Name == toolListTeams {
			listTeams = definition
			break
		}
	}
	if listTeams.Name == "" {
		t.Fatal("list_teams definition is missing")
	}

	payload, err := json.Marshal(listTeams)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(payload), `"required":null`) {
		t.Fatalf("list_teams schema contains a null required value: %s", payload)
	}
	if !strings.Contains(string(payload), `"required":[]`) {
		t.Fatalf("list_teams schema does not contain an empty required array: %s", payload)
	}
}

func TestFortyOneToolExecutorListTeamsUsesJoinedOnlyScope(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	joinedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
		Name:      "Product",
		Code:      "PRO",
		Workspace: scope.WorkspaceID,
	}
	adminVisibleOnly := teams.CoreTeam{
		ID:        uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002"),
		Name:      "Admin only",
		Code:      "ADM",
		Workspace: scope.WorkspaceID,
	}
	teamsService := &teamsServiceStub{
		joined: []teams.CoreTeam{joinedTeam},
		all:    []teams.CoreTeam{joinedTeam, adminVisibleOnly},
	}
	executor := newToolExecutorForTest(t, teamsService, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeams, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("execute list teams: %v", err)
	}
	var result listTeamsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Total != 1 || len(result.Teams) != 1 || result.Teams[0].ID != joinedTeam.ID {
		t.Fatalf("expected only the joined team, got %#v", result)
	}
	call := teamsService.lastCall()
	if call.workspaceID != scope.WorkspaceID || call.userID != scope.UserID || !call.filter.JoinedOnly {
		t.Fatalf("team query was not strictly joined-only: %#v", call)
	}
}

func TestFortyOneToolExecutorIntersectsJoinedTeamsWithChannelScope(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	allowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000011"),
		Name:      "Public product",
		Workspace: scope.WorkspaceID,
	}
	disallowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000012"),
		Name:      "Private operations",
		Workspace: scope.WorkspaceID,
	}
	executor := newToolExecutorForTest(t, &teamsServiceStub{joined: []teams.CoreTeam{allowedTeam, disallowedTeam}}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})

	scope.AllowedTeamIDs = []uuid.UUID{allowedTeam.ID}
	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeams, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("execute scoped list teams: %v", err)
	}
	var result listTeamsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Total != 1 || len(result.Teams) != 1 || result.Teams[0].ID != allowedTeam.ID {
		t.Fatalf("channel team scope was not enforced: %#v", result)
	}

	scope.AllowedTeamIDs = []uuid.UUID{}
	raw, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeams, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("execute empty scoped list teams: %v", err)
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode empty result: %v", err)
	}
	if result.Total != 0 || len(result.Teams) != 0 {
		t.Fatalf("explicit empty channel scope must deny every team: %#v", result)
	}
}

func TestFortyOneToolExecutorMyTasksSetsActorAndFiltersUnjoinedTeams(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	joinedTeamID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	unjoinedTeamID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	foreignWorkspaceID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")
	currentUserID := scope.UserID
	teamsService := &teamsServiceStub{joined: []teams.CoreTeam{{
		ID:        joinedTeamID,
		Code:      "ENG",
		Workspace: scope.WorkspaceID,
	}}}
	storiesService := &storiesServiceStub{items: []stories.CoreStoryList{
		{ID: uuid.New(), SequenceID: 12, Title: "Joined", Team: joinedTeamID, Workspace: scope.WorkspaceID, Assignee: &currentUserID},
		{ID: uuid.New(), SequenceID: 13, Title: "Admin but unjoined", Team: unjoinedTeamID, Workspace: scope.WorkspaceID, Assignee: &currentUserID},
		{ID: uuid.New(), SequenceID: 14, Title: "Foreign workspace", Team: joinedTeamID, Workspace: foreignWorkspaceID, Assignee: &currentUserID},
	}}
	executor := newToolExecutorForTest(t, teamsService, storiesService, &searchServiceStub{}, &objectivesServiceStub{})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolListMyTasks,
		Arguments: json.RawMessage(`{"limit":null}`),
	})
	if err != nil {
		t.Fatalf("execute list my tasks: %v", err)
	}
	var result listTasksResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if result.Total != 1 || len(result.Tasks) != 1 || result.Tasks[0].Title != "Joined" || result.Tasks[0].Reference != "ENG-12" {
		t.Fatalf("unexpected scoped task result %#v", result)
	}
	call := storiesService.lastCall()
	if call.workspaceID != scope.WorkspaceID || call.actorID != scope.UserID || call.actorErr != nil {
		t.Fatalf("MyStories did not receive the authoritative actor context: %#v", call)
	}
}

func TestFortyOneToolExecutorRejectsSearchForUnjoinedTeam(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	joinedTeamID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	unjoinedTeamID := uuid.MustParse("cccccccc-0000-0000-0000-000000000002")
	teamsService := &teamsServiceStub{joined: []teams.CoreTeam{{
		ID:        joinedTeamID,
		Workspace: scope.WorkspaceID,
	}}}
	searchService := &searchServiceStub{}
	executor := newToolExecutorForTest(t, teamsService, &storiesServiceStub{}, searchService, &objectivesServiceStub{})

	arguments, _ := json.Marshal(map[string]any{
		"query":   "launch",
		"team_id": unjoinedTeamID.String(),
		"kind":    nil,
		"limit":   nil,
	})
	_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolSearchWork, Arguments: arguments})
	if !errors.Is(err, ErrTeamNotAccessible) {
		t.Fatalf("expected inaccessible team error, got %v", err)
	}
	if searchService.callCount() != 0 {
		t.Fatal("search must not run for an unjoined team")
	}
}

func TestFortyOneToolExecutorSearchAndObjectivesFilterDefensively(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	joinedTeamID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")
	unjoinedTeamID := uuid.MustParse("dddddddd-0000-0000-0000-000000000002")
	foreignWorkspaceID := uuid.MustParse("dddddddd-0000-0000-0000-000000000003")
	teamsService := &teamsServiceStub{joined: []teams.CoreTeam{{
		ID:        joinedTeamID,
		Code:      "OPS",
		Workspace: scope.WorkspaceID,
	}}}
	searchService := &searchServiceStub{result: search.CoreSearchResult{
		Stories: []search.CoreSearchStory{
			{ID: uuid.New(), SequenceID: 7, Title: "Allowed story", Team: joinedTeamID, Workspace: scope.WorkspaceID},
			{ID: uuid.New(), SequenceID: 8, Title: "Unjoined story", Team: unjoinedTeamID, Workspace: scope.WorkspaceID},
			{ID: uuid.New(), SequenceID: 9, Title: "Foreign story", Team: joinedTeamID, Workspace: foreignWorkspaceID},
		},
		Objectives: []search.CoreSearchObjective{
			{ID: uuid.New(), Name: "Allowed objective", Team: joinedTeamID, Workspace: scope.WorkspaceID},
			{ID: uuid.New(), Name: "Unjoined objective", Team: unjoinedTeamID, Workspace: scope.WorkspaceID},
		},
	}}
	objectivesService := &objectivesServiceStub{items: []objectives.CoreObjective{
		{ID: uuid.New(), Name: "Allowed objective", Team: joinedTeamID, Workspace: scope.WorkspaceID},
		{ID: uuid.New(), Name: "Unjoined objective", Team: unjoinedTeamID, Workspace: scope.WorkspaceID},
		{ID: uuid.New(), Name: "Foreign objective", Team: joinedTeamID, Workspace: foreignWorkspaceID},
	}}
	executor := newToolExecutorForTest(t, teamsService, &storiesServiceStub{}, searchService, objectivesService)

	searchOutput, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolSearchWork,
		Arguments: json.RawMessage(`{"query":"launch","team_id":null,"kind":"all","limit":20}`),
	})
	if err != nil {
		t.Fatalf("search work: %v", err)
	}
	var searchResult searchWorkResult
	if err := json.Unmarshal(searchOutput, &searchResult); err != nil {
		t.Fatalf("decode search result: %v", err)
	}
	if len(searchResult.Stories) != 1 || searchResult.Stories[0].Title != "Allowed story" || len(searchResult.Objectives) != 1 || searchResult.Objectives[0].Name != "Allowed objective" {
		t.Fatalf("search leaked data outside joined teams: %#v", searchResult)
	}
	searchCall := searchService.lastCall()
	if searchCall.workspaceID != scope.WorkspaceID || searchCall.userID != scope.UserID || searchCall.params.PageSize != 20 {
		t.Fatalf("search received incorrect scope %#v", searchCall)
	}

	objectiveOutput, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolListObjectives,
		Arguments: json.RawMessage(`{"team_id":null,"query":null,"limit":20}`),
	})
	if err != nil {
		t.Fatalf("list objectives: %v", err)
	}
	var objectiveResult listObjectivesResult
	if err := json.Unmarshal(objectiveOutput, &objectiveResult); err != nil {
		t.Fatalf("decode objectives result: %v", err)
	}
	if objectiveResult.Count != 1 || len(objectiveResult.Objectives) != 1 || objectiveResult.Objectives[0].Name != "Allowed objective" {
		t.Fatalf("objectives leaked data outside joined teams: %#v", objectiveResult)
	}
	objectiveCall := objectivesService.lastCall()
	if objectiveCall.workspaceID != scope.WorkspaceID || objectiveCall.userID != scope.UserID || objectiveCall.filters["limit"] != 20 {
		t.Fatalf("objectives received incorrect scope %#v", objectiveCall)
	}
}
