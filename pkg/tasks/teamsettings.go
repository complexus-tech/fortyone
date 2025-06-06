package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

// Team settings automation task types
const (
	TypeSprintAutoCreation = "automation:sprints:create"
	TypeStoryAutoArchive   = "automation:stories:archive"
	TypeStoryAutoClose     = "automation:stories:close"
)

// EnqueueSprintAutoCreation enqueues a task to auto-create sprints for teams.
func (s *Service) EnqueueSprintAutoCreation(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue SprintAutoCreation task")

	defaultOpts := []asynq.Option{
		asynq.Queue("automation"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeSprintAutoCreation, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue SprintAutoCreation task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued SprintAutoCreation task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueStoryAutoArchive enqueues a task to auto-archive completed stories.
func (s *Service) EnqueueStoryAutoArchive(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue StoryAutoArchive task")

	defaultOpts := []asynq.Option{
		asynq.Queue("automation"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeStoryAutoArchive, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue StoryAutoArchive task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued StoryAutoArchive task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}

// EnqueueStoryAutoClose enqueues a task to auto-close inactive stories.
func (s *Service) EnqueueStoryAutoClose(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue StoryAutoClose task")

	defaultOpts := []asynq.Option{
		asynq.Queue("automation"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeStoryAutoClose, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue StoryAutoClose task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued StoryAutoClose task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}
