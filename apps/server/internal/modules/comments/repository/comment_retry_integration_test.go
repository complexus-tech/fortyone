//go:build integration

package commentsrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func TestExternalCommentRetriesCreateOneCommentAndEvent(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "retry")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "retry-author")
	actor := commentIntegrationActor(t, authorID, workspaceID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "comment-retry"), postgres.Pool))
	creationID := uuid.New()
	command := comments.CreateCommentCommand{
		CreationID: &creationID, WorkspaceID: workspaceID, StoryID: storyID,
		Actor: actor, Content: "Imported campaign discussion",
	}
	var workers sync.WaitGroup
	errorsByWorker := make(chan error, 8)
	for range 8 {
		workers.Go(func() {
			comment, err := service.CreateComment(ctx, command)
			if err == nil && comment.ID != creationID {
				err = errors.New("retry returned a different comment")
			}
			errorsByWorker <- err
		})
	}
	workers.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		if err != nil {
			t.Fatalf("concurrent retry: %v", err)
		}
	}
	assertCommentEventContract(t, ctx, postgres.Pool, creationID, "comment.created", 1)
	var count int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM story_comments WHERE story_id = $1`, storyID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("comment count = %d, error %v; want one", count, err)
	}
	command.Content = "Different request"
	if _, err := service.CreateComment(ctx, command); !errors.Is(err, comments.ErrCreationConflict) {
		t.Fatalf("changed retry error = %v, want conflict", err)
	}
	command.Content = "Imported campaign discussion"
	if _, err := postgres.Pool.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, workspaceID, authorID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateComment(ctx, command); !errors.Is(err, comments.ErrNotFound) {
		t.Fatalf("revoked retry error = %v, want not found", err)
	}
}
