package agentreadinesshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type oauthSessionAccounts struct{}

func (oauthSessionAccounts) ResolveActiveBrowserSessionVersion(context.Context, uuid.UUID) (int64, bool, error) {
	return 1, true, nil
}

func testOAuthBrowserSessions(store mid.SessionStore) mid.SessionResolver {
	return mid.NewBrowserSessionResolver(store, oauthSessionAccounts{})
}

func TestApprovalPageUsesApplicationThemeAndEscapesClientName(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()

	require.NoError(t, renderApprovalPage(recorder, approvalPageData{
		ClientName:  `<script>alert("xss")</script>`,
		Approval:    "approval-token",
		RedirectURI: "https://chatgpt.com/connector/oauth/fortyone",
	}))

	body := recorder.Body.String()
	require.Contains(t, body, "@media (prefers-color-scheme: dark)")
	require.Contains(t, body, `--font-body: -apple-system, BlinkMacSystemFont, "Inter", sans-serif`)
	require.Contains(t, body, "width: min(100%, 416px)")
	require.Contains(t, body, "--color-border: oklch(0.92 0.004 95)")
	require.Contains(t, body, "--radius-scale: 0.75")
	require.Contains(t, body, "--radius-scale: 2.5")
	require.Contains(t, body, "border-radius: var(--radius-xl)")
	require.Contains(t, body, "corner-shape: squircle")
	require.Contains(t, body, "oklch(0.6522 0.2135 38)")
	require.Contains(t, body, "would like to connect to FortyOne")
	require.Contains(t, body, "This connection will be able to:")
	require.Contains(t, body, "changes require your approval in the connected client")
	require.Contains(t, body, "font-size: 14px")
	require.Contains(t, body, "Agree &amp; allow")
	require.NotContains(t, body, "Agree &amp; allow access")
	require.Contains(t, body, `&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`)
	require.NotContains(t, body, `<script>alert("xss")</script>`)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	contentSecurityPolicy := recorder.Header().Get("Content-Security-Policy")
	require.Contains(t, contentSecurityPolicy, "form-action 'self' https://chatgpt.com")
	require.NotContains(t, contentSecurityPolicy, "connector/oauth/fortyone")
	require.Contains(t, contentSecurityPolicy, "frame-ancestors 'none'")
	require.Equal(t, "no-referrer", recorder.Header().Get("Referrer-Policy"))
}

func TestStoryListFiltersUseScopedAdvancedQueryContract(t *testing.T) {
	t.Parallel()
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
		Search:       " launch task ",
		StatusID:     statusID.String(),
		KeyResultID:  keyResultID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{teamID}, filters.TeamIDs)
	require.Equal(t, []uuid.UUID{sprintID}, filters.SprintIDs)
	require.Equal(t, objectiveID, *filters.Objective)
	require.Equal(t, []uuid.UUID{assigneeID}, filters.AssigneeIDs)
	require.NotNil(t, filters.AssignedToMe)
	require.True(t, *filters.AssignedToMe)
	expectedDueOn := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	require.Equal(t, expectedDueOn, *filters.DeadlineAfter)
	require.Equal(t, expectedDueOn, *filters.DeadlineBefore)
	require.Equal(t, "launch task", *filters.TitleContains)
	require.Equal(t, []uuid.UUID{statusID}, filters.StatusIDs)
	require.Equal(t, keyResultID, *filters.KeyResult)
}

func TestMCPPaginationIsBoundedAndReportsMoreResults(t *testing.T) {
	t.Parallel()
	page, pageSize, offset, limit := normalizePagination(2, 1000)
	require.Equal(t, 2, page)
	require.Equal(t, maximumMCPPageSize, pageSize)
	require.Equal(t, maximumMCPPageSize, offset)
	require.Equal(t, maximumMCPPageSize+1, limit)

	items := make([]int, 205)
	pageItems, hasMore := pageSlice(items, offset, limit, pageSize)
	require.Len(t, pageItems, maximumMCPPageSize)
	require.True(t, hasMore)

	lastPage, hasMore := pageSlice(items, 200, 101, 100)
	require.Len(t, lastPage, 5)
	require.False(t, hasMore)

	shortSection, hasMore := pageSliceWithMore([]int{1, 2}, 0, 101, 100, true)
	require.Equal(t, []int{1, 2}, shortSection)
	require.True(t, hasMore, "pagination must preserve hasMore from another analysis section")

	_, _, hugeOffset, hugeLimit := normalizePagination(int(^uint(0)>>1), maximumMCPPageSize)
	require.GreaterOrEqual(t, hugeOffset, 0)
	require.LessOrEqual(t, hugeOffset, int(^uint(0)>>1)-hugeLimit)
}

