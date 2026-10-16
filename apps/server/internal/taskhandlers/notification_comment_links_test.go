package taskhandlers

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCommentDigestLinksKeepTheSpecificCommentTarget(t *testing.T) {
	commentID := uuid.New()
	for _, test := range []struct {
		name      string
		commentID string
		want      string
	}{
		{"comment", commentID.String(), commentID.String()},
		{"legacy", "", ""},
		{"invalid", "not-a-uuid", ""},
		{"empty identity", uuid.Nil.String(), ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			message, err := json.Marshal(NotificationMessage{
				Template: "{actor} replied: {content}",
				Variables: map[string]Variable{
					"actor":     {Value: "Ada", Type: "actor"},
					"content":   {Value: "Ready to review", Type: "plain_text"},
					"commentId": {Value: test.commentID, Type: "comment"},
				},
			})
			require.NoError(t, err)
			storyID, notificationID := uuid.New(), uuid.New()
			input, err := buildNotificationDigestCopyInput(NotificationEmailDigestData{
				RecipientID: uuid.New(), WorkspaceID: uuid.New(), WorkspaceName: "Product",
				Items: []NotificationEmailDigestItem{{NotificationID: notificationID, NotificationType: "comment_reply", EntityType: "story", EntityID: storyID, Title: "Checkout spacing", Message: message}},
			}, "https://product.fortyone.app")
			require.NoError(t, err)
			require.NotEmpty(t, input.Fallback.Rows)
			link, err := url.Parse(input.Fallback.Rows[0].URL)
			require.NoError(t, err)
			require.Equal(t, test.want, link.Query().Get("comment"))
			require.Equal(t, storyID.String(), link.Query().Get("entityId"))
			require.Equal(t, "/notifications/"+notificationID.String(), link.Path)
			require.NotContains(t, input.Fallback.Rows[0].Text, commentID.String())
		})
	}
}
