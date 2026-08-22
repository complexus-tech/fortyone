package agentreadinesshttp

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type memoryOAuthStore struct {
	mu    sync.Mutex
	items map[string][]byte
}

func (s *memoryOAuthStore) Set(_ context.Context, key string, value any, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.items[key] = data
	return nil
}

func (s *memoryOAuthStore) Get(_ context.Context, key string, dest any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.items[key]
	if !ok {
		return errors.New("not found")
	}
	return json.Unmarshal(data, dest)
}

func (s *memoryOAuthStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func TestApprovalPageUsesApplicationThemeAndEscapesClientName(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()

	require.NoError(t, renderApprovalPage(recorder, approvalPageData{
		ClientName: `<script>alert("xss")</script>`,
		Approval:   "approval-token",
	}))

	body := recorder.Body.String()
	require.Contains(t, body, "@media (prefers-color-scheme: dark)")
	require.Contains(t, body, `font-family: -apple-system, BlinkMacSystemFont, "Inter"`)
	require.Contains(t, body, "corner-shape: squircle")
	require.Contains(t, body, "oklch(0.6522 0.2135 38)")
	require.Contains(t, body, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`)
	require.NotContains(t, body, `<script>alert("xss")</script>`)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.Contains(t, recorder.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
	require.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
}

func TestStoryListFiltersUseScopedAdvancedQueryContract(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	teamID := uuid.New()
	sprintID := uuid.New()
	objectiveID := uuid.New()
	assigneeID := uuid.New()
	statusID := uuid.New()
	keyResultID := uuid.New()
	dueOn := "2026-08-22"

	filters, err := storyListFilters(storyListInput{
		TeamID:       teamID.String(),
		SprintID:     sprintID.String(),
		ObjectiveID:  objectiveID.String(),
		AssigneeID:   assigneeID.String(),
		AssignedToMe: true,
		DueOn:        dueOn,
		StatusID:     statusID.String(),
		KeyResultID:  keyResultID.String(),
	}, userID)
	require.NoError(t, err)
	require.Equal(t, userID, filters["current_user_id"])
	require.Equal(t, []uuid.UUID{teamID}, filters["team_ids"])
	require.Equal(t, []uuid.UUID{sprintID}, filters["sprint_ids"])
	require.Equal(t, objectiveID, filters["objective_id"])
	require.Equal(t, []uuid.UUID{assigneeID}, filters["assignee_ids"])
	require.Equal(t, true, filters["assigned_to_me"])
	expectedDueOn := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDueOn, filters["deadline_after"])
	require.Equal(t, expectedDueOn, filters["deadline_before"])
	require.Equal(t, []uuid.UUID{statusID}, filters["status_ids"])
	require.Equal(t, keyResultID, filters["key_result_id"])
}

func TestStoryListFiltersRejectInvalidUUID(t *testing.T) {
	t.Parallel()
	_, err := storyListFilters(storyListInput{TeamID: "not-a-uuid"}, uuid.New())
	require.EqualError(t, err, "teamId must be a valid UUID")
}

func TestStoryListFiltersRejectInvalidDueDate(t *testing.T) {
	t.Parallel()
	_, err := storyListFilters(storyListInput{DueOn: "today"}, uuid.New())
	require.EqualError(t, err, "dueOn: date must use YYYY-MM-DD")
}

func TestOpenAPIIsValidJSONWithUniqueOperationIDs(t *testing.T) {
	t.Parallel()
	handler := New(Config{SecretKey: "test", APIPublicURL: "https://api.fortyone.app"})
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.OpenAPI(context.Background(), recorder, httptest.NewRequest(http.MethodGet, "/openapi.json", nil)))
	require.Equal(t, http.StatusOK, recorder.Code)
	var document map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &document))
	require.Equal(t, "3.1.1", document["openapi"])
}

func TestOAuthDiscoveryDescribesRemoteMCPResource(t *testing.T) {
	t.Parallel()
	handler := New(Config{SecretKey: "test", APIPublicURL: "https://api.fortyone.app"})
	protected := httptest.NewRecorder()
	require.NoError(t, handler.ProtectedResourceMetadata(context.Background(), protected, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)))
	require.Contains(t, protected.Body.String(), `"resource":"https://api.fortyone.app/mcp"`)
	require.Contains(t, protected.Body.String(), `"authorization_servers":["https://api.fortyone.app"]`)

	authorization := httptest.NewRecorder()
	require.NoError(t, handler.AuthorizationServerMetadata(context.Background(), authorization, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)))
	require.Contains(t, authorization.Body.String(), `"code_challenge_methods_supported":["S256"]`)
	require.Contains(t, authorization.Body.String(), `"registration_endpoint":"https://api.fortyone.app/oauth/register"`)
}

func TestMCPRejectsMissingBearerTokenWithDiscoveryHint(t *testing.T) {
	t.Parallel()
	handler := New(Config{SecretKey: "test", APIPublicURL: "https://api.fortyone.app"})
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.MCP(context.Background(), recorder, request))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Header().Get("WWW-Authenticate"), "oauth-protected-resource")
}

func TestMCPListsSchedulingAndPlanningTools(t *testing.T) {
	t.Parallel()
	handler := New(Config{SecretKey: "test", APIPublicURL: "https://api.fortyone.app"})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mcpClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "e1e76f7c-2832-43b6-88f7-0af378bde150", Audience: jwt.ClaimStrings{"https://api.fortyone.app/mcp"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}, Scope: mcpScope})
	rawToken, err := token.SignedString(handler.signingKey())
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("Authorization", "Bearer "+rawToken)
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.MCP(context.Background(), recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code)
	for _, expected := range []string{"create_story", "estimatedDurationMinutes", "minimumFocusBlockMinutes", "autoSchedulingEnabled", "create_sprint", "create_objective", "create_key_result", "analyze_work"} {
		require.Contains(t, recorder.Body.String(), expected)
	}
}

func TestRedirectURIValidation(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateRedirectURI("https://chatgpt.com/aip/callback"))
	require.NoError(t, validateRedirectURI("http://localhost:8123/callback"))
	require.Error(t, validateRedirectURI("http://example.com/callback"))
	require.Error(t, validateRedirectURI("https://example.com/callback#fragment"))
}

func TestOAuthAuthorizationCodeExchangeUsesPKCEAndResourceAudience(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	handler := New(Config{SecretKey: "test", APIPublicURL: "https://api.fortyone.app", Cache: store})
	verifier := "correct-horse-battery-staple"
	digest := sha256.Sum256([]byte(verifier))
	record := authorizationCode{
		ClientID:      "test-client",
		RedirectURI:   "https://client.example/callback",
		Scope:         "mcp:access offline_access",
		CodeChallenge: base64.RawURLEncoding.EncodeToString(digest[:]),
		UserID:        "e1e76f7c-2832-43b6-88f7-0af378bde150",
	}
	require.NoError(t, store.Set(context.Background(), oauthKey("code", hashToken("one-time-code")), record, time.Minute))

	form := "grant_type=authorization_code&code=one-time-code&client_id=test-client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&code_verifier=" + verifier + "&resource=https%3A%2F%2Fapi.fortyone.app%2Fmcp"
	request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.Token(context.Background(), recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotEmpty(t, response.RefreshToken)
	info, err := handler.verifyToken(context.Background(), response.AccessToken, request)
	require.NoError(t, err)
	require.Equal(t, record.UserID, info.UserID)
}

func TestOAuthUsesDedicatedLoginURLWithoutChangingWebsiteURL(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	handler := New(Config{
		SecretKey:    "test",
		APIPublicURL: "https://api.fortyone.app",
		LoginURL:     "https://cloud.fortyone.app",
		Cache:        store,
	})
	client := oauthClient{ClientID: "test-client", ClientName: "Test client", RedirectURIs: []string{"https://client.example/callback"}}
	require.NoError(t, store.Set(context.Background(), oauthKey("client", client.ClientID), client, time.Minute))

	request := httptest.NewRequest(http.MethodGet, "/oauth/authorize?response_type=code&client_id=test-client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&code_challenge=challenge&code_challenge_method=S256&resource=https%3A%2F%2Fapi.fortyone.app%2Fmcp", nil)
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.Authorize(context.Background(), recorder, request))
	require.Equal(t, http.StatusFound, recorder.Code)
	require.True(t, strings.HasPrefix(recorder.Header().Get("Location"), "https://cloud.fortyone.app/?callbackUrl="))
}
