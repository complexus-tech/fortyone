package feedbackrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var _ feedback.PublicationOutboxRepository = (*Repo)(nil)

const (
	lockUpdateForPublicationQuery = `
		SELECT fu.id, fu.workspace_id, fu.portal_id, workspace.slug AS portal_slug,
			fu.slug, fu.title, fu.summary, fu.body, fu.cover_image_url, fu.status,
			fu.published_at, fu.published_by_user_id, fu.publication_sequence,
			fu.created_at, fu.updated_at
		FROM feedback_updates fu
		INNER JOIN feedback_portals portal ON portal.id = fu.portal_id AND portal.workspace_id = fu.workspace_id
		-- Public feedback portal URLs use the owning workspace slug; the portal
		-- table intentionally has no independent slug column.
		INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id
		WHERE fu.workspace_id = $1 AND fu.id = $2
		FOR UPDATE OF fu
	`
	lockPublicationItemsQuery = `
		SELECT item.id
		FROM feedback_update_items link
		INNER JOIN feedback_items item ON item.id = link.item_id
		WHERE link.update_id = $1
		ORDER BY item.id
		FOR SHARE OF item
	`
	mutateUpdateForPublicationQuery = `
		UPDATE feedback_updates
		SET status = 'published', published_at = NOW(), published_by_user_id = $3,
			publication_sequence = publication_sequence + 1, updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2 AND status = 'draft'
		RETURNING published_at, published_by_user_id, publication_sequence, updated_at
	`
	insertPublicationOutboxQuery = `
		INSERT INTO feedback_update_publication_outbox (
			publication_event_id, update_id, workspace_id, portal_id, published_by_user_id,
			publication_sequence, published_at, event_payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, CAST($8 AS jsonb))
	`
	claimPublicationOutboxQuery = `
		WITH candidates AS (
			SELECT publication_event_id
			FROM feedback_update_publication_outbox
			WHERE (status IN ('pending', 'retrying') AND next_attempt_at <= NOW())
				OR (status = 'processing' AND claimed_at <= NOW() - CAST($2 AS interval))
			ORDER BY COALESCE(next_attempt_at, claimed_at), created_at, publication_event_id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		), claimed AS (
			UPDATE feedback_update_publication_outbox outbox
			SET status = 'processing', attempt_count = attempt_count + 1,
				next_attempt_at = NULL, claim_token = gen_random_uuid(), claimed_at = NOW(),
				completed_at = NULL, last_error = NULL, updated_at = NOW()
			FROM candidates
			WHERE outbox.publication_event_id = candidates.publication_event_id
			RETURNING outbox.*
		)
		SELECT publication_event_id, update_id, workspace_id, portal_id,
			published_by_user_id, publication_sequence, published_at, claim_token,
			attempt_count, event_payload
		FROM claimed
		ORDER BY published_at, publication_event_id
	`
	completePublicationOutboxQuery = `
		UPDATE feedback_update_publication_outbox
		SET status = 'completed', next_attempt_at = NULL, claim_token = NULL, claimed_at = NULL,
			completed_at = NOW(), last_error = NULL, updated_at = NOW()
		WHERE publication_event_id = $1 AND claim_token = $2 AND status = 'processing'
	`
	retryPublicationOutboxQuery = `
		UPDATE feedback_update_publication_outbox
		SET status = $3, next_attempt_at = $4, claim_token = NULL, claimed_at = NULL,
			completed_at = NULL, last_error = LEFT($5, 4000), updated_at = NOW()
		WHERE publication_event_id = $1 AND claim_token = $2 AND status = 'processing'
	`
	snapshotPublicationContributorAudienceQuery = `
		WITH linked_items AS (
			SELECT item.id
			FROM feedback_items item
			WHERE item.portal_id = $1 AND item.id = ANY(CAST($2 AS uuid[]))
				AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		), recipient_ids AS (
			SELECT follower.contributor_id
			FROM linked_items linked
			INNER JOIN feedback_item_followers follower
				ON follower.item_id = linked.id AND follower.unsubscribed_at IS NULL
			UNION
			SELECT follower.contributor_id
			FROM feedback_portal_followers follower
			WHERE follower.portal_id = $1 AND follower.unsubscribed_at IS NULL
		)
		SELECT contributor.id
		FROM (SELECT DISTINCT contributor_id FROM recipient_ids) recipient
		INNER JOIN feedback_contributors contributor ON contributor.id = recipient.contributor_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
		WHERE contributor.portal_id = $1
			AND contributor.kind IN ('verified_guest', 'external')
			AND contributor.email IS NOT NULL AND contributor.blocked_at IS NULL
			AND preference.email_unsubscribed_at IS NULL
		ORDER BY contributor.id
	`
	snapshotPublicationAccountAudienceQuery = `
		WITH linked_items AS (
			SELECT item.id AS item_id
			FROM feedback_items item
			WHERE item.portal_id = $1 AND item.id = ANY(CAST($2 AS uuid[]))
				AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		), candidates AS (
			SELECT contributor.user_id, follower.item_id
			FROM linked_items linked
			INNER JOIN feedback_item_followers follower
				ON follower.item_id = linked.item_id AND follower.unsubscribed_at IS NULL
			INNER JOIN feedback_contributors contributor
				ON contributor.id = follower.contributor_id AND contributor.portal_id = $1
			WHERE contributor.kind = 'account' AND contributor.user_id IS NOT NULL
				AND contributor.blocked_at IS NULL
			UNION ALL
			SELECT contributor.user_id, linked.item_id
			FROM feedback_portal_followers follower
			INNER JOIN feedback_contributors contributor
				ON contributor.id = follower.contributor_id AND contributor.portal_id = follower.portal_id
			CROSS JOIN LATERAL (SELECT item_id FROM linked_items ORDER BY item_id LIMIT 1) linked
			WHERE follower.portal_id = $1 AND follower.unsubscribed_at IS NULL
				AND contributor.kind = 'account' AND contributor.user_id IS NOT NULL
				AND contributor.blocked_at IS NULL
		)
		SELECT DISTINCT candidate.user_id, candidate.item_id
		FROM candidates candidate
		INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
		ORDER BY candidate.user_id, candidate.item_id
	`
	listPublicationDeliveryRecipientsQuery = `
		SELECT contributor.id AS contributor_id, contributor.email,
			COALESCE(NULLIF(trim(contributor.display_name), ''), 'there') AS display_name,
			contributor.kind
		FROM unnest(CAST($2 AS uuid[])) audience(contributor_id)
		INNER JOIN feedback_contributors contributor ON contributor.id = audience.contributor_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
		WHERE contributor.portal_id = $1
			AND contributor.kind IN ('verified_guest', 'external')
			AND contributor.email IS NOT NULL AND contributor.blocked_at IS NULL
			AND preference.email_unsubscribed_at IS NULL
			AND (
				EXISTS (
					SELECT 1
					FROM feedback_items published_item
					INNER JOIN feedback_items canonical_item
						ON canonical_item.id = COALESCE(published_item.merged_into_item_id, published_item.id)
					INNER JOIN feedback_item_followers follower
						ON follower.item_id = canonical_item.id
						AND follower.contributor_id = contributor.id
						AND follower.unsubscribed_at IS NULL
					WHERE published_item.portal_id = $1
						AND published_item.id = ANY(CAST($3 AS uuid[]))
						AND canonical_item.deleted_at IS NULL
						AND canonical_item.merged_into_item_id IS NULL
				)
				OR EXISTS (
					SELECT 1
					FROM feedback_portal_followers follower
					WHERE follower.portal_id = $1 AND follower.contributor_id = contributor.id
						AND follower.unsubscribed_at IS NULL
				)
			)
		ORDER BY contributor.id
	`
	listAccountPublicationRecipientsQuery = `
		WITH audience AS (
			SELECT users.user_id, items.item_id
			FROM unnest(CAST($2 AS uuid[])) WITH ORDINALITY users(user_id, ordinal)
			INNER JOIN unnest(CAST($3 AS uuid[])) WITH ORDINALITY items(item_id, ordinal)
				ON items.ordinal = users.ordinal
		), current_items AS (
			SELECT audience.user_id, canonical_item.id AS item_id
			FROM audience
			INNER JOIN feedback_items published_item
				ON published_item.id = audience.item_id AND published_item.portal_id = $1
			INNER JOIN feedback_items canonical_item
				ON canonical_item.id = COALESCE(published_item.merged_into_item_id, published_item.id)
			WHERE canonical_item.portal_id = $1 AND canonical_item.deleted_at IS NULL
				AND canonical_item.merged_into_item_id IS NULL
		)
		SELECT DISTINCT current.user_id, current.item_id
		FROM current_items current
		INNER JOIN users account ON account.user_id = current.user_id AND account.is_active = true
		WHERE EXISTS (
			SELECT 1
			FROM feedback_contributors contributor
			WHERE contributor.portal_id = $1 AND contributor.kind = 'account'
				AND contributor.user_id = current.user_id AND contributor.blocked_at IS NULL
				AND (
					EXISTS (
						SELECT 1
						FROM feedback_item_followers follower
						WHERE follower.item_id = current.item_id
							AND follower.contributor_id = contributor.id
							AND follower.unsubscribed_at IS NULL
					)
					OR EXISTS (
						SELECT 1
						FROM feedback_portal_followers follower
						WHERE follower.portal_id = $1
							AND follower.contributor_id = contributor.id
							AND follower.unsubscribed_at IS NULL
					)
				)
		)
		ORDER BY current.user_id, current.item_id
	`
)

