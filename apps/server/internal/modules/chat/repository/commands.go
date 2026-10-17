package chatrepository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"slices"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Creation receipts hash the original canonical request, independently of later edits.
func creationHash(c domain.CreateMessageCommand) ([]byte, error) {
	var rich any
	if len(c.RichContent) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(c.RichContent))
		decoder.UseNumber()
		if err := decoder.Decode(&rich); err != nil {
			return nil, domain.ErrInvalid
		}
	}
	ids := func(values []uuid.UUID) []string {
		result := make([]string, 0, len(values))
		for _, id := range values {
			result = append(result, id.String())
		}
		slices.Sort(result)
		return result
	}
	agents := append([]string{}, c.AgentMentions...)
	slices.Sort(agents)
	raw, err := json.Marshal(struct {
		ThreadRootID *uuid.UUID
		Text         string
		Rich         any
		Mentions     []string
		Agents       []string
		Attachments  []string
	}{c.ThreadRootID, c.Text, rich, ids(c.MentionedUserIDs), agents, ids(c.AttachmentIDs)})
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(raw)
	return hash[:], nil
}

func (r *Repository) CreateMessage(ctx context.Context, c domain.CreateMessageCommand, authorID uuid.UUID) (domain.Message, error) {
	kind, err := r.authorize(ctx, c.Scope, c.ConversationID, true)
	if err != nil {
		return domain.Message{}, err
	}
	if kind != "team" && len(c.AgentMentions) > 0 {
		return domain.Message{}, domain.ErrInvalid
	}
	hash, err := creationHash(c)
	if err != nil {
		return domain.Message{}, err
	}
	q := r.query(ctx)
	existing, err := q.FindMessageCreation(ctx, chatsql.FindMessageCreationParams{WorkspaceID: c.Scope.WorkspaceID, ConversationID: c.ConversationID, AuthorID: authorID, CreationID: c.CreationID})
	if err == nil {
		if !bytes.Equal(hash, existing.CreationHash) {
			return domain.Message{}, domain.ErrConflict
		}
		message, err := r.message(ctx, c.Scope, existing.ID)
		message.Replayed = true
		return message, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Message{}, err
	}
	if c.ThreadRootID != nil {
		root, err := r.identity(ctx, c.Scope, *c.ThreadRootID)
		if err != nil {
			return domain.Message{}, err
		}
		if root.ConversationID != c.ConversationID || root.ThreadRootID != nil {
			return domain.Message{}, domain.ErrNotFound
		}
		if root.DeletedAt != nil {
			message, err := r.message(ctx, c.Scope, root.ID)
			if err != nil {
				return domain.Message{}, err
			}
			if message.ReplyCount == 0 {
				return domain.Message{}, domain.ErrNotFound
			}
		}
	}
	sequence, err := q.AllocateMessageSequence(ctx, chatsql.AllocateMessageSequenceParams{WorkspaceID: c.Scope.WorkspaceID, ConversationID: c.ConversationID})
	if err != nil {
		return domain.Message{}, err
	}
	agents := append([]string{}, c.AgentMentions...)
	id, err := q.InsertMessage(ctx, chatsql.InsertMessageParams{WorkspaceID: c.Scope.WorkspaceID, ConversationID: c.ConversationID, Sequence: sequence, ThreadRootID: c.ThreadRootID, AuthorID: authorID, CreationID: c.CreationID, CreationHash: hash, Text: c.Text, RichContent: c.RichContent, AgentMentions: agents})
	if err != nil {
		return domain.Message{}, err
	}
	if err = r.replaceReferences(ctx, c.Scope, c.ConversationID, id, authorID, c.MentionedUserIDs, c.AttachmentIDs, c.Text, c.RichContent); err != nil {
		return domain.Message{}, err
	}
	if err = r.event(ctx, c.Scope, id, authorID, "message.created"); err != nil {
		return domain.Message{}, err
	}
	return r.message(ctx, c.Scope, id)
}

