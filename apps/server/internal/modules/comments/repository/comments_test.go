package commentsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeCommentQueries struct {
	appendEvent    func(context.Context, commentsql.AppendCommentMutationEventParams) ([]uuid.UUID, error)
	create         func(context.Context, commentsql.CreateCommentForActorParams) (commentsql.CreateCommentForActorRow, error)
	deleteComment  func(context.Context, commentsql.DeleteCommentForAuthorParams) (commentsql.DeleteCommentForAuthorRow, error)
	deleteMentions func(context.Context, commentsql.DeleteCommentMentionsForAuthorParams) (commentsql.DeleteCommentMentionsForAuthorRow, error)
	get            func(context.Context, commentsql.GetCommentForWorkspaceParams) (commentsql.GetCommentForWorkspaceRow, error)
	insertMentions func(context.Context, commentsql.InsertCommentMentionsForAuthorParams) (int64, error)
	update         func(context.Context, commentsql.UpdateCommentForAuthorParams) (commentsql.UpdateCommentForAuthorRow, error)
}

func (fake fakeCommentQueries) AppendCommentMutationEvent(ctx context.Context, params commentsql.AppendCommentMutationEventParams) ([]uuid.UUID, error) {
	if fake.appendEvent == nil {
		panic("unexpected AppendCommentMutationEvent call")
	}
	return fake.appendEvent(ctx, params)
}

func (fake fakeCommentQueries) CreateCommentForActor(ctx context.Context, params commentsql.CreateCommentForActorParams) (commentsql.CreateCommentForActorRow, error) {
	if fake.create == nil {
		panic("unexpected CreateCommentForActor call")
	}
	return fake.create(ctx, params)
}

func (fake fakeCommentQueries) DeleteCommentForAuthor(ctx context.Context, params commentsql.DeleteCommentForAuthorParams) (commentsql.DeleteCommentForAuthorRow, error) {
	if fake.deleteComment == nil {
		panic("unexpected DeleteCommentForAuthor call")
	}
	return fake.deleteComment(ctx, params)
}

func (fake fakeCommentQueries) DeleteCommentMentionsForAuthor(ctx context.Context, params commentsql.DeleteCommentMentionsForAuthorParams) (commentsql.DeleteCommentMentionsForAuthorRow, error) {
	if fake.deleteMentions == nil {
		panic("unexpected DeleteCommentMentionsForAuthor call")
	}
	return fake.deleteMentions(ctx, params)
}

func (fake fakeCommentQueries) GetCommentForWorkspace(ctx context.Context, params commentsql.GetCommentForWorkspaceParams) (commentsql.GetCommentForWorkspaceRow, error) {
	if fake.get == nil {
		panic("unexpected GetCommentForWorkspace call")
	}
	return fake.get(ctx, params)
}

func (fake fakeCommentQueries) InsertCommentMentionsForAuthor(ctx context.Context, params commentsql.InsertCommentMentionsForAuthorParams) (int64, error) {
	if fake.insertMentions == nil {
		panic("unexpected InsertCommentMentionsForAuthor call")
	}
	return fake.insertMentions(ctx, params)
}

func (fake fakeCommentQueries) UpdateCommentForAuthor(ctx context.Context, params commentsql.UpdateCommentForAuthorParams) (commentsql.UpdateCommentForAuthorRow, error) {
	if fake.update == nil {
		panic("unexpected UpdateCommentForAuthor call")
	}
	return fake.update(ctx, params)
}

