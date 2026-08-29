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
	TypeFigmaWebhook         = "figma:webhook:process"
	TypeFigmaWebhookRecovery = "figma:webhook:recovery"
)

// FigmaWebhookPayload contains only the durable inbox identity. Provider
// payloads and passcodes remain encrypted in PostgreSQL and never enter Redis.
type FigmaWebhookPayload struct {
	InboxID uuid.UUID `json:"inboxId"`
}

func (s *Service) EnqueueFigmaWebhook(ctx context.Context, payload FigmaWebhookPayload) error {
	if s == nil || s.asynqClient == nil {
		return errors.New("tasks: Figma webhook queue is not configured")
	}
	if payload.InboxID == uuid.Nil {
		return errors.New("tasks: Figma webhook inbox identity is invalid")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeFigmaWebhook, err)
	}
	_, err = s.asynqClient.Enqueue(
		asynq.NewTask(TypeFigmaWebhook, encoded),
		asynq.Queue("integrations"),
		asynq.MaxRetry(8),
		asynq.Timeout(90*time.Second),
		asynq.Retention(24*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("tasks: enqueue %s task: %w", TypeFigmaWebhook, err)
	}
	if s.log != nil {
		s.log.Info(ctx, "Figma webhook enqueued", "inbox_id", payload.InboxID)
	}
	return nil
}
