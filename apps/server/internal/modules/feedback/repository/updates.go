package feedbackrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type feedbackUpdateProjection struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	PortalID          uuid.UUID
	Slug              string
	Title             string
	Summary           *string
	Body              string
	CoverImageURL     *string
	Status            string
	PublishedAt       *time.Time
	PublishedByUserID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type publicationOutboxPayload struct {
	SchemaVersion       int                       `json:"schemaVersion"`
	PublicationEventID  uuid.UUID                 `json:"publicationEventId"`
	UpdateID            uuid.UUID                 `json:"updateId"`
	WorkspaceID         uuid.UUID                 `json:"workspaceId"`
	PortalID            uuid.UUID                 `json:"portalId"`
	PortalSlug          string                    `json:"portalSlug"`
	PublishedByUserID   uuid.UUID                 `json:"publishedByUserId"`
	PublicationSequence int64                     `json:"publicationSequence"`
	PublishedAt         time.Time                 `json:"publishedAt"`
	Slug                string                    `json:"slug"`
	Title               string                    `json:"title"`
	Summary             *string                   `json:"summary,omitempty"`
	Body                string                    `json:"body"`
	CoverImageURL       *string                   `json:"coverImageUrl,omitempty"`
	LinkedItems         []feedback.CoreUpdateItem `json:"linkedItems"`
	ContributorAudience []uuid.UUID               `json:"contributorAudience"`
	AccountAudience     []publicationAccountPair  `json:"accountAudience"`
}

type publicationAccountPair struct {
	UserID uuid.UUID `json:"userId"`
	ItemID uuid.UUID `json:"itemId"`
}

func (r *Repo) ListWorkspaceUpdates(context.Context, uuid.UUID, int, int) (feedback.CoreUpdatesPage, error) {
	return feedback.CoreUpdatesPage{}, feedback.ErrForbidden
}

func (r *Repo) ListWorkspaceUpdatesScoped(ctx context.Context, scope feedback.CoreAccessScope, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreUpdatesPage{}, feedback.ErrForbidden
	}
	offset, limit, err := pageBounds(page, pageSize)
	if err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
	rows, err := r.queries.ListWorkspaceFeedbackUpdates(ctx, feedbacksql.ListWorkspaceFeedbackUpdatesParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, RowOffset: offset, RowLimit: limit,
	})
	if err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
	projections := make([]feedbackUpdateProjection, 0, len(rows))
	for _, row := range rows {
		projections = append(projections, feedbackUpdateProjection{row.ID, row.WorkspaceID, row.PortalID, row.Slug,
			row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status, row.PublishedAt,
			row.PublishedByUserID, row.CreatedAt, row.UpdatedAt})
	}
	return r.hydrateUpdatePage(ctx, projections, page, pageSize)
}

func (r *Repo) GetWorkspaceUpdate(context.Context, uuid.UUID, uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
}

func (r *Repo) GetWorkspaceUpdateScoped(ctx context.Context, scope feedback.CoreAccessScope, updateID uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
	}
	row, err := r.queries.GetWorkspaceFeedbackUpdate(ctx, feedbacksql.GetWorkspaceFeedbackUpdateParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, UpdateID: updateID,
	})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, normalizeError(err)
	}
	updates, err := r.hydrateUpdates(ctx, []feedbackUpdateProjection{{row.ID, row.WorkspaceID, row.PortalID, row.Slug,
		row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status, row.PublishedAt,
		row.PublishedByUserID, row.CreatedAt, row.UpdatedAt}})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return updates[0], nil
}

func (r *Repo) CreateUpdate(ctx context.Context, input feedback.CoreUpdateInput) (feedback.CoreFeedbackUpdate, error) {
	if err := input.Access.Validate(); err != nil || input.Access.WorkspaceID != input.WorkspaceID || input.Access.ActorID != input.ActorID {
		return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
	}
	var update feedback.CoreFeedbackUpdate
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		updateID, err := q.CreateFeedbackUpdate(ctx, feedbacksql.CreateFeedbackUpdateParams{
			Title: input.Title, Body: input.Body, Slug: input.Slug, Summary: input.Summary,
			CoverImageURL: input.CoverImageURL, ActorID: input.Access.ActorID,
			WorkspaceID: input.WorkspaceID, PortalID: input.PortalID,
		})
		if err != nil {
			return normalizeUpdateWriteError(err)
		}
		if err := replaceFeedbackUpdateItems(ctx, q, input, updateID); err != nil {
			return err
		}
		row, err := q.GetWorkspaceFeedbackUpdate(ctx, feedbacksql.GetWorkspaceFeedbackUpdateParams{
			ActorID: input.Access.ActorID, WorkspaceID: input.WorkspaceID, UpdateID: updateID,
		})
		if err != nil {
			return normalizeError(err)
		}
		updates, err := hydrateUpdatesWithQueries(ctx, q, []feedbackUpdateProjection{{row.ID, row.WorkspaceID,
			row.PortalID, row.Slug, row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status,
			row.PublishedAt, row.PublishedByUserID, row.CreatedAt, row.UpdatedAt}})
		if err == nil {
			update = updates[0]
		}
		return err
	})
	return update, err
}

