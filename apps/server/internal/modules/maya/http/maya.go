package mayahttp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
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
const defaultRealtimeModel = "gpt-realtime-2.1-mini"
const defaultRealtimeTranscriptionModel = "gpt-4o-mini-transcribe"
const defaultRealtimeVoice = "marin"
const realtimeMonthlyVoiceLimit = 10 * time.Minute
const realtimeMaxSessionDuration = 5 * time.Minute

var ErrMayaAccessRequired = errors.New("maya agent is available on paid plans and active trials")
var ErrMayaRealtimeNotConfigured = errors.New("maya realtime voice is not configured")
var ErrMayaRealtimeToolNotConfigured = errors.New("maya realtime tools are not configured")
var ErrMayaRealtimeMonthlyLimitExceeded = errors.New("monthly realtime voice limit reached")
var ErrMayaRealtimeSessionInactive = errors.New("realtime voice session is not active")
var ErrMayaRealtimeToolCallConflict = errors.New("realtime tool call conflicts with an existing call")
var ErrMayaRealtimeToolCallInProgress = errors.New("realtime tool call is already processing")

var realtimeStoryPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type Handlers struct {
	db            *sqlx.DB
	log           *logger.Logger
	cache         *cache.Service
	service       *maya.Service
	workspaces    *workspaces.Service
	stories       *stories.Service
	states        *states.Service
	teams         *teams.Service
	users         *users.Service
	objectives    *objectives.Service
	keyResults    *keyresults.Service
	search        *search.Service
	activities    *activities.Service
	feedback      *feedback.Service
	notifications *notifications.Service
	reports       *reports.Service
	sprints       *sprints.Service
	secretKey     string
	aiAPIKey      string
	baseURL       string
	client        *http.Client
	now           func() time.Time
}

