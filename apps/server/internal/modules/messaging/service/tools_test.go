package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
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

func TestFortyOneToolExecutorOperationalCatalogIsOptionalAndStrict(t *testing.T) {
	t.Parallel()

	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{},
		&storiesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{},
			Users:  &usersServiceStub{},
		}),
	)
	definitions := executor.Definitions()
	if len(definitions) != 6 {
		t.Fatalf("expected six tools without a story reader, got %d", len(definitions))
	}
	if err := validateToolDefinitions(definitions); err != nil {
		t.Fatalf("operational catalog is not strict: %v", err)
	}
	expected := map[string]bool{toolListStatuses: false, toolListTeamMembers: false}
	for _, definition := range definitions {
		if _, operational := expected[definition.Name]; operational {
			expected[definition.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("operational tool %q is missing", name)
		}
	}
}

func TestFortyOneToolExecutorListStatusesIntersectsChannelScopeAndFiltersForgedRows(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	allowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("11111111-0000-0000-0000-000000000001"),
		Name:      "Web",
		Code:      "web",
		Workspace: scope.WorkspaceID,
	}
	disallowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("11111111-0000-0000-0000-000000000002"),
		Name:      "Operations",
		Code:      "OPS",
		Workspace: scope.WorkspaceID,
	}
	foreignWorkspaceID := uuid.MustParse("11111111-0000-0000-0000-000000000003")
	visibleStatus := states.CoreState{
		ID:         uuid.MustParse("11111111-0000-0000-0000-000000000011"),
		Name:       "In Progress",
		Category:   "started",
		OrderIndex: 3000,
		Team:       allowedTeam.ID,
		Workspace:  scope.WorkspaceID,
	}
	statesService := &statesServiceStub{items: []states.CoreState{
		visibleStatus,
		visibleStatus,
		{ID: uuid.New(), Name: "Secret", Category: "started", Team: disallowedTeam.ID, Workspace: scope.WorkspaceID},
		{ID: uuid.New(), Name: "Foreign", Category: "started", Team: allowedTeam.ID, Workspace: foreignWorkspaceID},
		{ID: uuid.New(), Name: " ", Category: "started", Team: allowedTeam.ID, Workspace: scope.WorkspaceID},
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{allowedTeam, disallowedTeam}},
		&storiesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: statesService, Users: &usersServiceStub{}}),
	)
	scope.AllowedTeamIDs = []uuid.UUID{allowedTeam.ID}

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolListStatuses,
		Arguments: json.RawMessage(`{"team_id":null,"limit":20}`),
	})
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	var result listStatusesResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode statuses: %v", err)
	}
	if result.Total != 1 || len(result.Statuses) != 1 || result.Statuses[0].ID != visibleStatus.ID {
		t.Fatalf("statuses escaped the effective team scope: %#v", result)
	}
	if result.Statuses[0].TeamName != "Web" || result.Statuses[0].TeamCode != "WEB" {
		t.Fatalf("status was not enriched with its human team: %#v", result.Statuses[0])
	}
	call := statesService.lastCall()
	if call.workspaceID != scope.WorkspaceID || call.userID != scope.UserID {
		t.Fatalf("statuses received incorrect actor scope: %#v", call)
	}

	callCount := statesService.callCount()
	arguments, _ := json.Marshal(map[string]any{"team_id": disallowedTeam.ID.String(), "limit": 20})
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolListStatuses, Arguments: arguments})
	if !errors.Is(err, ErrTeamNotAccessible) {
		t.Fatalf("expected forged team to be rejected, got %v", err)
	}
	if statesService.callCount() != callCount {
		t.Fatal("statuses service must not run for a team outside the channel audience")
	}
}

