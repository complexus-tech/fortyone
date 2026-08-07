package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type storyMediaLifecycleRepo struct {
	Repository
	orphanedAttachmentIDs []uuid.UUID
	err                   error
	receivedStoryIDs      []uuid.UUID
	receivedWorkspaceID   uuid.UUID
}

func (r *storyMediaLifecycleRepo) HardBulkDelete(_ context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	r.receivedStoryIDs = append([]uuid.UUID(nil), storyIDs...)
	r.receivedWorkspaceID = workspaceID
	return append([]uuid.UUID(nil), r.orphanedAttachmentIDs...), r.err
}

func TestHardBulkDeleteReturnsOrphanedStoryMedia(t *testing.T) {
	storyIDs := []uuid.UUID{uuid.New(), uuid.New()}
	workspaceID := uuid.New()
	orphanedAttachmentIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &storyMediaLifecycleRepo{orphanedAttachmentIDs: orphanedAttachmentIDs}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)

	got, err := service.HardBulkDelete(context.Background(), storyIDs, workspaceID)
	if err != nil {
		t.Fatalf("hard bulk delete: %v", err)
	}
	if len(got) != len(orphanedAttachmentIDs) {
		t.Fatalf("orphaned attachments = %v, want %v", got, orphanedAttachmentIDs)
	}
	for index := range got {
		if got[index] != orphanedAttachmentIDs[index] {
			t.Fatalf("orphaned attachments = %v, want %v", got, orphanedAttachmentIDs)
		}
	}
	if repo.receivedWorkspaceID != workspaceID || len(repo.receivedStoryIDs) != len(storyIDs) {
		t.Fatalf("repository received stories=%v workspace=%s", repo.receivedStoryIDs, repo.receivedWorkspaceID)
	}
}

func TestHardBulkDeleteDoesNotReturnMediaWhenDeletionFails(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repo := &storyMediaLifecycleRepo{
		orphanedAttachmentIDs: []uuid.UUID{uuid.New()},
		err:                   wantErr,
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)

	got, err := service.HardBulkDelete(context.Background(), []uuid.UUID{uuid.New()}, uuid.New())
	if !errors.Is(err, wantErr) {
		t.Fatalf("hard bulk delete error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Fatalf("orphaned attachments = %v, want nil", got)
	}
}
