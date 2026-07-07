package adminhttp

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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

type updateWorkspaceTrialRequest struct {
	TrialEndsOn time.Time `json:"trialEndsOn"`
	Reason      string    `json:"reason"`
}

type updateWorkspaceDeletedRequest struct {
	Deleted bool   `json:"deleted"`
	Reason  string `json:"reason"`
}

type reasonRequest struct {
	Reason string `json:"reason"`
}

type updateUserStateRequest struct {
	IsActive   *bool  `json:"isActive"`
	IsInternal *bool  `json:"isInternal"`
	Reason     string `json:"reason"`
}

type createAdminNoteRequest struct {
	TargetType  string     `json:"targetType"`
	TargetID    uuid.UUID  `json:"targetId"`
	WorkspaceID *uuid.UUID `json:"workspaceId"`
	Body        string     `json:"body"`
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

	page, limit := paginationParams(r)
	result, err := h.admin.ListWorkspaces(ctx, userID, admin.ListWorkspacesInput{
		Pagination: admin.PaginationInput{Page: page, Limit: limit},
		Query:      r.URL.Query().Get("q"),
		Status:     r.URL.Query().Get("status"),
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

	page, limit := paginationParams(r)
	result, err := h.admin.ListUsers(ctx, userID, admin.ListUsersInput{
		Pagination: admin.PaginationInput{Page: page, Limit: limit},
		Query:      r.URL.Query().Get("q"),
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
		IsActive:   req.IsActive,
		IsInternal: req.IsInternal,
		Reason:     req.Reason,
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

	workspaceID, err := optionalUUIDQuery(r, "workspaceId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	page, limit := paginationParams(r)
	result, err := h.admin.ListAuditLogs(ctx, userID, admin.ListAuditLogsInput{
		Pagination:  admin.PaginationInput{Page: page, Limit: limit},
		WorkspaceID: workspaceID,
		TargetType:  r.URL.Query().Get("targetType"),
		Query:       r.URL.Query().Get("q"),
		Action:      r.URL.Query().Get("action"),
		ActorQuery:  r.URL.Query().Get("actor"),
		From:        optionalTimeQuery(r, "from"),
		To:          optionalTimeQuery(r, "to"),
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "list_audit_logs", err)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) ExportAuditLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := optionalUUIDQuery(r, "workspaceId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	result, err := h.admin.ListAuditLogs(ctx, userID, admin.ListAuditLogsInput{
		Pagination:  admin.PaginationInput{Page: 1, Limit: maxPageSize},
		WorkspaceID: workspaceID,
		TargetType:  r.URL.Query().Get("targetType"),
		Query:       r.URL.Query().Get("q"),
		Action:      r.URL.Query().Get("action"),
		ActorQuery:  r.URL.Query().Get("actor"),
		From:        optionalTimeQuery(r, "from"),
		To:          optionalTimeQuery(r, "to"),
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "export_audit_logs", err)
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="admin-audit-logs.csv"`)
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"actor", "target_type", "target_id", "workspace", "action", "field", "old_value", "new_value", "reason", "created_at"}); err != nil {
		return err
	}
	for _, entry := range result.Items {
		if err := writer.Write([]string{
			entry.ActorName,
			entry.TargetType,
			uuidString(entry.TargetID),
			stringValue(entry.WorkspaceName),
			entry.Action,
			entry.FieldName,
			csvValue(entry.OldValue),
			csvValue(entry.NewValue),
			entry.Reason,
			entry.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
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

	page, limit := paginationParams(r)
	result, err := h.admin.ListAdminNotes(ctx, userID, admin.ListAdminNotesInput{
		Pagination:  admin.PaginationInput{Page: page, Limit: limit},
		TargetType:  r.URL.Query().Get("targetType"),
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

func (h *Handlers) respondAdminError(ctx context.Context, w http.ResponseWriter, operation string, err error) error {
	status := adminErrorStatus(err)
	if status >= http.StatusInternalServerError && h.log != nil {
		h.log.Error(ctx, "admin request failed", "operation", operation, "error", err)
	}
	return web.RespondError(ctx, w, err, status)
}

func adminErrorStatus(err error) int {
	switch {
	case errors.Is(err, admin.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, admin.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, admin.ErrInvalidAdminAction):
		return http.StatusBadRequest
	case errors.Is(err, admin.ErrInvalidAdminNote):
		return http.StatusBadRequest
	case errors.Is(err, admin.ErrReasonRequired):
		return http.StatusBadRequest
	case errors.Is(err, admin.ErrSelfMutation):
		return http.StatusBadRequest
	case errors.Is(err, admin.ErrInvalidTrialEndsOn):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func paginationParams(r *http.Request) (int, int) {
	page := positiveIntQuery(r, "page", 1)
	limit := positiveIntQuery(r, "limit", defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}
	return page, limit
}

func positiveIntQuery(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func optionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalTimeQuery(r *http.Request, key string) *time.Time {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", raw)
	}
	if err != nil {
		return nil
	}
	return &parsed
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func csvValue(value any) string {
	if value == nil {
		return ""
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}
