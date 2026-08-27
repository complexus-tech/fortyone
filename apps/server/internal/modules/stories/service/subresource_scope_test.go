package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type scopedSubresourceRepo struct {
	Repository
	story              CoreSingleStory
	getErr             error
	commentErr         error
	workspaceID        uuid.UUID
	commentsCalls      int
	activitiesCalls    int
	linksCalls         int
	createCommentCalls int
}

func (r *scopedSubresourceRepo) Get(_ context.Context, _ uuid.UUID, workspaceID uuid.UUID) (CoreSingleStory, error) {
	r.workspaceID = workspaceID
	return r.story, r.getErr
}

func (r *scopedSubresourceRepo) GetComments(_ context.Context, _, workspaceID uuid.UUID, _, _ int) ([]comments.CoreComment, bool, error) {
	r.workspaceID = workspaceID
	r.commentsCalls++
	return nil, false, nil
}

func (r *scopedSubresourceRepo) GetActivitiesWithUser(_ context.Context, _, workspaceID uuid.UUID, _, _ int) ([]CoreActivityWithUser, bool, error) {
	r.workspaceID = workspaceID
	r.activitiesCalls++
	return nil, false, nil
}

func (r *scopedSubresourceRepo) GetStoryLinks(_ context.Context, _, workspaceID uuid.UUID) ([]links.CoreLink, error) {
	r.workspaceID = workspaceID
	r.linksCalls++
	return nil, nil
}

func (r *scopedSubresourceRepo) GetComment(_ context.Context, _, _, workspaceID uuid.UUID) (comments.CoreComment, error) {
	r.workspaceID = workspaceID
	return comments.CoreComment{}, r.commentErr
}

func (r *scopedSubresourceRepo) CreateComment(_ context.Context, comment CoreNewComment) (comments.CoreComment, error) {
	r.createCommentCalls++
	return comments.CoreComment{StoryID: comment.StoryID}, nil
}

func newScopedSubresourceService(repo Repository) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil, nil)
}

func TestStorySubresourceReadsRequireWorkspaceStory(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	repo := &scopedSubresourceRepo{getErr: ErrNotFound}
	service := newScopedSubresourceService(repo)

	if _, _, err := service.GetComments(context.Background(), storyID, workspaceID, 1, 20); !errors.Is(err, ErrNotFound) {
		t.Fatalf("comments error = %v, want not found", err)
	}
	if _, _, err := service.GetActivitiesWithUser(context.Background(), storyID, workspaceID, 1, 20); !errors.Is(err, ErrNotFound) {
		t.Fatalf("activities error = %v, want not found", err)
	}
	if _, err := service.GetStoryLinks(context.Background(), storyID, workspaceID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("links error = %v, want not found", err)
	}
	if repo.commentsCalls != 0 || repo.activitiesCalls != 0 || repo.linksCalls != 0 {
		t.Fatalf("unauthorized subresource query executed: comments=%d activities=%d links=%d", repo.commentsCalls, repo.activitiesCalls, repo.linksCalls)
	}
	if repo.workspaceID != workspaceID {
		t.Fatalf("workspace = %s, want %s", repo.workspaceID, workspaceID)
	}
}

func TestCreateCommentRejectsParentOutsideStoryBeforeInsert(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	parentID := uuid.New()
	repo := &scopedSubresourceRepo{
		story:      CoreSingleStory{ID: storyID, Workspace: workspaceID},
		commentErr: ErrNotFound,
	}
	service := newScopedSubresourceService(repo)

	_, err := service.CreateComment(context.Background(), workspaceID, CoreNewComment{
		StoryID: storyID,
		Parent:  &parentID,
		UserID:  uuid.New(),
		Comment: "reply",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create reply error = %v, want not found", err)
	}
	if repo.createCommentCalls != 0 {
		t.Fatal("invalid parent relation reached comment insertion")
	}
}
