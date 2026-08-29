package feedbackrepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	feedbackDeliveryStaleAfter       = 15 * time.Minute
	feedbackUnsubscribeTokenLifetime = 30 * 24 * time.Hour
	feedbackDeliveryRetryingStatus   = "retrying"
	feedbackDeliveryFailedStatus     = "failed"
)

type contributorDeliveryEventPayload struct {
	ItemID    *uuid.UUID `json:"itemId,omitempty"`
	UpdateID  *uuid.UUID `json:"updateId,omitempty"`
	EventType string     `json:"eventType"`
}

func (r *Repo) CreateContributorDelivery(ctx context.Context, input feedback.CoreCreateDeliveryInput) (feedback.CoreDelivery, bool, error) {
	payload, err := json.Marshal(contributorDeliveryEventPayload{ItemID: input.ItemID, UpdateID: input.UpdateID, EventType: input.EventType})
	if err != nil {
		return feedback.CoreDelivery{}, false, err
	}
	var delivery feedback.CoreDelivery
	var created bool
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		row, queryErr := q.CreateFeedbackContributorDelivery(ctx, feedbacksql.CreateFeedbackContributorDeliveryParams{
			DeliveryID: input.DeliveryID, ItemID: input.ItemID, UpdateID: input.UpdateID, EventType: input.EventType,
			DedupeKey: input.DedupeKey, Subject: input.Subject, Message: input.Message,
			DestinationURL: input.DestinationURL, EventPayload: payload, PortalID: input.PortalID,
			ContributorID: input.ContributorID,
		})
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return nil
		}
		if queryErr != nil {
			return queryErr
		}
		delivery = feedback.CoreDelivery{ID: row.ID, PortalID: row.PortalID, ContributorID: row.ContributorID,
			Email: row.RecipientEmail, DisplayName: row.DisplayName, PortalName: row.PortalName, PortalSlug: row.PortalSlug,
			ItemID: row.ItemID, UpdateID: row.UpdateID, EventType: row.EventType, DedupeKey: row.DedupeKey,
			Subject: row.Subject, Message: row.Message, DestinationURL: row.DestinationURL, Status: row.Status,
			AttemptCount: int(row.AttemptCount), FinalFailureReason: valueOrZero(row.FinalFailureReason), CreatedAt: row.CreatedAt}
		created = row.WasCreated
		if !created {
			return nil
		}
		return q.CreateFeedbackUnsubscribeToken(ctx, feedbacksql.CreateFeedbackUnsubscribeTokenParams{
			TokenHash: input.TokenHash, ExpiresAt: time.Now().UTC().Add(feedbackUnsubscribeTokenLifetime),
			DeliveryID: row.ID, PortalID: input.PortalID, ContributorID: input.ContributorID,
		})
	})
	return delivery, created, err
}

func (r *Repo) ClaimContributorDelivery(ctx context.Context, deliveryID uuid.UUID) (feedback.CoreClaimedContributorDelivery, bool, error) {
	staleBefore := time.Now().UTC().Add(-feedbackDeliveryStaleAfter)
	row, err := r.queries.ClaimFeedbackContributorDelivery(ctx, feedbacksql.ClaimFeedbackContributorDeliveryParams{
		DeliveryID: deliveryID, StaleBefore: &staleBefore,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return feedback.CoreClaimedContributorDelivery{}, false, nil
	}
	if err != nil {
		return feedback.CoreClaimedContributorDelivery{}, false, err
	}
	return feedback.CoreClaimedContributorDelivery{ID: row.ID, RecipientEmail: row.RecipientEmail,
		DisplayName: row.DisplayName, PortalName: row.PortalName, PortalSlug: row.PortalSlug,
		Subject: row.Subject, Message: row.Message, DestinationURL: row.DestinationURL,
		TokenHash: append([]byte(nil), row.TokenHash...)}, true, nil
}

func (r *Repo) MarkContributorDeliverySent(ctx context.Context, deliveryID uuid.UUID) error {
	count, err := r.queries.MarkFeedbackContributorDeliverySent(ctx, feedbacksql.MarkFeedbackContributorDeliverySentParams{DeliveryID: deliveryID})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) MarkContributorDeliveryFailed(ctx context.Context, failure feedback.CoreContributorDeliveryFailure) error {
	status := feedbackDeliveryRetryingStatus
	if failure.Terminal {
		status = feedbackDeliveryFailedStatus
	}
	count, err := r.queries.MarkFeedbackContributorDeliveryFailed(ctx, feedbacksql.MarkFeedbackContributorDeliveryFailedParams{
		Status: status, Reason: failure.Reason, DeliveryID: failure.DeliveryID,
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) ListRecoverableContributorDeliveries(ctx context.Context, limit int) ([]feedback.CoreRecoverableContributorDelivery, error) {
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	staleBefore := time.Now().UTC().Add(-feedbackDeliveryStaleAfter)
	rows, err := r.queries.ListRecoverableFeedbackContributorDeliveries(ctx, feedbacksql.ListRecoverableFeedbackContributorDeliveriesParams{
		StaleBefore: &staleBefore, RowLimit: rowLimit,
	})
	if err != nil {
		return nil, err
	}
	deliveries := make([]feedback.CoreRecoverableContributorDelivery, 0, len(rows))
	for _, row := range rows {
		deliveries = append(deliveries, feedback.CoreRecoverableContributorDelivery{DeliveryID: row.DeliveryID,
			TokenHash: append([]byte(nil), row.TokenHash...)})
	}
	return deliveries, nil
}
