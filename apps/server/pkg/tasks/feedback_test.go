package tasks

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFeedbackContributorDeliveryPayloadContainsOnlyTheDeliveryIdentity(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.New()
	payload, err := json.Marshal(FeedbackContributorDeliveryPayload{DeliveryID: deliveryID})
	require.NoError(t, err)

	var envelope map[string]any
	require.NoError(t, json.Unmarshal(payload, &envelope))
	require.Equal(t, map[string]any{"deliveryId": deliveryID.String()}, envelope)
	require.NotContains(t, string(payload), "token")
	require.NotContains(t, string(payload), "secret")
}
