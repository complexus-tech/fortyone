package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackAPISenderUsesNativeStandardMarkdownForAssistantText(t *testing.T) {
	t.Parallel()

	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"ts":"171.200"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	sender := &slackAPISender{client: client}
	text := "That's **WEB-545**.\n\n- **Status:** Todo"
	messageID, err := sender.Send(context.Background(), "xoxb-test", SlackOutboundMessage{
		ChannelID:        "D123",
		Text:             text,
		StandardMarkdown: true,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if messageID != "171.200" {
		t.Fatalf("message ID = %q", messageID)
	}
	if payload["markdown_text"] != text {
		t.Fatalf("markdown_text = %#v", payload["markdown_text"])
	}
	if _, exists := payload["text"]; exists {
		t.Fatalf("payload unexpectedly contains text alongside markdown_text: %#v", payload)
	}
}

func TestSlackAPISenderRejectsMarkdownWithInteractiveBlocks(t *testing.T) {
	t.Parallel()

	sender := &slackAPISender{client: newSlackWebClient(http.DefaultClient)}
	_, err := sender.Send(context.Background(), "xoxb-test", SlackOutboundMessage{
		ChannelID:        "D123",
		Text:             "Confirm this change",
		StandardMarkdown: true,
		ProviderPayload: SlackProviderPayload{Blocks: []SlackBlock{{
			Type: "section",
			Text: &SlackTextObject{Type: "mrkdwn", Text: "Confirm this change"},
		}}},
	})
	if err == nil {
		t.Fatal("Send() error = nil, want Markdown and blocks conflict")
	}
}
