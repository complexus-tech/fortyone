package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
	"github.com/google/uuid"
)

func TestAdapterRepositoryCatalogUsesVaultTokenBoundaryAndKeysetCursor(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v4/projects" {
			t.Errorf("request path = %q", request.URL.Path)
		}
		if got := request.Header.Get("PRIVATE-TOKEN"); got != "private-secret" {
			t.Errorf("PRIVATE-TOKEN = %q", got)
		}
		if got := request.URL.Query().Get("id_after"); got != "41" {
			t.Errorf("id_after = %q", got)
		}
		if request.URL.Query().Get("pagination") != "keyset" || request.URL.Query().Get("membership") != "true" {
			t.Errorf("query = %v", request.URL.Query())
		}
		response.Header().Set("Link", `<https://gitlab.example/api/v4/projects?pagination=keyset&per_page=25&order_by=id&sort=asc&id_after=42>; rel="next"`)
		_ = json.NewEncoder(response).Encode([]map[string]any{{
			"id": 42, "name": "fortyone", "path_with_namespace": "complexus/fortyone",
			"web_url": "https://gitlab.example/complexus/fortyone", "default_branch": "main",
			"visibility": "private", "archived": false,
			"namespace": map[string]string{"full_path": "complexus"},
		}})
	}))
	defer server.Close()
	adapter := newTestAdapter(t, server, TokenPrivate)
	page, err := adapter.ListRepositories(t.Context(), testInstallation(), codehost.Cursor{Value: "41", Limit: 25})
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(page.Repositories) != 1 || page.Repositories[0].ExternalID != "42" ||
		page.Repositories[0].Owner != "complexus" || !page.Repositories[0].Private || page.NextCursor != "42" {
		t.Fatalf("ListRepositories() = %#v", page)
	}
}

func TestAccessTokenRedactsCommonFormattingAndStructuredLogging(t *testing.T) {
	t.Parallel()
	token := AccessToken{Kind: TokenPrivate, Value: "private-secret"}
	for _, formatted := range []string{fmt.Sprint(token), fmt.Sprintf("%+v", token), fmt.Sprintf("%#v", token), token.LogValue().String()} {
		if strings.Contains(formatted, token.Value) || formatted != "[REDACTED]" {
			t.Fatalf("formatted token = %q", formatted)
		}
	}
}

func TestAdapterRejectsInvalidProjectContinuation(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Link", `<https://gitlab.example/api/v4/projects?pagination=keyset>; rel="next"`)
		_, _ = response.Write([]byte(`[]`))
	}))
	defer server.Close()
	adapter := newTestAdapter(t, server, TokenPrivate)
	if _, err := adapter.ListRepositories(t.Context(), testInstallation(), codehost.Cursor{Limit: 25}); err == nil {
		t.Fatal("ListRepositories() error = nil, want invalid continuation error")
	}
	if _, err := adapter.ListRepositories(t.Context(), testInstallation(), codehost.Cursor{Value: "not-an-id", Limit: 25}); !errors.Is(err, codehost.ErrInvalidInput) {
		t.Fatalf("ListRepositories(invalid cursor) error = %v, want ErrInvalidInput", err)
	}
}

func TestAdapterMapsIssueAndCommentWithoutLeakingProviderTypes(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer private-secret" {
			t.Errorf("Authorization = %q", got)
		}
		switch request.URL.Path {
		case "/api/v4/projects/42/issues":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"id": 501, "iid": 7, "project_id": 42, "title": "Typed integration",
				"description": "Neutral contract", "state": "opened",
				"web_url": "https://gitlab.example/complexus/fortyone/-/issues/7",
			})
		case "/api/v4/projects/42/issues/7/notes":
			_ = json.NewEncoder(response).Encode(map[string]any{
				"id": 9001, "body": "Normalized comment", "created_at": createdAt,
				"author": map[string]any{"id": 19, "username": "joseph"},
			})
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()
	adapter := newTestAdapter(t, server, TokenOAuthBearer)
	repository := testRepository()
	item, err := adapter.CreateWorkItem(t.Context(), testInstallation(), codehost.CreateWorkItem{
		Repository: repository,
		Title:      "Typed integration",
		Body:       "Neutral contract",
	})
	if err != nil {
		t.Fatalf("CreateWorkItem() error = %v", err)
	}
	if item.ExternalID != "501" || item.Number != 7 || item.State != codehost.WorkItemStateOpen {
		t.Fatalf("CreateWorkItem() = %#v", item)
	}
	comment, err := adapter.AddComment(t.Context(), testInstallation(), codehost.AddComment{
		WorkItem: item,
		Body:     "Normalized comment",
	})
	if err != nil {
		t.Fatalf("AddComment() error = %v", err)
	}
	if comment.ExternalID != "9001" || comment.AuthorLogin != "joseph" ||
		!comment.CreatedAt.Equal(createdAt) || comment.WorkItem.Number != 7 {
		t.Fatalf("AddComment() = %#v", comment)
	}
}

func TestAdapterMakesUnsupportedAndAuthenticationFailuresExplicit(t *testing.T) {
	t.Parallel()
	if _, err := NewAdapter(Config{BaseURL: "http://gitlab.example/api/v4"}); !errors.Is(err, codehost.ErrInvalidInput) {
		t.Fatalf("NewAdapter() error = %v, want ErrInvalidInput", err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	adapter := newTestAdapter(t, server, TokenOAuthBearer)
	if err := adapter.Authorize(t.Context(), testInstallation()); !errors.Is(err, codehost.ErrAuthentication) {
		t.Fatalf("Authorize() error = %v, want ErrAuthentication", err)
	}
	if err := adapter.Capabilities().Require(codehost.Capability("merge_request_writer")); !errors.Is(err, codehost.ErrCapabilityUnsupported) {
		t.Fatalf("Require() error = %v, want ErrCapabilityUnsupported", err)
	}
}

func TestAdapterBoundsSuccessfulResponsesWithoutDecodeTargets(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(strings.Repeat("x", maxGitLabResponseBytes+1)))
	}))
	defer server.Close()
	adapter := newTestAdapter(t, server, TokenOAuthBearer)
	if err := adapter.Authorize(t.Context(), testInstallation()); err == nil || !strings.Contains(err.Error(), "exceeds one mebibyte") {
		t.Fatalf("Authorize() error = %v, want bounded-response error", err)
	}
}

func newTestAdapter(t *testing.T, server *httptest.Server, kind TokenKind) *Adapter {
	t.Helper()
	adapter, err := NewAdapter(Config{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Tokens: TokenSourceFunc(func(context.Context, codehost.InstallationRef) (AccessToken, error) {
			return AccessToken{Kind: kind, Value: "private-secret"}, nil
		}),
	})
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}
	return adapter
}

func testInstallation() codehost.InstallationRef {
	return codehost.InstallationRef{
		Provider:               ProviderKey,
		WorkspaceID:            uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		InstallationID:         uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		ExternalInstallationID: "gitlab-account-41",
		Generation:             uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
}

func testRepository() codehost.RepositoryRef {
	return codehost.RepositoryRef{
		ExternalID: "42", Owner: "complexus", Name: "fortyone", FullName: "complexus/fortyone",
		WebURL: "https://gitlab.example/complexus/fortyone", DefaultBranch: "main", Private: true,
	}
}
