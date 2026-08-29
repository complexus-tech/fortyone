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

const defaultRequestsPageSize = 25
const maxRequestsPageSize = 100

func New(requests *integrationrequests.Service, log *logger.Logger) *Handlers {
	return &Handlers{requests: requests, log: log}
}

func (h *Handlers) ListTeamRequests(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	query, err := parseRequestListQuery(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	requests, err := h.requests.ListByTeam(ctx, workspace.ID, teamID, userID, query.Filter)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	totalCount, err := h.requests.CountByTeam(ctx, workspace.ID, teamID, userID, query.Filter)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	hasMore := len(requests) > query.PageSize
	if hasMore {
		requests = requests[:query.PageSize]
	}
	return web.Respond(ctx, w, toAppRequestsResponse(requests, query.Page, query.PageSize, totalCount, hasMore), http.StatusOK)
}

func (h *Handlers) GetRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	request, err := h.requests.GetForUser(ctx, workspace.ID, requestID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if integrationrequests.IsNotFound(err) {
			status = http.StatusNotFound
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, toAppRequest(request), http.StatusOK)
}

func (h *Handlers) UpdateRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppUpdateIntegrationRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	request, err := h.requests.UpdatePending(ctx, workspace.ID, requestID, userID, integrationrequests.CoreUpdateRequestInput{
		Title:                    input.Title,
		Description:              toCoreOptionalValue(input.Description),
		StatusID:                 toCoreOptionalValue(input.StatusID),
		Priority:                 input.Priority,
		AssigneeID:               toCoreOptionalValue(input.AssigneeID),
		EstimateValue:            toCoreOptionalValue(input.EstimateValue),
		EstimatedDurationMinutes: toCoreOptionalValue(input.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: toCoreOptionalValue(input.MinimumFocusBlockMinutes),
		ObjectiveID:              toCoreOptionalValue(input.ObjectiveID),
		KeyResultID:              toCoreOptionalValue(input.KeyResultID),
		SprintID:                 toCoreOptionalValue(input.SprintID),
		StartDate:                toCoreOptionalDate(input.StartDate),
		EndDate:                  toCoreOptionalDate(input.EndDate),
		LabelIDs:                 input.LabelIDs,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppRequest(request), http.StatusOK)
}

func (h *Handlers) GetRequestThreadActivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	activity, err := h.requests.GetThreadActivityForRequest(ctx, workspace.ID, requestID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppThreadActivity(activity), http.StatusOK)
}

func (h *Handlers) CreateRequestComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppCreateIntegrationRequestComment
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comment, err := h.requests.CreateComment(ctx, integrationrequests.CoreCreateCommentInput{
		WorkspaceID:          workspace.ID,
		RequestID:            requestID,
		AuthorID:             userID,
		ClientIdempotencyKey: input.IdempotencyKey,
		Body:                 input.Body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)
}

func (h *Handlers) GetStoryProviderThreads(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	threads, err := h.requests.ListProviderThreadsForStory(ctx, workspace.ID, storyID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	response := make([]AppProviderThread, 0, len(threads))
	for _, thread := range threads {
		response = append(response, toAppProviderThread(thread))
	}
	return web.Respond(ctx, w, response, http.StatusOK)
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

func (h *Handlers) AcceptAllTeamRequests(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.requests.AcceptAllPendingByTeam(ctx, workspace.ID, teamID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppBulkRequestResult(result), http.StatusOK)
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

func (h *Handlers) DeclineAllTeamRequests(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.requests.DeclineAllPendingByTeam(ctx, workspace.ID, teamID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
	}
	return web.Respond(ctx, w, toAppBulkRequestResult(result), http.StatusOK)
}

func requestErrorStatus(err error) int {
	switch {
	case integrationrequests.IsNotFound(err):
		return http.StatusNotFound
	case errors.Is(err, integrationrequests.ErrProviderThreadNotFound):
		return http.StatusNotFound
	case errors.Is(err, integrationrequests.ErrRequestNotPending):
		return http.StatusConflict
	case errors.Is(err, integrationrequests.ErrIdempotencyConflict):
		return http.StatusConflict
	case errors.Is(err, integrationrequests.ErrInvalidRequestProperty):
		return http.StatusBadRequest
	case errors.Is(err, integrationrequests.ErrUnsupportedProvider):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
