package tasks

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestFigmaWebhookPayloadContainsOnlyInboxIdentity(t *testing.T) {
	t.Parallel()

	inboxID := uuid.New()
	payload, err := json.Marshal(FigmaWebhookPayload{InboxID: inboxID})
	if err != nil {
		t.Fatalf("marshal Figma webhook payload: %v", err)
	}
	if got, want := string(payload), `{"inboxId":"`+inboxID.String()+`"}`; got != want {
		t.Fatalf("payload = %s, want %s", got, want)
	}
}
