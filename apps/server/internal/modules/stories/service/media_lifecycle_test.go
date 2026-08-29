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
	hardDeleteResult      HardBulkDeleteResult
	err                   error
	receivedStoryIDs      []uuid.UUID
	receivedWorkspaceID   uuid.UUID
	receivedAuthorization BulkDeleteAuthorization
}

func (r *storyMediaLifecycleRepo) HardBulkDelete(_ context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID, authorization BulkDeleteAuthorization) (HardBulkDeleteResult, error) {
	r.receivedStoryIDs = append([]uuid.UUID(nil), storyIDs...)
	r.receivedWorkspaceID = workspaceID
	r.receivedAuthorization = authorization
	return r.hardDeleteResult, r.err
}

func TestHardBulkDeleteReturnsOrphanedStoryMedia(t *testing.T) {
	storyIDs := []uuid.UUID{uuid.New(), uuid.New()}
	workspaceID := uuid.New()
	actorID := uuid.New()
	orphanedAttachmentIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &storyMediaLifecycleRepo{hardDeleteResult: HardBulkDeleteResult{
		StoryIDs:              storyIDs,
		OrphanedAttachmentIDs: orphanedAttachmentIDs,
	}}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)

	authorization := BulkDeleteAuthorization{ActorID: actorID}
	got, err := service.HardBulkDelete(context.Background(), storyIDs, workspaceID, authorization)
	if err != nil {
		t.Fatalf("hard bulk delete: %v", err)
	}
	if len(got.OrphanedAttachmentIDs) != len(orphanedAttachmentIDs) {
		t.Fatalf("orphaned attachments = %v, want %v", got.OrphanedAttachmentIDs, orphanedAttachmentIDs)
	}
	for index := range got.OrphanedAttachmentIDs {
		if got.OrphanedAttachmentIDs[index] != orphanedAttachmentIDs[index] {
			t.Fatalf("orphaned attachments = %v, want %v", got.OrphanedAttachmentIDs, orphanedAttachmentIDs)
		}
	}
	if repo.receivedWorkspaceID != workspaceID || len(repo.receivedStoryIDs) != len(storyIDs) {
		t.Fatalf("repository received stories=%v workspace=%s", repo.receivedStoryIDs, repo.receivedWorkspaceID)
	}
	if repo.receivedAuthorization != authorization {
		t.Fatalf("repository authorization = %#v, want %#v", repo.receivedAuthorization, authorization)
	}
	if got.AttachmentObjectDeletionDeferred {
		t.Fatal("legacy hard-delete path unexpectedly claimed durable attachment cleanup")
	}
}

func TestHardBulkDeleteDoesNotReturnMediaWhenDeletionFails(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repo := &storyMediaLifecycleRepo{
		hardDeleteResult: HardBulkDeleteResult{OrphanedAttachmentIDs: []uuid.UUID{uuid.New()}},
		err:              wantErr,
	}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)

	got, err := service.HardBulkDelete(context.Background(), []uuid.UUID{uuid.New()}, uuid.New(), BulkDeleteAuthorization{ActorID: uuid.New()})
	if !errors.Is(err, wantErr) {
		t.Fatalf("hard bulk delete error = %v, want %v", err, wantErr)
	}
	if len(got.OrphanedAttachmentIDs) != 0 || len(got.StoryIDs) != 0 {
		t.Fatalf("hard-delete result = %#v, want empty", got)
	}
}
