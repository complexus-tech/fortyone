package commentshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type collaborationServiceStub struct {
	collaborationService
	reaction func(context.Context, comments.ReactionCommand) (comments.CommentState, error)
	read     func(context.Context, comments.ReadCommand) (comments.StoryState, error)
}

func (stub collaborationServiceStub) SetReaction(ctx context.Context, command comments.ReactionCommand) (comments.CommentState, error) {
	return stub.reaction(ctx, command)
}

func (stub collaborationServiceStub) MarkRead(ctx context.Context, command comments.ReadCommand) (comments.StoryState, error) {
	return stub.read(ctx, command)
}

func TestReactionHTTPUsesAuthenticatedActorAndFiniteIntent(t *testing.T) {
	t.Parallel()
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			workspaceID, actorID, commentID := uuid.New(), uuid.New(), uuid.New()
			called := false
			handler := newCollaborationHandlers(collaborationServiceStub{
				reaction: func(_ context.Context, command comments.ReactionCommand) (comments.CommentState, error) {
					called = true
					require.Equal(t, workspaceID, command.Scope.WorkspaceID)
					require.Equal(t, actorID, command.Scope.Actor.PrincipalID)
					require.Equal(t, commentID, command.Scope.CommentID)
					require.Equal(t, "👍", command.Emoji)
					require.Equal(t, method == http.MethodPut, command.Active)
					return comments.CommentState{CommentID: commentID}, nil
				},
			})
			setCollaborationHandlerScope(t, handler, workspaceID, actorID)
			app := newCommentHTTPTestApp()
			app.Put("/comments/{id}/reactions", handler.SetReaction)
			app.Delete("/comments/{id}/reactions", handler.SetReaction)
			request := httptest.NewRequest(method, "/comments/"+commentID.String()+"/reactions", strings.NewReader(`{"emoji":"👍"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.True(t, called)
		})
	}
}

func TestReadHTTPPreservesObservedCutoff(t *testing.T) {
	t.Parallel()
	workspaceID, actorID, storyID := uuid.New(), uuid.New(), uuid.New()
	cutoff := time.Date(2026, time.October, 16, 12, 0, 0, 123000, time.UTC)
	handler := newCollaborationHandlers(collaborationServiceStub{
		read: func(_ context.Context, command comments.ReadCommand) (comments.StoryState, error) {
			require.Equal(t, workspaceID, command.Scope.WorkspaceID)
			require.Equal(t, actorID, command.Scope.Actor.PrincipalID)
			require.Equal(t, storyID, command.Scope.StoryID)
			require.Equal(t, cutoff, command.ObservedAt)
			return comments.StoryState{ObservedAt: cutoff}, nil
		},
	})
	setCollaborationHandlerScope(t, handler, workspaceID, actorID)
	app := newCommentHTTPTestApp()
	app.Put("/stories/{storyId}/comments/read", handler.MarkRead)
	request := httptest.NewRequest(http.MethodPut, "/stories/"+storyID.String()+"/comments/read",
		strings.NewReader(`{"observedAt":"`+cutoff.Format(time.RFC3339Nano)+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
}

func TestResolutionHTTPRequiresExplicitResolvedFlag(t *testing.T) {
	t.Parallel()
	handler := newCollaborationHandlers(collaborationServiceStub{})
	setCollaborationHandlerScope(t, handler, uuid.New(), uuid.New())
	app := newCommentHTTPTestApp()
	app.Put("/comments/{id}/resolution", handler.SetResolution)
	request := httptest.NewRequest(http.MethodPut, "/comments/"+uuid.NewString()+"/resolution", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func setCollaborationHandlerScope(t *testing.T, handler *collaborationHandlers, workspaceID, actorID uuid.UUID) {
	t.Helper()
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	handler.workspaceID = func(context.Context) (uuid.UUID, error) { return workspaceID, nil }
	handler.actor = func(context.Context) (platformauth.Actor, error) { return actor, nil }
}
