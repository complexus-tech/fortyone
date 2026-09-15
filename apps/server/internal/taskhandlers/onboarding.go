package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

var (
	BrevoOnboardingList = int64(6)
	BrevoTrialList      = int64(12)
)

// HandleWorkspaceTrialEnd processes the workspace trial end task.
func (h *handlers) HandleWorkspaceTrialEnd(ctx context.Context, t *asynq.Task) error {
	var p tasks.WorkspaceTrialEndPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal WorkspaceTrialEndPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing WorkspaceTrialEnd task",
		"email", p.Email,
		"task_id", t.ResultWriter().TaskID(),
	)

	// Remove from trial list using Brevo's contact removal
	err := h.brevoService.RemoveContactFromList(ctx, BrevoTrialList, p.Email)
	if err != nil {
		return fmt.Errorf("failed to remove contact from trial list: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed WorkspaceTrialEnd task", "email", p.Email)
	return nil
}
