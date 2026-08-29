package epicshttp

import (
	"context"
	"errors"
	"net/http"

	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	epics *epics.Service
}

func New(epics *epics.Service) *Handlers {
	return &Handlers{
		epics: epics,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	return h.listForWorkspace(ctx, w, workspace.ID)
}

func (h *Handlers) listForWorkspace(ctx context.Context, w http.ResponseWriter, workspaceID uuid.UUID) error {
	if err := h.epics.List(ctx, workspaceID); err != nil {
		if errors.Is(err, epics.ErrNotImplemented) {
			return web.RespondError(ctx, w, err, http.StatusNotImplemented)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}