func (r *Repo) UpdateUpdate(ctx context.Context, input feedback.CoreUpdateInput) (feedback.CoreFeedbackUpdate, error) {
	if err := input.Access.Validate(); err != nil || input.Access.WorkspaceID != input.WorkspaceID || input.Access.ActorID != input.ActorID {
		return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
	}
	var update feedback.CoreFeedbackUpdate
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		count, err := q.UpdateDraftFeedbackUpdate(ctx, feedbacksql.UpdateDraftFeedbackUpdateParams{
			Title: input.Title, Body: input.Body, Slug: input.Slug, Summary: input.Summary,
			CoverImageURL: input.CoverImageURL, PortalID: input.PortalID, ActorID: input.Access.ActorID,
			WorkspaceID: input.WorkspaceID, UpdateID: input.UpdateID,
		})
		if err != nil {
			return normalizeUpdateWriteError(err)
		}
		if err := requireRowsAffected(count); err != nil {
			return err
		}
		if err := replaceFeedbackUpdateItems(ctx, q, input, input.UpdateID); err != nil {
			return err
		}
		row, err := q.GetWorkspaceFeedbackUpdate(ctx, feedbacksql.GetWorkspaceFeedbackUpdateParams{
			ActorID: input.Access.ActorID, WorkspaceID: input.WorkspaceID, UpdateID: input.UpdateID,
		})
		if err != nil {
			return normalizeError(err)
		}
		updates, err := hydrateUpdatesWithQueries(ctx, q, []feedbackUpdateProjection{{row.ID, row.WorkspaceID,
			row.PortalID, row.Slug, row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status,
			row.PublishedAt, row.PublishedByUserID, row.CreatedAt, row.UpdatedAt}})
		if err == nil {
			update = updates[0]
		}
		return err
	})
	return update, err
}

func replaceFeedbackUpdateItems(ctx context.Context, q feedbacksql.Querier, input feedback.CoreUpdateInput, updateID uuid.UUID) error {
	if err := q.DeleteFeedbackUpdateItems(ctx, feedbacksql.DeleteFeedbackUpdateItemsParams{UpdateID: updateID}); err != nil {
		return err
	}
	if len(input.ItemIDs) == 0 {
		return nil
	}
	count, err := q.InsertFeedbackUpdateItems(ctx, feedbacksql.InsertFeedbackUpdateItemsParams{
		ItemIds: input.ItemIDs, ActorID: input.Access.ActorID, WorkspaceID: input.WorkspaceID,
		PortalID: input.PortalID, UpdateID: updateID, AllTeams: input.Access.AllTeams,
		CredentialTeamIds: input.Access.CredentialTeamIDs,
	})
	if err != nil {
		return err
	}
	if count != int64(len(input.ItemIDs)) {
		return feedback.ErrNotFound
	}
	return nil
}

func (r *Repo) DeleteUpdate(context.Context, uuid.UUID, uuid.UUID) error {
	return feedback.ErrForbidden
}

