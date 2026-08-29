package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/hibiken/asynq"
)

// UserLifecycleStore combines the warning and deactivation phases of the
// account inactivity state machine behind the users module boundary.
type UserLifecycleStore interface {
	jobs.UserInactivityWarningStore
	jobs.InactiveUserDeactivator
}

type UserLifecycleHandlerDependencies struct {
	Log    *logger.Logger
	Store  UserLifecycleStore
	Mailer mailer.Service
}

// UserLifecycleHandlers owns account inactivity warnings and the later
// deactivation transition. Keeping both phases together makes the policy easy
// to trace without mixing it into generic data-retention cleanup.
type UserLifecycleHandlers struct {
	log    *logger.Logger
	store  UserLifecycleStore
	mailer mailer.Service
}

func NewUserLifecycleHandlers(dependencies UserLifecycleHandlerDependencies) *UserLifecycleHandlers {
	return &UserLifecycleHandlers{
		log:    dependencies.Log,
		store:  dependencies.Store,
		mailer: dependencies.Mailer,
	}
}

func (h *UserLifecycleHandlers) HandleUserInactivityWarning(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(ctx, task, "UserInactivityWarning", "user inactivity warning", func(ctx context.Context) error {
		return jobs.ProcessUserInactivityWarning(ctx, h.store, h.log, h.mailer)
	})
}

func (h *UserLifecycleHandlers) HandleUserDeactivation(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(ctx, task, "UserDeactivation", "user deactivation", func(ctx context.Context) error {
		return jobs.ProcessUserDeactivation(ctx, h.store, h.log)
	})
}

func (h *UserLifecycleHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "user lifecycle", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "user lifecycle", taskName, operationName, operation)
}
