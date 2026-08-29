package usershttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	verificationRateLimitWindow = 10 * time.Minute

	verificationSendEmailLimit   int64 = 3
	verificationSendNetworkLimit int64 = 20

	verificationConfirmEmailLimit   int64 = 10
	verificationConfirmTokenLimit   int64 = 5
	verificationConfirmNetworkLimit int64 = 50
)

type verificationRateLimitStore interface {
	IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type verificationRateLimitIdentity struct {
	kind  string
	value string
	limit int64
}

func (h *Handlers) enforceVerificationRateLimits(
	ctx context.Context,
	w http.ResponseWriter,
	scope string,
	identities ...verificationRateLimitIdentity,
) (bool, error) {
	if h.verificationRateLimits == nil {
		return false, web.RespondError(ctx, w, errors.New("verification rate limit unavailable"), http.StatusServiceUnavailable)
	}

	for _, identity := range identities {
		key, err := h.users.VerificationRateLimitKey(scope, identity.kind, identity.value)
		if err != nil {
			return false, web.RespondError(ctx, w, errors.New("verification rate limit unavailable"), http.StatusServiceUnavailable)
		}
		count, err := h.verificationRateLimits.IncrementWithTTL(ctx, key, verificationRateLimitWindow)
		if err != nil {
			return false, web.RespondError(ctx, w, errors.New("verification rate limit unavailable"), http.StatusServiceUnavailable)
		}
		web.SetRateLimitHeaders(
			w,
			"verification-"+scope+"-"+identity.kind,
			identity.limit,
			identity.limit-count,
			verificationRateLimitWindow,
		)
		if count > identity.limit {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(verificationRateLimitWindow/time.Second), 10))
			return false, web.RespondError(ctx, w, errors.New("too many verification attempts"), http.StatusTooManyRequests)
		}
	}

	return true, nil
}

func verificationNetworkIdentity(r *http.Request) string {
	remoteAddress := strings.TrimSpace(r.RemoteAddr)
	host, _, err := net.SplitHostPort(remoteAddress)
	if err == nil {
		remoteAddress = host
	}
	if ip := net.ParseIP(strings.TrimSpace(remoteAddress)); ip != nil {
		return ip.String()
	}
	// A single opaque bucket is safer than placing malformed/raw address data in
	// cache keys. Real deployments provide RemoteAddr through net/http.
	return "unknown-network"
}