func (r *Repository) UpdateMessage(ctx context.Context, c domain.UpdateMessageCommand) (domain.Message, error) {
	identity, err := r.identity(ctx, c.Scope, c.MessageID)
	if err != nil {
		return domain.Message{}, err
	}
	kind, err := r.authorize(ctx, c.Scope, identity.ConversationID, true)
	if err != nil {
		return domain.Message{}, err
	}
	if identity.AuthorID != c.Scope.Actor.PrincipalID {
		return domain.Message{}, domain.ErrForbidden
	}
	if kind != "team" && len(c.AgentMentions) > 0 {
		return domain.Message{}, domain.ErrInvalid
	}
	affected, err := r.query(ctx).UpdateMessage(ctx, chatsql.UpdateMessageParams{WorkspaceID: c.Scope.WorkspaceID, MessageID: c.MessageID, UserID: c.Scope.Actor.PrincipalID, Version: int32(c.Version), Text: c.Text, RichContent: c.RichContent, AgentMentions: append([]string{}, c.AgentMentions...)})
	if err != nil {
		return domain.Message{}, err
	}
	if affected != 1 {
		return domain.Message{}, domain.ErrConflict
	}
	if err = r.replaceReferences(ctx, c.Scope, identity.ConversationID, c.MessageID, c.Scope.Actor.PrincipalID, c.MentionedUserIDs, c.AttachmentIDs, c.Text, c.RichContent); err != nil {
		return domain.Message{}, err
	}
	if err = r.event(ctx, c.Scope, c.MessageID, c.Scope.Actor.PrincipalID, "message.updated"); err != nil {
		return domain.Message{}, err
	}
	return r.message(ctx, c.Scope, c.MessageID)
}

func (r *Repository) DeleteMessage(ctx context.Context, c domain.DeleteMessageCommand) (domain.DeletedMessage, error) {
	identity, err := r.identity(ctx, c.Scope, c.MessageID)
	if err != nil {
		return domain.DeletedMessage{}, err
	}
	if _, err = r.authorize(ctx, c.Scope, identity.ConversationID, true); err != nil {
		return domain.DeletedMessage{}, err
	}
	if identity.AuthorID != c.Scope.Actor.PrincipalID {
		return domain.DeletedMessage{}, domain.ErrForbidden
	}
	if identity.DeletedAt == nil {
		affected, err := r.query(ctx).DeleteMessage(ctx, chatsql.DeleteMessageParams{WorkspaceID: c.Scope.WorkspaceID, MessageID: c.MessageID, UserID: c.Scope.Actor.PrincipalID, Version: int32(c.Version)})
		if err != nil {
			return domain.DeletedMessage{}, err
		}
		if affected != 1 {
			return domain.DeletedMessage{}, domain.ErrConflict
		}
		if err = r.replaceReferences(ctx, c.Scope, identity.ConversationID, c.MessageID, c.Scope.Actor.PrincipalID, nil, nil, "", nil); err != nil {
			return domain.DeletedMessage{}, err
		}
		if err = r.query(ctx).DeleteMessageReactions(ctx, chatsql.DeleteMessageReactionsParams{MessageID: c.MessageID}); err != nil {
			return domain.DeletedMessage{}, err
		}
		if err = r.event(ctx, c.Scope, c.MessageID, c.Scope.Actor.PrincipalID, "message.deleted"); err != nil {
			return domain.DeletedMessage{}, err
		}
	}
	message, err := r.message(ctx, c.Scope, c.MessageID)
	if err != nil {
		return domain.DeletedMessage{}, err
	}
	return domain.DeletedMessage{ID: message.ID, ConversationID: message.ConversationID, ThreadRootID: message.ThreadRootID, Version: message.Version, DeletedAt: *message.DeletedAt}, nil
}

