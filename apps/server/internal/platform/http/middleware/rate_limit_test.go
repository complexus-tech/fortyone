package mid

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPublicFeedbackRateLimitUsesSignedAnonymousIdentity(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	store := &rateLimitStoreStub{count: 1}
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-item", AuthenticatedLimit: 10, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret", Now: func() time.Time { return now },
	}
	nextCalled := false
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		nextCalled = true
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.SetPathValue("portalSlug", "city-roads")
	setFeedbackIngressProof(request, "test-secret", "city-roads", strings.Repeat("ab", 32), now)

	recorder := httptest.NewRecorder()
	require.NoError(t, handler(request.Context(), recorder, request))
	require.True(t, nextCalled)
	require.Equal(t, "rate-limit:feedback-item:portal:city-roads:anonymous:"+strings.Repeat("ab", 32), store.key)
	require.Equal(t, []string{
		"rate-limit:feedback-item:anonymous:" + strings.Repeat("ab", 32),
		"rate-limit:feedback-item:portal:city-roads:anonymous:" + strings.Repeat("ab", 32),
	}, store.keys)
	require.Equal(t, []string{
		`"feedback-item-global";q=12;w=3600`,
		`"feedback-item-portal";q=3;w=3600`,
	}, recorder.Header().Values("RateLimit-Policy"))
	require.Equal(t, []string{
		`"feedback-item-global";r=11;t=3600`,
		`"feedback-item-portal";r=2;t=3600`,
	}, recorder.Header().Values("RateLimit"))
}

func TestPublicFeedbackRateLimitUsesPortalAndUser(t *testing.T) {
	store := &rateLimitStoreStub{count: 1}
	config := PublicFeedbackRateLimitConfig{Scope: "feedback-item", AuthenticatedLimit: 10, AnonymousLimit: 3, AnonymousGlobalLimit: 12, Window: time.Hour, IngressSecret: "test-secret"}
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error { return nil })
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.SetPathValue("portalSlug", "city-roads")
	userID := uuid.New()
	ctx := platformauth.SetUserID(request.Context(), userID)

	require.NoError(t, handler(ctx, httptest.NewRecorder(), request))
	require.Equal(t, "rate-limit:feedback-item:portal:city-roads:user:"+userID.String(), store.key)
}

func TestPublicFeedbackRateLimitUsesValidatedContributorSessionWithoutIngressProof(t *testing.T) {
	store := &rateLimitStoreStub{count: 1}
	contributorID := uuid.NewString()
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-comment", AuthenticatedLimit: 60, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret",
		ContributorResolver: func(_ context.Context, portalSlug, authorization string) (string, error) {
			require.Equal(t, "city-roads", portalSlug)
			require.Equal(t, "FeedbackSession opaque", authorization)
			return contributorID, nil
		},
	}
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error { return nil })
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items/item/comments", nil)
	request.SetPathValue("portalSlug", "city-roads")
	request.Header.Set("Authorization", "FeedbackSession opaque")

	require.NoError(t, handler(request.Context(), httptest.NewRecorder(), request))
	require.Equal(t, "rate-limit:feedback-comment:portal:city-roads:contributor:"+contributorID, store.key)
	require.Len(t, store.keys, 1, "validated contributors must not consume the anonymous global limit")
}

func TestPublicFeedbackRateLimitRejectsInvalidContributorSessionWithoutAnonymousFallback(t *testing.T) {
	store := &rateLimitStoreStub{}
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-comment", AuthenticatedLimit: 60, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret",
		ContributorResolver: func(context.Context, string, string) (string, error) {
			return "", errors.New("invalid session")
		},
	}
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items/item/comments", nil)
	request.SetPathValue("portalSlug", "city-roads")
	request.Header.Set("Authorization", "FeedbackSession attacker-controlled")
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(request.Context(), recorder, request))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Empty(t, store.keys)
}

func TestPublicFeedbackRateLimitRejectsMissingOrExpiredIngressProof(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-item", AuthenticatedLimit: 10, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret", Now: func() time.Time { return now },
	}
	handler := PublicFeedbackRateLimit(nil, &rateLimitStoreStub{}, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	for _, signedAt := range []time.Time{time.Time{}, now.Add(-3 * time.Minute)} {
		request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
		request.SetPathValue("portalSlug", "city-roads")
		if !signedAt.IsZero() {
			setFeedbackIngressProof(request, "test-secret", "city-roads", strings.Repeat("ab", 32), signedAt)
		}
		recorder := httptest.NewRecorder()

		require.NoError(t, handler(request.Context(), recorder, request))
		require.Equal(t, http.StatusForbidden, recorder.Code)
	}
}

