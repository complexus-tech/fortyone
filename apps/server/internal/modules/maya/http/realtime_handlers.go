package mayahttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

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
		if cleanupErr := h.endRealtimeVoiceSession(ctx, workspace.ID, userID, sessionID); cleanupErr != nil {
			h.log.Error(
				ctx,
				"failed to release realtime maya session after client secret failure",
				"error", cleanupErr,
				"workspace_id", workspace.ID,
				"user_id", userID,
				"session_id", sessionID,
			)
		}
		h.log.Error(ctx, "failed to create realtime maya session", "error", err, "workspace_id", workspace.ID, "user_id", userID)
		return web.RespondError(ctx, w, err, http.StatusBadGateway)
	}
	session.SessionID = sessionID
	session.MaxSessionSeconds = durationSeconds(maxSessionDuration)
	session.RemainingSeconds = durationSeconds(remainingDuration)
	session.MonthlyLimitSeconds = durationSeconds(realtimeMonthlyVoiceLimit)

	return web.Respond(ctx, w, session, http.StatusCreated)
}

func (h *Handlers) RetryScheduleIssue(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.RetryScheduleIssue(ctx, workspace.ID, storyID, userID); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) OverrideScheduleIssue(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var req AppManualScheduleStoryRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if _, err := h.service.ManuallyScheduleStory(ctx, maya.ManualScheduleStoryInput{
		WorkspaceID: workspace.ID,
		StoryID:     storyID,
		UserID:      userID,
		StartAt:     req.StartAt,
		Timezone:    req.Timezone,
	}); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
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
	case errors.Is(err, maya.ErrMayaAccessDenied), errors.Is(err, ErrMayaAccessRequired):
		return http.StatusPaymentRequired
	case errors.Is(err, stories.ErrAutoSchedulingUnavailable):
		return http.StatusPaymentRequired
	case errors.Is(err, stories.ErrAutoSchedulingAccessCheckFailed):
		return http.StatusServiceUnavailable
	case errors.Is(err, maya.ErrPlanNotFound):
		return http.StatusNotFound
	case errors.Is(err, stories.ErrAutoSchedulingOwnerLocked),
		errors.Is(err, stories.ErrAutoSchedulingLockEmpty),
		errors.Is(err, stories.ErrStoryChanged),
		errors.Is(err, calendar.ErrCalendarScheduleStalePlan),
		errors.Is(err, calendar.ErrCalendarScheduleConflict):
		return http.StatusConflict
	case errors.Is(err, maya.ErrInvalidPlanInput):
		return http.StatusBadRequest
	case errors.Is(err, stories.ErrMayaAssignmentRequiresScheduling),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDuration),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDeliveryDate):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handlers) workspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	return h.service.WorkspaceCanUseMaya(ctx, workspaceID)
}

func (h *Handlers) startRealtimeVoiceSession(ctx context.Context, workspaceID, userID uuid.UUID) (uuid.UUID, time.Duration, time.Duration, error) {
	lease, err := h.service.BeginRealtimeVoiceSession(ctx, workspaceID, userID)
	if err != nil {
		return uuid.Nil, 0, 0, err
	}
	return lease.SessionID, lease.MaxDuration, lease.RemainingDuration, nil
}

func (h *Handlers) endRealtimeVoiceSession(ctx context.Context, workspaceID, userID, sessionID uuid.UUID) error {
	return h.service.EndRealtimeVoiceSession(ctx, workspaceID, userID, sessionID)
}

func durationSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 0
	}
	return int((duration + time.Second - 1) / time.Second)
}