func (r *Repository) SetReaction(ctx context.Context, scope domain.ActorScope, id uuid.UUID, emoji string, active bool) (domain.Message, error) {
	identity, err := r.identity(ctx, scope, id)
	if err != nil {
		return domain.Message{}, err
	}
	if _, err = r.authorize(ctx, scope, identity.ConversationID, true); err != nil {
		return domain.Message{}, err
	}
	if identity.DeletedAt != nil {
		return domain.Message{}, domain.ErrNotFound
	}
	var affected int64
	if active {
		affected, err = r.query(ctx).AddReaction(ctx, chatsql.AddReactionParams{MessageID: id, UserID: scope.Actor.PrincipalID, Emoji: emoji})
	} else {
		affected, err = r.query(ctx).RemoveReaction(ctx, chatsql.RemoveReactionParams{MessageID: id, UserID: scope.Actor.PrincipalID, Emoji: emoji})
	}
	if err != nil {
		return domain.Message{}, err
	}
	if affected > 0 {
		var addedEmoji *string
		if active {
			addedEmoji = &emoji
		}
		if err = r.query(ctx).InsertReactionEvent(ctx, chatsql.InsertReactionEventParams{
			WorkspaceID: scope.WorkspaceID, MessageID: id, ActorID: scope.Actor.PrincipalID, ReactionEmoji: addedEmoji,
		}); err != nil {
			return domain.Message{}, err
		}
	}
	return r.message(ctx, scope, id)
}

func (r *Repository) MarkRead(ctx context.Context, c domain.ReadCommand) (domain.ReadState, error) {
	result := domain.ReadState{ConversationID: c.ConversationID, ThreadRootID: c.ThreadRootID}
	if _, err := r.authorize(ctx, c.Scope, c.ConversationID, true); err != nil {
		return result, err
	}
	message, err := r.identity(ctx, c.Scope, c.ThroughMessageID)
	if err != nil {
		return result, err
	}
	if message.ConversationID != c.ConversationID || !sameID(message.ThreadRootID, c.ThreadRootID) {
		return result, domain.ErrInvalid
	}
	q := r.query(ctx)
	if c.ThreadRootID == nil {
		result.LastReadSequence, err = q.MarkConversationRead(ctx, chatsql.MarkConversationReadParams{ConversationID: c.ConversationID, UserID: c.Scope.Actor.PrincipalID, Sequence: message.Sequence})
	} else {
		result.LastReadSequence, err = q.MarkThreadRead(ctx, chatsql.MarkThreadReadParams{ThreadRootID: *c.ThreadRootID, UserID: c.Scope.Actor.PrincipalID, Sequence: message.Sequence})
	}
	advanced := !errors.Is(err, pgx.ErrNoRows)
	if !advanced {
		if c.ThreadRootID == nil {
			result.LastReadSequence, err = q.GetConversationRead(ctx, chatsql.GetConversationReadParams{ConversationID: c.ConversationID, UserID: c.Scope.Actor.PrincipalID})
		} else {
			result.LastReadSequence, err = q.GetThreadRead(ctx, chatsql.GetThreadReadParams{ThreadRootID: *c.ThreadRootID, UserID: c.Scope.Actor.PrincipalID})
		}
	}
	if err != nil {
		return result, err
	}
	count, err := q.CountUnreadMessages(ctx, chatsql.CountUnreadMessagesParams{WorkspaceID: c.Scope.WorkspaceID, ConversationID: c.ConversationID, ThreadRootID: c.ThreadRootID, UserID: c.Scope.Actor.PrincipalID, Sequence: result.LastReadSequence})
	if err != nil {
		return result, err
	}
	result.UnreadCount = int(count)
	if advanced {
		if err = r.event(ctx, c.Scope, c.ThroughMessageID, c.Scope.Actor.PrincipalID, "read.updated"); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (r *Repository) event(ctx context.Context, scope domain.ActorScope, message, actor uuid.UUID, kind string) error {
	return r.query(ctx).InsertEvent(ctx, chatsql.InsertEventParams{WorkspaceID: scope.WorkspaceID, MessageID: message, ActorID: actor, EventType: kind})
}