func (r *Repo) DeleteUpdateScoped(ctx context.Context, scope feedback.CoreAccessScope, updateID uuid.UUID) error {
	if err := scope.Validate(); err != nil {
		return feedback.ErrForbidden
	}
	count, err := r.queries.DeleteDraftFeedbackUpdate(ctx, feedbacksql.DeleteDraftFeedbackUpdateParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, UpdateID: updateID,
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) PublishUpdate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (feedback.CoreFeedbackUpdate, bool, error) {
	return feedback.CoreFeedbackUpdate{}, false, feedback.ErrForbidden
}

func (r *Repo) PublishUpdateScoped(ctx context.Context, scope feedback.CoreAccessScope, updateID uuid.UUID) (feedback.CoreFeedbackUpdate, bool, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, feedback.ErrForbidden
	}
	var update feedback.CoreFeedbackUpdate
	created := false
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		locked, err := q.LockFeedbackUpdateForPublication(ctx, feedbacksql.LockFeedbackUpdateForPublicationParams{
			ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, UpdateID: updateID,
		})
		if err != nil {
			return normalizeError(err)
		}
		if _, err := q.LockFeedbackPublicationItems(ctx, feedbacksql.LockFeedbackPublicationItemsParams{UpdateID: updateID}); err != nil {
			return err
		}
		linkedItems, err := listCanonicalPublicationItems(ctx, q, updateID)
		if err != nil {
			return err
		}
		update = lockedUpdateCore(locked, linkedItems)
		if locked.Status == feedback.FeedbackUpdateStatusPublished {
			return nil
		}
		published, err := q.PublishFeedbackUpdate(ctx, feedbacksql.PublishFeedbackUpdateParams{
			ActorID: uuidPointer(scope.ActorID), WorkspaceID: scope.WorkspaceID, UpdateID: updateID,
		})
		if err != nil {
			return normalizeError(err)
		}
		if published.PublishedAt == nil || published.PublishedByUserID == nil {
			return errors.New("feedback publication mutation returned an incomplete envelope")
		}
		update.Status = feedback.FeedbackUpdateStatusPublished
		update.PublishedAt = published.PublishedAt
		update.PublishedByUserID = published.PublishedByUserID
		update.UpdatedAt = published.UpdatedAt
		itemIDs := make([]uuid.UUID, 0, len(linkedItems))
		for _, item := range linkedItems {
			itemIDs = append(itemIDs, item.ID)
		}
		contributors, err := q.SnapshotFeedbackPublicationContributorAudience(ctx, feedbacksql.SnapshotFeedbackPublicationContributorAudienceParams{
			PortalID: locked.PortalID, ItemIds: itemIDs,
		})
		if err != nil {
			return fmt.Errorf("snapshot contributor publication audience: %w", err)
		}
		accountRows, err := q.SnapshotFeedbackPublicationAccountAudience(ctx, feedbacksql.SnapshotFeedbackPublicationAccountAudienceParams{
			PortalID: locked.PortalID, ItemIds: itemIDs,
		})
		if err != nil {
			return fmt.Errorf("snapshot account publication audience: %w", err)
		}
		accounts := make([]publicationAccountPair, 0, len(accountRows))
		for _, row := range accountRows {
			if row.UserID != nil {
				accounts = append(accounts, publicationAccountPair{UserID: *row.UserID, ItemID: row.ItemID})
			}
		}
		eventID := uuid.New()
		payload, err := json.Marshal(publicationOutboxPayload{SchemaVersion: 1, PublicationEventID: eventID,
			UpdateID: locked.ID, WorkspaceID: locked.WorkspaceID, PortalID: locked.PortalID,
			PortalSlug: locked.PortalSlug, PublishedByUserID: scope.ActorID,
			PublicationSequence: published.PublicationSequence, PublishedAt: *published.PublishedAt,
			Slug: locked.Slug, Title: locked.Title, Summary: locked.Summary, Body: locked.Body,
			CoverImageURL: locked.CoverImageURL, LinkedItems: linkedItems,
			ContributorAudience: contributors, AccountAudience: accounts})
		if err != nil {
			return fmt.Errorf("marshal feedback publication snapshot: %w", err)
		}
		if err := q.InsertFeedbackPublicationOutbox(ctx, feedbacksql.InsertFeedbackPublicationOutboxParams{
			EventID: eventID, UpdateID: locked.ID, WorkspaceID: locked.WorkspaceID, PortalID: locked.PortalID,
			ActorID: uuidPointer(scope.ActorID), PublicationSequence: published.PublicationSequence,
			PublishedAt: *published.PublishedAt, EventPayload: payload,
		}); err != nil {
			return err
		}
		created = true
		return nil
	})
	return update, created, err
}

func (r *Repo) UnpublishUpdate(context.Context, uuid.UUID, uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
}

func (r *Repo) UnpublishUpdateScoped(ctx context.Context, scope feedback.CoreAccessScope, updateID uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreFeedbackUpdate{}, feedback.ErrForbidden
	}
	count, err := r.queries.UnpublishFeedbackUpdate(ctx, feedbacksql.UnpublishFeedbackUpdateParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, UpdateID: updateID,
	})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	if err := requireRowsAffected(count); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return r.GetWorkspaceUpdateScoped(ctx, scope, updateID)
}

