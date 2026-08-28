package safehttp

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewRejectsUnboundedConfiguration(t *testing.T) {
	t.Parallel()
	tests := []Config{
		{Timeout: 31 * time.Second},
		{Timeout: time.Second, TLSHandshakeTimeout: 2 * time.Second},
		{MaxResponseBytes: (1 << 20) + 1},
		{MaxResponseHeaderBytes: (64 << 10) + 1},
	}
	for _, config := range tests {
		if _, err := New(config); !errors.Is(err, ErrUnsupportedRequest) {
			t.Fatalf("New(%+v) error = %v, want ErrUnsupportedRequest", config, err)
		}
	}
}

func TestParseRetryAfterIsBounded(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	if got := parseRetryAfter("30", now); got != 30*time.Second {
		t.Fatalf("seconds Retry-After = %s", got)
	}
	if got := parseRetryAfter("86401", now); got != 0 {
		t.Fatalf("unbounded Retry-After = %s", got)
	}
	if got := parseRetryAfter(now.Add(5*time.Minute).Format(http.TimeFormat), now); got != 5*time.Minute {
		t.Fatalf("date Retry-After = %s", got)
	}
	if got := parseRetryAfter("not-a-date", now); got != 0 {
		t.Fatalf("invalid Retry-After = %s", got)
	}
}
