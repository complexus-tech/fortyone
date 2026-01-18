package chatsessions

import (
	"time"

	"github.com/google/uuid"
)

type CoreChatSession struct {
	ID          string     `json:"id"`
	UserID      uuid.UUID  `json:"userId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type CoreNewChatSession struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Title       string    `json:"title"`
	Messages    []any     `json:"messages"`
}
