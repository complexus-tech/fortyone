package agentreadinesshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/stretchr/testify/require"
)

func TestOAuthGlobalRateLimitFailsClosedAndUsesProtocolError(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	middleware := oauthGlobalRateLimit(Config{Cache: store}, "registration-test", 1, time.Minute)
	next := middleware(func(_ context.Context, writer http.ResponseWriter, _ *http.Request) error {
		writer.WriteHeader(http.StatusNoContent)
		return nil
	})

	first := httptest.NewRecorder()
	require.NoError(t, next(context.Background(), first, httptest.NewRequest(http.MethodPost, "/oauth/register", nil)))
	require.Equal(t, http.StatusNoContent, first.Code)

	second := httptest.NewRecorder()
	require.NoError(t, next(context.Background(), second, httptest.NewRequest(http.MethodPost, "/oauth/register", nil)))
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Equal(t, "60", second.Header().Get("Retry-After"))
	require.Contains(t, second.Body.String(), `"error":"temporarily_unavailable"`)

	unavailable := oauthGlobalRateLimit(Config{}, "unavailable-test", 1, time.Minute)(next)
	recorder := httptest.NewRecorder()
	require.NoError(t, unavailable(context.Background(), recorder, httptest.NewRequest(http.MethodPost, "/oauth/register", nil)))
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestMCPRateLimitIsPerDurableGrant(t *testing.T) {
	t.Parallel()
	store := &memoryOAuthStore{items: make(map[string][]byte)}
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", Cache: store})
	next := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})
	limited := mcpauth.RequireBearerToken(
		func(context.Context, string, *http.Request) (*mcpauth.TokenInfo, error) {
			return &mcpauth.TokenInfo{
				Scopes: []string{mcpScope}, Expiration: time.Now().Add(time.Minute),
				Extra: map[string]any{"grant_id": "grant-123"},
			}, nil
		},
		&mcpauth.RequireBearerTokenOptions{Scopes: []string{mcpScope}},
	)(handler.mcpGrantRateLimit(next, 1, time.Minute))

	request := func() *http.Request {
		value := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		value.Header.Set("Authorization", "Bearer opaque")
		return value
	}
	first := httptest.NewRecorder()
	limited.ServeHTTP(first, request())
	require.Equal(t, http.StatusNoContent, first.Code)

	second := httptest.NewRecorder()
	limited.ServeHTTP(second, request())
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Equal(t, "60", second.Header().Get("Retry-After"))
	require.NotContains(t, second.Body.String(), "opaque")
}
