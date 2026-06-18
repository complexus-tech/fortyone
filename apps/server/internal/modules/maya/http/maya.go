package mayahttp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/billing"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const defaultPlanningWindow = 14 * 24 * time.Hour
const defaultRealtimeBaseURL = "https://api.openai.com/v1"
const defaultRealtimeModel = "gpt-realtime-2"
const defaultRealtimeVoice = "marin"

var ErrMayaAccessRequired = errors.New("maya agent is available on paid plans and active trials")
var ErrMayaRealtimeNotConfigured = errors.New("maya realtime voice is not configured")
var ErrMayaRealtimeToolNotConfigured = errors.New("maya realtime tools are not configured")

var realtimeStoryPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type Handlers struct {
	db         *sqlx.DB
	log        *logger.Logger
	cache      *cache.Service
	service    *maya.Service
	workspaces *workspaces.Service
	stories    *stories.Service
	states     *states.Service
	teams      *teams.Service
	users      *users.Service
	objectives *objectives.Service
	keyResults *keyresults.Service
	search     *search.Service
	aiAPIKey   string
	baseURL    string
	client     *http.Client
	now        func() time.Time
}

func New(db *sqlx.DB, log *logger.Logger, cacheService *cache.Service, service *maya.Service, workspacesService *workspaces.Service, storiesService *stories.Service, statesService *states.Service, teamsService *teams.Service, usersService *users.Service, objectivesService *objectives.Service, keyResultsService *keyresults.Service, searchService *search.Service, aiAPIKey string) *Handlers {
	return &Handlers{
		db:         db,
		log:        log,
		cache:      cacheService,
		service:    service,
		workspaces: workspacesService,
		stories:    storiesService,
		states:     statesService,
		teams:      teamsService,
		users:      usersService,
		objectives: objectivesService,
		keyResults: keyResultsService,
		search:     searchService,
		aiAPIKey:   strings.TrimSpace(aiAPIKey),
		baseURL:    defaultRealtimeBaseURL,
		client:     &http.Client{Timeout: 20 * time.Second},
		now:        time.Now,
	}
}

func (h *Handlers) CreateWorkPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if ok, err := h.workspaceCanUseMaya(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	} else if !ok {
		return web.RespondError(ctx, w, ErrMayaAccessRequired, http.StatusPaymentRequired)
	}

	var req AppCreateWorkPlanRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	windowStart := h.now().UTC()
	if req.WindowStart != nil {
		windowStart = req.WindowStart.UTC()
	}
	windowEnd := windowStart.Add(defaultPlanningWindow)
	if req.WindowEnd != nil {
		windowEnd = req.WindowEnd.UTC()
	}

	plan, err := h.service.CreateWorkPlan(ctx, maya.CreateWorkPlanInput{
		WorkspaceID:      workspace.ID,
		StoryID:          req.StoryID,
		TriggeredBy:      userID,
		Trigger:          maya.RunTriggerManual,
		WindowStart:      windowStart,
		WindowEnd:        windowEnd,
		DurationMinutes:  req.DurationMinutes,
		CandidateUserIDs: req.CandidateUserIDs,
		AutoApply:        req.AutoApply,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppWorkPlan(plan), http.StatusCreated)
}

func (h *Handlers) CreateRealtimeSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if ok, err := h.workspaceCanUseMaya(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	} else if !ok {
		return web.RespondError(ctx, w, ErrMayaAccessRequired, http.StatusPaymentRequired)
	}
	if h.aiAPIKey == "" {
		return web.RespondError(ctx, w, ErrMayaRealtimeNotConfigured, http.StatusServiceUnavailable)
	}
	if h.workspaces == nil || h.teams == nil {
		return web.RespondError(ctx, w, ErrMayaRealtimeToolNotConfigured, http.StatusServiceUnavailable)
	}

	session, err := h.createRealtimeClientSecret(ctx, workspace.ID, userID)
	if err != nil {
		h.log.Error(ctx, "failed to create realtime maya session", "error", err, "workspace_id", workspace.ID, "user_id", userID)
		return web.RespondError(ctx, w, err, http.StatusBadGateway)
	}

	return web.Respond(ctx, w, session, http.StatusCreated)
}

