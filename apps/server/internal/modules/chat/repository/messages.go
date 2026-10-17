package chatrepository

import (
	"context"
	"math"
	"slices"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) GetMessage(ctx context.Context, scope domain.ActorScope, id uuid.UUID) (domain.Message, error) {
	identity, err := r.identity(ctx, scope, id)
	if err != nil {
		return domain.Message{}, err
	}
	if _, err = r.authorize(ctx, scope, identity.ConversationID, false); err != nil {
		return domain.Message{}, err
	}
	message, err := r.message(ctx, scope, id)
	if err == nil && message.DeletedAt != nil && (message.ThreadRootID != nil || message.ReplyCount == 0) {
		return domain.Message{}, domain.ErrNotFound
	}
	return message, err
}

func (r *Repository) ListMessages(ctx context.Context, q domain.MessageQuery) (domain.MessagePage, error) {
	result := domain.MessagePage{Items: []domain.Message{}}
	if _, err := r.authorize(ctx, q.Scope, q.ConversationID, false); err != nil {
		return result, err
	}
	if q.ThreadRootID != nil {
		root, err := r.identity(ctx, q.Scope, *q.ThreadRootID)
		if err != nil {
			return result, err
		}
		if root.ConversationID != q.ConversationID || root.ThreadRootID != nil {
			return result, domain.ErrNotFound
		}
	}
	params := chatsql.ListReplyMessagesParams{WorkspaceID: q.Scope.WorkspaceID, ConversationID: q.ConversationID, UserID: q.Scope.Actor.PrincipalID, ThreadRootID: q.ThreadRootID, BeforeSequence: math.MaxInt64, Ascending: q.ThreadRootID != nil, PageLimit: int32(q.Limit)}
	var err error
	if q.Before != "" {
		params.BeforeSequence, err = sequenceCursor(q.Before, q.ConversationID, q.ThreadRootID)
		params.Ascending = false
	}
	if q.After != "" {
		params.AfterSequence, err = sequenceCursor(q.After, q.ConversationID, q.ThreadRootID)
		params.Ascending = true
	}
	if err != nil {
		return result, err
	}
	var rows []chatsql.ListReplyMessagesRow
	if q.Around != nil {
		rows, err = r.around(ctx, q, params)
	} else {
		rows, err = r.listMessageRows(ctx, params)
	}
	if err != nil {
		return result, err
	}
	slices.SortFunc(rows, func(a, b chatsql.ListReplyMessagesRow) int {
		if a.Sequence < b.Sequence {
			return -1
		}
		if a.Sequence > b.Sequence {
			return 1
		}
		return 0
	})
	for _, row := range rows {
		message, err := decode[domain.Message](row.Value)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, message)
	}
	if len(rows) == 0 {
		return result, nil
	}
	first, last := rows[0].Sequence, rows[len(rows)-1].Sequence
	bounds, err := r.pageBounds(ctx, chatsql.ReplyPageBoundsParams{WorkspaceID: q.Scope.WorkspaceID, ConversationID: q.ConversationID, ThreadRootID: q.ThreadRootID, FirstSequence: first, LastSequence: last})
	if err != nil {
		return result, err
	}
	result.HasOlder, result.HasNewer = bounds.HasOlder, bounds.HasNewer
	result.OlderCursor = cursorString(cursor{ConversationID: q.ConversationID, ThreadRootID: q.ThreadRootID, Sequence: first})
	result.NewerCursor = cursorString(cursor{ConversationID: q.ConversationID, ThreadRootID: q.ThreadRootID, Sequence: last})
	return result, nil
}

func (r *Repository) around(ctx context.Context, q domain.MessageQuery, base chatsql.ListReplyMessagesParams) ([]chatsql.ListReplyMessagesRow, error) {
	anchor, err := r.identity(ctx, q.Scope, *q.Around)
	if err != nil {
		return nil, err
	}
	if anchor.ConversationID != q.ConversationID {
		return nil, domain.ErrNotFound
	}
	if q.ThreadRootID == nil && anchor.ThreadRootID != nil {
		anchor, err = r.identity(ctx, q.Scope, *anchor.ThreadRootID)
		if err != nil {
			return nil, err
		}
	}
	if !sameID(anchor.ThreadRootID, q.ThreadRootID) {
		return nil, domain.ErrNotFound
	}
	older := base
	older.BeforeSequence = anchor.Sequence
	older.Ascending = false
	older.PageLimit = int32(q.Limit / 2)
	before, err := r.listMessageRows(ctx, older)
	if err != nil {
		return nil, err
	}
	newer := base
	newer.AfterSequence = anchor.Sequence - 1
	newer.Ascending = true
	newer.PageLimit = int32(q.Limit - len(before))
	after, err := r.listMessageRows(ctx, newer)
	if err != nil {
		return nil, err
	}
	if len(before)+len(after) < q.Limit && len(before) == int(older.PageLimit) {
		older.PageLimit = int32(q.Limit - len(after))
		before, err = r.listMessageRows(ctx, older)
		if err != nil {
			return nil, err
		}
	}
	return append(before, after...), nil
}

