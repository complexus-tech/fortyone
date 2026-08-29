package feedbackrepository

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mergeSnapshot struct {
	SchemaVersion        int         `json:"schemaVersion"`
	MergeEventID         uuid.UUID   `json:"mergeEventId"`
	WorkspaceID          uuid.UUID   `json:"workspaceId"`
	PortalID             uuid.UUID   `json:"portalId"`
	SourceItemID         uuid.UUID   `json:"sourceItemId"`
	TargetItemID         uuid.UUID   `json:"targetItemId"`
	MergedByUserID       uuid.UUID   `json:"mergedByUserId"`
	MergedAt             time.Time   `json:"mergedAt"`
	SourceTitle          string      `json:"sourceTitle"`
	SourceSlug           string      `json:"sourceSlug"`
	TargetTitle          string      `json:"targetTitle"`
	TargetSlug           string      `json:"targetSlug"`
	MovedFollowerCount   int32       `json:"movedFollowerCount"`
	MovedUpdateLinkCount int32       `json:"movedUpdateLinkCount"`
	MovedStoryLinkCount  int32       `json:"movedStoryLinkCount"`
	SourceFollowerIDs    []uuid.UUID `json:"sourceFollowerIds"`
}

func (r *Repo) ListItemCandidates(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, int) (feedback.CoreMergeCandidatesPage, error) {
	return feedback.CoreMergeCandidatesPage{}, feedback.ErrForbidden
}

func (r *Repo) ListItemCandidatesScoped(ctx context.Context, scope feedback.CoreAccessScope, portalID, excludedItemID uuid.UUID, search string, limit int) (feedback.CoreMergeCandidatesPage, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreMergeCandidatesPage{}, feedback.ErrForbidden
	}
	rowLimit, err := safecast.Int32(limit + 1)
	if err != nil {
		return feedback.CoreMergeCandidatesPage{}, err
	}
	rows, err := r.queries.ListFeedbackMergeCandidates(ctx, feedbacksql.ListFeedbackMergeCandidatesParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
		ExcludedItemID: excludedItemID, Search: strings.TrimSpace(search), AllTeams: scope.AllTeams,
		CredentialTeamIds: scope.CredentialTeamIDs, RowLimit: rowLimit,
	})
	if err != nil {
		return feedback.CoreMergeCandidatesPage{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	candidates := make([]feedback.CoreMergeCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, feedback.CoreMergeCandidate{ID: row.ID, Slug: row.Slug,
			Title: row.Title, Status: row.Status, VoteCount: int(row.VoteCount), CommentCount: int(row.CommentCount)})
	}
	return feedback.CoreMergeCandidatesPage{Candidates: candidates, HasMore: hasMore}, nil
}

func (r *Repo) MergeItems(context.Context, feedback.CoreMergeItemInput) (feedback.CoreMergeItemResult, error) {
	return feedback.CoreMergeItemResult{}, feedback.ErrForbidden
}

