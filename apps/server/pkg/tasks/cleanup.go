package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// Cleanup task types
const (
	TypeTokenCleanup        = "cleanup:tokens"
	TypeDeleteStories       = "cleanup:deleted_stories"
	TypeWebhookCleanup      = "cleanup:stripe_webhooks"
	TypeWorkspaceCleanup    = "cleanup:deleted_workspaces"
	TypeChatSessionsCleanup = "cleanup:deleted_chat_sessions"
)

// EnqueueDeleteStories enqueues a task to cleanup deleted stories.
func (s *Service) EnqueueDeleteStories(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue DeleteStories task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeDeleteStories, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue DeleteStories task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued DeleteStories task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueTokenCleanup enqueues a task to cleanup expired verification tokens.
func (s *Service) EnqueueTokenCleanup(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue TokenCleanup task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 5),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeTokenCleanup, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue TokenCleanup task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued TokenCleanup task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueWebhookCleanup enqueues a task to cleanup old stripe webhook events.
func (s *Service) EnqueueWebhookCleanup(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue WebhookCleanup task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 10),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeWebhookCleanup, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue WebhookCleanup task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued WebhookCleanup task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueWorkspaceCleanup enqueues a task to cleanup deleted workspaces.
func (s *Service) EnqueueWorkspaceCleanup(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue WorkspaceCleanup task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeWorkspaceCleanup, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue WorkspaceCleanup task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued WorkspaceCleanup task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueChatSessionsCleanup enqueues a task to cleanup deleted chat sessions.
func (s *Service) EnqueueChatSessionsCleanup(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue ChatSessionsCleanup task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeChatSessionsCleanup, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue ChatSessionsCleanup task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued ChatSessionsCleanup task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}
