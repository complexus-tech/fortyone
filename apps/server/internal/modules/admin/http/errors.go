package adminhttp

import (
	"context"
	"errors"
	"net/http"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

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
	case errors.Is(err, admin.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, admin.ErrIntegrationUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, admin.ErrInvalidAdminAction),
		errors.Is(err, admin.ErrInvalidFilter),
		errors.Is(err, admin.ErrInvalidAdminNote),
		errors.Is(err, admin.ErrReasonRequired),
		errors.Is(err, admin.ErrSelfMutation),
		errors.Is(err, admin.ErrInvalidTrialEndsOn):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
