package taskhandlers

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

// runScheduledTask provides one fail-fast logging and error boundary for
// scheduled handlers. Domain-specific handler groups keep ownership of their
// dependencies and operation names while sharing the transport ceremony.
func runScheduledTask(
	ctx context.Context,
	task *asynq.Task,
	log *logger.Logger,
	handlerName string,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if ctx == nil {
		return fmt.Errorf("%s handler context is required", handlerName)
	}
	if log == nil || operation == nil {
		return fmt.Errorf("%s handler dependencies are required", handlerName)
	}

	taskID := scheduledTaskID(task)
	log.Info(ctx, "HANDLER: Processing "+taskName+" task", "task_id", taskID)
	if err := operation(ctx); err != nil {
		log.Error(ctx, "Failed to process "+taskName+" task", "error", err, "task_id", taskID)
		return fmt.Errorf("%s failed: %w", operationName, err)
	}

	log.Info(ctx, "HANDLER: Successfully processed "+taskName+" task", "task_id", taskID)
	return nil
}

func scheduledTaskID(task *asynq.Task) string {
	if task == nil || task.ResultWriter() == nil {
		return ""
	}
	return task.ResultWriter().TaskID()
}
