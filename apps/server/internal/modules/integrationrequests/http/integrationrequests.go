package integrationrequestshttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

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
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	status := r.URL.Query().Get("status")
	provider := r.URL.Query().Get("provider")
	priority := r.URL.Query().Get("priority")
	assigneeID, err := optionalUUIDQuery(r, "assigneeId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	createdAfter, err := optionalDateQuery(r, "createdAfter")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	createdBefore, err := optionalDateQuery(r, "createdBefore")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page, pageSize := paginationParams(r, defaultRequestsPageSize, maxRequestsPageSize)
	filter := integrationrequests.CoreListRequestsFilter{
		Status:        status,
		Provider:      provider,
		Priority:      priority,
		AssigneeID:    assigneeID,
		CreatedAfter:  createdAfter,
		CreatedBefore: createdBefore,
		Page:          page,
		PageSize:      pageSize + 1,
	}
	requests, err := h.requests.ListByTeam(ctx, workspace.ID, teamID, filter)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	totalCount, err := h.requests.CountByTeam(ctx, workspace.ID, teamID, filter)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	hasMore := len(requests) > pageSize
	if hasMore {
		requests = requests[:pageSize]
	}
	return web.Respond(ctx, w, toAppRequestsResponse(requests, page, pageSize, totalCount, hasMore), http.StatusOK)
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

func (h *Handlers) UpdateRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
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
	request, err := h.requests.UpdatePending(ctx, workspace.ID, requestID, integrationrequests.CoreUpdateRequestInput{
		Title:         input.Title,
		Description:   input.Description,
		StatusID:      input.StatusID,
		Priority:      input.Priority,
		AssigneeID:    input.AssigneeID,
		EstimateValue: input.EstimateValue,
		ObjectiveID:   input.ObjectiveID,
		KeyResultID:   input.KeyResultID,
		SprintID:      input.SprintID,
		StartDate:     input.StartDate.TimePtr(),
		EndDate:       input.EndDate.TimePtr(),
	})
	if err != nil {
		return web.RespondError(ctx, w, err, requestErrorStatus(err))
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
	case errors.Is(err, integrationrequests.ErrRequestNotPending):
		return http.StatusConflict
	case errors.Is(err, integrationrequests.ErrUnsupportedProvider):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func paginationParams(r *http.Request, defaultPageSize, maxPageSize int) (int, int) {
	page := 1
	pageSize := defaultPageSize
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func optionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalDateQuery(r *http.Request, key string) (*time.Time, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
