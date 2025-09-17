package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// Lifecycle management task types
const (
	// Notification tasks
	TypeWorkspaceInactivityWarning = "lifecycle:workspace:inactivity_warning"
	TypeUserInactivityWarning      = "lifecycle:user:inactivity_warning"

	// Deletion tasks
	TypeWorkspaceDeletion = "lifecycle:workspace:deletion"
	TypeUserDeactivation  = "lifecycle:user:deactivation"
)

// EnqueueWorkspaceInactivityWarning enqueues a task to send inactivity warnings to workspace admins.
func (s *Service) EnqueueWorkspaceInactivityWarning(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue WorkspaceInactivityWarning task")

	defaultOpts := []asynq.Option{
		asynq.Queue("notifications"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 5),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeWorkspaceInactivityWarning, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue WorkspaceInactivityWarning task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued WorkspaceInactivityWarning task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueUserInactivityWarning enqueues a task to send inactivity warnings to users.
func (s *Service) EnqueueUserInactivityWarning(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue UserInactivityWarning task")

	defaultOpts := []asynq.Option{
		asynq.Queue("notifications"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 5),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeUserInactivityWarning, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue UserInactivityWarning task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued UserInactivityWarning task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueWorkspaceDeletion enqueues a task to delete inactive workspaces.
func (s *Service) EnqueueWorkspaceDeletion(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue WorkspaceDeletion task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 10),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeWorkspaceDeletion, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue WorkspaceDeletion task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued WorkspaceDeletion task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueUserDeactivation enqueues a task to deactivate inactive users.
func (s *Service) EnqueueUserDeactivation(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue UserDeactivation task")

	defaultOpts := []asynq.Option{
		asynq.Queue("cleanup"),
		asynq.MaxRetry(3),
		asynq.ProcessIn(time.Minute * 10),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeUserDeactivation, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue UserDeactivation task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued UserDeactivation task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}
