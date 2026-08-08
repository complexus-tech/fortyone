package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlackWebClientHonorsRetryAfter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	err := client.callJSON(context.Background(), "xoxb-test", "chat.postMessage", map[string]string{"text": "hello"}, nil)
	var rateLimit *RateLimitError
	if !errors.As(err, &rateLimit) {
		t.Fatalf("callJSON() error = %v, want RateLimitError", err)
	}
	if rateLimit.RetryAfter != 7*time.Second {
		t.Fatalf("RetryAfter = %s, want 7s", rateLimit.RetryAfter)
	}
}

func TestSlackWebClientRejectsOKFalse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	if err := client.callJSON(context.Background(), "xoxb-test", "auth.test", map[string]string{}, nil); err == nil {
		t.Fatal("callJSON() error = nil, want Slack envelope error")
	}
}
