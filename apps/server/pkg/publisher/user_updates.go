package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	CalendarUpdatedType    = "calendar.updated"
	userUpdatesChannelBase = "user-updates:"
)

type UserUpdate struct {
	Type         string    `json:"type"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
	UserID       uuid.UUID `json:"userId"`
	ConnectionID uuid.UUID `json:"connectionId,omitempty"`
	SyncedAt     time.Time `json:"syncedAt,omitempty"`
}

func UserUpdatesChannel(userID uuid.UUID) string {
	return userUpdatesChannelBase + userID.String()
}

func (p *Publisher) PublishCalendarUpdated(ctx context.Context, workspaceID, userID, connectionID uuid.UUID, syncedAt time.Time) error {
	update := UserUpdate{
		Type:         CalendarUpdatedType,
		WorkspaceID:  workspaceID,
		UserID:       userID,
		ConnectionID: connectionID,
		SyncedAt:     syncedAt.UTC(),
	}
	payload, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("marshal calendar update: %w", err)
	}
	if err := p.redis.Publish(ctx, UserUpdatesChannel(userID), payload).Err(); err != nil {
		return fmt.Errorf("publish calendar update: %w", err)
	}
	return nil
}
