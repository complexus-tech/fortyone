package adminhttp

import (
	"context"
	"net/http"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Handlers struct {
	admin *admin.Service
	log   *logger.Logger
}

func New(log *logger.Logger, adminService *admin.Service) *Handlers {
	return &Handlers{admin: adminService, log: log}
}

func (h *Handlers) GetCurrentAdmin(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.admin.GetCurrentAdmin(ctx, userID)
	if err != nil {
		return h.respondAdminError(ctx, w, "get_current_admin", err)
	}
	return web.Respond(ctx, w, user, http.StatusOK)
}

func (h *Handlers) GetDashboardSummary(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	summary, err := h.admin.GetDashboardSummary(ctx, userID)
	if err != nil {
		return h.respondAdminError(ctx, w, "get_dashboard_summary", err)
	}
	return web.Respond(ctx, w, summary, http.StatusOK)
}

func (h *Handlers) ListWorkspaces(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	page, limit, err := paginationParams(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	query, err := optionalTextQuery(r, "q", maximumAdminSearchRunes)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	status, err := optionalTextQuery(r, "status", maximumAdminFilterRunes)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.admin.ListWorkspaces(ctx, userID, admin.ListWorkspacesInput{
		Pagination: admin.PaginationInput{Page: page, Limit: limit},
		Query:      query,
		Status:     status,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "list_workspaces", err)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) GetWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspace, err := h.admin.GetWorkspaceOverview(ctx, userID, workspaceID)
	if err != nil {
		return h.respondAdminError(ctx, w, "get_workspace", err)
	}
	return web.Respond(ctx, w, workspace, http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceTrial(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req updateWorkspaceTrialRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspace, err := h.admin.UpdateWorkspaceTrial(ctx, userID, workspaceID, admin.UpdateWorkspaceTrialInput{
		TrialEndsOn: req.TrialEndsOn,
		Reason:      req.Reason,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "update_workspace_trial", err)
	}
	return web.Respond(ctx, w, workspace, http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceDeleted(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req updateWorkspaceDeletedRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspace, err := h.admin.UpdateWorkspaceDeleted(ctx, userID, workspaceID, admin.UpdateWorkspaceDeletedInput{
		Deleted: req.Deleted,
		Reason:  req.Reason,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "update_workspace_deleted", err)
	}
	return web.Respond(ctx, w, workspace, http.StatusOK)
}

func (h *Handlers) RequestWorkspaceSubscriptionSync(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req reasonRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspace, err := h.admin.RequestWorkspaceSubscriptionSync(ctx, userID, workspaceID, admin.RequestWorkspaceSubscriptionSyncInput{
		Reason: req.Reason,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "request_workspace_subscription_sync", err)
	}
	return web.Respond(ctx, w, workspace, http.StatusOK)
}

func (h *Handlers) ListUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	page, limit, err := paginationParams(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	query, err := optionalTextQuery(r, "q", maximumAdminSearchRunes)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.admin.ListUsers(ctx, userID, admin.ListUsersInput{
		Pagination: admin.PaginationInput{Page: page, Limit: limit},
		Query:      query,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "list_users", err)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := uuid.Parse(web.Params(r, "userID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	user, err := h.admin.GetUserOverview(ctx, actorID, userID)
	if err != nil {
		return h.respondAdminError(ctx, w, "get_user", err)
	}
	return web.Respond(ctx, w, user, http.StatusOK)
}

func (h *Handlers) UpdateUserState(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := uuid.Parse(web.Params(r, "userID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req updateUserStateRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	user, err := h.admin.UpdateUserState(ctx, actorID, userID, admin.UpdateUserStateInput{
		Patch:  userStatePatch(req),
		Reason: req.Reason,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "update_user_state", err)
	}
	return web.Respond(ctx, w, user, http.StatusOK)
}

func (h *Handlers) RequestUserSessionRevocation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := uuid.Parse(web.Params(r, "userID"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req reasonRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	user, err := h.admin.RequestUserSessionRevocation(ctx, actorID, userID, admin.RequestUserSessionRevocationInput{
		Reason: req.Reason,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "request_user_session_revocation", err)
	}
	return web.Respond(ctx, w, user, http.StatusOK)
}

func (h *Handlers) ListAuditLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters, err := parseAuditLogQuery(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	page, limit, err := paginationParams(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.admin.ListAuditLogs(ctx, userID, admin.ListAuditLogsInput{
		Pagination:  admin.PaginationInput{Page: page, Limit: limit},
		WorkspaceID: filters.WorkspaceID,
		TargetType:  filters.TargetType,
		Query:       filters.Query,
		Action:      filters.Action,
		ActorQuery:  filters.ActorQuery,
		From:        filters.From,
		To:          filters.To,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "list_audit_logs", err)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) ListAdminNotes(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	targetID, err := optionalUUIDQuery(r, "targetId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	workspaceID, err := optionalUUIDQuery(r, "workspaceId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	page, limit, err := paginationParams(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	targetType, err := optionalTextQuery(r, "targetType", maximumAdminFilterRunes)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.admin.ListAdminNotes(ctx, userID, admin.ListAdminNotesInput{
		Pagination:  admin.PaginationInput{Page: page, Limit: limit},
		TargetType:  targetType,
		TargetID:    targetID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "list_admin_notes", err)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) CreateAdminNote(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req createAdminNoteRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	note, err := h.admin.CreateAdminNote(ctx, userID, admin.CreateAdminNoteInput{
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		WorkspaceID: req.WorkspaceID,
		Body:        req.Body,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "create_admin_note", err)
	}
	return web.Respond(ctx, w, note, http.StatusCreated)
}
