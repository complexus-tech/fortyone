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

const TypeBrevoEmailReply = "brevo:email_reply:process"

// BrevoEmailReplyPayload identifies a durable messaging inbox entry. The
// external workspace ID is the inbox's opaque workspace-and-thread scope, not
// a bare workspace UUID. Email content remains encrypted in Postgres and is
// never copied into Redis.
type BrevoEmailReplyPayload struct {
	ExternalWorkspaceID string `json:"externalWorkspaceId"`
	EventID             string `json:"eventId"`
	RecoveryAttempt     int    `json:"recoveryAttempt,omitempty"`
}

func (s *Service) EnqueueBrevoEmailReply(ctx context.Context, payload BrevoEmailReplyPayload) error {
	if s == nil || s.asynqClient == nil {
		return errors.New("tasks: Brevo email reply queue is not configured")
	}
	payload.ExternalWorkspaceID = strings.TrimSpace(payload.ExternalWorkspaceID)
	payload.EventID = strings.TrimSpace(payload.EventID)
	if payload.ExternalWorkspaceID == "" {
		return errors.New("tasks: Brevo email reply external workspace id is required")
	}
	if payload.EventID == "" {
		return errors.New("tasks: Brevo email reply event id is required")
	}
	if payload.RecoveryAttempt < 0 {
		return errors.New("tasks: Brevo email reply recovery attempt cannot be negative")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeBrevoEmailReply, err)
	}
	task := asynq.NewTask(TypeBrevoEmailReply, encoded)
	_, err = s.asynqClient.Enqueue(task,
		asynq.Queue("integrations"),
		asynq.TaskID(brevoEmailReplyTaskID(payload.ExternalWorkspaceID, payload.EventID, payload.RecoveryAttempt)),
		asynq.MaxRetry(6),
		asynq.Timeout(45*time.Second),
		asynq.Retention(24*time.Hour),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("tasks: enqueue %s task: %w", TypeBrevoEmailReply, err)
	}
	if s.log != nil {
		s.log.Info(ctx, "Brevo email reply enqueued", "event_id", payload.EventID)
	}
	return nil
}

func brevoEmailReplyTaskID(externalWorkspaceID, eventID string, recoveryAttempt int) string {
	seed := strings.TrimSpace(externalWorkspaceID) + ":" + strings.TrimSpace(eventID)
	if recoveryAttempt > 0 {
		seed += fmt.Sprintf(":recovery:%d", recoveryAttempt)
	}
	digest := sha256.Sum256([]byte(seed))
	return "brevo-email-reply-" + hex.EncodeToString(digest[:])
}
