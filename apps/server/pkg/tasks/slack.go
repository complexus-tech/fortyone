package tasks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TypeSlackEvent = "slack:event:process"

type SlackEventPayload struct {
	ExternalWorkspaceID string `json:"externalWorkspaceId"`
	EventID             string `json:"eventId"`
	RecoveryAttempt     int    `json:"recoveryAttempt,omitempty"`
}

func (s *Service) EnqueueSlackEvent(ctx context.Context, payload SlackEventPayload) error {
	if s == nil || s.asynqClient == nil {
		return errors.New("tasks: slack event queue is not configured")
	}
	payload.EventID = strings.TrimSpace(payload.EventID)
	payload.ExternalWorkspaceID = strings.TrimSpace(payload.ExternalWorkspaceID)
	if payload.ExternalWorkspaceID == "" {
		return errors.New("tasks: Slack external workspace id is required")
	}
	if payload.EventID == "" {
		return errors.New("tasks: slack event id is required")
	}
	if payload.RecoveryAttempt < 0 {
		return errors.New("tasks: Slack event recovery attempt cannot be negative")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeSlackEvent, err)
	}
	taskID := slackEventTaskID(payload.ExternalWorkspaceID, payload.EventID, payload.RecoveryAttempt)
	task := asynq.NewTask(TypeSlackEvent, encoded)
	_, err = s.asynqClient.Enqueue(task,
		asynq.Queue("integrations"),
		asynq.TaskID(taskID),
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
		s.log.Info(ctx, "slack event enqueued", "event_id", payload.EventID)
	}
	return nil
}

func slackEventTaskID(externalWorkspaceID, eventID string, recoveryAttempt int) string {
	seed := strings.TrimSpace(externalWorkspaceID) + ":" + strings.TrimSpace(eventID)
	if recoveryAttempt > 0 {
		seed += fmt.Sprintf(":recovery:%d", recoveryAttempt)
	}
	digest := sha256.Sum256([]byte(seed))
	return "slack-event-" + hex.EncodeToString(digest[:])
}
