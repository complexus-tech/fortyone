package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeGitHubStorySync = "github:story:sync"

type GitHubStorySyncPayload struct {
	StoryID     uuid.UUID `json:"storyId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
}

func (s *Service) EnqueueGitHubStorySync(payload GitHubStorySyncPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue GitHubStorySync task", "story_id", payload.StoryID, "workspace_id", payload.WorkspaceID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal GitHubStorySyncPayload", "error", err, "story_id", payload.StoryID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeGitHubStorySync, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("integrations"),
		asynq.MaxRetry(10),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeGitHubStorySync, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue GitHubStorySync task", "error", err, "story_id", payload.StoryID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeGitHubStorySync, err)
	}

	s.log.Info(ctx, "Successfully enqueued GitHubStorySync task", "task_id", info.ID, "queue", info.Queue, "story_id", payload.StoryID)
	return info, nil
}