func New(cfg Config) *Handlers {
	return &Handlers{
		db:            cfg.DB,
		log:           cfg.Log,
		cache:         cfg.Cache,
		service:       cfg.Service,
		workspaces:    cfg.Workspaces,
		stories:       cfg.Stories,
		states:        cfg.States,
		teams:         cfg.Teams,
		users:         cfg.Users,
		objectives:    cfg.Objectives,
		keyResults:    cfg.KeyResults,
		search:        cfg.Search,
		activities:    cfg.Activities,
		feedback:      cfg.Feedback,
		notifications: cfg.Notifications,
		reports:       cfg.Reports,
		sprints:       cfg.Sprints,
		secretKey:     cfg.SecretKey,
		aiAPIKey:      strings.TrimSpace(cfg.AIAPIKey),
		baseURL:       defaultRealtimeBaseURL,
		client:        &http.Client{Timeout: 20 * time.Second},
		now:           time.Now,
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
	if h.workspaces == nil || h.teams == nil || h.users == nil {
		return web.RespondError(ctx, w, ErrMayaRealtimeToolNotConfigured, http.StatusServiceUnavailable)
	}
	var req AppRealtimeSessionRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	sessionID, maxSessionDuration, remainingDuration, err := h.startRealtimeVoiceSession(ctx, workspace.ID, userID)
	if err != nil {
		if errors.Is(err, ErrMayaRealtimeMonthlyLimitExceeded) {
			return web.RespondError(ctx, w, err, http.StatusTooManyRequests)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	session, err := h.createRealtimeClientSecret(ctx, workspace.ID, userID, req)
	if err != nil {
		h.endRealtimeVoiceSession(ctx, workspace.ID, userID, sessionID)
		h.log.Error(ctx, "failed to create realtime maya session", "error", err, "workspace_id", workspace.ID, "user_id", userID)
		return web.RespondError(ctx, w, err, http.StatusBadGateway)
	}
	session.SessionID = sessionID
	session.MaxSessionSeconds = durationSeconds(maxSessionDuration)
	session.RemainingSeconds = durationSeconds(remainingDuration)
	session.MonthlyLimitSeconds = durationSeconds(realtimeMonthlyVoiceLimit)

	return web.Respond(ctx, w, session, http.StatusCreated)
}

func (h *Handlers) EndRealtimeSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var req AppRealtimeEndSessionRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if req.SessionID == uuid.Nil {
		return web.RespondError(ctx, w, errors.New("sessionId is required"), http.StatusBadRequest)
	}

	if err := h.endRealtimeVoiceSession(ctx, workspace.ID, userID, req.SessionID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, map[string]bool{"success": true}, http.StatusOK)
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
	if h.stories == nil || h.states == nil || h.teams == nil || h.users == nil || h.objectives == nil || h.keyResults == nil || h.search == nil ||
		h.activities == nil || h.feedback == nil || h.notifications == nil || h.reports == nil || h.sprints == nil {
		return web.RespondError(ctx, w, ErrMayaRealtimeToolNotConfigured, http.StatusServiceUnavailable)
	}
	var req AppRealtimeToolRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.validateRealtimeVoiceSession(ctx, workspace.ID, userID, req.SessionID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusConflict)
	}
	cached, claimed, err := h.claimRealtimeToolCall(ctx, req)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusConflict)
	}
	if !claimed {
		return web.Respond(ctx, w, cached, http.StatusOK)
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
		result, err = h.executeCreateTask(ctx, workspace.ID, userID, req.SessionID, req.Arguments)
	case "navigate":
		result, err = h.executeNavigate(ctx, workspace.ID, userID, req.Arguments)
	case "set_theme":
		result = executeSetTheme(req.Arguments)
	case "get_story":
		result, err = h.executeGetStory(ctx, workspace.ID, userID, req.Arguments)
	case "update_story":
		result, err = h.executeUpdateStory(ctx, workspace.ID, userID, req.SessionID, req.Arguments)
	case "story_comments":
		result, err = h.executeStoryComments(ctx, workspace.ID, userID, req.SessionID, req.Arguments)
	case "sprints":
		result, err = h.executeSprints(ctx, workspace.ID, userID, req.Arguments)
	case "workload":
		result, err = h.executeWorkload(ctx, workspace.ID, userID, req.Arguments)
	case "recent_activity":
		result, err = h.executeRecentActivity(ctx, workspace.ID, userID, req.Arguments)
	case "notifications":
		result, err = h.executeNotifications(ctx, workspace.ID, userID, req.SessionID, req.Arguments)
	case "customer_feedback":
		result, err = h.executeCustomerFeedback(ctx, workspace.ID, userID, req.Arguments)
	case "workspace_briefing":
		result, err = h.executeWorkspaceBriefing(ctx, workspace.ID, userID, req.Arguments)
	case "end_conversation":
		result = AppRealtimeToolResponse{
			Success: true,
			Message: "End the realtime voice conversation now.",
		}
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
	if err := h.completeRealtimeToolCall(ctx, req, result); err != nil {
		h.log.Error(ctx, "failed to persist maya realtime tool result", "tool", req.Name, "call_id", req.CallID, "error", err)
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
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

func (h *Handlers) startRealtimeVoiceSession(ctx context.Context, workspaceID, userID uuid.UUID) (uuid.UUID, time.Duration, time.Duration, error) {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, 0, 0, fmt.Errorf("begin realtime voice session transaction: %w", err)
	}
	defer tx.Rollback()

	var lockedWorkspaceID uuid.UUID
	if err := tx.GetContext(ctx, &lockedWorkspaceID, `
		SELECT workspace_id
		FROM workspaces
		WHERE workspace_id = $1
			AND deleted_at IS NULL
		FOR UPDATE
	`, workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, 0, 0, ErrMayaRealtimeMonthlyLimitExceeded
		}
		return uuid.Nil, 0, 0, fmt.Errorf("lock workspace for realtime voice session: %w", err)
	}

	var usedSeconds float64
	if err := tx.GetContext(ctx, &usedSeconds, `
		SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (
			CASE
				WHEN ended_at IS NULL THEN started_at + ($2 * INTERVAL '1 second')
				ELSE LEAST(ended_at, started_at + ($2 * INTERVAL '1 second'))
			END - started_at
		))), 0)
		FROM maya_realtime_voice_sessions
		WHERE workspace_id = $1
			AND started_at >= date_trunc('month', NOW())
			AND started_at < date_trunc('month', NOW()) + INTERVAL '1 month'
	`, workspaceID, durationSeconds(realtimeMaxSessionDuration)); err != nil {
		return uuid.Nil, 0, 0, fmt.Errorf("read realtime voice monthly usage: %w", err)
	}

	usedDuration := time.Duration(usedSeconds * float64(time.Second))
	remainingDuration := realtimeMonthlyVoiceLimit - usedDuration
	if remainingDuration < time.Second {
		return uuid.Nil, 0, 0, ErrMayaRealtimeMonthlyLimitExceeded
	}

	maxSessionDuration := remainingDuration
	if maxSessionDuration > realtimeMaxSessionDuration {
		maxSessionDuration = realtimeMaxSessionDuration
	}

	var sessionID uuid.UUID
	if err := tx.GetContext(ctx, &sessionID, `
		INSERT INTO maya_realtime_voice_sessions (workspace_id, user_id)
		VALUES ($1, $2)
		RETURNING session_id
	`, workspaceID, userID); err != nil {
		return uuid.Nil, 0, 0, fmt.Errorf("create realtime voice session record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, 0, 0, fmt.Errorf("commit realtime voice session transaction: %w", err)
	}

	return sessionID, maxSessionDuration, remainingDuration, nil
}

func (h *Handlers) endRealtimeVoiceSession(ctx context.Context, workspaceID, userID, sessionID uuid.UUID) error {
	_, err := h.db.ExecContext(ctx, `
		UPDATE maya_realtime_voice_sessions
		SET ended_at = COALESCE(ended_at, LEAST(NOW(), started_at + ($4 * INTERVAL '1 second'))),
			updated_at = NOW()
		WHERE workspace_id = $1
			AND user_id = $2
			AND session_id = $3
	`, workspaceID, userID, sessionID, durationSeconds(realtimeMaxSessionDuration))
	if err != nil {
		return fmt.Errorf("end realtime voice session: %w", err)
	}
	return nil
}

func durationSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 0
	}
	return int((duration + time.Second - 1) / time.Second)
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
	currentUser, err := h.currentRealtimeUser(ctx, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	terminology := h.realtimeTerminology(ctx, workspaceID)

	return AppRealtimeToolResponse{
		Success:     true,
		Teams:       toRealtimeVoiceTeams(workspaceTeams),
		User:        &currentUser,
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

func (h *Handlers) executeCreateTask(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeCreateTaskArguments
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid create_task arguments: %w", err)
		}
	}

	args.Title = strings.TrimSpace(args.Title)
	args.Description = strings.TrimSpace(args.Description)
	args.TeamName = strings.TrimSpace(args.TeamName)
	args.AssigneeName = strings.TrimSpace(args.AssigneeName)
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.BlockedByRef = strings.TrimSpace(args.BlockedByRef)
	args.BlockingRef = strings.TrimSpace(args.BlockingRef)
	args.RelatedRef = strings.TrimSpace(args.RelatedRef)
	args.Priority = normalizePriority(args.Priority)
	terminology := h.realtimeTerminology(ctx, workspaceID)
	if args.Title == "" {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user for the %s title before creating it.", terminology.Story),
			Terminology: &terminology,
		}, nil
	}
	isConfirmed := args.Confirmed
	providedConfirmationToken := strings.TrimSpace(args.ConfirmationToken)

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

	currentUser, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get current user: %w", err)
	}
	loc := userLocation(currentUser)
	now := h.now().In(loc)
	assigneeID, assigneeName, assigneeResponse, err := h.resolveRealtimeAssignee(ctx, workspaceID, userID, team, args)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if assigneeResponse != nil {
		assigneeResponse.Terminology = &terminology
		return *assigneeResponse, nil
	}
	startDate, err := parseRealtimeDate(args.StartDate, loc, now)
	if err != nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user to clarify the start date. %s", err.Error()),
			Terminology: &terminology,
		}, nil
	}
	endDate, err := parseRealtimeDate(args.EndDate, loc, now)
	if err != nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user to clarify the due date. %s", err.Error()),
			Terminology: &terminology,
		}, nil
	}

	blockedByID, blockedByRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.BlockedByRef, "blocker")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}
	blockingID, blockingRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.BlockingRef, "blocked work item")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}
	relatedID, relatedRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.RelatedRef, "related work item")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}

	confirmationInput := struct {
		Title         string
		Description   string
		TeamID        uuid.UUID
		StatusID      uuid.UUID
		AssigneeID    *uuid.UUID
		Priority      string
		EstimateValue *int16
		StartDate     *time.Time
		EndDate       *time.Time
		BlockedByID   *uuid.UUID
		BlockingID    *uuid.UUID
		RelatedID     *uuid.UUID
	}{
		Title: args.Title, Description: args.Description, TeamID: team.ID,
		StatusID: status.ID, AssigneeID: assigneeID, Priority: args.Priority,
		EstimateValue: args.EstimateValue, StartDate: startDate, EndDate: endDate,
		BlockedByID: blockedByID, BlockingID: blockingID, RelatedID: relatedID,
	}
	expectedConfirmationToken, err := h.confirmationToken(sessionID, "create_task", confirmationInput)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !isConfirmed {
		return AppRealtimeToolResponse{
			Success:              false,
			RequiresConfirmation: true,
			Message:              fmt.Sprintf("Ask the user to confirm before creating the %s %q in %s.", terminology.Story, args.Title, team.Name),
			Terminology:          &terminology,
			ConfirmationToken:    expectedConfirmationToken,
			Confirmation: &AppRealtimeConfirmation{
				Title:         args.Title,
				Description:   args.Description,
				TeamName:      team.Name,
				AssigneeName:  assigneeName,
				Priority:      args.Priority,
				EstimateValue: args.EstimateValue,
				StartDate:     args.StartDate,
				EndDate:       args.EndDate,
				BlockedByRef:  blockedByRef,
				BlockingRef:   blockingRef,
				RelatedRef:    relatedRef,
			},
		}, nil
	}
	confirmed, err := h.validateConfirmationToken(sessionID, "create_task", confirmationInput, providedConfirmationToken)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !confirmed {
		response := changedConfirmationResponse(expectedConfirmationToken)
		response.Terminology = &terminology
		return response, nil
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
		Assignee:        assigneeID,
		BlockedBy:       blockedByID,
		Blocking:        blockingID,
		Related:         relatedID,
		Reporter:        &userID,
		Priority:        args.Priority,
		EstimateValue:   args.EstimateValue,
		StartDate:       startDate,
		EndDate:         endDate,
		Team:            team.ID,
	}, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("create story: %w", err)
	}

	h.invalidateStoryListCaches(ctx, workspaceID)

	voiceStory := toRealtimeVoiceCreatedStory(story, *team, *status, assigneeName)
	voiceStory.BlockedBy = blockedByRef
	voiceStory.Blocking = blockingRef
	voiceStory.Related = relatedRef
	return AppRealtimeToolResponse{
		Success:     true,
		Story:       &voiceStory,
		Message:     fmt.Sprintf("Created the %s %q in %s.", terminology.Story, story.Title, team.Name),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) createRealtimeClientSecret(ctx context.Context, workspaceID, userID uuid.UUID, sessionRequest AppRealtimeSessionRequest) (AppRealtimeSession, error) {
	terminology := h.realtimeTerminology(ctx, workspaceID)
	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("list teams for realtime context: %w", err)
	}
	currentUser, err := h.currentRealtimeUser(ctx, userID)
	if err != nil {
		return AppRealtimeSession{}, err
	}

	payload := openAIRealtimeClientSecretRequest{
		Session: newRealtimeSessionConfig(terminology, workspaceTeams, currentUser, sessionRequest),
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

func newRealtimeSessionConfig(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam, currentUser AppRealtimeVoiceUser, sessionRequest AppRealtimeSessionRequest) openAIRealtimeSessionConfig {
	return openAIRealtimeSessionConfig{
		Type:             "realtime",
		Model:            defaultRealtimeModel,
		Instructions:     realtimeInstructions(terminology, workspaceTeams, currentUser, sessionRequest),
		MaxOutputTokens:  500,
		OutputModalities: []string{"audio"},
		Tools:            realtimeTools(),
		ToolChoice:       "auto",
		Audio: openAIRealtimeAudioConfig{
			Input: openAIRealtimeAudioInputConfig{
				NoiseReduction: openAIRealtimeNoiseReductionConfig{
					Type: "near_field",
				},
				Transcription: openAIRealtimeTranscriptionConfig{
					Language: "en",
					Model:    defaultRealtimeTranscriptionModel,
					Prompt:   realtimeTranscriptionPrompt(terminology, workspaceTeams),
				},
				TurnDetection: openAIRealtimeTurnDetectionConfig{
					Type:              "server_vad",
					Threshold:         0.75,
					PrefixPaddingMs:   300,
					SilenceDurationMs: 700,
					CreateResponse:    true,
					InterruptResponse: true,
				},
			},
			Output: openAIRealtimeAudioOutputConfig{
				Voice: defaultRealtimeVoice,
			},
		},
	}
}

func realtimeInstructions(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam, currentUser AppRealtimeVoiceUser, sessionRequest AppRealtimeSessionRequest) string {
	instructions := []string{
		"You are Maya, the project management assistant inside FortyOne.",
		"Your job is to help users manage work in FortyOne: work items, teams, priorities, assignments, workload, objectives, key results, activity, and workspace insights.",
		"In voice mode, be concise, natural, and direct. Prefer one to three spoken sentences unless the user asks for detail.",
		"Sound warm, sharp, curious, and genuinely enjoyable to talk to. Let personality come through as natural, context-dependent banter rather than a scripted joke.",
		"Use more playful energy for casual conversation and a lighter touch for professional or operational requests. Keep confirmations, failures, permissions, and sensitive topics straightforward and respectful.",
		"Avoid puns, dad jokes, forced analogies, corporate wordplay, fixed joke templates, and unrelated quips.",
		"Stay focused on project management inside FortyOne. Briefly redirect off-topic requests back to project-management help.",
		"Use available tools whenever facts, permissions, current state, IDs, or state changes are involved.",
		fmt.Sprintf("The current authenticated user is %s (@%s). When the user says me, my, or assign to me, resolve that to this user.", currentUser.Name, currentUser.Username),
		fmt.Sprintf("The user's timezone is %s. Today is %s and the current local time is %s. Interpret relative dates like today, tomorrow, this Friday, and next week in this timezone.", currentUser.Timezone, currentUser.Today, currentUser.Now),
		fmt.Sprintf("Use this workspace's preferred terminology when speaking: stories are called %q/%q, sprints are called %q/%q, objectives are called %q/%q, and key results are called %q/%q.", terminology.Story, terminology.Stories, terminology.Sprint, terminology.Sprints, terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		"Understand all common aliases even when you do not speak them back: story, task, issue, work item, objective, goal, project, key result, milestone, focus area, KPI, sprint, cycle, and iteration.",
		"Use get_context when you need current terminology or team context.",
		"Use list_my_tasks when the user asks about their assigned work, current work, plate, priorities, deadlines, overdue work, what they have today, or what to focus on.",
		"Use list_teams or list_team_members for team questions.",
		"Use search_work when the user asks to find or look up work by name, description, topic, or keyword.",
		fmt.Sprintf("Use list_objectives for %s/%s questions and list_key_results for %s/%s questions.", terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		fmt.Sprintf("Use create_task when the user asks you to create a %s, task, story, issue, or work item.", terminology.Story),
		"Use navigate to open FortyOne pages or records, and set_theme to change the application's appearance.",
		"Use get_story and update_story for story details and confirmed field changes. Use story_comments to read comments or add one after confirmation.",
		"Use sprints for running sprint lists and sprint summaries, workload for workload or capacity questions, recent_activity for recent workspace changes, notifications for notification questions and confirmed read actions, customer_feedback for customer feedback, and workspace_briefing for a concise operational overview.",
		"When the user clearly ends the conversation with phrases like bye, goodbye, that's all, thanks that's all, or talk later, say a brief goodbye and call end_conversation.",
		"Do not guess teams, statuses, permissions, or results. Ask a short clarifying question when the target is ambiguous.",
		teamSelectionInstruction(workspaceTeams),
		"Never expose raw UUIDs. Use human-readable names and story references.",
		"Keep tool usage internal. Do not mention tool names, parameters, or implementation details to the user.",
		"Never claim an action succeeded unless the tool result clearly shows success.",
		fmt.Sprintf("For %s creation: gather the title and target team if needed, resolve assignees from team members, convert natural dates to startDate/endDate, draft a concise title and useful description, ask for explicit confirmation, then call create_task with confirmed=true only after the user confirms the exact %s.", terminology.Story, terminology.Story),
		"For assignment during creation: set assignToMe=true when the user says me, myself, or assign to me. Set assigneeName when the user names another person; the backend resolves that name against team members.",
		"For estimates during creation: set estimateValue only when the user gives a numeric estimate such as 1, 2, 3, 5, or 8. If the estimate is non-numeric or unclear, ask a short clarifying question.",
		"For blockers and related work during creation: set blockedByRef when the new item is blocked by existing work, blockingRef when the new item blocks existing work, and relatedRef for related existing work. Use a human-readable story reference or title; the backend resolves it.",
		"If a tool returns requiresConfirmation, ask the requested confirmation in plain language. If the user confirms, repeat the exact same action details with confirmed=true and the returned confirmationToken. Never invent or reuse a token for different details.",
		"If a tool returns needsTeam, ask the requested clarification in plain language.",
		"If a tool returns needsAssignee, ask which team member should be assigned.",
		"If a tool returns needsStoryReference, ask which existing work item the user meant, using the returned references and titles.",
		"If a tool fails, repeat the useful error briefly. Do not invent a fallback workflow.",
	}

	if currentPath := strings.TrimSpace(sessionRequest.CurrentPath); currentPath != "" {
		instructions = append(instructions, fmt.Sprintf("The user started voice mode from the FortyOne path %q. Use it only to resolve references such as this page or this story when the conversation supports that interpretation.", currentPath))
	}
	if recentConversation := realtimeConversationContext(sessionRequest.Messages); recentConversation != "" {
		instructions = append(instructions, "Continue naturally from this recent typed and voice conversation. Do not repeat information the user already received:\n"+recentConversation)
	}

	return strings.Join(instructions, " ")
}

func realtimeConversationContext(messages []AppRealtimeConversationMessage) string {
	if len(messages) == 0 {
		return ""
	}

	lines := make([]string, 0, len(messages))
	for _, message := range messages {
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		speaker := "User"
		if message.Role == "assistant" {
			speaker = "Maya"
		}
		lines = append(lines, speaker+": "+text)
	}
	return strings.Join(lines, "\n")
}

func realtimeTranscriptionPrompt(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam) string {
	terms := []string{
		"FortyOne",
		"Maya",
		terminology.Story,
		terminology.Stories,
		terminology.Sprint,
		terminology.Sprints,
		terminology.Objective,
		terminology.Objectives,
		terminology.KeyResult,
		terminology.KeyResults,
	}
	for _, team := range workspaceTeams {
		if name := strings.TrimSpace(team.Name); name != "" {
			terms = append(terms, name)
		}
	}
	return "Expect FortyOne workspace terminology, names, and references including: " + strings.Join(terms, ", ") + "."
}

func realtimeTools() []openAIRealtimeTool {
	tools := []openAIRealtimeTool{
		{
			Type:        "function",
			Name:        "end_conversation",
			Description: "End the realtime voice conversation when the user says goodbye or clearly indicates they are done.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
		},
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
					"assigneeName": map[string]any{
						"type":        "string",
						"description": "Optional assignee name, username, or 'me'. The backend resolves this against the selected team's members.",
					},
					"assignToMe": map[string]any{
						"type":        "boolean",
						"description": "Set true when the user asks to assign the item to themselves.",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"No Priority", "Low", "Medium", "High", "Urgent"},
						"description": "Optional story priority. Defaults to No Priority.",
					},
					"estimateValue": map[string]any{
						"type":        "integer",
						"description": "Optional numeric estimate value. Use only when the user gives a numeric estimate.",
					},
					"startDate": map[string]any{
						"type":        "string",
						"description": "Optional start date. Prefer YYYY-MM-DD, but natural phrases like today, tomorrow, this Friday, or next Friday are accepted.",
					},
					"endDate": map[string]any{
						"type":        "string",
						"description": "Optional due/deadline date. Prefer YYYY-MM-DD, but natural phrases like today, tomorrow, this Friday, or next Friday are accepted.",
					},
					"blockedByRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title that blocks this new story.",
					},
					"blockingRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title that this new story blocks.",
					},
					"relatedRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title related to this new story.",
					},
					"confirmed": map[string]any{
						"type":        "boolean",
						"description": "True only after the user explicitly confirms creating this story.",
					},
					"confirmationToken": map[string]any{
						"type":        "string",
						"description": "The exact token returned by the preceding confirmation request. Include only after the user confirms without changing details.",
					},
				},
				"required": []string{"title", "confirmed"},
			},
		},
	}
	return append(tools, realtimeExtendedTools()...)
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
		Reference:     storyReference(teamCode, story.SequenceID),
		Title:         story.Title,
		Priority:      story.Priority,
		EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue,
		Team:          teamName,
		Status:        status,
		StartDate:     story.StartDate,
		EndDate:       story.EndDate,
		CompletedAt:   story.CompletedAt,
	}
}

