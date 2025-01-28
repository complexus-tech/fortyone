package objectivesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	objectives *objectives.Service
	keyResults *keyresults.Service
	// audit  *audit.Service
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
)

// New constructs a new objectives handlers instance.
func New(objectives *objectives.Service, keyResults *keyresults.Service) *Handlers {
	return &Handlers{
		objectives: objectives,
		keyResults: keyResults,
	}
}

// List returns a list of objectives.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	objectives, err := h.objectives.List(ctx, workspaceId, filters)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppObjectives(objectives), http.StatusOK)
	return nil
}

// Get returns an objective by ID.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	objective, err := h.objectives.Get(ctx, objID, wsID)
	if err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppObjective(objective), http.StatusOK)
	return nil
}

// Update updates an objective in the system
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var uo AppUpdateObjective
	if err := web.Decode(r, &uo); err != nil {
		return err
	}

	updates := make(map[string]any)
	if uo.Name != nil {
		updates["name"] = *uo.Name
	}
	if uo.Description != nil {
		updates["description"] = *uo.Description
	}
	if uo.LeadUser != nil {
		updates["lead_user_id"] = *uo.LeadUser
	}
	if uo.StartDate != nil {
		updates["start_date"] = *uo.StartDate
	}
	if uo.EndDate != nil {
		updates["end_date"] = *uo.EndDate
	}
	if uo.IsPrivate != nil {
		updates["is_private"] = *uo.IsPrivate
	}
	if uo.Status != nil {
		updates["status_id"] = *uo.Status
	}
	if uo.Priority != nil {
		updates["priority"] = *uo.Priority
	}

	if err := h.objectives.Update(ctx, objID, wsID, updates); err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Delete removes an objective from the system
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.objectives.Delete(ctx, objID, wsID); err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// GetKeyResults returns all key results for an objective.
func (h *Handlers) GetKeyResults(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	krs, err := h.keyResults.List(ctx, objID, wsID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppKeyResults(krs), http.StatusOK)
	return nil
}