func TestFortyOneToolExecutorListTeamMembersRejectsForgedTeamAndOmitsInactiveSensitiveFields(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	allowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("22222222-0000-0000-0000-000000000001"),
		Name:      "Engineering",
		Code:      "eng",
		Workspace: scope.WorkspaceID,
	}
	disallowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("22222222-0000-0000-0000-000000000002"),
		Name:      "Finance",
		Code:      "FIN",
		Workspace: scope.WorkspaceID,
	}
	activeMember := users.CoreUser{
		ID:              uuid.MustParse("22222222-0000-0000-0000-000000000011"),
		FullName:        "Ada Lovelace",
		Username:        "ada",
		Email:           "secret@example.com",
		IsActive:        true,
		TeamAIRoleTitle: "Staff Engineer",
	}
	usersService := &usersServiceStub{items: []users.CoreUser{
		activeMember,
		activeMember,
		{ID: uuid.New(), FullName: "Inactive Person", Username: "inactive", Email: "inactive@example.com", IsActive: false},
		{ID: uuid.New(), FullName: "System Actor", Username: "system", Email: "system@example.com", IsActive: true, IsSystem: true},
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{allowedTeam, disallowedTeam}},
		&storiesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: &statesServiceStub{}, Users: usersService}),
	)
	scope.AllowedTeamIDs = []uuid.UUID{allowedTeam.ID}
	arguments, _ := json.Marshal(map[string]any{
		"team_id": allowedTeam.ID.String(),
		"query":   " Ada ",
		"limit":   2,
	})
	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamMembers, Arguments: arguments})
	if err != nil {
		t.Fatalf("list team members: %v", err)
	}
	var result listTeamMembersResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode team members: %v", err)
	}
	if result.Total != 1 || len(result.Members) != 1 {
		t.Fatalf("inactive, system, or duplicate members were returned: %#v", result)
	}
	member := result.Members[0]
	if member.DisplayName != "Ada Lovelace" || member.Username != "ada" || !member.Active || member.RoleTitle != "Staff Engineer" {
		t.Fatalf("unexpected member projection: %#v", member)
	}
	if strings.Contains(string(raw), "secret@example.com") || strings.Contains(string(raw), `"email"`) {
		t.Fatalf("member output leaked an email field: %s", raw)
	}
	call := usersService.lastListCall()
	if call.workspaceID != scope.WorkspaceID || call.filter.TeamID == nil || *call.filter.TeamID != allowedTeam.ID || call.filter.Search != "Ada" || call.filter.Limit != 3 {
		t.Fatalf("members query was not constrained to the authorized team: %#v", call)
	}

	callCount := usersService.listCallCount()
	forgedArguments, _ := json.Marshal(map[string]any{
		"team_id": disallowedTeam.ID.String(),
		"query":   nil,
		"limit":   nil,
	})
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamMembers, Arguments: forgedArguments})
	if !errors.Is(err, ErrTeamNotAccessible) {
		t.Fatalf("expected forged team to be rejected, got %v", err)
	}
	if usersService.listCallCount() != callCount {
		t.Fatal("users service must not run for a team outside the channel audience")
	}
}

