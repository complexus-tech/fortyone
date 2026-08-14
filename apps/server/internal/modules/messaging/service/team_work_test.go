package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestFortyOneToolExecutorTeamWorkCatalogRequiresBoundedReaderAndOperationalServices(t *testing.T) {
	t.Parallel()

	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{},
		&teamWorkOnlyStoriesServiceStub{},
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{},
			Users:  &usersServiceStub{},
		}),
	)
	definitions := executor.Definitions()
	if len(definitions) != 7 {
		t.Fatalf("expected seven tools with a bounded team work reader but no completed-work reader, got %d", len(definitions))
	}
	if err := validateToolDefinitions(definitions); err != nil {
		t.Fatalf("team work catalog is not strict: %v", err)
	}
	var teamWorkDefinition ToolDefinition
	for _, definition := range definitions {
		if definition.Name == toolListTeamWork {
			teamWorkDefinition = definition
			break
		}
	}
	if teamWorkDefinition.Name == "" {
		t.Fatal("list_team_work definition is missing")
	}
	for _, definition := range definitions {
		if definition.Name == toolListCompleted {
			t.Fatal("list_completed_tasks must not be registered without its own reader interface")
		}
	}
	required, ok := teamWorkDefinition.Parameters["required"].([]string)
	if !ok || len(required) != 9 {
		t.Fatalf("list_team_work must require its complete strict contract: %#v", teamWorkDefinition.Parameters)
	}
}

func TestFortyOneToolExecutorListTeamWorkMeUsesAllowedScopeAndBoundedQuery(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.AllowedTeamIDs = []uuid.UUID{uuid.MustParse("66666666-0000-0000-0000-000000000001")}
	scope.SharedTeamIDs = []uuid.UUID{}
	scope.WebsiteURL = "https://fortyone.app"
	scope.WorkspaceSlug = "acme"
	team := teams.CoreTeam{
		ID:        scope.AllowedTeamIDs[0],
		Name:      "Product",
		Code:      "pro",
		Workspace: scope.WorkspaceID,
	}
	startedStatusID := uuid.MustParse("66666666-0000-0000-0000-000000000011")
	updatedAt := time.Date(2026, time.August, 14, 8, 0, 0, 0, time.UTC)
	storyReader := &completedStoriesServiceStub{groups: []stories.CoreStoryGroup{{
		Key:         teamWorkGroupNone,
		LoadedCount: 1,
		TotalCount:  1,
		Stories: []stories.CoreStoryList{{
			ID:         uuid.New(),
			SequenceID: 71,
			Title:      "Ship team summaries",
			Team:       team.ID,
			Workspace:  scope.WorkspaceID,
			Status:     &startedStatusID,
			Assignee:   &scope.UserID,
			UpdatedAt:  updatedAt,
		}},
	}}}
	usersService := &usersServiceStub{items: []users.CoreUser{{
		ID:       scope.UserID,
		FullName: "Joseph Mukorivo",
		Username: "joseph",
		IsActive: true,
	}}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{items: []states.CoreState{{
				ID:        startedStatusID,
				Name:      "In Progress",
				Category:  "started",
				Team:      team.ID,
				Workspace: scope.WorkspaceID,
			}}},
			Users: usersService,
		}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeMe,
		"assignee_ids":       nil,
		"mode":               teamWorkModeInProgress,
		"start_date":         nil,
		"end_date":           nil,
		"group_by":           teamWorkGroupNone,
		"limit":              nil,
		"limit_per_assignee": nil,
	})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if err != nil {
		t.Fatalf("list personal team work: %v", err)
	}
	var result listTeamWorkResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode personal team work: %v", err)
	}
	if result.Access != teamWorkAccessGranted || result.Total != 1 || result.Returned != 1 || result.Truncated || len(result.Tasks) != 1 {
		t.Fatalf("unexpected personal team work result: %#v", result)
	}
	if result.Tasks[0].Reference != "PRO-71" || result.Tasks[0].URL != "https://acme.fortyone.app/work/PRO-71" || result.Tasks[0].AssigneeName != "Joseph Mukorivo" {
		t.Fatalf("personal team work was not human-readable: %#v", result.Tasks[0])
	}
	if result.Team.SharedWorkEnabled {
		t.Fatal("personal team work must not imply that cross-assignee access is enabled")
	}
	if storyReader.groupedCallCount() != 1 {
		t.Fatalf("expected exactly one bounded stories query, got %d", storyReader.groupedCallCount())
	}
	call := storyReader.lastGroupedCall()
	if call.actorID != scope.UserID || call.actorErr != nil || call.query.GroupBy != teamWorkGroupNone || call.query.StoriesPerGroup != defaultTeamWorkLimit || call.query.Page != 1 {
		t.Fatalf("team work query was not bounded or actor-scoped: %#v", call)
	}
	if len(call.query.Filters.TeamIDs) != 1 || call.query.Filters.TeamIDs[0] != team.ID || len(call.query.Filters.AssigneeIDs) != 1 || call.query.Filters.AssigneeIDs[0] != scope.UserID {
		t.Fatalf("team work query was not restricted to the exact team and assignee: %#v", call.query.Filters)
	}
	if len(call.query.Filters.Categories) != 1 || call.query.Filters.Categories[0] != "started" || call.query.Filters.ShowSubStories == nil || *call.query.Filters.ShowSubStories {
		t.Fatalf("in-progress query semantics are incorrect: %#v", call.query.Filters)
	}
	if call.query.Filters.IsNotCompleted == nil || !*call.query.Filters.IsNotCompleted || call.query.Filters.IsCompleted != nil {
		t.Fatalf("in-progress query did not exclude completed stories: %#v", call.query.Filters)
	}
}

