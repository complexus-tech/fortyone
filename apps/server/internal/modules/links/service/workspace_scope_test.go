package links

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type scopedLinkRepo struct {
	created       CoreNewLink
	actorID       uuid.UUID
	updatedLinkID uuid.UUID
	deletedLinkID uuid.UUID
	workspaceID   uuid.UUID
}

func (r *scopedLinkRepo) CreateLink(_ context.Context, actorID uuid.UUID, input CoreNewLink) (CoreLink, error) {
	r.actorID = actorID
	r.created = input
	return CoreLink{StoryID: input.StoryID}, nil
}

func (r *scopedLinkRepo) UpdateLink(_ context.Context, actorID, linkID, workspaceID uuid.UUID, _ CoreUpdateLink) error {
	r.actorID = actorID
	r.updatedLinkID = linkID
	r.workspaceID = workspaceID
	return nil
}

func (r *scopedLinkRepo) DeleteLink(_ context.Context, actorID, linkID, workspaceID uuid.UUID) error {
	r.actorID = actorID
	r.deletedLinkID = linkID
	r.workspaceID = workspaceID
	return nil
}

func TestLinkMutationsPropagateWorkspaceIdentity(t *testing.T) {
	repo := &scopedLinkRepo{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo)
	storyID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	linkID := uuid.New()

	if _, err := service.CreateLink(context.Background(), actorID, CoreNewLink{StoryID: storyID, WorkspaceID: workspaceID, URL: "https://example.com"}); err != nil {
		t.Fatalf("create link: %v", err)
	}
	if repo.created.WorkspaceID != workspaceID || repo.actorID != actorID {
		t.Fatalf("create scope = %s/%s, want %s/%s", repo.created.WorkspaceID, repo.actorID, workspaceID, actorID)
	}
	if err := service.UpdateLink(context.Background(), actorID, linkID, workspaceID, CoreUpdateLink{}); err != nil {
		t.Fatalf("update link: %v", err)
	}
	if repo.updatedLinkID != linkID || repo.workspaceID != workspaceID || repo.actorID != actorID {
		t.Fatalf("update target = %s/%s/%s, want %s/%s/%s", repo.updatedLinkID, repo.workspaceID, repo.actorID, linkID, workspaceID, actorID)
	}
	if err := service.DeleteLink(context.Background(), actorID, linkID, workspaceID); err != nil {
		t.Fatalf("delete link: %v", err)
	}
	if repo.deletedLinkID != linkID || repo.workspaceID != workspaceID || repo.actorID != actorID {
		t.Fatalf("delete target = %s/%s/%s, want %s/%s/%s", repo.deletedLinkID, repo.workspaceID, repo.actorID, linkID, workspaceID, actorID)
	}
}
