package notificationshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestNotificationHTTPErrorStatusContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid", err: notifications.ErrInvalid, status: http.StatusBadRequest},
		{name: "forbidden", err: notifications.ErrForbidden, status: http.StatusForbidden},
		{name: "not found", err: notifications.ErrNotificationNotFound, status: http.StatusNotFound},
		{name: "conflict", err: notifications.ErrConflict, status: http.StatusConflict},
		{name: "internal", err: errors.New("database unavailable"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			if err := respondNotificationError(t.Context(), recorder, test.err); err != nil {
				t.Fatalf("respondNotificationError() error = %v", err)
			}
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, test.status, recorder.Body.String())
			}
		})
	}
}

func TestNotificationInboxHTTPBindsActorWorkspaceAndLegacyPagination(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	repository := &notificationHTTPRepositoryStub{}
	for range 3 {
		repository.listResult = append(repository.listResult, notificationsdomain.Notification{
			ID: uuid.New(), RecipientID: actorID, WorkspaceID: workspaceID,
			Type: notificationsdomain.NotificationTypeStoryUpdate, EntityType: notificationsdomain.EntityTypeStory,
			EntityID: uuid.New(), ActorID: uuid.New(), Title: "Story updated",
			Message:   notificationsdomain.NotificationMessage{Template: "updated", Variables: map[string]notificationsdomain.Variable{}},
			CreatedAt: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC),
		})
	}
	handler := newNotificationHTTPHandlers(repository)
	resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Name: "Workspace", Slug: "workspace", UserRole: "member",
	}}
	wrapped := mid.Workspace(notificationHTTPLogger(), resolver)(handler.List)
	request := httptest.NewRequest(http.MethodGet, "/workspaces/workspace/notifications?limit=2&offset=4&search=%20updated%20", nil)
	request.SetPathValue("workspaceSlug", "workspace")
	recorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("List() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.listQuery.Access.ActorID != actorID || repository.listQuery.Access.WorkspaceID != workspaceID ||
		repository.listQuery.Search != "updated" || repository.listQuery.Limit != 3 || repository.listQuery.Offset != 4 {
		t.Fatalf("repository list query = %#v", repository.listQuery)
	}
	var body struct {
		Data AppNotificationsResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(body.Data.Notifications) != 2 || !body.Data.Pagination.HasMore ||
		body.Data.Pagination.Page != 3 || body.Data.Pagination.PageSize != 2 || body.Data.Pagination.NextPage != 4 {
		t.Fatalf("notification response = %#v", body.Data)
	}
}

func TestNotificationListQueryParsingIsStrictBoundedAndValueSafe(t *testing.T) {
	t.Parallel()

	query, err := parseInboxListQuery(url.Values{
		"page":     {"9"},
		"pageSize": {"50"},
		"limit":    {"2"},
		"offset":   {"4"},
		"search":   {" updated "},
	})
	if err != nil {
		t.Fatalf("parseInboxListQuery() error = %v", err)
	}
	if query.Pagination.Page != 3 || query.Pagination.PageSize != 2 || query.Offset != 4 || query.Search != "updated" {
		t.Fatalf("parseInboxListQuery() = %#v", query)
	}

	portalQuery, err := parsePortalListQuery(url.Values{
		"page":       {"2"},
		"pageSize":   {"500"},
		"unreadOnly": {"true"},
	})
	if err != nil {
		t.Fatalf("parsePortalListQuery() error = %v", err)
	}
	if portalQuery.Pagination.Page != 2 || portalQuery.Pagination.PageSize != 100 || !portalQuery.UnreadOnly {
		t.Fatalf("parsePortalListQuery() = %#v", portalQuery)
	}

	for name, values := range map[string]url.Values{
		"repeated search":    {"search": {"first-sensitive", "second-sensitive"}},
		"oversized search":   {"search": {strings.Repeat("sensitive", 65)}},
		"invalid search":     {"search": {string([]byte{0xff})}},
		"nul search":         {"search": {"sensitive\x00search"}},
		"malformed page":     {"page": {"sensitive-page"}},
		"overflowing offset": {"offset": {"2147483648"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseInboxListQuery(values)
			if err == nil {
				t.Fatal("parseInboxListQuery() error = nil")
			}
			assertQueryErrorDoesNotExposeValues(t, err, values)
		})
	}

	for name, values := range map[string]url.Values{
		"repeated unread flag":  {"unreadOnly": {"true", "false"}},
		"malformed unread flag": {"unreadOnly": {"secret-value"}},
		"negative page":         {"page": {"-1"}},
		"overflowing page":      {"page": {strings.Repeat("9", 21)}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parsePortalListQuery(values)
			if err == nil {
				t.Fatal("parsePortalListQuery() error = nil")
			}
			assertQueryErrorDoesNotExposeValues(t, err, values)
		})
	}
}