func TestFortyOneToolExecutorListTeamWorkReturnsStructuredSharedAccessDenial(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{
		ID:        uuid.MustParse("77777777-0000-0000-0000-000000000001"),
		Name:      "Product",
		Code:      "PRO",
		Workspace: scope.WorkspaceID,
	}
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SharedTeamIDs = []uuid.UUID{}
	storyReader := &completedStoriesServiceStub{}
	usersService := &usersServiceStub{}
	statesService := &statesServiceStub{}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: statesService, Users: usersService}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeAll,
		"assignee_ids":       nil,
		"mode":               teamWorkModeCompleted,
		"start_date":         "2026-08-13",
		"end_date":           "2026-08-13",
		"group_by":           teamWorkGroupAssignee,
		"limit":              20,
		"limit_per_assignee": 5,
	})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if err != nil {
		t.Fatalf("expected a model-readable denial, got %v", err)
	}
	var result listTeamWorkResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode denied team work: %v", err)
	}
	if result.Access != teamWorkAccessDenied || result.AccessReason != "shared_team_scope_required" || result.Team.SharedWorkEnabled {
		t.Fatalf("unexpected shared-work denial: %#v", result)
	}
	if usersService.listCallCount() != 0 || statesService.callCount() != 0 || storyReader.groupedCallCount() != 0 {
		t.Fatal("denied shared work must not query members, statuses, or stories")
	}

	teamsRaw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeams, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("list teams: %v", err)
	}
	var teamsResult listTeamsResult
	if err := json.Unmarshal(teamsRaw, &teamsResult); err != nil {
		t.Fatalf("decode teams: %v", err)
	}
	if len(teamsResult.Teams) != 1 || teamsResult.Teams[0].SharedWorkEnabled {
		t.Fatalf("list_teams did not expose the shared-work denial: %#v", teamsResult)
	}
}

func TestFortyOneToolExecutorListTeamWorkRejectsSelectedNonMemberBeforeStoryQuery(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{
		ID:        uuid.MustParse("88888888-0000-0000-0000-000000000001"),
		Workspace: scope.WorkspaceID,
	}
	member := users.CoreUser{ID: uuid.MustParse("88888888-0000-0000-0000-000000000011"), FullName: "Ada", IsActive: true}
	nonMemberID := uuid.MustParse("88888888-0000-0000-0000-000000000012")
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	storyReader := &completedStoriesServiceStub{}
	usersService := &usersServiceStub{items: []users.CoreUser{member}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{States: &statesServiceStub{}, Users: usersService}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeSelected,
		"assignee_ids":       []string{nonMemberID.String()},
		"mode":               teamWorkModeActive,
		"start_date":         nil,
		"end_date":           nil,
		"group_by":           teamWorkGroupNone,
		"limit":              20,
		"limit_per_assignee": nil,
	})

	_, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if !errors.Is(err, ErrInvalidToolArguments) {
		t.Fatalf("expected exact member validation, got %v", err)
	}
	if usersService.listCallCount() != 1 || storyReader.groupedCallCount() != 0 {
		t.Fatal("a forged selected member must be rejected before the stories query")
	}
}