func (h *Handlers) ExecuteRealtimeTool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if ok, err := h.workspaceCanUseMaya(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	} else if !ok {
		return web.RespondError(ctx, w, ErrMayaAccessRequired, http.StatusPaymentRequired)
	}
	if h.stories == nil || h.states == nil || h.teams == nil || h.users == nil || h.objectives == nil || h.keyResults == nil || h.search == nil {
		return web.RespondError(ctx, w, ErrMayaRealtimeToolNotConfigured, http.StatusServiceUnavailable)
	}

	var req AppRealtimeToolRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var result AppRealtimeToolResponse
	switch req.Name {
	case "get_context":
		result, err = h.executeGetContext(ctx, workspace.ID, userID)
	case "list_teams":
		result, err = h.executeListTeams(ctx, workspace.ID, userID, req.Arguments)
	case "list_team_members":
		result, err = h.executeListTeamMembers(ctx, workspace.ID, userID, req.Arguments)
	case "list_my_tasks":
		result, err = h.executeListMyTasks(ctx, workspace.ID, userID, req.Arguments)
	case "search_work":
		result, err = h.executeSearchWork(ctx, workspace.ID, userID, req.Arguments)
	case "list_objectives":
		result, err = h.executeListObjectives(ctx, workspace.ID, userID, req.Arguments)
	case "list_key_results":
		result, err = h.executeListKeyResults(ctx, workspace.ID, userID, req.Arguments)
	case "create_task":
		result, err = h.executeCreateTask(ctx, workspace.ID, userID, req.Arguments)
	default:
		result = AppRealtimeToolResponse{
			Success: false,
			Error:   fmt.Sprintf("Unsupported realtime tool: %s.", req.Name),
		}
	}
	if err != nil {
		h.log.Error(ctx, "failed to execute maya realtime tool", "tool", req.Name, "error", err, "workspace_id", workspace.ID, "user_id", userID)
		result = AppRealtimeToolResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	if result.Terminology == nil {
		terminology := h.realtimeTerminology(ctx, workspace.ID)
		result.Terminology = &terminology
	}

	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) statusCode(err error) int {
	switch {
	case errors.Is(err, maya.ErrNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, ErrMayaRealtimeNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, ErrMayaRealtimeToolNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, maya.ErrInvalidPlanInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handlers) workspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM workspaces w
			WHERE w.workspace_id = $1
				AND ` + billing.WorkspaceMayaAccessSQL("w") + `
		)
	`

	var allowed bool
	if err := h.db.GetContext(ctx, &allowed, query, workspaceID); err != nil {
		return false, fmt.Errorf("check maya access: %w", err)
	}
	return allowed, nil
}

func (h *Handlers) executeListMyTasks(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListMyTasksArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_my_tasks arguments: %w", err)
		}
	}
	if args.Limit <= 0 {
		args.Limit = 10
	}
	if args.Limit > 25 {
		args.Limit = 25
	}

	statuses, err := h.states.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list statuses: %w", err)
	}
	statusesByID := make(map[uuid.UUID]states.CoreState, len(statuses))
	for _, status := range statuses {
		statusesByID[status.ID] = status
	}

	allStories, err := h.stories.MyStories(ctx, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list my stories: %w", err)
	}

	voiceStories := make([]AppRealtimeVoiceStory, 0, min(args.Limit, len(allStories)))
	for _, story := range allStories {
		if !args.IncludeCompleted && story.CompletedAt != nil {
			continue
		}
		if story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}
		voiceStories = append(voiceStories, toRealtimeVoiceStory(story, statusesByID))
		if len(voiceStories) >= args.Limit {
			break
		}
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	message := fmt.Sprintf("No assigned %s matched the request.", terminology.Stories)
	if len(voiceStories) == 1 {
		message = fmt.Sprintf("Found 1 assigned %s.", terminology.Story)
	} else if len(voiceStories) > 1 {
		message = fmt.Sprintf("Found %d assigned %s.", len(voiceStories), terminology.Stories)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Stories:     voiceStories,
		Count:       len(voiceStories),
		Message:     message,
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeGetContext(ctx context.Context, workspaceID, userID uuid.UUID) (AppRealtimeToolResponse, error) {
	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	terminology := h.realtimeTerminology(ctx, workspaceID)

	return AppRealtimeToolResponse{
		Success:     true,
		Teams:       toRealtimeVoiceTeams(workspaceTeams),
		Count:       len(workspaceTeams),
		Message:     teamContextMessage(workspaceTeams),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeListTeams(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListTeamsArguments{Limit: 25}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_teams arguments: %w", err)
		}
	}
	limit := clampLimit(args.Limit, 25)

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID, teams.CoreListTeamsFilter{
		Search: strings.TrimSpace(args.Search),
		Limit:  limit,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Teams:       toRealtimeVoiceTeams(workspaceTeams),
		Count:       len(workspaceTeams),
		Message:     fmt.Sprintf("Found %d team%s.", len(workspaceTeams), pluralSuffix(len(workspaceTeams))),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeListTeamMembers(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListTeamMembersArguments{Limit: 25}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_team_members arguments: %w", err)
		}
	}
	limit := clampLimit(args.Limit, 25)

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
	if team == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			NeedsTeam:   true,
			Teams:       toRealtimeVoiceTeams(workspaceTeams),
			Message:     "Ask which team to list members for.",
			Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
		}, nil
	}

	members, err := h.users.List(ctx, workspaceID, users.CoreListUsersFilter{
		TeamID: &team.ID,
		Search: strings.TrimSpace(args.Search),
		Limit:  limit,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list team members: %w", err)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Members:     toRealtimeVoiceMembers(members),
		Count:       len(members),
		Message:     fmt.Sprintf("Found %d member%s in %s.", len(members), pluralSuffix(len(members)), team.Name),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeSearchWork(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeSearchArguments{Type: "all", Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid search_work arguments: %w", err)
		}
	}
	args.Query = strings.TrimSpace(args.Query)
	if args.Query == "" {
		return AppRealtimeToolResponse{
			Success: false,
			Message: "Ask what to search for.",
		}, nil
	}

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	var teamID *uuid.UUID
	if strings.TrimSpace(args.TeamName) != "" {
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to search in.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		teamID = &team.ID
	}

	searchType := search.SearchTypeAll
	switch strings.ToLower(strings.TrimSpace(args.Type)) {
	case "stories", "tasks", "issues", "work_items", "work-items", "work items":
		searchType = search.SearchTypeStories
	case "objectives", "goals", "projects":
		searchType = search.SearchTypeObjectives
	}
	searchResult, err := h.search.Search(ctx, workspaceID, userID, search.SearchParams{
		Type:     searchType,
		Query:    args.Query,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: clampLimit(args.Limit, 10),
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("search work: %w", err)
	}

	teamsByID := indexTeamsByID(workspaceTeams)
	statusesByID, err := h.statusesByID(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}

	stories := make([]AppRealtimeVoiceStory, 0, len(searchResult.Stories))
	for _, story := range searchResult.Stories {
		stories = append(stories, toRealtimeVoiceSearchStory(story, teamsByID, statusesByID))
	}
	objectives := make([]AppRealtimeVoiceObjective, 0, len(searchResult.Objectives))
	for _, objective := range searchResult.Objectives {
		objectives = append(objectives, toRealtimeVoiceSearchObjective(objective, teamsByID))
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Stories:     stories,
		Objectives:  objectives,
		Count:       searchResult.TotalStories + searchResult.TotalObjectives,
		Message:     fmt.Sprintf("Found %d matching work items.", searchResult.TotalStories+searchResult.TotalObjectives),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeListObjectives(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListObjectivesArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_objectives arguments: %w", err)
		}
	}

	filters := map[string]any{
		"limit":  clampLimit(args.Limit, 10),
		"offset": 0,
	}
	if search := strings.TrimSpace(args.Search); search != "" {
		filters["search"] = search
	}

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	if strings.TrimSpace(args.TeamName) != "" {
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to list objectives for.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		filters["team_id"] = team.ID
	}

	objectiveList, err := h.objectives.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list objectives: %w", err)
	}
	teamsByID := indexTeamsByID(workspaceTeams)
	result := make([]AppRealtimeVoiceObjective, 0, len(objectiveList))
	for _, objective := range objectiveList {
		result = append(result, toRealtimeVoiceObjective(objective, teamsByID))
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	return AppRealtimeToolResponse{
		Success:     true,
		Objectives:  result,
		Count:       len(result),
		Message:     fmt.Sprintf("Found %d %s.", len(result), termForCount(terminology.Objective, terminology.Objectives, len(result))),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeListKeyResults(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListKeyResultsArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_key_results arguments: %w", err)
		}
	}

	filters := keyresults.CoreKeyResultFilters{
		WorkspaceID:    workspaceID,
		CurrentUserID:  userID,
		Page:           1,
		PageSize:       clampLimit(args.Limit, 10),
		OrderBy:        "updated_at",
		OrderDirection: "desc",
	}

	if strings.TrimSpace(args.TeamName) != "" {
		workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
		}
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to list key results for.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		filters.TeamIDs = []uuid.UUID{team.ID}
	}

	response, err := h.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list key results: %w", err)
	}

	keyResults := make([]AppRealtimeVoiceKeyResult, 0, len(response.KeyResults))
	for _, keyResult := range response.KeyResults {
		keyResults = append(keyResults, toRealtimeVoiceKeyResult(keyResult))
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	return AppRealtimeToolResponse{
		Success:     true,
		KeyResults:  keyResults,
		Count:       len(keyResults),
		Message:     fmt.Sprintf("Found %d %s.", len(keyResults), termForCount(terminology.KeyResult, terminology.KeyResults, len(keyResults))),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeCreateTask(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeCreateTaskArguments
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid create_task arguments: %w", err)
		}
	}

	args.Title = strings.TrimSpace(args.Title)
	args.Description = strings.TrimSpace(args.Description)
	args.TeamName = strings.TrimSpace(args.TeamName)
	args.Priority = normalizePriority(args.Priority)
	terminology := h.realtimeTerminology(ctx, workspaceID)
	if args.Title == "" {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user for the %s title before creating it.", terminology.Story),
			Terminology: &terminology,
		}, nil
	}
	if !args.Confirmed {
		return AppRealtimeToolResponse{
			Success:              false,
			RequiresConfirmation: true,
			Message:              fmt.Sprintf("Ask the user to confirm before creating the %s %q.", terminology.Story, args.Title),
			Terminology:          &terminology,
			Confirmation: &AppRealtimeConfirmation{
				Title:       args.Title,
				Description: args.Description,
				TeamName:    args.TeamName,
				Priority:    args.Priority,
			},
		}, nil
	}

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
	if team == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			NeedsTeam:   true,
			Teams:       toRealtimeVoiceTeams(workspaceTeams),
			Message:     fmt.Sprintf("Ask the user which team this %s should be created in before creating it.", terminology.Story),
			Terminology: &terminology,
		}, nil
	}

	statuses, err := h.states.TeamList(ctx, workspaceID, team.ID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list team statuses: %w", err)
	}
	status := findDefaultRealtimeStatus(statuses)
	if status == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Error:       fmt.Sprintf("No statuses are configured for %s.", team.Name),
			Terminology: &terminology,
		}, nil
	}

	description := args.Description
	descriptionHTML := ""
	if description != "" {
		descriptionHTML = "<p>" + html.EscapeString(description) + "</p>"
	}

	story, err := h.stories.Create(ctx, stories.CoreNewStory{
		Title:           args.Title,
		Description:     optionalString(description),
		DescriptionHTML: optionalString(descriptionHTML),
		Status:          &status.ID,
		Reporter:        &userID,
		Priority:        args.Priority,
		Team:            team.ID,
	}, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("create story: %w", err)
	}

	h.invalidateStoryListCaches(ctx, workspaceID)

	voiceStory := toRealtimeVoiceCreatedStory(story, *team, *status)
	return AppRealtimeToolResponse{
		Success:     true,
		Story:       &voiceStory,
		Message:     fmt.Sprintf("Created the %s %q in %s.", terminology.Story, story.Title, team.Name),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) createRealtimeClientSecret(ctx context.Context, workspaceID, userID uuid.UUID) (AppRealtimeSession, error) {
	terminology := h.realtimeTerminology(ctx, workspaceID)
	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("list teams for realtime context: %w", err)
	}

	payload := openAIRealtimeClientSecretRequest{
		Session: openAIRealtimeSessionConfig{
			Type:         "realtime",
			Model:        defaultRealtimeModel,
			Instructions: realtimeInstructions(terminology, workspaceTeams),
			Tools:        realtimeTools(),
			ToolChoice:   "auto",
			Audio: openAIRealtimeAudioConfig{
				Output: openAIRealtimeAudioOutputConfig{
					Voice: defaultRealtimeVoice,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("marshal realtime session request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(h.baseURL, "/")+"/realtime/client_secrets",
		bytes.NewReader(body),
	)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("create realtime session request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.aiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Safety-Identifier", safetyIdentifier(userID))

	res, err := h.client.Do(req)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("call realtime session endpoint: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("read realtime session response: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return AppRealtimeSession{}, fmt.Errorf("realtime session endpoint returned %s: %s", res.Status, strings.TrimSpace(string(data)))
	}

	var response openAIRealtimeClientSecretResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return AppRealtimeSession{}, fmt.Errorf("decode realtime session response: %w", err)
	}

	clientSecret := strings.TrimSpace(response.Value)
	expiresAt := response.ExpiresAt
	if clientSecret == "" && response.ClientSecret != nil {
		clientSecret = strings.TrimSpace(response.ClientSecret.Value)
		expiresAt = response.ClientSecret.ExpiresAt
	}
	if clientSecret == "" {
		return AppRealtimeSession{}, errors.New("realtime session response did not include a client secret")
	}

	return AppRealtimeSession{
		ClientSecret: clientSecret,
		ExpiresAt:    expiresAt,
		Model:        defaultRealtimeModel,
		Voice:        defaultRealtimeVoice,
	}, nil
}

func realtimeInstructions(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam) string {
	return strings.Join([]string{
		"You are Maya, the project management assistant inside FortyOne.",
		"Your job is to help users manage work in FortyOne: work items, teams, priorities, assignments, workload, objectives, key results, activity, and workspace insights.",
		"In voice mode, be concise, natural, and direct. Prefer one to three spoken sentences unless the user asks for detail.",
		"Stay focused on project management inside FortyOne. Briefly redirect off-topic requests back to project-management help.",
		"Use available tools whenever facts, permissions, current state, IDs, or state changes are involved.",
		fmt.Sprintf("Use this workspace's preferred terminology when speaking: stories are called %q/%q, sprints are called %q/%q, objectives are called %q/%q, and key results are called %q/%q.", terminology.Story, terminology.Stories, terminology.Sprint, terminology.Sprints, terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		"Understand all common aliases even when you do not speak them back: story, task, issue, work item, objective, goal, project, key result, milestone, focus area, KPI, sprint, cycle, and iteration.",
		"Use get_context when you need current terminology or team context.",
		"Use list_my_tasks when the user asks about their assigned work, current work, plate, priorities, deadlines, overdue work, what they have today, or what to focus on.",
		"Use list_teams or list_team_members for team questions.",
		"Use search_work when the user asks to find or look up work by name, description, topic, or keyword.",
		fmt.Sprintf("Use list_objectives for %s/%s questions and list_key_results for %s/%s questions.", terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		fmt.Sprintf("Use create_task when the user asks you to create a %s, task, story, issue, or work item.", terminology.Story),
		"Do not guess teams, statuses, permissions, or results. Ask a short clarifying question when the target is ambiguous.",
		teamSelectionInstruction(workspaceTeams),
		"Never expose raw UUIDs. Use human-readable names and story references.",
		"Keep tool usage internal. Do not mention tool names, parameters, or implementation details to the user.",
		"Never claim an action succeeded unless the tool result clearly shows success.",
		fmt.Sprintf("For %s creation: gather the title and target team if needed, draft a concise title and useful description, ask for explicit confirmation, then call create_task with confirmed=true only after the user confirms the exact %s.", terminology.Story, terminology.Story),
		"If a tool returns requiresConfirmation or needsTeam, ask the requested clarification in plain language.",
		"If a tool fails, repeat the useful error briefly. Do not invent a fallback workflow.",
	}, " ")
}

func realtimeTools() []openAIRealtimeTool {
	return []openAIRealtimeTool{
		{
			Type:        "function",
			Name:        "get_context",
			Description: "Get the current workspace terminology and the teams available to the user.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_teams",
			Description: "List teams the current user belongs to. Use for team questions or to resolve team names.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"search": map[string]any{
						"type":        "string",
						"description": "Optional team name or code search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of teams to return. Defaults to 25.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_team_members",
			Description: "List members of a team. If the user belongs to exactly one team, teamName can be omitted.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"teamName": map[string]any{
						"type":        "string",
						"description": "Team name or code. Omit only when the current user has one team.",
					},
					"search": map[string]any{
						"type":        "string",
						"description": "Optional member name, username, or email search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of members to return. Defaults to 25.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_my_tasks",
			Description: "List current user's assigned work items. Understands task/story/issue/work item terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"includeCompleted": map[string]any{
						"type":        "boolean",
						"description": "Whether completed stories should be included. Defaults to false.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of stories to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "search_work",
			Description: "Search across work items and objectives by topic, name, or description. Use for find/look up questions.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The text to search for.",
					},
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"all", "stories", "objectives"},
						"description": "Content type to search. Defaults to all.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code to restrict the search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Type:        "function",
			Name:        "list_objectives",
			Description: "List objectives/goals/projects accessible to the user, respecting the workspace's selected terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"search": map[string]any{
						"type":        "string",
						"description": "Optional objective/goal/project name search.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_key_results",
			Description: "List key results/milestones/focus areas/KPIs accessible to the user, respecting workspace terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "create_task",
			Description: "Create a new FortyOne work item after the user has confirmed the exact item. Understands task/story/issue/work item terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"title": map[string]any{
						"type":        "string",
						"description": "The story title.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Optional story description.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "The target team name when the workspace has more than one team or the user mentioned a team.",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"No Priority", "Low", "Medium", "High", "Urgent"},
						"description": "Optional story priority. Defaults to No Priority.",
					},
					"confirmed": map[string]any{
						"type":        "boolean",
						"description": "True only after the user explicitly confirms creating this story.",
					},
				},
				"required": []string{"title", "confirmed"},
			},
		},
	}
}

func (h *Handlers) realtimeTerminology(ctx context.Context, workspaceID uuid.UUID) AppRealtimeTerminology {
	settings := workspaces.CoreWorkspaceSettings{
		StoryTerm:     "story",
		SprintTerm:    "sprint",
		ObjectiveTerm: "objective",
		KeyResultTerm: "key result",
	}
	if h.workspaces != nil {
		if fetched, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspaceID); err == nil {
			settings = fetched
		}
	}

	return AppRealtimeTerminology{
		Story:      settings.StoryTerm,
		Stories:    pluralizeTerm(settings.StoryTerm),
		Sprint:     settings.SprintTerm,
		Sprints:    pluralizeTerm(settings.SprintTerm),
		Objective:  settings.ObjectiveTerm,
		Objectives: pluralizeTerm(settings.ObjectiveTerm),
		KeyResult:  settings.KeyResultTerm,
		KeyResults: pluralizeTerm(settings.KeyResultTerm),
	}
}

func pluralizeTerm(term string) string {
	term = strings.TrimSpace(term)
	if term == "" {
		return term
	}
	if strings.HasSuffix(term, "y") {
		return strings.TrimSuffix(term, "y") + "ies"
	}
	if term == "focus area" {
		return "focus areas"
	}
	return term + "s"
}

func termForCount(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func teamContextMessage(teamList []teams.CoreTeam) string {
	switch len(teamList) {
	case 0:
		return "The user does not belong to any teams."
	case 1:
		return fmt.Sprintf("The user belongs to one team: %s. Default to this team when team selection is needed.", teamList[0].Name)
	default:
		return fmt.Sprintf("The user belongs to %d teams. Resolve close team name matches by name or code, and ask only when ambiguous.", len(teamList))
	}
}

func teamSelectionInstruction(teamList []teams.CoreTeam) string {
	switch len(teamList) {
	case 0:
		return "The user currently belongs to no teams, so team-scoped actions may not be possible."
	case 1:
		return fmt.Sprintf("If team selection is needed and the user does not specify a team, use %s.", teamList[0].Name)
	default:
		return fmt.Sprintf("The user belongs to these teams: %s. Resolve close team matches by name or code; ask which team only when the match is missing or ambiguous.", strings.Join(teamNames(teamList), ", "))
	}
}

func teamNames(teamList []teams.CoreTeam) []string {
	names := make([]string, len(teamList))
	for i, team := range teamList {
		if strings.TrimSpace(team.Code) == "" {
			names[i] = team.Name
			continue
		}
		names[i] = fmt.Sprintf("%s (%s)", team.Name, team.Code)
	}
	return names
}

func clampLimit(limit, fallback int) int {
	if fallback <= 0 {
		fallback = 10
	}
	if limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func ptr[T any](value T) *T {
	return &value
}

func normalizePriority(priority string) string {
	priority = strings.TrimSpace(priority)
	if priority == "" {
		return "No Priority"
	}
	if _, ok := realtimeStoryPriorities[priority]; ok {
		return priority
	}
	return "No Priority"
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func normalizeName(value string) string {
	var b strings.Builder
	lastWasSpace := true
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasSpace = false
		case unicode.IsSpace(r):
			if !lastWasSpace {
				b.WriteByte(' ')
				lastWasSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func resolveRealtimeTeam(teamList []teams.CoreTeam, teamName string) *teams.CoreTeam {
	if strings.TrimSpace(teamName) == "" {
		if len(teamList) != 1 {
			return nil
		}
		return &teamList[0]
	}

	normalizedTeamName := normalizeName(teamName)
	for i := range teamList {
		if normalizeName(teamList[i].Name) == normalizedTeamName || normalizeName(teamList[i].Code) == normalizedTeamName {
			return &teamList[i]
		}
	}

	var matches []int
	for i := range teamList {
		if strings.Contains(normalizeName(teamList[i].Name), normalizedTeamName) || strings.Contains(normalizeName(teamList[i].Code), normalizedTeamName) {
			matches = append(matches, i)
		}
	}
	if len(matches) != 1 {
		return nil
	}
	return &teamList[matches[0]]
}

func findDefaultRealtimeStatus(statuses []states.CoreState) *states.CoreState {
	if len(statuses) == 0 {
		return nil
	}
	for i := range statuses {
		if statuses[i].IsDefault {
			return &statuses[i]
		}
	}
	for _, category := range []string{"unstarted", "backlog"} {
		for i := range statuses {
			if statuses[i].Category == category {
				return &statuses[i]
			}
		}
	}
	return &statuses[0]
}

func toRealtimeVoiceTeams(teamList []teams.CoreTeam) []AppRealtimeVoiceTeam {
	result := make([]AppRealtimeVoiceTeam, len(teamList))
	for i, team := range teamList {
		result[i] = AppRealtimeVoiceTeam{
			Name:        team.Name,
			Code:        team.Code,
			MemberCount: team.MemberCount,
			IsPrivate:   team.IsPrivate,
		}
	}
	return result
}

func toRealtimeVoiceStatus(status states.CoreState) *AppRealtimeVoiceStatus {
	return &AppRealtimeVoiceStatus{
		Name:     status.Name,
		Category: status.Category,
	}
}

func storyReference(teamCode string, sequenceID int) string {
	teamCode = strings.TrimSpace(teamCode)
	if teamCode == "" || sequenceID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID)
}

func toRealtimeVoiceStory(story stories.CoreStoryList, statusesByID map[uuid.UUID]states.CoreState) AppRealtimeVoiceStory {
	var teamName, teamCode string
	if story.TeamSummary != nil {
		teamName = story.TeamSummary.Name
		teamCode = story.TeamSummary.Code
	}

	var status *AppRealtimeVoiceStatus
	if story.Status != nil {
		if matchedStatus, ok := statusesByID[*story.Status]; ok {
			status = toRealtimeVoiceStatus(matchedStatus)
		}
	}

	return AppRealtimeVoiceStory{
		Reference:   storyReference(teamCode, story.SequenceID),
		Title:       story.Title,
		Priority:    story.Priority,
		Team:        teamName,
		Status:      status,
		StartDate:   story.StartDate,
		EndDate:     story.EndDate,
		CompletedAt: story.CompletedAt,
	}
}

func toRealtimeVoiceCreatedStory(story stories.CoreSingleStory, team teams.CoreTeam, status states.CoreState) AppRealtimeVoiceStory {
	return AppRealtimeVoiceStory{
		Reference:   storyReference(team.Code, story.SequenceID),
		Title:       story.Title,
		Priority:    story.Priority,
		Team:        team.Name,
		Status:      toRealtimeVoiceStatus(status),
		StartDate:   story.StartDate,
		EndDate:     story.EndDate,
		CompletedAt: story.CompletedAt,
	}
}

func toRealtimeVoiceMembers(memberList []users.CoreUser) []AppRealtimeVoiceMember {
	result := make([]AppRealtimeVoiceMember, len(memberList))
	for i, member := range memberList {
		name := strings.TrimSpace(member.FullName)
		if name == "" {
			name = member.Username
		}
		roleTitle := strings.TrimSpace(member.TeamAIRoleTitle)
		if roleTitle == "" {
			roleTitle = strings.TrimSpace(member.InferredTeamAIRoleTitle)
		}
		role := ""
		if member.Role != nil {
			role = *member.Role
		}
		result[i] = AppRealtimeVoiceMember{
			Name:         name,
			Username:     member.Username,
			Role:         role,
			RoleTitle:    roleTitle,
			LastActiveAt: member.LastStoryActivityAt,
		}
	}
	return result
}

func toRealtimeVoiceObjective(objective objectives.CoreObjective, teamsByID map[uuid.UUID]teams.CoreTeam) AppRealtimeVoiceObjective {
	teamName := ""
	if team, ok := teamsByID[objective.Team]; ok {
		teamName = team.Name
	}
	health := ""
	if objective.Health != nil {
		health = string(*objective.Health)
	}
	return AppRealtimeVoiceObjective{
		Name:             objective.Name,
		Description:      objective.Description,
		Team:             teamName,
		Priority:         objective.Priority,
		Health:           health,
		StartDate:        objective.StartDate,
		EndDate:          objective.EndDate,
		TotalStories:     objective.TotalStories,
		CompletedStories: objective.CompletedStories,
	}
}

func toRealtimeVoiceKeyResult(keyResult keyresults.CoreKeyResultWithObjective) AppRealtimeVoiceKeyResult {
	return AppRealtimeVoiceKeyResult{
		Name:            keyResult.Name,
		ObjectiveName:   keyResult.ObjectiveName,
		Team:            keyResult.TeamName,
		MeasurementType: keyResult.MeasurementType,
		StartValue:      keyResult.StartValue,
		CurrentValue:    keyResult.CurrentValue,
		TargetValue:     keyResult.TargetValue,
		StartDate:       keyResult.StartDate,
		EndDate:         keyResult.EndDate,
	}
}

func toRealtimeVoiceSearchStory(story search.CoreSearchStory, teamsByID map[uuid.UUID]teams.CoreTeam, statusesByID map[uuid.UUID]states.CoreState) AppRealtimeVoiceStory {
	teamName, teamCode := "", ""
	if team, ok := teamsByID[story.Team]; ok {
		teamName = team.Name
		teamCode = team.Code
	}
	var status *AppRealtimeVoiceStatus
	if story.Status != nil {
		if matchedStatus, ok := statusesByID[*story.Status]; ok {
			status = toRealtimeVoiceStatus(matchedStatus)
		}
	}
	return AppRealtimeVoiceStory{
		Reference: storyReference(teamCode, story.SequenceID),
		Title:     story.Title,
		Priority:  story.Priority,
		Team:      teamName,
		Status:    status,
		StartDate: story.StartDate,
		EndDate:   story.EndDate,
	}
}

func toRealtimeVoiceSearchObjective(objective search.CoreSearchObjective, teamsByID map[uuid.UUID]teams.CoreTeam) AppRealtimeVoiceObjective {
	teamName := ""
	if team, ok := teamsByID[objective.Team]; ok {
		teamName = team.Name
	}
	health := ""
	if objective.Health != nil {
		health = *objective.Health
	}
	return AppRealtimeVoiceObjective{
		Name:        objective.Name,
		Description: objective.Description,
		Team:        teamName,
		Priority:    objective.Priority,
		Health:      health,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
	}
}

func indexTeamsByID(teamList []teams.CoreTeam) map[uuid.UUID]teams.CoreTeam {
	teamsByID := make(map[uuid.UUID]teams.CoreTeam, len(teamList))
	for _, team := range teamList {
		teamsByID[team.ID] = team
	}
	return teamsByID
}

func (h *Handlers) statusesByID(ctx context.Context, workspaceID, userID uuid.UUID) (map[uuid.UUID]states.CoreState, error) {
	statusList, err := h.states.List(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("list statuses: %w", err)
	}
	statusesByID := make(map[uuid.UUID]states.CoreState, len(statusList))
	for _, status := range statusList {
		statusesByID[status.ID] = status
	}
	return statusesByID, nil
}

func (h *Handlers) invalidateStoryListCaches(ctx context.Context, workspaceID uuid.UUID) {
	if h.cache == nil {
		return
	}

	storyListCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceID.String())
	h.cache.DeleteByPattern(ctx, storyListCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)
}

func safetyIdentifier(userID uuid.UUID) string {
	sum := sha256.Sum256([]byte(userID.String()))
	return hex.EncodeToString(sum[:])
}
