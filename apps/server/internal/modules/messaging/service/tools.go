package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"
	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	toolListTeams           = "list_teams"
	toolListMyTasks         = "list_my_tasks"
	toolListCompleted       = "list_completed_tasks"
	toolListTeamWork        = "list_team_work"
	toolSearchWork          = "search_work"
	toolListObjectives      = "list_objectives"
	toolListSprints         = "list_sprints"
	toolGetSprintSummary    = "get_sprint_summary"
	toolGetObjectiveSummary = "get_objective_summary"
	toolGetWorkloadSummary  = "get_workload_summary"
	toolListStatuses        = "list_statuses"
	toolListTeamMembers     = "list_team_members"
	toolGetStory            = "get_story"
	toolCreateStory         = "create_story"
	toolCreateStories       = "create_stories"
	toolUpdateStory         = "update_story"
	toolAddComment          = "add_story_comment"
	toolAddRelationship     = "add_story_relationship"

	defaultToolLimit         = 20
	maxToolLimit             = 50
	maxSearchRunes           = 200
	maxStoryReferenceRunes   = 64
	maxStoryDescriptionRunes = 8_000
	maxCompletedTaskDays     = 366
)

// TeamsService is the subset of the teams domain used by assistant tools.
type TeamsService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters ...teamsdomain.ListFilter) ([]teamsdomain.Team, error)
}

// StoriesService is the subset of the stories domain used by assistant tools.
type StoriesService interface {
	MyStories(ctx context.Context, workspaceID uuid.UUID) ([]storydomain.StoryList, error)
}

// CompletedStoryReaderService is the optional stories-domain surface used to
// query completed work by the authenticated user's assignment and completion
// date. Keeping this separate preserves compatibility with narrow test and
// provider adapters that only support active-work listing.
type CompletedStoryReaderService interface {
	List(ctx context.Context, workspaceID uuid.UUID, filters storydomain.StoryFilters) ([]storydomain.StoryList, error)
}

// TeamWorkStoryReaderService is the bounded stories-domain surface used for
// shared team summaries. The grouped query caps story rows per requested group
// in SQL before the messaging layer applies its global model-output cap.
type TeamWorkStoryReaderService interface {
	ListGroupedStories(ctx context.Context, query storydomain.StoryQuery) ([]storydomain.StoryGroup, error)
}

// StoryReaderService is implemented by story services that support resolving a
// human-readable FortyOne reference. The base executor remains compatible with
// narrower story implementations that only support list_my_tasks.
type StoryReaderService interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (storydomain.Story, error)
}

// StoryMutationService is the narrow stories-domain surface needed to propose
// and explicitly confirm assistant story writes.
type StoryMutationService interface {
	StoriesService
	Get(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) (storydomain.Story, error)
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (storydomain.Story, error)
	CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, story storydomain.NewStory, workspaceID uuid.UUID) (storydomain.Story, error)
	UpdateExternalUserActionIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) error
	CreateCommentExternal(ctx context.Context, actorID, workspaceID uuid.UUID, comment storydomain.NewComment) (storydomain.Comment, error)
	UpdateLabels(ctx context.Context, id, workspaceID uuid.UUID, labels []uuid.UUID) error
	AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (storydomain.StoryAssociation, error)
	FindFirstStatusByCategory(ctx context.Context, teamID, workspaceID uuid.UUID, category string) (*uuid.UUID, error)
}

// SearchService is the subset of the search domain used by assistant tools.
type SearchService interface {
	Search(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, params searchdomain.SearchParams) (searchdomain.CoreSearchResult, error)
}

// ObjectivesService is the subset of the objectives domain used by assistant tools.
type ObjectivesService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters map[string]any) ([]objectivesdomain.Objective, error)
}

// StatesService is the membership-aware subset of the statuses domain used by
// optional operational tools and result enrichment.
type StatesService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]statesdomain.State, error)
}

// UsersService is the subset of the users domain used by optional operational
// tools. List is always called with an already-authorized team filter before
// another member is returned to the model.
type UsersService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (usersdomain.User, error)
	List(ctx context.Context, workspaceID uuid.UUID, filter usersdomain.ListUsersFilter) ([]usersdomain.User, error)
}

// OperationalToolServices groups the optional read-only domain services used
// for status/member lookups and human-readable story enrichment.
type OperationalToolServices struct {
	States   StatesService
	Users    UsersService
	Workload WorkloadService
}

// PlanningToolServices groups the read-only planning services used to resolve
// sprint and objective summaries from natural-language names.
type PlanningToolServices struct {
	Sprints SprintsService
}

// The messaging capability layer centralizes its caller-owned names here so
// individual tool use cases depend only on stable domain contracts.
type (
	messagingTeam           = teamsdomain.Team
	messagingTeamFilter     = teamsdomain.ListFilter
	messagingStory          = storydomain.Story
	messagingStoryFilters   = storydomain.StoryFilters
	messagingNewStory       = storydomain.NewStory
	messagingNewComment     = storydomain.NewComment
	messagingState          = statesdomain.State
	messagingUser           = usersdomain.User
	messagingUserListFilter = usersdomain.ListUsersFilter
	workSearchParams        = searchdomain.SearchParams
)

