package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

const (
	TypeMayaWorkFocusInference = "maya:work_focus:infer"
	TypeMayaBatchAssignment    = "maya:assignment:batch"
)

func (s *Service) EnqueueMayaWorkFocusInference(opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue MayaWorkFocusInference task")

	defaultOpts := []asynq.Option{
		asynq.Queue("automation"),
		asynq.TaskID("maya_work_focus_inference"),
		asynq.MaxRetry(3),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeMayaWorkFocusInference, nil, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue MayaWorkFocusInference task", "error", err)
		return nil, err
	}

	s.log.Info(ctx, "Successfully enqueued MayaWorkFocusInference task", "task_id", info.ID, "queue", info.Queue)
	return info, nil
}
