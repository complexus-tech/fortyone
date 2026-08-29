package taskhandlers

import (
	"context"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

// RetentionHandlerDependencies contains only the persistence capabilities used
// by scheduled data-retention work. Keeping these ports explicit prevents the
// worker from bypassing module repositories with ad hoc SQL.
type RetentionHandlerDependencies struct {
	Log                 *logger.Logger
	ChatSessions        jobs.ChatSessionPurger
	VerificationTokens  jobs.VerificationTokenPurger
	StripeWebhookEvents jobs.StripeWebhookEventPurger
	MessagingData       jobs.MessagingDataPurger
	Feedback            feedback.MaintenanceStore
}

// RetentionHandlers owns scheduled retention and inactivity-policy operations.
type RetentionHandlers struct {
	log                 *logger.Logger
	chatSessions        jobs.ChatSessionPurger
	verificationTokens  jobs.VerificationTokenPurger
	stripeWebhookEvents jobs.StripeWebhookEventPurger
	messagingData       jobs.MessagingDataPurger
	feedback            feedback.MaintenanceStore
}

func NewRetentionHandlers(dependencies RetentionHandlerDependencies) *RetentionHandlers {
	return &RetentionHandlers{
		log:                 dependencies.Log,
		chatSessions:        dependencies.ChatSessions,
		verificationTokens:  dependencies.VerificationTokens,
		stripeWebhookEvents: dependencies.StripeWebhookEvents,
		messagingData:       dependencies.MessagingData,
		feedback:            dependencies.Feedback,
	}
}

// HandleTokenCleanup retires expired verification and feedback security
// artifacts according to their module-owned retention policies.
func (h *RetentionHandlers) HandleTokenCleanup(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "TokenCleanup", "token cleanup", func(ctx context.Context) error {
		return jobs.PurgeExpiredTokens(ctx, h.verificationTokens, h.feedback, h.log)
	})
}

// HandleChatSessionsCleanup retires expired soft-deleted chat sessions.
func (h *RetentionHandlers) HandleChatSessionsCleanup(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "ChatSessionsCleanup", "chat sessions cleanup", func(ctx context.Context) error {
		return jobs.PurgeDeletedChatSessions(ctx, h.chatSessions, h.log)
	})
}

// HandleWebhookCleanup retires terminal Stripe webhook receipts.
func (h *RetentionHandlers) HandleWebhookCleanup(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "WebhookCleanup", "webhook cleanup", func(ctx context.Context) error {
		return jobs.PurgeOldStripeWebhookEvents(ctx, h.stripeWebhookEvents, h.log)
	})
}

// HandleMessagingCleanup retires expired integration nonces and provider
// messaging data through one bounded, transaction-backed repository port.
func (h *RetentionHandlers) HandleMessagingCleanup(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "MessagingCleanup", "messaging cleanup", func(ctx context.Context) error {
		return jobs.PurgeMessagingData(ctx, h.messagingData, h.log)
	})
}

// HandleDeleteFeedback retires feedback records after their module-owned
// deletion grace period and contributor delivery safeguards have elapsed.
func (h *RetentionHandlers) HandleDeleteFeedback(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "DeleteFeedback", "delete feedback cleanup", func(ctx context.Context) error {
		return jobs.PurgeDeletedFeedback(ctx, h.feedback, h.log)
	})
}

func (h *RetentionHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "retention", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "retention", taskName, operationName, operation)
}
