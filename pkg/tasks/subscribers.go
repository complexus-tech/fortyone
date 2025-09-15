package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeSubscriberUpdate = "email:subscription:user"
const TypeSubscriberDelete = "email:subscription:user:delete"

type SubscriberUpdatePayload struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

type SubscriberDeletePayload struct {
	Email string `json:"email"`
}

// EnqueueSubscriberUpdate enqueues a task to update a subscriber.
func (s *Service) EnqueueSubscriberUpdate(payload SubscriberUpdatePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue SubscriberUpdate task", "email", payload.Email)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal SubscriberUpdatePayload", "error", err, "email", payload.Email)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeSubscriberUpdate, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"),
		asynq.MaxRetry(2),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeSubscriberUpdate, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue SubscriberUpdate task", "error", err, "email", payload.Email)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeSubscriberUpdate, err)
	}

	s.log.Info(ctx, "Successfully enqueued SubscriberUpdate task", "task_id", info.ID, "queue", info.Queue, "email", payload.Email)
	return info, nil
}

// EnqueueSubscriberDelete enqueues a task to delete a subscriber.
func (s *Service) EnqueueSubscriberDelete(payload SubscriberDeletePayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue SubscriberDelete task", "email", payload.Email)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal SubscriberDeletePayload", "error", err, "email", payload.Email)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeSubscriberDelete, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("onboarding"),
		asynq.MaxRetry(2),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeSubscriberDelete, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue SubscriberDelete task", "error", err, "email", payload.Email)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeSubscriberDelete, err)
	}

	s.log.Info(ctx, "Successfully enqueued SubscriberDelete task", "task_id", info.ID, "queue", info.Queue, "email", payload.Email)
	return info, nil
}
