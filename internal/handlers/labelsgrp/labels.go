package labelsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/labels"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidLabelID     = errors.New("label id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	labels *labels.Service
	log    *logger.Logger
}

func New(labels *labels.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		labels: labels,
		log:    log,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	labels, err := h.labels.GetLabels(ctx, workspace.ID, filters)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, toAppLabels(labels), http.StatusOK)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	labelIdParam := web.Params(r, "id")
	labelId, err := uuid.Parse(labelIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLabelID, http.StatusBadRequest)
		return nil
	}

	label, err := h.labels.GetLabel(ctx, labelId, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, toAppLabel(label), http.StatusOK)
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppNewLabel
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	input := labels.CoreNewLabel{
		Name:        req.Name,
		TeamID:      req.TeamID,
		WorkspaceID: workspace.ID,
		Color:       req.Color,
	}

	label, err := h.labels.CreateLabel(ctx, input)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, toAppLabel(label), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	labelIdParam := web.Params(r, "id")
	labelId, err := uuid.Parse(labelIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLabelID, http.StatusBadRequest)
		return nil
	}

	var req AppUpdateLabel
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	label, err := h.labels.UpdateLabel(ctx, labelId, workspace.ID, req.Name, req.Color)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, toAppLabel(label), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	labelIdParam := web.Params(r, "id")
	labelId, err := uuid.Parse(labelIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLabelID, http.StatusBadRequest)
		return nil
	}

	if err := h.labels.DeleteLabel(ctx, labelId, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
