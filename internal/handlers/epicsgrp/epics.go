package epicsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/epics"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	epics *epics.Service
	// audit  *audit.Service
}

// New constructs a new epics handlers instance.
func New(epics *epics.Service) *Handlers {
	return &Handlers{
		epics: epics,
	}
}

// List returns a list of epics.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	epics, err := h.epics.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppEpics(epics), http.StatusOK)
	return nil
}
