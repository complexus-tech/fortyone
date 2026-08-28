package taskhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestRetentionHandlersDelegateToModuleMaintenancePorts(t *testing.T) {
	tests := []struct {
		name   string
		handle func(*RetentionHandlers) error
		assert func(*testing.T, *retentionStoreStub)
	}{
		{
			name: "verification and feedback security artifacts",
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleTokenCleanup(context.Background(), asynq.NewTask("test", nil))
			},
			assert: func(t *testing.T, store *retentionStoreStub) {
				require.Equal(t, 1, store.verificationTokenCalls)
				require.Equal(t, 1, store.feedbackArtifactCalls)
			},
		},
		{
			name: "deleted chat sessions",
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleChatSessionsCleanup(context.Background(), asynq.NewTask("test", nil))
			},
			assert: func(t *testing.T, store *retentionStoreStub) {
				require.Equal(t, 1, store.chatSessionCalls)
			},
		},
		{
			name: "terminal Stripe webhooks",
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleWebhookCleanup(context.Background(), asynq.NewTask("test", nil))
			},
			assert: func(t *testing.T, store *retentionStoreStub) {
				require.Equal(t, 1, store.stripeWebhookCalls)
			},
		},
		{
			name: "messaging retention",
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleMessagingCleanup(context.Background(), asynq.NewTask("test", nil))
			},
			assert: func(t *testing.T, store *retentionStoreStub) {
				require.Equal(t, 1, store.messagingDataCalls)
			},
		},
		{
			name: "deleted feedback",
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleDeleteFeedback(context.Background(), asynq.NewTask("test", nil))
			},
			assert: func(t *testing.T, store *retentionStoreStub) {
				require.Equal(t, 1, store.deletedFeedbackCalls)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &retentionStoreStub{}
			handlers := newTestRetentionHandlers(store)

			require.NoError(t, test.handle(handlers))
			test.assert(t, store)
		})
	}
}

func TestRetentionHandlersPreserveMaintenanceFailures(t *testing.T) {
	sentinel := errors.New("maintenance store unavailable")
	tests := []struct {
		name      string
		configure func(*retentionStoreStub)
		handle    func(*RetentionHandlers) error
	}{
		{
			name:      "verification tokens",
			configure: func(store *retentionStoreStub) { store.verificationTokenErr = sentinel },
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleTokenCleanup(context.Background(), asynq.NewTask("test", nil))
			},
		},
		{
			name:      "chat sessions",
			configure: func(store *retentionStoreStub) { store.chatSessionErr = sentinel },
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleChatSessionsCleanup(context.Background(), asynq.NewTask("test", nil))
			},
		},
		{
			name:      "Stripe webhooks",
			configure: func(store *retentionStoreStub) { store.stripeWebhookErr = sentinel },
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleWebhookCleanup(context.Background(), asynq.NewTask("test", nil))
			},
		},
		{
			name:      "messaging retention",
			configure: func(store *retentionStoreStub) { store.messagingDataErr = sentinel },
			handle: func(handlers *RetentionHandlers) error {
				return handlers.HandleMessagingCleanup(context.Background(), asynq.NewTask("test", nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &retentionStoreStub{}
			test.configure(store)

			err := test.handle(newTestRetentionHandlers(store))
			require.ErrorIs(t, err, sentinel)
		})
	}
}

func TestRetentionHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *RetentionHandlers
	err := handlers.HandleChatSessionsCleanup(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewRetentionHandlers(RetentionHandlerDependencies{Log: testTaskLogger()})
	err = handlers.HandleChatSessionsCleanup(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "chat session maintenance store is required")
}

func newTestRetentionHandlers(store *retentionStoreStub) *RetentionHandlers {
	return NewRetentionHandlers(RetentionHandlerDependencies{
		Log:                 testTaskLogger(),
		ChatSessions:        store,
		VerificationTokens:  store,
		StripeWebhookEvents: store,
		MessagingData:       store,
		Feedback:            store,
	})
}

type retentionStoreStub struct {
	chatSessionCalls       int
	verificationTokenCalls int
	inactiveUserCalls      int
	stripeWebhookCalls     int
	messagingDataCalls     int
	feedbackArtifactCalls  int
	deletedFeedbackCalls   int

	chatSessionErr       error
	verificationTokenErr error
	inactiveUserErr      error
	stripeWebhookErr     error
	messagingDataErr     error
}

func (store *retentionStoreStub) PurgeDeletedChatSessions(
	context.Context,
	time.Time,
	int,
) (int64, error) {
	store.chatSessionCalls++
	return 0, store.chatSessionErr
}

func (store *retentionStoreStub) PurgeExpiredVerificationTokens(
	context.Context,
	time.Time,
	int,
) (int64, error) {
	store.verificationTokenCalls++
	return 0, store.verificationTokenErr
}

func (store *retentionStoreStub) DeactivateInactiveUsers(
	context.Context,
	time.Time,
	time.Time,
	time.Time,
	int,
) (int64, error) {
	store.inactiveUserCalls++
	return 0, store.inactiveUserErr
}

func (store *retentionStoreStub) PurgeTerminalStripeWebhookEvents(
	context.Context,
	time.Time,
	int,
) (int64, error) {
	store.stripeWebhookCalls++
	return 0, store.stripeWebhookErr
}

func (store *retentionStoreStub) PurgeMessagingDataBatch(
	context.Context,
	messagingdomain.RetentionCutoffs,
	int,
) (messagingdomain.RetentionPurgeResult, error) {
	store.messagingDataCalls++
	return messagingdomain.RetentionPurgeResult{}, store.messagingDataErr
}

func (store *retentionStoreStub) PurgeExpiredContributorArtifacts(
	context.Context,
	feedback.CoreContributorArtifactCutoffs,
) (feedback.CoreContributorArtifactPurgeResult, error) {
	store.feedbackArtifactCalls++
	return feedback.CoreContributorArtifactPurgeResult{}, nil
}

func (store *retentionStoreStub) PurgeDeletedFeedback(
	context.Context,
	time.Time,
) (feedback.CoreDeletedFeedbackPurgeResult, error) {
	store.deletedFeedbackCalls++
	return feedback.CoreDeletedFeedbackPurgeResult{}, nil
}
