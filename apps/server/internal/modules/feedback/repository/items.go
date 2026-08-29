package feedbackrepository

import (
	"context"
	"errors"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const feedbackTrashRecoveryWindow = 30 * 24 * time.Hour

func (r *Repo) ListItems(ctx context.Context, input feedback.CoreListItemsInput) (feedback.CoreItemsPage, error) {
	return r.listItems(ctx, feedback.CoreAccessScope{WorkspaceID: input.WorkspaceID, ActorID: input.ViewerID, AllTeams: true}, input)
}

func (r *Repo) ListItemsScoped(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreListItemsInput) (feedback.CoreItemsPage, error) {
	if err := scope.Validate(); err != nil || input.WorkspaceID != scope.WorkspaceID {
		return feedback.CoreItemsPage{}, feedback.ErrForbidden
	}
	input.ViewerID = scope.ActorID
	return r.listItems(ctx, scope, input)
}

func (r *Repo) listItems(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreListItemsInput) (feedback.CoreItemsPage, error) {
	offset, limit, err := pageBounds(input.Page, input.PageSize)
	if err != nil {
		return feedback.CoreItemsPage{}, err
	}
	statusMode := int16(0)
	if input.Status == "active" {
		statusMode = 1
	} else if input.Status != "" && input.Status != "all" {
		statusMode = 2
	}
	viewerID := uuidPointer(input.ViewerID)
	teamID := uuid.Nil
	if input.TeamID != nil {
		teamID = *input.TeamID
	}
	boardID := uuid.Nil
	if input.BoardID != nil {
		boardID = *input.BoardID
	}
	var recoveryCutoff *time.Time
	if input.DeletedOnly {
		cutoff := time.Now().Add(-feedbackTrashRecoveryWindow)
		recoveryCutoff = &cutoff
	}
	rows, err := r.queries.ListFeedbackItems(ctx, feedbacksql.ListFeedbackItemsParams{
		ViewerID: viewerID, FilterPortal: input.PortalID != uuid.Nil, PortalID: input.PortalID,
		FilterItem: input.ItemID != uuid.Nil, ItemID: input.ItemID, FilterTeam: input.TeamID != nil,
		WorkspaceID: input.WorkspaceID, TeamID: teamID, DeletedOnly: input.DeletedOnly,
		RecoveryCutoff: recoveryCutoff, StatusMode: statusMode, Status: input.Status,
		FilterBoard: input.BoardID != nil, BoardID: boardID, FilterAuthor: input.AuthorID != uuid.Nil,
		AuthorID: uuidPointer(input.AuthorID), FilterSearch: strings.TrimSpace(input.Search) != "",
		Search: strings.TrimSpace(input.Search), RequireMember: input.WorkspaceID != uuid.Nil,
		ActorID: scope.ActorID, AllTeams: scope.AllTeams, CredentialTeamIds: scope.CredentialTeamIDs,
		SortKey: input.Sort, RowOffset: offset, RowLimit: limit,
	})
	if err != nil {
		return feedback.CoreItemsPage{}, err
	}
	hasMore := len(rows) > input.PageSize
	if hasMore {
		rows = rows[:input.PageSize]
	}
	items := make([]feedback.CoreItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, listItemProjection(row).core())
	}
	return feedback.CoreItemsPage{Items: items, HasMore: hasMore}, nil
}

func (r *Repo) GetItemScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) (feedback.CoreItem, error) {
	page, err := r.ListItemsScoped(ctx, scope, feedback.CoreListItemsInput{
		WorkspaceID: scope.WorkspaceID,
		ItemID:      itemID,
		Status:      "all",
		Sort:        "newest",
		Page:        1,
		PageSize:    1,
	})
	if err != nil {
		return feedback.CoreItem{}, err
	}
	if len(page.Items) != 1 {
		return feedback.CoreItem{}, feedback.ErrNotFound
	}
	return page.Items[0], nil
}

func (r *Repo) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (feedback.CoreItem, error) {
	row, err := r.queries.GetWorkspaceFeedbackItem(ctx, feedbacksql.GetWorkspaceFeedbackItemParams{WorkspaceID: workspaceID, ItemID: itemID})
	if err != nil {
		return feedback.CoreItem{}, normalizeError(err)
	}
	return workspaceItemProjection(row).core(), nil
}

func (r *Repo) GetItemByPortal(ctx context.Context, portalID, itemID uuid.UUID) (feedback.CoreItem, error) {
	row, err := r.queries.GetPublicFeedbackItem(ctx, feedbacksql.GetPublicFeedbackItemParams{PortalID: portalID, ItemID: itemID})
	if err != nil {
		return feedback.CoreItem{}, normalizeError(err)
	}
	return publicItemProjection(row).core(), nil
}

