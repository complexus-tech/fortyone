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
	updatedLinkID uuid.UUID
	deletedLinkID uuid.UUID
	workspaceID   uuid.UUID
}

func (r *scopedLinkRepo) CreateLink(_ context.Context, input CoreNewLink) (CoreLink, error) {
	r.created = input
	return CoreLink{StoryID: input.StoryID}, nil
}

func (r *scopedLinkRepo) UpdateLink(_ context.Context, linkID, workspaceID uuid.UUID, _ CoreUpdateLink) error {
	r.updatedLinkID = linkID
	r.workspaceID = workspaceID
	return nil
}

func (r *scopedLinkRepo) DeleteLink(_ context.Context, linkID, workspaceID uuid.UUID) error {
	r.deletedLinkID = linkID
	r.workspaceID = workspaceID
	return nil
}

func TestLinkMutationsPropagateWorkspaceIdentity(t *testing.T) {
	repo := &scopedLinkRepo{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo)
	storyID := uuid.New()
	workspaceID := uuid.New()
	linkID := uuid.New()

	if _, err := service.CreateLink(context.Background(), CoreNewLink{StoryID: storyID, WorkspaceID: workspaceID, URL: "https://example.com"}); err != nil {
		t.Fatalf("create link: %v", err)
	}
	if repo.created.WorkspaceID != workspaceID {
		t.Fatalf("create workspace = %s, want %s", repo.created.WorkspaceID, workspaceID)
	}
	if err := service.UpdateLink(context.Background(), linkID, workspaceID, CoreUpdateLink{}); err != nil {
		t.Fatalf("update link: %v", err)
	}
	if repo.updatedLinkID != linkID || repo.workspaceID != workspaceID {
		t.Fatalf("update target = %s/%s, want %s/%s", repo.updatedLinkID, repo.workspaceID, linkID, workspaceID)
	}
	if err := service.DeleteLink(context.Background(), linkID, workspaceID); err != nil {
		t.Fatalf("delete link: %v", err)
	}
	if repo.deletedLinkID != linkID || repo.workspaceID != workspaceID {
		t.Fatalf("delete target = %s/%s, want %s/%s", repo.deletedLinkID, repo.workspaceID, linkID, workspaceID)
	}
}
