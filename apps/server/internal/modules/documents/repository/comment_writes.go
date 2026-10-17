package documentsrepository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"

	domain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	queries "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/documentmentions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) commentQueries(ctx context.Context) *queries.Queries {
	if tx, ok := database.UnitTransaction(ctx); ok {
		return queries.New(tx)
	}
	return r.queries
}
func (r *Repository) commentTransaction(ctx context.Context, operation func(context.Context) error) error {
	if err := r.configured(); err != nil {
		return err
	}
	if _, ok := database.UnitTransaction(ctx); ok {
		return operation(ctx)
	}
	var err error
	for attempt := 0; attempt < maximumTransactionAttempts; attempt++ {
		err = database.WithinUnitOfWork(ctx, r.pool, func(txctx context.Context, _ pgx.Tx) error { return operation(txctx) })
		if !database.IsRetryableTransactionError(err) {
			return err
		}
	}
	return err
}
func creationHash(input any) []byte {
	value, _ := json.Marshal(input)
	sum := sha256.Sum256(value)
	return sum[:]
}

func (r *Repository) replayComment(ctx context.Context, w, actor, author, document uuid.UUID, id *uuid.UUID, hash []byte) (*domain.CommentContext, error) {
	if id == nil {
		return nil, nil
	}
	row, err := r.commentQueries(ctx).GetDocumentCommentCreation(ctx, queries.GetDocumentCommentCreationParams{WorkspaceID: w, ActorID: actor, AuthorID: author, CreationID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if row.DocumentID != document || !bytes.Equal(row.CreationHash, hash) {
		return nil, domain.ErrConflict
	}
	value, err := r.GetComment(ctx, w, actor, document, row.CommentID)
	return &value, err
}
func (r *Repository) canonicalComment(ctx context.Context, w, actor, document uuid.UUID, body string) (documentmentions.Result, error) {
	members, err := r.ListCommentMentions(ctx, w, actor, document)
	if err != nil {
		return documentmentions.Result{}, err
	}
	names := make(map[uuid.UUID]string, len(members))
	for _, member := range members {
		names[member.ID] = member.Name
	}
	return documentmentions.Canonicalize(body, names), nil
}
func (r *Repository) notifyComment(ctx context.Context, w, document uuid.UUID, comment domain.Comment, reply bool) error {
	q := r.commentQueries(ctx)
	if err := q.RecordDocumentCommentMentions(ctx, queries.RecordDocumentCommentMentionsParams{CommentID: comment.ID, WorkspaceID: w, DocumentID: document, UserIds: comment.MentionedUserIDs}); err != nil {
		return err
	}
	return q.NotifyDocumentComment(ctx, queries.NotifyDocumentCommentParams{CommentID: comment.ID, IsReply: reply})
}

func (r *Repository) CreateComment(ctx context.Context, input domain.CreateCommentInput) (domain.CommentThread, error) {
	var result domain.CommentThread
	err := r.commentTransaction(ctx, func(txctx context.Context) error {
		hash := creationHash(input)
		replay, err := r.replayComment(txctx, input.WorkspaceID, input.UserID, input.UserID, input.DocumentID, input.CreationID, hash)
		if err != nil {
			return err
		}
		if replay != nil {
			result = replay.Thread
			return nil
		}
		canonical, err := r.canonicalComment(txctx, input.WorkspaceID, input.UserID, input.DocumentID, input.Body)
		if err != nil {
			return err
		}
		row, err := r.commentQueries(txctx).CreateDocumentCommentThread(txctx, queries.CreateDocumentCommentThreadParams{ActorID: input.UserID, DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID, Quote: input.Quote, AnchorStart: input.AnchorStart, AnchorEnd: input.AnchorEnd, Body: canonical.Body, CreationID: input.CreationID, CreationHash: hash})
		if err != nil {
			return mapNotFound("create document comment", err)
		}
		comment := commentFromRow(row.CommentID, row.Body, row.CreatedBy, row.AuthorName, row.AuthorAvatar, row.CreatedAt, row.UpdatedAt, row.ReplyToCommentID)
		comment.MentionedUserIDs = canonical.UserIDs
		result = domain.CommentThread{ID: row.ThreadID, DocumentID: row.DocumentID, Quote: row.Quote, AnchorStart: row.AnchorStart, AnchorEnd: row.AnchorEnd, CreatedBy: row.ThreadCreatedBy, ResolvedAt: row.ResolvedAt, ResolvedBy: row.ResolvedBy, CreatedAt: row.ThreadCreatedAt, Comments: []domain.Comment{comment}}
		return r.notifyComment(txctx, input.WorkspaceID, input.DocumentID, comment, false)
	})
	return result, err
}
func (r *Repository) ReplyToComment(ctx context.Context, input domain.ReplyCommentInput) (domain.Comment, error) {
	return r.writeCommentReply(ctx, input, input.UserID, creationHash(input))
}
func (r *Repository) CreateAgentReply(ctx context.Context, input domain.AgentReplyInput) (domain.Comment, error) {
	return r.writeCommentReply(ctx, domain.ReplyCommentInput{WorkspaceID: input.WorkspaceID, DocumentID: input.DocumentID, UserID: input.RequesterID, ThreadID: input.ThreadID, Body: input.Body, CreationID: &input.CreationID, ReplyToCommentID: input.ReplyToCommentID}, input.AuthorID, creationHash(input))
}
func (r *Repository) writeCommentReply(ctx context.Context, input domain.ReplyCommentInput, author uuid.UUID, hash []byte) (domain.Comment, error) {
	var result domain.Comment
	err := r.commentTransaction(ctx, func(txctx context.Context) error {
		replay, err := r.replayComment(txctx, input.WorkspaceID, input.UserID, author, input.DocumentID, input.CreationID, hash)
		if err != nil {
			return err
		}
		if replay != nil {
			if replay.Thread.ID != input.ThreadID {
				return domain.ErrConflict
			}
			result = replay.Comment
			return nil
		}
		canonical, err := r.canonicalComment(txctx, input.WorkspaceID, input.UserID, input.DocumentID, input.Body)
		if err != nil {
			return err
		}
		row, err := r.commentQueries(txctx).AddDocumentCommentReply(txctx, queries.AddDocumentCommentReplyParams{ActorID: input.UserID, AuthorID: author, DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID, ThreadID: input.ThreadID, Body: canonical.Body, CreationID: input.CreationID, CreationHash: hash, ReplyToCommentID: input.ReplyToCommentID})
		if err != nil {
			return mapNotFound("reply to document comment", err)
		}
		result = commentFromRow(row.CommentID, row.Body, row.CreatedBy, row.AuthorName, row.AuthorAvatar, row.CreatedAt, row.UpdatedAt, row.ReplyToCommentID)
		result.MentionedUserIDs = canonical.UserIDs
		return r.notifyComment(txctx, input.WorkspaceID, input.DocumentID, result, true)
	})
	return result, err
}
func (r *Repository) GetComment(ctx context.Context, w, actor, document, id uuid.UUID) (domain.CommentContext, error) {
	doc, err := r.Get(ctx, w, actor, document)
	if err != nil {
		return domain.CommentContext{}, err
	}
	threads, err := r.ListComments(ctx, w, actor, document)
	if err != nil {
		return domain.CommentContext{}, err
	}
	for _, thread := range threads {
		for _, comment := range thread.Comments {
			if comment.ID == id {
				ids, err := r.commentQueries(ctx).ListDocumentCommentMentionIDs(ctx, queries.ListDocumentCommentMentionIDsParams{CommentID: id})
				comment.MentionedUserIDs = ids
				return domain.CommentContext{Document: doc, Thread: thread, Comment: comment}, err
			}
		}
	}
	return domain.CommentContext{}, domain.ErrNotFound
}
func (r *Repository) ListCommentMentions(ctx context.Context, w, actor, document uuid.UUID) ([]domain.CommentMention, error) {
	if _, err := r.Get(ctx, w, actor, document); err != nil {
		return nil, err
	}
	rows, err := r.commentQueries(ctx).ListDocumentCommentMentions(ctx, queries.ListDocumentCommentMentionsParams{WorkspaceID: w, ActorID: actor, DocumentID: document})
	if err != nil {
		return nil, err
	}
	result := make([]domain.CommentMention, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.CommentMention{ID: row.UserID, Name: row.Name, Username: row.Username, AvatarURL: row.AvatarURL})
	}
	return result, nil
}
