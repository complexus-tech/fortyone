package mid

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type PublicFeedbackRateLimitConfig struct {
	Scope                string
	AuthenticatedLimit   int64
	AnonymousLimit       int64
	AnonymousGlobalLimit int64
	Window               time.Duration
	IngressSecret        string
	ContributorResolver  FeedbackContributorIdentityResolver
	Now                  func() time.Time
}

// FeedbackContributorIdentityResolver validates a portal-scoped contributor
// session and returns its stable, non-secret rate-limit identity.
type FeedbackContributorIdentityResolver func(context.Context, string, string) (string, error)

type RateLimitStore interface {
	IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

const (
	feedbackIngressVersion         = "v1"
	feedbackIngressIdentityHeader  = "X-FortyOne-Feedback-Identity"
	feedbackIngressSignatureHeader = "X-FortyOne-Feedback-Signature"
	feedbackIngressTimestampHeader = "X-FortyOne-Feedback-Timestamp"
	feedbackIngressVersionHeader   = "X-FortyOne-Feedback-Version"
	feedbackIngressMaxClockSkew    = 2 * time.Minute
)

var feedbackPortalSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,253}[a-z0-9]$`)

// PublicFeedbackRateLimit accepts anonymous identity only when the trusted web
// ingress signs it. This avoids collapsing all visitors into the Next.js
// server's shared RemoteAddr while refusing spoofable forwarded-IP headers.
func PublicFeedbackRateLimit(log *logger.Logger, store RateLimitStore, config PublicFeedbackRateLimitConfig) web.Middleware {
	if strings.TrimSpace(config.Scope) == "" || config.AuthenticatedLimit <= 0 || config.AnonymousLimit <= 0 || config.AnonymousGlobalLimit <= 0 || config.Window <= 0 || strings.TrimSpace(config.IngressSecret) == "" {
		panic("valid public feedback rate limit configuration is required")
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if store == nil {
				return web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}
			portalSlug := strings.ToLower(strings.TrimSpace(web.Params(r, "portalSlug")))
			if !feedbackPortalSlugPattern.MatchString(portalSlug) {
				return web.RespondError(ctx, w, errors.New("invalid feedback portal slug"), http.StatusBadRequest)
			}
			identity := ""
			limit := config.AnonymousLimit
			if userID, err := GetUserID(ctx); err == nil {
				identity = "user:" + userID.String()
				limit = config.AuthenticatedLimit
			} else if hasFeedbackSessionAuthorization(r.Header.Get("Authorization")) && config.ContributorResolver != nil {
				contributorID, err := config.ContributorResolver(ctx, portalSlug, r.Header.Get("Authorization"))
				if err != nil || strings.TrimSpace(contributorID) == "" {
					return web.RespondError(ctx, w, errors.New("invalid feedback contributor session"), http.StatusUnauthorized)
				}
				identity = "contributor:" + strings.TrimSpace(contributorID)
				limit = config.AuthenticatedLimit
			} else {
				fingerprint, err := validateFeedbackIngressProof(r, portalSlug, config.IngressSecret, now().UTC())
				if err != nil {
					return web.RespondError(ctx, w, err, http.StatusForbidden)
				}
				identity = "anonymous:" + fingerprint
				globalKey := fmt.Sprintf("rate-limit:%s:anonymous:%s", config.Scope, fingerprint)
				allowed, err := incrementPublicFeedbackRateLimit(
					ctx,
					w,
					log,
					store,
					config,
					globalKey,
					config.Scope+"-global",
					config.AnonymousGlobalLimit,
				)
				if err != nil || !allowed {
					return err
				}
			}
			key := fmt.Sprintf("rate-limit:%s:portal:%s:%s", config.Scope, portalSlug, identity)
			allowed, err := incrementPublicFeedbackRateLimit(
				ctx,
				w,
				log,
				store,
				config,
				key,
				config.Scope+"-portal",
				limit,
			)
			if err != nil || !allowed {
				return err
			}
			return next(ctx, w, r)
		}
	}
}

func hasFeedbackSessionAuthorization(value string) bool {
	parts := strings.Fields(strings.TrimSpace(value))
	return len(parts) == 2 && strings.EqualFold(parts[0], "FeedbackSession") && parts[1] != ""
}

func incrementPublicFeedbackRateLimit(
	ctx context.Context,
	w http.ResponseWriter,
	log *logger.Logger,
	store RateLimitStore,
	config PublicFeedbackRateLimitConfig,
	key string,
	policy string,
	limit int64,
) (bool, error) {
	count, err := store.IncrementWithTTL(ctx, key, config.Window)
	if err != nil {
		if log != nil {
			log.Error(ctx, "failed to enforce public feedback rate limit", "scope", config.Scope, "error", err)
		}
		return false, web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
	}
	web.SetRateLimitHeaders(w, policy, limit, limit-count, config.Window)
	if count > limit {
		retryAfterSeconds := max(int64(config.Window/time.Second), 1)
		w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
		return false, web.RespondError(ctx, w, errors.New("too many requests"), http.StatusTooManyRequests)
	}
	return true, nil
}

func validateFeedbackIngressProof(r *http.Request, portalSlug, secret string, now time.Time) (string, error) {
	if r.Header.Get(feedbackIngressVersionHeader) != feedbackIngressVersion {
		return "", errors.New("invalid feedback ingress proof")
	}
	fingerprint := strings.ToLower(strings.TrimSpace(r.Header.Get(feedbackIngressIdentityHeader)))
	fingerprintBytes, err := hex.DecodeString(fingerprint)
	if err != nil || len(fingerprintBytes) != sha256.Size {
		return "", errors.New("invalid feedback ingress proof")
	}
	timestampText := strings.TrimSpace(r.Header.Get(feedbackIngressTimestampHeader))
	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return "", errors.New("invalid feedback ingress proof")
	}
	signedAt := time.Unix(timestamp, 0).UTC()
	if signedAt.Before(now.Add(-feedbackIngressMaxClockSkew)) || signedAt.After(now.Add(feedbackIngressMaxClockSkew)) {
		return "", errors.New("expired feedback ingress proof")
	}
	providedSignature, err := hex.DecodeString(strings.TrimSpace(r.Header.Get(feedbackIngressSignatureHeader)))
	if err != nil || len(providedSignature) != sha256.Size {
		return "", errors.New("invalid feedback ingress proof")
	}
	expectedSignature := feedbackIngressSignature(secret, portalSlug, fingerprint, timestampText)
	if !hmac.Equal(providedSignature, expectedSignature) {
		return "", errors.New("invalid feedback ingress proof")
	}
	return fingerprint, nil
}

func feedbackIngressSignature(secret, portalSlug, fingerprint, timestamp string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.Join([]string{
		feedbackIngressVersion,
		strings.ToLower(strings.TrimSpace(portalSlug)),
		strings.ToLower(strings.TrimSpace(fingerprint)),
		strings.TrimSpace(timestamp),
	}, "\n")))
	return mac.Sum(nil)
}

type AuthenticatedUserRateLimitConfig struct {
	Scope  string
	Limit  int64
	Window time.Duration
}

func AuthenticatedUserRateLimit(
	log *logger.Logger,
	store RateLimitStore,
	config AuthenticatedUserRateLimitConfig,
) web.Middleware {
	if strings.TrimSpace(config.Scope) == "" {
		panic("authenticated user rate limit scope is required")
	}
	if config.Limit <= 0 {
		panic("authenticated user rate limit must be positive")
	}
	if config.Window <= 0 {
		panic("authenticated user rate limit window must be positive")
	}

	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			userID, err := GetUserID(ctx)
			if err != nil {
				return web.RespondError(ctx, w, errors.New("user not authenticated"), http.StatusUnauthorized)
			}
			if store == nil {
				if log != nil {
					log.Error(ctx, "rate limit store is not configured", "scope", config.Scope)
				}
				return web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}

			portalSlug := strings.ToLower(strings.TrimSpace(web.Params(r, "portalSlug")))
			key := fmt.Sprintf("rate-limit:%s:portal:%s:user:%s", config.Scope, portalSlug, userID)
			count, err := store.IncrementWithTTL(ctx, key, config.Window)
			if err != nil {
				if log != nil {
					log.Error(ctx, "failed to enforce request rate limit", "scope", config.Scope, "error", err)
				}
				return web.RespondError(ctx, w, errors.New("rate limit unavailable"), http.StatusServiceUnavailable)
			}
			web.SetRateLimitHeaders(w, config.Scope, config.Limit, config.Limit-count, config.Window)
			if count > config.Limit {
				retryAfterSeconds := int64(config.Window / time.Second)
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
				return web.RespondError(ctx, w, errors.New("too many requests"), http.StatusTooManyRequests)
			}

			return next(ctx, w, r)
		}
	}
}