func (r *Repo) MergeItemsScoped(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreMergeItemInput) (feedback.CoreMergeItemResult, error) {
	if err := scope.Validate(); err != nil || input.Access.WorkspaceID != scope.WorkspaceID ||
		input.Access.ActorID != scope.ActorID || input.Access.AllTeams != scope.AllTeams ||
		!slices.Equal(input.Access.CredentialTeamIDs, scope.CredentialTeamIDs) ||
		input.WorkspaceID != scope.WorkspaceID || input.ActorID != scope.ActorID {
		return feedback.CoreMergeItemResult{}, feedback.ErrForbidden
	}
	var result feedback.CoreMergeItemResult
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		orderedIDs := []uuid.UUID{input.SourceItemID, input.TargetItemID}
		if orderedIDs[1].String() < orderedIDs[0].String() {
			orderedIDs[0], orderedIDs[1] = orderedIDs[1], orderedIDs[0]
		}
		locked, err := q.LockFeedbackMergeItems(ctx, feedbacksql.LockFeedbackMergeItemsParams{
			ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemIds: orderedIDs,
			AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
		})
		if err != nil {
			return err
		}
		if len(locked) != 2 {
			return feedback.ErrNotFound
		}
		items := make(map[uuid.UUID]feedbacksql.LockFeedbackMergeItemsRow, 2)
		for _, item := range locked {
			items[item.ID] = item
		}
		source, sourceFound := items[input.SourceItemID]
		target, targetFound := items[input.TargetItemID]
		if !sourceFound || !targetFound {
			return feedback.ErrNotFound
		}
		if source.MergedIntoItemID != nil && *source.MergedIntoItemID == target.ID {
			stored, err := q.GetCompletedFeedbackMerge(ctx, feedbacksql.GetCompletedFeedbackMergeParams{
				SourceItemID: source.ID, TargetItemID: target.ID,
			})
			if err != nil {
				return normalizeError(err)
			}
			result = feedback.CoreMergeItemResult{SourceItemID: stored.SourceItemID, TargetItemID: stored.TargetItemID,
				PortalID: stored.PortalID, MergedAt: stored.MergedAt, MergedByUserID: stored.MergedByUserID,
				MovedFollowerCount: int(stored.MovedFollowerCount), MovedUpdateLinkCount: int(stored.MovedUpdateLinkCount),
				MovedStoryLinkCount: int(stored.MovedStoryLinkCount)}
			return hydrateMergeTarget(ctx, q, scope.WorkspaceID, target.ID, &result)
		}
		if source.ID == target.ID || source.PortalID != target.PortalID || source.MergedIntoItemID != nil ||
			target.MergedIntoItemID != nil || source.DeletedAt != nil || target.DeletedAt != nil {
			return feedback.ErrMergeConflict
		}
		hasInbound, err := q.FeedbackItemHasInboundMerges(ctx, feedbacksql.FeedbackItemHasInboundMergesParams{
			WorkspaceID: source.WorkspaceID, PortalID: source.PortalID, ItemID: uuidPointer(source.ID),
		})
		if err != nil {
			return err
		}
		if hasInbound {
			return feedback.ErrMergeConflict
		}
		followerIDs, err := q.ListActiveFeedbackFollowerIDs(ctx, feedbacksql.ListActiveFeedbackFollowerIDsParams{ItemID: source.ID})
		if err != nil {
			return err
		}
		followerCount, err := q.CountFeedbackFollowersMovedByMerge(ctx, feedbacksql.CountFeedbackFollowersMovedByMergeParams{
			TargetItemID: target.ID, SourceItemID: source.ID,
		})
		if err != nil {
			return err
		}
		if err := q.CopyFeedbackFollowersForMerge(ctx, feedbacksql.CopyFeedbackFollowersForMergeParams{
			TargetItemID: target.ID, SourceItemID: source.ID,
		}); err != nil {
			return err
		}
		updateLinkCount, err := q.CopyFeedbackUpdateLinksForMerge(ctx, feedbacksql.CopyFeedbackUpdateLinksForMergeParams{
			TargetItemID: target.ID, SourceItemID: source.ID,
		})
		if err != nil {
			return err
		}
		storyLinkCount, err := moveFeedbackStoryLinksForMerge(ctx, q, source.ID, target.ID)
		if err != nil {
			return err
		}
		mergedAt, err := q.MarkFeedbackItemMerged(ctx, feedbacksql.MarkFeedbackItemMergedParams{
			ActorID: uuidPointer(scope.ActorID), SourceItemID: source.ID, TargetItemID: target.ID,
		})
		if err != nil {
			return normalizeError(err)
		}
		if mergedAt == nil {
			return feedback.ErrMergeConflict
		}
		if err := q.UnsubscribeSourceFeedbackFollowersAfterMerge(ctx, feedbacksql.UnsubscribeSourceFeedbackFollowersAfterMergeParams{
			MergedAt: mergedAt, SourceItemID: source.ID,
		}); err != nil {
			return err
		}
		eventID := uuid.New()
		payload, err := json.Marshal(mergeSnapshot{SchemaVersion: 1, MergeEventID: eventID,
			WorkspaceID: source.WorkspaceID, PortalID: source.PortalID, SourceItemID: source.ID,
			TargetItemID: target.ID, MergedByUserID: scope.ActorID, MergedAt: *mergedAt,
			SourceTitle: source.Title, SourceSlug: source.Slug, TargetTitle: target.Title, TargetSlug: target.Slug,
			MovedFollowerCount: followerCount, MovedUpdateLinkCount: updateLinkCount,
			MovedStoryLinkCount: storyLinkCount, SourceFollowerIDs: followerIDs})
		if err != nil {
			return err
		}
		if err := q.InsertFeedbackMergeOutbox(ctx, feedbacksql.InsertFeedbackMergeOutboxParams{
			EventID: eventID, SourceItemID: source.ID, TargetItemID: target.ID, WorkspaceID: source.WorkspaceID,
			PortalID: source.PortalID, ActorID: scope.ActorID, MergedAt: *mergedAt, EventPayload: payload,
		}); err != nil {
			return err
		}
		result = feedback.CoreMergeItemResult{SourceItemID: source.ID, TargetItemID: target.ID,
			PortalID: source.PortalID, MergedAt: *mergedAt, MergedByUserID: scope.ActorID,
			MovedFollowerCount: int(followerCount), MovedUpdateLinkCount: int(updateLinkCount),
			MovedStoryLinkCount: int(storyLinkCount)}
		return hydrateMergeTarget(ctx, q, scope.WorkspaceID, target.ID, &result)
	})
	return result, err
}