func (r *Repo) ListPublicUpdates(ctx context.Context, portalID uuid.UUID, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	offset, limit, err := pageBounds(page, pageSize)
	if err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
	rows, err := r.queries.ListPublicFeedbackUpdates(ctx, feedbacksql.ListPublicFeedbackUpdatesParams{
		PortalID: portalID, RowOffset: offset, RowLimit: limit,
	})
	if err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
	projections := make([]feedbackUpdateProjection, 0, len(rows))
	for _, row := range rows {
		projections = append(projections, feedbackUpdateProjection{row.ID, row.WorkspaceID, row.PortalID, row.Slug,
			row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status, row.PublishedAt,
			row.PublishedByUserID, row.CreatedAt, row.UpdatedAt})
	}
	return r.hydrateUpdatePage(ctx, projections, page, pageSize)
}

func (r *Repo) GetPublicUpdate(ctx context.Context, portalID uuid.UUID, slug string) (feedback.CoreFeedbackUpdate, error) {
	row, err := r.queries.GetPublicFeedbackUpdate(ctx, feedbacksql.GetPublicFeedbackUpdateParams{PortalID: portalID, Slug: slug})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, normalizeError(err)
	}
	updates, err := r.hydrateUpdates(ctx, []feedbackUpdateProjection{{row.ID, row.WorkspaceID, row.PortalID, row.Slug,
		row.Title, row.Summary, row.Body, row.CoverImageURL, row.Status, row.PublishedAt,
		row.PublishedByUserID, row.CreatedAt, row.UpdatedAt}})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return updates[0], nil
}

