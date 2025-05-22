package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	asynqClient *asynq.Client
	log         *logger.Logger
}

// New creates a new tasks Service instance.
func New(existingRdb redis.UniversalClient, log *logger.Logger) (*Service, error) {
	client := asynq.NewClientFromRedisClient(existingRdb)

	if client == nil { // Should not happen if existingRdb is valid, but good to check.
		return nil, fmt.Errorf("tasks: failed to create Asynq client using NewClientFromRedisClient (returned nil)")
	}
	s := &Service{
		asynqClient: client,
		log:         log,
	}
	return s, nil
}

func (s *Service) Close() error {
	if s.asynqClient != nil {
		return s.asynqClient.Close()
	}
	return nil
}

// --- User Onboarding Task ---
const TypeUserOnboardingStart = "user:onboarding:start"

// UserOnboardingStartPayload defines the data for starting a user's onboarding process.
type UserOnboardingStartPayload struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

// EnqueueUserOnboardingStart enqueues a task to initiate the user onboarding process.
func (s *Service) EnqueueUserOnboardingStart(payload UserOnboardingStartPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal UserOnboardingStartPayload", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeUserOnboardingStart, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"),
		// asynq.MaxRetry(3), // Example: set default retry for this task type
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeUserOnboardingStart, payloadBytes, finalOpts...)

	// Pass a context if you have one available, otherwise nil is acceptable for Enqueue
	// but EnqueueContext is preferred if a context (e.g. from an HTTP request) is available.
	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue UserOnboardingStart task", "error", err, "user_id", payload.UserID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeUserOnboardingStart, err)
	}
	return info, nil
}
