package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/hibiken/asynq"
)

// WorkspaceLifecycleStore combines the two workspace-owned persistence ports
// needed by the warning and deletion phases of the inactivity policy.
type WorkspaceLifecycleStore interface {
	jobs.WorkspaceInactivityWarningStore
	jobs.InactiveWorkspaceDeleter
	jobs.SoftDeletedWorkspacePurger
}

type WorkspaceLifecycleHandlerDependencies struct {
	Log    *logger.Logger
	Store  WorkspaceLifecycleStore
	Mailer mailer.Service
}

// WorkspaceLifecycleHandlers owns the warning and deletion phases of the
// workspace inactivity state machine.
type WorkspaceLifecycleHandlers struct {
	log    *logger.Logger
	store  WorkspaceLifecycleStore
	mailer mailer.Service
}

func NewWorkspaceLifecycleHandlers(
	dependencies WorkspaceLifecycleHandlerDependencies,
) *WorkspaceLifecycleHandlers {
	return &WorkspaceLifecycleHandlers{
		log:    dependencies.Log,
		store:  dependencies.Store,
		mailer: dependencies.Mailer,
	}
}

func (h *WorkspaceLifecycleHandlers) HandleWorkspaceInactivityWarning(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(ctx, task, "WorkspaceInactivityWarning", "workspace inactivity warning", func(ctx context.Context) error {
		return jobs.ProcessWorkspaceInactivityWarning(ctx, h.store, h.log, h.mailer)
	})
}

func (h *WorkspaceLifecycleHandlers) HandleWorkspaceDeletion(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(ctx, task, "WorkspaceDeletion", "workspace deletion", func(ctx context.Context) error {
		return jobs.ProcessWorkspaceDeletion(ctx, h.store, h.log)
	})
}

func (h *WorkspaceLifecycleHandlers) HandleWorkspaceCleanup(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(ctx, task, "WorkspaceCleanup", "workspace cleanup", func(ctx context.Context) error {
		return jobs.PurgeDeletedWorkspaces(ctx, h.store, h.log)
	})
}

func (h *WorkspaceLifecycleHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "workspace lifecycle", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "workspace lifecycle", taskName, operationName, operation)
}