func TestFortyOneToolExecutorMyTasksReturnsOnlyActiveAssignmentsWithHumanFields(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{
		ID:        uuid.MustParse("33333333-0000-0000-0000-000000000001"),
		Name:      "Web Platform",
		Code:      "web",
		Workspace: scope.WorkspaceID,
	}
	activeStatusID := uuid.MustParse("33333333-0000-0000-0000-000000000011")
	completedStatusID := uuid.MustParse("33333333-0000-0000-0000-000000000012")
	otherUserID := uuid.MustParse("33333333-0000-0000-0000-000000000013")
	currentUserID := scope.UserID
	now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	storiesService := &storiesServiceStub{items: []stories.CoreStoryList{
		{ID: uuid.New(), SequenceID: 41, Title: "Active assignment", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &currentUserID, UpdatedAt: now},
		{ID: uuid.New(), SequenceID: 42, Title: "Completed by status", Team: team.ID, Workspace: scope.WorkspaceID, Status: &completedStatusID, Assignee: &currentUserID},
		{ID: uuid.New(), SequenceID: 43, Title: "Completed timestamp", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &currentUserID, CompletedAt: &now},
		{ID: uuid.New(), SequenceID: 44, Title: "Deleted", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &currentUserID, DeletedAt: &now},
		{ID: uuid.New(), SequenceID: 45, Title: "Archived", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &currentUserID, ArchivedAt: &now},
		{ID: uuid.New(), SequenceID: 46, Title: "Someone else's", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &otherUserID},
	}}
	statesService := &statesServiceStub{items: []states.CoreState{
		{ID: activeStatusID, Name: "In Progress", Category: "started", Team: team.ID, Workspace: scope.WorkspaceID},
		{ID: completedStatusID, Name: "Done", Category: "completed", Team: team.ID, Workspace: scope.WorkspaceID},
	}}
	usersService := &usersServiceStub{current: users.CoreUser{
		ID:       scope.UserID,
		FullName: "Joseph Mukorivo",
		Username: "joseph",
		IsActive: true,
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storiesService,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: statesService, Users: usersService}),
	)

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListMyTasks, Arguments: json.RawMessage(`{"limit":20}`)})
	if err != nil {
		t.Fatalf("list my tasks: %v", err)
	}
	var result listTasksResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if result.Total != 1 || len(result.Tasks) != 1 {
		t.Fatalf("non-active or non-assigned tasks were returned: %#v", result)
	}
	task := result.Tasks[0]
	if task.Reference != "WEB-41" || task.TeamName != "Web Platform" || task.TeamCode != "WEB" || task.StatusName != "In Progress" || task.StatusCategory != "started" || task.AssigneeName != "Joseph Mukorivo" || task.AssigneeUsername != "joseph" {
		t.Fatalf("task did not contain human-readable enrichment: %#v", task)
	}
	call := storiesService.lastCall()
	if call.actorID != scope.UserID || call.actorErr != nil {
		t.Fatalf("MyStories did not receive the authenticated actor: %#v", call)
	}
}

