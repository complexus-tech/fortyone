package reportshttp

import (
	"context"
	"net/http"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) TrackWorkspaceAnalyticsEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.TrackWorkspaceAnalyticsEvent")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppTrackWorkspaceAnalyticsEventRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	input := reports.CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		EventName:   req.EventName,
		Surface:     req.Surface,
		TeamID:      req.TeamID,
		StoryID:     req.StoryID,
		ObjectiveID: req.ObjectiveID,
		SprintID:    req.SprintID,
		KeyResultID: req.KeyResultID,
		Properties:  req.Properties,
	}
	if req.OccurredAt != nil {
		input.OccurredAt = *req.OccurredAt
	}

	event, err := h.reports.TrackWorkspaceAnalyticsEvent(ctx, input)
	if err != nil {
		return respondReportError(ctx, w, err)
	}

	return web.Respond(ctx, w, AppTrackWorkspaceAnalyticsEventResponse{
		EventName:  event.EventName,
		Surface:    event.Surface,
		OccurredAt: event.OccurredAt,
	}, http.StatusCreated)
}
