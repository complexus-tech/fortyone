package objectivesgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	objectives    *objectives.Service
	keyResults    *keyresults.Service
	okrActivities *okractivities.Service
	attachments   *attachments.Service
	cache         *cache.Service
	log           *logger.Logger
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
)

func New(objectives *objectives.Service, keyResults *keyresults.Service, okrActivities *okractivities.Service, attachments *attachments.Service, cacheService *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		objectives:    objectives,
		keyResults:    keyResults,
		okrActivities: okrActivities,
		attachments:   attachments,
		cache:         cacheService,
		log:           log,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, 24*time.Hour)
	if err != nil {
		return ""
	}
	return resolved
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.List")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
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

	filtersStr := ""
	for k, v := range filters {
		filtersStr += fmt.Sprintf("%s:%v;", k, v)
	}

	cacheKey := cache.ObjectiveListCacheKey(workspace.ID, filtersStr)
	var cachedObjectives []objectives.CoreObjective

	if err := h.cache.Get(ctx, cacheKey, &cachedObjectives); err == nil {
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppObjectives(cachedObjectives), http.StatusOK)
		return nil
	}

	objectivesList, err := h.objectives.List(ctx, workspace.ID, userID, filters)
	if err != nil {
		return err
	}

	if err := h.cache.Set(ctx, cacheKey, objectivesList, cache.ListTTL); err != nil {
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppObjectives(objectivesList), http.StatusOK)
	return nil
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Get")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	cacheKey := cache.ObjectiveDetailCacheKey(workspace.ID, objID)
	var cachedObjective objectives.CoreObjective

	if err := h.cache.Get(ctx, cacheKey, &cachedObjective); err == nil {
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppObjective(cachedObjective), http.StatusOK)
		return nil
	}

	objective, err := h.objectives.Get(ctx, objID, workspace.ID)
	if err != nil {
		if errors.Is(err, objectives.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	if err := h.cache.Set(ctx, cacheKey, objective, cache.DetailTTL); err != nil {
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppObjective(objective), http.StatusOK)
	return nil
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Update")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	var uo AppUpdateObjective
	if err := web.Decode(r, &uo); err != nil {
		return err
	}

	comment := ""
	if uo.Comment != nil {
		comment = *uo.Comment
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
		updates["start_date"] = *uo.StartDate.TimePtr()
	}
	if uo.EndDate != nil {
		updates["end_date"] = *uo.EndDate.TimePtr()
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

	if err := h.objectives.Update(ctx, objID, workspace.ID, userID, comment, updates); err != nil {
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

	cacheKeys := cache.InvalidateObjectiveKeys(workspace.ID, objID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Delete")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	cacheKeys := cache.InvalidateObjectiveKeys(workspace.ID, objID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	if err := h.objectives.Delete(ctx, objID, workspace.ID); err != nil {
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

func (h *Handlers) GetKeyResults(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.GetKeyResults")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	cacheKey := cache.KeyResultsListCacheKey(workspace.ID, objID)
	var cachedKeyResults []keyresults.CoreKeyResult

	if err := h.cache.Get(ctx, cacheKey, &cachedKeyResults); err == nil {
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppKeyResults(cachedKeyResults), http.StatusOK)
		return nil
	}

	krs, err := h.keyResults.List(ctx, objID, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	if err := h.cache.Set(ctx, cacheKey, krs, cache.ListTTL); err != nil {
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppKeyResults(krs), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.Create")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var newObj AppNewObjective
	if err := web.Decode(r, &newObj); err != nil {
		return err
	}

	var keyResults []keyresults.CoreNewKeyResult
	for _, kr := range newObj.KeyResults {
		keyResults = append(keyResults, keyresults.CoreNewKeyResult{
			Name:            kr.Name,
			MeasurementType: kr.MeasurementType,
			StartValue:      kr.StartValue,
			CurrentValue:    kr.CurrentValue,
			TargetValue:     kr.TargetValue,
			CreatedBy:       userID,
			Lead:            kr.Lead,
			Contributors:    kr.Contributors,
			StartDate:       kr.StartDate.TimePtr(),
			EndDate:         kr.EndDate.TimePtr(),
		})
	}

	objective, createdKRs, err := h.objectives.Create(ctx, toCoreNewObjective(newObj, userID), workspace.ID, keyResults)
	if err != nil {
		if errors.Is(err, objectives.ErrNameExists) {
			web.RespondError(ctx, w, err, http.StatusConflict)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	listCachePattern := fmt.Sprintf(cache.ObjectiveListKey+"*", workspace.ID.String())
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

func (h *Handlers) GetAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectivesgrp.handlers.GetAnalytics")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	analytics, err := h.objectives.GetAnalytics(ctx, objID, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	for i := range analytics.TeamAllocation {
		if analytics.TeamAllocation[i].AvatarURL != nil {
			resolved := h.resolveUserAvatarURL(ctx, *analytics.TeamAllocation[i].AvatarURL)
			if resolved == "" {
				analytics.TeamAllocation[i].AvatarURL = nil
			} else {
				analytics.TeamAllocation[i].AvatarURL = &resolved
			}
		}
	}

	web.Respond(ctx, w, toAppObjectiveAnalytics(analytics), http.StatusOK)
	return nil
}

func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID := web.Params(r, "id")
	objID, err := uuid.Parse(objectiveID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidObjectiveID, http.StatusBadRequest)
		return nil
	}

	// Use web.GetFilters like your existing List method
	var af AppFilters
	_, err = web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	// Use the typed af struct directly
	page := af.Page
	if page <= 0 {
		page = 1
	}

	pageSize := af.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	activities, hasMore, err := h.okrActivities.GetObjectiveActivities(ctx, objID, page, pageSize)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	for i := range activities {
		activities[i].User.AvatarURL = h.resolveUserAvatarURL(ctx, activities[i].User.AvatarURL)
	}

	response := map[string]any{
		"activities": toAppObjectiveActivities(activities),
		"pagination": map[string]any{
			"page":     page,
			"pageSize": pageSize,
			"hasMore":  hasMore,
		},
	}

	web.Respond(ctx, w, response, http.StatusOK)
	return nil
}