func TestFortyOneToolExecutorListCompletedTasksUsesLocalDateRangeAndStoryLinks(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.Timezone = "Africa/Harare"
	scope.WebsiteURL = "https://fortyone.app"
	scope.WorkspaceSlug = "acme"
	team := teams.CoreTeam{
		ID:        uuid.MustParse("55555555-0000-0000-0000-000000000001"),
		Name:      "Web Platform",
		Code:      "web",
		Workspace: scope.WorkspaceID,
	}
	completedStatusID := uuid.MustParse("55555555-0000-0000-0000-000000000011")
	activeStatusID := uuid.MustParse("55555555-0000-0000-0000-000000000012")
	currentUserID := scope.UserID
	completedAtLatest := time.Date(2026, time.August, 12, 21, 30, 0, 0, time.UTC)
	completedAtEarlier := time.Date(2026, time.August, 12, 1, 0, 0, 0, time.UTC)
	completedAtNextDay := time.Date(2026, time.August, 13, 21, 0, 0, 0, time.UTC)
	completedStories := &completedStoriesServiceStub{items: []stories.CoreStoryList{
		{ID: uuid.New(), SequenceID: 41, Title: "Completed latest", Team: team.ID, Workspace: scope.WorkspaceID, Status: &completedStatusID, Assignee: &currentUserID, CompletedAt: &completedAtLatest},
		{ID: uuid.New(), SequenceID: 42, Title: "Completed earlier", Team: team.ID, Workspace: scope.WorkspaceID, Status: &completedStatusID, Assignee: &currentUserID, CompletedAt: &completedAtEarlier},
		{ID: uuid.New(), SequenceID: 43, Title: "Completed outside requested day", Team: team.ID, Workspace: scope.WorkspaceID, Status: &completedStatusID, Assignee: &currentUserID, CompletedAt: &completedAtNextDay},
		{ID: uuid.New(), SequenceID: 44, Title: "Active timestamp", Team: team.ID, Workspace: scope.WorkspaceID, Status: &activeStatusID, Assignee: &currentUserID, CompletedAt: &completedAtLatest},
		{ID: uuid.New(), SequenceID: 45, Title: "Other team", Team: uuid.New(), Workspace: scope.WorkspaceID, Status: &completedStatusID, Assignee: &currentUserID, CompletedAt: &completedAtLatest},
	}}
	statesService := &statesServiceStub{items: []states.CoreState{
		{ID: completedStatusID, Name: "Done", Category: "completed", Team: team.ID, Workspace: scope.WorkspaceID},
		{ID: activeStatusID, Name: "In Progress", Category: "started", Team: team.ID, Workspace: scope.WorkspaceID},
	}}
	usersService := &usersServiceStub{current: users.CoreUser{
		ID:       scope.UserID,
		FullName: "Joseph Mukorivo",
		Username: "joseph",
		IsActive: true,
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		completedStories,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: statesService, Users: usersService}),
	)

	raw, err := executor.Execute(context.Background(), scope, ToolCall{
		Name:      toolListCompleted,
		Arguments: json.RawMessage(`{"start_date":"2026-08-12","end_date":"2026-08-12","limit":20}`),
	})
	if err != nil {
		t.Fatalf("list completed tasks: %v", err)
	}
	var result listTasksResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode completed tasks: %v", err)
	}
	if result.Total != 2 || len(result.Tasks) != 2 {
		t.Fatalf("unexpected completed task result: %#v", result)
	}
	if result.Tasks[0].Reference != "WEB-41" || result.Tasks[0].CompletedAt == nil || !result.Tasks[0].CompletedAt.Equal(completedAtLatest) {
		t.Fatalf("completed tasks were not ordered by completion time: %#v", result.Tasks)
	}
	if result.Tasks[1].Reference != "WEB-42" || result.Tasks[1].StatusCategory != "completed" {
		t.Fatalf("completed task enrichment is incorrect: %#v", result.Tasks[1])
	}
	if result.Tasks[0].URL != "https://acme.fortyone.app/work/WEB-41" {
		t.Fatalf("completed task URL = %q", result.Tasks[0].URL)
	}

	call := completedStories.lastCall()
	if call.actorID != scope.UserID || call.actorErr != nil || call.workspaceID != scope.WorkspaceID {
		t.Fatalf("completed task query did not receive the authenticated scope: %#v", call)
	}
	completedAfter, ok := call.filters["completed_after"].(time.Time)
	if !ok || !completedAfter.Equal(time.Date(2026, time.August, 11, 22, 0, 0, 0, time.UTC)) {
		t.Fatalf("completed_after did not use the user's local timezone: %#v", call.filters)
	}
	completedBefore, ok := call.filters["completed_before"].(time.Time)
	if !ok || !completedBefore.Equal(time.Date(2026, time.August, 12, 21, 59, 59, 999999999, time.UTC)) {
		t.Fatalf("completed_before did not close the user's local day: %#v", call.filters)
	}
	if call.filters["assigned_to_me"] != true || call.filters["current_user_id"] != scope.UserID {
		t.Fatalf("completed task query was not restricted to the current assignee: %#v", call.filters)
	}
}

