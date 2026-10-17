package chatrepository

import (
	"context"
	"errors"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) MarkUnread(ctx context.Context, command domain.MarkUnreadCommand) (domain.MarkUnreadState, error) {
	result := domain.MarkUnreadState{ReadState: domain.ReadState{ConversationID: command.ConversationID, ThreadRootID: command.ThreadRootID}}
	if _, err := r.authorize(ctx, command.Scope, command.ConversationID, true); err != nil {
		return result, err
	}
	message, err := r.identity(ctx, command.Scope, command.FromMessageID)
	if err != nil {
		return result, err
	}
	if message.DeletedAt != nil {
		return result, domain.ErrNotFound
	}
	if message.ConversationID != command.ConversationID || !sameID(message.ThreadRootID, command.ThreadRootID) {
		return result, domain.ErrInvalid
	}

	// Conversation locking serializes reads, unread rewinds and message writes.
	// An earlier unread cursor must remain unread; this operation never acknowledges
	// messages. An absent cursor already means that the stream is entirely unread.
	q := r.query(ctx)
	userID := command.Scope.Actor.PrincipalID
	sequence := message.Sequence - 1
	if command.ThreadRootID == nil {
		result.LastReadSequence, err = q.MarkConversationUnread(ctx, chatsql.MarkConversationUnreadParams{
			ConversationID: command.ConversationID, UserID: userID, Sequence: sequence,
		})
	} else {
		result.LastReadSequence, err = q.MarkThreadUnread(ctx, chatsql.MarkThreadUnreadParams{
			ThreadRootID: *command.ThreadRootID, UserID: userID, Sequence: sequence,
		})
	}
	changed := !errors.Is(err, pgx.ErrNoRows)
	if !changed {
		if command.ThreadRootID == nil {
			result.LastReadSequence, err = q.GetConversationRead(ctx, chatsql.GetConversationReadParams{ConversationID: command.ConversationID, UserID: userID})
		} else {
			result.LastReadSequence, err = q.GetThreadRead(ctx, chatsql.GetThreadReadParams{ThreadRootID: *command.ThreadRootID, UserID: userID})
		}
		if errors.Is(err, pgx.ErrNoRows) {
			result.LastReadSequence, err = 0, nil
		}
	}
	if err != nil {
		return result, err
	}
	count, err := q.CountUnreadMessages(ctx, chatsql.CountUnreadMessagesParams{
		WorkspaceID: command.Scope.WorkspaceID, ConversationID: command.ConversationID,
		ThreadRootID: command.ThreadRootID, UserID: userID, Sequence: result.LastReadSequence,
	})
	if err != nil {
		return result, err
	}
	result.UnreadCount = int(count)
	if command.ThreadRootID != nil && count > 0 {
		firstUnreadID, err := q.FirstUnreadThreadReply(ctx, chatsql.FirstUnreadThreadReplyParams{
			WorkspaceID: command.Scope.WorkspaceID, ConversationID: command.ConversationID,
			ThreadRootID: command.ThreadRootID, UserID: userID, Sequence: result.LastReadSequence,
		})
		if err != nil {
			return result, err
		}
		result.FirstUnreadReplyID = &firstUnreadID
	}
	if changed {
		if err = r.event(ctx, command.Scope, command.FromMessageID, userID, "read.updated"); err != nil {
			return result, err
		}
	}
	return result, nil
}
