package chatrepository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) replaceReferences(ctx context.Context, scope domain.ActorScope, conversationID, messageID, authorID uuid.UUID, mentions, attachments []uuid.UUID, text string, rich json.RawMessage) error {
	q := r.query(ctx)
	if err := q.ReplaceMessageMentions(ctx, chatsql.ReplaceMessageMentionsParams{MessageID: messageID}); err != nil {
		return err
	}
	for _, userID := range mentions {
		count, err := q.InsertMessageMention(ctx, chatsql.InsertMessageMentionParams{WorkspaceID: scope.WorkspaceID, ConversationID: conversationID, MessageID: messageID, UserID: userID})
		if err != nil {
			return err
		}
		if count != 1 {
			return domain.ErrInvalid
		}
	}
	for _, id := range attachments {
		count, err := q.ClaimAttachment(ctx, chatsql.ClaimAttachmentParams{WorkspaceID: scope.WorkspaceID, ConversationID: conversationID, MessageID: &messageID, UserID: authorID, AttachmentID: id})
		if err != nil {
			return err
		}
		if count != 1 {
			return domain.ErrInvalid
		}
	}
	removed, err := q.RemoveMessageAttachments(ctx, chatsql.RemoveMessageAttachmentsParams{MessageID: &messageID, KeepIds: append([]uuid.UUID{}, attachments...)})
	if err != nil {
		return err
	}
	if err := r.retireRemovedAttachments(ctx, scope.WorkspaceID, removed); err != nil {
		return err
	}
	if err := q.DeleteMessageLinks(ctx, chatsql.DeleteMessageLinksParams{MessageID: messageID}); err != nil {
		return err
	}
	for _, link := range extractLinks(text, rich) {
		if err := q.InsertMessageLink(ctx, chatsql.InsertMessageLinkParams{MessageID: messageID, URL: link, Title: link}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) LinkPendingAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID, metadata domain.AttachmentMetadata) error {
	if _, err := r.authorize(ctx, scope.ActorScope, scope.ConversationID, true); err != nil {
		return err
	}
	count, err := r.query(ctx).LinkPendingAttachment(ctx, chatsql.LinkPendingAttachmentParams{WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID, UserID: &scope.Actor.PrincipalID, AttachmentID: id, Kind: metadata.Kind, DurationMs: metadata.DurationMS})
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) AuthorizeAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) error {
	if _, err := r.authorize(ctx, scope.ActorScope, scope.ConversationID, false); err != nil {
		return err
	}
	allowed, err := r.query(ctx).AuthorizeAttachment(ctx, chatsql.AuthorizeAttachmentParams{WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID, UserID: scope.Actor.PrincipalID, AttachmentID: id})
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) UnlinkPendingAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) error {
	if _, err := r.authorize(ctx, scope.ActorScope, scope.ConversationID, true); err != nil {
		return err
	}
	removed, err := r.query(ctx).UnlinkPendingAttachment(ctx, chatsql.UnlinkPendingAttachmentParams{WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID, UserID: scope.Actor.PrincipalID, AttachmentID: id})
	if err != nil {
		return err
	}
	return r.retireRemovedAttachments(ctx, scope.WorkspaceID, removed)
}

func (r *Repository) retireRemovedAttachments(ctx context.Context, workspaceID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 || r.retireAttachments == nil {
		return nil
	}
	return r.retireAttachments(ctx, workspaceID, ids)
}

func (r *Repository) Resources(ctx context.Context, q domain.SearchQuery) (domain.ResourcePage, error) {
	result := domain.ResourcePage{Items: []domain.Resource{}}
	if _, err := r.authorize(ctx, q.Scope, q.ConversationID, false); err != nil {
		return result, err
	}
	params := chatsql.ListResourcesParams{WorkspaceID: q.Scope.WorkspaceID, ConversationID: q.ConversationID, Kind: q.Kind, Search: q.Query, OldestFirst: q.Sort == "oldest", PageLimit: int32(q.Limit + 1)}
	resourceScope := resourceCursorScope(q)
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil || c.ConversationID != q.ConversationID || c.Time == nil || c.ID == nil || c.ResourceScope != resourceScope {
			return result, domain.ErrInvalid
		}
		params.CursorTime, params.CursorID = c.Time, c.ID
	}
	rows, err := r.query(ctx).ListResources(ctx, params)
	if err != nil {
		return result, err
	}
	if len(rows) > q.Limit {
		last := rows[q.Limit-1]
		result.NextCursor = cursorString(cursor{ConversationID: q.ConversationID, Time: &last.CreatedAt, ID: &last.ID, ResourceScope: resourceScope})
		rows = rows[:q.Limit]
	}
	for _, row := range rows {
		resource, err := decode[domain.Resource](row.Value)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, resource)
	}
	return result, nil
}

func resourceCursorScope(q domain.SearchQuery) string {
	// Keep cursors bounded even for a maximum-length Unicode search.
	digest := sha256.Sum256([]byte(q.Kind + "\x00" + q.Sort + "\x00" + q.Query))
	return hex.EncodeToString(digest[:])
}