func TestStoryToolResultDoesNotExposeInternalCreationKey(t *testing.T) {
	t.Parallel()
	creationKey := "mcp:user:private-retry-token"
	result := storyToolResult(stories.CoreSingleStory{ID: uuid.New(), Title: "Public title", CreationKey: &creationKey})

	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	require.Contains(t, string(encoded), "Public title")
	require.NotContains(t, string(encoded), creationKey)
	require.NotContains(t, string(encoded), "CreationKey")
	require.NotContains(t, string(encoded), "creationKey")
}

func TestStoryListFiltersRejectInvalidUUID(t *testing.T) {
	t.Parallel()
	_, err := storyListFilters(storyListInput{TeamID: "not-a-uuid"})
	require.EqualError(t, err, "teamId must be a valid UUID")
}

func TestStoryListFiltersRejectInvalidDueDate(t *testing.T) {
	t.Parallel()
	_, err := storyListFilters(storyListInput{DueOn: "today"})
	require.EqualError(t, err, "dueOn: date must use YYYY-MM-DD")
}

func TestOpenAPIIsValidJSONWithUniqueOperationIDs(t *testing.T) {
	t.Parallel()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app"})
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.OpenAPI(context.Background(), recorder, httptest.NewRequest(http.MethodGet, "/openapi.json", nil)))
	require.Equal(t, http.StatusOK, recorder.Code)
	var document map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &document))
	require.Equal(t, "3.1.1", document["openapi"])
}

func TestOAuthDiscoveryDescribesRemoteMCPAndAPIResources(t *testing.T) {
	t.Parallel()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: newStubOAuthPlatform()})
	protected := httptest.NewRecorder()
	require.NoError(t, handler.ProtectedResourceMetadata(context.Background(), protected, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)))
	require.Contains(t, protected.Body.String(), `"resource":"https://api.fortyone.app/mcp"`)
	require.Contains(t, protected.Body.String(), `"authorization_servers":["https://api.fortyone.app"]`)
	require.Contains(t, protected.Body.String(), `"mcp:access"`)

	apiProtected := httptest.NewRecorder()
	require.NoError(t, handler.APIProtectedResourceMetadata(context.Background(), apiProtected, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource/api/v1", nil)))
	require.Contains(t, apiProtected.Body.String(), `"resource":"https://api.fortyone.app/api/v1"`)
	require.Contains(t, apiProtected.Body.String(), `"stories:read"`)
	require.NotContains(t, apiProtected.Body.String(), `"mcp:access"`)

	authorization := httptest.NewRecorder()
	require.NoError(t, handler.AuthorizationServerMetadata(context.Background(), authorization, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)))
	require.Contains(t, authorization.Body.String(), `"code_challenge_methods_supported":["S256"]`)
	require.Contains(t, authorization.Body.String(), `"registration_endpoint":"https://api.fortyone.app/oauth/register"`)
	require.Contains(t, authorization.Body.String(), `"authorization_response_iss_parameter_supported":true`)
	require.Contains(t, authorization.Body.String(), `"stories:write"`)
	require.Contains(t, authorization.Body.String(), `"client_credentials"`)
	require.Contains(t, authorization.Body.String(), `"client_secret_basic"`)
}

func TestOAuthApprovalPreservesExactAPIResource(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	oauth := newStubOAuthPlatform()
	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app", Cache: store, OAuth: oauth,
		BrowserSessions: testOAuthBrowserSessions(store),
	})
	const approval = "api-resource-approval"
	const session = "api-resource-session"
	require.NoError(t, store.Set(context.Background(), oauthApprovalKey(approval), authorizationRequest{
		ClientID: oauth.application.ClientID, RedirectURI: oauth.application.RedirectURIs[0],
		Resource: oauth.apiResource, Scope: "offline_access stories:read", CodeChallenge: "challenge",
		UserID: oauth.identity.UserID.String(),
	}, approvalTTL))
	require.NoError(t, store.Set(context.Background(), cache.AuthSessionCacheKey(session), platformauth.BrowserSession{
		UserID: oauth.identity.UserID, Version: 1,
	}, time.Hour))

	request := httptest.NewRequest(http.MethodPost, "/oauth/authorize", strings.NewReader("approval="+approval+"&decision=allow"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: "fortyone_session", Value: session})
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.Authorize(context.Background(), recorder, request))
	require.Equal(t, http.StatusFound, recorder.Code)
	require.Equal(t, oauth.apiResource, oauth.authorizedRequest().Resource)
}

