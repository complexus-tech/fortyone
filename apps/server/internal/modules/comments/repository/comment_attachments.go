package commentsrepository

import (
	"context"
	"fmt"
	"io"
	"strings"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

func commentAttachmentIDs(content string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	tokens := html.NewTokenizer(strings.NewReader(content))
	for {
		switch tokens.Next() {
		case html.ErrorToken:
			if tokens.Err() == io.EOF {
				return ids, nil
			}
			return nil, fmt.Errorf("parse comment attachments: %w", tokens.Err())
		case html.StartTagToken, html.SelfClosingTagToken:
			for _, attribute := range tokens.Token().Attr {
				if attribute.Key != "data-comment-attachment-id" {
					continue
				}
				id, err := uuid.Parse(attribute.Val)
				if err != nil || id == uuid.Nil {
					return nil, fmt.Errorf("%w: attachment identifier is invalid", commentsdomain.ErrInvalidComment)
				}
				if _, exists := seen[id]; !exists {
					ids = append(ids, id)
					seen[id] = struct{}{}
				}
			}
		}
	}
}

func syncCommentAttachments(ctx context.Context, queries commentsql.Querier, comment commentsdomain.Comment, workspaceID uuid.UUID) error {
	ids, err := commentAttachmentIDs(comment.Comment)
	if err != nil {
		return err
	}
	retained, err := queries.ReplaceCommentAttachmentReferences(ctx, commentsql.ReplaceCommentAttachmentReferencesParams{
		CommentID: comment.ID, WorkspaceID: workspaceID, AttachmentIds: ids,
	})
	if err != nil {
		return fmt.Errorf("persist comment attachment references: %w", err)
	}
	if retained != int64(len(ids)) {
		return fmt.Errorf("%w: one or more attachments are unavailable for this story", commentsdomain.ErrInvalidComment)
	}
	return nil
}
