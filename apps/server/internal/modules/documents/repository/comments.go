package documentsrepository

import (
	"context"
	"fmt"
	"time"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
)

func (repository *Repository) ListComments(ctx context.Context, workspaceID, userID, documentID uuid.UUID) ([]documentdomain.CommentThread, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.commentQueries(ctx).ListDocumentComments(ctx, documentssql.ListDocumentCommentsParams{
		ActorID: userID, DocumentID: documentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list document comments: %w", err)
	}
	threads := make([]documentdomain.CommentThread, 0)
	indexes := make(map[uuid.UUID]int)
	for _, row := range rows {
		index, exists := indexes[row.ThreadID]
		if !exists {
			index = len(threads)
			indexes[row.ThreadID] = index
			threads = append(threads, documentdomain.CommentThread{
				ID: row.ThreadID, DocumentID: row.DocumentID, Quote: row.Quote,
				AnchorStart: row.AnchorStart, AnchorEnd: row.AnchorEnd,
				CreatedBy: row.ThreadCreatedBy, ResolvedAt: row.ResolvedAt,
				ResolvedBy: row.ResolvedBy, CreatedAt: row.ThreadCreatedAt,
				Comments: make([]documentdomain.Comment, 0, 1),
			})
		}
		threads[index].Comments = append(threads[index].Comments, commentFromRow(
			row.CommentID, row.Body, row.CreatedBy, row.AuthorName, row.AuthorAvatar, row.CreatedAt, row.UpdatedAt, row.ReplyToCommentID,
		))
	}
	return threads, nil
}

func (repository *Repository) ResolveComment(ctx context.Context, input documentdomain.ResolveCommentInput) error {
	if err := repository.configured(); err != nil {
		return err
	}
	actorID := input.UserID
	_, err := repository.commentQueries(ctx).SetDocumentCommentResolved(ctx, documentssql.SetDocumentCommentResolvedParams{
		Resolved: input.Resolved, ActorID: &actorID, ThreadID: input.ThreadID,
		DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
	})
	if err != nil {
		return mapNotFound("resolve document comment", err)
	}
	return nil
}

func commentFromRow(id uuid.UUID, body string, createdBy uuid.UUID, authorName string, authorAvatar *string, createdAt, updatedAt time.Time, replyTo *uuid.UUID) documentdomain.Comment {
	return documentdomain.Comment{
		ID: id, Body: body, CreatedBy: createdBy, AuthorName: authorName, ReplyToCommentID: replyTo,
		AuthorAvatar: authorAvatar, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}
