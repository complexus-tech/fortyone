package commentshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type seenServiceStub struct {
	collaborationService
	seen func(context.Context, comments.StoryScope, uuid.UUID) error
}

func (stub seenServiceStub) MarkSeen(ctx context.Context, scope comments.StoryScope, commentID uuid.UUID) error {
	return stub.seen(ctx, scope, commentID)
}

func TestMarkSeenHTTPUsesAuthenticatedScopeAndSupportsEmptyReplay(t *testing.T) {
	t.Parallel()
	workspaceID, actorID, storyID, commentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	calls := 0
	handler := newCollaborationHandlers(seenServiceStub{
		seen: func(_ context.Context, scope comments.StoryScope, id uuid.UUID) error {
			calls++
			require.Equal(t, workspaceID, scope.WorkspaceID)
			require.Equal(t, workspaceID, scope.Actor.WorkspaceID)
			require.Equal(t, actorID, scope.Actor.PrincipalID)
			require.Equal(t, storyID, scope.StoryID)
			require.Equal(t, commentID, id)
			return nil
		},
	})
	setCollaborationHandlerScope(t, handler, workspaceID, actorID)
	app := newCommentHTTPTestApp()
	app.Put("/workspaces/{workspaceSlug}/stories/{storyId}/comments/{commentId}/seen", handler.MarkSeen)
	path := "/workspaces/acme/stories/" + storyID.String() + "/comments/" + commentID.String() + "/seen"
	for range 2 {
		response := httptest.NewRecorder()
		app.ServeHTTP(response, httptest.NewRequest(http.MethodPut, path, nil))
		require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
		require.Empty(t, response.Body.String())
	}
	// Unused body/query fields cannot override the authenticated actor or route.
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, path+"?commentId="+uuid.NewString(), strings.NewReader(
		`{"workspaceId":"`+uuid.NewString()+`","actorId":"`+uuid.NewString()+`","storyId":"`+uuid.NewString()+`","commentId":"`+uuid.NewString()+`"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(response, request)
	require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
	require.Equal(t, 3, calls)
}

func TestMarkSeenHTTPRejectsMalformedTargetsBeforeService(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"/stories/not-a-uuid/comments/" + uuid.NewString() + "/seen",
		"/stories/" + uuid.NewString() + "/comments/not-a-uuid/seen",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			handler := newCollaborationHandlers(seenServiceStub{seen: func(context.Context, comments.StoryScope, uuid.UUID) error {
				t.Fatal("invalid target reached service")
				return nil
			}})
			setCollaborationHandlerScope(t, handler, uuid.New(), uuid.New())
			app := newCommentHTTPTestApp()
			app.Put("/stories/{storyId}/comments/{commentId}/seen", handler.MarkSeen)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, httptest.NewRequest(http.MethodPut, path, nil))
			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestMarkSeenHTTPRequiresAuthenticatedWorkspace(t *testing.T) {
	t.Parallel()
	for _, missing := range []string{"workspace", "actor"} {
		t.Run(missing, func(t *testing.T) {
			t.Parallel()
			handler := newCollaborationHandlers(seenServiceStub{seen: func(context.Context, comments.StoryScope, uuid.UUID) error {
				t.Fatal("unauthenticated request reached service")
				return nil
			}})
			setCollaborationHandlerScope(t, handler, uuid.New(), uuid.New())
			if missing == "workspace" {
				handler.workspaceID = func(context.Context) (uuid.UUID, error) { return uuid.Nil, errors.New("workspace unavailable") }
			} else {
				handler.actor = func(context.Context) (platformauth.Actor, error) {
					return platformauth.Actor{}, platformauth.ErrActorNotFound
				}
			}
			app := newCommentHTTPTestApp()
			app.Put("/stories/{storyId}/comments/{commentId}/seen", handler.MarkSeen)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/stories/"+uuid.NewString()+"/comments/"+uuid.NewString()+"/seen", nil))
			require.Equal(t, http.StatusUnauthorized, response.Code)
		})
	}
}

func TestMarkSeenHTTPPreservesAccessAndPersistenceErrors(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "hidden or wrong-story comment", err: comments.ErrNotFound, status: http.StatusNotFound},
		{name: "forbidden actor", err: comments.ErrForbidden, status: http.StatusForbidden},
		{name: "invalid target", err: comments.ErrInvalidComment, status: http.StatusBadRequest},
		{name: "persistence failure", err: errors.New("database unavailable"), status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			handler := newCollaborationHandlers(seenServiceStub{seen: func(context.Context, comments.StoryScope, uuid.UUID) error { return test.err }})
			setCollaborationHandlerScope(t, handler, uuid.New(), uuid.New())
			app := newCommentHTTPTestApp()
			app.Put("/stories/{storyId}/comments/{commentId}/seen", handler.MarkSeen)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/stories/"+uuid.NewString()+"/comments/"+uuid.NewString()+"/seen", nil))
			require.Equal(t, test.status, response.Code, response.Body.String())
		})
	}
}
