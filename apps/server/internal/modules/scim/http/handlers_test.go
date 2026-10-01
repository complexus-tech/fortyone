package scimhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	"github.com/google/uuid"
)

type stub struct {
	Service
	calls int
	token string
	user  domain.User
}

func (s *stub) Authenticate(_ context.Context, _ string, token string) (domain.Scope, error) {
	s.token = token
	if token != "scoped-token" {
		return domain.Scope{}, domain.ErrUnauthorized
	}
	return domain.Scope{}, nil
}
func (s *stub) Create(context.Context, domain.Scope, domain.Input) (domain.User, error) {
	s.calls++
	return s.user, nil
}
func TestProtocolUsesDedicatedBearerAndBoundedKnownBodies(t *testing.T) {
	backend := &stub{user: domain.User{ID: uuid.New(), UserName: "member", Active: true, Profile: []byte(`{}`), CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	h := New(backend, "https://api.example.com")
	for _, test := range []struct {
		name, auth, body string
		status           int
	}{
		{"cookie cannot replace token", "", `{}`, 401},
		{"first party bearer denied", "Bearer first-party-access", `{}`, 401},
		{"unknown attribute", "Bearer scoped-token", `{"schemas":["` + domain.UserSchema + `"],"userName":"member","password":"secret"}`, 400},
		{"oversize", "Bearer scoped-token", `{"displayName":"` + strings.Repeat("x", 65536) + `"}`, 400},
		{"extra json", "Bearer scoped-token", `{} {}`, 400},
		{"valid shape", "Bearer scoped-token", `{"schemas":["` + domain.UserSchema + `"],"userName":"member"}`, 201},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "https://api.example.com/scim/v2/acme/Users", strings.NewReader(test.body))
			r.SetPathValue("workspaceSlug", "acme")
			r.Header.Set("Content-Type", "application/scim+json")
			if test.auth != "" {
				r.Header.Set("Authorization", test.auth)
			}
			r.AddCookie(&http.Cookie{Name: "session", Value: "browser-session"})
			w := httptest.NewRecorder()
			if err := h.Users(t.Context(), w, r); err != nil {
				t.Fatal(err)
			}
			if w.Code != test.status {
				t.Fatal(w.Code, w.Body.String())
			}
			if w.Header().Get("Content-Type") != "application/scim+json" || strings.Contains(w.Body.String(), `"data"`) {
				t.Fatal("incorrect protocol response")
			}
		})
	}
	if backend.calls != 1 {
		t.Fatal("invalid request reached mutation", backend.calls)
	}
}
func TestQueriesCannotChangeScopeOrInjectSQL(t *testing.T) {
	for _, query := range []string{`filter=userName+eq+%22a%22+or+active+eq+true`, `filter=password+eq+%22secret%22`, `workspaceId=other`, `count=1&count=2`, `attributes=password`, `attributes=userName&excludedAttributes=active`, `sortBy=userName`} {
		values, err := url.ParseQuery(query)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = parsePage(values); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatal("invalid query accepted", query, err)
		}
	}
	page, err := parsePage(url.Values{"startIndex": {"-20"}, "count": {"-1"}, "filter": {`userName eq "O'Brien@example.com"`}})
	if err != nil || page.StartIndex != 1 || page.Count != 0 || page.Filter.Value != "O'Brien@example.com" {
		t.Fatal("safe exact filter", page, err)
	}
}
func TestProjectionKeepsRequiredFieldsAndRemovesExcludedProfile(t *testing.T) {
	resource := map[string]any{"schemas": []string{domain.UserSchema}, "id": uuid.NewString(), "userName": "member", "active": true, "name": domain.Name{GivenName: "Name"}}
	result, err := project(resource, url.Values{"attributes": {"userName"}})
	if err != nil || len(result) != 3 || result["active"] != nil {
		t.Fatal(result, err)
	}
	body, _ := json.Marshal(result)
	if !strings.Contains(string(body), `"schemas"`) || !strings.Contains(string(body), `"id"`) {
		t.Fatal("always-returned fields omitted")
	}
}

func TestNestedProjectionKeepsOnlyRequestedSubattributes(t *testing.T) {
	resource := map[string]any{"schemas": []string{domain.UserSchema}, "id": uuid.NewString(), "name": domain.Name{GivenName: "Included", FamilyName: "Excluded"}, "emails": []domain.Email{{Value: "work@example.com", Type: "work", Primary: true}}}
	result, err := project(resource, url.Values{"attributes": {"name.givenName,emails.value"}})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(result)
	if strings.Contains(string(body), "Excluded") || strings.Contains(string(body), `"primary"`) || !strings.Contains(string(body), "Included") || !strings.Contains(string(body), "work@example.com") {
		t.Fatal(string(body))
	}
}
