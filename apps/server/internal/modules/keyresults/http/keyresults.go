package keyresultshttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidKeyResultID = errors.New("key result id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidObjectiveID = errors.New("objective id is not in its proper form")
	ErrInvalidFilters     = errors.New("invalid key result filters")
)

type Handlers struct {
	keyResults    *keyresults.Service
	okrActivities *okractivities.Service
	attachments   *attachments.Service
	log           *logger.Logger
	cache         *cache.Service
}

func New(keyResults *keyresults.Service, okrActivities *okractivities.Service, attachments *attachments.Service, cache *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		keyResults:    keyResults,
		okrActivities: okrActivities,
		attachments:   attachments,
		log:           log,
		cache:         cache,
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

func (h *Handlers) invalidateObjectiveCache(ctx context.Context, workspaceID, objectiveID uuid.UUID) {
	if h.cache == nil {
		return
	}
	for _, key := range cache.InvalidateObjectiveKeys(workspaceID, objectiveID) {
		if strings.Contains(key, "*") {
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				if h.log != nil {
					h.log.Error(ctx, "failed to delete objective cache pattern", "key", key, "error", err)
				}
			}
			continue
		}
		if err := h.cache.Delete(ctx, key); err != nil {
			if h.log != nil {
				h.log.Error(ctx, "failed to delete objective cache", "key", key, "error", err)
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

	kr, err := h.keyResults.Create(ctx, toCoreNewKeyResult(nkr, userID), workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, keyResultErrorStatus(err))
		return nil
	}

	h.invalidateObjectiveCache(ctx, workspace.ID, nkr.ObjectiveID)

	web.Respond(ctx, w, toAppKeyResult(kr), http.StatusCreated)
	return nil
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

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
	comment := ""
	if ukr.Comment != nil {
		comment = *ukr.Comment
	}

	patch := toKeyResultPatch(ukr)
	currentKeyResult, err := h.keyResults.GetForActor(ctx, id, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, keyResultErrorStatus(err))
	}

	if err := h.keyResults.UpdateIntent(ctx, id, workspace.ID, userID, patch, comment); err != nil {
		web.RespondError(ctx, w, err, keyResultErrorStatus(err))
		return nil
	}

	h.invalidateObjectiveCache(ctx, workspace.ID, currentKeyResult.ObjectiveID)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

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
	currentKeyResult, err := h.keyResults.GetForActor(ctx, id, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, keyResultErrorStatus(err))
	}

	if err := h.keyResults.Delete(ctx, id, workspace.ID, userID); err != nil {
		web.RespondError(ctx, w, err, keyResultErrorStatus(err))
		return nil
	}

	h.invalidateObjectiveCache(ctx, workspace.ID, currentKeyResult.ObjectiveID)

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
		web.RespondError(ctx, w, err, keyResultErrorStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppKeyResults(krs), http.StatusOK)
	return nil
}

func (h *Handlers) ListPaginated(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "keyresultshttp.handlers.ListPaginated")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters, err := parseKeyResultFilters(r, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	response, err := h.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		web.RespondError(ctx, w, err, keyResultErrorStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppKeyResultListResponse(response), http.StatusOK)
	return nil
}

func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	keyResultID := web.Params(r, "id")
	id, err := uuid.Parse(keyResultID)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidKeyResultID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if _, err := h.keyResults.GetForActor(ctx, id, workspace.ID, userID); err != nil {
		return web.RespondError(ctx, w, err, keyResultErrorStatus(err))
	}

	pageRequest, err := pagination.ParseOffsetQuery(r.URL.Query(), pagination.OffsetQueryConfig{
		DefaultPageSize: okractivities.DefaultPageSize,
		MaximumPageSize: okractivities.MaximumPageSize,
		MaximumOffset:   maximumKeyResultQueryOffset,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page := pageRequest.Page
	pageSize := pageRequest.PageSize

	activities, hasMore, err := h.okrActivities.GetKeyResultActivities(ctx, id, workspace.ID, page, pageSize)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, okractivities.ErrInvalid) {
			status = http.StatusBadRequest
		} else if errors.Is(err, okractivities.ErrForbidden) {
			status = http.StatusForbidden
		}
		web.RespondError(ctx, w, err, status)
		return nil
	}

	for i := range activities {
		activities[i].User.AvatarURL = h.resolveUserAvatarURL(ctx, activities[i].User.AvatarURL)
	}

	response := AppKeyResultActivitiesResponse{
		Activities: toAppKeyResultActivities(activities),
		Pagination: AppKeyResultActivityPagination{
			Page: page, PageSize: pageSize, HasMore: hasMore,
		},
	}

	web.Respond(ctx, w, response, http.StatusOK)
	return nil
}

func toKeyResultPatch(update AppUpdateKeyResult) keyresults.KeyResultPatch {
	patch := keyresults.KeyResultPatch{}
	if update.Name != "" {
		patch.Name = keyresults.SetField(update.Name)
	}
	if update.MeasurementType != "" {
		patch.MeasurementType = keyresults.SetField(update.MeasurementType)
	}
	if update.StartValue != nil {
		patch.StartValue = keyresults.SetField(*update.StartValue)
	}
	if update.CurrentValue != nil {
		patch.CurrentValue = keyresults.SetField(*update.CurrentValue)
	}
	if update.TargetValue != nil {
		patch.TargetValue = keyresults.SetField(*update.TargetValue)
	}
	if update.ClearLead {
		patch.Lead = keyresults.ClearField[uuid.UUID]()
	} else if update.Lead != nil {
		patch.Lead = keyresults.SetField(update.Lead)
	}
	if update.Contributors != nil {
		patch.Contributors = keyresults.SetField(*update.Contributors)
	}
	if update.StartDate != nil {
		patch.StartDate = keyresults.SetField(update.StartDate.TimePtr())
	}
	if update.EndDate != nil {
		patch.EndDate = keyresults.SetField(update.EndDate.TimePtr())
	}
	return patch
}

func keyResultErrorStatus(err error) int {
	switch {
	case errors.Is(err, keyresults.ErrInvalid), errors.Is(err, keyresults.ErrInvalidReference):
		return http.StatusBadRequest
	case errors.Is(err, keyresults.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, keyresults.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, keyresults.ErrVersionConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
