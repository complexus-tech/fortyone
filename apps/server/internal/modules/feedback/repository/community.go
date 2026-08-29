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

func (r *Repo) ListContributorActivity(ctx context.Context, input feedback.CoreListContributorActivityInput) (feedback.CoreContributorActivityPage, error) {
	offset, limit, err := pageBounds(input.Page, input.PageSize)
	if err != nil {
		return feedback.CoreContributorActivityPage{}, err
	}
	userID := uuidPointer(input.UserID)
	rows, err := r.queries.ListContributorFeedbackActivity(ctx, feedbacksql.ListContributorFeedbackActivityParams{
		ActivityType: input.ActivityType, RowOffset: offset, RowLimit: limit, UserID: userID,
	})
	if err != nil {
		return feedback.CoreContributorActivityPage{}, err
	}
	stats, err := r.queries.GetContributorFeedbackActivityStats(ctx, feedbacksql.GetContributorFeedbackActivityStatsParams{UserID: userID})
	if err != nil {
		return feedback.CoreContributorActivityPage{}, err
	}
	hasMore := len(rows) > input.PageSize
	if hasMore {
		rows = rows[:input.PageSize]
	}
	activities := make([]feedback.CoreContributorActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, feedback.CoreContributorActivity{ID: row.ID, Type: row.ActivityType,
			FeedbackID: row.FeedbackID, FeedbackTitle: row.FeedbackTitle, FeedbackSlug: row.FeedbackSlug,
			Body: row.Body, BoardName: row.BoardName, Status: row.Status, VoteCount: int(row.VoteCount),
			CommentCount: int(row.CommentCount), PortalSlug: row.PortalSlug, WorkspaceName: row.WorkspaceName,
			WorkspaceSlug: row.WorkspaceSlug, CreatedAt: row.CreatedAt})
	}
	return feedback.CoreContributorActivityPage{Activities: activities, Page: input.Page, PageSize: input.PageSize,
		HasMore: hasMore, FeedbackCount: int(stats.FeedbackCount), CommentCount: int(stats.CommentCount),
		VoteScore: int(stats.VoteScore), PortalCount: int(stats.PortalCount)}, nil
}

func (r *Repo) GetContributor(ctx context.Context, portalID, authorID uuid.UUID) (feedback.CoreContributor, error) {
	row, err := r.queries.GetPublicFeedbackContributor(ctx, feedbacksql.GetPublicFeedbackContributorParams{PortalID: portalID, AuthorID: authorID})
	if err != nil {
		return feedback.CoreContributor{}, normalizeError(err)
	}
	return feedback.CoreContributor{ID: row.ID, PortalID: portalID, UserID: row.ID, Kind: feedback.ContributorKindAccount,
		Name: row.Name, AvatarURL: row.AvatarURL, JoinedAt: row.JoinedAt,
		Stats: feedback.CoreContributorStats{FeedbackCount: int(row.FeedbackCount), CommentCount: int(row.CommentCount), VoteScore: int(row.VoteScore)}}, nil
}

func (r *Repo) ContributorExists(ctx context.Context, portalID, authorID uuid.UUID) (bool, error) {
	return r.queries.PublicFeedbackContributorExists(ctx, feedbacksql.PublicFeedbackContributorExistsParams{PortalID: portalID, AuthorID: uuidPointer(authorID)})
}