func TestCompletedTaskDateRangeDefaultsToTodayInUserTimezone(t *testing.T) {
	t.Parallel()

	start, end, err := completedTaskDateRange(nil, nil, "America/Los_Angeles", time.Date(2026, time.August, 13, 0, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("completedTaskDateRange() error = %v", err)
	}
	if !start.Equal(time.Date(2026, time.August, 12, 7, 0, 0, 0, time.UTC)) || !end.Equal(time.Date(2026, time.August, 13, 6, 59, 59, 999999999, time.UTC)) {
		t.Fatalf("default local-day range = %s to %s", start, end)
	}
}

func TestCompletedTaskDateRangeRejectsInvalidRanges(t *testing.T) {
	t.Parallel()

	startDate := "2026-08-13"
	endDate := "2026-08-12"
	if _, _, err := completedTaskDateRange(&startDate, &endDate, "UTC", time.Time{}); err == nil {
		t.Fatal("expected reversed completed task dates to fail")
	}
	tooEarly := "2024-01-01"
	tooLate := "2026-08-13"
	if _, _, err := completedTaskDateRange(&tooEarly, &tooLate, "UTC", time.Time{}); err == nil {
		t.Fatal("expected an overlong completed task range to fail")
	}
}

func TestFortyOneToolExecutorGetStoryResolvesAuthorizedHumanReferenceAndEnriches(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	allowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("44444444-0000-0000-0000-000000000001"),
		Name:      "Web",
		Code:      "web",
		Workspace: scope.WorkspaceID,
	}
	disallowedTeam := teams.CoreTeam{
		ID:        uuid.MustParse("44444444-0000-0000-0000-000000000002"),
		Name:      "Operations",
		Code:      "OPS",
		Workspace: scope.WorkspaceID,
	}
	statusID := uuid.MustParse("44444444-0000-0000-0000-000000000011")
	assigneeID := uuid.MustParse("44444444-0000-0000-0000-000000000012")
	description := "Investigate the login failure."
	reader := &storyReaderServiceStub{
		storiesServiceStub: storiesServiceStub{},
		story: stories.CoreSingleStory{
			ID:          uuid.MustParse("44444444-0000-0000-0000-000000000013"),
			SequenceID:  42,
			Title:       "Fix login",
			Description: &description,
			Team:        allowedTeam.ID,
			Workspace:   scope.WorkspaceID,
			Status:      &statusID,
			Assignee:    &assigneeID,
			Priority:    "High",
		},
	}
	statesService := &statesServiceStub{items: []states.CoreState{{
		ID: statusID, Name: "In Progress", Category: "started", Team: allowedTeam.ID, Workspace: scope.WorkspaceID,
	}}}
	usersService := &usersServiceStub{items: []users.CoreUser{{
		ID: assigneeID, FullName: "Ada Lovelace", Username: "ada", Email: "ada@example.com", IsActive: true,
	}}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{allowedTeam, disallowedTeam}},
		reader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: statesService, Users: usersService}),
	)
	scope.AllowedTeamIDs = []uuid.UUID{allowedTeam.ID}

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolGetStory, Arguments: json.RawMessage(`{"story_reference":" web 42 "}`)})
	if err != nil {
		t.Fatalf("get story: %v", err)
	}
	var result storyDetailsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode story: %v", err)
	}
	if result.Reference != "WEB-42" || result.Title != "Fix login" || result.TeamName != "Web" || result.StatusName != "In Progress" || result.AssigneeName != "Ada Lovelace" || result.AssigneeUsername != "ada" {
		t.Fatalf("story did not contain authorized human-readable details: %#v", result)
	}
	if len(reader.calls) != 1 || reader.calls[0].reference != "WEB-42" || reader.calls[0].actorID != scope.UserID || reader.calls[0].actorErr != nil {
		t.Fatalf("story reader received an incorrect reference or actor: %#v", reader.calls)
	}
	if strings.Contains(string(raw), "ada@example.com") || strings.Contains(string(raw), `"email"`) {
		t.Fatalf("story enrichment leaked an email field: %s", raw)
	}

	forgedCalls := len(reader.calls)
	_, err = executor.Execute(context.Background(), scope, ToolCall{Name: toolGetStory, Arguments: json.RawMessage(`{"story_reference":"OPS-42"}`)})
	if !errors.Is(err, ErrTeamNotAccessible) {
		t.Fatalf("expected inaccessible reference team error, got %v", err)
	}
	if len(reader.calls) != forgedCalls {
		t.Fatal("story reader must not run for a reference outside the channel audience")
	}
}

