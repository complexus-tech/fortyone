package epicsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/epics"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
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
	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	epics, err := h.epics.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppEpics(epics), http.StatusOK)
	return nil
}
