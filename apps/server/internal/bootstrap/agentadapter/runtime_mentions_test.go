package agentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMayaPublishesCanonicalMentionsUnderRequesterAuthority(t *testing.T) {
	t.Parallel()
	runtime, run, _ := runtimeFixture()
	memberID, deniedID := uuid.New(), uuid.New()
	writer := &runtimeWriter{targets: []comments.MentionTarget{{ID: memberID, Name: "Alex & Co"}}}
	runtime.Comments = writer
	text := "Please review [@Forged label](mention:" + memberID.String() + "). [@Unavailable](mention:" + deniedID.String() + ")"
	id, err := runtime.Publish(t.Context(), run, agentruns.Outcome{Reply: text})
	require.NoError(t, err)
	require.Equal(t, uuid.NewSHA1(run.ID, []byte("reply")), id)
	require.Equal(t, run.RequesterID, writer.scope.Requester.PrincipalID)
	require.Equal(t, auth.PrincipalHumanUser, writer.scope.Requester.Kind)
	require.Equal(t, runtime.MayaID, writer.scope.Author.PrincipalID)
	require.Equal(t, auth.PrincipalSystem, writer.scope.Author.Kind)
	require.Equal(t, []uuid.UUID{run.TeamID}, writer.scope.Requester.TeamAccess.RestrictedTeamIDs())
	require.Equal(t, []uuid.UUID{memberID}, writer.calls[0].MentionedUserIDs)
	require.Contains(t, writer.calls[0].Content, `data-mention-kind="user"`)
	require.Contains(t, writer.calls[0].Content, "@Alex &amp; Co")
	require.Contains(t, writer.calls[0].Content, "@Unavailable")
	require.NotContains(t, writer.calls[0].Content, "Forged label")
	require.NotContains(t, writer.calls[0].Content, "mention:")
	require.NotContains(t, writer.calls[0].Content, "data-agent-key")
}

type replyNotificationWriter struct {
	inputs   []notifications.CoreNewNotification
	failAt   int
	deferred []uuid.UUID
}

func (w *replyNotificationWriter) DeferMayaReply(ctx context.Context, input notifications.CoreNewNotification, commentID uuid.UUID) error {
	w.deferred = append(w.deferred, commentID)
	_, err := w.Create(ctx, input)
	return err
}

func (w *replyNotificationWriter) Create(_ context.Context, input notifications.CoreNewNotification) (notifications.CoreNotification, error) {
	w.inputs = append(w.inputs, input)
	if len(w.inputs) == w.failAt {
		return notifications.CoreNotification{}, errors.New("delivery temporarily unavailable after persistence")
	}
	return notifications.CoreNotification{}, nil
}

func TestMayaMentionDeliveryUsesCommittedRecipientsAndStableRetryKeys(t *testing.T) {
	t.Parallel()
	runtime, run, stories := runtimeFixture()
	replyID, memberID, followerID := uuid.New(), uuid.New(), uuid.New()
	run.CreatedAt = time.Now().UTC()
	run.Payload, _ = json.Marshal(commentPayload{StoryTitle: "Review this"})
	stories.audience = []uuid.UUID{run.RequesterID, memberID, followerID}
	writer := &runtimeWriter{reply: comments.AgentReply{Comment: comments.CoreComment{
		ID: replyID, StoryID: run.EntityID, UserID: runtime.MayaID, Comment: "<p>Persisted answer for @Alex</p>",
	}, MentionedUserIDs: []uuid.UUID{run.RequesterID, memberID}}}
	notificationWriter := &replyNotificationWriter{failAt: 2}
	runtime.Comments, runtime.Notifications = writer, notificationWriter
	runtime.NotificationRules = notifications.NewRules(nil, nil, nil, nil)
	outcome := agentruns.Outcome{Reply: "Never re-render this changed model output"}
	require.Error(t, runtime.Deliver(t.Context(), run, outcome, replyID))
	require.NoError(t, runtime.Deliver(t.Context(), run, outcome, replyID))
	require.Len(t, notificationWriter.inputs, 5)
	require.Equal(t, []uuid.UUID{replyID, replyID}, notificationWriter.deferred, "only non-mention reply delivery waits for visibility")
	require.Equal(t, notificationWriter.inputs[0], notificationWriter.inputs[2])
	require.Equal(t, notificationWriter.inputs[1], notificationWriter.inputs[3])
	byRecipient := make(map[uuid.UUID]notifications.CoreNewNotification)
	for _, input := range notificationWriter.inputs[2:] {
		require.NoError(t, input.Validate())
		require.NotContains(t, byRecipient, input.RecipientID, "mention must replace a duplicate reply notification")
		byRecipient[input.RecipientID] = input
		require.Equal(t, runtime.MayaID, input.ActorID)
		require.Equal(t, "Maya", input.Message.Variables["actor"].Value)
		require.Contains(t, input.Message.Variables["content"].Value, "Persisted answer")
		require.Equal(t, replyID.String(), input.Message.Variables["commentId"].Value)
		require.Equal(t, run.CreatedAt, input.OccurredAt)
	}
	require.Equal(t, "comment_reply", string(byRecipient[followerID].Type))
	for _, id := range []uuid.UUID{run.RequesterID, memberID} {
		require.Equal(t, "mention", string(byRecipient[id].Type))
		require.Equal(t, "maya-mention:"+replyID.String()+":"+id.String(), byRecipient[id].DedupeKey)
	}
}