func TestFortyOneToolExecutorRejectsMissingStrictArguments(t *testing.T) {
	t.Parallel()

	executor := newToolExecutorForTest(t, &teamsServiceStub{}, &storiesServiceStub{}, &searchServiceStub{}, &objectivesServiceStub{})
	_, err := executor.Execute(context.Background(), testToolScope(), ToolCall{
		Name:      toolSearchWork,
		Arguments: json.RawMessage(`{"query":"launch"}`),
	})
	if !errors.Is(err, ErrInvalidToolArguments) {
		t.Fatalf("expected strict argument error, got %v", err)
	}
}

func newToolExecutorForTest(
	t *testing.T,
	teamsService TeamsService,
	storiesService StoriesService,
	searchService SearchService,
	objectivesService ObjectivesService,
	options ...FortyOneToolExecutorOption,
) *FortyOneToolExecutor {
	t.Helper()
	executor, err := NewFortyOneToolExecutor(teamsService, storiesService, searchService, objectivesService, options...)
	if err != nil {
		t.Fatalf("construct executor: %v", err)
	}
	return executor
}

func testToolScope() ToolScope {
	return ToolScope{
		WorkspaceID: uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
		UserID:      uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
	}
}

type teamsServiceCall struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
	filter      teams.CoreListTeamsFilter
}

type teamsServiceStub struct {
	joined []teams.CoreTeam
	all    []teams.CoreTeam
	calls  []teamsServiceCall
}

func (s *teamsServiceStub) List(_ context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error) {
	filter := teams.CoreListTeamsFilter{}
	if len(filters) > 0 {
		filter = filters[0]
	}
	s.calls = append(s.calls, teamsServiceCall{workspaceID: workspaceID, userID: userID, filter: filter})
	if filter.JoinedOnly {
		return append([]teams.CoreTeam(nil), s.joined...), nil
	}
	return append([]teams.CoreTeam(nil), s.all...), nil
}

