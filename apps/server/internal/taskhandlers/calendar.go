package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type CalendarSyncProcessor interface {
	SyncConnectionFromNotification(ctx context.Context, connectionID uuid.UUID) error
	RenewExpiringNotificationChannels(ctx context.Context) (int, error)
	DispatchReadyScheduleEventOutbox(ctx context.Context) (int, error)
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

func (h *handlers) HandleCalendarScheduleOutboxDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.calendar == nil {
		return fmt.Errorf("calendar sync processor is not configured")
	}
	_, err := h.calendar.DispatchReadyScheduleEventOutbox(ctx)
	return err
}

func (h *handlers) HandleCalendarScheduleReconcile(ctx context.Context, task *asynq.Task) error {
	if h.mayaService == nil {
		return fmt.Errorf("Maya schedule reconciler is not configured")
	}
	var payload tasks.CalendarScheduleReconcilePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode calendar schedule reconciliation payload: %w", err)
	}
	if (payload.UserID == nil || *payload.UserID == uuid.Nil) && (payload.StoryID == nil || *payload.StoryID == uuid.Nil) {
		return fmt.Errorf("calendar schedule reconciliation target is required")
	}
	return h.mayaService.ReconcileSchedule(ctx, maya.ReconcileScheduleInput{WorkspaceID: payload.WorkspaceID, UserID: payload.UserID, StoryID: payload.StoryID})
}
