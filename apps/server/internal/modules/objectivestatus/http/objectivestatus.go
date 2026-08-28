package objectivestatushttp

import (
	"context"
	"errors"
	"net/http"

	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidStatusID    = errors.New("objective status id is not in its proper form")
)

type Handlers struct {
	objectiveStatus *objectivestatus.Service
}

func New(objectiveStatus *objectivestatus.Service) *Handlers {
	return &Handlers{
		objectiveStatus: objectiveStatus,
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req NewObjectiveStatus
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	status, err := h.objectiveStatus.Create(ctx, actorID, workspace.ID, toCoreNewObjectiveStatus(req))
	if err != nil {
		return web.RespondError(ctx, w, err, objectiveStatusErrorStatus(err))
	}

	return web.Respond(ctx, w, toAppObjectiveStatus(status), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	statusIdParam := web.Params(r, "statusId")
	statusId, err := uuid.Parse(statusIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStatusID, http.StatusBadRequest)
	}

	var req UpdateObjectiveStatus
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	status, err := h.objectiveStatus.Update(ctx, actorID, workspace.ID, statusId, toCoreUpdateObjectiveStatus(req))
	if err != nil {
		return web.RespondError(ctx, w, err, objectiveStatusErrorStatus(err))
	}

	return web.Respond(ctx, w, toAppObjectiveStatus(status), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	statusIdParam := web.Params(r, "statusId")
	statusId, err := uuid.Parse(statusIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStatusID, http.StatusBadRequest)
	}

	if err := h.objectiveStatus.Delete(ctx, actorID, workspace.ID, statusId); err != nil {
		return web.RespondError(ctx, w, err, objectiveStatusErrorStatus(err))
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	statuses, err := h.objectiveStatus.ListForMember(ctx, actorID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppObjectiveStatuses(statuses), http.StatusOK)
}

func objectiveStatusErrorStatus(err error) int {
	switch {
	case errors.Is(err, objectivestatus.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, objectivestatus.ErrNoFields), errors.Is(err, objectivestatus.ErrInvalidOrder):
		return http.StatusBadRequest
	case errors.Is(err, objectivestatus.ErrStatusHasObjectives), errors.Is(err, objectivestatus.ErrLastInCategory):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
