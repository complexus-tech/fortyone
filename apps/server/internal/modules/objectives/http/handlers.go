package objectiveshttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
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

func New(
	objectiveService *objectives.Service,
	keyResultService *keyresults.Service,
	activityService *okractivities.Service,
	attachmentService *attachments.Service,
	cacheService *cache.Service,
	log *logger.Logger,
) *Handlers {
	return &Handlers{
		objectives: objectiveService, keyResults: keyResultService,
		okrActivities: activityService, attachments: attachmentService,
		cache: cacheService, log: log,
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
		var err error
		if strings.Contains(key, "*") {
			err = h.cache.DeleteByPattern(ctx, key)
		} else {
			err = h.cache.Delete(ctx, key)
		}
		if err != nil && h.log != nil {
			h.log.Error(ctx, "failed to invalidate objective cache",
				"key", key, "workspace_id", workspaceID, "objective_id", objectiveID, "error", err)
		}
	}
}

func respondObjectiveError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, objectives.ErrInvalid), errors.Is(err, objectives.ErrInvalidReference):
		status = http.StatusBadRequest
	case errors.Is(err, objectives.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, objectives.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, objectives.ErrNameExists), errors.Is(err, objectives.ErrVersionConflict):
		status = http.StatusConflict
	}
	return web.RespondError(ctx, w, err, status)
}

func parseObjectivePathID(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := web.UUIDPathParameter(r, name)
	if err != nil {
		_ = web.RespondError(ctx, w, err, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
