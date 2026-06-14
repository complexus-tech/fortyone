package mayahttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/internal/platform/billing"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const defaultPlanningWindow = 14 * 24 * time.Hour

var ErrMayaAccessRequired = errors.New("maya agent is available on paid plans and active trials")

type Handlers struct {
	db      *sqlx.DB
	log     *logger.Logger
	service *maya.Service
	now     func() time.Time
}

func New(db *sqlx.DB, log *logger.Logger, service *maya.Service) *Handlers {
	return &Handlers{
		db:      db,
		log:     log,
		service: service,
		now:     time.Now,
	}
}

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

func (h *Handlers) statusCode(err error) int {
	switch {
	case errors.Is(err, maya.ErrNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, maya.ErrInvalidPlanInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handlers) workspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM workspaces w
			WHERE w.workspace_id = $1
				AND ` + billing.WorkspaceMayaAccessSQL("w") + `
		)
	`

	var allowed bool
	if err := h.db.GetContext(ctx, &allowed, query, workspaceID); err != nil {
		return false, fmt.Errorf("check maya access: %w", err)
	}
	return allowed, nil
}
