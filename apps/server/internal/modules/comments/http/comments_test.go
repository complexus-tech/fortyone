package commentshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace/noop"
)

type commentServiceFake struct {
	update func(context.Context, comments.UpdateCommentCommand) error
	delete func(context.Context, comments.AuthorScope) error
}

func (fake commentServiceFake) UpdateComment(ctx context.Context, command comments.UpdateCommentCommand) error {
	if fake.update == nil {
		panic("unexpected UpdateComment call")
	}
	return fake.update(ctx, command)
}

func (fake commentServiceFake) DeleteComment(ctx context.Context, scope comments.AuthorScope) error {
	if fake.delete == nil {
		panic("unexpected DeleteComment call")
	}
	return fake.delete(ctx, scope)
}

func TestUpdateCommentThreadsAuthenticatedScope(t *testing.T) {
	t.Parallel()

	commentID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	mentionedUserID := uuid.New()
	handler := New(commentServiceFake{
		update: func(_ context.Context, command comments.UpdateCommentCommand) error {
			if command.Scope.CommentID != commentID || command.Scope.WorkspaceID != workspaceID || command.Scope.Actor.PrincipalID != actorID {
				t.Fatalf("scope = %#v", command.Scope)
			}
			if command.Content != "Updated" || len(command.MentionedUserIDs) != 1 || command.MentionedUserIDs[0] != mentionedUserID {
				t.Fatalf("payload = %q/%v", command.Content, command.MentionedUserIDs)
			}
			return nil
		},
	})
	setHandlerScope(handler, workspaceID, actorID)
	app := newCommentHTTPTestApp()
	app.Put("/workspaces/{workspaceSlug}/comments/{id}", handler.UpdateComment)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/workspaces/acme/comments/"+commentID.String(),
		strings.NewReader(`{"content":"Updated","mentions":["`+mentionedUserID.String()+`"]}`),
	)
	request.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestDeleteCommentThreadsAuthenticatedScope(t *testing.T) {
	t.Parallel()

	commentID := uuid.New()
	workspaceID := uuid.New()
	actorID := uuid.New()
	handler := New(commentServiceFake{
		delete: func(_ context.Context, scope comments.AuthorScope) error {
			if scope.CommentID != commentID || scope.WorkspaceID != workspaceID || scope.Actor.PrincipalID != actorID {
				t.Fatalf("scope = %#v", scope)
			}
			return nil
		},
	})
	setHandlerScope(handler, workspaceID, actorID)
	app := newCommentHTTPTestApp()
	app.Delete("/workspaces/{workspaceSlug}/comments/{id}", handler.DeleteComment)

	recorder := httptest.NewRecorder()
	app.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodDelete, "/workspaces/acme/comments/"+commentID.String(), nil),
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestCommentMutationErrorsUseNonEnumeratingStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "hidden target", err: comments.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid mention", err: comments.ErrInvalidMention, wantStatus: http.StatusBadRequest},
		{name: "database failure", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := commentMutationStatus(test.err); got != test.wantStatus {
				t.Fatalf("status = %d, want %d", got, test.wantStatus)
			}
		})
	}
}

func TestUpdateCommentHidesInaccessibleTargetAsNotFound(t *testing.T) {
	t.Parallel()

	handler := New(commentServiceFake{
		update: func(context.Context, comments.UpdateCommentCommand) error {
			return comments.ErrNotFound
		},
	})
	setHandlerScope(handler, uuid.New(), uuid.New())
	app := newCommentHTTPTestApp()
	app.Put("/workspaces/{workspaceSlug}/comments/{id}", handler.UpdateComment)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/workspaces/acme/comments/"+uuid.NewString(),
		strings.NewReader(`{"content":"Updated","mentions":[]}`),
	)
	request.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"message":"comment not found"`) {
		t.Fatalf("response did not use the generic not-found error: %s", recorder.Body.String())
	}
}

func TestUpdateCommentValidatesBoundedMentionInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input UpdateComment
	}{
		{name: "missing content", input: UpdateComment{}},
		{name: "nil mention", input: UpdateComment{Content: "Comment", Mentions: []uuid.UUID{uuid.Nil}}},
		{name: "too many mentions", input: UpdateComment{Content: "Comment", Mentions: make([]uuid.UUID, maxCommentMentions+1)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.input.Validate(); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func setHandlerScope(handler *Handlers, workspaceID, actorID uuid.UUID) {
	handler.workspaceID = func(context.Context) (uuid.UUID, error) { return workspaceID, nil }
	handler.actor = func(context.Context) (platformauth.Actor, error) {
		return platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	}
}

func newCommentHTTPTestApp() *web.App {
	return web.New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("comments-http-test"))
}
