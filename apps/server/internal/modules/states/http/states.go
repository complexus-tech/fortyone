package stateshttp

import (
	"context"
	"errors"
	"net/http"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidStateID     = errors.New("state id is not in its proper form")
)

type Handlers struct {
	states *states.Service
}

func New(states *states.Service) *Handlers {
	return &Handlers{
		states: states,
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req NewState
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.states.Create(ctx, workspace.ID, toCoreNewState(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppState(state), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	stateIdParam := web.Params(r, "stateId")
	stateId, err := uuid.Parse(stateIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStateID, http.StatusBadRequest)
	}

	var req UpdateState
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.states.Update(ctx, workspace.ID, stateId, toCoreUpdateState(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppState(state), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	stateIdParam := web.Params(r, "stateId")
	stateId, err := uuid.Parse(stateIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStateID, http.StatusBadRequest)
	}

	if err := h.states.Delete(ctx, workspace.ID, stateId); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIdParam := r.URL.Query().Get("teamId")
	if teamIdParam != "" {
		teamId, err := uuid.Parse(teamIdParam)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid team ID"), http.StatusBadRequest)
		}

		states, err := h.states.TeamList(ctx, workspace.ID, teamId)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}

		return web.Respond(ctx, w, toAppStates(states), http.StatusOK)
	}

	states, err := h.states.List(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppStates(states), http.StatusOK)
}