func (s *teamsServiceStub) lastCall() teamsServiceCall {
	if len(s.calls) == 0 {
		return teamsServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

type storiesServiceCall struct {
	workspaceID uuid.UUID
	actorID     uuid.UUID
	actorErr    error
}

type storiesServiceStub struct {
	items []stories.CoreStoryList
	calls []storiesServiceCall
}

type completedStoriesServiceCall struct {
	workspaceID uuid.UUID
	actorID     uuid.UUID
	actorErr    error
	filters     map[string]any
}

type completedStoriesServiceStub struct {
	storiesServiceStub
	items []stories.CoreStoryList
	calls []completedStoriesServiceCall
}

func (s *completedStoriesServiceStub) List(ctx context.Context, workspaceID uuid.UUID, filters map[string]any) ([]stories.CoreStoryList, error) {
	actorID, actorErr := platformauth.GetUserID(ctx)
	s.calls = append(s.calls, completedStoriesServiceCall{
		workspaceID: workspaceID,
		actorID:     actorID,
		actorErr:    actorErr,
		filters:     filters,
	})
	return append([]stories.CoreStoryList(nil), s.items...), nil
}

func (s *completedStoriesServiceStub) lastCall() completedStoriesServiceCall {
	if len(s.calls) == 0 {
		return completedStoriesServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

func (s *storiesServiceStub) MyStories(ctx context.Context, workspaceID uuid.UUID) ([]stories.CoreStoryList, error) {
	actorID, err := platformauth.GetUserID(ctx)
	s.calls = append(s.calls, storiesServiceCall{workspaceID: workspaceID, actorID: actorID, actorErr: err})
	return append([]stories.CoreStoryList(nil), s.items...), nil
}

func (s *storiesServiceStub) lastCall() storiesServiceCall {
	if len(s.calls) == 0 {
		return storiesServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

type storyReaderCall struct {
	workspaceID uuid.UUID
	reference   string
	actorID     uuid.UUID
	actorErr    error
}

type storyReaderServiceStub struct {
	storiesServiceStub
	story stories.CoreSingleStory
	err   error
	calls []storyReaderCall
}

func (s *storyReaderServiceStub) QueryByRef(ctx context.Context, workspaceID uuid.UUID, reference string) (stories.CoreSingleStory, error) {
	actorID, actorErr := platformauth.GetUserID(ctx)
	s.calls = append(s.calls, storyReaderCall{
		workspaceID: workspaceID,
		reference:   reference,
		actorID:     actorID,
		actorErr:    actorErr,
	})
	return s.story, s.err
}

type searchServiceCall struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
	params      search.SearchParams
}

type searchServiceStub struct {
	result search.CoreSearchResult
	calls  []searchServiceCall
}

func (s *searchServiceStub) Search(_ context.Context, workspaceID uuid.UUID, userID uuid.UUID, params search.SearchParams) (search.CoreSearchResult, error) {
	s.calls = append(s.calls, searchServiceCall{workspaceID: workspaceID, userID: userID, params: params})
	return s.result, nil
}

func (s *searchServiceStub) callCount() int {
	return len(s.calls)
}

func (s *searchServiceStub) lastCall() searchServiceCall {
	if len(s.calls) == 0 {
		return searchServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

type objectivesServiceCall struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
	filters     map[string]any
}

type objectivesServiceStub struct {
	items []objectives.CoreObjective
	calls []objectivesServiceCall
}

func (s *objectivesServiceStub) List(_ context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error) {
	filterCopy := make(map[string]any, len(filters))
	for key, value := range filters {
		filterCopy[key] = value
	}
	s.calls = append(s.calls, objectivesServiceCall{workspaceID: workspaceID, userID: userID, filters: filterCopy})
	return append([]objectives.CoreObjective(nil), s.items...), nil
}

func (s *objectivesServiceStub) lastCall() objectivesServiceCall {
	if len(s.calls) == 0 {
		return objectivesServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

type statesServiceCall struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
}

type statesServiceStub struct {
	items []states.CoreState
	err   error
	calls []statesServiceCall
}

func (s *statesServiceStub) List(_ context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]states.CoreState, error) {
	s.calls = append(s.calls, statesServiceCall{workspaceID: workspaceID, userID: userID})
	return append([]states.CoreState(nil), s.items...), s.err
}

func (s *statesServiceStub) callCount() int {
	return len(s.calls)
}

func (s *statesServiceStub) lastCall() statesServiceCall {
	if len(s.calls) == 0 {
		return statesServiceCall{}
	}
	return s.calls[len(s.calls)-1]
}

type usersServiceCall struct {
	workspaceID uuid.UUID
	filter      users.CoreListUsersFilter
}

type usersServiceStub struct {
	current   users.CoreUser
	items     []users.CoreUser
	getErr    error
	listErr   error
	getCalls  []uuid.UUID
	listCalls []usersServiceCall
}

func (s *usersServiceStub) GetUser(_ context.Context, userID uuid.UUID) (users.CoreUser, error) {
	s.getCalls = append(s.getCalls, userID)
	return s.current, s.getErr
}

func (s *usersServiceStub) List(_ context.Context, workspaceID uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error) {
	filterCopy := filter
	if filter.TeamID != nil {
		teamID := *filter.TeamID
		filterCopy.TeamID = &teamID
	}
	s.listCalls = append(s.listCalls, usersServiceCall{workspaceID: workspaceID, filter: filterCopy})
	return append([]users.CoreUser(nil), s.items...), s.listErr
}

func (s *usersServiceStub) listCallCount() int {
	return len(s.listCalls)
}

func (s *usersServiceStub) lastListCall() usersServiceCall {
	if len(s.listCalls) == 0 {
		return usersServiceCall{}
	}
	return s.listCalls[len(s.listCalls)-1]
}
