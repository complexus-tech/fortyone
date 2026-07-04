package notificationsrepository

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultPreferencesIncludeDigestChannels(t *testing.T) {
	preferences := getDefaultPreferences()

	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["reminders"])
	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["weekly_digest"])
}

func TestToCoreNotificationPreferencesBackfillsMissingDefaults(t *testing.T) {
	dbPreferences := dbNotificationPreferences{
		Preferences: json.RawMessage(`{"story_update":{"email":false,"in_app":true}}`),
	}

	corePreferences, err := toCoreNotificationPreferences(dbPreferences)

	require.NoError(t, err)
	require.Equal(t, false, corePreferences.Preferences["story_update"].(map[string]interface{})["email"])
	require.Equal(t, true, corePreferences.Preferences["reminders"].(map[string]interface{})["email"])
	require.Equal(t, true, corePreferences.Preferences["weekly_digest"].(map[string]interface{})["email"])
}
