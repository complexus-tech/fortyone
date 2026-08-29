package web

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSetRateLimitHeadersAdvertisesBoundedFixedWindow(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	SetRateLimitHeaders(recorder, "Feedback Item / Global", 12, 11, time.Hour)
	SetRateLimitHeaders(recorder, "Feedback Item / Portal", 3, -4, 1500*time.Millisecond)

	policyHeaders := recorder.Header().Values("RateLimit-Policy")
	if len(policyHeaders) != 2 {
		t.Fatalf("RateLimit-Policy values = %v", policyHeaders)
	}
	if policyHeaders[0] != `"feedback-item-global";q=12;w=3600` {
		t.Fatalf("first RateLimit-Policy = %q", policyHeaders[0])
	}
	if policyHeaders[1] != `"feedback-item-portal";q=3;w=2` {
		t.Fatalf("second RateLimit-Policy = %q", policyHeaders[1])
	}

	limitHeaders := recorder.Header().Values("RateLimit")
	if len(limitHeaders) != 2 {
		t.Fatalf("RateLimit values = %v", limitHeaders)
	}
	if limitHeaders[0] != `"feedback-item-global";r=11;t=3600` {
		t.Fatalf("first RateLimit = %q", limitHeaders[0])
	}
	if limitHeaders[1] != `"feedback-item-portal";r=0;t=2` {
		t.Fatalf("second RateLimit = %q", limitHeaders[1])
	}
}

func TestSetRateLimitHeadersNormalizesUntrustedPolicyNamesAndClampsRemaining(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	SetRateLimitHeaders(
		recorder,
		"\r\nAuthorization: Bearer secret "+strings.Repeat("x", 100),
		10,
		99,
		time.Minute,
	)

	policy := recorder.Header().Get("RateLimit-Policy")
	if strings.ContainsAny(policy, "\r\n") || len(policy) > 96 {
		t.Fatalf("unsafe RateLimit-Policy = %q", policy)
	}
	if got := recorder.Header().Get("RateLimit"); !strings.Contains(got, ";r=10;") {
		t.Fatalf("RateLimit remaining quota was not clamped: %q", got)
	}
}

func TestSetRateLimitHeadersIgnoresInvalidQuota(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	SetRateLimitHeaders(recorder, "invalid", 0, 0, time.Minute)
	SetRateLimitHeaders(recorder, "invalid", 10, 0, 0)

	if len(recorder.Header()) != 0 {
		t.Fatalf("invalid quota produced headers: %v", recorder.Header())
	}
}