func TestFeedbackIngressProofMatchesNextGoldenVector(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.Header.Set(feedbackIngressVersionHeader, "v1")
	request.Header.Set(feedbackIngressIdentityHeader, "4de3dc762d2ec7ee7c2bdc0c9dccb06a6575d6baf14fd7780977fd7fb4e3c8d2")
	request.Header.Set(feedbackIngressTimestampHeader, "1800000000")
	request.Header.Set(feedbackIngressSignatureHeader, "e46dd5329fcf300bda60d49584a0637ae2107c78da79d57718901911d0acdb4b")

	fingerprint, err := validateFeedbackIngressProof(
		request,
		"city-roads",
		"test-feedback-ingress-secret-32-bytes",
		now,
	)

	require.NoError(t, err)
	require.Equal(t, "4de3dc762d2ec7ee7c2bdc0c9dccb06a6575d6baf14fd7780977fd7fb4e3c8d2", fingerprint)
}

func TestFeedbackIngressProofRejectsTamperingAndFutureTimestamps(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	tests := map[string]func(*http.Request){
		"future timestamp": func(request *http.Request) {
			setFeedbackIngressProof(request, "test-secret", "city-roads", strings.Repeat("ab", 32), now.Add(3*time.Minute))
		},
		"malformed signature": func(request *http.Request) {
			setFeedbackIngressProof(request, "test-secret", "city-roads", strings.Repeat("ab", 32), now)
			request.Header.Set(feedbackIngressSignatureHeader, "not-hex")
		},
		"wrong portal": func(request *http.Request) {
			setFeedbackIngressProof(request, "test-secret", "other-portal", strings.Repeat("ab", 32), now)
		},
		"wrong secret": func(request *http.Request) {
			setFeedbackIngressProof(request, "wrong-secret", "city-roads", strings.Repeat("ab", 32), now)
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
			mutate(request)

			_, err := validateFeedbackIngressProof(request, "city-roads", "test-secret", now)

			require.Error(t, err)
		})
	}
}

func TestPublicFeedbackRateLimitCapsAnonymousTrafficAcrossPortals(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	store := &rateLimitStoreStub{count: 13}
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-item", AuthenticatedLimit: 10, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret", Now: func() time.Time { return now },
	}
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/city-roads/feedback/items", nil)
	request.SetPathValue("portalSlug", "city-roads")
	setFeedbackIngressProof(request, "test-secret", "city-roads", strings.Repeat("ab", 32), now)
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(request.Context(), recorder, request))
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.Equal(t, []string{"rate-limit:feedback-item:anonymous:" + strings.Repeat("ab", 32)}, store.keys)
}

func TestPublicFeedbackRateLimitRejectsInvalidPortalSlugBeforeAllocatingKey(t *testing.T) {
	config := PublicFeedbackRateLimitConfig{
		Scope: "feedback-item", AuthenticatedLimit: 10, AnonymousLimit: 3, AnonymousGlobalLimit: 12,
		Window: time.Hour, IngressSecret: "test-secret",
	}
	store := &rateLimitStoreStub{}
	handler := PublicFeedbackRateLimit(nil, store, config)(func(context.Context, http.ResponseWriter, *http.Request) error {
		t.Fatal("next handler should not be called")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/portals/invalid/feedback/items", nil)
	request.SetPathValue("portalSlug", "../../users")
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(request.Context(), recorder, request))
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Empty(t, store.keys)
}

func setFeedbackIngressProof(request *http.Request, secret, portalSlug, fingerprint string, signedAt time.Time) {
	timestamp := strconv.FormatInt(signedAt.Unix(), 10)
	request.Header.Set(feedbackIngressVersionHeader, feedbackIngressVersion)
	request.Header.Set(feedbackIngressIdentityHeader, fingerprint)
	request.Header.Set(feedbackIngressTimestampHeader, timestamp)
	request.Header.Set(feedbackIngressSignatureHeader, hex.EncodeToString(feedbackIngressSignature(secret, portalSlug, fingerprint, timestamp)))
}

type rateLimitStoreStub struct {
	count int64
	err   error
	key   string
	keys  []string
	ttl   time.Duration
}

func (s *rateLimitStoreStub) IncrementWithTTL(_ context.Context, key string, ttl time.Duration) (int64, error) {
	s.key = key
	s.keys = append(s.keys, key)
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
	require.Equal(t, `"feedback-item";q=10;w=3600`, recorder.Header().Get("RateLimit-Policy"))
	require.Equal(t, `"feedback-item";r=0;t=3600`, recorder.Header().Get("RateLimit"))
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
