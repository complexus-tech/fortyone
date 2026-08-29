package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type feedbackContributorDeliveryStoreStub struct {
	delivery    feedback.CoreClaimedContributorDelivery
	deliverable bool
	err         error
	claimedID   uuid.UUID
	failure     feedback.CoreContributorDeliveryFailure
}

func (s *feedbackContributorDeliveryStoreStub) ClaimContributorDelivery(_ context.Context, deliveryID uuid.UUID) (feedback.CoreClaimedContributorDelivery, bool, error) {
	s.claimedID = deliveryID
	return s.delivery, s.deliverable, s.err
}

func (s *feedbackContributorDeliveryStoreStub) MarkContributorDeliverySent(context.Context, uuid.UUID) error {
	return nil
}

func (s *feedbackContributorDeliveryStoreStub) MarkContributorDeliveryFailed(_ context.Context, failure feedback.CoreContributorDeliveryFailure) error {
	s.failure = failure
	return nil
}

func (s *feedbackContributorDeliveryStoreStub) ListRecoverableContributorDeliveries(context.Context, int) ([]feedback.CoreRecoverableContributorDelivery, error) {
	return nil, nil
}

type failingFeedbackMailer struct {
	err error
}

func (m failingFeedbackMailer) Send(context.Context, mailer.Email) error {
	return m.err
}

func (m failingFeedbackMailer) SendTemplated(context.Context, mailer.TemplatedEmail) error {
	return m.err
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

func TestFeedbackDeliveryReconstructsAndVerifiesHashedUnsubscribeToken(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.New()
	token, hash, err := feedbacksecurity.DeriveUnsubscribeToken("auth-secret", deliveryID)
	require.NoError(t, err)

	reconstructed, err := feedbackUnsubscribeToken("auth-secret", deliveryID, hash)
	require.NoError(t, err)
	require.Equal(t, token, reconstructed)

	_, err = feedbackUnsubscribeToken("wrong-secret", deliveryID, hash)
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
		DeliveryID: deliveryID,
	})
	require.NoError(t, err)

	require.NoError(t, handler.HandleFeedbackContributorDelivery(context.Background(), asynq.NewTask("feedback:delivery", payload)))
	require.Equal(t, deliveryID, store.claimedID)

}

func TestFeedbackDeliveryFailsClosedWithoutStore(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(tasks.FeedbackContributorDeliveryPayload{DeliveryID: uuid.New()})
	require.NoError(t, err)

	err = (&handlers{}).HandleFeedbackContributorDelivery(context.Background(), asynq.NewTask("feedback:delivery", payload))
	require.ErrorContains(t, err, "store is unavailable")
}

func TestFeedbackDeliveryFailureDoesNotPersistOrReturnProviderDetail(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.New()
	_, tokenHash, err := feedbacksecurity.DeriveUnsubscribeToken("auth-secret", deliveryID)
	require.NoError(t, err)
	store := &feedbackContributorDeliveryStoreStub{
		deliverable: true,
		delivery: feedback.CoreClaimedContributorDelivery{
			ID:             deliveryID,
			RecipientEmail: "contributor@example.com",
			DisplayName:    "Contributor",
			PortalSlug:     "workspace",
			Subject:        "Feedback changed",
			Message:        "A status changed",
			DestinationURL: "https://workspace.fortyone.app/portal/workspace/feedback/item",
			TokenHash:      tokenHash,
		},
	}
	providerDetail := "smtp rejected contributor@example.com with credential token=secret"
	handler := &handlers{
		feedbackDeliveries:  store,
		feedbackSecurityKey: "auth-secret",
		mailerService:       failingFeedbackMailer{err: errors.New(providerDetail)},
	}
	payload, err := json.Marshal(tasks.FeedbackContributorDeliveryPayload{DeliveryID: deliveryID})
	require.NoError(t, err)

	err = handler.HandleFeedbackContributorDelivery(context.Background(), asynq.NewTask("feedback:delivery", payload))
	require.Error(t, err)
	require.NotContains(t, err.Error(), providerDetail)
	require.Equal(t, feedbackDeliveryFailureReason, store.failure.Reason)
	require.NotContains(t, store.failure.Reason, "contributor@example.com")
	require.NotContains(t, store.failure.Reason, "secret")
}
