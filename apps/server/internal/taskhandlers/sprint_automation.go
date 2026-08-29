package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

// SprintAutomationHandlerDependencies contains the single typed persistence
// capability shared by sprint creation and inactivity shutdown policies.
type SprintAutomationHandlerDependencies struct {
	Log   *logger.Logger
	Store jobs.SprintAutomationStore
}

// SprintAutomationHandlers owns both scheduled phases of sprint automation.
type SprintAutomationHandlers struct {
	log   *logger.Logger
	store jobs.SprintAutomationStore
}

// NewSprintAutomationHandlers creates the sprint automation handler group.
func NewSprintAutomationHandlers(dependencies SprintAutomationHandlerDependencies) *SprintAutomationHandlers {
	return &SprintAutomationHandlers{log: dependencies.Log, store: dependencies.Store}
}

// HandleSprintAutoCreation reconciles and replenishes automated sprints.
func (h *SprintAutomationHandlers) HandleSprintAutoCreation(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "SprintAutoCreation", "sprint auto-creation", func(ctx context.Context) error {
		return jobs.ProcessSprintAutoCreation(ctx, h.store, h.log)
	})
}

// HandleDisableInactiveAutomation disables stale automated sprint policies.
func (h *SprintAutomationHandlers) HandleDisableInactiveAutomation(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(
		ctx,
		task,
		"DisableInactiveAutomation",
		"disable inactive automation",
		func(ctx context.Context) error {
			return jobs.DisableAutomationForInactiveTeams(ctx, h.store, h.log)
		},
	)
}

func (h *SprintAutomationHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "sprint automation", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "sprint automation", taskName, operationName, operation)
}
