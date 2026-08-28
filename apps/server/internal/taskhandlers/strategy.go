package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type StrategyHandlerDependencies struct {
	Log          *logger.Logger
	Store        jobs.StrategyCommunicationsStore
	Notifier     jobs.StrategyNotificationCreator
	SystemUserID uuid.UUID
}

// StrategyHandlers owns scheduled planning reminders, weekly check-ins, and
// monthly summaries. Its persistence and notification ports are explicit so
// the worker cannot reach around either module with ad hoc SQL.
type StrategyHandlers struct {
	log          *logger.Logger
	store        jobs.StrategyCommunicationsStore
	notifier     jobs.StrategyNotificationCreator
	systemUserID uuid.UUID
}

func NewStrategyHandlers(dependencies StrategyHandlerDependencies) *StrategyHandlers {
	return &StrategyHandlers{
		log:          dependencies.Log,
		store:        dependencies.Store,
		notifier:     dependencies.Notifier,
		systemUserID: dependencies.SystemUserID,
	}
}

func (h *StrategyHandlers) HandleStrategyCommunications(
	ctx context.Context,
	task *asynq.Task,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "strategy", "StrategyCommunications", "strategy communications", nil)
	}
	return runScheduledTask(
		ctx,
		task,
		h.log,
		"strategy",
		"StrategyCommunications",
		"strategy communications",
		func(ctx context.Context) error {
			return jobs.ProcessStrategyCommunications(
				ctx,
				h.store,
				h.log,
				h.notifier,
				h.systemUserID,
			)
		},
	)
}