const (
	maximumEstimatedDurationMinutes    = storydomain.MaximumEstimatedDurationMinutes
	storyAutoSchedulingStatusOff       = storydomain.AutoSchedulingStatusOff
	storyAutoSchedulingStatusScheduled = storydomain.AutoSchedulingStatusScheduled
	storyAutoSchedulingStatusAtRisk    = storydomain.AutoSchedulingStatusAtRisk
	workSearchTypeAll                  = searchdomain.SearchTypeAll
	workSearchTypeStories              = searchdomain.SearchTypeStories
	workSearchTypeObjectives           = searchdomain.SearchTypeObjectives
	workSearchSortByRelevance          = searchdomain.SortByRelevance
)

func validateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes *int) error {
	return storydomain.ValidateScheduling(estimatedDurationMinutes, minimumFocusBlockMinutes)
}

func validateStoryAutoSchedulingContract(enabled, locked bool, status string) error {
	return storydomain.ValidateStoryAutoSchedulingContract(enabled, locked, status)
}

func storyChangedError() error {
	return storydomain.ErrStoryChanged
}

func storyAutoSchedulingLockEmptyError() error {
	return storydomain.ErrAutoSchedulingLockEmpty
}

func storyFocusBlockRequiresDurationError() error {
	return storydomain.ErrFocusBlockRequiresDuration
}

// FortyOneToolExecutor exposes the deliberately small FortyOne tool catalog
// used by messaging assistants. Mutation tools only produce signed proposals;
// ConfirmStoryMutation is the sole write boundary.
type FortyOneToolExecutor struct {
	teams       TeamsService
	stories     StoriesService
	completed   CompletedStoryReaderService
	teamWork    TeamWorkStoryReaderService
	storyReader StoryReaderService
	search      SearchService
	objectives  ObjectivesService
	sprints     SprintsService
	states      StatesService
	users       UsersService
	workload    WorkloadService
	mutations   *storyMutationExecutor
	definitions []ToolDefinition
}

type fortyOneToolExecutorConfig struct {
	storyMutationSecret string
	storyMutationStore  StoryMutationConfirmationStore
	operational         *OperationalToolServices
	planning            *PlanningToolServices
}

// FortyOneToolExecutorOption configures optional capabilities without
// widening the default read-only executor.
type FortyOneToolExecutorOption func(*fortyOneToolExecutorConfig) error

// WithStoryMutations enables signed create/update proposals and explicit
// confirmation. secret must be the dedicated messaging-mutation HMAC key shared
// by every API and worker replica so confirmations survive retries and
// deployments without reusing browser-session key material.
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

// WithPlanningTools enables read-only sprint and objective summaries. The
// stories service remains the source of associated work, while Sprints scopes
// sprint resolution and analytics to the authenticated workspace member.
func WithPlanningTools(services PlanningToolServices) FortyOneToolExecutorOption {
	return func(config *fortyOneToolExecutorConfig) error {
		if services.Sprints == nil {
			return errors.New("sprints service is required for planning assistant tools")
		}
		configuredServices := services
		config.planning = &configuredServices
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
	completedReader, _ := storiesService.(CompletedStoryReaderService)
	teamWorkReader, _ := storiesService.(TeamWorkStoryReaderService)
	executor := &FortyOneToolExecutor{
		teams:       teamsService,
		stories:     storiesService,
		completed:   completedReader,
		teamWork:    teamWorkReader,
		storyReader: storyReader,
		search:      searchService,
		objectives:  objectivesService,
		definitions: fortyOneToolDefinitions(),
	}
	if config.planning != nil {
		executor.sprints = config.planning.Sprints
		executor.definitions = append(executor.definitions, planningToolDefinitions()...)
	}
	if config.operational != nil {
		executor.states = config.operational.States
		executor.users = config.operational.Users
		executor.workload = config.operational.Workload
		executor.definitions = append(executor.definitions, operationalToolDefinitions(storyReader != nil)...)
		if executor.workload != nil {
			executor.definitions = append(executor.definitions, workloadToolDefinitions()...)
		}
	}
	if completedReader != nil {
		executor.definitions = append(executor.definitions, completedTaskToolDefinitions()...)
	}
	if executor.teamWork != nil && executor.states != nil && executor.users != nil {
		executor.definitions = append(executor.definitions, teamWorkToolDefinitions()...)
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
	case toolListCompleted:
		if e.completed == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.listCompletedTasks(ctx, scope, call.Arguments)
	case toolListTeamWork:
		if e.teamWork == nil || e.states == nil || e.users == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.listTeamWork(ctx, scope, call.Arguments)
	case toolSearchWork:
		return e.searchWork(ctx, scope, call.Arguments)
	case toolListObjectives:
		return e.listObjectives(ctx, scope, call.Arguments)
	case toolListSprints:
		if e.sprints == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.listSprints(ctx, scope, call.Arguments)
	case toolGetSprintSummary:
		if e.sprints == nil || e.completed == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.getSprintSummary(ctx, scope, call.Arguments)
	case toolGetObjectiveSummary:
		if e.sprints == nil || e.completed == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.getObjectiveSummary(ctx, scope, call.Arguments)
	case toolGetWorkloadSummary:
		if e.workload == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		return e.getWorkloadSummary(ctx, scope, call.Arguments)
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
	case toolCreateStories:
		if e.mutations == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
		}
		if !scope.AllowMutations {
			return nil, ErrMutationNotAllowed
		}
		return e.mutations.proposeCreateBatch(ctx, e, scope, call.Arguments)
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
