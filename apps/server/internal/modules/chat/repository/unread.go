package chatrepository

import (
	"context"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) Unread(ctx context.Context, scope domain.ActorScope) (domain.UnreadSummary, error) {
	result := domain.UnreadSummary{Conversations: []domain.ConversationUnread{}}
	rows, err := r.query(ctx).ChatUnreadSummary(ctx, chatsql.ChatUnreadSummaryParams{
		WorkspaceID: scope.WorkspaceID, UserID: scope.Actor.PrincipalID,
		Unrestricted: scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.TotalUnread += int(row.UnreadCount)
		result.UnreadThreads += int(row.UnreadThreads)
		result.Conversations = append(result.Conversations, domain.ConversationUnread{
			ConversationID: row.ConversationID, UnreadCount: int(row.UnreadCount),
			UnreadRootCount: int(row.UnreadRootCount), UnreadReplyCount: int(row.UnreadReplyCount),
		})
	}
	return result, nil
}

func (r *Repository) UnreadThreadReadTargets(ctx context.Context, scope domain.ActorScope) ([]domain.ThreadReadTarget, error) {
	rows, err := r.query(ctx).ListUnreadChatThreadReadTargets(ctx, chatsql.ListUnreadChatThreadReadTargetsParams{
		WorkspaceID: scope.WorkspaceID, UserID: scope.Actor.PrincipalID,
		Unrestricted: scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: scope.Actor.TeamAccess.RestrictedTeamIDs(),
	})
	if err != nil {
		return nil, err
	}
	targets := make([]domain.ThreadReadTarget, 0, len(rows))
	for _, row := range rows {
		targets = append(targets, domain.ThreadReadTarget{
			ConversationID: row.ConversationID, ThreadRootID: row.ThreadRootID, ThroughMessageID: row.ThroughMessageID,
		})
	}
	return targets, nil
}

func (r *Repository) UnreadThreads(ctx context.Context, query domain.UnreadThreadQuery) (domain.UnreadThreadPage, error) {
	result := domain.UnreadThreadPage{Items: []domain.UnreadThread{}}
	if query.ConversationID != nil {
		if _, err := r.authorize(ctx, query.Scope, *query.ConversationID, false); err != nil {
			return result, err
		}
	}
	params := chatsql.ListUnreadChatThreadsParams{
		WorkspaceID: query.Scope.WorkspaceID, UserID: query.Scope.Actor.PrincipalID,
		Unrestricted: query.Scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: query.Scope.Actor.TeamAccess.RestrictedTeamIDs(),
		ConversationID: query.ConversationID, PageLimit: int32(query.Limit + 1),
	}
	cursorScope := unreadThreadCursorScope(query)
	if query.Cursor != "" {
		decoded, err := decodeUnreadThreadCursor(query.Cursor, cursorScope)
		if err != nil {
			return result, err
		}
		params.CursorTime, params.CursorID = decoded.Time, decoded.ID
	}
	rows, err := r.query(ctx).ListUnreadChatThreads(ctx, params)
	if err != nil {
		return result, err
	}
	if len(rows) > query.Limit {
		last := rows[query.Limit-1]
		result.NextCursor = cursorString(cursor{Time: &last.ActivityAt, ID: &last.ID, ResourceScope: cursorScope})
		rows = rows[:query.Limit]
	}
	for _, row := range rows {
		item, err := decode[domain.UnreadThread](row.Value)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func unreadThreadCursorScope(query domain.UnreadThreadQuery) string {
	conversationID := uuid.Nil
	if query.ConversationID != nil {
		conversationID = *query.ConversationID
	}
	return fmt.Sprintf("unread-threads:%s:%s:%s", query.Scope.WorkspaceID, query.Scope.Actor.PrincipalID, conversationID)
}

func decodeUnreadThreadCursor(value, scope string) (cursor, error) {
	decoded, err := decodeCursor(value)
	if err != nil || decoded.ResourceScope != scope || decoded.Time == nil || decoded.Time.IsZero() || decoded.ID == nil || *decoded.ID == uuid.Nil {
		return cursor{}, domain.ErrInvalid
	}
	return decoded, nil
}
