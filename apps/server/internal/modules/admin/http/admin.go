package adminhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Handlers struct {
	admin *admin.Service
}

type updateWorkspaceTrialRequest struct {
	TrialEndsOn time.Time `json:"trialEndsOn"`
	Reason      string    `json:"reason"`
}

func New(adminService *admin.Service) *Handlers {
	return &Handlers{admin: adminService}
}

func (h *Handlers) GetCurrentAdmin(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.admin.GetCurrentAdmin(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
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
	})
	if err != nil {
		return web.RespondError(ctx, w, err, adminErrorStatus(err))
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func adminErrorStatus(err error) int {
	switch {
	case errors.Is(err, admin.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, admin.ErrNotFound):
		return http.StatusNotFound
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
