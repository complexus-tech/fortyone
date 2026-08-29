package messaging

import (
	"context"
	"testing"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

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
	filters     stories.CoreStoryFilters
}

type completedStoriesServiceStub struct {
	storiesServiceStub
	items        []stories.CoreStoryList
	groups       []stories.CoreStoryGroup
	calls        []completedStoriesServiceCall
	groupedCalls []groupedStoriesServiceCall
}

func (s *completedStoriesServiceStub) List(ctx context.Context, workspaceID uuid.UUID, filters stories.CoreStoryFilters) ([]stories.CoreStoryList, error) {
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
