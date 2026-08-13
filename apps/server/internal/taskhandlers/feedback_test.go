package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type feedbackContributorDeliveryStoreStub struct {
	delivery    feedbackContributorDeliveryData
	deliverable bool
	err         error
	claimedID   uuid.UUID
}

func (s *feedbackContributorDeliveryStoreStub) ClaimContributorDelivery(_ context.Context, deliveryID uuid.UUID) (feedbackContributorDeliveryData, bool, error) {
	s.claimedID = deliveryID
	return s.delivery, s.deliverable, s.err
}

type feedbackOutboxProcessorStub struct {
	called bool
	err    error
}

func (p *feedbackOutboxProcessorStub) DispatchReadyOutboxEvents(context.Context) error {
	p.called = true
	return p.err
}

func TestHandleFeedbackOutboxDispatchDelegatesToProcessor(t *testing.T) {
	t.Parallel()
	processor := &feedbackOutboxProcessorStub{}
	handler := &handlers{feedbackOutbox: processor}
	require.NoError(t, handler.HandleFeedbackOutboxDispatch(context.Background(), asynq.NewTask("test", nil)))
	require.True(t, processor.called)

	processor.err = errors.New("database unavailable")
	err := handler.HandleFeedbackOutboxDispatch(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "database unavailable")
}

func TestFeedbackUnsubscribeURLIsTokenFreeExceptPreferenceExchange(t *testing.T) {
	t.Parallel()
	result, err := feedbackUnsubscribeURL(
		"https://app.example.com/portal/roads/feedback/dark-mode?source=email#status",
		"roads and paths",
		"opaque+/token",
	)
	require.NoError(t, err)
	require.Equal(t, "https://app.example.com/portal/roads%20and%20paths/feedback/preferences/exchange?token=opaque%2B%2Ftoken", result)
}

func TestFeedbackDeliveryRecoveryReconstructsHashedUnsubscribeToken(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.New()
	token, hash, err := feedbacksecurity.DeriveUnsubscribeToken("auth-secret", deliveryID)
	require.NoError(t, err)

	payload, err := feedbackDeliveryRecoveryPayload("auth-secret", deliveryID, hash)
	require.NoError(t, err)
	require.Equal(t, deliveryID, payload.DeliveryID)
	require.Equal(t, token, payload.UnsubscribeToken)

	_, err = feedbackDeliveryRecoveryPayload("wrong-secret", deliveryID, hash)
	require.Error(t, err)
}

func TestFeedbackUnsubscribeURLRejectsRelativeDestinations(t *testing.T) {
	t.Parallel()
	_, err := feedbackUnsubscribeURL("/portal/roads/feedback/item", "roads", "token")
	require.Error(t, err)
}

func TestFeedbackDeliveryClaimSuppressesBlockedOrUnsubscribedRecipients(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.New()
	store := &feedbackContributorDeliveryStoreStub{deliverable: false}
	handler := &handlers{feedbackDeliveries: store}
	payload, err := json.Marshal(tasks.FeedbackContributorDeliveryPayload{
		DeliveryID: deliveryID, UnsubscribeToken: "opaque-token",
	})
	require.NoError(t, err)

	require.NoError(t, handler.HandleFeedbackContributorDelivery(context.Background(), asynq.NewTask("feedback:delivery", payload)))
	require.Equal(t, deliveryID, store.claimedID)

	for _, contract := range []string{
		"contributor.blocked_at IS NULL",
		"preference.email_unsubscribed_at IS NULL",
		"CASE WHEN candidate.eligible THEN 'processing' ELSE 'suppressed' END",
		"FOR UPDATE OF delivery",
	} {
		require.True(t, strings.Contains(feedbackContributorDeliveryClaimQuery, contract), contract)
	}
}