func TestNotificationInboxHTTPRejectsMalformedPaginationBeforeRepository(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	repository := &notificationHTTPRepositoryStub{}
	handler := newNotificationHTTPHandlers(repository)
	resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Name: "Workspace", Slug: "workspace", UserRole: "member",
	}}
	wrapper := mid.Workspace(notificationHTTPLogger(), resolver)(handler.List)
	request := httptest.NewRequest(http.MethodGet, "/workspaces/workspace/notifications?page=secret-page", nil)
	request.SetPathValue("workspaceSlug", "workspace")
	recorder := httptest.NewRecorder()
	if err := wrapper(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("List() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.listQuery != (notificationsdomain.ListQuery{}) {
		t.Fatalf("repository called with %#v", repository.listQuery)
	}
	if strings.Contains(recorder.Body.String(), "secret-page") {
		t.Fatalf("response exposes query value: %s", recorder.Body.String())
	}
}

func assertQueryErrorDoesNotExposeValues(t *testing.T, err error, values url.Values) {
	t.Helper()
	if _, ok := err.(*web.QueryParameterError); !ok {
		t.Fatalf("error type = %T, want *web.QueryParameterError", err)
	}
	for _, supplied := range values {
		for _, value := range supplied {
			if value != "" && strings.Contains(err.Error(), value) {
				t.Fatalf("error %q exposes query value", err)
			}
		}
	}
}

func TestNotificationMutationHTTPCarriesExactScopeAndStatuses(t *testing.T) {
	t.Parallel()

	actorID, workspaceID, notificationID := uuid.New(), uuid.New(), uuid.New()
	repository := &notificationHTTPRepositoryStub{}
	handler := newNotificationHTTPHandlers(repository)
	resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Name: "Workspace", Slug: "workspace", UserRole: "guest",
	}}
	wrapped := mid.Workspace(notificationHTTPLogger(), resolver)(handler.MarkAsRead)
	request := httptest.NewRequest(http.MethodPut, "/workspaces/workspace/notifications/"+notificationID.String()+"/read", nil)
	request.SetPathValue("workspaceSlug", "workspace")
	request.SetPathValue("id", notificationID.String())
	recorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("MarkAsRead() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.mutation.Access.ActorID != actorID || repository.mutation.Access.WorkspaceID != workspaceID ||
		repository.mutation.NotificationID != notificationID || repository.mutation.Kind != notificationsdomain.NotificationMutationRead {
		t.Fatalf("repository mutation = %#v", repository.mutation)
	}

	repository.mutateErr = notificationsdomain.ErrNotFound
	notFoundRecorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), notFoundRecorder, request); err != nil {
		t.Fatalf("not-found MarkAsRead() error = %v", err)
	}
	if notFoundRecorder.Code != http.StatusNotFound {
		t.Fatalf("not-found status = %d, body=%s", notFoundRecorder.Code, notFoundRecorder.Body.String())
	}

	invalidRequest := httptest.NewRequest(http.MethodPut, "/workspaces/workspace/notifications/not-a-uuid/read", nil)
	invalidRequest.SetPathValue("workspaceSlug", "workspace")
	invalidRequest.SetPathValue("id", "not-a-uuid")
	invalidRecorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), invalidRecorder, invalidRequest); err != nil {
		t.Fatalf("invalid MarkAsRead() error = %v", err)
	}
	if invalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid ID status = %d, body=%s", invalidRecorder.Code, invalidRecorder.Body.String())
	}
}