func TestOAuthRegistrationRejectsUnknownAndTrailingJSON(t *testing.T) {
	t.Parallel()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: newStubOAuthPlatform()})

	for _, body := range []string{
		`{"client_name":"Client","redirect_uris":["https://client.example/callback"],"trusted":true}`,
		`{"client_name":"Client","redirect_uris":["https://client.example/callback"]} {}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		require.NoError(t, handler.RegisterClient(context.Background(), recorder, request))
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Contains(t, recorder.Body.String(), `"error":"invalid_client_metadata"`)
	}
}

func TestOAuthInvalidResourceCannotRedirectThroughUnregisteredURI(t *testing.T) {
	t.Parallel()
	oauth := newStubOAuthPlatform()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})
	request := httptest.NewRequest(
		http.MethodGet,
		"/oauth/authorize?response_type=code&client_id="+url.QueryEscape(oauth.application.ClientID)+
			"&redirect_uri="+url.QueryEscape("https://attacker.example/callback")+
			"&code_challenge="+strings.Repeat("a", 43)+
			"&code_challenge_method=S256&resource="+url.QueryEscape("https://wrong.example/mcp"),
		nil,
	)
	recorder := httptest.NewRecorder()

	require.NoError(t, handler.Authorize(context.Background(), recorder, request))
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Empty(t, recorder.Header().Get("Location"))
	require.Contains(t, recorder.Body.String(), `"error":"invalid_request"`)
}

func TestOAuthApprovalIsConsumedExactlyOnceUnderConcurrency(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	oauth := newStubOAuthPlatform()
	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app", Cache: store, OAuth: oauth,
		BrowserSessions: testOAuthBrowserSessions(store),
	})
	userID := oauth.identity.UserID
	const approval = "single-use-approval"
	const session = "browser-session"
	pending := authorizationRequest{
		ClientID: oauth.application.ClientID, RedirectURI: oauth.application.RedirectURIs[0],
		State: "opaque-state", Scope: mcpScope + " offline_access", CodeChallenge: "challenge",
		UserID: userID.String(),
	}
	require.NoError(t, store.Set(context.Background(), oauthApprovalKey(approval), pending, approvalTTL))
	require.NoError(t, store.Set(context.Background(), cache.AuthSessionCacheKey(session), platformauth.BrowserSession{
		UserID: userID, Version: 1,
	}, time.Hour))

	start := make(chan struct{})
	statusCodes := make([]int, 2)
	locations := make([]string, 2)
	var wait sync.WaitGroup
	for index := range statusCodes {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			request := httptest.NewRequest(
				http.MethodPost,
				"/oauth/authorize",
				strings.NewReader("approval="+approval+"&decision=allow"),
			)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.AddCookie(&http.Cookie{Name: "fortyone_session", Value: session})
			recorder := httptest.NewRecorder()
			require.NoError(t, handler.Authorize(context.Background(), recorder, request))
			statusCodes[index] = recorder.Code
			locations[index] = recorder.Header().Get("Location")
		}()
	}
	close(start)
	wait.Wait()

	redirects := 0
	rejections := 0
	for index, status := range statusCodes {
		switch status {
		case http.StatusFound:
			redirects++
			require.Contains(t, locations[index], "code=one-time-code")
		case http.StatusBadRequest:
			rejections++
		default:
			t.Fatalf("approval status = %d, want redirect or replay rejection", status)
		}
	}
	require.Equal(t, 1, redirects)
	require.Equal(t, 1, rejections)
}

func TestOAuthMismatchedSessionCannotConsumeAnotherUsersApproval(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	oauth := newStubOAuthPlatform()
	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app", Cache: store, OAuth: oauth,
		BrowserSessions: testOAuthBrowserSessions(store),
	})
	const approval = "session-bound-approval"
	require.NoError(t, store.Set(context.Background(), oauthApprovalKey(approval), authorizationRequest{
		ClientID: oauth.application.ClientID, RedirectURI: oauth.application.RedirectURIs[0],
		Scope: mcpScope, CodeChallenge: "challenge", UserID: oauth.identity.UserID.String(),
	}, approvalTTL))
	require.NoError(t, store.Set(context.Background(), cache.AuthSessionCacheKey("wrong-session"), platformauth.BrowserSession{
		UserID: uuid.New(), Version: 1,
	}, time.Hour))
	require.NoError(t, store.Set(context.Background(), cache.AuthSessionCacheKey("right-session"), platformauth.BrowserSession{
		UserID: oauth.identity.UserID, Version: 1,
	}, time.Hour))

	request := func(session string) *http.Request {
		value := httptest.NewRequest(http.MethodPost, "/oauth/authorize", strings.NewReader("approval="+approval+"&decision=allow"))
		value.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		value.AddCookie(&http.Cookie{Name: "fortyone_session", Value: session})
		return value
	}
	wrong := httptest.NewRecorder()
	require.NoError(t, handler.Authorize(context.Background(), wrong, request("wrong-session")))
	require.Equal(t, http.StatusUnauthorized, wrong.Code)

	right := httptest.NewRecorder()
	require.NoError(t, handler.Authorize(context.Background(), right, request("right-session")))
	require.Equal(t, http.StatusFound, right.Code)
	require.Contains(t, right.Header().Get("Location"), "code=one-time-code")
}

func TestMCPRejectsMissingBearerTokenWithDiscoveryHint(t *testing.T) {
	t.Parallel()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app"})
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.MCP(context.Background(), recorder, request))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Header().Get("WWW-Authenticate"), "oauth-protected-resource")
}

func TestMCPListsSchedulingAndPlanningTools(t *testing.T) {
	t.Parallel()
	oauth := newStubOAuthPlatform()
	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app",
		OAuth:        oauth,
		Cache:        &memoryOAuthStore{items: make(map[string][]byte)},
	})
	rawToken := oauth.pair.AccessToken.Reveal()

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
	require.Error(t, validateRedirectURI("https://user@example.com/callback"))

	origin, err := oauthRedirectOrigin("https://chatgpt.com/connector/oauth/fortyone?state=opaque")
	require.NoError(t, err)
	require.Equal(t, "https://chatgpt.com", origin)
}

func TestOAuthAuthorizationCodeExchangeReturnsBoundedTokenResponse(t *testing.T) {
	t.Parallel()
	oauth := newStubOAuthPlatform()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})

	form := "grant_type=authorization_code&code=one-time-code&client_id=test-client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&code_verifier=verifier&resource=https%3A%2F%2Fapi.fortyone.app%2Fmcp"
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
	require.Equal(t, oauth.pair.RefreshToken.Reveal(), response.RefreshToken)
	info, err := handler.verifyToken(context.Background(), response.AccessToken, request)
	require.NoError(t, err)
	require.Equal(t, oauth.identity.UserID.String(), info.UserID)
}

func TestOAuthTokenRejectionsAreLoggedWithoutCredentialValues(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	oauth := newStubOAuthPlatform()
	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app",
		OAuth:        oauth,
		Log:          logger.NewWithJSON(&logs, slog.LevelDebug, "agent-readiness-test"),
	})
	request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader("grant_type=authorization_code&code=sensitive-code&resource=https%3A%2F%2Fwrong.example%2Fmcp"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	require.NoError(t, handler.Token(context.Background(), recorder, request))
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"error":"invalid_target"`)
	require.Contains(t, logs.String(), `"msg":"MCP OAuth request rejected"`)
	require.Contains(t, logs.String(), `"operation":"exchange_code"`)
	require.Contains(t, logs.String(), `"oauth_error":"invalid_target"`)
	require.NotContains(t, logs.String(), "sensitive-code")
}

