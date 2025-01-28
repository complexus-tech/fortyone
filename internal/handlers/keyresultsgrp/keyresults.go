package keyresultsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// Set of error variables for key result operations
var (
	ErrInvalidKeyResultID = errors.New("key result id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
)

// Handlers manages the set of key result endpoints
type Handlers struct {
	keyResults *keyresults.Service
	log        *logger.Logger
}

// New creates a new key results handlers
func New(keyResults *keyresults.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		keyResults: keyResults,
		log:        log,
	}
}

// Create adds a new key result to the system
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	var nkr AppNewKeyResult
	if err := web.Decode(r, &nkr); err != nil {
		return err
	}

	if err := nkr.Validate(); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	kr, err := h.keyResults.Create(ctx, toCoreNewKeyResult(nkr))
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppKeyResult(kr), http.StatusCreated)
	return nil
}

// Update modifies a key result in the system
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	keyResultID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var ukr AppUpdateKeyResult
	if err := web.Decode(r, &ukr); err != nil {
		return err
	}

	updates := make(map[string]any)
	if ukr.Name != "" {
		updates["name"] = ukr.Name
	}
	if ukr.MeasurementType != "" {
		updates["measurement_type"] = ukr.MeasurementType
	}
	if ukr.StartValue != nil {
		updates["start_value"] = ukr.StartValue
	}
	if ukr.TargetValue != nil {
		updates["target_value"] = ukr.TargetValue
	}

	if err := h.keyResults.Update(ctx, id, wsID, updates); err != nil {
		if errors.Is(err, keyresults.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Delete removes a key result from the system
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	keyResultID := web.Params(r, "id")
	workspaceID := web.Params(r, "workspaceId")

	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
		return nil
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.keyResults.Delete(ctx, id, wsID); err != nil {
		if errors.Is(err, keyresults.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// List returns all key results for an objective
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "objectiveId")
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
