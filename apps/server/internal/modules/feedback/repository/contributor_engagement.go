package feedbackrepository

import (
	"context"
	"errors"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) CreateContributorItemAndFollow(ctx context.Context, input feedback.CoreContributorItemInput) (feedback.CoreItem, error) {
	if input.Participant.ID == uuid.Nil || input.Participant.PortalID != input.Item.PortalID {
		return feedback.CoreItem{}, feedback.ErrAuthenticationRequired
	}
	var item feedback.CoreItem
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		itemInput := input.Item
		itemInput.ContributorID = input.Participant.ID
		var err error
		item, err = createItemWithQueries(ctx, q, itemInput, false)
		if err != nil {
			return err
		}
		count, err := q.FollowFeedbackItem(ctx, feedbacksql.FollowFeedbackItemParams{ContributorID: input.Participant.ID, ItemID: item.ID})
		if err != nil {
			return err
		}
		if err = requireRowsAffected(count); err != nil {
			return err
		}
		if err = q.EnsureFeedbackContributorPreferences(ctx, feedbacksql.EnsureFeedbackContributorPreferencesParams{
			PortalID: input.Participant.PortalID, ContributorID: input.Participant.ID,
		}); err != nil {
			return err
		}
		item.Following = true
		return nil
	})
	return item, err
}

