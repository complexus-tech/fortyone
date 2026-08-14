package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type CalendarSyncProcessor interface {
	SyncConnectionFromNotification(ctx context.Context, connectionID uuid.UUID) error
	RenewExpiringNotificationChannels(ctx context.Context) (int, error)
}

func (h *handlers) HandleCalendarSync(ctx context.Context, task *asynq.Task) error {
	if h.calendar == nil {
		return fmt.Errorf("calendar sync processor is not configured")
	}
	var payload tasks.CalendarSyncPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode calendar sync payload: %w", err)
	}
	if payload.ConnectionID == uuid.Nil {
		return fmt.Errorf("calendar sync connection ID is required")
	}
	return h.calendar.SyncConnectionFromNotification(ctx, payload.ConnectionID)
}

func (h *handlers) HandleCalendarWatchRenewal(ctx context.Context, _ *asynq.Task) error {
	if h.calendar == nil {
		return fmt.Errorf("calendar sync processor is not configured")
	}
	_, err := h.calendar.RenewExpiringNotificationChannels(ctx)
	return err
}