func TestFortyOneToolExecutorListTeamWorkGroupsCompletedAssignmentsWithTruncation(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.Timezone = "Africa/Harare"
	scope.WebsiteURL = "https://fortyone.app"
	scope.WorkspaceSlug = "acme"
	team := teams.CoreTeam{
		ID:        uuid.MustParse("99999999-0000-0000-0000-000000000001"),
		Name:      "Product",
		Code:      "pro",
		Workspace: scope.WorkspaceID,
	}
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	ada := users.CoreUser{ID: uuid.MustParse("99999999-0000-0000-0000-000000000011"), FullName: "Ada Lovelace", Username: "ada", IsActive: true}
	grace := users.CoreUser{ID: uuid.MustParse("99999999-0000-0000-0000-000000000012"), FullName: "Grace Hopper", Username: "grace", IsActive: true}
	linus := users.CoreUser{ID: uuid.MustParse("99999999-0000-0000-0000-000000000013"), FullName: "Linus Torvalds", Username: "linus", IsActive: true}
	doneStatusID := uuid.MustParse("99999999-0000-0000-0000-000000000021")
	completedTimes := []time.Time{
		time.Date(2026, time.August, 13, 20, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 13, 19, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 13, 18, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 13, 17, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 13, 16, 0, 0, 0, time.UTC),
	}
	adaStories := []stories.CoreStoryList{
		{ID: uuid.New(), SequenceID: 81, Title: "Ada one", Team: team.ID, Workspace: scope.WorkspaceID, Status: &doneStatusID, Assignee: &ada.ID, CompletedAt: &completedTimes[0], UpdatedAt: completedTimes[0]},
		{ID: uuid.New(), SequenceID: 82, Title: "Ada two", Team: team.ID, Workspace: scope.WorkspaceID, Status: &doneStatusID, Assignee: &ada.ID, CompletedAt: &completedTimes[1], UpdatedAt: completedTimes[1]},
	}
	graceStories := []stories.CoreStoryList{
		{ID: uuid.New(), SequenceID: 84, Title: "Grace one", Team: team.ID, Workspace: scope.WorkspaceID, Status: &doneStatusID, Assignee: &grace.ID, CompletedAt: &completedTimes[3], UpdatedAt: completedTimes[3]},
		{ID: uuid.New(), SequenceID: 85, Title: "Grace two", Team: team.ID, Workspace: scope.WorkspaceID, Status: &doneStatusID, Assignee: &grace.ID, CompletedAt: &completedTimes[4], UpdatedAt: completedTimes[4]},
	}
	forgedAssigneeID := uuid.New()
	storyReader := &completedStoriesServiceStub{groups: []stories.CoreStoryGroup{
		{Key: ada.ID.String(), LoadedCount: 2, TotalCount: 3, HasMore: true, Stories: adaStories},
		{Key: grace.ID.String(), LoadedCount: 2, TotalCount: 2, Stories: graceStories},
		{Key: linus.ID.String(), LoadedCount: 0, TotalCount: 0, Stories: []stories.CoreStoryList{}},
		{
			Key:         forgedAssigneeID.String(),
			LoadedCount: 1,
			TotalCount:  1,
			Stories: []stories.CoreStoryList{{
				ID: uuid.New(), SequenceID: 86, Title: "Forged assignee", Team: team.ID, Workspace: scope.WorkspaceID, Status: &doneStatusID, Assignee: &forgedAssigneeID, CompletedAt: &completedTimes[4]},
			},
		},
	}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{items: []states.CoreState{{ID: doneStatusID, Name: "Done", Category: "completed", Team: team.ID, Workspace: scope.WorkspaceID}}},
			Users:  &usersServiceStub{items: []users.CoreUser{linus, grace, ada, ada}},
		}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeAll,
		"assignee_ids":       nil,
		"mode":               teamWorkModeCompleted,
		"start_date":         "2026-08-13",
		"end_date":           "2026-08-13",
		"group_by":           teamWorkGroupAssignee,
		"limit":              3,
		"limit_per_assignee": 2,
	})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if err != nil {
		t.Fatalf("list completed team work: %v", err)
	}
	var result listTeamWorkResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode completed team work: %v", err)
	}
	if result.Access != teamWorkAccessGranted || result.Total != 5 || result.Returned != 3 || !result.Truncated || result.AssigneeTotal != 3 || result.AssigneeReturned != 3 || len(result.Groups) != 3 {
		t.Fatalf("unexpected grouped team work result: %#v", result)
	}
	if result.Groups[0].AssigneeID != ada.ID || result.Groups[0].Returned != 2 || !result.Groups[0].Truncated || result.Groups[1].AssigneeID != grace.ID || result.Groups[1].Returned != 1 {
		t.Fatalf("completed work was not grouped and capped by assignee: %#v", result.Groups)
	}
	if result.Groups[2].AssigneeID != linus.ID || result.Groups[2].Total != 0 || result.Groups[2].Returned != 0 || len(result.Groups[2].Tasks) != 0 || result.Groups[2].Truncated {
		t.Fatalf("zero-result assignee was not represented: %#v", result.Groups[2])
	}
	if result.Groups[0].Tasks[0].Reference != "PRO-81" || result.Groups[0].Tasks[0].URL != "https://acme.fortyone.app/work/PRO-81" {
		t.Fatalf("completed work did not include canonical story links: %#v", result.Groups[0].Tasks[0])
	}
	if strings.Contains(string(raw), "Forged assignee") {
		t.Fatalf("defensive post-filtering leaked a forged story: %s", raw)
	}
	if storyReader.groupedCallCount() != 1 {
		t.Fatalf("expected one bounded completed-work query, got %d", storyReader.groupedCallCount())
	}
	call := storyReader.lastGroupedCall()
	if call.query.GroupBy != teamWorkGroupAssignee || call.query.StoriesPerGroup != 2 || call.query.PageSize != 3 || call.query.OrderBy != "completed" || call.query.OrderDirection != "desc" {
		t.Fatalf("completed-work query was not bounded per assignee: %#v", call.query)
	}
	if call.query.Filters.IsCompleted == nil || !*call.query.Filters.IsCompleted || call.query.Filters.IsNotCompleted != nil {
		t.Fatalf("completed-work query did not require completed stories: %#v", call.query.Filters)
	}
	if call.query.Filters.CompletedAfter == nil || !call.query.Filters.CompletedAfter.Equal(time.Date(2026, time.August, 12, 22, 0, 0, 0, time.UTC)) {
		t.Fatalf("completed_after did not use the local date: %#v", call.query.Filters.CompletedAfter)
	}
	if call.query.Filters.CompletedBefore == nil || !call.query.Filters.CompletedBefore.Equal(time.Date(2026, time.August, 13, 21, 59, 59, 999999999, time.UTC)) {
		t.Fatalf("completed_before did not close the local date: %#v", call.query.Filters.CompletedBefore)
	}
	if len(call.query.Filters.AssigneeIDs) != 3 || call.query.Filters.AssigneeIDs[0] != ada.ID || call.query.Filters.AssigneeIDs[1] != grace.ID || call.query.Filters.AssigneeIDs[2] != linus.ID {
		t.Fatalf("completed-work assignees were not exact active team members: %#v", call.query.Filters.AssigneeIDs)
	}
}

