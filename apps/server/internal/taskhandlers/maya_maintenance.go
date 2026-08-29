package taskhandlers

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

// MayaMaintenanceHandlerDependencies contains only the dependencies required
// by Maya's scheduled learning policies.
type MayaMaintenanceHandlerDependencies struct {
	Log   *logger.Logger
	Store jobs.MayaWorkFocusStore
}

// MayaMaintenanceHandlers owns Maya's scheduled maintenance boundary.
type MayaMaintenanceHandlers struct {
	log   *logger.Logger
	store jobs.MayaWorkFocusStore
}

// NewMayaMaintenanceHandlers creates the Maya maintenance handler group.
func NewMayaMaintenanceHandlers(dependencies MayaMaintenanceHandlerDependencies) *MayaMaintenanceHandlers {
	return &MayaMaintenanceHandlers{
		log:   dependencies.Log,
		store: dependencies.Store,
	}
}

// HandleWorkFocusInference applies Maya's bounded work-focus learning policy.
func (h *MayaMaintenanceHandlers) HandleWorkFocusInference(ctx context.Context, task *asynq.Task) error {
	if h == nil {
		return runScheduledTask(
			ctx,
			task,
			nil,
			"Maya work-focus inference",
			"MayaWorkFocusInference",
			"maya work-focus inference",
			nil,
		)
	}
	return runScheduledTask(
		ctx,
		task,
		h.log,
		"MayaWorkFocusInference",
		"MayaWorkFocusInference",
		"maya work-focus inference",
		func(ctx context.Context) error {
			return jobs.ProcessMayaWorkFocusInference(ctx, h.store, h.log, time.Now().UTC())
		},
	)
}
