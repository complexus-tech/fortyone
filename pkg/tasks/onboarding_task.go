package tasks

import (
	"context" // You added this in your manual edit, keeping it for logger compatibility
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	// Logger is part of the Service struct, no direct import needed here unless for standalone functions
)

// --- User Onboarding Task ---
const TypeUserOnboardingStart = "user:onboarding:start"

// UserOnboardingStartPayload defines the data for starting a user's onboarding process.
// Matches the fields you updated: UserID, Email, FullName.
type UserOnboardingStartPayload struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

// EnqueueUserOnboardingStart enqueues a task to initiate the user onboarding process.
// This is a method on the tasks.Service struct.
func (s *Service) EnqueueUserOnboardingStart(payload UserOnboardingStartPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// Use a context for logging, as you did in your manual edit.
	// If this method could take a context from the caller (e.g., from an HTTP request),
	// that would be even better for tracing and cancellation.
	ctx := context.Background() // Or pass ctx as an argument to this method

	// s.log is available from the Service struct (s *Service)
	s.log.Info(ctx, "Attempting to enqueue UserOnboardingStart task", "user_id", payload.UserID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal UserOnboardingStartPayload", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeUserOnboardingStart, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"), // Default queue for this task type
		// asynq.MaxRetry(3),      // Example: set default retry
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
