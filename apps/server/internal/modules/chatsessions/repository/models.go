package chatsessionsrepository

import (
	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
)

func toCoreChatSession(s chatsessionssql.ChatSession) chatsessions.CoreChatSession {
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

func toCoreChatSessions(sessions []chatsessionssql.ChatSession) []chatsessions.CoreChatSession {
	result := make([]chatsessions.CoreChatSession, len(sessions))
	for i, session := range sessions {
		result[i] = toCoreChatSession(session)
	}
	return result
}
