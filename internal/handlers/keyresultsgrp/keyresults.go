package keyresultsgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidKeyResultID = errors.New("key result id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
)

type Handlers struct {
	keyResults    *keyresults.Service
	okrActivities *okractivities.Service
	log           *logger.Logger
	cache         *cache.Service
}

func New(keyResults *keyresults.Service, okrActivities *okractivities.Service, cache *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		keyResults:    keyResults,
		okrActivities: okrActivities,
		log:           log,
		cache:         cache,
	}
}

func (h *Handlers) invalidateCache(ctx context.Context, workspaceID uuid.UUID) {
	cacheKeys := cache.InvalidateKeyResultKeys(workspaceID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
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

	h.invalidateCache(ctx, workspace.ID)

	web.Respond(ctx, w, toAppKeyResult(kr), http.StatusCreated)
	return nil
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	keyResultID := web.Params(r, "id")
	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
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
	if ukr.Lead != nil {
		updates["lead"] = ukr.Lead
	}
	if ukr.StartDate != nil {
		updates["start_date"] = ukr.StartDate
	}
	if ukr.EndDate != nil {
		updates["end_date"] = ukr.EndDate
	}
	if ukr.Contributors != nil {
		updates["contributors"] = *ukr.Contributors
	}

	if err := h.keyResults.Update(ctx, id, workspace.ID, updates); err != nil {
		if errors.Is(err, keyresults.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	h.invalidateCache(ctx, workspace.ID)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	keyResultID := web.Params(r, "id")
	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
		return nil
	}

	if err := h.keyResults.Delete(ctx, id, workspace.ID); err != nil {
		if errors.Is(err, keyresults.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	h.invalidateCache(ctx, workspace.ID)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "objectiveId")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	krs, err := h.keyResults.List(ctx, objID, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppKeyResults(krs), http.StatusOK)
	return nil
}

func (h *Handlers) ListPaginated(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "keyresultsgrp.handlers.ListPaginated")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters := parseKeyResultFilters(r, workspace.ID, userID)

	response, err := h.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppKeyResultListResponse(response), http.StatusOK)
	return nil
}

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

	if teamIDs := query["teamIds"]; len(teamIDs) > 0 {
		filters.TeamIDs = parseUUIDArray(teamIDs)
	}

	if objectiveIDs := query["objectiveIds"]; len(objectiveIDs) > 0 {
		filters.ObjectiveIDs = parseUUIDArray(objectiveIDs)
	}

	if measurementTypes := query["measurementTypes"]; len(measurementTypes) > 0 {
		filters.MeasurementTypes = measurementTypes
	}

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

func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	keyResultID := web.Params(r, "id")
	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
		return nil
	}

	// Use your existing pagination helper functions
	page := getIntParam(r.URL.Query(), "page", 1)
	pageSize := getIntParam(r.URL.Query(), "pageSize", 20)

	activities, hasMore, err := h.okrActivities.GetKeyResultActivities(ctx, id, page, pageSize)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	response := map[string]any{
		"activities": toAppKeyResultActivities(activities),
		"pagination": map[string]any{
			"page":     page,
			"pageSize": pageSize,
			"hasMore":  hasMore,
		},
	}

	web.Respond(ctx, w, response, http.StatusOK)
	return nil
}
