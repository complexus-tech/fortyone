package tasks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlackEventTaskID(t *testing.T) {
	t.Parallel()

	initial := slackEventTaskID("T1", "Ev123", 0)
	recovery := slackEventTaskID("T1", "Ev123", 7)

	require.Equal(t, initial, slackEventTaskID(" T1 ", " Ev123 ", 0))
	require.Equal(t, recovery, slackEventTaskID("T1", "Ev123", 7))
	require.NotEqual(t, initial, recovery)
	require.NotEqual(t, recovery, slackEventTaskID("T1", "Ev123", 8))
	require.NotEqual(t, recovery, slackEventTaskID("T2", "Ev123", 7))
}

func TestSlackEventPayloadContainsNoProviderMessageBody(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(SlackEventPayload{ExternalWorkspaceID: "T1", EventID: "Ev123", RecoveryAttempt: 2})
	require.NoError(t, err)
	require.JSONEq(t, `{"externalWorkspaceId":"T1","eventId":"Ev123","recoveryAttempt":2}`, string(payload))
	require.NotContains(t, strings.ToLower(string(payload)), "body")
}
