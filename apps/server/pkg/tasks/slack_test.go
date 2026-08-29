package tasks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSlackEventPayloadContainsNoProviderMessageBody(t *testing.T) {
	t.Parallel()

	inboxID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	payload, err := json.Marshal(SlackEventPayload{Provider: "slack", InboxID: inboxID})
	require.NoError(t, err)
	require.JSONEq(t, `{"provider":"slack","inboxId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}`, string(payload))
	require.NotContains(t, strings.ToLower(string(payload)), "body")
	require.NotContains(t, string(payload), "Ev123")
	require.NotContains(t, string(payload), "T1")
}