func (r *Repo) ResolveCanonicalItem(ctx context.Context, portalID uuid.UUID, itemReference string) (feedback.CoreCanonicalItem, error) {
	row, err := r.queries.ResolveCanonicalFeedbackItem(ctx, feedbacksql.ResolveCanonicalFeedbackItemParams{
		PortalID: portalID, ItemReference: strings.TrimSpace(itemReference),
	})
	if err != nil {
		return feedback.CoreCanonicalItem{}, normalizeError(err)
	}
	return feedback.CoreCanonicalItem{ItemID: row.ItemID, ItemSlug: row.ItemSlug, Merged: row.Merged}, nil
}

func (r *Repo) ListSimilarItems(ctx context.Context, portalID uuid.UUID, title, description string, limit int) ([]feedback.CoreSimilarItem, error) {
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSimilarFeedbackItems(ctx, feedbacksql.ListSimilarFeedbackItemsParams{
		RowLimit: rowLimit, Title: title, Description: description, PortalID: portalID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]feedback.CoreSimilarItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedback.CoreSimilarItem{ID: row.ID, Slug: row.Slug, Title: row.Title,
			AuthorID: uuidPointer(row.AuthorID), AuthorName: row.AuthorName, AuthorAvatar: nonEmptyStringPointer(row.AuthorAvatar),
			Status: row.Status, VoteCount: int(row.VoteCount), CommentCount: int(row.CommentCount), Confidence: row.Confidence})
	}
	return items, nil
}

func (r *Repo) CreateItem(ctx context.Context, input feedback.CoreItemInput) (feedback.CoreItem, error) {
	return r.createItem(ctx, input, true)
}

func (r *Repo) CreateAnonymousItem(ctx context.Context, input feedback.CoreItemInput) (feedback.CoreItem, error) {
	var item feedback.CoreItem
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		contributorID, err := q.CreateAnonymousFeedbackContributor(ctx, feedbacksql.CreateAnonymousFeedbackContributorParams{PortalID: input.PortalID})
		if err != nil {
			return normalizeError(err)
		}
		input.ContributorID = contributorID
		item, err = createItemWithQueries(ctx, q, input, false)
		return err
	})
	return item, err
}

func (r *Repo) createItem(ctx context.Context, input feedback.CoreItemInput, requireActor bool) (feedback.CoreItem, error) {
	var item feedback.CoreItem
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		var err error
		item, err = createItemWithQueries(ctx, q, input, requireActor)
		return err
	})
	return item, err
}

func createItemWithQueries(ctx context.Context, q feedbacksql.Querier, input feedback.CoreItemInput, requireActor bool) (feedback.CoreItem, error) {
	itemID, err := q.CreateFeedbackItem(ctx, feedbacksql.CreateFeedbackItemParams{
		Title: input.Title, Description: input.Description, Slug: input.Slug, SubmissionSource: input.Source,
		BoardID: input.BoardID, ContributorID: input.ContributorID, WorkspaceID: input.WorkspaceID,
		PortalID: input.PortalID, RequireActor: requireActor, ActorID: uuidPointer(input.AuthorID),
	})
	if err != nil {
		return feedback.CoreItem{}, normalizeError(err)
	}
	row, err := q.GetWorkspaceFeedbackItem(ctx, feedbacksql.GetWorkspaceFeedbackItemParams{WorkspaceID: input.WorkspaceID, ItemID: itemID})
	if err != nil {
		return feedback.CoreItem{}, normalizeError(err)
	}
	return workspaceItemProjection(row).core(), nil
}

func (r *Repo) GetOrCreateAccountContributor(ctx context.Context, portalID, userID uuid.UUID) (feedback.CoreContributor, error) {
	row, err := r.queries.GetOrCreateAccountFeedbackContributor(ctx, feedbacksql.GetOrCreateAccountFeedbackContributorParams{UserID: userID, PortalID: portalID})
	if err != nil {
		return feedback.CoreContributor{}, normalizeError(err)
	}
	accountID := uuid.Nil
	if row.UserID != nil {
		accountID = *row.UserID
	}
	return feedback.CoreContributor{ID: row.ID, PortalID: row.PortalID, UserID: accountID,
		Kind: row.Kind, JoinedAt: row.CreatedAt}, nil
}

