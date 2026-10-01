package scimhttp

import (
	"context"
	"errors"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func adminScope(ctx context.Context) (domain.AdminScope, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return domain.AdminScope{}, domain.ErrForbidden
	}
	actor, err := mid.GetUserID(ctx)
	if err != nil {
		return domain.AdminScope{}, domain.ErrForbidden
	}
	return domain.AdminScope{WorkspaceID: workspace.ID, ActorID: actor}, nil
}
func adminRespond(ctx context.Context, w http.ResponseWriter, result any, err error) error {
	w.Header().Set("Cache-Control", "private, no-store")
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrForbidden):
			status = http.StatusForbidden
		case errors.Is(err, domain.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, domain.ErrInvalidInput):
			status = http.StatusBadRequest
		case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrChanged):
			status = http.StatusConflict
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}
func (h *Handlers) AdminStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := adminScope(ctx)
	if err != nil {
		return adminRespond(ctx, w, nil, err)
	}
	result, err := h.service.Status(ctx, scope)
	return adminRespond(ctx, w, result, err)
}
func (h *Handlers) Mint(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := adminScope(ctx)
	if err != nil {
		return adminRespond(ctx, w, nil, err)
	}
	var input domain.MintInput
	if err = web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Mint(ctx, scope, input)
	return adminRespond(ctx, w, result, err)
}
func (h *Handlers) Revoke(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := adminScope(ctx)
	if err != nil {
		return adminRespond(ctx, w, nil, err)
	}
	id, err := uuid.Parse(web.Params(r, "credentialId"))
	if err != nil || id == uuid.Nil {
		return adminRespond(ctx, w, nil, domain.ErrInvalidInput)
	}
	return adminRespond(ctx, w, nil, h.service.Revoke(ctx, scope, id))
}
func (h *Handlers) RetrySeats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := adminScope(ctx)
	if err != nil {
		return adminRespond(ctx, w, nil, err)
	}
	return adminRespond(ctx, w, nil, h.service.RetrySeats(ctx, scope))
}
