package usershttp

import (
	"context"
	"errors"
	"net/http"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) DeleteAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	// Browser cookies can change in another tab after a destructive dialog opens.
	// This is an optional precondition, never an account-selection parameter.
	if r.Body != nil && r.Body != http.NoBody && r.ContentLength != 0 {
		var request struct {
			ExpectedUserID *uuid.UUID `json:"expectedUserId"`
		}
		if err := web.Decode(r, &request); err != nil {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
		if request.ExpectedUserID != nil && *request.ExpectedUserID != userID {
			return web.RespondError(ctx, w, errors.New("Your signed-in account changed. Refresh settings before trying again."), http.StatusConflict)
		}
	}
	pending, err := h.users.DeleteAccount(ctx, userID)
	if err != nil {
		var conflict *usersdomain.AccountDeletionConflict
		switch {
		case errors.As(err, &conflict):
			return web.RespondError(ctx, w, conflict, http.StatusConflict)
		case errors.Is(err, usersdomain.ErrNotFound):
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		case errors.Is(err, usersdomain.ErrAccountDeletionUnavailable):
			return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
		default:
			if h.log != nil {
				h.log.Error(ctx, "permanent account deletion failed", "error", err)
			}
			return web.RespondError(ctx, w, errors.New("Your account could not be deleted. Please try again."), http.StatusInternalServerError)
		}
	}
	// Account state is authoritative for every browser/mobile session, including
	// sessions on other devices. A cache outage cannot revive the erased account.
	h.clearSessionCookie(w, r)
	if pending {
		return web.Respond(ctx, w, map[string]string{"status": "cleanup_pending"}, http.StatusAccepted)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