func (r *Repo) ListContributorComments(ctx context.Context, input feedback.CoreListContributorCommentsInput) (feedback.CoreContributorCommentsPage, error) {
	offset, limit, err := pageBounds(input.Page, input.PageSize)
	if err != nil {
		return feedback.CoreContributorCommentsPage{}, err
	}
	rows, err := r.queries.ListPublicContributorComments(ctx, feedbacksql.ListPublicContributorCommentsParams{
		PortalID: input.PortalID, AuthorID: uuidPointer(input.AuthorID), RowOffset: offset, RowLimit: limit,
	})
	if err != nil {
		return feedback.CoreContributorCommentsPage{}, err
	}
	hasMore := len(rows) > input.PageSize
	if hasMore {
		rows = rows[:input.PageSize]
	}
	comments := make([]feedback.CoreContributorComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, feedback.CoreContributorComment{ID: row.ID, ItemID: row.ItemID,
			FeedbackTitle: row.FeedbackTitle, FeedbackSlug: row.FeedbackSlug, Body: row.Body,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return feedback.CoreContributorCommentsPage{Comments: comments, Page: input.Page, PageSize: input.PageSize, HasMore: hasMore}, nil
}

func (r *Repo) ListComments(ctx context.Context, portalID uuid.UUID, itemIDs []uuid.UUID) ([]feedback.CoreComment, error) {
	if len(itemIDs) == 0 {
		return []feedback.CoreComment{}, nil
	}
	rows, err := r.queries.ListPublicFeedbackComments(ctx, feedbacksql.ListPublicFeedbackCommentsParams{PortalID: portalID, ItemIds: itemIDs})
	if err != nil {
		return nil, err
	}
	comments := make([]feedback.CoreComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, publicCommentProjection(row).core())
	}
	return comments, nil
}

func (r *Repo) ListItemComments(ctx context.Context, workspaceID, itemID uuid.UUID) ([]feedback.CoreComment, error) {
	rows, err := r.queries.ListInternalFeedbackComments(ctx, feedbacksql.ListInternalFeedbackCommentsParams{
		ActorID: workspaceID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
	if err != nil {
		return nil, err
	}
	comments := make([]feedback.CoreComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, internalCommentProjection(row).core())
	}
	return comments, nil
}

func (r *Repo) ListItemCommentsScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) ([]feedback.CoreComment, error) {
	rows, err := r.queries.ListInternalFeedbackComments(ctx, feedbacksql.ListInternalFeedbackCommentsParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return nil, err
	}
	comments := make([]feedback.CoreComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, internalCommentProjection(row).core())
	}
	return comments, nil
}

func (r *Repo) GetComment(ctx context.Context, workspaceID, itemID, commentID uuid.UUID) (feedback.CoreComment, error) {
	row, err := r.queries.GetInternalFeedbackComment(ctx, feedbacksql.GetInternalFeedbackCommentParams{
		ActorID: workspaceID, WorkspaceID: workspaceID, ItemID: itemID, CommentID: commentID, AllTeams: true,
	})
	if err != nil {
		return feedback.CoreComment{}, normalizeError(err)
	}
	return getCommentProjection(row).core(), nil
}

func (r *Repo) GetCommentScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID, commentID uuid.UUID) (feedback.CoreComment, error) {
	row, err := r.queries.GetInternalFeedbackComment(ctx, feedbacksql.GetInternalFeedbackCommentParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID, CommentID: commentID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return feedback.CoreComment{}, normalizeError(err)
	}
	return getCommentProjection(row).core(), nil
}

func (r *Repo) CreateComment(ctx context.Context, input feedback.CoreCommentInput) (feedback.CoreComment, error) {
	row, err := r.queries.CreateAccountFeedbackComment(ctx, feedbacksql.CreateAccountFeedbackCommentParams{
		ActorID: input.AuthorID, WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, AllTeams: true,
		ParentID: input.ParentID, Body: input.Body,
	})
	if err != nil {
		return feedback.CoreComment{}, normalizeError(err)
	}
	return accountCommentProjection(row).core(), nil
}

func (r *Repo) CreateCommentScoped(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreCommentInput) (feedback.CoreComment, error) {
	if input.WorkspaceID != scope.WorkspaceID || input.AuthorID != scope.ActorID {
		return feedback.CoreComment{}, feedback.ErrForbidden
	}
	row, err := r.queries.CreateAccountFeedbackComment(ctx, feedbacksql.CreateAccountFeedbackCommentParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: input.ItemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs, ParentID: input.ParentID, Body: input.Body,
	})
	if err != nil {
		return feedback.CoreComment{}, normalizeError(err)
	}
	return accountCommentProjection(row).core(), nil
}

