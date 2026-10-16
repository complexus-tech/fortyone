package commentsrepository

import (
	"context"
	"errors"
	"testing"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	"github.com/google/uuid"
)

func TestCommentAttachmentIDsParseAndDeduplicateStructuredReferences(t *testing.T) {
	id := uuid.New()
	ids, err := commentAttachmentIDs(`<p>Example</p><div data-comment-attachment-id="` + id.String() + `"></div><div data-comment-attachment-id='` + id.String() + `'></div>`)
	if err != nil || len(ids) != 1 || ids[0] != id {
		t.Fatalf("attachment IDs = %v, error %v", ids, err)
	}
	for _, value := range []string{"invalid", uuid.Nil.String()} {
		if _, err := commentAttachmentIDs(`<div data-comment-attachment-id="` + value + `"></div>`); !errors.Is(err, commentsdomain.ErrInvalidComment) {
			t.Fatalf("invalid attachment %q accepted: %v", value, err)
		}
	}
}

type commentAttachmentQueries struct {
	commentsql.Querier
	params commentsql.ReplaceCommentAttachmentReferencesParams
	count  int64
}

func (q *commentAttachmentQueries) ReplaceCommentAttachmentReferences(_ context.Context, params commentsql.ReplaceCommentAttachmentReferencesParams) (int64, error) {
	q.params = params
	return q.count, nil
}

func TestCommentAttachmentReferencesRequireEveryScopedAttachment(t *testing.T) {
	workspaceID, attachmentID := uuid.New(), uuid.New()
	comment := commentsdomain.Comment{ID: uuid.New(), Comment: `<div data-comment-attachment-id="` + attachmentID.String() + `"></div>`}
	queries := &commentAttachmentQueries{}
	if err := syncCommentAttachments(t.Context(), queries, comment, workspaceID); !errors.Is(err, commentsdomain.ErrInvalidComment) {
		t.Fatalf("missing scoped attachment error = %v", err)
	}
	queries.count = 1
	if err := syncCommentAttachments(t.Context(), queries, comment, workspaceID); err != nil {
		t.Fatal(err)
	}
	if queries.params.WorkspaceID != workspaceID || queries.params.CommentID != comment.ID || len(queries.params.AttachmentIds) != 1 {
		t.Fatalf("reference scope = %#v", queries.params)
	}
	queries.count = 0
	comment.Comment = "Removed attachment"
	if err := syncCommentAttachments(t.Context(), queries, comment, workspaceID); err != nil || len(queries.params.AttachmentIds) != 0 {
		t.Fatalf("remove references = %#v, %v", queries.params, err)
	}
}
