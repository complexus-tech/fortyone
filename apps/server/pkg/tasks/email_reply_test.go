package tasks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrevoEmailReplyTaskID(t *testing.T) {
	t.Parallel()

	initial := brevoEmailReplyTaskID("workspace-id", "message-id", 0)
	recovery := brevoEmailReplyTaskID("workspace-id", "message-id", 3)

	require.Equal(t, initial, brevoEmailReplyTaskID(" workspace-id ", " message-id ", 0))
	require.Equal(t, recovery, brevoEmailReplyTaskID("workspace-id", "message-id", 3))
	require.NotEqual(t, initial, recovery)
	require.NotEqual(t, recovery, brevoEmailReplyTaskID("workspace-id", "message-id", 4))
	require.NotEqual(t, recovery, brevoEmailReplyTaskID("other-workspace", "message-id", 3))
}

func TestBrevoEmailReplyPayloadContainsNoEmailContent(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(BrevoEmailReplyPayload{
		ExternalWorkspaceID: "workspace-id",
		EventID:             "message-id",
		RecoveryAttempt:     2,
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"externalWorkspaceId":"workspace-id","eventId":"message-id","recoveryAttempt":2}`, string(payload))
	require.NotContains(t, strings.ToLower(string(payload)), "body")
	require.NotContains(t, strings.ToLower(string(payload)), "subject")
	require.NotContains(t, strings.ToLower(string(payload)), "sender")
}
