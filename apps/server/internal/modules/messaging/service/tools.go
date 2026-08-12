package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	toolListTeams       = "list_teams"
	toolListMyTasks     = "list_my_tasks"
	toolSearchWork      = "search_work"
	toolListObjectives  = "list_objectives"
	toolListStatuses    = "list_statuses"
	toolListTeamMembers = "list_team_members"
	toolGetStory        = "get_story"
	toolCreateStory     = "create_story"
	toolUpdateStory     = "update_story"
	toolAddComment      = "add_story_comment"
	toolAddRelationship = "add_story_relationship"

	defaultToolLimit         = 20
	maxToolLimit             = 50
	maxSearchRunes           = 200
	maxStoryReferenceRunes   = 64
	maxStoryDescriptionRunes = 8_000
)

// TeamsService is the subset of the teams domain used by assistant tools.
type TeamsService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error)
}

// StoriesService is the subset of the stories domain used by assistant tools.
type StoriesService interface {
	MyStories(ctx context.Context, workspaceID uuid.UUID) ([]stories.CoreStoryList, error)
}

// StoryReaderService is implemented by story services that support resolving a
// human-readable FortyOne reference. The base executor remains compatible with
// narrower story implementations that only support list_my_tasks.
type StoryReaderService interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error)
}

// StoryMutationService is the narrow stories-domain surface needed to propose
// and explicitly confirm assistant story writes.
type StoryMutationService interface {
	StoriesService
	Get(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error)
	CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, story stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateExternalUserActionIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) error
	CreateCommentExternal(ctx context.Context, actorID, workspaceID uuid.UUID, comment stories.CoreNewComment) (comments.CoreComment, error)
	UpdateLabels(ctx context.Context, id, workspaceID uuid.UUID, labels []uuid.UUID) error
	AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (stories.CoreStoryAssociation, error)
	FindFirstStatusByCategory(ctx context.Context, teamID, workspaceID uuid.UUID, category string) (*uuid.UUID, error)
}

// SearchService is the subset of the search domain used by assistant tools.
type SearchService interface {
	Search(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, params search.SearchParams) (search.CoreSearchResult, error)
}

// ObjectivesService is the subset of the objectives domain used by assistant tools.
type ObjectivesService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error)
}

// StatesService is the membership-aware subset of the statuses domain used by
// optional operational tools and result enrichment.
type StatesService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]states.CoreState, error)
}

// UsersService is the subset of the users domain used by optional operational
// tools. List is always called with an already-authorized team filter before
// another member is returned to the model.
type UsersService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error)
	List(ctx context.Context, workspaceID uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error)
}

// OperationalToolServices groups the optional read-only domain services used
// for status/member lookups and human-readable story enrichment.
type OperationalToolServices struct {
	States StatesService
	Users  UsersService
}

var (
	_ StoryReaderService = (*stories.Service)(nil)
	_ StatesService      = (*states.Service)(nil)
	_ UsersService       = (*users.Service)(nil)
)

// FortyOneToolExecutor exposes the deliberately small FortyOne tool catalog
// used by messaging assistants. Mutation tools only produce signed proposals;
// ConfirmStoryMutation is the sole write boundary.
type FortyOneToolExecutor struct {
	teams       TeamsService
	stories     StoriesService
	storyReader StoryReaderService
	search      SearchService
	objectives  ObjectivesService
	states      StatesService
	users       UsersService
	mutations   *storyMutationExecutor
	definitions []ToolDefinition
}

type fortyOneToolExecutorConfig struct {
	storyMutationSecret string
	storyMutationStore  StoryMutationConfirmationStore
	operational         *OperationalToolServices
}

// FortyOneToolExecutorOption configures optional capabilities without
// widening the default read-only executor.
type FortyOneToolExecutorOption func(*fortyOneToolExecutorConfig) error

// WithStoryMutations enables signed create/update proposals and explicit
// confirmation. secret must be an application secret shared by every worker
// replica so confirmations survive retries and deployments.
func WithStoryMutations(secret string) FortyOneToolExecutorOption {
	secret = strings.TrimSpace(secret)
	return func(config *fortyOneToolExecutorConfig) error {
		if secret == "" {
			return errors.New("story mutation confirmation secret is required")
		}
		config.storyMutationSecret = secret
		return nil
	}
}

// WithStoryMutationConfirmationStore supplies the durable state machine used
// to arbitrate confirmation and cancellation across every API/worker replica.
func WithStoryMutationConfirmationStore(store StoryMutationConfirmationStore) FortyOneToolExecutorOption {
	return func(config *fortyOneToolExecutorConfig) error {
		if store == nil {
			return errors.New("story mutation confirmation store is required")
		}
		config.storyMutationStore = store
		return nil
	}
}