func TestFortyOneToolExecutorListTeamWorkDueUsesCalendarDatesAndActiveStatuses(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	scope.Timezone = "America/Los_Angeles"
	team := teams.CoreTeam{ID: uuid.MustParse("aaaaaaaa-1000-0000-0000-000000000001"), Name: "Web", Code: "WEB", Workspace: scope.WorkspaceID}
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	unstartedStatusID := uuid.MustParse("aaaaaaaa-1000-0000-0000-000000000011")
	deadline := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.August, 12, 20, 0, 0, 0, time.UTC)
	storyReader := &completedStoriesServiceStub{groups: []stories.CoreStoryGroup{{
		Key:         teamWorkGroupNone,
		LoadedCount: 2,
		TotalCount:  2,
		Stories: []stories.CoreStoryList{
			{ID: uuid.New(), SequenceID: 91, Title: "Due today", Team: team.ID, Workspace: scope.WorkspaceID, Status: &unstartedStatusID, Assignee: &scope.UserID, EndDate: &deadline},
			{ID: uuid.New(), SequenceID: 92, Title: "Already completed", Team: team.ID, Workspace: scope.WorkspaceID, Status: &unstartedStatusID, Assignee: &scope.UserID, EndDate: &deadline, CompletedAt: &completedAt},
		},
	}}}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{items: []states.CoreState{{ID: unstartedStatusID, Name: "To Do", Category: "unstarted", Team: team.ID, Workspace: scope.WorkspaceID}}},
			Users:  &usersServiceStub{items: []users.CoreUser{{ID: scope.UserID, FullName: "Joseph", IsActive: true}}},
		}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeMe,
		"assignee_ids":       nil,
		"mode":               teamWorkModeDue,
		"start_date":         "2026-08-13",
		"end_date":           "2026-08-13",
		"group_by":           teamWorkGroupNone,
		"limit":              10,
		"limit_per_assignee": nil,
	})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if err != nil {
		t.Fatalf("list due team work: %v", err)
	}
	var result listTeamWorkResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode due team work: %v", err)
	}
	if result.Returned != 1 || len(result.Tasks) != 1 || result.Tasks[0].Reference != "WEB-91" || result.StartDate == nil || *result.StartDate != "2026-08-13" {
		t.Fatalf("due work did not use calendar-date semantics: %#v", result)
	}
	if result.Total != 1 || result.TotalIsExact || !result.Truncated {
		t.Fatalf("defensive filtering must mark repository totals as inexact: %#v", result)
	}
	call := storyReader.lastGroupedCall()
	if call.query.OrderBy != "deadline" || call.query.OrderDirection != "asc" || call.query.Filters.DeadlineAfter == nil || call.query.Filters.DeadlineBefore == nil || !call.query.Filters.DeadlineAfter.Equal(deadline) || !call.query.Filters.DeadlineBefore.Equal(deadline) {
		t.Fatalf("due query did not preserve the local calendar date: %#v", call.query)
	}
	if strings.Join(call.query.Filters.Categories, ",") != "backlog,unstarted,started,paused" {
		t.Fatalf("due query did not use every active status category: %#v", call.query.Filters.Categories)
	}
	if call.query.Filters.IsNotCompleted == nil || !*call.query.Filters.IsNotCompleted || call.query.Filters.IsCompleted != nil {
		t.Fatalf("due query did not exclude completed stories: %#v", call.query.Filters)
	}
}

