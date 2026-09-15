package notifications

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationMessageIdentityReferencesPersistWithoutChangingPublicMessage(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	message := NotificationMessage{
		Template: "{actor} reassigned task to {assignee}",
		Variables: map[string]Variable{
			"actor":    {Value: "alex", Type: "actor"},
			"assignee": {Value: "alex", Type: "assignee"},
		},
		IdentityReferences: map[string]uuid.UUID{"assignee": assigneeID},
	}
	stored, err := json.Marshal(message)
	require.NoError(t, err)
	var persisted map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(stored, &persisted))
	require.JSONEq(t, `{"assignee":"`+assigneeID.String()+`"}`, string(persisted["identityReferences"]))

	var decoded NotificationMessage
	require.NoError(t, json.Unmarshal(stored, &decoded))
	require.Equal(t, message, decoded)

	public := (Notification{EntityType: EntityTypeStory, Message: decoded}).Public()
	publicJSON, err := json.Marshal(public)
	require.NoError(t, err)
	require.Nil(t, public.Message.IdentityReferences)
	require.NotContains(t, string(publicJSON), "identityReferences")
	require.NotContains(t, string(publicJSON), assigneeID.String())
	require.Equal(t, message.Template, public.Message.Template)
	require.Equal(t, message.Variables, public.Message.Variables)
	require.Equal(t, assigneeID, decoded.IdentityReferences["assignee"], "public redaction must not erase persisted attribution")
}

func TestLegacyNotificationMessageOmitsIdentityReferences(t *testing.T) {
	t.Parallel()

	var message NotificationMessage
	require.NoError(t, json.Unmarshal([]byte(`{"template":"Task updated","variables":{}}`), &message))
	require.Nil(t, message.IdentityReferences)
	encoded, err := json.Marshal(message)
	require.NoError(t, err)
	require.JSONEq(t, `{"template":"Task updated","variables":{}}`, string(encoded))
}

func TestNotificationMessagePublicRemovesAllInternalSnapshots(t *testing.T) {
	t.Parallel()

	message := NotificationMessage{
		Template:           "Internal snapshot",
		Strategy:           &StrategyNotificationSnapshot{Version: 1},
		IdentityReferences: map[string]uuid.UUID{"assignee": uuid.New()},
	}
	public := message.Public()
	require.Nil(t, public.Strategy)
	require.Nil(t, public.IdentityReferences)
	require.NotNil(t, message.Strategy)
	require.NotEmpty(t, message.IdentityReferences)
}
