package labelsgrp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/labels"
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
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	labels, err := h.labels.GetLabels(ctx, workspaceId, filters)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppLabels(labels), http.StatusOK)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	labelIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	labelId, err := uuid.Parse(labelIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLabelID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	label, err := h.labels.GetLabel(ctx, labelId, workspaceId)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppLabel(label), http.StatusOK)
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var req AppNewLabel
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	input := labels.CoreNewLabel{
		Name:        req.Name,
		TeamID:      req.TeamID,
		WorkspaceID: workspaceId,
		Color:       req.Color,
	}

	label, err := h.labels.CreateLabel(ctx, input)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppLabel(label), http.StatusCreated)
}