func (r *Repo) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input feedback.CoreUpdateItemStatusInput) (feedback.CoreItem, bool, error) {
	item, changed, updated, err := r.updateItemStatus(ctx, workspaceID, itemID, time.Time{}, input, false)
	if err != nil {
		return feedback.CoreItem{}, false, err
	}
	if !updated {
		return feedback.CoreItem{}, false, feedback.ErrNotFound
	}
	return item, changed, nil
}

func (r *Repo) UpdateItemStatusIfUnchanged(ctx context.Context, workspaceID, itemID uuid.UUID, expectedUpdatedAt time.Time, input feedback.CoreUpdateItemStatusInput) (feedback.CoreItem, bool, bool, error) {
	return r.updateItemStatus(ctx, workspaceID, itemID, expectedUpdatedAt, input, true)
}

func (r *Repo) updateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, expected time.Time, input feedback.CoreUpdateItemStatusInput, requireUnchanged bool) (feedback.CoreItem, bool, bool, error) {
	var item feedback.CoreItem
	var changed, updated bool
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		row, err := q.UpdateFeedbackItemStatus(ctx, feedbacksql.UpdateFeedbackItemStatusParams{ActorID: input.ActorID,
			WorkspaceID: workspaceID, ItemID: itemID, AllowLinked: input.AllowLinked, RequireUnchanged: requireUnchanged,
			ExpectedUpdatedAt: expected.UTC(), Status: input.Status, RoadmapSummary: input.RoadmapSummary})
		if errors.Is(err, pgx.ErrNoRows) {
			if requireUnchanged {
				return nil
			}
			if !input.AllowLinked {
				managed, checkErr := q.FeedbackItemStoryManaged(ctx, feedbacksql.FeedbackItemStoryManagedParams{WorkspaceID: workspaceID, ItemID: itemID})
				if checkErr != nil {
					return checkErr
				}
				if managed {
					return feedback.ErrStoryManaged
				}
			}
			return nil
		}
		if err != nil {
			return err
		}
		itemRow, err := q.GetWorkspaceFeedbackItem(ctx, feedbacksql.GetWorkspaceFeedbackItemParams{WorkspaceID: workspaceID, ItemID: row.ID})
		if err != nil {
			return normalizeError(err)
		}
		item = workspaceItemProjection(itemRow).core()
		changed, updated = row.StatusChanged, true
		return nil
	})
	return item, changed, updated, err
}

func (r *Repo) TrashItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	count, err := r.queries.TrashFeedbackItem(ctx, feedbacksql.TrashFeedbackItemParams{ActorID: actorIDForItem(itemID), WorkspaceID: workspaceID, ItemID: itemID})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	state, err := r.queries.GetFeedbackItemProtection(ctx, feedbacksql.GetFeedbackItemProtectionParams{WorkspaceID: workspaceID, ItemID: itemID})
	if err != nil {
		return err
	}
	if state.MergeProtected {
		return feedback.ErrMergeConflict
	}
	if state.StoryManaged {
		return feedback.ErrStoryManaged
	}
	return feedback.ErrNotFound
}

func (r *Repo) TrashItemScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) error {
	if _, err := r.GetItemScoped(ctx, scope, itemID); err != nil {
		return err
	}
	return r.trashItem(ctx, scope.ActorID, scope.WorkspaceID, itemID)
}

func (r *Repo) trashItem(ctx context.Context, actorID, workspaceID, itemID uuid.UUID) error {
	count, err := r.queries.TrashFeedbackItem(ctx, feedbacksql.TrashFeedbackItemParams{ActorID: actorID, WorkspaceID: workspaceID, ItemID: itemID})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	state, err := r.queries.GetFeedbackItemProtection(ctx, feedbacksql.GetFeedbackItemProtectionParams{WorkspaceID: workspaceID, ItemID: itemID})
	if err != nil {
		return err
	}
	if state.MergeProtected {
		return feedback.ErrMergeConflict
	}
	if state.StoryManaged {
		return feedback.ErrStoryManaged
	}
	return feedback.ErrNotFound
}