func TestCreateCommentWritesStateMentionsAndMinimalEventInOneUnit(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	storyID := uuid.New()
	commentID := uuid.New()
	userID := uuid.New()
	mentionID := uuid.New()
	eventID := uuid.New()
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	actor := repositoryTestActor(t, userID, workspaceID)
	queries := fakeCommentQueries{
		create: func(_ context.Context, params commentsql.CreateCommentForActorParams) (commentsql.CreateCommentForActorRow, error) {
			if params.ActorID != userID || params.WorkspaceID != workspaceID || params.StoryID != storyID || !params.TeamAccessUnrestricted {
				t.Fatalf("create params = %#v", params)
			}
			return commentsql.CreateCommentForActorRow{CommentID: commentID, StoryID: storyID, CommenterID: userID, Content: params.Content, CreatedAt: now, UpdatedAt: now}, nil
		},
		deleteMentions: func(_ context.Context, params commentsql.DeleteCommentMentionsForAuthorParams) (commentsql.DeleteCommentMentionsForAuthorRow, error) {
			return commentsql.DeleteCommentMentionsForAuthorRow{CommentFound: params.CommentID == commentID}, nil
		},
		insertMentions: func(_ context.Context, params commentsql.InsertCommentMentionsForAuthorParams) (int64, error) {
			if len(params.MentionedUserIds) != 1 || params.MentionedUserIds[0] != mentionID {
				t.Fatalf("mention params = %#v", params)
			}
			return 1, nil
		},
		appendEvent: func(_ context.Context, params commentsql.AppendCommentMutationEventParams) ([]uuid.UUID, error) {
			if params.EventID != eventID || params.EventType != string(commentsdomain.EventCreated) || params.CommentID != commentID {
				t.Fatalf("event params = %#v", params)
			}
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(params.Payload, &payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if len(payload) != 3 || payload["comment_id"] == nil || payload["story_id"] == nil || payload["parent_id"] == nil {
				t.Fatalf("event payload fields = %v", payload)
			}
			if string(params.PayloadBody) == "" || params.ActorID != userID || params.ActorCredentialID != nil {
				t.Fatalf("event envelope/actor = %q/%#v", params.PayloadBody, params)
			}
			return []uuid.UUID{uuid.New()}, nil
		},
	}
	repository := newCommentTestRepository(queries)

	created, err := repository.CreateComment(t.Context(), commentsdomain.CreateCommand{
		WorkspaceID: workspaceID, StoryID: storyID, Actor: actor,
		Content: "Typed comment", MentionedUserIDs: []uuid.UUID{mentionID},
	}, eventID)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if created.ID != commentID || created.SubComments == nil {
		t.Fatalf("created comment = %#v", created)
	}
}

func TestUpdateRollsBackIntentWhenMentionReplacementFails(t *testing.T) {
	t.Parallel()

	wantErr := commentsdomain.ErrInvalidMention
	workspaceID := uuid.New()
	userID := uuid.New()
	commentID := uuid.New()
	now := time.Now().UTC()
	queries := fakeCommentQueries{
		update: func(context.Context, commentsql.UpdateCommentForAuthorParams) (commentsql.UpdateCommentForAuthorRow, error) {
			return commentsql.UpdateCommentForAuthorRow{CommentID: commentID, StoryID: uuid.New(), CommenterID: userID, Content: "updated", CreatedAt: now, UpdatedAt: now}, nil
		},
		deleteMentions: func(context.Context, commentsql.DeleteCommentMentionsForAuthorParams) (commentsql.DeleteCommentMentionsForAuthorRow, error) {
			return commentsql.DeleteCommentMentionsForAuthorRow{CommentFound: true}, nil
		},
		insertMentions: func(context.Context, commentsql.InsertCommentMentionsForAuthorParams) (int64, error) {
			return 0, nil
		},
	}
	repository := newCommentTestRepository(queries)
	repository.runTransaction = func(ctx context.Context, operation func(commentsql.Querier) error) error {
		return operation(queries)
	}

	err := repository.UpdateComment(t.Context(), commentsdomain.UpdateCommand{
		Scope:   commentsdomain.ActorScope{CommentID: commentID, WorkspaceID: workspaceID, Actor: repositoryTestActor(t, userID, workspaceID)},
		Content: "updated", MentionedUserIDs: []uuid.UUID{uuid.New()},
	}, uuid.New())
	if !errors.Is(err, wantErr) {
		t.Fatalf("update error = %v, want %v", err, wantErr)
	}
}

func TestHiddenMutationsMapToNotFound(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	repository := newCommentTestRepository(fakeCommentQueries{
		deleteComment: func(context.Context, commentsql.DeleteCommentForAuthorParams) (commentsql.DeleteCommentForAuthorRow, error) {
			return commentsql.DeleteCommentForAuthorRow{}, pgx.ErrNoRows
		},
	})
	err := repository.DeleteComment(t.Context(), commentsdomain.ActorScope{
		CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: repositoryTestActor(t, userID, workspaceID),
	}, uuid.New())
	if !errors.Is(err, commentsdomain.ErrNotFound) {
		t.Fatalf("delete error = %v, want not found", err)
	}
}

func TestGetCommentMapsTypedScopeAndHiddenRows(t *testing.T) {
	t.Parallel()

	commentID := uuid.New()
	workspaceID := uuid.New()
	repository := newCommentTestRepository(fakeCommentQueries{
		get: func(_ context.Context, params commentsql.GetCommentForWorkspaceParams) (commentsql.GetCommentForWorkspaceRow, error) {
			if params.CommentID != commentID || params.WorkspaceID != workspaceID {
				t.Fatalf("get params = %#v", params)
			}
			return commentsql.GetCommentForWorkspaceRow{}, pgx.ErrNoRows
		},
	})
	_, err := repository.GetComment(t.Context(), commentsdomain.GetQuery{CommentID: commentID, WorkspaceID: workspaceID})
	if !errors.Is(err, commentsdomain.ErrNotFound) {
		t.Fatalf("get error = %v, want not found", err)
	}
}

func newCommentTestRepository(queries commentsql.Querier) *Repository {
	return newWithQueries(
		logger.NewWithText(io.Discard, slog.LevelError, "comments-repository-test"),
		queries,
	)
}

func repositoryTestActor(t *testing.T, userID, workspaceID uuid.UUID) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewHumanActor(userID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor: %v", err)
	}
	return actor
}
