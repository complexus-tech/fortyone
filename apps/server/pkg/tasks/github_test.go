package tasks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGitHubWebhookPayloadContainsOnlyInboxIdentity(t *testing.T) {
	t.Parallel()

	inboxID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	payload, err := json.Marshal(GitHubWebhookPayload{InboxID: inboxID})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(payload) != `{"inboxId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}` {
		t.Fatalf("payload = %s", payload)
	}
	if strings.Contains(strings.ToLower(string(payload)), "provider") ||
		strings.Contains(strings.ToLower(string(payload)), "body") {
		t.Fatalf("payload contains non-identity data: %s", payload)
	}
}