type publicationUpdateRow struct {
	ID                  uuid.UUID  `db:"id"`
	WorkspaceID         uuid.UUID  `db:"workspace_id"`
	PortalID            uuid.UUID  `db:"portal_id"`
	PortalSlug          string     `db:"portal_slug"`
	Slug                string     `db:"slug"`
	Title               string     `db:"title"`
	Summary             *string    `db:"summary"`
	Body                string     `db:"body"`
	CoverImageURL       *string    `db:"cover_image_url"`
	Status              string     `db:"status"`
	PublishedAt         *time.Time `db:"published_at"`
	PublishedByUserID   *uuid.UUID `db:"published_by_user_id"`
	PublicationSequence int64      `db:"publication_sequence"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
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
	UserID uuid.UUID `db:"user_id" json:"userId"`
	ItemID uuid.UUID `db:"item_id" json:"itemId"`
}

type publicationClaimRow struct {
	EventID             uuid.UUID  `db:"publication_event_id"`
	UpdateID            uuid.UUID  `db:"update_id"`
	WorkspaceID         uuid.UUID  `db:"workspace_id"`
	PortalID            uuid.UUID  `db:"portal_id"`
	ActorID             *uuid.UUID `db:"published_by_user_id"`
	PublicationSequence int64      `db:"publication_sequence"`
	PublishedAt         time.Time  `db:"published_at"`
	ClaimToken          uuid.UUID  `db:"claim_token"`
	AttemptCount        int        `db:"attempt_count"`
	Payload             []byte     `db:"event_payload"`
}

func (r *Repo) publishUpdateWithOutbox(ctx context.Context, workspaceID, updateID, actorID uuid.UUID) (feedback.CoreFeedbackUpdate, bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	defer tx.Rollback()

	var row publicationUpdateRow
	if err := tx.GetContext(ctx, &row, lockUpdateForPublicationQuery, workspaceID, updateID); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	if err := lockPublicationItems(ctx, tx, updateID); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	linkedItems, err := publicationLinkedItems(ctx, tx, updateID)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	if row.Status == feedback.FeedbackUpdateStatusPublished {
		if err := tx.Commit(); err != nil {
			return feedback.CoreFeedbackUpdate{}, false, err
		}
		return feedback.CoreFeedbackUpdate{
			ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
			Slug: row.Slug, Title: row.Title, Summary: row.Summary, Body: row.Body,
			CoverImageURL: row.CoverImageURL, Status: row.Status,
			PublishedAt: row.PublishedAt, PublishedByUserID: row.PublishedByUserID,
			LinkedItems: linkedItems, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}, false, nil
	}

	var published struct {
		PublishedAt         time.Time `db:"published_at"`
		PublishedByUserID   uuid.UUID `db:"published_by_user_id"`
		PublicationSequence int64     `db:"publication_sequence"`
		UpdatedAt           time.Time `db:"updated_at"`
	}
	if err := tx.GetContext(ctx, &published, mutateUpdateForPublicationQuery, workspaceID, updateID, actorID); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	row.Status = feedback.FeedbackUpdateStatusPublished
	row.PublishedAt = &published.PublishedAt
	row.PublishedByUserID = &published.PublishedByUserID
	row.PublicationSequence = published.PublicationSequence
	row.UpdatedAt = published.UpdatedAt

	itemIDs := make([]uuid.UUID, 0, len(linkedItems))
	for _, item := range linkedItems {
		itemIDs = append(itemIDs, item.ID)
	}
	contributorAudience, accountAudience, err := publicationAudience(ctx, tx, row.PortalID, itemIDs)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, false, fmt.Errorf("snapshot feedback publication audience: %w", err)
	}

	eventID := uuid.New()
	payload, err := json.Marshal(publicationOutboxPayload{
		SchemaVersion: 1, PublicationEventID: eventID, UpdateID: row.ID,
		WorkspaceID: row.WorkspaceID, PortalID: row.PortalID, PortalSlug: row.PortalSlug,
		PublishedByUserID: actorID, PublicationSequence: row.PublicationSequence,
		PublishedAt: published.PublishedAt, Slug: row.Slug, Title: row.Title,
		Summary: row.Summary, Body: row.Body, CoverImageURL: row.CoverImageURL,
		LinkedItems: linkedItems, ContributorAudience: contributorAudience,
		AccountAudience: accountAudience,
	})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, false, fmt.Errorf("marshal feedback publication snapshot: %w", err)
	}
	if _, err := tx.ExecContext(ctx, insertPublicationOutboxQuery,
		eventID, row.ID, row.WorkspaceID, row.PortalID, actorID,
		row.PublicationSequence, published.PublishedAt, payload,
	); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreFeedbackUpdate{}, false, err
	}

	return feedback.CoreFeedbackUpdate{
		ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
		Slug: row.Slug, Title: row.Title, Summary: row.Summary, Body: row.Body,
		CoverImageURL: row.CoverImageURL, Status: row.Status,
		PublishedAt: row.PublishedAt, PublishedByUserID: row.PublishedByUserID,
		LinkedItems: linkedItems, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, true, nil
}

func lockPublicationItems(ctx context.Context, tx *sqlx.Tx, updateID uuid.UUID) error {
	lockedItemIDs := make([]uuid.UUID, 0)
	return tx.SelectContext(ctx, &lockedItemIDs, lockPublicationItemsQuery, updateID)
}

func publicationLinkedItems(ctx context.Context, tx *sqlx.Tx, updateID uuid.UUID) ([]feedback.CoreUpdateItem, error) {
	var rows []updateItemRow
	if err := tx.SelectContext(ctx, &rows, `
		WITH canonical_links AS (
			SELECT DISTINCT link.update_id, COALESCE(source.merged_into_item_id, source.id) AS item_id
			FROM feedback_update_items link
			INNER JOIN feedback_items source ON source.id = link.item_id
			WHERE link.update_id = $1 AND source.deleted_at IS NULL
		)
		SELECT link.update_id, item.id, item.slug, item.title, `+projectedFeedbackStatus+` AS status
		FROM canonical_links link
		INNER JOIN feedback_items item ON item.id = link.item_id
		LEFT JOIN LATERAL (
			SELECT story_link.id, story_link.story_id
			FROM feedback_story_links story_link
			WHERE story_link.item_id = item.id AND story_link.is_primary = true
			LIMIT 1
		) primary_link ON true
		LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
		LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
		WHERE item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		ORDER BY item.title, item.id
	`, updateID); err != nil {
		return nil, err
	}
	items := make([]feedback.CoreUpdateItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedback.CoreUpdateItem{ID: row.ID, Slug: row.Slug, Title: row.Title, Status: row.Status})
	}
	return items, nil
}

func publicationAudience(ctx context.Context, tx *sqlx.Tx, portalID uuid.UUID, itemIDs []uuid.UUID) ([]uuid.UUID, []publicationAccountPair, error) {
	contributorAudience := make([]uuid.UUID, 0)
	if err := tx.SelectContext(
		ctx,
		&contributorAudience,
		snapshotPublicationContributorAudienceQuery,
		portalID,
		pq.Array(itemIDs),
	); err != nil {
		return nil, nil, err
	}

	accountAudience := make([]publicationAccountPair, 0)
	if err := tx.SelectContext(
		ctx,
		&accountAudience,
		snapshotPublicationAccountAudienceQuery,
		portalID,
		pq.Array(itemIDs),
	); err != nil {
		return nil, nil, err
	}
	return contributorAudience, accountAudience, nil
}

func (r *Repo) ClaimPublicationOutboxEvents(ctx context.Context, limit int, staleAfter time.Duration) ([]feedback.CorePublicationOutboxEvent, error) {
	if limit <= 0 {
		return []feedback.CorePublicationOutboxEvent{}, nil
	}
	var rows []publicationClaimRow
	if err := r.db.SelectContext(ctx, &rows, claimPublicationOutboxQuery, limit, intervalLiteral(staleAfter)); err != nil {
		return nil, err
	}
	result := make([]feedback.CorePublicationOutboxEvent, 0, len(rows))
	for _, row := range rows {
		actorID := uuid.Nil
		if row.ActorID != nil {
			actorID = *row.ActorID
		}
		result = append(result, feedback.CorePublicationOutboxEvent{
			EventID: row.EventID, UpdateID: row.UpdateID, WorkspaceID: row.WorkspaceID,
			PortalID: row.PortalID, ActorID: actorID,
			PublicationSequence: row.PublicationSequence, PublishedAt: row.PublishedAt,
			ClaimToken: row.ClaimToken, AttemptCount: row.AttemptCount,
			Payload: append(json.RawMessage(nil), row.Payload...),
		})
	}
	return result, nil
}

func (r *Repo) CompletePublicationOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, completePublicationOutboxQuery, eventID, claimToken)
	if err != nil {
		return err
	}
	return requireAffected(result)
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
		failure = "feedback Update publication dispatch failed"
	}
	result, err := r.db.ExecContext(ctx, retryPublicationOutboxQuery,
		eventID, claimToken, status, nextAttempt, failure,
	)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *Repo) ListPublicationDeliveryRecipients(ctx context.Context, portalID uuid.UUID, contributorIDs, itemIDs []uuid.UUID) ([]feedback.CoreDeliveryRecipient, error) {
	if len(contributorIDs) == 0 {
		return []feedback.CoreDeliveryRecipient{}, nil
	}
	var rows []struct {
		ContributorID uuid.UUID `db:"contributor_id"`
		Email         string    `db:"email"`
		DisplayName   string    `db:"display_name"`
		Kind          string    `db:"kind"`
	}
	if err := r.db.SelectContext(
		ctx,
		&rows,
		listPublicationDeliveryRecipientsQuery,
		portalID,
		pq.Array(contributorIDs),
		pq.Array(itemIDs),
	); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreDeliveryRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreDeliveryRecipient{
			ContributorID: row.ContributorID, Email: row.Email,
			DisplayName: row.DisplayName, Kind: row.Kind,
		})
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
	var rows []struct {
		UserID uuid.UUID `db:"user_id"`
		ItemID uuid.UUID `db:"item_id"`
	}
	if err := r.db.SelectContext(
		ctx,
		&rows,
		listAccountPublicationRecipientsQuery,
		portalID,
		pq.Array(userIDs),
		pq.Array(itemIDs),
	); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreAccountUpdateRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreAccountUpdateRecipient{UserID: row.UserID, ItemID: row.ItemID})
	}
	return result, nil
}

func (r *Repo) ResolveNotificationActor(ctx context.Context, actorID, fallbackID uuid.UUID) (uuid.UUID, error) {
	var resolved uuid.UUID
	if err := r.db.GetContext(ctx, &resolved, `
		SELECT candidate.user_id
		FROM unnest(CAST($1 AS uuid[])) WITH ORDINALITY candidate(user_id, priority)
		INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
		WHERE candidate.user_id <> '00000000-0000-0000-0000-000000000000'
		ORDER BY candidate.priority
		LIMIT 1
	`, pq.Array([]uuid.UUID{actorID, fallbackID})); err != nil {
		return uuid.Nil, err
	}
	return resolved, nil
}
