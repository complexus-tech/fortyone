package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

var (
	BrevoOnboardingList = int64(6)
	BrevoTrialList      = int64(12)
)

// HandleUserOnboardingStart processes the user onboarding start task.
func (h *handlers) HandleUserOnboardingStart(ctx context.Context, t *asynq.Task) error {
	var p tasks.UserOnboardingStartPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal UserOnboardingStartPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing UserOnboardingStart task",
		"user_id", p.UserID,
		"email", p.Email,
		"full_name", p.FullName,
		"task_id", t.ResultWriter().TaskID(),
	)

	// Brevo integration
	_, err := h.brevoService.CreateOrUpdateContact(ctx, brevo.CreateOrUpdateContactRequest{
		Email: p.Email,
		Attributes: map[string]any{
			"NAME": p.FullName,
		},
		ListIDs: []int64{BrevoOnboardingList},
	})
	if err != nil {
		return fmt.Errorf("failed to create/update contact: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed UserOnboardingStart task", "user_id", p.UserID)
	return nil
}

// HandleWorkspaceTrialStart processes the workspace trial start task.
func (h *handlers) HandleWorkspaceTrialStart(ctx context.Context, t *asynq.Task) error {
	var p tasks.WorkspaceTrialStartPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal WorkspaceTrialStartPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing WorkspaceTrialStart task",
		"user_id", p.UserID,
		"email", p.Email,
		"full_name", p.FullName,
		"workspace_slug", p.WorkspaceSlug,
		"workspace_name", p.WorkspaceName,
		"task_id", t.ResultWriter().TaskID(),
	)

	// Brevo integration - add to trial list
	_, err := h.brevoService.CreateOrUpdateContact(ctx, brevo.CreateOrUpdateContactRequest{
		Email: p.Email,
		Attributes: map[string]any{
			"NAME":           p.FullName,
			"WORKSPACE_NAME": p.WorkspaceName,
			"WORKSPACE_SLUG": p.WorkspaceSlug,
		},
		ListIDs: []int64{BrevoTrialList},
	})
	if err != nil {
		return fmt.Errorf("failed to create/update contact for trial: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed WorkspaceTrialStart task", "user_id", p.UserID, "workspace_slug", p.WorkspaceSlug)
	return nil
}

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