func (r *Repo) GetPrivateAuthor(ctx context.Context, workspaceID, itemID uuid.UUID) (feedback.CorePrivateAuthor, error) {
	row, err := r.queries.GetFeedbackPrivateAuthor(ctx, feedbacksql.GetFeedbackPrivateAuthorParams{
		ActorID: workspaceID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
	if err != nil {
		return feedback.CorePrivateAuthor{}, normalizeError(err)
	}
	masked := false
	if row.PublicMasked != nil {
		masked = *row.PublicMasked
	}
	return feedback.CorePrivateAuthor{ContributorID: row.ContributorID, UserID: row.UserID, Kind: row.Kind,
		DisplayName: row.DisplayName, Email: nonEmptyStringPointer(row.Email), AvatarURL: nonEmptyStringPointer(row.AvatarURL), PublicMasked: masked}, nil
}

func (r *Repo) GetPrivateAuthorScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) (feedback.CorePrivateAuthor, error) {
	row, err := r.queries.GetFeedbackPrivateAuthor(ctx, feedbacksql.GetFeedbackPrivateAuthorParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return feedback.CorePrivateAuthor{}, normalizeError(err)
	}
	return feedback.CorePrivateAuthor{ContributorID: row.ContributorID, UserID: row.UserID, Kind: row.Kind,
		DisplayName: row.DisplayName, Email: nonEmptyStringPointer(row.Email), AvatarURL: nonEmptyStringPointer(row.AvatarURL),
		PublicMasked: row.PublicMasked != nil && *row.PublicMasked}, nil
}

func (r *Repo) GetItemReadAt(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (*time.Time, error) {
	readAt, err := r.queries.GetFeedbackItemReadAt(ctx, feedbacksql.GetFeedbackItemReadAtParams{
		ActorID: userID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &readAt, nil
}

func (r *Repo) GetItemReadAtScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) (*time.Time, error) {
	readAt, err := r.queries.GetFeedbackItemReadAt(ctx, feedbacksql.GetFeedbackItemReadAtParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &readAt, nil
}

func (r *Repo) ListTeamSummaries(ctx context.Context, workspaceID, userID uuid.UUID) ([]feedback.CoreTeamSummary, error) {
	rows, err := r.queries.ListFeedbackTeamSummaries(ctx, feedbacksql.ListFeedbackTeamSummariesParams{ActorID: userID, WorkspaceID: workspaceID, AllTeams: true})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreTeamSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreTeamSummary{TeamID: row.TeamID, Enabled: row.Enabled,
			TotalCount: int(row.TotalCount), UnreadCount: int(row.UnreadCount)})
	}
	return result, nil
}

func (r *Repo) ListTeamSummariesScoped(ctx context.Context, scope feedback.CoreAccessScope) ([]feedback.CoreTeamSummary, error) {
	rows, err := r.queries.ListFeedbackTeamSummaries(ctx, feedbacksql.ListFeedbackTeamSummariesParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreTeamSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreTeamSummary{TeamID: row.TeamID, Enabled: row.Enabled,
			TotalCount: int(row.TotalCount), UnreadCount: int(row.UnreadCount)})
	}
	return result, nil
}

func (r *Repo) MarkItemRead(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (time.Time, error) {
	return r.queries.MarkFeedbackItemRead(ctx, feedbacksql.MarkFeedbackItemReadParams{
		ActorID: userID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
}

func (r *Repo) MarkItemReadScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) (time.Time, error) {
	return r.queries.MarkFeedbackItemRead(ctx, feedbacksql.MarkFeedbackItemReadParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
}

func (r *Repo) MarkItemUnread(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	_, err := r.queries.MarkFeedbackItemUnread(ctx, feedbacksql.MarkFeedbackItemUnreadParams{
		ActorID: userID, WorkspaceID: workspaceID, ItemID: itemID, AllTeams: true,
	})
	return err
}

func (r *Repo) MarkItemUnreadScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) error {
	count, err := r.queries.MarkFeedbackItemUnread(ctx, feedbacksql.MarkFeedbackItemUnreadParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
		AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (feedback.CoreVoteResult, error) {
	direction, err := safecast.Int16(vote)
	if err != nil {
		return feedback.CoreVoteResult{}, err
	}
	var result feedback.CoreVoteResult
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		participant, err := q.GetOrCreateAccountFeedbackContributor(ctx, feedbacksql.GetOrCreateAccountFeedbackContributorParams{UserID: userID, PortalID: uuid.Nil})
		if err != nil {
			// The vote query resolves the item portal and the matching account
			// contributor itself. A zero portal lookup is therefore optional.
			participant.ID = uuid.Nil
		}
		current := int32(0)
		if participant.ID != uuid.Nil {
			current, err = q.GetFeedbackVote(ctx, feedbacksql.GetFeedbackVoteParams{WorkspaceID: workspaceID, ItemID: itemID, ContributorID: participant.ID})
			if errors.Is(err, pgx.ErrNoRows) {
				err = nil
			}
			if err != nil {
				return err
			}
		}
		if int(current) == vote && participant.ID != uuid.Nil {
			count, deleteErr := q.DeleteFeedbackVote(ctx, feedbacksql.DeleteFeedbackVoteParams{WorkspaceID: workspaceID, ItemID: itemID, ContributorID: participant.ID})
			if deleteErr != nil {
				return deleteErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = 0
		} else {
			count, upsertErr := q.UpsertAccountFeedbackVote(ctx, feedbacksql.UpsertAccountFeedbackVoteParams{
				Direction: direction, ActorID: userID, WorkspaceID: workspaceID, ItemID: itemID,
			})
			if upsertErr != nil {
				return upsertErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = vote
		}
		count, countErr := q.GetFeedbackVoteCount(ctx, feedbacksql.GetFeedbackVoteCountParams{ItemID: itemID})
		result.VoteCount = int(count)
		return countErr
	})
	return result, err
}

func (r *Repo) ToggleVoteScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID, vote int) (feedback.CoreVoteResult, error) {
	direction, err := safecast.Int16(vote)
	if err != nil {
		return feedback.CoreVoteResult{}, err
	}
	var result feedback.CoreVoteResult
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		params := feedbacksql.GetAccountFeedbackVoteParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID,
			ItemID: itemID, AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs}
		current, err := q.GetAccountFeedbackVote(ctx, params)
		if err != nil {
			return normalizeError(err)
		}
		if current == int32(direction) {
			count, deleteErr := q.DeleteAccountFeedbackVote(ctx, feedbacksql.DeleteAccountFeedbackVoteParams{
				ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
				AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
			})
			if deleteErr != nil {
				return deleteErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = 0
		} else {
			count, upsertErr := q.UpsertAccountFeedbackVote(ctx, feedbacksql.UpsertAccountFeedbackVoteParams{
				Direction: direction, ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID,
				AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
			})
			if upsertErr != nil {
				return upsertErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			result.Vote = vote
		}
		count, countErr := q.GetFeedbackVoteCount(ctx, feedbacksql.GetFeedbackVoteCountParams{ItemID: itemID})
		if countErr != nil {
			return countErr
		}
		result.VoteCount, err = safecast.Int64(int64(count))
		return err
	})
	return result, err
}

func publicCommentProjection(row feedbacksql.ListPublicFeedbackCommentsRow) commentProjection {
	return commentProjection{row.ID, row.WorkspaceID, row.ItemID, row.ContributorID, row.AuthorID, row.ParentID,
		row.AuthorName, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked, row.Body, row.CreatedAt, row.UpdatedAt}
}

func internalCommentProjection(row feedbacksql.ListInternalFeedbackCommentsRow) commentProjection {
	return commentProjection{row.ID, row.WorkspaceID, row.ItemID, row.ContributorID, row.AuthorID, row.ParentID,
		row.AuthorName, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked, row.Body, row.CreatedAt, row.UpdatedAt}
}

func getCommentProjection(row feedbacksql.GetInternalFeedbackCommentRow) commentProjection {
	return commentProjection{row.ID, row.WorkspaceID, row.ItemID, row.ContributorID, row.AuthorID, row.ParentID,
		row.AuthorName, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked, row.Body, row.CreatedAt, row.UpdatedAt}
}

func accountCommentProjection(row feedbacksql.CreateAccountFeedbackCommentRow) commentProjection {
	return commentProjection{row.ID, row.WorkspaceID, row.ItemID, row.ContributorID, row.AuthorID, row.ParentID,
		row.AuthorName, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked, row.Body, row.CreatedAt, row.UpdatedAt}
}
