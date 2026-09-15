package workerbootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	accountDeletionFinalizationTask = "cleanup:account_deletion_finalization"
	accountDeletionBatchSize        = 50
)

type accountDeletionFinalizer interface {
	FinalizePending(context.Context, int) (int, error)
}

type accountSubscriberCleanup interface {
	Dispatch(context.Context) (int, error)
	UpdateActiveAccount(context.Context, usersdomain.SubscriberUpdate) error
}

func subscriberUpdateHandler(cleanup accountSubscriberCleanup) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		if cleanup == nil {
			return errors.New("subscriber lifecycle cleanup is not configured")
		}
		// This shape includes the common email and the additional identity/list
		// context used by onboarding and trial-start jobs. Queued names are ignored.
		var payload tasks.WorkspaceTrialStartPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil || payload.Email == "" {
			return fmt.Errorf("invalid subscriber update payload: %w", asynq.SkipRetry)
		}
		update := usersdomain.SubscriberUpdate{Email: payload.Email}
		switch task.Type() {
		case tasks.TypeUserOnboardingStart, tasks.TypeWorkspaceTrialStart:
			userID, err := uuid.Parse(payload.UserID)
			if err != nil || userID == uuid.Nil {
				return fmt.Errorf("invalid subscriber account identity: %w", asynq.SkipRetry)
			}
			update.UserID = userID
			if task.Type() == tasks.TypeUserOnboardingStart {
				update.ListIDs = []int64{taskhandlers.BrevoOnboardingList}
			} else {
				update.ListIDs = []int64{taskhandlers.BrevoTrialList}
				update.Attributes = map[string]string{"WORKSPACE_NAME": payload.WorkspaceName, "WORKSPACE_SLUG": payload.WorkspaceSlug}
			}
		}
		return cleanup.UpdateActiveAccount(ctx, update)
	}
}

// Calendar cleanup has its own durable dispatcher. This scheduled sweep finishes
// pending account deletions once that cleanup is complete, including after a
// worker restart. The manager owns locking and safe retries across workers.
func registerAccountDeletionFinalization(mux *asynq.ServeMux, scheduler scheduleRegistrar, manager accountDeletionFinalizer, subscribers accountSubscriberCleanup, log *logger.Logger) error {
	if mux == nil || scheduler == nil || manager == nil || subscribers == nil || log == nil {
		return errors.New("account deletion worker dependencies are required")
	}
	mux.HandleFunc(accountDeletionFinalizationTask, func(ctx context.Context, _ *asynq.Task) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		finalized, finalizeErr := manager.FinalizePending(ctx, accountDeletionBatchSize)
		if finalized > 0 {
			log.Info(ctx, "Pending account deletions finalized", "count", finalized)
		}
		cleaned, cleanupErr := subscribers.Dispatch(ctx)
		if cleaned > 0 {
			log.Info(ctx, "Account subscriber cleanup completed", "count", cleaned)
		}
		return errors.Join(finalizeErr, cleanupErr)
	})
	_, err := scheduler.Register("*/1 * * * *", asynq.NewTask(accountDeletionFinalizationTask, nil),
		asynq.Queue("cleanup"), asynq.MaxRetry(2), asynq.Timeout(4*time.Minute), asynq.Unique(55*time.Second))
	if err != nil {
		return fmt.Errorf("register account deletion finalization: %w", err)
	}
	return nil
}
