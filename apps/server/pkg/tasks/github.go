package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeGitHubStorySync       = "github:story:sync"
	TypeGitHubWebhook         = "github:webhook:process"
	TypeGitHubWebhookRecovery = "github:webhook:recovery"
)

type GitHubWebhookPayload struct {
	InboxID uuid.UUID `json:"inboxId"`
}

func (s *Service) EnqueueGitHubWebhook(ctx context.Context, payload GitHubWebhookPayload) error {
	if s == nil || s.asynqClient == nil {
		return errors.New("tasks: GitHub webhook queue is not configured")
	}
	if payload.InboxID == uuid.Nil {
		return errors.New("tasks: GitHub webhook inbox identity is invalid")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeGitHubWebhook, err)
	}
	task := asynq.NewTask(TypeGitHubWebhook, encoded)
	// The durable inbox owns deduplication. A deterministic Asynq task ID would
	// let a retained or archived task suppress a later recovery dispatch.
	_, err = s.asynqClient.Enqueue(task,
		asynq.Queue("integrations"),
		asynq.MaxRetry(8),
		asynq.Timeout(90*time.Second),
		asynq.Retention(24*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("tasks: enqueue %s task: %w", TypeGitHubWebhook, err)
	}
	if s.log != nil {
		s.log.Info(ctx, "GitHub webhook enqueued", "inbox_id", payload.InboxID)
	}
	return nil
}

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
