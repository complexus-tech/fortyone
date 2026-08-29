package comments

import (
	"context"
	"errors"
	"testing"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type fixedEventIDs struct{ id uuid.UUID }

func (generator fixedEventIDs) New() uuid.UUID { return generator.id }

type serviceRepositoryFake struct {
	create func(context.Context, commentsdomain.CreateCommand, uuid.UUID) (commentsdomain.Comment, error)
	update func(context.Context, commentsdomain.UpdateCommand, uuid.UUID) error
	delete func(context.Context, commentsdomain.ActorScope, uuid.UUID) error
	get    func(context.Context, commentsdomain.GetQuery) (commentsdomain.Comment, error)
}

func (fake serviceRepositoryFake) CreateComment(ctx context.Context, command commentsdomain.CreateCommand, eventID uuid.UUID) (commentsdomain.Comment, error) {
	if fake.create == nil {
		panic("unexpected CreateComment call")
	}
	return fake.create(ctx, command, eventID)
}

func (fake serviceRepositoryFake) UpdateComment(ctx context.Context, command commentsdomain.UpdateCommand, eventID uuid.UUID) error {
	if fake.update == nil {
		panic("unexpected UpdateComment call")
	}
	return fake.update(ctx, command, eventID)
}

func (fake serviceRepositoryFake) DeleteComment(ctx context.Context, scope commentsdomain.ActorScope, eventID uuid.UUID) error {
	if fake.delete == nil {
		panic("unexpected DeleteComment call")
	}
	return fake.delete(ctx, scope, eventID)
}

func (fake serviceRepositoryFake) GetComment(ctx context.Context, query commentsdomain.GetQuery) (commentsdomain.Comment, error) {
	if fake.get == nil {
		panic("unexpected GetComment call")
	}
	return fake.get(ctx, query)
}

func TestCreateCommentValidatesAndThreadsTypedIntent(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	storyID := uuid.New()
	parentID := uuid.New()
	authorID := uuid.New()
	mentionID := uuid.New()
	eventID := uuid.New()
	actor := commentTestActor(t, authorID, workspaceID)
	repository := serviceRepositoryFake{
		create: func(_ context.Context, command commentsdomain.CreateCommand, gotEventID uuid.UUID) (commentsdomain.Comment, error) {
			if gotEventID != eventID || command.WorkspaceID != workspaceID || command.StoryID != storyID || command.Actor.PrincipalID != authorID {
				t.Fatalf("create intent = %#v, event=%s", command, gotEventID)
			}
			if command.ParentID == nil || *command.ParentID != parentID || len(command.MentionedUserIDs) != 1 || command.MentionedUserIDs[0] != mentionID {
				t.Fatalf("normalized create input = %#v", command)
			}
			return commentsdomain.Comment{ID: uuid.New(), StoryID: storyID, UserID: authorID, Comment: command.Content, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}
	service := newService(repository, fixedEventIDs{id: eventID})

	created, err := service.CreateComment(t.Context(), CreateCommentCommand{
		WorkspaceID: workspaceID, StoryID: storyID, ParentID: &parentID,
		Actor: actor, Content: "Created", MentionedUserIDs: []uuid.UUID{mentionID, mentionID},
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if created.StoryID != storyID || created.Comment != "Created" {
		t.Fatalf("created comment = %#v", created)
	}
}

func TestUpdateAndDeleteUseStableCallerGeneratedEventIDs(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	commentID := uuid.New()
	authorID := uuid.New()
	mentionID := uuid.New()
	eventID := uuid.New()
	actor := commentTestActor(t, authorID, workspaceID)
	repository := serviceRepositoryFake{
		update: func(_ context.Context, command commentsdomain.UpdateCommand, gotEventID uuid.UUID) error {
			if gotEventID != eventID || command.Scope.CommentID != commentID || command.Scope.Actor.PrincipalID != authorID {
				t.Fatalf("update intent = %#v, event=%s", command, gotEventID)
			}
			if len(command.MentionedUserIDs) != 1 || command.MentionedUserIDs[0] != mentionID {
				t.Fatalf("deduplicated mentions = %v", command.MentionedUserIDs)
			}
			return nil
		},
		delete: func(_ context.Context, scope commentsdomain.ActorScope, gotEventID uuid.UUID) error {
			if gotEventID != eventID || scope.CommentID != commentID || scope.WorkspaceID != workspaceID || scope.Actor.PrincipalID != authorID {
				t.Fatalf("delete intent = %#v, event=%s", scope, gotEventID)
			}
			return nil
		},
	}
	service := newService(repository, fixedEventIDs{id: eventID})
	scope := AuthorScope{CommentID: commentID, WorkspaceID: workspaceID, Actor: actor}
	if err := service.UpdateComment(t.Context(), UpdateCommentCommand{
		Scope: scope, Content: "Updated", MentionedUserIDs: []uuid.UUID{mentionID, mentionID},
	}); err != nil {
		t.Fatalf("update comment: %v", err)
	}
	if err := service.DeleteComment(t.Context(), scope); err != nil {
		t.Fatalf("delete comment: %v", err)
	}
}

func TestMutationsFailClosedBeforePersistence(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	validActor := commentTestActor(t, uuid.New(), workspaceID)
	readOnly, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalOAuthUser, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeCommentsRead),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("create read-only actor: %v", err)
	}
	readOnly, err = readOnly.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind read-only actor: %v", err)
	}
	service := newService(serviceRepositoryFake{}, fixedEventIDs{id: uuid.New()})

	tests := []struct {
		name    string
		command UpdateCommentCommand
	}{
		{name: "blank content", command: UpdateCommentCommand{Scope: AuthorScope{CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: validActor}, Content: "  "}},
		{name: "wrong workspace", command: UpdateCommentCommand{Scope: AuthorScope{CommentID: uuid.New(), WorkspaceID: uuid.New(), Actor: validActor}, Content: "content"}},
		{name: "missing write scope", command: UpdateCommentCommand{Scope: AuthorScope{CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: readOnly}, Content: "content"}},
		{name: "zero mention", command: UpdateCommentCommand{Scope: AuthorScope{CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: validActor}, Content: "content", MentionedUserIDs: []uuid.UUID{uuid.Nil}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := service.UpdateComment(t.Context(), test.command); err == nil {
				t.Fatal("UpdateComment() error = nil")
			}
		})
	}
}

func TestRepositoryErrorsRemainDiscoverable(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("database unavailable")
	workspaceID := uuid.New()
	actor := commentTestActor(t, uuid.New(), workspaceID)
	service := newService(serviceRepositoryFake{
		delete: func(context.Context, commentsdomain.ActorScope, uuid.UUID) error { return wantErr },
	}, fixedEventIDs{id: uuid.New()})

	err := service.DeleteComment(t.Context(), AuthorScope{
		CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: actor,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("delete error = %v, want wrapped database error", err)
	}
}

func commentTestActor(t *testing.T, userID, workspaceID uuid.UUID) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor workspace: %v", err)
	}
	return actor
}
