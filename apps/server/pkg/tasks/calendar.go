package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCalendarSync         = "calendar:sync"
	TypeCalendarWatchRenewal = "calendar:watch:renewal"
)

type CalendarSyncPayload struct {
	ConnectionID uuid.UUID `json:"connectionId"`
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
