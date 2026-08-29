package taskhandlers

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const feedbackDeliveryRecoveryBatchSize = 500

const feedbackDeliveryFailureReason = "feedback contributor email delivery failed"

func (h *handlers) HandleFeedbackOutboxDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.feedbackOutbox == nil {
		return errors.New("feedback outbox processor is unavailable")
	}
	if err := h.feedbackOutbox.DispatchReadyOutboxEvents(ctx); err != nil {
		return fmt.Errorf("dispatch feedback outbox events: %w", err)
	}
	return nil
}

type feedbackContributorDeliveryStore = feedback.ContributorDeliveryStore

func (h *handlers) HandleFeedbackContributorDelivery(ctx context.Context, task *asynq.Task) error {
	var payload tasks.FeedbackContributorDeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil || payload.DeliveryID == uuid.Nil {
		return fmt.Errorf("decode feedback contributor delivery: %w", asynq.SkipRetry)
	}
	if h.feedbackDeliveries == nil {
		return errors.New("feedback contributor delivery store is unavailable")
	}
	delivery, deliverable, err := h.feedbackDeliveries.ClaimContributorDelivery(ctx, payload.DeliveryID)
	if err != nil {
		return fmt.Errorf("claim feedback contributor delivery: %w", err)
	}
	if !deliverable {
		return nil
	}
	unsubscribeToken, err := feedbackUnsubscribeToken(h.feedbackSecurityKey, delivery.ID, delivery.TokenHash)
	if err != nil {
		return fmt.Errorf("verify feedback unsubscribe token: %w: %w", err, asynq.SkipRetry)
	}
	unsubscribeURL, err := feedbackUnsubscribeURL(delivery.DestinationURL, delivery.PortalSlug, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("build feedback unsubscribe URL: %w: %w", err, asynq.SkipRetry)
	}
	body := fmt.Sprintf(`<!doctype html><html><body><p>Hi %s,</p><p>%s</p><p><a href="%s">View feedback</a></p><p style="font-size:12px;color:#666"><a href="%s">Manage feedback emails</a></p></body></html>`,
		html.EscapeString(delivery.DisplayName),
		html.EscapeString(delivery.Message),
		html.EscapeString(delivery.DestinationURL),
		html.EscapeString(unsubscribeURL),
	)
	if err := h.mailerService.Send(ctx, mailer.Email{
		To:            []string{delivery.RecipientEmail},
		Subject:       delivery.Subject,
		Body:          body,
		PlainTextBody: delivery.Message + "\n\n" + delivery.DestinationURL + "\n\nManage feedback emails: " + unsubscribeURL,
		IsHTML:        true,
	}); err != nil {
		retryCount, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)
		terminal := retryCount >= maxRetry
		updateErr := h.feedbackDeliveries.MarkContributorDeliveryFailed(ctx, feedback.CoreContributorDeliveryFailure{
			DeliveryID: delivery.ID,
			Reason:     feedbackDeliveryFailureReason,
			Terminal:   terminal,
		})
		deliveryErr := errors.New(feedbackDeliveryFailureReason)
		if updateErr != nil {
			return errors.Join(deliveryErr, updateErr)
		}
		return deliveryErr
	}
	if err := h.feedbackDeliveries.MarkContributorDeliverySent(ctx, delivery.ID); err != nil {
		return fmt.Errorf("mark feedback contributor delivery sent: %w", err)
	}
	return nil
}

func (h *handlers) HandleFeedbackContributorDeliveryRecovery(ctx context.Context, _ *asynq.Task) error {
	if h.feedbackDeliveries == nil || h.feedbackTasks == nil || strings.TrimSpace(h.feedbackSecurityKey) == "" {
		return errors.New("feedback delivery recovery is not configured")
	}
	rows, err := h.feedbackDeliveries.ListRecoverableContributorDeliveries(ctx, feedbackDeliveryRecoveryBatchSize)
	if err != nil {
		return fmt.Errorf("load recoverable feedback deliveries: %w", err)
	}
	var recoveryErr error
	for _, row := range rows {
		_, err := feedbackUnsubscribeToken(h.feedbackSecurityKey, row.DeliveryID, row.TokenHash)
		if err != nil {
			if updateErr := h.feedbackDeliveries.MarkContributorDeliveryFailed(ctx, feedback.CoreContributorDeliveryFailure{
				DeliveryID: row.DeliveryID,
				Reason:     "unsubscribe token integrity check failed",
				Terminal:   true,
			}); updateErr != nil {
				recoveryErr = errors.Join(recoveryErr, updateErr)
			}
			continue
		}
		if err := h.feedbackTasks.EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload{DeliveryID: row.DeliveryID}); err != nil {
			recoveryErr = errors.Join(recoveryErr, fmt.Errorf("re-enqueue feedback delivery %s: %w", row.DeliveryID, err))
		}
	}
	return recoveryErr
}

func feedbackUnsubscribeToken(authSecret string, deliveryID uuid.UUID, storedHash []byte) (string, error) {
	token, derivedHash, err := feedbacksecurity.DeriveUnsubscribeToken(authSecret, deliveryID)
	if err != nil {
		return "", err
	}
	if len(storedHash) != len(derivedHash) || !hmac.Equal(storedHash, derivedHash) {
		return "", errors.New("feedback unsubscribe token hash mismatch")
	}
	return token, nil
}

func feedbackUnsubscribeURL(destinationURL, portalSlug, token string) (string, error) {
	destination, err := url.Parse(strings.TrimSpace(destinationURL))
	if err != nil || destination.Scheme == "" || destination.Host == "" {
		return "", errors.New("delivery destination must be an absolute URL")
	}
	destination.Path = "/portal/" + strings.TrimSpace(portalSlug) + "/feedback/preferences/exchange"
	destination.RawQuery = url.Values{"token": []string{token}}.Encode()
	destination.Fragment = ""
	return destination.String(), nil
}
