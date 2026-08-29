package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

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
	if member.ID != activeMember.ID || member.DisplayName != "Ada Lovelace" || member.Username != "ada" || !member.Active || member.RoleTitle != "Staff Engineer" {
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
	completedAfter := call.filters.CompletedAfter
	if completedAfter == nil || !completedAfter.Equal(time.Date(2026, time.August, 11, 22, 0, 0, 0, time.UTC)) {
		t.Fatalf("completed_after did not use the user's local timezone: %#v", call.filters)
	}
	completedBefore := call.filters.CompletedBefore
	if completedBefore == nil || !completedBefore.Equal(time.Date(2026, time.August, 12, 21, 59, 59, 999999999, time.UTC)) {
		t.Fatalf("completed_before did not close the user's local day: %#v", call.filters)
	}
	if call.filters.AssignedToMe == nil || !*call.filters.AssignedToMe {
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
	estimatedDurationMinutes := 90
	minimumFocusBlockMinutes := 30
	reader := &storyReaderServiceStub{
		storiesServiceStub: storiesServiceStub{},
		story: stories.CoreSingleStory{
			ID:                       uuid.MustParse("44444444-0000-0000-0000-000000000013"),
			SequenceID:               42,
			Title:                    "Fix login",
			Description:              &description,
			Team:                     allowedTeam.ID,
			Workspace:                scope.WorkspaceID,
			Status:                   &statusID,
			Assignee:                 &assigneeID,
			Priority:                 "High",
			EstimatedDurationMinutes: &estimatedDurationMinutes,
			MinimumFocusBlockMinutes: &minimumFocusBlockMinutes,
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
	if result.EstimatedDurationMinutes == nil || *result.EstimatedDurationMinutes != estimatedDurationMinutes || result.MinimumFocusBlockMinutes == nil || *result.MinimumFocusBlockMinutes != minimumFocusBlockMinutes {
		t.Fatalf("story did not expose its explicit time contract: %#v", result)
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
