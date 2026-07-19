package notificationsrepository

import (
	"encoding/json"
	"testing"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDefaultPreferencesIncludeDigestChannels(t *testing.T) {
	preferences := getDefaultPreferences()

	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["reminders"])
	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["weekly_digest"])
	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["feedback_comment"])
	require.Equal(t, map[string]bool{"email": true, "in_app": true}, preferences["feedback_status_update"])
}

func TestToDBNewNotificationUsesExplicitEventDedupeKey(t *testing.T) {
	input := notifications.CoreNewNotification{
		DedupeKey:   "feedback-comment:comment-id:recipient-id",
		RecipientID: uuid.New(),
		WorkspaceID: uuid.New(),
		Type:        "feedback_comment",
		EntityType:  "feedback",
		EntityID:    uuid.New(),
		ActorID:     uuid.New(),
	}

	result, err := toDBNewNotification(input)

	require.NoError(t, err)
	require.Equal(t, input.DedupeKey, result.DedupeKey)
}

func TestToDBNewNotificationDoesNotCollapseDistinctLegacyEvents(t *testing.T) {
	input := notifications.CoreNewNotification{
		RecipientID: uuid.New(),
		WorkspaceID: uuid.New(),
		Type:        "story_update",
		EntityType:  "story",
		EntityID:    uuid.New(),
		ActorID:     uuid.New(),
	}

	first, err := toDBNewNotification(input)
	require.NoError(t, err)
	second, err := toDBNewNotification(input)

	require.NoError(t, err)
	require.NotEmpty(t, first.DedupeKey)
	require.NotEmpty(t, second.DedupeKey)
	require.NotEqual(t, first.DedupeKey, second.DedupeKey)
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
	require.Equal(t, true, corePreferences.Preferences["feedback_comment"].(map[string]interface{})["email"])
}
