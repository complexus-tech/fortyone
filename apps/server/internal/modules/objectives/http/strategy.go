package objectiveshttp

import (
	"context"
	"net/http"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetStrategyMap(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	strategy, err := h.objectives.GetStrategyMap(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppStrategyMap(strategy), http.StatusOK)
}

func (h *Handlers) UpdateStrategy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var req AppStrategyUpdate
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.objectives.UpdateStrategy(ctx, workspace.ID, objectives.CoreStrategyUpdate{UltimateGoal: req.UltimateGoal, Description: req.Description}); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) CreateStrategicPillar(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var req AppNewStrategicPillar
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	pillar, err := h.objectives.CreateStrategicPillar(ctx, workspace.ID, objectives.CoreNewStrategicPillar{Name: req.Name, Description: req.Description, OrderIndex: req.OrderIndex})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppStrategicPillar(pillar), http.StatusCreated)
}

func (h *Handlers) UpdateStrategicPillar(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	pillarID, err := uuid.Parse(web.Params(r, "pillarId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var req AppUpdateStrategicPillar
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	pillar, err := h.objectives.UpdateStrategicPillar(ctx, workspace.ID, pillarID, objectives.CoreUpdateStrategicPillar{Name: req.Name, Description: req.Description, OrderIndex: req.OrderIndex})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppStrategicPillar(pillar), http.StatusOK)
}

func (h *Handlers) DeleteStrategicPillar(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	pillarID, err := uuid.Parse(web.Params(r, "pillarId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.objectives.DeleteStrategicPillar(ctx, workspace.ID, pillarID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) AlignObjective(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var req AppObjectiveAlignment
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.objectives.AlignObjective(ctx, workspace.ID, objectiveID, req.PillarID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
