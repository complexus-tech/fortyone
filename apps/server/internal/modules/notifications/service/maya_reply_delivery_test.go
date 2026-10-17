package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mayaReplyStoreStub struct {
	Repository
	inputs []domain.NewNotification
	err    error
}

func (s *mayaReplyStoreStub) DeferMayaReply(_ context.Context, input domain.NewNotification, _ uuid.UUID) error {
	s.inputs = append(s.inputs, input)
	return s.err
}

func (*mayaReplyStoreStub) DispatchMayaReplies(context.Context, int, func(context.Context, domain.NewNotification) error) (int, error) {
	return 0, nil
}

func TestMayaReplyDeferralAcceptsOnlyBoundReplyEvents(t *testing.T) {
	t.Parallel()
	commentID, recipientID := uuid.New(), uuid.New()
	input := domain.NewNotification{
		DedupeKey:  "maya-reply:" + commentID.String() + ":" + recipientID.String(),
		OccurredAt: time.Now(), RecipientID: recipientID, WorkspaceID: uuid.New(), EntityID: uuid.New(), ActorID: uuid.New(),
		Type: domain.NotificationTypeCommentReply, EntityType: domain.EntityTypeStory, Title: "Story",
		Message: domain.NotificationMessage{Variables: map[string]domain.Variable{"commentId": {Value: commentID.String(), Type: "comment"}}},
	}
	store := &mayaReplyStoreStub{}
	service := New(nil, store, nil, nil)
	require.NoError(t, service.DeferMayaReply(t.Context(), input, commentID))
	require.Equal(t, []domain.NewNotification{input}, store.inputs, "enqueue must not invoke realtime, email, push, or immediate Create")
	for _, modify := range []func(*domain.NewNotification){
		func(n *domain.NewNotification) { n.Type = domain.NotificationTypeMention },
		func(n *domain.NewNotification) { n.DedupeKey = "arbitrary" },
		func(n *domain.NewNotification) { n.RecipientID = uuid.New() },
		func(n *domain.NewNotification) { n.OccurredAt = time.Time{} },
		func(n *domain.NewNotification) { enabled := true; n.InAppEnabled = &enabled },
	} {
		invalid := input
		modify(&invalid)
		require.ErrorIs(t, service.DeferMayaReply(t.Context(), invalid, commentID), domain.ErrInvalid)
	}
	require.ErrorIs(t, service.DeferMayaReply(t.Context(), input, uuid.New()), domain.ErrInvalid)
	require.Len(t, store.inputs, 1)
	store.err = errors.New("database unavailable")
	require.ErrorIs(t, service.DeferMayaReply(t.Context(), input, commentID), store.err, "invocation must retry until its delivery decision is durable")
}

func TestMayaReplyOutboxRetriesRealtimeAfterInboxAlreadyCommitted(t *testing.T) {
	t.Parallel()
	input, notification := validNotification()
	store := &repositoryStub{created: notification, inserted: false}
	queue := &taskStub{}
	service := newTestService(store, queue)
	published := 0
	service.publishRealtime = func(_ context.Context, n CoreNotification) error {
		require.Equal(t, notification.ID, n.ID)
		published++
		if published == 1 {
			return errors.New("Redis temporarily unavailable")
		}
		return nil
	}
	require.Error(t, service.deliverMayaReply(t.Context(), input))
	require.Empty(t, queue.pushPayloads)
	require.Empty(t, queue.payloads)
	require.NoError(t, service.deliverMayaReply(t.Context(), input))
	require.Equal(t, 2, published)
	require.Len(t, queue.pushPayloads, 1)
	require.Len(t, queue.payloads, 1)
}

func TestMayaReplyRetryHonorsLiveInAppOptOut(t *testing.T) {
	t.Parallel()
	input, notification := validNotification()
	disabled := false
	input.InAppEnabled = &disabled
	queue := &taskStub{}
	service := newTestService(&repositoryStub{created: notification, inserted: false}, queue)
	service.publishRealtime = func(context.Context, CoreNotification) error {
		t.Fatal("live preference must override the persisted in-app snapshot on retry")
		return nil
	}
	require.NoError(t, service.deliverMayaReply(t.Context(), input))
	require.Empty(t, queue.pushPayloads)
	require.Len(t, queue.payloads, 1, "email retains its own independently checked preference")
}
