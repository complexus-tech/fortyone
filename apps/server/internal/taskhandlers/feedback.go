package taskhandlers

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

const feedbackDeliveryRecoveryBatchSize = 500

func (h *handlers) HandleFeedbackOutboxDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.feedbackOutbox == nil {
		return errors.New("feedback outbox processor is unavailable")
	}
	if err := h.feedbackOutbox.DispatchReadyOutboxEvents(ctx); err != nil {
		return fmt.Errorf("dispatch feedback outbox events: %w", err)
	}
	return nil
}

type feedbackContributorDeliveryData struct {
	ID             uuid.UUID `db:"id"`
	RecipientEmail string    `db:"recipient_email"`
	DisplayName    string    `db:"display_name"`
	PortalName     string    `db:"portal_name"`
	PortalSlug     string    `db:"portal_slug"`
	Subject        string    `db:"subject"`
	Message        string    `db:"message"`
	DestinationURL string    `db:"destination_url"`
	Status         string    `db:"status"`
}

type feedbackContributorDeliveryStore interface {
	ClaimContributorDelivery(context.Context, uuid.UUID) (feedbackContributorDeliveryData, bool, error)
}

type databaseFeedbackContributorDeliveryStore struct {
	db *sqlx.DB
}

const feedbackContributorDeliveryClaimQuery = `
	WITH candidate AS (
		SELECT delivery.id,
			(
				contributor.kind IN ('verified_guest', 'external')
				AND contributor.email IS NOT NULL
				AND contributor.blocked_at IS NULL
				AND preference.email_unsubscribed_at IS NULL
			) AS eligible
		FROM feedback_contributor_deliveries delivery
		INNER JOIN feedback_contributors contributor ON contributor.id = delivery.contributor_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = delivery.portal_id AND preference.contributor_id = delivery.contributor_id
		WHERE delivery.id = $1
			AND (
				delivery.status IN ('queued', 'retrying')
				OR (delivery.status = 'processing' AND delivery.last_attempt_at <= NOW() - INTERVAL '15 minutes')
			)
		FOR UPDATE OF delivery
	), claimed AS (
		UPDATE feedback_contributor_deliveries delivery
		SET status = CASE WHEN candidate.eligible THEN 'processing' ELSE 'suppressed' END,
			next_attempt_at = NULL,
			last_attempt_at = CASE WHEN candidate.eligible THEN NOW() ELSE delivery.last_attempt_at END,
			final_failure_reason = CASE
				WHEN candidate.eligible THEN NULL
				ELSE 'recipient blocked or unsubscribed before delivery'
			END,
			updated_at = NOW()
		FROM candidate
		WHERE delivery.id = candidate.id
		RETURNING delivery.*, candidate.eligible
	)
	SELECT claimed.id,
		claimed.recipient_email,
		COALESCE(NULLIF(trim(contributor.display_name), ''), 'there') AS display_name,
		workspace.name AS portal_name,
		workspace.slug AS portal_slug,
		claimed.subject,
		claimed.message,
		claimed.destination_url,
		claimed.status
	FROM claimed
	INNER JOIN feedback_contributors contributor ON contributor.id = claimed.contributor_id
	INNER JOIN feedback_portals portal ON portal.id = claimed.portal_id
	INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id
	WHERE claimed.eligible
`

func (s *databaseFeedbackContributorDeliveryStore) ClaimContributorDelivery(ctx context.Context, deliveryID uuid.UUID) (feedbackContributorDeliveryData, bool, error) {
	if s == nil || s.db == nil {
		return feedbackContributorDeliveryData{}, false, errors.New("feedback delivery store is unavailable")
	}
	var delivery feedbackContributorDeliveryData
	if err := s.db.GetContext(ctx, &delivery, feedbackContributorDeliveryClaimQuery, deliveryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return feedbackContributorDeliveryData{}, false, nil
		}
		return feedbackContributorDeliveryData{}, false, err
	}
	return delivery, true, nil
}

