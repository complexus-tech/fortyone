package notificationshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/google/uuid"
)

func TestNotificationReadHTTPCarriesObservedVersion(t *testing.T) {
	t.Parallel()
	actorID, workspaceID, notificationID := uuid.New(), uuid.New(), uuid.New()
	observed := time.Date(2026, time.October, 3, 20, 0, 0, 123456789, time.UTC)
	repository := &notificationHTTPRepositoryStub{}
	handler := newNotificationHTTPHandlers(repository)
	resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Slug: "workspace", UserRole: "member",
	}}
	wrapped := mid.Workspace(notificationHTTPLogger(), resolver)(handler.MarkAsRead)
	request := httptest.NewRequest(http.MethodPut,
		"/workspaces/workspace/notifications/"+notificationID.String()+"/read?"+
			url.Values{"observedCreatedAt": {observed.Format(time.RFC3339Nano)}}.Encode(), nil)
	request.SetPathValue("workspaceSlug", "workspace")
	request.SetPathValue("id", notificationID.String())
	recorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	mutation := repository.mutation
	if mutation.ExpectedCreatedAt == nil || !mutation.ExpectedCreatedAt.Equal(observed) ||
		mutation.Access.ActorID != actorID || mutation.Access.WorkspaceID != workspaceID || mutation.NotificationID != notificationID {
		t.Fatalf("repository mutation = %#v", mutation)
	}
	repository.mutateErr = notificationsdomain.ErrConflict
	conflict := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), conflict, request); err != nil {
		t.Fatalf("conflict MarkAsRead() error = %v", err)
	}
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d, body=%s", conflict.Code, conflict.Body.String())
	}
}

func TestNotificationReadHTTPRejectsInvalidVersionsBeforeMutation(t *testing.T) {
	t.Parallel()
	for name, values := range map[string]url.Values{
		"empty":     {"observedCreatedAt": {""}},
		"malformed": {"observedCreatedAt": {"not-a-timestamp"}},
		"repeated":  {"observedCreatedAt": {"2026-10-03T20:00:00Z", "2026-10-03T20:01:00Z"}},
		"zero":      {"observedCreatedAt": {"0001-01-01T00:00:00Z"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actorID, workspaceID, notificationID := uuid.New(), uuid.New(), uuid.New()
			repository := &notificationHTTPRepositoryStub{}
			handler := newNotificationHTTPHandlers(repository)
			resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{ID: workspaceID, Slug: "workspace", UserRole: "member"}}
			wrapped := mid.Workspace(notificationHTTPLogger(), resolver)(handler.MarkAsRead)
			request := httptest.NewRequest(http.MethodPut, "/workspaces/workspace/notifications/"+notificationID.String()+"/read?"+values.Encode(), nil)
			request.SetPathValue("workspaceSlug", "workspace")
			request.SetPathValue("id", notificationID.String())
			recorder := httptest.NewRecorder()
			if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
				t.Fatalf("MarkAsRead() error = %v", err)
			}
			if recorder.Code != http.StatusBadRequest || repository.mutation.NotificationID != uuid.Nil {
				t.Fatalf("status = %d, mutation=%#v, body=%s", recorder.Code, repository.mutation, recorder.Body.String())
			}
		})
	}
}
