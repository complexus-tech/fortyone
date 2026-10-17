package chatrepository

import (
	"context"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) GetConversation(ctx context.Context, scope domain.ActorScope, id, mayaID uuid.UUID) (domain.Conversation, error) {
	if _, err := r.authorize(ctx, scope, id, false); err != nil {
		return domain.Conversation{}, err
	}
	return r.conversation(ctx, scope, id, mayaID)
}

func (r *Repository) conversation(ctx context.Context, scope domain.ActorScope, id, mayaID uuid.UUID) (domain.Conversation, error) {
	raw, err := r.query(ctx).GetConversation(ctx, chatsql.GetConversationParams{WorkspaceID: scope.WorkspaceID, ConversationID: id, UserID: scope.Actor.PrincipalID, MayaID: mayaID})
	if err != nil {
		return domain.Conversation{}, scopedError(err)
	}
	return decode[domain.Conversation](raw)
}

func (r *Repository) ListConversations(ctx context.Context, q domain.ListQuery, mayaID uuid.UUID) (domain.ConversationPage, error) {
	result := domain.ConversationPage{Items: []domain.Conversation{}}
	params := chatsql.ListConversationsParams{WorkspaceID: q.Scope.WorkspaceID, UserID: q.Scope.Actor.PrincipalID, MayaID: mayaID, Unrestricted: q.Scope.Actor.TeamAccess.IsUnrestricted(), TeamIds: q.Scope.Actor.TeamAccess.RestrictedTeamIDs(), Search: q.Search, PageLimit: int32(q.Limit + 1)}
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil || c.Time == nil || c.ID == nil {
			return result, domain.ErrInvalid
		}
		params.CursorTime, params.CursorID = c.Time, c.ID
	}
	rows, err := r.query(ctx).ListConversations(ctx, params)
	if err != nil {
		return result, err
	}
	if len(rows) > q.Limit {
		last := rows[q.Limit-1]
		result.NextCursor = cursorString(cursor{Time: last.ActivityAt, ID: &last.ID})
		rows = rows[:q.Limit]
	}
	for _, row := range rows {
		item, err := decode[domain.Conversation](row.Value)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (r *Repository) CreateDirect(ctx context.Context, scope domain.ActorScope, peer, mayaID uuid.UUID) (domain.Conversation, error) {
	if _, err := r.query(ctx).LockChatWorkspace(ctx, chatsql.LockChatWorkspaceParams{WorkspaceID: scope.WorkspaceID}); err != nil {
		return domain.Conversation{}, scopedError(err)
	}
	id, err := r.query(ctx).CreateDirectConversation(ctx, chatsql.CreateDirectConversationParams{WorkspaceID: scope.WorkspaceID, UserID: scope.Actor.PrincipalID, PeerID: peer})
	if err != nil {
		return domain.Conversation{}, scopedError(err)
	}
	return r.conversation(ctx, scope, id, mayaID)
}
