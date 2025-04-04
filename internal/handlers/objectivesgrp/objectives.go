package objectivesgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	objectives *objectives.Service
	keyResults *keyresults.Service
	cache      *cache.Service
	log        *logger.Logger
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
)

// New constructs a new objectives handlers instance.
func New(objectives *objectives.Service, keyResults *keyresults.Service, cacheService *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		objectives: objectives,
		keyResults: keyResults,
		cache:      cacheService,
		log:        log,
	}
}

// List returns a list of objectives.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.List")
	defer span.End()
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	// Try to get from cache first
	filtersStr := ""
	if filters != nil {
		// Convert filters to string representation for cache key
		// This is a simple implementation - you might want to make this more sophisticated
		for k, v := range filters {
			filtersStr += fmt.Sprintf("%s:%v;", k, v)
		}
	}
	cacheKey := cache.ObjectiveListCacheKey(workspaceId, filtersStr)
	var cachedObjectives []objectives.CoreObjective

	if err := h.cache.Get(ctx, cacheKey, &cachedObjectives); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppObjectives(cachedObjectives), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	objectivesList, err := h.objectives.List(ctx, workspaceId, userID, filters)
	if err != nil {
		return err
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, objectivesList, cache.ListTTL); err != nil {
		// Log error but continue - cache failure shouldn't affect response
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppObjectives(objectivesList), http.StatusOK)
	return nil
}

// Get returns an objective by ID.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Get")
	defer span.End()
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

	// Try to get from cache first
	cacheKey := cache.ObjectiveDetailCacheKey(wsID, objID)
	var cachedObjective objectives.CoreObjective

	if err := h.cache.Get(ctx, cacheKey, &cachedObjective); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppObjective(cachedObjective), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	objective, err := h.objectives.Get(ctx, objID, wsID)
	if err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, objective, cache.DetailTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppObjective(objective), http.StatusOK)
	return nil
}

// Update updates an objective in the system
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Update")
	defer span.End()
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
	if uo.Health != nil {
		health := objectives.ObjectiveHealth(*uo.Health)
		updates["health"] = health
	}

	if err := h.objectives.Update(ctx, objID, wsID, updates); err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		if errors.Is(err, objectives.ErrNameExists) {
			web.RespondError(ctx, w, err, http.StatusConflict)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Invalidate cache after successful update
	cacheKeys := cache.InvalidateObjectiveKeys(wsID, objID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			h.cache.DeleteByPattern(ctx, key)
		} else {
			// Handle exact key deletion
			h.cache.Delete(ctx, key)
		}
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Delete removes an objective from the system
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Delete")
	defer span.End()
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

	// Invalidate cache before deleting the objective
	cacheKeys := cache.InvalidateObjectiveKeys(wsID, objID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			h.cache.DeleteByPattern(ctx, key)
		} else {
			// Handle exact key deletion
			h.cache.Delete(ctx, key)
		}
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
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.GetKeyResults")
	defer span.End()
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

	// Try to get from cache first
	cacheKey := cache.KeyResultsListCacheKey(wsID, objID)
	var cachedKeyResults []keyresults.CoreKeyResult

	if err := h.cache.Get(ctx, cacheKey, &cachedKeyResults); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppKeyResults(cachedKeyResults), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	krs, err := h.keyResults.List(ctx, objID, wsID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, krs, cache.ListTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppKeyResults(krs), http.StatusOK)
	return nil
}

// Create creates a new objective with optional key results
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Create")
	defer span.End()
	workspaceID := web.Params(r, "workspaceId")
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var newObj AppNewObjective
	if err := web.Decode(r, &newObj); err != nil {
		return err
	}

	// Convert key results if they exist
	var keyResults []keyresults.CoreNewKeyResult
	for _, kr := range newObj.KeyResults {
		keyResults = append(keyResults, keyresults.CoreNewKeyResult{
			Name:            kr.Name,
			MeasurementType: kr.MeasurementType,
			StartValue:      kr.StartValue,
			CurrentValue:    kr.CurrentValue,
			TargetValue:     kr.TargetValue,
			CreatedBy:       userID,
		})
	}

	objective, createdKRs, err := h.objectives.Create(ctx, toCoreNewObjective(newObj, userID), wsID, keyResults)
	if err != nil {
		if errors.Is(err, objectives.ErrNameExists) {
			web.RespondError(ctx, w, err, http.StatusConflict)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// After successful creation, invalidate list cache
	listCachePattern := fmt.Sprintf(cache.ObjectiveListKey+"*", wsID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	response := struct {
		Objective  AppObjectiveList `json:"objective"`
		KeyResults []AppKeyResult   `json:"keyResults,omitempty"`
	}{
		Objective:  toAppObjective(objective),
		KeyResults: toAppKeyResults(createdKRs),
	}

	web.Respond(ctx, w, response, http.StatusCreated)
	return nil
}
