package stories

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type scopedSubresourceRepo struct {
	Repository
	story           CoreSingleStory
	getErr          error
	commentErr      error
	workspaceID     uuid.UUID
	commentsCalls   int
	activitiesCalls int
	linksCalls      int
}

type scopedCommentCreator struct{ calls int }

func (creator *scopedCommentCreator) CreateComment(_ context.Context, command CreateCommentCommand) (CoreComment, error) {
	creator.calls++
	return CoreComment{StoryID: command.StoryID}, nil
}

func (r *scopedSubresourceRepo) Get(_ context.Context, _ uuid.UUID, workspaceID uuid.UUID) (CoreSingleStory, error) {
	r.workspaceID = workspaceID
	return r.story, r.getErr
}

func (r *scopedSubresourceRepo) GetComments(_ context.Context, _, workspaceID uuid.UUID, _, _ int) ([]CoreComment, bool, error) {
	r.workspaceID = workspaceID
	r.commentsCalls++
	return nil, false, nil
}

func (r *scopedSubresourceRepo) GetActivitiesWithUser(_ context.Context, _, workspaceID uuid.UUID, _, _ int) ([]CoreActivityWithUser, bool, error) {
	r.workspaceID = workspaceID
	r.activitiesCalls++
	return nil, false, nil
}

func (r *scopedSubresourceRepo) GetStoryLinks(_ context.Context, _, workspaceID uuid.UUID) ([]storydomain.StoryLink, error) {
	r.workspaceID = workspaceID
	r.linksCalls++
	return nil, nil
}

func (r *scopedSubresourceRepo) GetComment(_ context.Context, _, _, workspaceID uuid.UUID) (CoreComment, error) {
	r.workspaceID = workspaceID
	return CoreComment{}, r.commentErr
}

func newScopedSubresourceService(repo Repository) *Service {
	return New(logger.NewWithText(io.Discard, slog.LevelError, "test"), repo, nil, nil)
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
	creator := &scopedCommentCreator{}
	service.ConfigureCommentCreator(creator)
	ctx := platformauth.SetUserID(context.Background(), uuid.New())

	_, err := service.CreateComment(ctx, workspaceID, CoreNewComment{
		StoryID: storyID,
		Parent:  &parentID,
		UserID:  uuid.New(),
		Comment: "reply",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create reply error = %v, want not found", err)
	}
	if creator.calls != 0 {
		t.Fatal("invalid parent relation reached delegated comment insertion")
	}
}
