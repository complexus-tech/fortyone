package notificationsrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateNotificationQueryPreservesExistingDeliveryStateOnReplay(t *testing.T) {
	require.Contains(t, createNotificationQuery, "ON CONFLICT (dedupe_key) DO NOTHING")
	require.Contains(t, getNotificationByDedupeKeyQuery, "WHERE dedupe_key = $1 AND recipient_id = $2")
	require.NotContains(t, createNotificationQuery, "read_at =")
	require.NotContains(t, createNotificationQuery, "email_sent_at =")
	require.NotContains(t, createNotificationQuery, "created_at =")
}
