package enterprisessohttp

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

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
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
	return New(service, states, writer, "https://api.example.com/auth/sso/callback", "https://app.example.com", "", deployment.Production), service, states, writer
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

func TestDesktopSSOReturnIsBoundedAndSurvivesSuccessfulCallback(t *testing.T) {
	h, _, _, writer := setup()
	proof := strings.Repeat("s", 43)
	target := "/auth/desktop?state=" + proof + "&code_challenge=" + proof
	r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/acme?returnTo="+url.QueryEscape(target), nil)
	r.SetPathValue("workspaceSlug", "acme")
	w := httptest.NewRecorder()
	if err := h.Start(t.Context(), w, r); err != nil || w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("start: %v %s", err, w.Body.String())
	}
	location, _ := url.Parse(w.Header().Get("Location"))
	result := callback(t, h, t.Context(), location.Query().Get("state"), w.Result().Cookies()[0])
	canonical, valid := desktopReturnPath(target, deployment.Production)
	if !valid || result.Header().Get("Location") != "https://app.example.com"+canonical || writer.calls != 1 {
		t.Fatalf("desktop callback: %s", result.Header().Get("Location"))
	}
}

func TestDesktopSSOReturnRejectsOpenRedirectsAndExtraProofFields(t *testing.T) {
	proof := strings.Repeat("s", 43)
	valid := "/auth/desktop?state=" + proof + "&code_challenge=" + proof
	for _, target := range []string{"https://evil.example" + valid, "//evil.example" + valid, valid + "#fragment", valid + "&redirect=evil", valid + "&state=" + proof, "/auth/desktop?state=short&code_challenge=" + proof, "/auth/desktop%3F?state=" + proof + "&code_challenge=" + proof, "/auth/desktop\\?state=" + proof + "&code_challenge=" + proof} {
		if _, allowed := desktopReturnPath(target, deployment.Production); allowed {
			t.Fatalf("allowed unsafe target %q", target)
		}
	}
}

func TestDesktopDevelopmentSSOReturnPreservesItsCallback(t *testing.T) {
	h, _, _, writer := setup()
	h.deploymentMode = deployment.Development
	proof := strings.Repeat("s", 43)
	target := "/auth/desktop?state=" + proof + "&code_challenge=" + proof + "&redirect_uri=" + url.QueryEscape(auth.DesktopDevelopmentRedirectURI)
	r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/acme?returnTo="+url.QueryEscape(target), nil)
	r.SetPathValue("workspaceSlug", "acme")
	w := httptest.NewRecorder()
	if err := h.Start(t.Context(), w, r); err != nil || w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("start: %v %s", err, w.Body.String())
	}
	location, _ := url.Parse(w.Header().Get("Location"))
	result := callback(t, h, t.Context(), location.Query().Get("state"), w.Result().Cookies()[0])
	canonical, valid := desktopReturnPath(target, deployment.Development)
	if !valid || result.Header().Get("Location") != "https://app.example.com"+canonical || writer.calls != 1 {
		t.Fatalf("development callback: %s", result.Header().Get("Location"))
	}
}

func TestDesktopSSOReturnRedirectAllowlistUsesDeploymentMode(t *testing.T) {
	proof := strings.Repeat("s", 43)
	base := "/auth/desktop?state=" + proof + "&code_challenge=" + proof
	for _, mode := range []deployment.Mode{deployment.Development, deployment.Test, deployment.Staging, deployment.Production, ""} {
		t.Run(string(mode), func(t *testing.T) {
			valid := []string{base, base + "&redirect_uri=" + url.QueryEscape(auth.DesktopRedirectURI)}
			invalid := []string{
				base + "&redirect_uri=",
				base + "&redirect_uri=" + url.QueryEscape(auth.DesktopDevelopmentRedirectURI+"/"),
				base + "&redirect_uri=" + url.QueryEscape("https://evil.example/login"),
				base + "&redirect_uri=" + url.QueryEscape(auth.DesktopRedirectURI) + "&redirect_uri=" + url.QueryEscape(auth.DesktopRedirectURI),
				base + "&redirect_uri=" + url.QueryEscape(auth.DesktopRedirectURI) + "&extra=field",
			}
			dev := base + "&redirect_uri=" + url.QueryEscape(auth.DesktopDevelopmentRedirectURI)
			if mode == deployment.Development {
				valid = append(valid, dev)
			} else {
				invalid = append(invalid, dev)
			}
			for _, target := range valid {
				if _, allowed := desktopReturnPath(target, mode); !allowed {
					t.Errorf("rejected allowed target %q", target)
				}
			}
			for _, target := range invalid {
				if _, allowed := desktopReturnPath(target, mode); allowed {
					t.Errorf("allowed invalid target %q", target)
				}
			}
			h, _, states, _ := setup()
			h.deploymentMode = mode
			r := httptest.NewRequest(http.MethodGet, "https://api.example.com/auth/sso/acme?returnTo="+url.QueryEscape(dev), nil)
			r.SetPathValue("workspaceSlug", "acme")
			w := httptest.NewRecorder()
			if err := h.Start(t.Context(), w, r); err != nil {
				t.Fatal(err)
			}
			if mode != deployment.Development && (w.Code != http.StatusBadRequest || len(states.values) != 0) {
				t.Fatal("non-development SSO persisted a development handoff")
			}
		})
	}
}