func (r *Repo) CreateContributorComment(ctx context.Context, input feedback.CoreContributorCommentInput) (feedback.CoreComment, error) {
	row, err := r.queries.CreateContributorFeedbackComment(ctx, feedbacksql.CreateContributorFeedbackCommentParams{
		ParentID: input.ParentID, Body: input.Body, ContributorID: input.Participant.ID,
		WorkspaceID: input.WorkspaceID, PortalID: input.PortalID, ItemID: input.ItemID,
	})
	if err != nil {
		return feedback.CoreComment{}, normalizeError(err)
	}
	return commentProjection{row.ID, row.WorkspaceID, row.ItemID, row.ContributorID, row.AuthorID, row.ParentID,
		row.AuthorName, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked,
		row.Body, row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) ToggleContributorVote(ctx context.Context, input feedback.CoreContributorVoteInput) (feedback.CoreVoteResult, error) {
	direction, err := safecast.Int16(input.Vote)
	if err != nil {
		return feedback.CoreVoteResult{}, err
	}
	var result feedback.CoreVoteResult
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		current, err := q.GetFeedbackVote(ctx, feedbacksql.GetFeedbackVoteParams{
			WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, ContributorID: input.Participant.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			current, err = 0, nil
		}
		if err != nil {
			return err
		}
		if int(current) == input.Vote {
			count, deleteErr := q.DeleteFeedbackVote(ctx, feedbacksql.DeleteFeedbackVoteParams{
				WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, ContributorID: input.Participant.ID,
			})
			if deleteErr != nil {
				return deleteErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = 0
		} else {
			count, upsertErr := q.UpsertContributorFeedbackVote(ctx, feedbacksql.UpsertContributorFeedbackVoteParams{
				Direction: direction, ContributorID: input.Participant.ID, WorkspaceID: input.WorkspaceID, ItemID: input.ItemID,
			})
			if upsertErr != nil {
				return upsertErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = input.Vote
		}
		count, countErr := q.GetFeedbackVoteCount(ctx, feedbacksql.GetFeedbackVoteCountParams{ItemID: input.ItemID})
		result.VoteCount = int(count)
		return countErr
	})
	return result, err
}

func (r *Repo) GetItemFollow(ctx context.Context, itemID, contributorID uuid.UUID) (feedback.CoreFollowState, error) {
	row, err := r.queries.GetFeedbackItemFollow(ctx, feedbacksql.GetFeedbackItemFollowParams{ContributorID: contributorID, ItemID: itemID})
	if err != nil {
		return feedback.CoreFollowState{}, normalizeError(err)
	}
	return feedback.CoreFollowState{ItemID: row.ItemID, ItemSlug: row.ItemSlug, Title: row.Title,
		ContributorID: row.ContributorID, Following: row.Following, CreatedAt: row.CreatedAt}, nil
}

func (r *Repo) SetItemFollow(ctx context.Context, itemID, contributorID uuid.UUID, following bool) (feedback.CoreFollowState, error) {
	if following {
		count, err := r.queries.FollowFeedbackItem(ctx, feedbacksql.FollowFeedbackItemParams{ContributorID: contributorID, ItemID: itemID})
		if err != nil {
			return feedback.CoreFollowState{}, err
		}
		if err = requireRowsAffected(count); err != nil {
			return feedback.CoreFollowState{}, err
		}
	} else {
		if _, err := r.GetItemFollow(ctx, itemID, contributorID); err != nil {
			return feedback.CoreFollowState{}, err
		}
		if _, err := r.queries.UnfollowFeedbackItem(ctx, feedbacksql.UnfollowFeedbackItemParams{ItemID: itemID, ContributorID: contributorID}); err != nil {
			return feedback.CoreFollowState{}, err
		}
	}
	return r.GetItemFollow(ctx, itemID, contributorID)
}

func (r *Repo) GetContributorPreferences(ctx context.Context, portalID, contributorID uuid.UUID) (feedback.CoreContributorPreferences, error) {
	preference, err := r.queries.GetFeedbackContributorPreference(ctx, feedbacksql.GetFeedbackContributorPreferenceParams{PortalID: portalID, ContributorID: contributorID})
	if err != nil {
		return feedback.CoreContributorPreferences{}, normalizeError(err)
	}
	rows, err := r.queries.ListFeedbackItemFollows(ctx, feedbacksql.ListFeedbackItemFollowsParams{PortalID: portalID, ContributorID: contributorID})
	if err != nil {
		return feedback.CoreContributorPreferences{}, err
	}
	follows := make([]feedback.CoreFollowState, 0, len(rows))
	for _, row := range rows {
		createdAt := row.CreatedAt
		follows = append(follows, feedback.CoreFollowState{ItemID: row.ItemID, ItemSlug: row.ItemSlug,
			Title: row.Title, ContributorID: contributorID, Following: row.Following, CreatedAt: &createdAt})
	}
	return feedback.CoreContributorPreferences{PortalID: portalID, ContributorID: contributorID,
		PortalEmailsEnabled: preference.PortalEmailsEnabled, ItemFollows: follows, UpdatedAt: preference.UpdatedAt}, nil
}

func (r *Repo) SetPortalEmailPreference(ctx context.Context, portalID, contributorID uuid.UUID, enabled bool) (feedback.CoreContributorPreferences, error) {
	var unsubscribedAt *time.Time
	if !enabled {
		now := time.Now().UTC()
		unsubscribedAt = &now
	}
	count, err := r.queries.SetFeedbackPortalEmailPreference(ctx, feedbacksql.SetFeedbackPortalEmailPreferenceParams{
		PortalID: portalID, ContributorID: contributorID, EmailUnsubscribedAt: unsubscribedAt,
	})
	if err != nil {
		return feedback.CoreContributorPreferences{}, err
	}
	if count == 0 {
		return feedback.CoreContributorPreferences{}, feedback.ErrNotFound
	}
	return r.GetContributorPreferences(ctx, portalID, contributorID)
}

func (r *Repo) GetUnreadUpdateCount(ctx context.Context, portalID, contributorID uuid.UUID) (int, error) {
	count, err := r.queries.GetUnreadFeedbackUpdateCount(ctx, feedbacksql.GetUnreadFeedbackUpdateCountParams{ContributorID: contributorID, PortalID: portalID})
	return int(count), err
}

func (r *Repo) MarkUpdatesSeen(ctx context.Context, portalID, contributorID uuid.UUID) (time.Time, error) {
	seenAt, err := r.queries.MarkFeedbackUpdatesSeen(ctx, feedbacksql.MarkFeedbackUpdatesSeenParams{PortalID: portalID, ContributorID: contributorID})
	if err != nil {
		return time.Time{}, normalizeError(err)
	}
	if seenAt == nil {
		return time.Time{}, feedback.ErrNotFound
	}
	return *seenAt, nil
}

func (r *Repo) ListDeliveryRecipients(ctx context.Context, portalID uuid.UUID, itemID, updateID *uuid.UUID, actorContributorID uuid.UUID) ([]feedback.CoreDeliveryRecipient, error) {
	rows, err := r.queries.ListFeedbackDeliveryRecipients(ctx, feedbacksql.ListFeedbackDeliveryRecipientsParams{
		PortalID: portalID, ActorContributorID: actorContributorID, ItemID: itemID, UpdateID: updateID,
	})
	if err != nil {
		return nil, err
	}
	recipients := make([]feedback.CoreDeliveryRecipient, 0, len(rows))
	for _, row := range rows {
		if row.Email == nil {
			continue
		}
		recipients = append(recipients, feedback.CoreDeliveryRecipient{ContributorID: row.ContributorID,
			Email: *row.Email, DisplayName: row.DisplayName, Kind: row.Kind})
	}
	return recipients, nil
}

func (r *Repo) ListAccountUpdateRecipients(ctx context.Context, portalID, updateID uuid.UUID) ([]feedback.CoreAccountUpdateRecipient, error) {
	rows, err := r.queries.ListAccountFeedbackUpdateRecipients(ctx, feedbacksql.ListAccountFeedbackUpdateRecipientsParams{UpdateID: updateID, PortalID: portalID})
	if err != nil {
		return nil, err
	}
	recipients := make([]feedback.CoreAccountUpdateRecipient, 0, len(rows))
	for _, row := range rows {
		if row.UserID != nil {
			recipients = append(recipients, feedback.CoreAccountUpdateRecipient{UserID: *row.UserID, ItemID: row.ItemID})
		}
	}
	return recipients, nil
}

func (r *Repo) ListAccountItemFollowers(ctx context.Context, portalID, itemID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries.ListAccountFeedbackItemFollowers(ctx, feedbacksql.ListAccountFeedbackItemFollowersParams{PortalID: portalID, ItemID: itemID})
	if err != nil {
		return nil, err
	}
	userIDs := make([]uuid.UUID, 0, len(rows))
	for _, userID := range rows {
		if userID != nil {
			userIDs = append(userIDs, *userID)
		}
	}
	return userIDs, nil
}
