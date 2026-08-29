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
	TypeCalendarSync                   = "calendar:sync"
	TypeCalendarWatchRenewal           = "calendar:watch:renewal"
	TypeCalendarScheduleReconcile      = "calendar:schedule:reconcile"
	TypeCalendarWorkspaceScheduleBatch = "calendar:schedule:workspace-batch"
	TypeCalendarScheduleOutbox         = "calendar:schedule:outbox:dispatch"

	CalendarWorkspaceScheduleBatchDelay = 15 * time.Minute
	calendarWorkspaceScheduleBatchTTL   = 30 * time.Minute
)

type CalendarSyncPayload struct {
	ConnectionID uuid.UUID `json:"connectionId"`
}

type CalendarScheduleReconcilePayload struct {
	WorkspaceID *uuid.UUID `json:"workspaceId,omitempty"`
	UserID      *uuid.UUID `json:"userId,omitempty"`
	StoryID     *uuid.UUID `json:"storyId,omitempty"`
}

type CalendarWorkspaceScheduleBatchPayload struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
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
		asynq.Unique(10*time.Second),
	)
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue calendar sync: %w", err)
	}
	return nil
}

func (s *Service) EnqueueCalendarScheduleReconcile(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("calendar schedule reconciliation user ID is required")
	}
	return s.enqueueCalendarScheduleReconcile(ctx, CalendarScheduleReconcilePayload{UserID: &userID}, 10*time.Second)
}

func (s *Service) EnqueueStoryScheduleReconcile(ctx context.Context, workspaceID, storyID uuid.UUID) error {
	if workspaceID == uuid.Nil || storyID == uuid.Nil {
		return fmt.Errorf("calendar schedule reconciliation workspace and story IDs are required")
	}
	body, err := json.Marshal(CalendarScheduleReconcilePayload{WorkspaceID: &workspaceID, StoryID: &storyID})
	if err != nil {
		return fmt.Errorf("marshal calendar schedule reconciliation payload: %w", err)
	}
	_, err = s.asynqClient.EnqueueContext(
		ctx,
		asynq.NewTask(TypeCalendarScheduleReconcile, body),
		asynq.Queue("integrations"),
		asynq.Unique(5*time.Second),
	)
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue calendar schedule reconciliation: %w", err)
	}
	return nil
}

// EnqueueCalendarWorkspaceScheduleBatch opens a collection window for story
// creation and planning changes. Asynq's uniqueness lock coalesces every
// request for the same workspace into the first task waiting in that window.
func (s *Service) EnqueueCalendarWorkspaceScheduleBatch(ctx context.Context, workspaceID uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return fmt.Errorf("calendar workspace schedule batch workspace ID is required")
	}
	body, err := json.Marshal(CalendarWorkspaceScheduleBatchPayload{WorkspaceID: workspaceID})
	if err != nil {
		return fmt.Errorf("marshal calendar workspace schedule batch payload: %w", err)
	}
	_, err = s.asynqClient.EnqueueContext(
		ctx,
		asynq.NewTask(TypeCalendarWorkspaceScheduleBatch, body),
		asynq.Queue("integrations"),
		asynq.ProcessIn(CalendarWorkspaceScheduleBatchDelay),
		asynq.Unique(calendarWorkspaceScheduleBatchTTL),
	)
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue calendar workspace schedule batch: %w", err)
	}
	return nil
}

func (s *Service) enqueueCalendarScheduleReconcile(ctx context.Context, payload CalendarScheduleReconcilePayload, delay time.Duration) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal calendar schedule reconciliation payload: %w", err)
	}
	options := []asynq.Option{asynq.Queue("integrations")}
	if delay > 0 {
		options = append(options, asynq.ProcessIn(delay))
	}
	_, err = s.asynqClient.EnqueueContext(ctx, asynq.NewTask(TypeCalendarScheduleReconcile, body), options...)
	if err != nil {
		return fmt.Errorf("enqueue calendar schedule reconciliation: %w", err)
	}
	return nil
}
