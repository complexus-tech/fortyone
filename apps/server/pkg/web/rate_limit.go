package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	maxRateLimitPolicyNameLength = 64
	defaultRateLimitPolicyName   = "default"
)

// SetRateLimitHeaders advertises one fixed-window request quota. Multiple
// calls append independent policies, which is useful when a request consumes
// both a global and a resource-scoped quota.
//
// The wire format follows draft-ietf-httpapi-ratelimit-headers-11. Retry-After
// remains the authoritative signal when a caller rejects a request with 429.
func SetRateLimitHeaders(
	w http.ResponseWriter,
	policy string,
	limit int64,
	remaining int64,
	window time.Duration,
) {
	if w == nil || limit <= 0 || window <= 0 {
		return
	}

	policy = normalizeRateLimitPolicyName(policy)
	remaining = max(min(remaining, limit), 0)
	windowSeconds := int64(window / time.Second)
	if window%time.Second != 0 {
		windowSeconds++
	}
	windowSeconds = max(windowSeconds, 1)

	w.Header().Add(
		"RateLimit-Policy",
		fmt.Sprintf("%q;q=%d;w=%d", policy, limit, windowSeconds),
	)
	w.Header().Add(
		"RateLimit",
		fmt.Sprintf("%q;r=%d;t=%d", policy, remaining, windowSeconds),
	)
}

func normalizeRateLimitPolicyName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var normalized strings.Builder
	normalized.Grow(min(len(value), maxRateLimitPolicyNameLength))

	separatorPending := false
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			requiredBytes := 1
			if separatorPending && normalized.Len() > 0 {
				requiredBytes++
			}
			if normalized.Len()+requiredBytes > maxRateLimitPolicyNameLength {
				return normalized.String()
			}
			if separatorPending && normalized.Len() > 0 {
				normalized.WriteByte('-')
			}
			separatorPending = false
			normalized.WriteRune(character)
		default:
			separatorPending = normalized.Len() > 0
		}
	}

	result := strings.TrimSuffix(normalized.String(), "-")
	if result == "" {
		return defaultRateLimitPolicyName
	}
	return result
}
