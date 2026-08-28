package feedbackrepository

import (
	"context"
	"errors"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) PurgeExpiredContributorArtifacts(ctx context.Context, cutoffs feedback.CoreContributorArtifactCutoffs) (feedback.CoreContributorArtifactPurgeResult, error) {
	var result feedback.CoreContributorArtifactPurgeResult
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		var err error
		result.VerificationsDeleted, err = q.PurgeFeedbackContributorVerifications(ctx, feedbacksql.PurgeFeedbackContributorVerificationsParams{RetainedBefore: cutoffs.RetainedBefore})
		if err != nil {
			return err
		}
		result.SessionsDeleted, err = q.PurgeFeedbackContributorSessions(ctx, feedbacksql.PurgeFeedbackContributorSessionsParams{RetainedBefore: cutoffs.RetainedBefore})
		if err != nil {
			return err
		}
		result.UnsubscribeTokens, err = q.PurgeFeedbackContributorUnsubscribeTokens(ctx, feedbacksql.PurgeFeedbackContributorUnsubscribeTokensParams{RetainedBefore: cutoffs.RetainedBefore})
		if err != nil {
			return err
		}
		result.WidgetNoncesDeleted, err = q.PurgeFeedbackWidgetAssertionNonces(ctx, feedbacksql.PurgeFeedbackWidgetAssertionNoncesParams{ExpiredBefore: cutoffs.ExpiredBefore})
		if err != nil {
			return err
		}
		result.SecretRotationsDeleted, err = q.PurgeFeedbackWidgetSecretRotations(ctx, feedbacksql.PurgeFeedbackWidgetSecretRotationsParams{RetainedBefore: cutoffs.RetainedBefore})
		return err
	})
	return result, err
}

func (r *Repo) PurgeDeletedFeedback(ctx context.Context, deletedBefore time.Time) (feedback.CoreDeletedFeedbackPurgeResult, error) {
	row, err := r.queries.PurgeDeletedFeedback(ctx, feedbacksql.PurgeDeletedFeedbackParams{DeletedBefore: &deletedBefore})
	if err != nil {
		return feedback.CoreDeletedFeedbackPurgeResult{}, err
	}
	return feedback.CoreDeletedFeedbackPurgeResult{ItemsDeleted: row.ItemsDeleted, ContributorsDeleted: row.ContributorsDeleted}, nil
}

func (r *Repo) ListDigestRecipients(ctx context.Context, cursor feedback.CoreDigestRecipientCursor) ([]feedback.CoreDigestRecipient, error) {
	rows, err := r.queries.ListFeedbackDigestRecipients(ctx, feedbacksql.ListFeedbackDigestRecipientsParams{
		HasCursor: cursor.HasCursor, AfterWorkspaceID: cursor.AfterWorkspaceID, AfterUserID: cursor.AfterUserID, RowLimit: cursor.Limit,
	})
	if err != nil {
		return nil, err
	}
	recipients := make([]feedback.CoreDigestRecipient, 0, len(rows))
	for _, row := range rows {
		recipients = append(recipients, feedback.CoreDigestRecipient{UserID: row.UserID, UserEmail: row.UserEmail,
			UserName: row.UserName, Timezone: row.Timezone, WorkspaceID: row.WorkspaceID,
			WorkspaceName: row.WorkspaceName, WorkspaceSlug: row.WorkspaceSlug})
	}
	return recipients, nil
}

func (r *Repo) ListDigestSubscriptions(ctx context.Context, recipientID, workspaceID uuid.UUID) ([]feedback.CoreDigestSubscription, error) {
	rows, err := r.queries.ListFeedbackDigestSubscriptions(ctx, feedbacksql.ListFeedbackDigestSubscriptionsParams{RecipientID: recipientID, WorkspaceID: workspaceID})
	if err != nil {
		return nil, err
	}
	subscriptions := make([]feedback.CoreDigestSubscription, 0, len(rows))
	for _, row := range rows {
		subscriptions = append(subscriptions, feedback.CoreDigestSubscription{BoardID: row.BoardID, TeamID: row.TeamID,
			EmailFrequency: row.EmailFrequency, CreatedAt: row.CreatedAt, LastDigestSentAt: row.LastDigestSentAt,
			LastDigestCursorAt: row.LastDigestCursorAt})
	}
	return subscriptions, nil
}

func (r *Repo) ClaimDigestDelivery(ctx context.Context, claim feedback.CoreDigestDeliveryClaim) (uuid.UUID, bool, error) {
	deliveryID, err := r.queries.ClaimFeedbackDigestDelivery(ctx, feedbacksql.ClaimFeedbackDigestDeliveryParams{
		LocalDate: claim.LocalDate, WindowStart: claim.WindowStart, WindowEnd: claim.WindowEnd,
		RecipientID: claim.RecipientID, WorkspaceID: claim.WorkspaceID, StaleBefore: claim.StaleBefore,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	return deliveryID, err == nil, err
}

func (r *Repo) ListDigestItems(ctx context.Context, query feedback.CoreDigestItemsQuery) ([]feedback.CoreDigestItem, error) {
	rows, err := r.queries.ListFeedbackDigestItems(ctx, feedbacksql.ListFeedbackDigestItemsParams{
		RowLimit: query.Limit, BoardIds: query.BoardIDs, WindowStarts: query.WindowStarts,
		RecipientID: query.RecipientID, WorkspaceID: query.WorkspaceID, WindowEnd: query.WindowEnd,
	})
	if err != nil {
		return nil, err
	}
	items := make([]feedback.CoreDigestItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedback.CoreDigestItem{ID: row.ID, TeamID: row.TeamID, Title: row.Title,
			Description: row.Description, AuthorName: row.AuthorName, TeamName: row.TeamName, Status: row.Status,
			CreatedAt: row.CreatedAt, TotalCount: row.TotalCount, PendingReviewCount: row.PendingReviewCount})
	}
	return items, nil
}

func (r *Repo) CompleteDigestDelivery(ctx context.Context, completion feedback.CoreDigestDeliveryCompletion) error {
	return r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		advanced, err := q.AdvanceFeedbackDigestSubscriptionCursors(ctx, feedbacksql.AdvanceFeedbackDigestSubscriptionCursorsParams{
			DeliveryAt: &completion.DeliveredAt, WindowEnd: &completion.WindowEnd, RecipientID: completion.RecipientID,
			WorkspaceID: completion.WorkspaceID, BoardIds: completion.BoardIDs,
		})
		if err != nil {
			return err
		}
		if len(completion.BoardIDs) > 0 && advanced == 0 {
			return feedback.ErrNotFound
		}
		completed, err := q.CompleteFeedbackDigestDelivery(ctx, feedbacksql.CompleteFeedbackDigestDeliveryParams{
			Status: string(completion.Status), ItemCount: completion.ItemCount, SentAt: &completion.DeliveredAt,
			DeliveryID: completion.DeliveryID, WorkspaceID: completion.WorkspaceID, RecipientID: completion.RecipientID,
		})
		if err != nil {
			return err
		}
		return requireRowsAffected(completed)
	})
}

func (r *Repo) FailDigestDelivery(ctx context.Context, deliveryID uuid.UUID, failure string) error {
	count, err := r.queries.FailFeedbackDigestDelivery(ctx, feedbacksql.FailFeedbackDigestDeliveryParams{Failure: failure, DeliveryID: deliveryID})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}
