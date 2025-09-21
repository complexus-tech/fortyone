package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeUserOnboardingStart = "user:onboarding:start"
const TypeWorkspaceTrialStart = "workspace:trial:start"

type UserOnboardingStartPayload struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

type WorkspaceTrialStartPayload struct {
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	FullName      string `json:"fullName"`
	WorkspaceSlug string `json:"workspaceSlug"`
	WorkspaceName string `json:"workspaceName"`
}

// EnqueueUserOnboardingStart enqueues a task to initiate the user onboarding process.
func (s *Service) EnqueueUserOnboardingStart(payload UserOnboardingStartPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue UserOnboardingStart task", "user_id", payload.UserID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal UserOnboardingStartPayload", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeUserOnboardingStart, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"),
		asynq.MaxRetry(2),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeUserOnboardingStart, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue UserOnboardingStart task", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeUserOnboardingStart, err)
	}

	s.log.Info(ctx, "Successfully enqueued UserOnboardingStart task", "task_id", info.ID, "queue", info.Queue, "user_id", payload.UserID)
	return info, nil
}

// EnqueueWorkspaceTrialStart enqueues a task to initiate the workspace trial process.
func (s *Service) EnqueueWorkspaceTrialStart(payload WorkspaceTrialStartPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue WorkspaceTrialStart task", "user_id", payload.UserID, "workspace_slug", payload.WorkspaceSlug)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal WorkspaceTrialStartPayload", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeWorkspaceTrialStart, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"),
		asynq.MaxRetry(2),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeWorkspaceTrialStart, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue WorkspaceTrialStart task", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeWorkspaceTrialStart, err)
	}

	s.log.Info(ctx, "Successfully enqueued WorkspaceTrialStart task", "task_id", info.ID, "queue", info.Queue, "user_id", payload.UserID)
	return info, nil
}
