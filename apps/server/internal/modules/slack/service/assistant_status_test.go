package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackAssistantStatusSetterUsesNativeThreadStatus(t *testing.T) {
	t.Parallel()

	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assistant.threads.setStatus" {
			t.Errorf("request path = %q", r.URL.Path)
		}
		if authorization := r.Header.Get("Authorization"); authorization != "Bearer xoxb-test" {
			t.Errorf("Authorization = %q", authorization)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	setter := &slackAssistantStatusSetter{client: client}
	if err := setter.SetStatus(context.Background(), "xoxb-test", "C123", "171.200", slackAssistantThinkingStatus); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	want := map[string]any{
		"channel_id": "C123",
		"thread_ts":  "171.200",
		"status":     slackAssistantThinkingStatus,
		"username":   slackAssistantStatusUsername,
	}
	for key, expected := range want {
		if payload[key] != expected {
			t.Errorf("payload[%q] = %#v, want %#v", key, payload[key], expected)
		}
	}
}
