package notifications

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentNotificationContentRemainsStructuredData(t *testing.T) {
	t.Parallel()

	const content = `<script>alert("xss")</script><img src=x onerror="alert(1)">&lt;entity&gt;`
	actorID := uuid.New()
	recipientID := uuid.New()
	workspaceID := uuid.New()
	storyID := uuid.New()
	rules := NewRules(nil, nil, nil, nil)

	tests := []struct {
		name         string
		wantTemplate string
		process      func() ([]CoreNewNotification, error)
	}{
		{
			name:         "comment",
			wantTemplate: "{actor} left a comment: {content}",
			process: func() ([]CoreNewNotification, error) {
				return rules.ProcessCommentCreated(context.Background(), events.CommentCreatedPayload{
					CommentID:   uuid.New(),
					StoryID:     storyID,
					StoryTitle:  "Test Story",
					AssigneeID:  &recipientID,
					WorkspaceID: workspaceID,
					Content:     content,
				}, actorID)
			},
		},
		{
			name:         "reply",
			wantTemplate: "{actor} replied: {content}",
			process: func() ([]CoreNewNotification, error) {
				return rules.ProcessCommentReplied(context.Background(), events.CommentRepliedPayload{
					CommentID:       uuid.New(),
					ParentCommentID: uuid.New(),
					ParentAuthorID:  recipientID,
					StoryID:         storyID,
					StoryTitle:      "Test Story",
					WorkspaceID:     workspaceID,
					Content:         content,
				}, actorID)
			},
		},
		{
			name:         "mention",
			wantTemplate: "{actor} mentioned you: {content}",
			process: func() ([]CoreNewNotification, error) {
				return rules.ProcessUserMentioned(context.Background(), events.UserMentionedPayload{
					CommentID:     uuid.New(),
					StoryID:       storyID,
					StoryTitle:    "Test Story",
					WorkspaceID:   workspaceID,
					MentionedUser: recipientID,
					Content:       content,
				}, actorID)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notifications, err := test.process()

			require.NoError(t, err)
			require.Len(t, notifications, 1)
			message := notifications[0].Message
			assert.Equal(t, test.wantTemplate, message.Template)
			assert.NotContains(t, message.Template, content)
			assert.Equal(t, Variable{Value: content, Type: "text"}, message.Variables["content"])
		})
	}
}