func toRealtimeVoiceCreatedStory(story stories.CoreSingleStory, team teams.CoreTeam, status states.CoreState, assigneeName string) AppRealtimeVoiceStory {
	return AppRealtimeVoiceStory{
		Reference:     storyReference(team.Code, story.SequenceID),
		Title:         story.Title,
		Priority:      story.Priority,
		EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue,
		Team:          team.Name,
		Assignee:      assigneeName,
		Status:        toRealtimeVoiceStatus(status),
		StartDate:     story.StartDate,
		EndDate:       story.EndDate,
		CompletedAt:   story.CompletedAt,
	}
}

func (h *Handlers) currentRealtimeUser(ctx context.Context, userID uuid.UUID) (AppRealtimeVoiceUser, error) {
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return AppRealtimeVoiceUser{}, fmt.Errorf("get current user: %w", err)
	}
	loc := userLocation(user)
	now := h.now().In(loc)
	name := strings.TrimSpace(user.FullName)
	if name == "" {
		name = user.Username
	}
	return AppRealtimeVoiceUser{
		Name:     name,
		Username: user.Username,
		Timezone: loc.String(),
		Today:    now.Format("2006-01-02"),
		Now:      now.Format("15:04"),
	}, nil
}

func userLocation(user users.CoreUser) *time.Location {
	timezone := strings.TrimSpace(user.Timezone)
	if timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (h *Handlers) resolveRealtimeAssignee(ctx context.Context, workspaceID, userID uuid.UUID, team *teams.CoreTeam, args AppRealtimeCreateTaskArguments) (*uuid.UUID, string, *AppRealtimeToolResponse, error) {
	assigneeName := strings.TrimSpace(args.AssigneeName)
	if args.AssignToMe || isSelfReference(assigneeName) {
		currentUser, err := h.users.GetUser(ctx, userID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("get current user: %w", err)
		}
		return &userID, displayUserName(currentUser), nil, nil
	}
	if assigneeName == "" {
		return nil, "", nil, nil
	}

	members, err := h.users.List(ctx, workspaceID, users.CoreListUsersFilter{
		TeamID: &team.ID,
		Search: assigneeName,
		Limit:  25,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("list team members: %w", err)
	}
	matches := resolveRealtimeMemberMatches(members, assigneeName)
	if len(matches) == 1 {
		member := matches[0]
		memberID := member.ID
		return &memberID, displayUserName(member), nil, nil
	}

	message := fmt.Sprintf("Ask which %s member should be assigned.", team.Name)
	if len(matches) == 0 {
		message = fmt.Sprintf("I could not find %q in %s. Ask for the assignee's name again.", assigneeName, team.Name)
	}
	return nil, "", &AppRealtimeToolResponse{
		Success:       false,
		NeedsAssignee: true,
		Members:       toRealtimeVoiceMembers(members),
		Message:       message,
	}, nil
}

func resolveRealtimeMemberMatches(memberList []users.CoreUser, assigneeName string) []users.CoreUser {
	normalized := normalizeName(assigneeName)
	var exact []users.CoreUser
	for _, member := range memberList {
		if normalizeName(member.FullName) == normalized || normalizeName(member.Username) == normalized || normalizeName(member.Email) == normalized {
			exact = append(exact, member)
		}
	}
	if len(exact) > 0 {
		return exact
	}

	var partial []users.CoreUser
	for _, member := range memberList {
		if strings.Contains(normalizeName(member.FullName), normalized) ||
			strings.Contains(normalizeName(member.Username), normalized) ||
			strings.Contains(normalizeName(member.Email), normalized) {
			partial = append(partial, member)
		}
	}
	return partial
}

func (h *Handlers) resolveRealtimeStoryLink(ctx context.Context, workspaceID, userID uuid.UUID, teamList []teams.CoreTeam, team *teams.CoreTeam, value, label string) (*uuid.UUID, string, *AppRealtimeToolResponse, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, "", nil, nil
	}

	teamsByID := indexTeamsByID(teamList)
	if storyRef, ok := normalizeRealtimeStoryReference(value, team); ok {
		story, err := h.stories.QueryByRef(ctx, workspaceID, storyRef)
		if err == nil {
			storyID := story.ID
			return &storyID, storyReference(story.TeamCode, story.SequenceID), nil, nil
		}
	}

	teamID := (*uuid.UUID)(nil)
	if team != nil {
		teamID = &team.ID
	}
	searchResult, err := h.search.Search(ctx, workspaceID, userID, search.SearchParams{
		Type:     search.SearchTypeStories,
		Query:    value,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: 5,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("search %s story reference: %w", label, err)
	}

	matches := resolveRealtimeStoryMatches(searchResult.Stories, value, teamsByID)
	if len(matches) == 1 {
		story := matches[0]
		storyID := story.ID
		return &storyID, realtimeSearchStoryReference(story, teamsByID), nil, nil
	}

	statusesByID, err := h.statusesByID(ctx, workspaceID, userID)
	if err != nil {
		return nil, "", nil, err
	}
	voiceStories := make([]AppRealtimeVoiceStory, 0, len(searchResult.Stories))
	for _, story := range searchResult.Stories {
		voiceStories = append(voiceStories, toRealtimeVoiceSearchStory(story, teamsByID, statusesByID))
	}

	message := fmt.Sprintf("I could not find %q. Ask for the existing %s reference or title.", value, label)
	if len(voiceStories) > 1 {
		message = fmt.Sprintf("Ask which existing %s the user meant.", label)
	}
	return nil, "", &AppRealtimeToolResponse{
		Success:             false,
		NeedsStoryReference: true,
		Stories:             voiceStories,
		Count:               len(voiceStories),
		Message:             message,
	}, nil
}

func resolveRealtimeStoryMatches(storyList []search.CoreSearchStory, value string, teamsByID map[uuid.UUID]teams.CoreTeam) []search.CoreSearchStory {
	if len(storyList) == 0 {
		return nil
	}

	normalizedValue := normalizeName(value)
	var exact []search.CoreSearchStory
	for _, story := range storyList {
		if normalizeName(story.Title) == normalizedValue || normalizeName(realtimeSearchStoryReference(story, teamsByID)) == normalizedValue {
			exact = append(exact, story)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	if len(storyList) == 1 {
		return storyList
	}
	return storyList
}

func normalizeRealtimeStoryReference(value string, team *teams.CoreTeam) (string, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}

	var letters strings.Builder
	var digits strings.Builder
	seenDigit := false
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			if seenDigit {
				return "", false
			}
			letters.WriteRune(r)
		case r >= '0' && r <= '9':
			seenDigit = true
			digits.WriteRune(r)
		case r == '-' || unicode.IsSpace(r):
			continue
		default:
			return "", false
		}
	}

	if digits.Len() == 0 {
		return "", false
	}
	sequenceID, err := strconv.Atoi(digits.String())
	if err != nil || sequenceID <= 0 {
		return "", false
	}

	teamCode := letters.String()
	if teamCode == "" {
		if team == nil || strings.TrimSpace(team.Code) == "" {
			return "", false
		}
		teamCode = strings.ToUpper(strings.TrimSpace(team.Code))
	}
	return storyReference(teamCode, sequenceID), true
}

func realtimeSearchStoryReference(story search.CoreSearchStory, teamsByID map[uuid.UUID]teams.CoreTeam) string {
	team, ok := teamsByID[story.Team]
	if !ok {
		return ""
	}
	return storyReference(team.Code, story.SequenceID)
}

func isSelfReference(value string) bool {
	switch normalizeName(value) {
	case "me", "myself", "self", "i":
		return true
	default:
		return false
	}
}

func displayUserName(user users.CoreUser) string {
	name := strings.TrimSpace(user.FullName)
	if name != "" {
		return name
	}
	return user.Username
}

func parseRealtimeDate(value string, loc *time.Location, now time.Time) (*time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", raw, loc); err == nil {
		return &parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		local := parsed.In(loc)
		date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
		return &date, nil
	}

	normalized := normalizeDatePhrase(raw)
	switch normalized {
	case "today":
		date := dateOnly(now, loc)
		return &date, nil
	case "tomorrow":
		date := dateOnly(now.AddDate(0, 0, 1), loc)
		return &date, nil
	case "next week":
		date := dateOnly(now.AddDate(0, 0, 7), loc)
		return &date, nil
	}

	if weekday, ok := weekdayFromPhrase(normalized); ok {
		date := dateOnly(nextWeekday(now, loc, weekday, strings.HasPrefix(normalized, "next ")), loc)
		return &date, nil
	}

	return nil, fmt.Errorf("use YYYY-MM-DD or a relative date like today, tomorrow, this Friday, or next Friday")
}

func normalizeDatePhrase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, prefix := range []string{"due ", "on ", "by ", "deadline ", "start ", "starting "} {
		value = strings.TrimPrefix(value, prefix)
	}
	return strings.Join(strings.Fields(value), " ")
}

func weekdayFromPhrase(value string) (time.Weekday, bool) {
	value = strings.TrimPrefix(value, "this ")
	value = strings.TrimPrefix(value, "next ")
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	weekday, ok := weekdays[value]
	return weekday, ok
}

func nextWeekday(now time.Time, loc *time.Location, weekday time.Weekday, forceNext bool) time.Time {
	local := now.In(loc)
	days := (int(weekday) - int(local.Weekday()) + 7) % 7
	if forceNext {
		if days == 0 {
			days = 7
		} else {
			days += 7
		}
	}
	return local.AddDate(0, 0, days)
}

func dateOnly(value time.Time, loc *time.Location) time.Time {
	local := value.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
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
