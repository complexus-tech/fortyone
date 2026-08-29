package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// StoryAutomationHandlerDependencies contains the typed persistence capability
// and durable actor shared by the three scheduled story automation policies.
type StoryAutomationHandlerDependencies struct {
	Log          *logger.Logger
	Store        jobs.StoryAutomationStore
	SystemUserID uuid.UUID
}

// StoryAutomationHandlers owns archive, close, and sprint-migration scheduling
// without exposing a database driver to the task layer.
type StoryAutomationHandlers struct {
	log          *logger.Logger
	store        jobs.StoryAutomationStore
	systemUserID uuid.UUID
}

// NewStoryAutomationHandlers creates the story automation handler group.
func NewStoryAutomationHandlers(dependencies StoryAutomationHandlerDependencies) *StoryAutomationHandlers {
	return &StoryAutomationHandlers{
		log:          dependencies.Log,
		store:        dependencies.Store,
		systemUserID: dependencies.SystemUserID,
	}
}

// HandleStoryAutoArchive archives completed or cancelled stories after the
// owning team's configured retention period.
func (h *StoryAutomationHandlers) HandleStoryAutoArchive(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "StoryAutoArchive", "story auto-archive", func(ctx context.Context) error {
		return jobs.ProcessStoryAutoArchive(ctx, h.store, h.log)
	})
}

// HandleStoryAutoClose closes inactive stories and records the system actor's
// activity in the same transaction.
func (h *StoryAutomationHandlers) HandleStoryAutoClose(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "StoryAutoClose", "story auto-close", func(ctx context.Context) error {
		return jobs.ProcessStoryAutoClose(ctx, h.store, h.log, h.systemUserID)
	})
}

// HandleSprintStoryMigration moves incomplete stories to the next sprint and
// records activity plus audit evidence atomically.
func (h *StoryAutomationHandlers) HandleSprintStoryMigration(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "SprintStoryMigration", "sprint story migration", func(ctx context.Context) error {
		return jobs.ProcessSprintStoryMigration(ctx, h.store, h.log, h.systemUserID)
	})
}

func (h *StoryAutomationHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "story automation", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "story automation", taskName, operationName, operation)
}
