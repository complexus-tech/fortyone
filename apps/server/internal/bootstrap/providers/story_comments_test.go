package providers

import (
	"context"
	"testing"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type storyCommentStoreStub struct {
	command commentsdomain.CreateCommand
	result  commentsdomain.Comment
}

func (store *storyCommentStoreStub) CreateComment(
	_ context.Context,
	command commentsdomain.CreateCommand,
	_ uuid.UUID,
) (commentsdomain.Comment, error) {
	store.command = command
	return store.result, nil
}

func (*storyCommentStoreStub) UpdateComment(context.Context, commentsdomain.UpdateCommand, uuid.UUID) error {
	return nil
}

func (*storyCommentStoreStub) DeleteComment(context.Context, commentsdomain.ActorScope, uuid.UUID) error {
	return nil
}

func (*storyCommentStoreStub) GetComment(context.Context, commentsdomain.GetQuery) (commentsdomain.Comment, error) {
	return commentsdomain.Comment{}, nil
}

func TestStoryCommentCreatorMapsCallerOwnedContracts(t *testing.T) {
	t.Parallel()

	workspaceID, storyID, actorID := uuid.New(), uuid.New(), uuid.New()
	parentID, commentID, replyID := uuid.New(), uuid.New(), uuid.New()
	mentionedIDs := []uuid.UUID{uuid.New(), uuid.New()}
	now := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind comment actor: %v", err)
	}
	store := &storyCommentStoreStub{result: commentsdomain.Comment{
		ID: commentID, StoryID: storyID, Parent: &parentID, UserID: actorID,
		Comment: "Created through the adapter", CreatedAt: now, UpdatedAt: now,
		SubComments: []commentsdomain.Comment{{
			ID: replyID, StoryID: storyID, Parent: &commentID, UserID: actorID,
			Comment: "Nested", CreatedAt: now, UpdatedAt: now,
		}},
	}}
	adapter, err := NewStoryCommentCreator(comments.New(store))
	if err != nil {
		t.Fatalf("construct story comment adapter: %v", err)
	}

	created, err := adapter.CreateComment(t.Context(), stories.CreateCommentCommand{
		WorkspaceID: workspaceID, StoryID: storyID, ParentID: &parentID,
		Actor: actor, Content: "Created through the adapter", MentionedUserIDs: mentionedIDs,
	})
	if err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if store.command.WorkspaceID != workspaceID || store.command.StoryID != storyID ||
		store.command.ParentID == nil || *store.command.ParentID != parentID || store.command.Actor.PrincipalID != actorID {
		t.Fatalf("mapped comment command = %#v", store.command)
	}
	if len(store.command.MentionedUserIDs) != len(mentionedIDs) || &store.command.MentionedUserIDs[0] == &mentionedIDs[0] {
		t.Fatalf("mention IDs were not defensively copied: %v", store.command.MentionedUserIDs)
	}
	if created.ID != commentID || created.Parent == nil || *created.Parent != parentID ||
		len(created.SubComments) != 1 || created.SubComments[0].ID != replyID {
		t.Fatalf("mapped story comment = %#v", created)
	}

	store.result.SubComments[0].Comment = "mutated"
	if created.SubComments[0].Comment == "mutated" {
		t.Fatal("adapter retained the comments service child slice")
	}
}
