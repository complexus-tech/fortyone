package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeSlackEvent = "slack:event:process"

type SlackEventPayload struct {
	Provider            string    `json:"provider,omitempty"`
	InboxID             uuid.UUID `json:"inboxId,omitempty"`
	ExternalWorkspaceID string    `json:"externalWorkspaceId,omitempty"`
	EventID             string    `json:"eventId,omitempty"`
	RecoveryAttempt     int       `json:"recoveryAttempt,omitempty"`
}

func (s *Service) EnqueueSlackEvent(ctx context.Context, payload SlackEventPayload) error {
	if s == nil || s.asynqClient == nil {
		return errors.New("tasks: slack event queue is not configured")
	}
	payload.Provider = strings.TrimSpace(payload.Provider)
	if payload.Provider != "slack" {
		return errors.New("tasks: Slack webhook provider is required")
	}
	if payload.InboxID == uuid.Nil {
		return errors.New("tasks: Slack webhook inbox id is required")
	}
	if strings.TrimSpace(payload.ExternalWorkspaceID) != "" || strings.TrimSpace(payload.EventID) != "" || payload.RecoveryAttempt != 0 {
		return errors.New("tasks: new Slack webhook tasks must contain inbox identity only")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeSlackEvent, err)
	}
	task := asynq.NewTask(TypeSlackEvent, encoded)
	_, err = s.asynqClient.Enqueue(task,
		asynq.Queue("integrations"),
		asynq.MaxRetry(6),
		asynq.Timeout(45*time.Second),
		asynq.Retention(24*time.Hour),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("tasks: enqueue %s task: %w", TypeSlackEvent, err)
	}
	if s.log != nil {
		s.log.Info(ctx, "slack event enqueued", "provider", payload.Provider, "inbox_id", payload.InboxID)
	}
	return nil
}