// WithOperationalTools enables the read-only status/member catalog and
// human-readable enrichment. Both services are required so enabled tools never
// return a partially enriched or inconsistently authorized response.
func WithOperationalTools(services OperationalToolServices) FortyOneToolExecutorOption {
	return func(config *fortyOneToolExecutorConfig) error {
		if services.States == nil || services.Users == nil {
			return errors.New("states and users services are required for operational assistant tools")
		}
		configuredServices := services
		config.operational = &configuredServices
		return nil
	}
}

// NewFortyOneToolExecutor constructs an executor backed by the existing domain
// services. Every execution re-establishes authoritative user context and
// resolves joined teams before returning data or proposing a mutation.
func NewFortyOneToolExecutor(
	teamsService TeamsService,
	storiesService StoriesService,
	searchService SearchService,
	objectivesService ObjectivesService,
	options ...FortyOneToolExecutorOption,
) (*FortyOneToolExecutor, error) {
	if teamsService == nil || storiesService == nil || searchService == nil || objectivesService == nil {
		return nil, errors.New("all FortyOne assistant tool services are required")
	}

	config := fortyOneToolExecutorConfig{}
	for _, option := range options {
		if option == nil {
			return nil, errors.New("FortyOne assistant tool option must not be nil")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	storyReader, _ := storiesService.(StoryReaderService)
	executor := &FortyOneToolExecutor{
		teams:       teamsService,
		stories:     storiesService,
		storyReader: storyReader,
		search:      searchService,
		objectives:  objectivesService,
		definitions: fortyOneToolDefinitions(),
	}
	if config.operational != nil {
		executor.states = config.operational.States
		executor.users = config.operational.Users
		executor.definitions = append(executor.definitions, operationalToolDefinitions(storyReader != nil)...)
	}
	if config.storyMutationSecret != "" {
		if config.storyMutationStore == nil {
			return nil, errors.New("story mutation confirmation store is required when story mutations are enabled")
		}
		mutationService, ok := storiesService.(StoryMutationService)
		if !ok {
			return nil, errors.New("stories service does not support assistant story mutations")
		}
		executor.mutations = newStoryMutationExecutor(mutationService, config.storyMutationSecret, config.storyMutationStore)
		executor.definitions = append(executor.definitions, storyMutationToolDefinitions()...)
	}
	return executor, nil
}

// Definitions returns a defensive copy of the configured catalog.
func (e *FortyOneToolExecutor) Definitions() []ToolDefinition {
	return cloneToolDefinitions(e.definitions)
}

// Execute runs one tool in the supplied authoritative scope. Mutation entries
// still only prepare proposals; confirmation remains a separate write boundary.
func (e *FortyOneToolExecutor) Execute(ctx context.Context, scope ToolScope, call ToolCall) (json.RawMessage, error) {
	if err := validateToolScope(&scope); err != nil {
		return nil, err
	}

	ctx = platformauth.SetUserID(ctx, scope.UserID)
	switch call.Name {
	case toolListTeams:
		return e.listTeams(ctx, scope, call.Arguments)
	case toolListMyTasks:
		return e.listMyTasks(ctx, scope, call.Arguments)
	case toolSearchWork:
		return e.searchWork(ctx, scope, call.Arguments)
	case toolListObjectives:
		return e.listObjectives(ctx, scope, call.Arguments)
	case toolListStatuses:
		if e.states == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.listStatuses(ctx, scope, call.Arguments)
	case toolListTeamMembers:
		if e.users == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.listTeamMembers(ctx, scope, call.Arguments)
	case toolGetStory:
		if e.states == nil || e.users == nil || e.storyReader == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.getStory(ctx, scope, call.Arguments)
	case toolCreateStory:
		if e.mutations == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		if !scope.AllowMutations {
			return nil, ErrMutationNotAllowed
		}
		return e.mutations.proposeCreate(ctx, e, scope, call.Arguments)
	case toolUpdateStory:
		if e.mutations == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		if !scope.AllowMutations {
			return nil, ErrMutationNotAllowed
		}
		return e.mutations.proposeUpdate(ctx, e, scope, call.Arguments)
	case toolAddComment:
		if e.mutations == nil || !scope.AllowMutations {
			return nil, ErrMutationNotAllowed
		}
		return e.mutations.proposeComment(ctx, e, scope, call.Arguments)
	case toolAddRelationship:
		if e.mutations == nil || !scope.AllowMutations {
			return nil, ErrMutationNotAllowed
		}
		return e.mutations.proposeRelationship(ctx, e, scope, call.Arguments)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
	}
}

func (e *FortyOneToolExecutor) listTeams(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct{}
	if err := decodeToolArguments(raw, &args); err != nil {
		return nil, err
	}

	joined, _, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	result := listTeamsResult{
		Total: len(joined),
		Teams: make([]teamResult, 0, min(len(joined), maxToolLimit)),
	}
	for _, team := range joined {
		if len(result.Teams) == maxToolLimit {
			result.Truncated = true
			break
		}
		result.Teams = append(result.Teams, teamResult{
			ID:             team.ID,
			Name:           team.Name,
			Code:           team.Code,
			IsPrivate:      team.IsPrivate,
			MemberCount:    team.MemberCount,
			SprintsEnabled: team.SprintsEnabled,
		})
	}
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) listMyTasks(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Limit *int `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	items, err := e.stories.MyStories(ctx, scope.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("list my tasks: %w", err)
	}

	var statusesByID map[uuid.UUID]states.CoreState
	if e.states != nil {
		_, statusesByID, err = e.scopedStatuses(ctx, scope, joinedByID)
		if err != nil {
			return nil, err
		}
	}

	var assigneeName string
	var assigneeUsername string
	if e.users != nil {
		currentUser, getUserErr := e.users.GetUser(ctx, scope.UserID)
		if getUserErr != nil {
			return nil, fmt.Errorf("load current user for task enrichment: %w", getUserErr)
		}
		if currentUser.ID != scope.UserID || !currentUser.IsActive || currentUser.IsSystem {
			return nil, errors.New("load current user for task enrichment: unexpected inactive or mismatched user")
		}
		assigneeName = memberDisplayName(currentUser)
		assigneeUsername = strings.TrimSpace(currentUser.Username)
	}

	filtered := make([]taskResult, 0, min(len(items), limit))
	total := 0
	for _, story := range items {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || story.Assignee == nil || *story.Assignee != scope.UserID {
			continue
		}
		if story.CompletedAt != nil || story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}
		var statusName string
		var statusCategory string
		if story.Status != nil && statusesByID != nil {
			if status, visible := statusesByID[*story.Status]; visible {
				statusName = status.Name
				statusCategory = status.Category
				if statusIsClosed(status.Category) {
					continue
				}
			}
		}
		total++
		if len(filtered) == limit {
			continue
		}
		filtered = append(filtered, taskResult{
			ID:               story.ID,
			Reference:        storyReference(team.Code, story.SequenceID),
			URL:              storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:            story.Title,
			TeamID:           story.Team,
			TeamName:         team.Name,
			TeamCode:         strings.ToUpper(strings.TrimSpace(team.Code)),
			StatusID:         story.Status,
			StatusName:       statusName,
			StatusCategory:   statusCategory,
			AssigneeID:       story.Assignee,
			AssigneeName:     assigneeName,
			AssigneeUsername: assigneeUsername,
			Priority:         story.Priority,
			EndDate:          story.EndDate,
			UpdatedAt:        story.UpdatedAt,
		})
	}

	return marshalToolResult(listTasksResult{
		Total:     total,
		Truncated: total > len(filtered),
		Tasks:     filtered,
	})
}

func (e *FortyOneToolExecutor) searchWork(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Query  string  `json:"query"`
		TeamID *string `json:"team_id"`
		Kind   *string `json:"kind"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "query", "team_id", "kind", "limit"); err != nil {
		return nil, err
	}
	query := strings.TrimSpace(args.Query)
	if query == "" || len([]rune(query)) > maxSearchRunes {
		return nil, fmt.Errorf("%w: query must contain 1-%d characters", ErrInvalidToolArguments, maxSearchRunes)
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}

	searchType := search.SearchTypeAll
	if args.Kind != nil {
		switch *args.Kind {
		case "all":
			searchType = search.SearchTypeAll
		case "stories":
			searchType = search.SearchTypeStories
		case "objectives":
			searchType = search.SearchTypeObjectives
		default:
			return nil, fmt.Errorf("%w: unsupported search kind %q", ErrInvalidToolArguments, *args.Kind)
		}
	}

	result, err := e.search.Search(ctx, scope.WorkspaceID, scope.UserID, search.SearchParams{
		Type:     searchType,
		Query:    query,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search work: %w", err)
	}

	storiesResult := make([]searchStoryResult, 0, min(len(result.Stories), limit))
	for _, story := range result.Stories {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || len(storiesResult) == limit {
			continue
		}
		storiesResult = append(storiesResult, searchStoryResult{
			ID:        story.ID,
			Reference: storyReference(team.Code, story.SequenceID),
			URL:       storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:     story.Title,
			TeamID:    story.Team,
			StatusID:  story.Status,
			Priority:  story.Priority,
			UpdatedAt: story.UpdatedAt,
		})
	}

	objectivesResult := make([]searchObjectiveResult, 0, min(len(result.Objectives), limit))
	for _, objective := range result.Objectives {
		if _, allowed := joinedByID[objective.Team]; !allowed || objective.Workspace != scope.WorkspaceID || len(objectivesResult) == limit {
			continue
		}
		objectivesResult = append(objectivesResult, searchObjectiveResult{
			ID:           objective.ID,
			Name:         objective.Name,
			ShortSummary: objective.ShortSummary,
			TeamID:       objective.Team,
			StatusID:     objective.Status,
			Priority:     objective.Priority,
			Health:       objective.Health,
			UpdatedAt:    objective.UpdatedAt,
		})
	}

	return marshalToolResult(searchWorkResult{
		Stories:    storiesResult,
		Objectives: objectivesResult,
	})
}

func (e *FortyOneToolExecutor) listObjectives(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID *string `json:"team_id"`
		Query  *string `json:"query"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "query", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}

	filters := map[string]any{"limit": limit}
	if teamID != nil {
		filters["team_id"] = *teamID
	}
	if args.Query != nil {
		query := strings.TrimSpace(*args.Query)
		if len([]rune(query)) > maxSearchRunes {
			return nil, fmt.Errorf("%w: query must not exceed %d characters", ErrInvalidToolArguments, maxSearchRunes)
		}
		if query != "" {
			filters["search"] = query
		}
	}

	items, err := e.objectives.List(ctx, scope.WorkspaceID, scope.UserID, filters)
	if err != nil {
		return nil, fmt.Errorf("list objectives: %w", err)
	}
	result := make([]objectiveResult, 0, min(len(items), limit))
	for _, objective := range items {
		if _, allowed := joinedByID[objective.Team]; !allowed || objective.Workspace != scope.WorkspaceID || len(result) == limit {
			continue
		}
		var health *string
		if objective.Health != nil {
			value := string(*objective.Health)
			health = &value
		}
		result = append(result, objectiveResult{
			ID:               objective.ID,
			SequenceID:       objective.SequenceID,
			Name:             objective.Name,
			ShortSummary:     objective.ShortSummary,
			TeamID:           objective.Team,
			StatusID:         objective.Status,
			Priority:         objective.Priority,
			Health:           health,
			StartDate:        objective.StartDate,
			EndDate:          objective.EndDate,
			KeyResultCount:   objective.KeyResultCount,
			TotalStories:     objective.TotalStories,
			CompletedStories: objective.CompletedStories,
			UpdatedAt:        objective.UpdatedAt,
		})
	}
	return marshalToolResult(listObjectivesResult{
		Count:      len(result),
		Objectives: result,
	})
}

func (e *FortyOneToolExecutor) listStatuses(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID *string `json:"team_id"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	if len(joinedByID) == 0 {
		return marshalToolResult(listStatusesResult{Statuses: []statusResult{}})
	}

	visible, _, err := e.scopedStatuses(ctx, scope, joinedByID)
	if err != nil {
		return nil, err
	}
	result := listStatusesResult{Statuses: make([]statusResult, 0, min(len(visible), limit))}
	for _, status := range visible {
		if teamID != nil && status.Team != *teamID {
			continue
		}
		team := joinedByID[status.Team]
		result.Total++
		if len(result.Statuses) == limit {
			continue
		}
		result.Statuses = append(result.Statuses, statusResult{
			ID:         status.ID,
			Name:       status.Name,
			Category:   status.Category,
			OrderIndex: status.OrderIndex,
			IsDefault:  status.IsDefault,
			TeamID:     team.ID,
			TeamName:   team.Name,
			TeamCode:   strings.ToUpper(strings.TrimSpace(team.Code)),
		})
	}
	result.Truncated = result.Total > len(result.Statuses)
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) listTeamMembers(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID string  `json:"team_id"`
		Query  *string `json:"query"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "query", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(&args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	team := joinedByID[*teamID]
	query := ""
	if args.Query != nil {
		query = strings.TrimSpace(*args.Query)
		if len([]rune(query)) > maxSearchRunes {
			return nil, fmt.Errorf("%w: query must not exceed %d characters", ErrInvalidToolArguments, maxSearchRunes)
		}
	}

	items, err := e.users.List(ctx, scope.WorkspaceID, users.CoreListUsersFilter{
		TeamID: teamID,
		Search: query,
		Limit:  limit + 1,
	})
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	result := listTeamMembersResult{
		TeamName: team.Name,
		TeamCode: strings.ToUpper(strings.TrimSpace(team.Code)),
		Members:  make([]teamMemberResult, 0, min(len(items), limit)),
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, member := range items {
		if member.ID == uuid.Nil || !member.IsActive || member.IsSystem {
			continue
		}
		if _, duplicate := seen[member.ID]; duplicate {
			continue
		}
		seen[member.ID] = struct{}{}
		displayName := memberDisplayName(member)
		username := strings.TrimSpace(member.Username)
		if displayName == "" && username == "" {
			continue
		}
		result.Total++
		if len(result.Members) == limit {
			continue
		}
		result.Members = append(result.Members, teamMemberResult{
			DisplayName: displayName,
			Username:    username,
			Active:      true,
			RoleTitle:   memberRoleTitle(member),
		})
	}
	result.Truncated = result.Total > len(result.Members)
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) getStory(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		StoryReference string `json:"story_reference"`
	}
	if err := decodeToolArguments(raw, &args, "story_reference"); err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	storyReference, expectedTeam, err := accessibleStoryReference(args.StoryReference, joinedByID)
	if err != nil {
		return nil, err
	}
	story, err := e.storyReader.QueryByRef(ctx, scope.WorkspaceID, storyReference)
	if err != nil {
		return nil, fmt.Errorf("get story: %w", err)
	}
	if story.Workspace != scope.WorkspaceID || story.Team != expectedTeam.ID {
		return nil, fmt.Errorf("%w: story reference resolved outside the authorized team", ErrTeamNotAccessible)
	}
	if story.DeletedAt != nil || story.ArchivedAt != nil {
		return nil, errors.New("get story: story is deleted or archived")
	}

	_, statusesByID, err := e.scopedStatuses(ctx, scope, joinedByID)
	if err != nil {
		return nil, err
	}
	var statusName string
	var statusCategory string
	if story.Status != nil {
		if status, visible := statusesByID[*story.Status]; visible {
			statusName = status.Name
			statusCategory = status.Category
		}
	}

	var assigneeName string
	var assigneeUsername string
	if story.Assignee != nil {
		member, memberErr := e.activeTeamMemberByID(ctx, scope.WorkspaceID, expectedTeam.ID, *story.Assignee)
		if memberErr != nil {
			return nil, memberErr
		}
		if member != nil {
			assigneeName = memberDisplayName(*member)
			assigneeUsername = strings.TrimSpace(member.Username)
		}
	}
	description, descriptionTruncated := boundedOptionalString(story.Description, maxStoryDescriptionRunes)
	var sprintName *string
	if story.SprintSummary != nil {
		name := strings.TrimSpace(story.SprintSummary.Name)
		if name != "" {
			sprintName = &name
		}
	}

	return marshalToolResult(storyDetailsResult{
		ID:                   story.ID,
		Reference:            storyReference,
		URL:                  storyURL(scope, storyReference),
		Title:                story.Title,
		Description:          description,
		DescriptionTruncated: descriptionTruncated,
		TeamID:               expectedTeam.ID,
		TeamName:             expectedTeam.Name,
		TeamCode:             strings.ToUpper(strings.TrimSpace(expectedTeam.Code)),
		StatusID:             story.Status,
		StatusName:           statusName,
		StatusCategory:       statusCategory,
		AssigneeID:           story.Assignee,
		AssigneeName:         assigneeName,
		AssigneeUsername:     assigneeUsername,
		Priority:             story.Priority,
		EstimateLabel:        story.EstimateLabel,
		EstimateValue:        story.EstimateValue,
		SprintName:           sprintName,
		StartDate:            story.StartDate,
		EndDate:              story.EndDate,
		CompletedAt:          story.CompletedAt,
		UpdatedAt:            story.UpdatedAt,
	})
}

func (e *FortyOneToolExecutor) scopedStatuses(
	ctx context.Context,
	scope ToolScope,
	joinedByID map[uuid.UUID]teams.CoreTeam,
) ([]states.CoreState, map[uuid.UUID]states.CoreState, error) {
	items, err := e.states.List(ctx, scope.WorkspaceID, scope.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("list statuses: %w", err)
	}
	visible := make([]states.CoreState, 0, len(items))
	byID := make(map[uuid.UUID]states.CoreState, len(items))
	for _, status := range items {
		if status.ID == uuid.Nil || status.Workspace != scope.WorkspaceID {
			continue
		}
		if _, allowed := joinedByID[status.Team]; !allowed {
			continue
		}
		if _, duplicate := byID[status.ID]; duplicate {
			continue
		}
		status.Name = strings.TrimSpace(status.Name)
		status.Category = strings.TrimSpace(status.Category)
		if status.Name == "" {
			continue
		}
		visible = append(visible, status)
		byID[status.ID] = status
	}
	return visible, byID, nil
}

func (e *FortyOneToolExecutor) activeTeamMemberByID(
	ctx context.Context,
	workspaceID, teamID, memberID uuid.UUID,
) (*users.CoreUser, error) {
	items, err := e.users.List(ctx, workspaceID, users.CoreListUsersFilter{TeamID: &teamID})
	if err != nil {
		return nil, fmt.Errorf("load story assignee: %w", err)
	}
	for _, member := range items {
		if member.ID == memberID && member.IsActive && !member.IsSystem {
			copy := member
			return &copy, nil
		}
	}
	return nil, nil
}

func accessibleStoryReference(raw string, joinedByID map[uuid.UUID]teams.CoreTeam) (string, teams.CoreTeam, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || len([]rune(trimmed)) > maxStoryReferenceRunes {
		return "", teams.CoreTeam{}, fmt.Errorf("%w: story_reference must contain 1-%d characters", ErrInvalidToolArguments, maxStoryReferenceRunes)
	}
	compact := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(trimmed, " ", ""), "-", ""))
	digitIndex := -1
	for index, char := range compact {
		if char >= '0' && char <= '9' {
			digitIndex = index
			break
		}
	}
	if digitIndex < 1 || digitIndex == len(compact) {
		return "", teams.CoreTeam{}, fmt.Errorf("%w: story_reference must look like WEB-123", ErrInvalidToolArguments)
	}
	teamCode := compact[:digitIndex]
	sequenceText := compact[digitIndex:]
	for _, char := range sequenceText {
		if char < '0' || char > '9' {
			return "", teams.CoreTeam{}, fmt.Errorf("%w: story_reference must look like WEB-123", ErrInvalidToolArguments)
		}
	}
	sequenceID, err := strconv.Atoi(sequenceText)
	if err != nil || sequenceID < 1 {
		return "", teams.CoreTeam{}, fmt.Errorf("%w: story_reference must contain a positive sequence number", ErrInvalidToolArguments)
	}
	for _, team := range joinedByID {
		normalizedCode := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(team.Code), "-", ""))
		if normalizedCode == teamCode {
			return fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(team.Code)), sequenceID), team, nil
		}
	}
	return "", teams.CoreTeam{}, fmt.Errorf("%w: team code %s", ErrTeamNotAccessible, teamCode)
}

func statusIsClosed(category string) bool {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "completed", "cancelled":
		return true
	default:
		return false
	}
}

func memberDisplayName(member users.CoreUser) string {
	if displayName := strings.TrimSpace(member.FullName); displayName != "" {
		return displayName
	}
	return strings.TrimSpace(member.Username)
}

func memberRoleTitle(member users.CoreUser) string {
	if roleTitle := strings.TrimSpace(member.TeamAIRoleTitle); roleTitle != "" {
		return roleTitle
	}
	return strings.TrimSpace(member.InferredTeamAIRoleTitle)
}

func boundedOptionalString(value *string, maximumRunes int) (*string, bool) {
	if value == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, false
	}
	runes := []rune(trimmed)
	if len(runes) <= maximumRunes {
		return &trimmed, false
	}
	truncated := string(runes[:maximumRunes])
	return &truncated, true
}

func (e *FortyOneToolExecutor) joinedTeams(ctx context.Context, scope ToolScope) ([]teams.CoreTeam, map[uuid.UUID]teams.CoreTeam, error) {
	items, err := e.teams.List(ctx, scope.WorkspaceID, scope.UserID, teams.CoreListTeamsFilter{JoinedOnly: true})
	if err != nil {
		return nil, nil, fmt.Errorf("list joined teams: %w", err)
	}
	joined := make([]teams.CoreTeam, 0, len(items))
	joinedByID := make(map[uuid.UUID]teams.CoreTeam, len(items))
	var allowedTeamIDs map[uuid.UUID]struct{}
	if scope.AllowedTeamIDs != nil {
		allowedTeamIDs = make(map[uuid.UUID]struct{}, len(scope.AllowedTeamIDs))
		for _, teamID := range scope.AllowedTeamIDs {
			if teamID != uuid.Nil {
				allowedTeamIDs[teamID] = struct{}{}
			}
		}
	}
	for _, team := range items {
		if team.ID == uuid.Nil || team.Workspace != scope.WorkspaceID {
			continue
		}
		if allowedTeamIDs != nil {
			if _, allowed := allowedTeamIDs[team.ID]; !allowed {
				continue
			}
		}
		if _, duplicate := joinedByID[team.ID]; duplicate {
			continue
		}
		joined = append(joined, team)
		joinedByID[team.ID] = team
	}
	return joined, joinedByID, nil
}

func accessibleTeamID(raw *string, joined map[uuid.UUID]teams.CoreTeam) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	teamID, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil || teamID == uuid.Nil {
		return nil, fmt.Errorf("%w: team_id must be a UUID or null", ErrInvalidToolArguments)
	}
	if _, ok := joined[teamID]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return &teamID, nil
}

func normalizedLimit(value *int) (int, error) {
	if value == nil {
		return defaultToolLimit, nil
	}
	if *value < 1 || *value > maxToolLimit {
		return 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidToolArguments, maxToolLimit)
	}
	return *value, nil
}

func decodeToolArguments(raw json.RawMessage, target any, required ...string) error {
	if len(raw) == 0 {
		return fmt.Errorf("%w: arguments are required", ErrInvalidToolArguments)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return fmt.Errorf("%w: arguments must be a JSON object", ErrInvalidToolArguments)
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%w: missing %s", ErrInvalidToolArguments, key)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: arguments contain trailing data", ErrInvalidToolArguments)
	}
	return nil
}

func marshalToolResult(value any) (json.RawMessage, error) {
	result, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}
	return result, nil
}

func storyReference(teamCode string, sequenceID int) string {
	teamCode = strings.ToUpper(strings.TrimSpace(teamCode))
	if teamCode == "" {
		return fmt.Sprintf("#%d", sequenceID)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID)
}

func fortyOneToolDefinitions() []ToolDefinition {
	nullableLimit := func() map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": "Maximum number of results, or null for the default.",
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}
	nullableTeamID := func() map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": "A team UUID returned by list_teams, or null for all joined teams.",
		}
	}

	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListTeams,
			Description: "List only the FortyOne teams the current user has joined.",
			Strict:      true,
			Parameters:  strictObjectSchema(map[string]any{}, []string{}),
		},
		{
			Type:        "function",
			Name:        toolListMyTasks,
			Description: "List active tasks assigned to the current user across only their joined teams. Completed, cancelled, deleted, and archived work is excluded.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"limit": nullableLimit(),
			}, []string{"limit"}),
		},
		{
			Type:        "function",
			Name:        toolSearchWork,
			Description: "Search task and objective titles within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Plain-language text to search for.",
					"minLength":   1,
					"maxLength":   maxSearchRunes,
				},
				"team_id": nullableTeamID(),
				"kind": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Limit the search to stories or objectives, or null/all for both.",
					"enum":        []any{"all", "stories", "objectives", nil},
				},
				"limit": nullableLimit(),
			}, []string{"query", "team_id", "kind", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolListObjectives,
			Description: "List objectives within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": nullableTeamID(),
				"query": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Optional objective-name search text, or null.",
					"maxLength":   maxSearchRunes,
				},
				"limit": nullableLimit(),
			}, []string{"team_id", "query", "limit"}),
		},
	}
}

func operationalToolDefinitions(includeStoryReader bool) []ToolDefinition {
	nullableLimit := func() map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": "Maximum number of results, or null for the default.",
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}
	nullableTeamID := func() map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": "A team UUID returned by list_teams, or null for every team visible in this conversation.",
		}
	}

	definitions := []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListStatuses,
			Description: "List human-readable task statuses from only the teams visible to the current user in this conversation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": nullableTeamID(),
				"limit":   nullableLimit(),
			}, []string{"team_id", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolListTeamMembers,
			Description: "List active human members of one team visible to the current user in this conversation. Results contain no email addresses.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"query": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Optional display-name or username search text, or null.",
					"maxLength":   maxSearchRunes,
				},
				"limit": nullableLimit(),
			}, []string{"team_id", "query", "limit"}),
		},
	}
	if includeStoryReader {
		definitions = append(definitions, ToolDefinition{
			Type:        "function",
			Name:        toolGetStory,
			Description: "Get current details for one active task by its human-readable reference, such as WEB-123, only when its team is visible in this conversation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"story_reference": map[string]any{
					"type":        "string",
					"description": "An exact human-readable task reference such as WEB-123.",
					"minLength":   1,
					"maxLength":   maxStoryReferenceRunes,
				},
			}, []string{"story_reference"}),
		})
	}
	return definitions
}

func strictObjectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

type teamResult struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	IsPrivate      bool      `json:"is_private"`
	MemberCount    int       `json:"member_count"`
	SprintsEnabled bool      `json:"sprints_enabled"`
}

type listTeamsResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Teams     []teamResult `json:"teams"`
}

type taskResult struct {
	ID               uuid.UUID  `json:"id"`
	Reference        string     `json:"reference"`
	URL              string     `json:"url,omitempty"`
	Title            string     `json:"title"`
	TeamID           uuid.UUID  `json:"team_id"`
	TeamName         string     `json:"team_name"`
	TeamCode         string     `json:"team_code"`
	StatusID         *uuid.UUID `json:"status_id"`
	StatusName       string     `json:"status_name"`
	StatusCategory   string     `json:"status_category"`
	AssigneeID       *uuid.UUID `json:"assignee_id"`
	AssigneeName     string     `json:"assignee_name"`
	AssigneeUsername string     `json:"assignee_username"`
	Priority         string     `json:"priority"`
	EndDate          *time.Time `json:"end_date"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type listTasksResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Tasks     []taskResult `json:"tasks"`
}

type searchStoryResult struct {
	ID        uuid.UUID  `json:"id"`
	Reference string     `json:"reference"`
	URL       string     `json:"url,omitempty"`
	Title     string     `json:"title"`
	TeamID    uuid.UUID  `json:"team_id"`
	StatusID  *uuid.UUID `json:"status_id"`
	Priority  string     `json:"priority"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func storyURL(scope ToolScope, reference string) string {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(scope.WebsiteURL), "/"))
	if err != nil || base.Hostname() == "" || strings.TrimSpace(scope.WorkspaceSlug) == "" || strings.TrimSpace(reference) == "" {
		return ""
	}
	base.Path = path.Join("/", scope.WorkspaceSlug, "work", reference)
	if !strings.EqualFold(base.Hostname(), "localhost") && !strings.EqualFold(base.Hostname(), "127.0.0.1") {
		base.Path = path.Join("/", "work", reference)
		if !strings.HasPrefix(base.Hostname(), scope.WorkspaceSlug+".") {
			base.Host = scope.WorkspaceSlug + "." + base.Host
		}
	}
	return base.String()
}

type searchObjectiveResult struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ShortSummary *string   `json:"short_summary"`
	TeamID       uuid.UUID `json:"team_id"`
	StatusID     uuid.UUID `json:"status_id"`
	Priority     *string   `json:"priority"`
	Health       *string   `json:"health"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type searchWorkResult struct {
	Stories    []searchStoryResult     `json:"stories"`
	Objectives []searchObjectiveResult `json:"objectives"`
}

type objectiveResult struct {
	ID               uuid.UUID  `json:"id"`
	SequenceID       int        `json:"sequence_id"`
	Name             string     `json:"name"`
	ShortSummary     *string    `json:"short_summary"`
	TeamID           uuid.UUID  `json:"team_id"`
	StatusID         uuid.UUID  `json:"status_id"`
	Priority         *string    `json:"priority"`
	Health           *string    `json:"health"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	KeyResultCount   int        `json:"key_result_count"`
	TotalStories     int        `json:"total_stories"`
	CompletedStories int        `json:"completed_stories"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type listObjectivesResult struct {
	Count      int               `json:"count"`
	Objectives []objectiveResult `json:"objectives"`
}

type statusResult struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	OrderIndex int       `json:"order_index"`
	IsDefault  bool      `json:"is_default"`
	TeamID     uuid.UUID `json:"team_id"`
	TeamName   string    `json:"team_name"`
	TeamCode   string    `json:"team_code"`
}

type listStatusesResult struct {
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated"`
	Statuses  []statusResult `json:"statuses"`
}

type teamMemberResult struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Active      bool   `json:"active"`
	RoleTitle   string `json:"role_title"`
}

type listTeamMembersResult struct {
	TeamName  string             `json:"team_name"`
	TeamCode  string             `json:"team_code"`
	Total     int                `json:"total"`
	Truncated bool               `json:"truncated"`
	Members   []teamMemberResult `json:"members"`
}

type storyDetailsResult struct {
	ID                   uuid.UUID  `json:"id"`
	Reference            string     `json:"reference"`
	URL                  string     `json:"url,omitempty"`
	Title                string     `json:"title"`
	Description          *string    `json:"description"`
	DescriptionTruncated bool       `json:"description_truncated"`
	TeamID               uuid.UUID  `json:"team_id"`
	TeamName             string     `json:"team_name"`
	TeamCode             string     `json:"team_code"`
	StatusID             *uuid.UUID `json:"status_id"`
	StatusName           string     `json:"status_name"`
	StatusCategory       string     `json:"status_category"`
	AssigneeID           *uuid.UUID `json:"assignee_id"`
	AssigneeName         string     `json:"assignee_name"`
	AssigneeUsername     string     `json:"assignee_username"`
	Priority             string     `json:"priority"`
	EstimateLabel        *string    `json:"estimate_label"`
	EstimateValue        *int16     `json:"estimate_value"`
	SprintName           *string    `json:"sprint_name"`
	StartDate            *time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date"`
	CompletedAt          *time.Time `json:"completed_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
