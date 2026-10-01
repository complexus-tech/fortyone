package enterprisessohttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type stateStore struct {
	values map[string][]byte
	takes  int
}

func (s *stateStore) Set(_ context.Context, key string, value any, _ time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.values[key] = encoded
	return nil
}
func (s *stateStore) Take(_ context.Context, key string, target any) error {
	value, ok := s.values[key]
	if !ok {
		return errors.New("missing state")
	}
	delete(s.values, key)
	s.takes++
	return json.Unmarshal(value, target)
}

type flowService struct {
	Service
	connection domain.Connection
	completed  int
	attempt    domain.Attempt
}

func (s *flowService) Public(context.Context, string) (domain.Connection, error) {
	return s.connection, nil
}
func (s *flowService) AuthorizationURL(_ context.Context, _ domain.Connection, attempt domain.Attempt, state string) (string, error) {
	s.attempt = attempt
	return "https://idp.example.com/authorize?state=" + url.QueryEscape(state), nil
}
func (s *flowService) Complete(context.Context, domain.Attempt, string) (uuid.UUID, domain.Identity, error) {
	s.completed++
	return uuid.New(), domain.Identity{AuthenticatedAt: time.Now().UTC()}, nil
}

type issuer struct{ calls int }

func (i *issuer) Issue(context.Context, http.ResponseWriter, *http.Request, uuid.UUID, auth.WorkspaceSSOAuthentication) error {
	i.calls++
	return nil
}
func setup() (*Handlers, *flowService, *stateStore, *issuer) {
	service := &flowService{connection: domain.Connection{ID: uuid.New(), WorkspaceID: uuid.New(), Generation: 1, Enabled: true}}
	states := &stateStore{values: map[string][]byte{}}
	writer := &issuer{}
	return New(service, states, writer, "https://api.example.com/auth/sso/callback", "https://app.example.com", "", true), service, states, writer
}
func begin(t *testing.T, h *Handlers, ctx context.Context) (string, *http.Cookie) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/acme", nil)
	r.SetPathValue("workspaceSlug", "acme")
	w := httptest.NewRecorder()
	if err := h.Start(ctx, w, r); err != nil || w.Code != http.StatusTemporaryRedirect {
		t.Fatal("start", err, w.Body.String())
	}
	location, _ := url.Parse(w.Header().Get("Location"))
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("SSO state cookie not protected")
	}
	return location.Query().Get("state"), cookies[0]
}
func callback(t *testing.T, h *Handlers, ctx context.Context, state string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/callback?state="+url.QueryEscape(state)+"&code=private-code", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	if err := h.Callback(ctx, w, r); err != nil {
		t.Fatal(err)
	}
	return w
}
func TestSSOStateBoundToBrowserAndConsumedOnce(t *testing.T) {
	h, service, states, writer := setup()
	state, cookie := begin(t, h, t.Context())
	if state == "" || service.attempt.Nonce == "" || service.attempt.Verifier == "" {
		t.Fatal("missing code-flow proofs")
	}
	response := callback(t, h, t.Context(), state, &http.Cookie{Name: stateCookie, Value: "attacker"})
	if response.Code != http.StatusSeeOther || service.completed != 0 || states.takes != 0 {
		t.Fatal("browser mismatch consumed valid state or reached provider")
	}
	response = callback(t, h, t.Context(), state, cookie)
	if response.Header().Get("Location") != "https://app.example.com/acme" || service.completed != 1 || writer.calls != 1 {
		t.Fatal("verified SSO did not issue scoped session")
	}
	response = callback(t, h, t.Context(), state, cookie)
	if service.completed != 1 || writer.calls != 1 || response.Header().Get("Location") != "https://app.example.com/?error=sso_expired" {
		t.Fatal("callback replay reached provider")
	}
}
func TestSSOInitialLinkRequiresTheSameAccountAtCallback(t *testing.T) {
	h, service, _, writer := setup()
	ctx := auth.SetUserID(t.Context(), uuid.New())
	state, cookie := begin(t, h, ctx)
	if service.attempt.LinkUserID == uuid.Nil {
		t.Fatal("initial link lost authenticated account proof")
	}
	response := callback(t, h, t.Context(), state, cookie)
	if service.completed != 0 || writer.calls != 0 || response.Header().Get("Location") != "https://app.example.com/?error=sso_failed" {
		t.Fatal("initial link survived account sign-out")
	}
	r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/acme?callbackUrl=https://attacker.example", nil)
	r.SetPathValue("workspaceSlug", "acme")
	w := httptest.NewRecorder()
	_ = h.Start(ctx, w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatal("arbitrary callback accepted")
	}
}

func TestSSORecoveryRetainsExistingAccountProof(t *testing.T) {
	h, service, _, writer := setup()
	actorID := uuid.New()
	ctx := auth.SetUserID(t.Context(), actorID)
	state, cookie := begin(t, h, ctx)
	if service.attempt.LinkUserID != actorID {
		t.Fatal("public SSO continuation lost the existing account identity")
	}
	response := callback(t, h, ctx, state, cookie)
	if response.Header().Get("Location") != "https://app.example.com/acme" || service.completed != 1 || writer.calls != 1 {
		t.Fatal("an authenticated unlinked account cannot complete SSO recovery")
	}
}