func (r *Repository) Search(ctx context.Context, q domain.SearchQuery) (domain.SearchPage, error) {
	result := domain.SearchPage{Items: []domain.Message{}}
	if _, err := r.authorize(ctx, q.Scope, q.ConversationID, false); err != nil {
		return result, err
	}
	before := int64(math.MaxInt64)
	if q.Cursor != "" {
		var err error
		before, err = sequenceCursor(q.Cursor, q.ConversationID, nil)
		if err != nil {
			return result, err
		}
	}
	rows, err := r.query(ctx).SearchMessages(ctx, chatsql.SearchMessagesParams{WorkspaceID: q.Scope.WorkspaceID, ConversationID: q.ConversationID, UserID: q.Scope.Actor.PrincipalID, BeforeSequence: before, Search: q.Query, PageLimit: int32(q.Limit + 1)})
	if err != nil {
		return result, err
	}
	if len(rows) > q.Limit {
		result.NextCursor = cursorString(cursor{ConversationID: q.ConversationID, Sequence: rows[q.Limit-1].Sequence})
		rows = rows[:q.Limit]
	}
	for _, row := range rows {
		item, err := decode[domain.Message](row.Value)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (r *Repository) BoundedContext(ctx context.Context, q domain.ContextQuery) ([]domain.Message, error) {
	if _, err := r.authorize(ctx, q.Scope, q.ConversationID, false); err != nil {
		return nil, err
	}
	if q.ThreadRootID != nil {
		root, err := r.identity(ctx, q.Scope, *q.ThreadRootID)
		if err != nil {
			return nil, err
		}
		if root.ConversationID != q.ConversationID || root.ThreadRootID != nil {
			return nil, domain.ErrNotFound
		}
	}
	before := int64(math.MaxInt64)
	if q.BeforeMessageID != nil {
		anchor, err := r.identity(ctx, q.Scope, *q.BeforeMessageID)
		if err != nil {
			return nil, err
		}
		if anchor.ConversationID != q.ConversationID {
			return nil, domain.ErrNotFound
		}
		before = anchor.Sequence + 1
	}
	rows, err := r.listMessageRows(ctx, chatsql.ListReplyMessagesParams{WorkspaceID: q.Scope.WorkspaceID, ConversationID: q.ConversationID, UserID: q.Scope.Actor.PrincipalID, ThreadRootID: q.ThreadRootID, BeforeSequence: before, PageLimit: int32(q.Limit)})
	if err != nil {
		return nil, err
	}
	result := make([]domain.Message, 0, len(rows))
	for _, row := range rows {
		item, err := decode[domain.Message](row.Value)
		if err != nil {
			return nil, err
		}
		if item.DeletedAt == nil {
			result = append(result, item)
		}
	}
	return result, nil
}

// Each window uses a fixed root/reply predicate and index order. Keeping ASC and
// DESC in bounded branches avoids a generic prepared plan sorting all history.
func (r *Repository) listMessageRows(ctx context.Context, p chatsql.ListReplyMessagesParams) ([]chatsql.ListReplyMessagesRow, error) {
	if p.ThreadRootID != nil {
		return r.query(ctx).ListReplyMessages(ctx, p)
	}
	rows, err := r.query(ctx).ListRootMessages(ctx, chatsql.ListRootMessagesParams{WorkspaceID: p.WorkspaceID, ConversationID: p.ConversationID, UserID: p.UserID, BeforeSequence: p.BeforeSequence, AfterSequence: p.AfterSequence, Ascending: p.Ascending, PageLimit: p.PageLimit})
	if err != nil {
		return nil, err
	}
	result := make([]chatsql.ListReplyMessagesRow, len(rows))
	for i, row := range rows {
		result[i] = chatsql.ListReplyMessagesRow{Sequence: row.Sequence, Value: row.Value}
	}
	return result, nil
}
func (r *Repository) pageBounds(ctx context.Context, p chatsql.ReplyPageBoundsParams) (chatsql.ReplyPageBoundsRow, error) {
	if p.ThreadRootID != nil {
		return r.query(ctx).ReplyPageBounds(ctx, p)
	}
	row, err := r.query(ctx).RootPageBounds(ctx, chatsql.RootPageBoundsParams{WorkspaceID: p.WorkspaceID, ConversationID: p.ConversationID, FirstSequence: p.FirstSequence, LastSequence: p.LastSequence})
	return chatsql.ReplyPageBoundsRow{HasOlder: row.HasOlder, HasNewer: row.HasNewer}, err
}
