package taskhandlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTaskDigestPreservesDiscussionsAndActiveScheduleIssueAlongsideLatestRoutineUpdate(t *testing.T) {
	storyID := uuid.New()
	base := time.Date(2026, time.October, 3, 9, 0, 0, 0, time.UTC)
	items := []NotificationEmailDigestItem{
		{NotificationID: uuid.New(), NotificationType: "story_update", EntityType: "story", EntityID: storyID, CreatedAt: base},
		{NotificationID: uuid.New(), NotificationType: "story_update", EntityType: "story", EntityID: storyID, CreatedAt: base.Add(time.Minute)},
		{NotificationID: uuid.New(), NotificationType: "story_update", EntityType: "story", EntityID: storyID, CreatedAt: base, Message: json.RawMessage(`{"scheduleIssue":{"id":"` + uuid.NewString() + `"}}`)},
		{NotificationID: uuid.New(), NotificationType: "story_comment", EntityType: "story", EntityID: storyID, CreatedAt: base},
		{NotificationID: uuid.New(), NotificationType: "story_mention", EntityType: "story", EntityID: storyID, CreatedAt: base},
	}

	require.Equal(t, items[1:], latestTaskDigestItems(items))
}

func TestTaskDigestProtectsLiteralPlainTextCommentValues(t *testing.T) {
	message := NotificationMessage{
		Template: "New comment: {content}",
		Variables: map[string]Variable{
			"content": {Value: "Use <example> &amp; literally", Type: "plain_text"},
		},
	}
	require.Equal(t, []string{"Use <example> &amp; literally"}, notificationVariableValues(message.Variables))
	require.Equal(t, []string{"New comment: Use <example> &amp; literally"}, notificationSemanticProtectedTokens(message, parseNotificationMessage(message).Text))
}