func TestOAuthTokenUsesOnlyBoundedBodyCredentials(t *testing.T) {
	t.Parallel()

	handler := New(Config{
		APIPublicURL: "https://api.fortyone.app",
		OAuth:        newStubOAuthPlatform(),
	})

	t.Run("query parameters cannot substitute for token form fields", func(t *testing.T) {
		request := httptest.NewRequest(
			http.MethodPost,
			"/oauth/token?grant_type=refresh_token&refresh_token=query-secret&client_id=query-client",
			strings.NewReader("grant_type=unsupported"),
		)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()

		require.NoError(t, handler.Token(context.Background(), recorder, request))
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Contains(t, recorder.Body.String(), `"error":"unsupported_grant_type"`)
	})

	t.Run("oversized form is rejected before credential lookup", func(t *testing.T) {
		request := httptest.NewRequest(
			http.MethodPost,
			"/oauth/token",
			strings.NewReader("grant_type=refresh_token&refresh_token="+strings.Repeat("a", oauthFormBodyLimit)),
		)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()

		require.NoError(t, handler.Token(context.Background(), recorder, request))
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Contains(t, recorder.Body.String(), `"error":"invalid_request"`)
	})
}

func TestOAuthUsesDedicatedLoginURLWithoutChangingWebsiteURL(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	oauth := newStubOAuthPlatform()
	oauth.application.ClientID = "test-client"
	handler := New(Config{
		APIPublicURL:    "https://api.fortyone.app",
		LoginURL:        "https://cloud.fortyone.app",
		Cache:           store,
		BrowserSessions: testOAuthBrowserSessions(store),
		OAuth:           oauth,
	})

	request := httptest.NewRequest(http.MethodGet, "/oauth/authorize?response_type=code&client_id=test-client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&code_challenge=challenge&code_challenge_method=S256&resource=https%3A%2F%2Fapi.fortyone.app%2Fmcp", nil)
	recorder := httptest.NewRecorder()
	require.NoError(t, handler.Authorize(context.Background(), recorder, request))
	require.Equal(t, http.StatusFound, recorder.Code)
	require.True(t, strings.HasPrefix(recorder.Header().Get("Location"), "https://cloud.fortyone.app/?callbackUrl="))
}
