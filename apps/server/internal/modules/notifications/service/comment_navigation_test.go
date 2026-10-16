package notifications

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCommentNotificationsRetainTheSpecificCommentTarget(t *testing.T) {
	rules := NewRules(nil, nil, nil, nil)
	actorID, recipientID, storyID, workspaceID, commentID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name   string
		create func() ([]CoreNewNotification, error)
	}{
		{"comment", func() ([]CoreNewNotification, error) {
			return rules.ProcessCommentCreated(context.Background(), events.CommentCreatedPayload{
				CommentID: commentID, StoryID: storyID, WorkspaceID: workspaceID, AssigneeID: &recipientID, Content: "Review this change",
			}, actorID)
		}},
		{"reply", func() ([]CoreNewNotification, error) {
			return rules.ProcessCommentReplied(context.Background(), events.CommentRepliedPayload{
				CommentID: commentID, ParentCommentID: uuid.New(), ParentAuthorID: recipientID, StoryID: storyID, WorkspaceID: workspaceID, Content: "Review this change",
			}, actorID)
		}},
		{"mention", func() ([]CoreNewNotification, error) {
			return rules.ProcessUserMentioned(context.Background(), events.UserMentionedPayload{
				CommentID: commentID, StoryID: storyID, WorkspaceID: workspaceID, MentionedUser: recipientID, Content: "Review this change",
			}, actorID)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.create()
			require.NoError(t, err)
			require.Len(t, result, 1)
			require.Equal(t, storyID, result[0].EntityID)
			require.Equal(t, Variable{Value: commentID.String(), Type: "comment"}, result[0].Message.Variables["commentId"])
			require.NotContains(t, result[0].Message.Template, "{commentId}")
		})
	}
}
