package statesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidStateID     = errors.New("state id is not in its proper form")
)

type Handlers struct {
	states *states.Service
	// audit  *audit.Service
}

// New constructs a new states handlers instance.
func New(states *states.Service) *Handlers {
	return &Handlers{
		states: states,
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	var req NewState
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.states.Create(ctx, workspaceId, toCoreNewState(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppState(state), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	stateIdParam := web.Params(r, "stateId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	stateId, err := uuid.Parse(stateIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStateID, http.StatusBadRequest)
	}

	var req UpdateState
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.states.Update(ctx, workspaceId, stateId, toCoreUpdateState(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppState(state), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	stateIdParam := web.Params(r, "stateId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	stateId, err := uuid.Parse(stateIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStateID, http.StatusBadRequest)
	}

	if err := h.states.Delete(ctx, workspaceId, stateId); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// List returns a list of states.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	states, err := h.states.List(ctx, workspaceId, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppStates(states), http.StatusOK)
}
