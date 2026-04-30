package integrationrequestshttp

import (
	"context"
	"errors"
	"net/http"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	requests *integrationrequests.Service
	log      *logger.Logger
}

func New(requests *integrationrequests.Service, log *logger.Logger) *Handlers {
	return &Handlers{requests: requests, log: log}
}

func (h *Handlers) ListTeamRequests(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	status := r.URL.Query().Get("status")
	requests, err := h.requests.ListByTeam(ctx, workspace.ID, teamID, integrationrequests.CoreListRequestsFilter{Status: status})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppRequests(requests), http.StatusOK)
}

func (h *Handlers) GetRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	requestID, err := uuid.Parse(web.Params(r, "requestId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	request, err := h.requests.Get(ctx, workspace.ID, requestID)
	if err != nil {
		status := http.StatusInternalServerError
		if integrationrequests.IsNotFound(err) {
			status = http.StatusNotFound
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, toAppRequest(request), http.StatusOK)
}

func (h *Handlers) AcceptRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	requestID, err := uuid.Parse(web.Params(r, "requestId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	request, err := h.requests.Accept(ctx, workspace.ID, requestID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppRequest(request), http.StatusOK)
}

func (h *Handlers) DeclineRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	requestID, err := uuid.Parse(web.Params(r, "requestId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	request, err := h.requests.Decline(ctx, workspace.ID, requestID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppRequest(request), http.StatusOK)
}

func requestErrorStatus(err error) int {
	switch {
	case integrationrequests.IsNotFound(err):
		return http.StatusNotFound
	case errors.Is(err, integrationrequests.ErrRequestNotPending):
		return http.StatusConflict
	case errors.Is(err, integrationrequests.ErrUnsupportedProvider):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
