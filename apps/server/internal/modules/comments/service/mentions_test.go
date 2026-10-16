package comments

import (
	"context"
	"errors"
	"testing"
	"time"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mentionEventPublisher struct {
	events []events.Event
	err    error
}

func (publisher *mentionEventPublisher) Publish(_ context.Context, event events.Event) error {
	publisher.events = append(publisher.events, event)
	return publisher.err
}

type mentionTestTransaction struct {
	pgx.Tx
	commitError error
}

func (tx *mentionTestTransaction) Commit(context.Context) error   { return tx.commitError }
func (tx *mentionTestTransaction) Rollback(context.Context) error { return nil }
func (tx *mentionTestTransaction) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return tx, nil
}

func TestEditedMentionsPublishOnlyAfterSuccessfulCommit(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		mutationErr error
		commitErr   error
		publishErr  error
		wantEvents  int
	}{
		{name: "committed", wantEvents: 1},
		{name: "publish failure leaves edit successful", publishErr: errors.New("stream unavailable"), wantEvents: 1},
		{name: "mutation rejected", mutationErr: commentsdomain.ErrInvalidMention},
		{name: "commit failed", commitErr: pgx.ErrTxCommitRollback},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspaceID, authorID, mentionID := uuid.New(), uuid.New(), uuid.New()
			comment := commentsdomain.Comment{
				ID: uuid.New(), StoryID: uuid.New(), UserID: authorID,
				Comment: "Updated comment", UpdatedAt: time.Now().UTC(),
			}
			publisher := &mentionEventPublisher{err: test.publishErr}
			service := New(serviceRepositoryFake{
				update: func(context.Context, commentsdomain.UpdateCommand, uuid.UUID) (commentsdomain.UpdateResult, error) {
					return commentsdomain.UpdateResult{
						Comment: comment, StoryTitle: "Scoped story",
						AddedMentionIDs: []uuid.UUID{mentionID, authorID, mentionID},
					}, test.mutationErr
				},
			}, WithEventPublisher(nil, publisher))
			tx := &mentionTestTransaction{commitError: test.commitErr}
			err := platformdatabase.WithinUnitOfWork(t.Context(), tx, func(ctx context.Context, _ pgx.Tx) error {
				err := service.UpdateComment(ctx, UpdateCommentCommand{
					Scope:   AuthorScope{CommentID: comment.ID, WorkspaceID: workspaceID, Actor: commentTestActor(t, authorID, workspaceID)},
					Content: comment.Comment, MentionedUserIDs: []uuid.UUID{mentionID, authorID},
				})
				if len(publisher.events) != 0 {
					t.Fatal("mention was published before the outer transaction committed")
				}
				return err
			})
			if want := errors.Join(test.mutationErr, test.commitErr); (want == nil) != (err == nil) {
				t.Fatalf("edit error = %v, want %v", err, want)
			}
			if len(publisher.events) != test.wantEvents {
				t.Fatalf("published %d events, want %d", len(publisher.events), test.wantEvents)
			}
			if test.wantEvents == 0 {
				return
			}
			event := publisher.events[0]
			payload, ok := event.Payload.(events.UserMentionedPayload)
			if !ok || event.Type != events.UserMentioned || event.ActorID != authorID ||
				payload.CommentID != comment.ID || payload.StoryID != comment.StoryID ||
				payload.WorkspaceID != workspaceID || payload.StoryTitle != "Scoped story" ||
				payload.MentionedUser != mentionID || payload.Content != comment.Comment ||
				!event.Timestamp.Equal(comment.UpdatedAt) {
				t.Fatalf("mention event = %#v", event)
			}
		})
	}
}