func TestFortyOneToolExecutorListTeamWorkCapsAllAssignees(t *testing.T) {
	t.Parallel()

	scope := testToolScope()
	team := teams.CoreTeam{ID: uuid.New(), Name: "Large Team", Code: "BIG", Workspace: scope.WorkspaceID}
	scope.AllowedTeamIDs = []uuid.UUID{team.ID}
	scope.SharedTeamIDs = []uuid.UUID{team.ID}
	members := make([]users.CoreUser, 0, maxTeamWorkAssignees+1)
	for index := 0; index < maxTeamWorkAssignees+1; index++ {
		members = append(members, users.CoreUser{
			ID:       uuid.New(),
			FullName: fmt.Sprintf("Member %02d", index),
			IsActive: true,
		})
	}
	storyReader := &completedStoriesServiceStub{}
	usersService := &usersServiceStub{items: members}
	executor := newToolExecutorForTest(
		t,
		&teamsServiceStub{joined: []teams.CoreTeam{team}},
		storyReader,
		&searchServiceStub{},
		&objectivesServiceStub{},
		WithOperationalTools(OperationalToolServices{
			States: &statesServiceStub{},
			Users:  usersService,
		}),
	)
	arguments, _ := json.Marshal(map[string]any{
		"team_id":            team.ID.String(),
		"assignee_scope":     teamWorkAssigneeAll,
		"assignee_ids":       nil,
		"mode":               teamWorkModeInProgress,
		"start_date":         nil,
		"end_date":           nil,
		"group_by":           teamWorkGroupAssignee,
		"limit":              1,
		"limit_per_assignee": maxToolLimit,
	})

	raw, err := executor.Execute(context.Background(), scope, ToolCall{Name: toolListTeamWork, Arguments: arguments})
	if err != nil {
		t.Fatalf("list large-team work: %v", err)
	}
	var result listTeamWorkResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode large-team work: %v", err)
	}
	if !result.AssigneesTruncated || !result.Truncated || result.TotalIsExact {
		t.Fatalf("large-team result did not disclose the assignee cap: %#v", result)
	}
	if result.Limit != 1 || result.LimitPerAssignee != 1 {
		t.Fatalf("effective output limits were not reported: %#v", result)
	}
	if result.AssigneeTotal != maxTeamWorkAssignees || result.AssigneeReturned != maxTeamWorkAssignees || len(result.Groups) != maxTeamWorkAssignees {
		t.Fatalf("every queried assignee must have an output group: %#v", result)
	}
	for _, group := range result.Groups {
		if group.Total != 0 || group.Returned != 0 || len(group.Tasks) != 0 || group.Truncated || !group.TotalIsExact {
			t.Fatalf("unexpected empty assignee group: %#v", group)
		}
	}
	if got := usersService.lastListCall().filter.Limit; got != maxTeamWorkAssignees+1 {
		t.Fatalf("member lookup limit = %d, want %d", got, maxTeamWorkAssignees+1)
	}
	call := storyReader.lastGroupedCall()
	if len(call.query.Filters.AssigneeIDs) != maxTeamWorkAssignees {
		t.Fatalf("story query assignees = %d, want %d", len(call.query.Filters.AssigneeIDs), maxTeamWorkAssignees)
	}
	if call.query.StoriesPerGroup != 1 {
		t.Fatalf("stories per assignee = %d, want global cap 1", call.query.StoriesPerGroup)
	}
}