func (h *handlers) HandleFeedbackContributorDelivery(ctx context.Context, task *asynq.Task) error {
	var payload tasks.FeedbackContributorDeliveryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil || payload.DeliveryID == uuid.Nil || strings.TrimSpace(payload.UnsubscribeToken) == "" {
		return fmt.Errorf("decode feedback contributor delivery: %w", asynq.SkipRetry)
	}
	deliveryStore := h.feedbackDeliveries
	if deliveryStore == nil {
		deliveryStore = &databaseFeedbackContributorDeliveryStore{db: h.db}
	}
	delivery, deliverable, err := deliveryStore.ClaimContributorDelivery(ctx, payload.DeliveryID)
	if err != nil {
		return fmt.Errorf("claim feedback contributor delivery: %w", err)
	}
	if !deliverable {
		return nil
	}
	unsubscribeURL, err := feedbackUnsubscribeURL(delivery.DestinationURL, delivery.PortalSlug, payload.UnsubscribeToken)
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
		status := "retrying"
		if terminal {
			status = "failed"
		}
		_, updateErr := h.db.ExecContext(ctx, `
			UPDATE feedback_contributor_deliveries
			SET status = $2,
				attempt_count = attempt_count + 1,
				last_attempt_at = NOW(),
				next_attempt_at = CASE WHEN $2 = 'retrying' THEN NOW() ELSE NULL END,
				final_failure_reason = CASE WHEN $2 = 'failed' THEN LEFT($3, 1000) ELSE final_failure_reason END,
				updated_at = NOW()
			WHERE id = $1 AND status = 'processing'
		`, delivery.ID, status, err.Error())
		if updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	_, err = h.db.ExecContext(ctx, `
		UPDATE feedback_contributor_deliveries
		SET status = 'sent', attempt_count = attempt_count + 1, last_attempt_at = NOW(), next_attempt_at = NULL, sent_at = NOW(), final_failure_reason = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, delivery.ID)
	if err != nil {
		return fmt.Errorf("mark feedback contributor delivery sent: %w", err)
	}
	return nil
}

func (h *handlers) HandleFeedbackContributorDeliveryRecovery(ctx context.Context, _ *asynq.Task) error {
	if h.db == nil || h.feedbackTasks == nil || strings.TrimSpace(h.feedbackAuthSecret) == "" {
		return errors.New("feedback delivery recovery is not configured")
	}
	var rows []struct {
		DeliveryID uuid.UUID `db:"delivery_id"`
		TokenHash  []byte    `db:"token_hash"`
	}
	if err := h.db.SelectContext(ctx, &rows, `
		SELECT delivery.id AS delivery_id, token.token_hash
		FROM feedback_contributor_deliveries delivery
		INNER JOIN feedback_contributor_unsubscribe_tokens token ON token.delivery_id = delivery.id
		WHERE (
				(
					delivery.status IN ('queued', 'retrying')
					AND (delivery.next_attempt_at IS NULL OR delivery.next_attempt_at <= NOW())
				)
				OR (
					delivery.status = 'processing'
					AND delivery.last_attempt_at <= NOW() - INTERVAL '15 minutes'
				)
			)
			AND token.consumed_at IS NULL AND token.expires_at > NOW()
		ORDER BY delivery.created_at, delivery.id
		LIMIT $1
	`, feedbackDeliveryRecoveryBatchSize); err != nil {
		return fmt.Errorf("load recoverable feedback deliveries: %w", err)
	}
	var recoveryErr error
	for _, row := range rows {
		payload, err := feedbackDeliveryRecoveryPayload(h.feedbackAuthSecret, row.DeliveryID, row.TokenHash)
		if err != nil {
			if _, updateErr := h.db.ExecContext(ctx, `
				UPDATE feedback_contributor_deliveries
				SET status = 'failed', final_failure_reason = $2, next_attempt_at = NULL, updated_at = NOW()
				WHERE id = $1 AND status IN ('queued', 'processing', 'retrying')
			`, row.DeliveryID, "unsubscribe token integrity check failed"); updateErr != nil {
				recoveryErr = errors.Join(recoveryErr, updateErr)
			}
			continue
		}
		if err := h.feedbackTasks.EnqueueFeedbackContributorDelivery(payload); err != nil {
			recoveryErr = errors.Join(recoveryErr, fmt.Errorf("re-enqueue feedback delivery %s: %w", row.DeliveryID, err))
		}
	}
	return recoveryErr
}

func feedbackDeliveryRecoveryPayload(authSecret string, deliveryID uuid.UUID, storedHash []byte) (tasks.FeedbackContributorDeliveryPayload, error) {
	token, derivedHash, err := feedbacksecurity.DeriveUnsubscribeToken(authSecret, deliveryID)
	if err != nil {
		return tasks.FeedbackContributorDeliveryPayload{}, err
	}
	if len(storedHash) != len(derivedHash) || !hmac.Equal(storedHash, derivedHash) {
		return tasks.FeedbackContributorDeliveryPayload{}, errors.New("feedback unsubscribe token hash mismatch")
	}
	return tasks.FeedbackContributorDeliveryPayload{DeliveryID: deliveryID, UnsubscribeToken: token}, nil
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