func (r *Repo) RestoreItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	cutoff := time.Now().Add(-feedbackTrashRecoveryWindow)
	count, err := r.queries.RestoreFeedbackItem(ctx, feedbacksql.RestoreFeedbackItemParams{
		ActorID: actorIDForItem(itemID), WorkspaceID: workspaceID, ItemID: itemID, RecoveryCutoff: &cutoff,
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) RestoreItemScoped(ctx context.Context, scope feedback.CoreAccessScope, itemID uuid.UUID) error {
	page, err := r.ListItemsScoped(ctx, scope, feedback.CoreListItemsInput{
		WorkspaceID: scope.WorkspaceID, ItemID: itemID, DeletedOnly: true,
		Status: "all", Sort: "newest", Page: 1, PageSize: 1,
	})
	if err != nil {
		return err
	}
	if len(page.Items) != 1 {
		return feedback.ErrNotFound
	}
	cutoff := time.Now().Add(-feedbackTrashRecoveryWindow)
	count, err := r.queries.RestoreFeedbackItem(ctx, feedbacksql.RestoreFeedbackItemParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, ItemID: itemID, RecoveryCutoff: &cutoff,
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func actorIDForItem(itemID uuid.UUID) uuid.UUID {
	// Existing public service signatures do not carry an actor for these two
	// mutations. The service fills an actor-aware contract before these methods
	// are exposed; this sentinel keeps the persistence adapter fail-closed.
	return itemID
}

func pageBounds(page, pageSize int) (int32, int32, error) {
	offset64 := int64(page-1) * int64(pageSize)
	offset, err := safecast.Int64ToInt32(offset64)
	if err != nil {
		return 0, 0, err
	}
	limit, err := safecast.Int32(pageSize + 1)
	if err != nil {
		return 0, 0, err
	}
	return offset, limit, nil
}

func listItemProjection(row feedbacksql.ListFeedbackItemsRow) itemProjection {
	return itemProjection{row.ID, row.WorkspaceID, row.PortalID, row.BoardID, row.ContributorID, row.AuthorID,
		row.AuthorName, row.AuthorEmail, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked,
		row.Following, row.MergedIntoItemID, row.MergedAt, row.MergedByUserID, row.Title, row.Description, row.Slug,
		row.Status, row.VoteCount, row.UpvoteCount, row.DownvoteCount, row.CommentCount, row.RoadmapSummary,
		row.BoardTeamID, row.BoardName, row.BoardSlug, row.BoardColor, row.BoardOrderIndex, row.BoardCreatedAt,
		row.BoardUpdatedAt, uuidPointer(row.PrimaryLinkID), uuidPointer(row.PrimaryStoryID), row.PrimaryStoryTitle,
		nonEmptyStringPointer(row.PrimaryRelationship), row.PrimaryCreatedByUserID, timePointer(row.PrimaryCreatedAt),
		row.ReadAt, row.DeletedAt, row.CreatedAt, row.UpdatedAt}
}

func workspaceItemProjection(row feedbacksql.GetWorkspaceFeedbackItemRow) itemProjection {
	return itemProjection{row.ID, row.WorkspaceID, row.PortalID, row.BoardID, row.ContributorID, row.AuthorID,
		row.AuthorName, row.AuthorEmail, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked,
		row.Following, row.MergedIntoItemID, row.MergedAt, row.MergedByUserID, row.Title, row.Description, row.Slug,
		row.Status, row.VoteCount, row.UpvoteCount, row.DownvoteCount, row.CommentCount, row.RoadmapSummary,
		row.BoardTeamID, row.BoardName, row.BoardSlug, row.BoardColor, row.BoardOrderIndex, row.BoardCreatedAt,
		row.BoardUpdatedAt, uuidPointer(row.PrimaryLinkID), uuidPointer(row.PrimaryStoryID), row.PrimaryStoryTitle,
		nonEmptyStringPointer(row.PrimaryRelationship), row.PrimaryCreatedByUserID, timePointer(row.PrimaryCreatedAt),
		row.ReadAt, row.DeletedAt, row.CreatedAt, row.UpdatedAt}
}

func publicItemProjection(row feedbacksql.GetPublicFeedbackItemRow) itemProjection {
	return itemProjection{row.ID, row.WorkspaceID, row.PortalID, row.BoardID, row.ContributorID, row.AuthorID,
		row.AuthorName, row.AuthorEmail, nonEmptyStringPointer(row.AuthorAvatar), row.ParticipantKind, row.AuthorMasked,
		row.Following, row.MergedIntoItemID, row.MergedAt, row.MergedByUserID, row.Title, row.Description, row.Slug,
		row.Status, row.VoteCount, row.UpvoteCount, row.DownvoteCount, row.CommentCount, row.RoadmapSummary,
		row.BoardTeamID, row.BoardName, row.BoardSlug, row.BoardColor, row.BoardOrderIndex, row.BoardCreatedAt,
		row.BoardUpdatedAt, uuidPointer(row.PrimaryLinkID), uuidPointer(row.PrimaryStoryID), row.PrimaryStoryTitle,
		nonEmptyStringPointer(row.PrimaryRelationship), row.PrimaryCreatedByUserID, timePointer(row.PrimaryCreatedAt),
		row.ReadAt, row.DeletedAt, row.CreatedAt, row.UpdatedAt}
}