func TestTeamWorkActiveModeMeansOpenAssignmentsRatherThanDateActivity(t *testing.T) {
	t.Parallel()

	if strings.Join(teamWorkCategories(teamWorkModeActive), ",") != "backlog,unstarted,started,paused" {
		t.Fatalf("active mode categories = %#v", teamWorkCategories(teamWorkModeActive))
	}
	if !teamWorkStoryMatchesDate(stories.CoreStoryList{}, teamWorkModeActive, nil) {
		t.Fatal("an open assignment should match active mode")
	}
	completedAt := time.Now()
	if teamWorkStoryMatchesDate(stories.CoreStoryList{CompletedAt: &completedAt}, teamWorkModeActive, nil) {
		t.Fatal("a completed assignment must not match active mode")
	}
	date := "2026-08-13"
	if err := validateTeamWorkArguments(teamWorkAssigneeMe, nil, teamWorkModeActive, &date, nil, teamWorkGroupNone, nil); !errors.Is(err, ErrInvalidToolArguments) {
		t.Fatalf("active mode must reject date filters, got %v", err)
	}
}

type teamWorkOnlyStoriesServiceStub struct {
	storiesServiceStub
}

func (s *teamWorkOnlyStoriesServiceStub) ListGroupedStories(_ context.Context, _ stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	return nil, nil
}

type groupedStoriesServiceCall struct {
	actorID  uuid.UUID
	actorErr error
	query    stories.CoreStoryQuery
}

func (s *completedStoriesServiceStub) ListGroupedStories(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	actorID, actorErr := platformauth.GetUserID(ctx)
	query.Filters.TeamIDs = append([]uuid.UUID(nil), query.Filters.TeamIDs...)
	query.Filters.AssigneeIDs = append([]uuid.UUID(nil), query.Filters.AssigneeIDs...)
	query.Filters.Categories = append([]string(nil), query.Filters.Categories...)
	s.groupedCalls = append(s.groupedCalls, groupedStoriesServiceCall{
		actorID:  actorID,
		actorErr: actorErr,
		query:    query,
	})
	result := make([]stories.CoreStoryGroup, len(s.groups))
	copy(result, s.groups)
	for index := range result {
		result[index].Stories = append([]stories.CoreStoryList(nil), result[index].Stories...)
	}
	return result, nil
}

func (s *completedStoriesServiceStub) groupedCallCount() int {
	return len(s.groupedCalls)
}

func (s *completedStoriesServiceStub) lastGroupedCall() groupedStoriesServiceCall {
	if len(s.groupedCalls) == 0 {
		return groupedStoriesServiceCall{}
	}
	return s.groupedCalls[len(s.groupedCalls)-1]
}
