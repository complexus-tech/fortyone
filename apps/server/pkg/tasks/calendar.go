package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCalendarSync              = "calendar:sync"
	TypeCalendarWatchRenewal      = "calendar:watch:renewal"
	TypeCalendarScheduleReconcile = "calendar:schedule:reconcile"
	TypeCalendarScheduleOutbox    = "calendar:schedule:outbox:dispatch"
)

type CalendarSyncPayload struct {
	ConnectionID uuid.UUID `json:"connectionId"`
}

type CalendarScheduleReconcilePayload struct {
	WorkspaceID *uuid.UUID `json:"workspaceId,omitempty"`
	UserID      *uuid.UUID `json:"userId,omitempty"`
	StoryID     *uuid.UUID `json:"storyId,omitempty"`
}

func (s *Service) EnqueueCalendarSync(ctx context.Context, connectionID uuid.UUID) error {
	payload, err := json.Marshal(CalendarSyncPayload{ConnectionID: connectionID})
	if err != nil {
		return fmt.Errorf("marshal calendar sync payload: %w", err)
	}

	_, err = s.asynqClient.EnqueueContext(
		ctx,
		asynq.NewTask(TypeCalendarSync, payload),
		asynq.Queue("integrations"),
	)
	if err != nil {
		return fmt.Errorf("enqueue calendar sync: %w", err)
	}
	return nil
}

func (s *Service) EnqueueCalendarScheduleReconcile(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("calendar schedule reconciliation user ID is required")
	}
	return s.enqueueCalendarScheduleReconcile(ctx, CalendarScheduleReconcilePayload{UserID: &userID})
}

func (s *Service) EnqueueStoryScheduleReconcile(ctx context.Context, workspaceID, storyID uuid.UUID) error {
	if workspaceID == uuid.Nil || storyID == uuid.Nil {
		return fmt.Errorf("calendar schedule reconciliation workspace and story IDs are required")
	}
	return s.enqueueCalendarScheduleReconcile(ctx, CalendarScheduleReconcilePayload{WorkspaceID: &workspaceID, StoryID: &storyID})
}

func (s *Service) enqueueCalendarScheduleReconcile(ctx context.Context, payload CalendarScheduleReconcilePayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal calendar schedule reconciliation payload: %w", err)
	}
	_, err = s.asynqClient.EnqueueContext(
		ctx,
		asynq.NewTask(TypeCalendarScheduleReconcile, body),
		asynq.Queue("integrations"),
		asynq.ProcessIn(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("enqueue calendar schedule reconciliation: %w", err)
	}
	return nil
}