func (r *Repo) hydrateUpdatePage(ctx context.Context, rows []feedbackUpdateProjection, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	updates, err := r.hydrateUpdates(ctx, rows)
	if err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
	return feedback.CoreUpdatesPage{Updates: updates, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

func (r *Repo) hydrateUpdates(ctx context.Context, rows []feedbackUpdateProjection) ([]feedback.CoreFeedbackUpdate, error) {
	return hydrateUpdatesWithQueries(ctx, r.queries, rows)
}

func hydrateUpdatesWithQueries(ctx context.Context, q feedbacksql.Querier, rows []feedbackUpdateProjection) ([]feedback.CoreFeedbackUpdate, error) {
	if len(rows) == 0 {
		return []feedback.CoreFeedbackUpdate{}, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	itemRows, err := q.ListFeedbackUpdateItems(ctx, feedbacksql.ListFeedbackUpdateItemsParams{UpdateIds: ids})
	if err != nil {
		return nil, err
	}
	itemsByUpdate := make(map[uuid.UUID][]feedback.CoreUpdateItem, len(rows))
	for _, row := range itemRows {
		itemsByUpdate[row.UpdateID] = append(itemsByUpdate[row.UpdateID], feedback.CoreUpdateItem{
			ID: row.ID, Slug: row.Slug, Title: row.Title, Status: row.Status,
		})
	}
	result := make([]feedback.CoreFeedbackUpdate, 0, len(rows))
	for _, row := range rows {
		items := itemsByUpdate[row.ID]
		if items == nil {
			items = []feedback.CoreUpdateItem{}
		}
		result = append(result, feedback.CoreFeedbackUpdate{ID: row.ID, WorkspaceID: row.WorkspaceID,
			PortalID: row.PortalID, Slug: row.Slug, Title: row.Title, Summary: row.Summary, Body: row.Body,
			CoverImageURL: row.CoverImageURL, Status: row.Status, PublishedAt: row.PublishedAt,
			PublishedByUserID: row.PublishedByUserID, LinkedItems: items,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return result, nil
}

func listCanonicalPublicationItems(ctx context.Context, q feedbacksql.Querier, updateID uuid.UUID) ([]feedback.CoreUpdateItem, error) {
	rows, err := q.ListCanonicalFeedbackPublicationItems(ctx, feedbacksql.ListCanonicalFeedbackPublicationItemsParams{UpdateID: updateID})
	if err != nil {
		return nil, err
	}
	items := make([]feedback.CoreUpdateItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedback.CoreUpdateItem{ID: row.ID, Slug: row.Slug, Title: row.Title, Status: row.Status})
	}
	return items, nil
}

func lockedUpdateCore(row feedbacksql.LockFeedbackUpdateForPublicationRow, linkedItems []feedback.CoreUpdateItem) feedback.CoreFeedbackUpdate {
	return feedback.CoreFeedbackUpdate{ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
		Slug: row.Slug, Title: row.Title, Summary: row.Summary, Body: row.Body, CoverImageURL: row.CoverImageURL,
		Status: row.Status, PublishedAt: row.PublishedAt, PublishedByUserID: row.PublishedByUserID,
		LinkedItems: linkedItems, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func normalizeUpdateWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "feedback_updates_portal_slug_unique" {
		return fmt.Errorf("%w: update slug already exists", feedback.ErrInvalidInput)
	}
	return normalizeError(err)
}

func (r *Repo) ClaimPublicationOutboxEvents(ctx context.Context, limit int, staleAfter time.Duration) ([]feedback.CorePublicationOutboxEvent, error) {
	if limit <= 0 {
		return []feedback.CorePublicationOutboxEvent{}, nil
	}
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	staleBefore := time.Now().UTC().Add(-staleAfter)
	rows, err := r.queries.ClaimFeedbackPublicationOutboxEvents(ctx, feedbacksql.ClaimFeedbackPublicationOutboxEventsParams{
		StaleBefore: &staleBefore, RowLimit: rowLimit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CorePublicationOutboxEvent, 0, len(rows))
	for _, row := range rows {
		if row.PublishedByUserID == nil || row.ClaimToken == nil {
			return nil, errors.New("claimed feedback publication has an incomplete actor or claim token")
		}
		result = append(result, feedback.CorePublicationOutboxEvent{EventID: row.PublicationEventID,
			UpdateID: row.UpdateID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
			ActorID: *row.PublishedByUserID, PublicationSequence: row.PublicationSequence,
			PublishedAt: row.PublishedAt, ClaimToken: *row.ClaimToken, AttemptCount: int(row.AttemptCount),
			Payload: append(json.RawMessage(nil), row.EventPayload...)})
	}
	return result, nil
}

func (r *Repo) CompletePublicationOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	count, err := r.queries.CompleteFeedbackPublicationOutboxEvent(ctx, feedbacksql.CompleteFeedbackPublicationOutboxEventParams{
		EventID: eventID, ClaimToken: uuidPointer(claimToken),
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) RetryPublicationOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID, failure string, retryAt time.Time, terminal bool) error {
	status := "retrying"
	var nextAttempt *time.Time
	if terminal {
		status = "failed"
	} else {
		nextAttempt = &retryAt
	}
	failure = strings.TrimSpace(failure)
	if failure == "" {
		failure = "feedback update publication dispatch failed"
	}
	count, err := r.queries.RetryFeedbackPublicationOutboxEvent(ctx, feedbacksql.RetryFeedbackPublicationOutboxEventParams{
		Status: status, NextAttemptAt: nextAttempt, Failure: failure, EventID: eventID,
		ClaimToken: uuidPointer(claimToken),
	})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) ListPublicationDeliveryRecipients(ctx context.Context, portalID uuid.UUID, contributorIDs, itemIDs []uuid.UUID) ([]feedback.CoreDeliveryRecipient, error) {
	if len(contributorIDs) == 0 {
		return []feedback.CoreDeliveryRecipient{}, nil
	}
	rows, err := r.queries.ListFeedbackPublicationDeliveryRecipients(ctx, feedbacksql.ListFeedbackPublicationDeliveryRecipientsParams{
		ContributorIds: contributorIDs, PortalID: portalID, ItemIds: itemIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreDeliveryRecipient, 0, len(rows))
	for _, row := range rows {
		if row.Email == nil {
			continue
		}
		result = append(result, feedback.CoreDeliveryRecipient{ContributorID: row.ContributorID,
			Email: *row.Email, DisplayName: row.DisplayName, Kind: row.Kind})
	}
	return result, nil
}

func (r *Repo) ListAccountPublicationRecipients(ctx context.Context, portalID uuid.UUID, audience []feedback.CoreAccountUpdateRecipient) ([]feedback.CoreAccountUpdateRecipient, error) {
	if len(audience) == 0 {
		return []feedback.CoreAccountUpdateRecipient{}, nil
	}
	userIDs := make([]uuid.UUID, 0, len(audience))
	itemIDs := make([]uuid.UUID, 0, len(audience))
	for _, recipient := range audience {
		userIDs = append(userIDs, recipient.UserID)
		itemIDs = append(itemIDs, recipient.ItemID)
	}
	rows, err := r.queries.ListAccountFeedbackPublicationRecipients(ctx, feedbacksql.ListAccountFeedbackPublicationRecipientsParams{
		PortalID: portalID, UserIds: userIDs, ItemIds: itemIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreAccountUpdateRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreAccountUpdateRecipient{UserID: row.UserID, ItemID: row.ItemID})
	}
	return result, nil
}

func (r *Repo) ResolveNotificationActor(ctx context.Context, actorID, fallbackID uuid.UUID) (uuid.UUID, error) {
	resolved, err := r.queries.ResolveFeedbackNotificationActor(ctx, feedbacksql.ResolveFeedbackNotificationActorParams{
		CandidateIds: []uuid.UUID{actorID, fallbackID},
	})
	return resolved, normalizeError(err)
}
