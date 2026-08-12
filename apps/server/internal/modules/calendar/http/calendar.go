package calendarhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	log     *logger.Logger
	service *calendar.Service
}

func New(log *logger.Logger, service *calendar.Service) *Handlers {
	return &Handlers{log: log, service: service}
}

func (h *Handlers) GetIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	connections, err := h.service.ListConnections(ctx, workspace.ID, &userID)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppIntegration(connections), http.StatusOK)
}

func (h *Handlers) CreateConnectSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	session, err := h.service.CreateConnectSession(ctx, workspace.ID, userID, workspace.Slug)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, AppCreateConnectSession{AuthURL: session.AuthURL}, http.StatusOK)
}

func (h *Handlers) SyncConnection(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	connectionID, err := uuid.Parse(web.Params(r, "connectionId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.SyncConnection(ctx, workspace.ID, userID, connectionID); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) RevokeConnection(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	connectionID, err := uuid.Parse(web.Params(r, "connectionId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.RevokeConnection(ctx, workspace.ID, userID, connectionID); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetSchedule(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	startAt, endAt, err := parseScheduleRange(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	schedule, err := h.service.ListCalendarView(ctx, workspace.ID, userID, startAt, endAt)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppSchedule(schedule), http.StatusOK)
}

func (h *Handlers) GetCalendarEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	eventID, err := uuid.Parse(web.Params(r, "eventId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	event, err := h.service.GetCalendarEvent(ctx, workspace.ID, userID, eventID)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppCalendarEvent(event), http.StatusOK)
}

func (h *Handlers) CreateScheduleBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var req AppScheduleBlockRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	block, err := h.service.CreateScheduleBlock(ctx, toCoreScheduleBlockInput(workspace.ID, userID, uuid.Nil, req))
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppScheduleBlock(block), http.StatusCreated)
}

func (h *Handlers) UpdateScheduleBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	blockID, err := uuid.Parse(web.Params(r, "blockId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var req AppScheduleBlockRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	block, err := h.service.UpdateScheduleBlock(ctx, toCoreScheduleBlockInput(workspace.ID, userID, blockID, req))
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppScheduleBlock(block), http.StatusOK)
}

func (h *Handlers) DeleteScheduleBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	blockID, err := uuid.Parse(web.Params(r, "blockId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.DeleteScheduleBlock(ctx, workspace.ID, userID, blockID); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) HandleGoogleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	providerError := r.URL.Query().Get("error")
	if providerError != "" {
		errorCode := "connection_failed"
		if providerError == "access_denied" {
			errorCode = "access_denied"
		}
		redirectURL, err := h.service.CalendarCallbackErrorURL(state, userID, errorCode)
		if err != nil {
			return web.RespondError(ctx, w, err, h.statusCode(err))
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return nil
	}
	_, redirectURL, err := h.service.CompleteConnect(ctx, userID, code, state)
	if err != nil {
		failureURL, redirectErr := h.service.CalendarCallbackErrorURL(state, userID, "connection_failed")
		if redirectErr != nil {
			return web.RespondError(ctx, w, err, h.statusCode(err))
		}
		http.Redirect(w, r, failureURL, http.StatusTemporaryRedirect)
		return nil
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) HandleGoogleNotification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	err := h.service.ProcessGoogleNotification(
		ctx,
		r.Header.Get("X-Goog-Channel-ID"),
		r.Header.Get("X-Goog-Resource-ID"),
		r.Header.Get("X-Goog-Resource-State"),
		r.Header.Get("X-Goog-Channel-Token"),
	)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handlers) statusCode(err error) int {
	switch {
	case errors.Is(err, calendar.ErrCalendarNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, calendar.ErrInvalidCalendarState):
		return http.StatusBadRequest
	case errors.Is(err, calendar.ErrCalendarNotFound):
		return http.StatusNotFound
	case errors.Is(err, calendar.ErrCalendarAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, calendar.ErrCalendarCredentialsIncomplete):
		return http.StatusBadRequest
	case errors.Is(err, calendar.ErrCalendarEventNotFound):
		return http.StatusNotFound
	case errors.Is(err, calendar.ErrCalendarSyncSuperseded):
		return http.StatusConflict
	case errors.Is(err, calendar.ErrInvalidScheduleRange):
		return http.StatusBadRequest
	case errors.Is(err, calendar.ErrInvalidScheduleBlock):
		return http.StatusBadRequest
	case errors.Is(err, calendar.ErrCalendarScheduleConflict):
		return http.StatusConflict
	case errors.Is(err, calendar.ErrCalendarScheduleBlockNotFound):
		return http.StatusNotFound
	case errors.Is(err, calendar.ErrInvalidCalendarNotification):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func parseScheduleRange(r *http.Request) (time.Time, time.Time, error) {
	startRaw := r.URL.Query().Get("start")
	endRaw := r.URL.Query().Get("end")
	startAt, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endAt, err := time.Parse(time.RFC3339, endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startAt, endAt, nil
}

func toCoreScheduleBlockInput(workspaceID, userID, blockID uuid.UUID, req AppScheduleBlockRequest) calendar.CoreScheduleBlockInput {
	isLocked := true
	if req.IsLocked != nil {
		isLocked = *req.IsLocked
	}
	return calendar.CoreScheduleBlockInput{
		ID:          blockID,
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     req.StoryID,
		BlockType:   calendar.ScheduleBlockType(req.BlockType),
		Title:       req.Title,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		IsLocked:    isLocked,
		Source:      calendar.ScheduleBlockSourceUser,
	}
}
