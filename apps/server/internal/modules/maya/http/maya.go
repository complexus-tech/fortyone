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

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
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
	db       *sqlx.DB
	log      *logger.Logger
	cache    *cache.Service
	service  *maya.Service
	stories  *stories.Service
	states   *states.Service
	teams    *teams.Service
	aiAPIKey string
	baseURL  string
	client   *http.Client
	now      func() time.Time
}

func New(db *sqlx.DB, log *logger.Logger, cacheService *cache.Service, service *maya.Service, storiesService *stories.Service, statesService *states.Service, teamsService *teams.Service, aiAPIKey string) *Handlers {
	return &Handlers{
		db:       db,
		log:      log,
		cache:    cacheService,
		service:  service,
		stories:  storiesService,
		states:   statesService,
		teams:    teamsService,
		aiAPIKey: strings.TrimSpace(aiAPIKey),
		baseURL:  defaultRealtimeBaseURL,
		client:   &http.Client{Timeout: 20 * time.Second},
		now:      time.Now,
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

	session, err := h.createRealtimeClientSecret(ctx, userID)
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
	if h.stories == nil || h.states == nil || h.teams == nil {
		return web.RespondError(ctx, w, ErrMayaRealtimeToolNotConfigured, http.StatusServiceUnavailable)
	}

	var req AppRealtimeToolRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var result AppRealtimeToolResponse
	switch req.Name {
	case "list_my_tasks":
		result, err = h.executeListMyTasks(ctx, workspace.ID, userID, req.Arguments)
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

	message := "No assigned stories matched the request."
	if len(voiceStories) == 1 {
		message = "Found 1 assigned story."
	} else if len(voiceStories) > 1 {
		message = fmt.Sprintf("Found %d assigned stories.", len(voiceStories))
	}

	return AppRealtimeToolResponse{
		Success: true,
		Stories: voiceStories,
		Count:   len(voiceStories),
		Message: message,
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
	if args.Title == "" {
		return AppRealtimeToolResponse{
			Success: false,
			Message: "Ask the user for the task title before creating it.",
		}, nil
	}
	if !args.Confirmed {
		return AppRealtimeToolResponse{
			Success:              false,
			RequiresConfirmation: true,
			Message:              fmt.Sprintf("Ask the user to confirm before creating %q.", args.Title),
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
			Success:   false,
			NeedsTeam: true,
			Teams:     toRealtimeVoiceTeams(workspaceTeams),
			Message:   "Ask the user which team this story should be created in before creating it.",
		}, nil
	}

	statuses, err := h.states.TeamList(ctx, workspaceID, team.ID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list team statuses: %w", err)
	}
	status := findDefaultRealtimeStatus(statuses)
	if status == nil {
		return AppRealtimeToolResponse{
			Success: false,
			Error:   fmt.Sprintf("No statuses are configured for %s.", team.Name),
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
		Success: true,
		Story:   &voiceStory,
		Message: fmt.Sprintf("Created %q in %s.", story.Title, team.Name),
	}, nil
}

func (h *Handlers) createRealtimeClientSecret(ctx context.Context, userID uuid.UUID) (AppRealtimeSession, error) {
	payload := openAIRealtimeClientSecretRequest{
		Session: openAIRealtimeSessionConfig{
			Type:         "realtime",
			Model:        defaultRealtimeModel,
			Instructions: realtimeInstructions(),
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

func realtimeInstructions() string {
	return strings.Join([]string{
		"You are Maya, the project management assistant inside FortyOne.",
		"Your job is to help users manage work in FortyOne: stories, tasks, teams, priorities, assignments, workload, activity, and workspace insights.",
		"In voice mode, be concise, natural, and direct. Prefer one to three spoken sentences unless the user asks for detail.",
		"Stay focused on project management inside FortyOne. Briefly redirect off-topic requests back to project-management help.",
		"Use available tools whenever facts, permissions, current state, IDs, or state changes are involved.",
		"Treat the user's words task and tasks as FortyOne stories unless the user clearly means something else.",
		"Use list_my_tasks when the user asks about their assigned work, current tasks, plate, priorities, deadlines, overdue work, or what to focus on.",
		"Use create_task when the user asks you to create a task or story.",
		"Do not guess teams, statuses, permissions, or results. Ask a short clarifying question when the target is ambiguous.",
		"Never expose raw UUIDs. Use human-readable names and story references.",
		"Keep tool usage internal. Do not mention tool names, parameters, or implementation details to the user.",
		"Never claim an action succeeded unless the tool result clearly shows success.",
		"For story creation: gather the title and target team if needed, draft a concise title and useful description, ask for explicit confirmation, then call create_task with confirmed=true only after the user confirms the exact task.",
		"If a tool returns requiresConfirmation or needsTeam, ask the requested clarification in plain language.",
		"If a tool fails, repeat the useful error briefly. Do not invent a fallback workflow.",
	}, " ")
}

func realtimeTools() []openAIRealtimeTool {
	return []openAIRealtimeTool{
		{
			Type:        "function",
			Name:        "list_my_tasks",
			Description: "List tasks/stories assigned to the current user in this FortyOne workspace.",
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
			Name:        "create_task",
			Description: "Create a new FortyOne task/story after the user has confirmed the exact task.",
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
		if normalizeName(teamList[i].Name) == normalizedTeamName {
			return &teamList[i]
		}
	}

	var matches []int
	for i := range teamList {
		if strings.Contains(normalizeName(teamList[i].Name), normalizedTeamName) {
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
			Name: team.Name,
			Code: team.Code,
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