func moveFeedbackStoryLinksForMerge(ctx context.Context, q feedbacksql.Querier, sourceID, targetID uuid.UUID) (int32, error) {
	sourcePrimary, sourceErr := q.GetPrimaryFeedbackStoryID(ctx, feedbacksql.GetPrimaryFeedbackStoryIDParams{ItemID: sourceID})
	targetPrimary, targetErr := q.GetPrimaryFeedbackStoryID(ctx, feedbacksql.GetPrimaryFeedbackStoryIDParams{ItemID: targetID})
	if sourceErr != nil && !errors.Is(sourceErr, pgx.ErrNoRows) {
		return 0, sourceErr
	}
	if targetErr != nil && !errors.Is(targetErr, pgx.ErrNoRows) {
		return 0, targetErr
	}
	sourceHasPrimary := sourceErr == nil
	targetHasPrimary := targetErr == nil
	if sourceHasPrimary && targetHasPrimary && sourcePrimary != targetPrimary {
		return 0, feedback.ErrMergeConflict
	}
	var moved int32
	if sourceHasPrimary && !targetHasPrimary {
		alreadyLinked, err := q.FeedbackTargetAlreadyLinksStory(ctx, feedbacksql.FeedbackTargetAlreadyLinksStoryParams{
			TargetItemID: targetID, StoryID: sourcePrimary,
		})
		if err != nil {
			return 0, err
		}
		if alreadyLinked {
			return 0, feedback.ErrMergeConflict
		}
		count, err := q.MovePrimaryFeedbackStoryLink(ctx, feedbacksql.MovePrimaryFeedbackStoryLinkParams{
			TargetItemID: targetID, SourceItemID: sourceID,
		})
		if err != nil {
			return 0, err
		}
		converted, err := safecast.Int64ToInt32(count)
		if err != nil {
			return 0, err
		}
		moved += converted
	}
	nonPrimary, err := q.CopyNonPrimaryFeedbackStoryLinksForMerge(ctx, feedbacksql.CopyNonPrimaryFeedbackStoryLinksForMergeParams{
		TargetItemID: targetID, SourceItemID: sourceID,
	})
	return moved + nonPrimary, err
}

func hydrateMergeTarget(ctx context.Context, q feedbacksql.Querier, workspaceID, targetID uuid.UUID, result *feedback.CoreMergeItemResult) error {
	row, err := q.GetWorkspaceFeedbackItem(ctx, feedbacksql.GetWorkspaceFeedbackItemParams{WorkspaceID: workspaceID, ItemID: targetID})
	if err != nil {
		return normalizeError(err)
	}
	result.Target = workspaceItemProjection(row).core()
	return nil
}

func (r *Repo) ClaimMergeOutboxEvents(ctx context.Context, limit int, staleAfter time.Duration) ([]feedback.CoreMergeOutboxEvent, error) {
	if limit <= 0 {
		return []feedback.CoreMergeOutboxEvent{}, nil
	}
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	staleBefore := time.Now().UTC().Add(-staleAfter)
	rows, err := r.queries.ClaimFeedbackMergeOutboxEvents(ctx, feedbacksql.ClaimFeedbackMergeOutboxEventsParams{
		StaleBefore: &staleBefore, RowLimit: rowLimit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreMergeOutboxEvent, 0, len(rows))
	for _, row := range rows {
		if row.ClaimToken == nil {
			return nil, errors.New("claimed feedback merge has no claim token")
		}
		result = append(result, feedback.CoreMergeOutboxEvent{EventID: row.EventID, WorkspaceID: row.WorkspaceID,
			PortalID: row.PortalID, SourceItemID: row.SourceItemID, TargetItemID: row.TargetItemID,
			ActorID: row.ActorID, MergedAt: row.MergedAt, ClaimToken: *row.ClaimToken,
			AttemptCount: int(row.AttemptCount), Payload: append(json.RawMessage(nil), row.EventPayload...)})
	}
	return result, nil
}

func (r *Repo) ListMergeRecipients(ctx context.Context, portalID, targetItemID uuid.UUID, contributorIDs []uuid.UUID) ([]feedback.CoreMergeRecipient, error) {
	if len(contributorIDs) == 0 {
		return []feedback.CoreMergeRecipient{}, nil
	}
	rows, err := r.queries.ListFeedbackMergeRecipients(ctx, feedbacksql.ListFeedbackMergeRecipientsParams{
		TargetItemID: targetItemID, PortalID: portalID, ContributorIds: contributorIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreMergeRecipient, 0, len(rows))
	for _, row := range rows {
		userID := uuid.Nil
		if row.UserID != nil {
			userID = *row.UserID
		}
		result = append(result, feedback.CoreMergeRecipient{ContributorID: row.ContributorID, UserID: userID, Kind: row.Kind})
	}
	return result, nil
}

func (r *Repo) CompleteMergeOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	count, err := r.queries.CompleteFeedbackMergeOutboxEvent(ctx, feedbacksql.CompleteFeedbackMergeOutboxEventParams{
		EventID: eventID, ClaimToken: uuidPointer(claimToken),
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) RetryMergeOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID, failure string, retryAt time.Time, terminal bool) error {
	status := "retrying"
	var nextAttempt *time.Time
	if terminal {
		status = "failed"
	} else {
		nextAttempt = &retryAt
	}
	count, err := r.queries.RetryFeedbackMergeOutboxEvent(ctx, feedbacksql.RetryFeedbackMergeOutboxEventParams{
		Status: status, NextAttemptAt: nextAttempt, Failure: strings.TrimSpace(failure),
		EventID: eventID, ClaimToken: uuidPointer(claimToken),
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}
