package chatsessionsrepository

import (
	"encoding/json"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

type dbChatSession struct {
	ID          string     `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	Title       string     `db:"title"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type dbChatMessage struct {
	ID        uuid.UUID       `db:"id"`
	SessionID string          `db:"session_id"`
	Messages  json.RawMessage `db:"messages"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

func toCoreChatSession(s dbChatSession) chatsessions.CoreChatSession {
	return chatsessions.CoreChatSession{
		ID:          s.ID,
		UserID:      s.UserID,
		WorkspaceID: s.WorkspaceID,
		Title:       s.Title,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		DeletedAt:   s.DeletedAt,
	}
}

func toCoreChatSessions(sessions []dbChatSession) []chatsessions.CoreChatSession {
	result := make([]chatsessions.CoreChatSession, len(sessions))
	for i, session := range sessions {
		result[i] = toCoreChatSession(session)
	}
	return result
}
