package keyresultsgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
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
	cache      *cache.Service
}

// New creates a new key results handlers
func New(keyResults *keyresults.Service, cache *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		keyResults: keyResults,
		log:        log,
		cache:      cache,
	}
}

// invalidateCache invalidates all relevant caches for a key result operation
func (h *Handlers) invalidateCache(ctx context.Context, workspaceID uuid.UUID) {
	cacheKeys := cache.InvalidateKeyResultKeys(workspaceID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			// Handle exact key deletion
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
	}
}

// Create adds a new key result to the system
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID := web.Params(r, "workspaceId")
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var nkr AppNewKeyResult
	if err := web.Decode(r, &nkr); err != nil {
		return err
	}

	if err := nkr.Validate(); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	kr, err := h.keyResults.Create(ctx, toCoreNewKeyResult(nkr, userID))
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Invalidate cache after successful creation
	h.invalidateCache(ctx, wsID)

	web.Respond(ctx, w, toAppKeyResult(kr), http.StatusCreated)
	return nil
}

// Update modifies a key result in the system
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

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
	if ukr.CurrentValue != nil {
		updates["current_value"] = ukr.CurrentValue
	}
	if ukr.TargetValue != nil {
		updates["target_value"] = ukr.TargetValue
	}
	updates["last_updated_by"] = userID

	if err := h.keyResults.Update(ctx, id, wsID, updates); err != nil {
		if errors.Is(err, keyresults.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Invalidate cache after successful update
	h.invalidateCache(ctx, wsID)

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

	// Invalidate cache after successful deletion
	h.invalidateCache(ctx, wsID)

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

// ListPaginated returns paginated key results for a workspace
func (h *Handlers) ListPaginated(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "keyresultsgrp.handlers.ListPaginated")
	defer span.End()

	// Get current user
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID := web.Params(r, "workspaceId")
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Parse query parameters
	filters := parseKeyResultFilters(r, wsID, userID)

	// Call service
	response, err := h.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppKeyResultListResponse(response), http.StatusOK)
	return nil
}

// Helper function to parse filters from request
func parseKeyResultFilters(r *http.Request, workspaceID, userID uuid.UUID) keyresultsrepo.CoreKeyResultFilters {
	query := r.URL.Query()

	filters := keyresultsrepo.CoreKeyResultFilters{
		WorkspaceID:    workspaceID,
		CurrentUserID:  userID,
		Page:           getIntParam(query, "page", 1),
		PageSize:       getIntParam(query, "pageSize", 20),
		OrderBy:        getStringParam(query, "orderBy", "created_at"),
		OrderDirection: getStringParam(query, "orderDirection", "desc"),
	}

	// Parse team IDs
	if teamIDs := query["teamIds"]; len(teamIDs) > 0 {
		filters.TeamIDs = parseUUIDArray(teamIDs)
	}

	// Parse objective IDs
	if objectiveIDs := query["objectiveIds"]; len(objectiveIDs) > 0 {
		filters.ObjectiveIDs = parseUUIDArray(objectiveIDs)
	}

	// Parse measurement types
	if measurementTypes := query["measurementTypes"]; len(measurementTypes) > 0 {
		filters.MeasurementTypes = measurementTypes
	}

	// Parse dates
	if createdAfter := query.Get("createdAfter"); createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			filters.CreatedAfter = &t
		}
	}

	if createdBefore := query.Get("createdBefore"); createdBefore != "" {
		if t, err := time.Parse(time.RFC3339, createdBefore); err == nil {
			filters.CreatedBefore = &t
		}
	}

	return filters
}

// Helper functions for parsing query parameters
func getIntParam(query map[string][]string, key string, defaultValue int) int {
	if values := query[key]; len(values) > 0 {
		if val, err := strconv.Atoi(values[0]); err == nil {
			return val
		}
	}
	return defaultValue
}

func getStringParam(query map[string][]string, key, defaultValue string) string {
	if values := query[key]; len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func parseUUIDArray(values []string) []uuid.UUID {
	var uuids []uuid.UUID
	for _, val := range values {
		if id, err := uuid.Parse(val); err == nil {
			uuids = append(uuids, id)
		}
	}
	return uuids
}
