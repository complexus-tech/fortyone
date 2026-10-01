package workautomationshttp

import (
	"context"
	"errors"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	workautomations "github.com/complexus-tech/projects-api/internal/modules/workautomations/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct{ service *workautomations.Service }

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actor, workspace, err := identities(ctx)
	if err != nil {
		return respondError(ctx, w, err)
	}
	query := r.URL.Query()
	if len(query) != 1 || len(query["teamId"]) != 1 {
		return respondError(ctx, w, domain.ErrInvalidInput)
	}
	team, err := uuid.Parse(query.Get("teamId"))
	if err != nil {
		return respondError(ctx, w, domain.ErrInvalidInput)
	}
	items, err := h.service.List(ctx, actor, workspace, team)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, items, http.StatusOK)
}
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actor, workspace, err := identities(ctx)
	if err != nil {
		return respondError(ctx, w, err)
	}
	var input domain.Input
	if err := web.DecodeWithLimit(r, &input, 70<<10); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	item, err := h.service.Create(ctx, actor, workspace, input)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, item, http.StatusCreated)
}
func (h *Handlers) Pause(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actor, workspace, err := identities(ctx)
	if err != nil {
		return respondError(ctx, w, err)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return respondError(ctx, w, domain.ErrInvalidInput)
	}
	var input domain.Update
	if err := web.DecodeWithLimit(r, &input, 1024); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	item, err := h.service.Pause(ctx, actor, workspace, id, *input.Paused)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, item, http.StatusOK)
}
func (h *Handlers) Archive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actor, workspace, err := identities(ctx)
	if err != nil {
		return respondError(ctx, w, err)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return respondError(ctx, w, domain.ErrInvalidInput)
	}
	if err := h.service.Archive(ctx, actor, workspace, id); err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
func (h *Handlers) Runs(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actor, workspace, err := identities(ctx)
	if err != nil {
		return respondError(ctx, w, err)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return respondError(ctx, w, domain.ErrInvalidInput)
	}
	items, err := h.service.Runs(ctx, actor, workspace, id)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, items, http.StatusOK)
}
func identities(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	actor, err := mid.GetUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	workspace, err := mid.GetWorkspace(ctx)
	return actor, workspace.ID, err
}
func respondError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	if errors.Is(err, domain.ErrInvalidInput) {
		status = http.StatusBadRequest
	}
	if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
	}
	return web.RespondError(ctx, w, err, status)
}
