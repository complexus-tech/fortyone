package mayahttp

import (
	"context"
	"net/http"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) CreateWorkPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if ok, err := h.workspaceCanUseMaya(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	} else if !ok {
		return web.RespondError(ctx, w, ErrMayaAccessRequired, http.StatusPaymentRequired)
	}

	var req AppCreateWorkPlanRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	windowStart := h.now().UTC()
	if req.WindowStart != nil {
		windowStart = req.WindowStart.UTC()
	}
	windowEnd := windowStart.Add(defaultPlanningWindow)
	if req.WindowEnd != nil {
		windowEnd = req.WindowEnd.UTC()
	}

	plan, err := h.service.CreateWorkPlan(ctx, maya.CreateWorkPlanInput{
		WorkspaceID:      workspace.ID,
		StoryID:          req.StoryID,
		TriggeredBy:      userID,
		Trigger:          maya.RunTriggerManual,
		WindowStart:      windowStart,
		WindowEnd:        windowEnd,
		DurationMinutes:  req.DurationMinutes,
		CandidateUserIDs: req.CandidateUserIDs,
		AutoApply:        req.AutoApply,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppWorkPlan(plan), http.StatusCreated)
}

func (h *Handlers) ApplyWorkPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if ok, err := h.workspaceCanUseMaya(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	} else if !ok {
		return web.RespondError(ctx, w, ErrMayaAccessRequired, http.StatusPaymentRequired)
	}
	runID, err := uuid.Parse(web.Params(r, "runId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	plan, err := h.service.ApplyWorkPlan(ctx, maya.ApplyWorkPlanInput{
		WorkspaceID: workspace.ID,
		RunID:       runID,
		TriggeredBy: userID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, toAppWorkPlan(plan), http.StatusOK)
}
