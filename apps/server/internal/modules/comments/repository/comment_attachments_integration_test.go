//go:build integration

package commentsrepository

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func TestCommentAttachmentReferencesProtectFilesAndReleaseOnEditOrDelete(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "files")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "file-author")
	actor := commentIntegrationActor(t, authorID, workspaceID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "comment-files"), postgres.Pool))
	attachments := attachmentsrepository.New(postgres.Pool)
	for _, release := range []string{"edit", "delete"} {
		t.Run(release, func(t *testing.T) {
			attachmentID := uuid.New()
			if _, err := postgres.Pool.Exec(ctx, `INSERT INTO attachments (attachment_id,filename,blob_name,size,mime_type,uploaded_by,workspace_id) VALUES ($1,'brief.pdf',$2,12,'application/pdf',$3,$4)`, attachmentID, attachmentID.String(), authorID, workspaceID); err != nil {
				t.Fatal(err)
			}
			if _, err := postgres.Pool.Exec(ctx, `INSERT INTO story_attachments (story_id,attachment_id) VALUES ($1,$2)`, storyID, attachmentID); err != nil {
				t.Fatal(err)
			}
			body := `<div data-comment-attachment-id="` + attachmentID.String() + `"><a>brief.pdf</a></div>`
			created, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: actor, Content: body})
			if err != nil {
				t.Fatal(err)
			}
			if err := attachments.DeleteAttachment(ctx, attachmentID, workspaceID); !errors.Is(err, attachmentdomain.ErrCommentReference) {
				t.Fatalf("referenced attachment deletion error = %v", err)
			}
			scope := comments.AuthorScope{CommentID: created.ID, WorkspaceID: workspaceID, Actor: actor}
			if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{Scope: scope, Content: `<div data-comment-attachment-id="` + uuid.NewString() + `"></div>`}); !errors.Is(err, comments.ErrInvalidComment) {
				t.Fatalf("unlinked attachment accepted: %v", err)
			}
			if err := attachments.DeleteAttachment(ctx, attachmentID, workspaceID); !errors.Is(err, attachmentdomain.ErrCommentReference) {
				t.Fatalf("failed edit lost file reference: %v", err)
			}
			if release == "edit" {
				if err := service.UpdateComment(ctx, comments.UpdateCommentCommand{Scope: scope, Content: "File removed"}); err != nil {
					t.Fatal(err)
				}
			} else if err := service.DeleteComment(ctx, scope); err != nil {
				t.Fatal(err)
			}
			if err := attachments.DeleteAttachment(ctx, attachmentID, workspaceID); err != nil {
				t.Fatalf("released file could not be removed: %v", err)
			}
		})
	}
}

func TestCommentAttachmentReferencesAllowWorkspaceCascadeDeletion(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "file-cascade")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "cascade-author")
	attachmentID := uuid.New()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO attachments (attachment_id, filename, blob_name, size, mime_type, uploaded_by, workspace_id)
		VALUES ($1, 'brief.pdf', $2, 12, 'application/pdf', $3, $4)
	`, attachmentID, attachmentID.String(), authorID, workspaceID); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.Pool.Exec(ctx, `INSERT INTO story_attachments (story_id, attachment_id) VALUES ($1, $2)`, storyID, attachmentID); err != nil {
		t.Fatal(err)
	}
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "comment-cascade"), postgres.Pool))
	created, err := service.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
		Actor:       commentIntegrationActor(t, authorID, workspaceID),
		Content:     `<div data-comment-attachment-id="` + attachmentID.String() + `"><a>brief.pdf</a></div>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attachmentsrepository.New(postgres.Pool).DeleteAttachment(ctx, attachmentID, workspaceID); !errors.Is(err, attachmentdomain.ErrCommentReference) {
		t.Fatalf("referenced attachment deletion error = %v", err)
	}
	// Workspace retention deletes the aggregate in one statement. Its attachment
	// and comment cascades must succeed regardless of their FK trigger order.
	if _, err := postgres.Pool.Exec(ctx, `DELETE FROM workspaces WHERE workspace_id = $1`, workspaceID); err != nil {
		t.Fatalf("delete workspace containing a comment attachment: %v", err)
	}
	var remains bool
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM stories WHERE id = $1)
		    OR EXISTS (SELECT 1 FROM story_comments WHERE comment_id = $2)
		    OR EXISTS (SELECT 1 FROM story_comment_attachments WHERE comment_id = $2)
		    OR EXISTS (SELECT 1 FROM attachments WHERE attachment_id = $3)
	`, storyID, created.ID, attachmentID).Scan(&remains); err != nil {
		t.Fatal(err)
	}
	if remains {
		t.Fatal("workspace deletion left a story, comment, attachment reference, or attachment behind")
	}
}
