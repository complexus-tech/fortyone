package chatsessionshttp

import (
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

// AppChatSession represents a chat session in the application layer
type AppChatSession struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AppNewChatSession represents a request to create a new chat session
type AppNewChatSession struct {
	ID       string `json:"id" validate:"required,len=16"`
	Title    string `json:"title" validate:"required"`
	Messages []any  `json:"messages"`
}

// AppSaveMessagesRequest represents a request to save messages for a session
type AppSaveMessagesRequest struct {
	ID       string `json:"id" validate:"required,len=16"`
	Messages []any  `json:"messages" validate:"required"`
}

// AppUpdateSessionRequest represents a request to update a session
type AppUpdateSessionRequest struct {
	Title string `json:"title" validate:"required"`
}

type GetUserMessageCountResponse struct {
	Count int `json:"count"`
}

// Conversion functions
func toAppChatSession(s chatsessions.CoreChatSession) AppChatSession {
	return AppChatSession{
		ID:          s.ID,
		UserID:      s.UserID,
		WorkspaceID: s.WorkspaceID,
		Title:       s.Title,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toAppChatSessions(sessions []chatsessions.CoreChatSession) []AppChatSession {
	result := make([]AppChatSession, len(sessions))
	for i, session := range sessions {
		result[i] = toAppChatSession(session)
	}
	return result
}