func TestNotificationPreferenceHTTPCreatesPresenceAwarePatch(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	repository := &notificationHTTPRepositoryStub{}
	handler := newNotificationHTTPHandlers(repository)
	resolver := notificationWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Name: "Workspace", Slug: "workspace", UserRole: "member",
	}}
	wrapped := mid.Workspace(notificationHTTPLogger(), resolver)(handler.UpdatePreference)
	request := httptest.NewRequest(
		http.MethodPut,
		"/workspaces/workspace/notification-preferences/mention",
		bytes.NewBufferString(`{"emailEnabled":false}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("workspaceSlug", "workspace")
	request.SetPathValue("type", "mention")
	recorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("UpdatePreference() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("UpdatePreference() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.preference.Access.ActorID != actorID || repository.preference.Access.WorkspaceID != workspaceID ||
		repository.preference.Type != notificationsdomain.PreferenceTypeMention {
		t.Fatalf("repository preference command = %#v", repository.preference)
	}
	email, emailSpecified := repository.preference.Patch.Email.Value()
	if !emailSpecified || email == nil || *email || repository.preference.Patch.InApp.Specified() {
		t.Fatalf("preference channel presence = email %v/%t, in-app %t", email, emailSpecified, repository.preference.Patch.InApp.Specified())
	}

	invalidRequest := httptest.NewRequest(
		http.MethodPut,
		"/workspaces/workspace/notification-preferences/arbitrary",
		bytes.NewBufferString(`{"emailEnabled":false}`),
	)
	invalidRequest.Header.Set("Content-Type", "application/json")
	invalidRequest.SetPathValue("workspaceSlug", "workspace")
	invalidRequest.SetPathValue("type", "arbitrary")
	invalidRecorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), invalidRecorder, invalidRequest); err != nil {
		t.Fatalf("invalid UpdatePreference() error = %v", err)
	}
	if invalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid preference type status = %d, body=%s", invalidRecorder.Code, invalidRecorder.Body.String())
	}
}

func TestPortalNotificationHTTPBindsAuthenticatedActorAndPortalOnly(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	repository := &notificationHTTPRepositoryStub{}
	handler := newNotificationHTTPHandlers(repository)
	request := httptest.NewRequest(http.MethodGet, "/portals/public-roadmap/notifications?unreadOnly=true&page=2&pageSize=5", nil)
	request.SetPathValue("portalSlug", "public-roadmap")
	recorder := httptest.NewRecorder()
	if err := handler.ListPortalFeedback(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("ListPortalFeedback() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListPortalFeedback() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.portalQuery.Access.ActorID != actorID || repository.portalQuery.Access.PortalSlug != "public-roadmap" ||
		!repository.portalQuery.UnreadOnly || repository.portalQuery.Limit != 6 || repository.portalQuery.Offset != 5 {
		t.Fatalf("repository portal query = %#v", repository.portalQuery)
	}
}

func newNotificationHTTPHandlers(repository notifications.Repository) *Handlers {
	service := notifications.New(notificationHTTPLogger(), repository, nil, nil)
	return New(service, nil, nil, notificationHTTPLogger())
}

func notificationHTTPLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelError, "notification-http-test")
}

type notificationWorkspaceResolverStub struct {
	workspace mid.WorkspaceInfo
}

func (resolver notificationWorkspaceResolverStub) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return resolver.workspace, nil
}

func (notificationWorkspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type notificationHTTPRepositoryStub struct {
	listQuery   notificationsdomain.ListQuery
	listResult  []notificationsdomain.Notification
	listErr     error
	mutation    notificationsdomain.NotificationMutation
	mutateErr   error
	preference  notificationsdomain.UpdatePreference
	portalQuery notificationsdomain.PortalListQuery
}

func (*notificationHTTPRepositoryStub) Create(context.Context, notificationsdomain.NewNotification) (notificationsdomain.Notification, bool, error) {
	return notificationsdomain.Notification{}, false, nil
}

func (repository *notificationHTTPRepositoryStub) List(_ context.Context, query notificationsdomain.ListQuery) ([]notificationsdomain.Notification, error) {
	repository.listQuery = query
	return repository.listResult, repository.listErr
}

func (*notificationHTTPRepositoryStub) CountUnread(context.Context, notificationsdomain.WorkspaceAccess) (int, error) {
	return 0, nil
}

func (repository *notificationHTTPRepositoryStub) Mutate(_ context.Context, command notificationsdomain.NotificationMutation) error {
	repository.mutation = command
	return repository.mutateErr
}

func (*notificationHTTPRepositoryStub) MutateAll(context.Context, notificationsdomain.WorkspaceMutation) (int, error) {
	return 0, nil
}

func (*notificationHTTPRepositoryStub) GetPreferences(context.Context, notificationsdomain.WorkspaceAccess) (notificationsdomain.Preferences, error) {
	return notificationsdomain.Preferences{}, nil
}

func (repository *notificationHTTPRepositoryStub) UpdatePreference(_ context.Context, command notificationsdomain.UpdatePreference) (notificationsdomain.Preferences, error) {
	repository.preference = command
	return notificationsdomain.Preferences{}, nil
}

func (repository *notificationHTTPRepositoryStub) ListPortalFeedback(_ context.Context, query notificationsdomain.PortalListQuery) ([]notificationsdomain.PortalNotification, error) {
	repository.portalQuery = query
	return nil, nil
}

func (*notificationHTTPRepositoryStub) CountUnreadPortalFeedback(context.Context, notificationsdomain.PortalAccess) (int, error) {
	return 0, nil
}

func (*notificationHTTPRepositoryStub) MarkPortalFeedbackRead(context.Context, notificationsdomain.PortalNotificationMutation) error {
	return nil
}

func (*notificationHTTPRepositoryStub) MarkAllPortalFeedbackRead(context.Context, notificationsdomain.PortalMutation) (int, error) {
	return 0, nil
}

func (*notificationHTTPRepositoryStub) ListKeyResultAudience(context.Context, notificationsdomain.KeyResultAudienceQuery) ([]notificationsdomain.KeyResultAudienceMember, error) {
	return nil, nil
}

func (*notificationHTTPRepositoryStub) GetEmailDelivery(context.Context, notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error) {
	return nil, nil
}

func (*notificationHTTPRepositoryStub) ListEmailDigest(context.Context, notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error) {
	return nil, nil
}

func (*notificationHTTPRepositoryStub) ListDeliveryTeamIDs(context.Context, notificationsdomain.DeliveryScope) ([]uuid.UUID, error) {
	return nil, nil
}

func (*notificationHTTPRepositoryStub) MarkEmailSent(context.Context, notificationsdomain.MarkEmailSent) error {
	return nil
}
