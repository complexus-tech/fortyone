package mid

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type rateLimitStoreStub struct {
	count int64
	err   error
	key   string
	ttl   time.Duration
}

func (s *rateLimitStoreStub) IncrementWithTTL(_ context.Context, key string, ttl time.Duration) (int64, error) {
	s.key = key
	s.ttl = ttl
	return s.count, s.err
}

func TestAuthenticatedUserRateLimitAllowsRequestWithinLimit(t *testing.T) {
	store := &rateLimitStoreStub{count: 3}
	config := AuthenticatedUserRateLimitConfig{Scope: "feedback-item", Limit: 10, Window: time.Hour}
	nextCalled := false
	next := func(context.Context, http.ResponseWriter, *http.Request) error {
		nextCalled = true
		return nil
	}
	handler := AuthenticatedUserRateLimit(nil, store, config)(next)
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.SetPathValue("portalSlug", "city-roads")
	userID := uuid.New()
	ctx := platformauth.SetUserID(request.Context(), userID)

	err := handler(ctx, httptest.NewRecorder(), request)

	require.NoError(t, err)
	require.True(t, nextCalled)
	require.Equal(t, "rate-limit:feedback-item:portal:city-roads:user:"+userID.String(), store.key)
	require.Equal(t, time.Hour, store.ttl)
}

func TestAuthenticatedUserRateLimitRejectsRequestAboveLimit(t *testing.T) {
	store := &rateLimitStoreStub{count: 11}
	config := AuthenticatedUserRateLimitConfig{Scope: "feedback-item", Limit: 10, Window: time.Hour}
	handler := AuthenticatedUserRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.SetPathValue("portalSlug", "city-roads")
	ctx := platformauth.SetUserID(request.Context(), uuid.New())
	recorder := httptest.NewRecorder()

	err := handler(ctx, recorder, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.Equal(t, "3600", recorder.Header().Get("Retry-After"))
}

func TestAuthenticatedUserRateLimitFailsClosedWhenStoreFails(t *testing.T) {
	store := &rateLimitStoreStub{err: errors.New("redis unavailable")}
	config := AuthenticatedUserRateLimitConfig{Scope: "feedback-comment", Limit: 60, Window: time.Hour}
	handler := AuthenticatedUserRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items/1/comments", nil)
	request.SetPathValue("portalSlug", "city-roads")
	ctx := platformauth.SetUserID(request.Context(), uuid.New())
	recorder := httptest.NewRecorder()

	err := handler(ctx, recorder, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}
