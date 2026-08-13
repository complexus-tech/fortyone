package feedbackrepository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const maxVerificationCodeAttempts = 5

var _ feedback.NextPhaseRepository = (*Repo)(nil)

type participantRow struct {
	ID              uuid.UUID  `db:"id"`
	PortalID        uuid.UUID  `db:"portal_id"`
	UserID          *uuid.UUID `db:"user_id"`
	Kind            string     `db:"kind"`
	Email           *string    `db:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	DisplayName     *string    `db:"display_name"`
	AvatarURL       *string    `db:"avatar_url"`
	ExternalID      *string    `db:"external_id"`
	PublicMasked    bool       `db:"public_masked"`
	BlockedAt       *time.Time `db:"blocked_at"`
	BlockedReason   *string    `db:"blocked_reason"`
	LastSeenAt      *time.Time `db:"last_seen_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type contributorSessionRow struct {
	ID            uuid.UUID  `db:"id"`
	PortalID      uuid.UUID  `db:"portal_id"`
	ContributorID uuid.UUID  `db:"contributor_id"`
	Source        string     `db:"source"`
	ExpiresAt     time.Time  `db:"expires_at"`
	RevokedAt     *time.Time `db:"revoked_at"`
	LastUsedAt    *time.Time `db:"last_used_at"`
	CreatedAt     time.Time  `db:"created_at"`
}

type verificationRow struct {
	ID           uuid.UUID  `db:"id"`
	PortalID     uuid.UUID  `db:"portal_id"`
	Email        string     `db:"email"`
	DisplayName  *string    `db:"display_name"`
	PublicMasked bool       `db:"public_masked"`
	TokenHash    []byte     `db:"token_hash"`
	CodeHash     []byte     `db:"code_hash"`
	Source       string     `db:"source"`
	ExpiresAt    time.Time  `db:"expires_at"`
	ConsumedAt   *time.Time `db:"consumed_at"`
	AttemptCount int        `db:"attempt_count"`
	CreatedAt    time.Time  `db:"created_at"`
}

type feedbackUpdateRow struct {
	ID                uuid.UUID  `db:"id"`
	WorkspaceID       uuid.UUID  `db:"workspace_id"`
	PortalID          uuid.UUID  `db:"portal_id"`
	Slug              string     `db:"slug"`
	Title             string     `db:"title"`
	Summary           *string    `db:"summary"`
	Body              string     `db:"body"`
	CoverImageURL     *string    `db:"cover_image_url"`
	Status            string     `db:"status"`
	PublishedAt       *time.Time `db:"published_at"`
	PublishedByUserID *uuid.UUID `db:"published_by_user_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type updateItemRow struct {
	UpdateID uuid.UUID `db:"update_id"`
	ID       uuid.UUID `db:"id"`
	Slug     string    `db:"slug"`
	Title    string    `db:"title"`
	Status   string    `db:"status"`
}

type widgetSettingsRow struct {
	PortalID               uuid.UUID      `db:"portal_id"`
	Enabled                bool           `db:"enabled"`
	WidgetKeyID            uuid.UUID      `db:"widget_key_id"`
	AllowedOrigins         pq.StringArray `db:"allowed_origins"`
	SigningSecretEncrypted *string        `db:"signing_secret_encrypted"`
	SigningSecretVersion   int            `db:"signing_secret_version"`
	PreviousExpiresAt      *time.Time     `db:"previous_version_expires_at"`
	CreatedAt              time.Time      `db:"created_at"`
	UpdatedAt              time.Time      `db:"updated_at"`
}

type deliveryRow struct {
	ID                 uuid.UUID  `db:"id"`
	PortalID           uuid.UUID  `db:"portal_id"`
	ContributorID      uuid.UUID  `db:"contributor_id"`
	Email              string     `db:"recipient_email"`
	DisplayName        string     `db:"display_name"`
	PortalName         string     `db:"portal_name"`
	PortalSlug         string     `db:"portal_slug"`
	ItemID             *uuid.UUID `db:"item_id"`
	UpdateID           *uuid.UUID `db:"update_id"`
	EventType          string     `db:"event_type"`
	DedupeKey          string     `db:"dedupe_key"`
	Subject            string     `db:"subject"`
	Message            string     `db:"message"`
	DestinationURL     string     `db:"destination_url"`
	Status             string     `db:"status"`
	AttemptCount       int        `db:"attempt_count"`
	FinalFailureReason *string    `db:"final_failure_reason"`
	CreatedAt          time.Time  `db:"created_at"`
	WasCreated         bool       `db:"was_created"`
}

func (r *Repo) CreateContributorVerification(ctx context.Context, input feedback.CoreVerificationRequest) (feedback.CoreVerificationChallenge, error) {
	var challenge feedback.CoreVerificationChallenge
	err := r.db.GetContext(ctx, &challenge, `
		INSERT INTO feedback_contributor_verifications (
			portal_id, email, display_name, public_masked, token_hash, code_hash, source, expires_at
		)
		SELECT $1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8
		WHERE (
			SELECT COUNT(*)
			FROM feedback_contributor_verifications recent
			WHERE recent.portal_id = $1
				AND recent.email = $2
				AND recent.created_at >= NOW() - INTERVAL '10 minutes'
		) < 3
		RETURNING id, expires_at
	`, input.PortalID, input.Email, input.DisplayName, input.PublicMasked, input.TokenHash, input.CodeHash, input.Source, input.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return feedback.CoreVerificationChallenge{}, feedback.ErrVerificationAttempts
	}
	return challenge, err
}

func (r *Repo) ConfirmContributorVerification(ctx context.Context, input feedback.CoreVerificationConfirmation) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	defer tx.Rollback()

	verification, err := lockVerification(ctx, tx, input)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	now := time.Now().UTC()
	if verification.ConsumedAt != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationConsumed
	}
	if !verification.ExpiresAt.After(now) {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationExpired
	}
	if verification.AttemptCount >= maxVerificationCodeAttempts {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationAttempts
	}
	if len(input.CodeHash) > 0 && !hmac.Equal(verification.CodeHash, input.CodeHash) {
		attempts := verification.AttemptCount + 1
		if _, updateErr := tx.ExecContext(ctx, `
			UPDATE feedback_contributor_verifications SET attempt_count = $2 WHERE id = $1
		`, verification.ID, attempts); updateErr != nil {
			return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, updateErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, commitErr
		}
		if attempts >= maxVerificationCodeAttempts {
			return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationAttempts
		}
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrContributorSessionInvalid
	}

	participant, err := upsertVerifiedParticipant(ctx, tx, verification, now)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE feedback_contributor_verifications
		SET consumed_at = $2
		WHERE id = $1 AND consumed_at IS NULL
	`, verification.ID, now); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
		VALUES ($1, $2)
		ON CONFLICT (portal_id, contributor_id) DO NOTHING
	`, input.PortalID, participant.ID); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	session, err := insertContributorSession(ctx, tx, input.PortalID, participant.ID, input.SessionTokenHash, input.Source, input.SessionExpiresAt)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	return participant, session, nil
}

func lockVerification(ctx context.Context, tx *sqlx.Tx, input feedback.CoreVerificationConfirmation) (verificationRow, error) {
	var row verificationRow
	query := `
		SELECT id, portal_id, email, display_name, public_masked, token_hash, code_hash, source,
			expires_at, consumed_at, attempt_count, created_at
		FROM feedback_contributor_verifications
		WHERE portal_id = $1 AND token_hash = $2
		FOR UPDATE
	`
	args := []any{input.PortalID, input.TokenHash}
	if len(input.CodeHash) > 0 {
		query = `
			SELECT id, portal_id, email, display_name, public_masked, token_hash, code_hash, source,
				expires_at, consumed_at, attempt_count, created_at
			FROM feedback_contributor_verifications
			WHERE portal_id = $1 AND email = $2 AND source = $3 AND consumed_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
			FOR UPDATE
		`
		args = []any{input.PortalID, input.Email, input.Source}
	}
	if err := tx.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return verificationRow{}, feedback.ErrContributorSessionInvalid
		}
		return verificationRow{}, err
	}
	if row.Source != input.Source {
		return verificationRow{}, feedback.ErrContributorSessionInvalid
	}
	return row, nil
}

func upsertVerifiedParticipant(ctx context.Context, tx *sqlx.Tx, verification verificationRow, now time.Time) (feedback.CoreParticipant, error) {
	var row participantRow
	err := tx.GetContext(ctx, &row, `
		WITH existing AS (
			SELECT id, kind
			FROM feedback_contributors
			WHERE portal_id = $1 AND lower(email) = lower($2) AND kind <> 'external'
			LIMIT 1
			FOR UPDATE
		), updated AS (
			UPDATE feedback_contributors contributor
			SET email = $2,
				email_verified_at = COALESCE(contributor.email_verified_at, $5),
				display_name = COALESCE(NULLIF($3, ''), contributor.display_name),
				public_masked = CASE WHEN contributor.kind = 'verified_guest' THEN $4 ELSE false END,
				last_seen_at = $5,
				updated_at = $5
			FROM existing
			WHERE contributor.id = existing.id
			RETURNING contributor.*
		), inserted AS (
			INSERT INTO feedback_contributors (
				portal_id, kind, email, email_verified_at, display_name, public_masked, last_seen_at
			)
			SELECT $1, 'verified_guest', $2, $5, NULLIF($3, ''), $4, $5
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING *
		)
		SELECT id, portal_id, user_id, kind, email, email_verified_at, display_name, avatar_url,
			external_id, public_masked, blocked_at, blocked_reason, last_seen_at, created_at, updated_at
		FROM updated
		UNION ALL
		SELECT id, portal_id, user_id, kind, email, email_verified_at, display_name, avatar_url,
			external_id, public_masked, blocked_at, blocked_reason, last_seen_at, created_at, updated_at
		FROM inserted
	`, verification.PortalID, verification.Email, pointerString(verification.DisplayName), verification.PublicMasked, now)
	if err != nil {
		return feedback.CoreParticipant{}, err
	}
	participant := toCoreParticipant(row)
	if participant.BlockedAt != nil {
		return feedback.CoreParticipant{}, feedback.ErrContributorBlocked
	}
	return participant, nil
}

func (r *Repo) GetContributorSession(ctx context.Context, portalID uuid.UUID, tokenHash []byte, source string) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	var row struct {
		participantRow
		SessionID         uuid.UUID  `db:"session_id"`
		SessionSource     string     `db:"session_source"`
		SessionExpiresAt  time.Time  `db:"session_expires_at"`
		SessionRevokedAt  *time.Time `db:"session_revoked_at"`
		SessionLastUsedAt *time.Time `db:"session_last_used_at"`
		SessionCreatedAt  time.Time  `db:"session_created_at"`
	}
	err := r.db.GetContext(ctx, &row, `
		WITH active AS (
			UPDATE feedback_contributor_sessions
			SET last_used_at = NOW()
			WHERE portal_id = $1 AND token_hash = $2
				AND revoked_at IS NULL AND expires_at > NOW()
				AND (
					($3 = '' AND source IN ('portal', 'widget'))
					OR source = $3
				)
			RETURNING *
		)
		SELECT contributor.id, contributor.portal_id, contributor.user_id, contributor.kind,
			COALESCE(contributor.email, account.email) AS email,
			contributor.email_verified_at,
			COALESCE(contributor.display_name, NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS display_name,
			COALESCE(contributor.avatar_url, account.avatar_url) AS avatar_url,
			contributor.external_id, contributor.public_masked, contributor.blocked_at, contributor.blocked_reason,
			contributor.last_seen_at, contributor.created_at, contributor.updated_at,
			active.id AS session_id, active.source AS session_source, active.expires_at AS session_expires_at,
			active.revoked_at AS session_revoked_at, active.last_used_at AS session_last_used_at,
			active.created_at AS session_created_at
		FROM active
		INNER JOIN feedback_contributors contributor ON contributor.id = active.contributor_id AND contributor.portal_id = active.portal_id
		LEFT JOIN users account ON account.user_id = contributor.user_id
	`, portalID, tokenHash, source)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	session := feedback.CoreParticipantSession{
		ID:            row.SessionID,
		PortalID:      row.PortalID,
		ContributorID: row.ID,
		Source:        row.SessionSource,
		ExpiresAt:     row.SessionExpiresAt,
		RevokedAt:     row.SessionRevokedAt,
		LastUsedAt:    row.SessionLastUsedAt,
		CreatedAt:     row.SessionCreatedAt,
	}
	return toCoreParticipant(row.participantRow), session, nil
}

func (r *Repo) RevokeContributorSession(ctx context.Context, portalID uuid.UUID, tokenHash []byte) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE feedback_contributor_sessions
		SET revoked_at = NOW()
		WHERE portal_id = $1 AND token_hash = $2 AND revoked_at IS NULL
	`, portalID, tokenHash)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *Repo) GetParticipantByUser(ctx context.Context, portalID, userID uuid.UUID) (feedback.CoreParticipant, error) {
	var row participantRow
	err := r.db.GetContext(ctx, &row, `
		WITH contributor AS (
			INSERT INTO feedback_contributors (portal_id, user_id, kind, last_seen_at)
			VALUES ($1, $2, 'account', NOW())
			ON CONFLICT (portal_id, user_id) WHERE user_id IS NOT NULL
			DO UPDATE SET last_seen_at = NOW(), updated_at = NOW()
			RETURNING *
		)
		SELECT contributor.id, contributor.portal_id, contributor.user_id, contributor.kind,
			account.email, CAST(NULL AS timestamptz) AS email_verified_at,
			COALESCE(NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS display_name,
			account.avatar_url, contributor.external_id, contributor.public_masked,
			contributor.blocked_at, contributor.blocked_reason, contributor.last_seen_at,
			contributor.created_at, contributor.updated_at
		FROM contributor
		INNER JOIN users account ON account.user_id = contributor.user_id
	`, portalID, userID)
	return toCoreParticipant(row), err
}

func insertContributorSession(ctx context.Context, tx *sqlx.Tx, portalID, contributorID uuid.UUID, tokenHash []byte, source string, expiresAt time.Time) (feedback.CoreParticipantSession, error) {
	var row contributorSessionRow
	err := tx.GetContext(ctx, &row, `
		INSERT INTO feedback_contributor_sessions (portal_id, contributor_id, token_hash, source, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, portal_id, contributor_id, source, expires_at, revoked_at, last_used_at, created_at
	`, portalID, contributorID, tokenHash, source, expiresAt)
	return toCoreContributorSession(row), err
}

func toCoreParticipant(row participantRow) feedback.CoreParticipant {
	participant := feedback.CoreParticipant{
		ID:              row.ID,
		PortalID:        row.PortalID,
		Kind:            row.Kind,
		EmailVerifiedAt: row.EmailVerifiedAt,
		AvatarURL:       row.AvatarURL,
		PublicMasked:    row.PublicMasked,
		BlockedAt:       row.BlockedAt,
		LastSeenAt:      row.LastSeenAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.UserID != nil {
		participant.UserID = *row.UserID
	}
	participant.Email = pointerString(row.Email)
	participant.DisplayName = pointerString(row.DisplayName)
	participant.ExternalID = pointerString(row.ExternalID)
	participant.BlockedReason = pointerString(row.BlockedReason)
	return participant
}

func toCoreContributorSession(row contributorSessionRow) feedback.CoreParticipantSession {
	return feedback.CoreParticipantSession{
		ID: row.ID, PortalID: row.PortalID, ContributorID: row.ContributorID, Source: row.Source,
		ExpiresAt: row.ExpiresAt, RevokedAt: row.RevokedAt, LastUsedAt: row.LastUsedAt, CreatedAt: row.CreatedAt,
	}
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func requireAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return feedback.ErrNotFound
	}
	return nil
}

func (r *Repo) CreateContributorItemAndFollow(ctx context.Context, input feedback.CoreContributorItemInput) (feedback.CoreItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreItem{}, err
	}
	defer tx.Rollback()
	itemInput := input.Item
	if input.Participant.ID == uuid.Nil || input.Participant.PortalID != itemInput.PortalID {
		return feedback.CoreItem{}, feedback.ErrAuthenticationRequired
	}
	var itemID uuid.UUID
	if err := tx.GetContext(ctx, &itemID, `
		INSERT INTO feedback_items (
			workspace_id, portal_id, board_id, contributor_id, author_id, title, description, slug, submission_source
		)
		SELECT $1, $2, $3, contributor.id, contributor.user_id, $5, $6, $7, $8
		FROM feedback_contributors contributor
		WHERE contributor.portal_id = $2 AND contributor.id = $4 AND contributor.blocked_at IS NULL
		RETURNING id
	`, itemInput.WorkspaceID, itemInput.PortalID, itemInput.BoardID, input.Participant.ID, itemInput.Title, itemInput.Description, itemInput.Slug, itemInput.Source); err != nil {
		return feedback.CoreItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_item_followers (item_id, contributor_id)
		VALUES ($1, $2)
		ON CONFLICT (item_id, contributor_id)
		DO UPDATE SET unsubscribed_at = NULL
	`, itemID, input.Participant.ID); err != nil {
		return feedback.CoreItem{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
		VALUES ($1, $2)
		ON CONFLICT (portal_id, contributor_id) DO NOTHING
	`, itemInput.PortalID, input.Participant.ID); err != nil {
		return feedback.CoreItem{}, err
	}
	var row itemRow
	if err := tx.GetContext(ctx, &row, itemSelectQuery()+`
		WHERE fi.id = $1
	`, itemID); err != nil {
		return feedback.CoreItem{}, err
	}
	row.Following = true
	if err := tx.Commit(); err != nil {
		return feedback.CoreItem{}, err
	}
	return toCoreItem(row), nil
}

func (r *Repo) CreateContributorComment(ctx context.Context, input feedback.CoreContributorCommentInput) (feedback.CoreComment, error) {
	var row commentRow
	err := r.db.GetContext(ctx, &row, `
		WITH inserted AS (
			INSERT INTO feedback_comments (workspace_id, item_id, author_id, contributor_id, parent_id, body)
			SELECT $1, item.id, contributor.user_id, contributor.id, $5, $6
			FROM feedback_items item
			INNER JOIN feedback_contributors contributor
				ON contributor.portal_id = item.portal_id AND contributor.id = $4
			WHERE item.workspace_id = $1 AND item.portal_id = $2 AND item.id = $3
				AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
				AND contributor.blocked_at IS NULL
			RETURNING *
		)
		SELECT inserted.id, inserted.workspace_id, inserted.item_id, inserted.author_id,
			inserted.contributor_id, inserted.parent_id,
			CASE
				WHEN contributor.kind = 'anonymous'
					OR (contributor.kind IN ('verified_guest', 'external')
						AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN 'Anonymous'
				WHEN contributor.kind IN ('verified_guest', 'external') THEN COALESCE(NULLIF(trim(contributor.display_name), ''), 'Guest')
				ELSE COALESCE(NULLIF(trim(account.full_name), ''), NULLIF(trim(account.username), ''), 'Anonymous')
			END AS author_name,
			CASE
				WHEN contributor.kind = 'anonymous'
					OR (contributor.kind IN ('verified_guest', 'external')
						AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests')) THEN NULL
				WHEN contributor.kind = 'account' THEN account.avatar_url
				ELSE contributor.avatar_url
			END AS author_avatar,
			contributor.kind AS participant_kind,
			(contributor.kind = 'anonymous' OR (contributor.kind IN ('verified_guest', 'external')
				AND (contributor.public_masked OR portal.guest_identity_policy = 'always_mask_guests'))) AS author_masked,
			inserted.body, inserted.created_at, inserted.updated_at
		FROM inserted
		INNER JOIN feedback_items item ON item.id = inserted.item_id
		INNER JOIN feedback_portals portal ON portal.id = item.portal_id
		INNER JOIN feedback_contributors contributor ON contributor.id = inserted.contributor_id
		LEFT JOIN users account ON account.user_id = contributor.user_id
	`, input.WorkspaceID, input.PortalID, input.ItemID, input.Participant.ID, input.ParentID, input.Body)
	if err != nil {
		return feedback.CoreComment{}, err
	}
	return toCoreComment(row), nil
}

func (r *Repo) ToggleContributorVote(ctx context.Context, input feedback.CoreContributorVoteInput) (feedback.CoreVoteResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreVoteResult{}, err
	}
	defer tx.Rollback()
	var current int
	if err := tx.GetContext(ctx, &current, `
		SELECT COALESCE((
			SELECT vote.direction
			FROM feedback_votes vote
			INNER JOIN feedback_items item ON item.id = vote.item_id
			WHERE vote.workspace_id = $1 AND vote.item_id = $2 AND vote.contributor_id = $3
				AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		), 0)
	`, input.WorkspaceID, input.ItemID, input.Participant.ID); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	resultingVote := input.Vote
	if current == input.Vote {
		result, err := tx.ExecContext(ctx, `
			DELETE FROM feedback_votes vote
			USING feedback_items item
			WHERE item.id = vote.item_id AND vote.workspace_id = $1 AND vote.item_id = $2
				AND vote.contributor_id = $3 AND item.deleted_at IS NULL
				AND item.merged_into_item_id IS NULL
		`, input.WorkspaceID, input.ItemID, input.Participant.ID)
		if err != nil {
			return feedback.CoreVoteResult{}, err
		}
		if err := requireAffected(result); err != nil {
			return feedback.CoreVoteResult{}, err
		}
		resultingVote = 0
	} else {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO feedback_votes (workspace_id, item_id, user_id, contributor_id, direction)
			SELECT item.workspace_id, item.id, contributor.user_id, contributor.id, $4
			FROM feedback_items item
			INNER JOIN feedback_contributors contributor
				ON contributor.portal_id = item.portal_id AND contributor.id = $3
			WHERE item.workspace_id = $1 AND item.id = $2 AND item.deleted_at IS NULL
				AND item.merged_into_item_id IS NULL
				AND contributor.blocked_at IS NULL AND contributor.kind <> 'anonymous'
			ON CONFLICT (item_id, contributor_id)
			DO UPDATE SET direction = EXCLUDED.direction
		`, input.WorkspaceID, input.ItemID, input.Participant.ID, input.Vote)
		if err != nil {
			return feedback.CoreVoteResult{}, err
		}
		if err := requireAffected(result); err != nil {
			return feedback.CoreVoteResult{}, err
		}
	}
	var count int
	if err := tx.GetContext(ctx, &count, `
		SELECT CAST(COALESCE(SUM(direction), 0) AS integer)
		FROM feedback_votes WHERE item_id = $1
	`, input.ItemID); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreVoteResult{}, err
	}
	return feedback.CoreVoteResult{Vote: resultingVote, VoteCount: count}, nil
}

func (r *Repo) GetItemFollow(ctx context.Context, itemID, contributorID uuid.UUID) (feedback.CoreFollowState, error) {
	var state struct {
		ItemID        uuid.UUID  `db:"item_id"`
		ItemSlug      string     `db:"item_slug"`
		Title         string     `db:"title"`
		ContributorID uuid.UUID  `db:"contributor_id"`
		Following     bool       `db:"following"`
		CreatedAt     *time.Time `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &state, `
		SELECT item.id AS item_id, item.slug AS item_slug, item.title, CAST($2 AS uuid) AS contributor_id,
			(follower.item_id IS NOT NULL AND follower.unsubscribed_at IS NULL) AS following,
			follower.created_at
		FROM feedback_items item
		LEFT JOIN feedback_item_followers follower
			ON follower.item_id = item.id AND follower.contributor_id = $2
		WHERE item.id = $1 AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
	`, itemID, contributorID)
	if err != nil {
		return feedback.CoreFollowState{}, err
	}
	return feedback.CoreFollowState{
		ItemID: state.ItemID, ItemSlug: state.ItemSlug, Title: state.Title,
		ContributorID: state.ContributorID, Following: state.Following, CreatedAt: state.CreatedAt,
	}, nil
}

func (r *Repo) SetItemFollow(ctx context.Context, itemID, contributorID uuid.UUID, following bool) (feedback.CoreFollowState, error) {
	if following {
		result, err := r.db.ExecContext(ctx, `
			INSERT INTO feedback_item_followers (item_id, contributor_id)
			SELECT item.id, contributor.id
			FROM feedback_items item
			INNER JOIN feedback_contributors contributor
				ON contributor.portal_id = item.portal_id AND contributor.id = $2
			WHERE item.id = $1 AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
				AND contributor.blocked_at IS NULL
			ON CONFLICT (item_id, contributor_id)
			DO UPDATE SET unsubscribed_at = NULL
		`, itemID, contributorID)
		if err != nil {
			return feedback.CoreFollowState{}, err
		}
		if err := requireAffected(result); err != nil {
			return feedback.CoreFollowState{}, err
		}
	} else {
		if _, err := r.GetItemFollow(ctx, itemID, contributorID); err != nil {
			return feedback.CoreFollowState{}, err
		}
		_, err := r.db.ExecContext(ctx, `
			UPDATE feedback_item_followers follower
			SET unsubscribed_at = COALESCE(follower.unsubscribed_at, NOW())
			FROM feedback_items item
			WHERE item.id = follower.item_id AND follower.item_id = $1 AND follower.contributor_id = $2
				AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		`, itemID, contributorID)
		if err != nil {
			return feedback.CoreFollowState{}, err
		}
	}
	return r.GetItemFollow(ctx, itemID, contributorID)
}

func (r *Repo) GetContributorPreferences(ctx context.Context, portalID, contributorID uuid.UUID) (feedback.CoreContributorPreferences, error) {
	var preferences struct {
		PortalEmailsEnabled bool      `db:"portal_emails_enabled"`
		UpdatedAt           time.Time `db:"updated_at"`
	}
	err := r.db.GetContext(ctx, &preferences, `
		WITH preference AS (
			INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
			VALUES ($1, $2)
			ON CONFLICT (portal_id, contributor_id)
			DO UPDATE SET updated_at = feedback_contributor_preferences.updated_at
			RETURNING email_unsubscribed_at, updated_at
		)
		SELECT email_unsubscribed_at IS NULL AS portal_emails_enabled, updated_at FROM preference
	`, portalID, contributorID)
	if err != nil {
		return feedback.CoreContributorPreferences{}, err
	}
	var rows []struct {
		ItemID    uuid.UUID  `db:"item_id"`
		ItemSlug  string     `db:"item_slug"`
		Title     string     `db:"title"`
		Following bool       `db:"following"`
		CreatedAt *time.Time `db:"created_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT item.id AS item_id, item.slug AS item_slug, item.title,
			(follower.unsubscribed_at IS NULL) AS following, follower.created_at
		FROM feedback_item_followers follower
		INNER JOIN feedback_items item ON item.id = follower.item_id
		WHERE item.portal_id = $1 AND follower.contributor_id = $2
			AND item.deleted_at IS NULL AND item.merged_into_item_id IS NULL
		ORDER BY item.created_at DESC, item.id DESC
	`, portalID, contributorID); err != nil {
		return feedback.CoreContributorPreferences{}, err
	}
	follows := make([]feedback.CoreFollowState, 0, len(rows))
	for _, row := range rows {
		follows = append(follows, feedback.CoreFollowState{
			ItemID: row.ItemID, ItemSlug: row.ItemSlug, Title: row.Title, ContributorID: contributorID,
			Following: row.Following, CreatedAt: row.CreatedAt,
		})
	}
	return feedback.CoreContributorPreferences{
		PortalID: portalID, ContributorID: contributorID, PortalEmailsEnabled: preferences.PortalEmailsEnabled,
		ItemFollows: follows, UpdatedAt: preferences.UpdatedAt,
	}, nil
}

func (r *Repo) SetPortalEmailPreference(ctx context.Context, portalID, contributorID uuid.UUID, enabled bool) (feedback.CoreContributorPreferences, error) {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback_contributor_preferences (portal_id, contributor_id, email_unsubscribed_at)
		VALUES ($1, $2, CASE WHEN $3 THEN NULL ELSE NOW() END)
		ON CONFLICT (portal_id, contributor_id)
		DO UPDATE SET email_unsubscribed_at = CASE WHEN $3 THEN NULL ELSE NOW() END, updated_at = NOW()
	`, portalID, contributorID, enabled); err != nil {
		return feedback.CoreContributorPreferences{}, err
	}
	return r.GetContributorPreferences(ctx, portalID, contributorID)
}

func (r *Repo) GetUnreadUpdateCount(ctx context.Context, portalID, contributorID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT CAST(COUNT(*) AS integer)
		FROM feedback_updates update_record
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = $1 AND preference.contributor_id = $2
		WHERE update_record.portal_id = $1
			AND update_record.status = 'published' AND update_record.published_at IS NOT NULL
			AND update_record.published_at > COALESCE(preference.last_seen_update_published_at, CAST('-infinity' AS timestamptz))
	`, portalID, contributorID)
	return count, err
}

func (r *Repo) MarkUpdatesSeen(ctx context.Context, portalID, contributorID uuid.UUID) (time.Time, error) {
	var seenAt time.Time
	err := r.db.GetContext(ctx, &seenAt, `
		INSERT INTO feedback_contributor_preferences (
			portal_id, contributor_id, last_seen_update_published_at
		)
		VALUES ($1, $2, NOW())
		ON CONFLICT (portal_id, contributor_id)
		DO UPDATE SET last_seen_update_published_at = NOW(), updated_at = NOW()
		RETURNING last_seen_update_published_at
	`, portalID, contributorID)
	return seenAt, err
}

func (r *Repo) ConsumeUnsubscribeToken(ctx context.Context, portalID uuid.UUID, tokenHash, sessionTokenHash []byte, sessionExpiresAt time.Time) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	defer tx.Rollback()
	var token struct {
		ID            uuid.UUID  `db:"id"`
		ContributorID uuid.UUID  `db:"contributor_id"`
		ItemID        *uuid.UUID `db:"item_id"`
		Purpose       string     `db:"purpose"`
		ExpiresAt     time.Time  `db:"expires_at"`
		ConsumedAt    *time.Time `db:"consumed_at"`
	}
	if err := tx.GetContext(ctx, &token, `
		SELECT id, contributor_id, item_id, purpose, expires_at, consumed_at
		FROM feedback_contributor_unsubscribe_tokens
		WHERE portal_id = $1 AND token_hash = $2
		FOR UPDATE
	`, portalID, tokenHash); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if token.ConsumedAt != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationConsumed
	}
	if !token.ExpiresAt.After(time.Now().UTC()) {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrVerificationExpired
	}
	if _, err := tx.ExecContext(ctx, `UPDATE feedback_contributor_unsubscribe_tokens SET consumed_at = NOW() WHERE id = $1`, token.ID); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	switch token.Purpose {
	case "unsubscribe_item":
		if token.ItemID != nil {
			if _, err := tx.ExecContext(ctx, `
				UPDATE feedback_item_followers SET unsubscribed_at = COALESCE(unsubscribed_at, NOW())
				WHERE item_id = $1 AND contributor_id = $2
			`, *token.ItemID, token.ContributorID); err != nil {
				return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
			}
		}
	case "unsubscribe_portal":
		if _, err := tx.ExecContext(ctx, `
			UPDATE feedback_portal_followers SET unsubscribed_at = COALESCE(unsubscribed_at, NOW())
			WHERE portal_id = $1 AND contributor_id = $2
		`, portalID, token.ContributorID); err != nil {
			return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
		}
	case "all_email":
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO feedback_contributor_preferences (portal_id, contributor_id, email_unsubscribed_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (portal_id, contributor_id)
			DO UPDATE SET email_unsubscribed_at = NOW(), updated_at = NOW()
		`, portalID, token.ContributorID); err != nil {
			return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
		}
	}
	participant, err := getParticipantTx(ctx, tx, portalID, token.ContributorID)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	session, err := insertContributorSession(ctx, tx, portalID, token.ContributorID, sessionTokenHash, feedback.ContributorSessionSourcePreferences, sessionExpiresAt)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	return participant, session, nil
}

func getParticipantTx(ctx context.Context, tx *sqlx.Tx, portalID, contributorID uuid.UUID) (feedback.CoreParticipant, error) {
	var row participantRow
	err := tx.GetContext(ctx, &row, `
		SELECT contributor.id, contributor.portal_id, contributor.user_id, contributor.kind,
			COALESCE(contributor.email, account.email) AS email, contributor.email_verified_at,
			COALESCE(contributor.display_name, NULLIF(account.full_name, ''), NULLIF(account.username, '')) AS display_name,
			COALESCE(contributor.avatar_url, account.avatar_url) AS avatar_url, contributor.external_id,
			contributor.public_masked, contributor.blocked_at, contributor.blocked_reason,
			contributor.last_seen_at, contributor.created_at, contributor.updated_at
		FROM feedback_contributors contributor
		LEFT JOIN users account ON account.user_id = contributor.user_id
		WHERE contributor.portal_id = $1 AND contributor.id = $2
	`, portalID, contributorID)
	return toCoreParticipant(row), err
}

func (r *Repo) ListDeliveryRecipients(ctx context.Context, portalID uuid.UUID, itemID, updateID *uuid.UUID, actorContributorID uuid.UUID) ([]feedback.CoreDeliveryRecipient, error) {
	var rows []struct {
		ContributorID uuid.UUID `db:"contributor_id"`
		Email         string    `db:"email"`
		DisplayName   string    `db:"display_name"`
		Kind          string    `db:"kind"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		WITH recipient_ids AS (
			SELECT follower.contributor_id
			FROM feedback_item_followers follower
			INNER JOIN feedback_items item ON item.id = follower.item_id
			WHERE item.portal_id = $1 AND follower.unsubscribed_at IS NULL
				AND (CAST($2 AS uuid) IS NOT NULL AND follower.item_id = $2)
			UNION
			SELECT follower.contributor_id
			FROM feedback_item_followers follower
			INNER JOIN feedback_items item ON item.id = follower.item_id
			INNER JOIN feedback_update_items linked ON linked.item_id = item.id
			WHERE item.portal_id = $1 AND follower.unsubscribed_at IS NULL
				AND (CAST($3 AS uuid) IS NOT NULL AND linked.update_id = $3)
			UNION
			SELECT contributor_id
			FROM feedback_portal_followers
			WHERE portal_id = $1 AND unsubscribed_at IS NULL
		), recipients AS (
			SELECT DISTINCT contributor_id FROM recipient_ids
		)
		SELECT contributor.id AS contributor_id, contributor.email,
			COALESCE(NULLIF(trim(contributor.display_name), ''), 'there') AS display_name,
			contributor.kind
		FROM recipients
		INNER JOIN feedback_contributors contributor ON contributor.id = recipients.contributor_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
		WHERE contributor.portal_id = $1
			AND contributor.id <> $4
			AND contributor.kind IN ('verified_guest', 'external')
			AND contributor.email IS NOT NULL
			AND contributor.blocked_at IS NULL
			AND preference.email_unsubscribed_at IS NULL
		ORDER BY contributor.id
	`, portalID, itemID, updateID, actorContributorID)
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreDeliveryRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreDeliveryRecipient{
			ContributorID: row.ContributorID, Email: row.Email, DisplayName: row.DisplayName, Kind: row.Kind,
		})
	}
	return result, nil
}

func (r *Repo) ListAccountUpdateRecipients(ctx context.Context, portalID, updateID uuid.UUID) ([]feedback.CoreAccountUpdateRecipient, error) {
	var rows []struct {
		UserID uuid.UUID `db:"user_id"`
		ItemID uuid.UUID `db:"item_id"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		WITH linked_items AS (
			SELECT link.item_id
			FROM feedback_update_items link
			INNER JOIN feedback_items item ON item.id = link.item_id
			WHERE link.update_id = $2 AND item.portal_id = $1 AND item.deleted_at IS NULL
		), candidates AS (
			SELECT contributor.user_id, follower.item_id
			FROM linked_items linked
			INNER JOIN feedback_item_followers follower
				ON follower.item_id = linked.item_id AND follower.unsubscribed_at IS NULL
			INNER JOIN feedback_contributors contributor
				ON contributor.id = follower.contributor_id AND contributor.portal_id = $1
			WHERE contributor.kind = 'account' AND contributor.user_id IS NOT NULL AND contributor.blocked_at IS NULL
			UNION ALL
			SELECT contributor.user_id, linked.item_id
			FROM feedback_portal_followers follower
			INNER JOIN feedback_contributors contributor
				ON contributor.id = follower.contributor_id AND contributor.portal_id = follower.portal_id
			CROSS JOIN LATERAL (SELECT item_id FROM linked_items ORDER BY item_id LIMIT 1) linked
			WHERE follower.portal_id = $1 AND follower.unsubscribed_at IS NULL
				AND contributor.kind = 'account' AND contributor.user_id IS NOT NULL AND contributor.blocked_at IS NULL
		)
		SELECT DISTINCT ON (candidate.user_id) candidate.user_id, candidate.item_id
		FROM candidates candidate
		INNER JOIN users account ON account.user_id = candidate.user_id AND account.is_active = true
		ORDER BY candidate.user_id, candidate.item_id
	`, portalID, updateID); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreAccountUpdateRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreAccountUpdateRecipient{UserID: row.UserID, ItemID: row.ItemID})
	}
	return result, nil
}

func (r *Repo) ListAccountItemFollowers(ctx context.Context, portalID, itemID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	if err := r.db.SelectContext(ctx, &userIDs, `
		SELECT DISTINCT contributor.user_id
		FROM feedback_item_followers follower
		INNER JOIN feedback_items item
			ON item.id = follower.item_id AND item.portal_id = $1
		INNER JOIN feedback_contributors contributor
			ON contributor.id = follower.contributor_id AND contributor.portal_id = item.portal_id
		INNER JOIN users account
			ON account.user_id = contributor.user_id AND account.is_active = true
		WHERE follower.item_id = $2
			AND follower.unsubscribed_at IS NULL
			AND contributor.kind = 'account'
			AND contributor.user_id IS NOT NULL
			AND contributor.blocked_at IS NULL
		ORDER BY contributor.user_id
	`, portalID, itemID); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *Repo) ListPrimaryStoryItems(ctx context.Context, workspaceID, storyID uuid.UUID) ([]feedback.CoreItem, error) {
	var rows []struct {
		ID             uuid.UUID  `db:"id"`
		WorkspaceID    uuid.UUID  `db:"workspace_id"`
		PortalID       uuid.UUID  `db:"portal_id"`
		ContributorID  uuid.UUID  `db:"contributor_id"`
		AuthorID       *uuid.UUID `db:"author_id"`
		Title          string     `db:"title"`
		Slug           string     `db:"slug"`
		Status         string     `db:"status"`
		RoadmapSummary *string    `db:"roadmap_summary"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fi.id, fi.workspace_id, fi.portal_id, fi.contributor_id, fi.author_id,
			fi.title, fi.slug, `+projectedFeedbackStatus+` AS status,
			fi.roadmap_summary, fi.updated_at
		FROM feedback_story_links primary_link
		INNER JOIN feedback_items fi ON fi.id = primary_link.item_id
		INNER JOIN stories projected_story ON projected_story.id = primary_link.story_id
		LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
		WHERE primary_link.workspace_id = $1 AND primary_link.story_id = $2
			AND primary_link.is_primary = true
			AND fi.deleted_at IS NULL
			AND fi.merged_into_item_id IS NULL
		ORDER BY fi.id
	`, workspaceID, storyID); err != nil {
		return nil, err
	}
	items := make([]feedback.CoreItem, 0, len(rows))
	for _, row := range rows {
		authorID := uuid.Nil
		if row.AuthorID != nil {
			authorID = *row.AuthorID
		}
		items = append(items, feedback.CoreItem{
			ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID, ContributorID: row.ContributorID,
			AuthorID: authorID, Title: row.Title, Slug: row.Slug, Status: row.Status,
			RoadmapSummary: row.RoadmapSummary, UpdatedAt: row.UpdatedAt,
		})
	}
	return items, nil
}

func (r *Repo) ListItemCandidates(ctx context.Context, workspaceID, portalID, excludedItemID uuid.UUID, search string, limit int) (feedback.CoreMergeCandidatesPage, error) {
	type mergeCandidateRow struct {
		ID           uuid.UUID `db:"id"`
		Slug         string    `db:"slug"`
		Title        string    `db:"title"`
		Status       string    `db:"status"`
		VoteCount    int       `db:"vote_count"`
		CommentCount int       `db:"comment_count"`
	}
	var rows []mergeCandidateRow
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT fi.id, fi.slug, fi.title, `+projectedFeedbackStatus+` AS status,
			CAST(COALESCE(votes.vote_count, 0) AS int) AS vote_count,
			CAST(COALESCE(comments.comment_count, 0) AS int) AS comment_count
		FROM feedback_items fi
		LEFT JOIN LATERAL (
			SELECT story_link.story_id
			FROM feedback_story_links story_link
			WHERE story_link.item_id = fi.id AND story_link.is_primary = true
			LIMIT 1
		) primary_link ON true
		LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
		LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(vote.direction), 0) AS vote_count
			FROM feedback_votes vote WHERE vote.item_id = fi.id
		) votes ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS comment_count
			FROM feedback_comments comment_record WHERE comment_record.item_id = fi.id
		) comments ON true
		WHERE fi.workspace_id = $1 AND fi.portal_id = $2
			AND ($3 = '00000000-0000-0000-0000-000000000000' OR fi.id <> $3)
			AND fi.deleted_at IS NULL AND fi.merged_into_item_id IS NULL
			AND ($4 = '' OR `+feedbackItemSearchVector+` @@ websearch_to_tsquery('english', $4))
		ORDER BY vote_count DESC, fi.created_at DESC, fi.id DESC
		LIMIT $5
	`, workspaceID, portalID, excludedItemID, strings.TrimSpace(search), limit+1); err != nil {
		return feedback.CoreMergeCandidatesPage{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	candidates := make([]feedback.CoreMergeCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, feedback.CoreMergeCandidate{
			ID: row.ID, Slug: row.Slug, Title: row.Title, Status: row.Status,
			VoteCount: row.VoteCount, CommentCount: row.CommentCount,
		})
	}
	return feedback.CoreMergeCandidatesPage{Candidates: candidates, HasMore: hasMore}, nil
}

func (r *Repo) MergeItems(ctx context.Context, input feedback.CoreMergeItemInput) (feedback.CoreMergeItemResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	defer tx.Rollback()

	orderedIDs := []uuid.UUID{input.SourceItemID, input.TargetItemID}
	if orderedIDs[1].String() < orderedIDs[0].String() {
		orderedIDs[0], orderedIDs[1] = orderedIDs[1], orderedIDs[0]
	}
	type lockedItem struct {
		ID               uuid.UUID  `db:"id"`
		WorkspaceID      uuid.UUID  `db:"workspace_id"`
		PortalID         uuid.UUID  `db:"portal_id"`
		Title            string     `db:"title"`
		Slug             string     `db:"slug"`
		MergedIntoItemID *uuid.UUID `db:"merged_into_item_id"`
		DeletedAt        *time.Time `db:"deleted_at"`
	}
	var locked []lockedItem
	if err := tx.SelectContext(ctx, &locked, `
		SELECT id, workspace_id, portal_id, title, slug, merged_into_item_id, deleted_at
		FROM feedback_items
		WHERE workspace_id = $1 AND id = ANY($2)
		ORDER BY id
		FOR UPDATE
	`, input.WorkspaceID, pq.Array(orderedIDs)); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if len(locked) != 2 {
		return feedback.CoreMergeItemResult{}, feedback.ErrNotFound
	}
	items := make(map[uuid.UUID]lockedItem, 2)
	for _, item := range locked {
		items[item.ID] = item
	}
	source, sourceFound := items[input.SourceItemID]
	target, targetFound := items[input.TargetItemID]
	if !sourceFound || !targetFound {
		return feedback.CoreMergeItemResult{}, feedback.ErrNotFound
	}

	// Retrying the exact same completed merge is a read-only idempotent path.
	if source.MergedIntoItemID != nil && *source.MergedIntoItemID == target.ID {
		var stored struct {
			SourceItemID         uuid.UUID `db:"source_item_id"`
			TargetItemID         uuid.UUID `db:"target_item_id"`
			PortalID             uuid.UUID `db:"portal_id"`
			MergedAt             time.Time `db:"merged_at"`
			MergedByUserID       uuid.UUID `db:"merged_by_user_id"`
			MovedFollowerCount   int       `db:"moved_follower_count"`
			MovedUpdateLinkCount int       `db:"moved_update_link_count"`
			MovedStoryLinkCount  int       `db:"moved_story_link_count"`
		}
		if err := tx.GetContext(ctx, &stored, `
			SELECT source_item_id, target_item_id, portal_id, merged_at, merged_by_user_id,
				CAST(COALESCE(CAST(event_payload->>'movedFollowerCount' AS int), 0) AS int) AS moved_follower_count,
				CAST(COALESCE(CAST(event_payload->>'movedUpdateLinkCount' AS int), 0) AS int) AS moved_update_link_count,
				CAST(COALESCE(CAST(event_payload->>'movedStoryLinkCount' AS int), 0) AS int) AS moved_story_link_count
			FROM feedback_item_merge_outbox WHERE source_item_id = $1 AND target_item_id = $2
		`, source.ID, target.ID); err != nil {
			return feedback.CoreMergeItemResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return feedback.CoreMergeItemResult{}, err
		}
		result := feedback.CoreMergeItemResult{
			SourceItemID: stored.SourceItemID, TargetItemID: stored.TargetItemID, PortalID: stored.PortalID,
			MergedAt: stored.MergedAt, MergedByUserID: stored.MergedByUserID,
			MovedFollowerCount: stored.MovedFollowerCount, MovedUpdateLinkCount: stored.MovedUpdateLinkCount,
			MovedStoryLinkCount: stored.MovedStoryLinkCount,
		}
		result.Target, err = r.GetItem(ctx, input.WorkspaceID, target.ID)
		return result, err
	}
	if source.ID == target.ID || source.PortalID != target.PortalID || source.MergedIntoItemID != nil || target.MergedIntoItemID != nil || source.DeletedAt != nil || target.DeletedAt != nil {
		return feedback.CoreMergeItemResult{}, feedback.ErrMergeConflict
	}
	var hasInboundMerges bool
	if err := tx.GetContext(ctx, &hasInboundMerges, `
		SELECT EXISTS (
			SELECT 1 FROM feedback_items
			WHERE workspace_id = $1 AND portal_id = $2 AND merged_into_item_id = $3
		)
	`, source.WorkspaceID, source.PortalID, source.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if hasInboundMerges {
		// V1 canonical resolution follows one hop. Reject merging an existing
		// canonical target again instead of creating an A -> B -> C chain.
		return feedback.CoreMergeItemResult{}, feedback.ErrMergeConflict
	}

	var sourceFollowerIDs []uuid.UUID
	if err := tx.SelectContext(ctx, &sourceFollowerIDs, `
		SELECT contributor_id FROM feedback_item_followers
		WHERE item_id = $1 AND unsubscribed_at IS NULL
		ORDER BY contributor_id
	`, source.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	var followerCount int
	if err := tx.GetContext(ctx, &followerCount, `
		SELECT CAST(COUNT(*) AS int)
		FROM feedback_item_followers source
		LEFT JOIN feedback_item_followers target
			ON target.item_id = $2 AND target.contributor_id = source.contributor_id
		WHERE source.item_id = $1 AND source.unsubscribed_at IS NULL
			AND (target.contributor_id IS NULL OR target.unsubscribed_at IS NOT NULL)
	`, source.ID, target.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_item_followers (item_id, contributor_id, created_at, unsubscribed_at)
		SELECT $2, contributor_id, created_at, NULL
		FROM feedback_item_followers WHERE item_id = $1 AND unsubscribed_at IS NULL
		ON CONFLICT (item_id, contributor_id) DO UPDATE
		SET unsubscribed_at = NULL,
			created_at = LEAST(feedback_item_followers.created_at, EXCLUDED.created_at)
	`, source.ID, target.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}

	var updateLinkCount int
	if err := tx.GetContext(ctx, &updateLinkCount, `
		WITH moved AS (
			INSERT INTO feedback_update_items (update_id, item_id)
			SELECT update_id, $2 FROM feedback_update_items WHERE item_id = $1
			ON CONFLICT (update_id, item_id) DO NOTHING
			RETURNING update_id
		)
		SELECT CAST(COUNT(*) AS int) FROM moved
	`, source.ID, target.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}

	var storyLinkCount int
	var sourcePrimaryStoryID uuid.UUID
	sourcePrimaryExists := true
	if err := tx.GetContext(ctx, &sourcePrimaryStoryID, `
		SELECT story_id FROM feedback_story_links WHERE item_id = $1 AND is_primary = true
	`, source.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return feedback.CoreMergeItemResult{}, err
		}
		sourcePrimaryExists = false
	}
	var targetPrimaryStoryID uuid.UUID
	targetPrimaryExists := true
	if err := tx.GetContext(ctx, &targetPrimaryStoryID, `
		SELECT story_id FROM feedback_story_links WHERE item_id = $1 AND is_primary = true
	`, target.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return feedback.CoreMergeItemResult{}, err
		}
		targetPrimaryExists = false
	}
	if sourcePrimaryExists && targetPrimaryExists && sourcePrimaryStoryID != targetPrimaryStoryID {
		return feedback.CoreMergeItemResult{}, feedback.ErrMergeConflict
	}
	if sourcePrimaryExists && !targetPrimaryExists {
		var targetAlreadyLinksStory bool
		if err := tx.GetContext(ctx, &targetAlreadyLinksStory, `
			SELECT EXISTS (SELECT 1 FROM feedback_story_links WHERE item_id = $1 AND story_id = $2)
		`, target.ID, sourcePrimaryStoryID); err != nil {
			return feedback.CoreMergeItemResult{}, err
		}
		if targetAlreadyLinksStory {
			return feedback.CoreMergeItemResult{}, feedback.ErrMergeConflict
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE feedback_story_links SET item_id = $2 WHERE item_id = $1 AND is_primary = true
		`, source.ID, target.ID)
		if err != nil {
			return feedback.CoreMergeItemResult{}, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return feedback.CoreMergeItemResult{}, err
		}
		storyLinkCount += int(count)
	}
	var nonPrimaryStoryLinkCount int
	if err := tx.GetContext(ctx, &nonPrimaryStoryLinkCount, `
		WITH moved AS (
			INSERT INTO feedback_story_links (workspace_id, item_id, story_id, relationship, is_primary, created_by_user_id, created_at)
			SELECT workspace_id, $2, story_id, relationship, false, created_by_user_id, created_at
			FROM feedback_story_links
			WHERE item_id = $1 AND is_primary = false
			ON CONFLICT (item_id, story_id) DO NOTHING
			RETURNING id
		)
		SELECT CAST(COUNT(*) AS int) FROM moved
	`, source.ID, target.ID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	storyLinkCount += nonPrimaryStoryLinkCount

	var mergedAt time.Time
	if err := tx.GetContext(ctx, &mergedAt, `
		UPDATE feedback_items
		SET merged_into_item_id = $2, merged_at = NOW(), merged_by_user_id = $3, updated_at = NOW()
		WHERE id = $1 AND merged_into_item_id IS NULL
		RETURNING merged_at
	`, source.ID, target.ID, input.ActorID); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE feedback_item_followers
		SET unsubscribed_at = $2
		WHERE item_id = $1 AND unsubscribed_at IS NULL
	`, source.ID, mergedAt); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	eventID := uuid.New()
	payload, err := json.Marshal(map[string]any{
		"schemaVersion": 1, "mergeEventId": eventID,
		"workspaceId": source.WorkspaceID, "portalId": source.PortalID,
		"sourceItemId": source.ID, "targetItemId": target.ID,
		"mergedByUserId": input.ActorID, "mergedAt": mergedAt,
		"sourceTitle": source.Title, "sourceSlug": source.Slug,
		"targetTitle": target.Title, "targetSlug": target.Slug,
		"movedFollowerCount": followerCount, "movedUpdateLinkCount": updateLinkCount,
		"movedStoryLinkCount": storyLinkCount,
		"sourceFollowerIds":   sourceFollowerIDs,
	})
	if err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_item_merge_outbox (
			merge_event_id, source_item_id, target_item_id, workspace_id, portal_id,
			merged_by_user_id, merged_at, event_payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, CAST($8 AS jsonb))
	`, eventID, source.ID, target.ID, source.WorkspaceID, source.PortalID, input.ActorID, mergedAt, payload); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	targetItem, err := r.GetItem(ctx, input.WorkspaceID, target.ID)
	if err != nil {
		return feedback.CoreMergeItemResult{}, err
	}
	return feedback.CoreMergeItemResult{
		SourceItemID: source.ID, TargetItemID: target.ID, PortalID: source.PortalID,
		MergedAt: mergedAt, MergedByUserID: input.ActorID, MovedFollowerCount: followerCount,
		MovedUpdateLinkCount: updateLinkCount, MovedStoryLinkCount: storyLinkCount, Target: targetItem,
	}, nil
}

func (r *Repo) ClaimMergeOutboxEvents(ctx context.Context, limit int, staleAfter time.Duration) ([]feedback.CoreMergeOutboxEvent, error) {
	if limit <= 0 {
		return []feedback.CoreMergeOutboxEvent{}, nil
	}
	type mergeOutboxRow struct {
		EventID      uuid.UUID `db:"event_id"`
		WorkspaceID  uuid.UUID `db:"workspace_id"`
		PortalID     uuid.UUID `db:"portal_id"`
		SourceItemID uuid.UUID `db:"source_item_id"`
		TargetItemID uuid.UUID `db:"target_item_id"`
		ActorID      uuid.UUID `db:"actor_id"`
		MergedAt     time.Time `db:"merged_at"`
		ClaimToken   uuid.UUID `db:"claim_token"`
		AttemptCount int       `db:"attempt_count"`
		Payload      []byte    `db:"event_payload"`
	}
	var rows []mergeOutboxRow
	err := r.db.SelectContext(ctx, &rows, `
		WITH candidates AS (
			SELECT merge_event_id
			FROM feedback_item_merge_outbox
			WHERE (status IN ('pending', 'retrying') AND next_attempt_at <= NOW())
				OR (status = 'processing' AND claimed_at <= NOW() - CAST($2 AS interval))
			ORDER BY COALESCE(next_attempt_at, claimed_at), created_at, merge_event_id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		), claimed AS (
			UPDATE feedback_item_merge_outbox outbox
			SET status = 'processing', attempt_count = attempt_count + 1,
				next_attempt_at = NULL, claim_token = gen_random_uuid(), claimed_at = NOW(),
				completed_at = NULL, last_error = NULL, updated_at = NOW()
			FROM candidates WHERE outbox.merge_event_id = candidates.merge_event_id
			RETURNING outbox.*
		)
		SELECT merge_event_id AS event_id, workspace_id, portal_id, source_item_id, target_item_id,
			merged_by_user_id AS actor_id, merged_at, claim_token, attempt_count, event_payload
		FROM claimed ORDER BY merged_at, merge_event_id
	`, limit, intervalLiteral(staleAfter))
	if err != nil {
		return nil, err
	}
	result := make([]feedback.CoreMergeOutboxEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreMergeOutboxEvent{
			EventID: row.EventID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID,
			SourceItemID: row.SourceItemID, TargetItemID: row.TargetItemID, ActorID: row.ActorID,
			MergedAt: row.MergedAt, ClaimToken: row.ClaimToken, AttemptCount: row.AttemptCount,
			Payload: append(json.RawMessage(nil), row.Payload...),
		})
	}
	return result, nil
}

func (r *Repo) ListMergeRecipients(ctx context.Context, portalID, targetItemID uuid.UUID, contributorIDs []uuid.UUID) ([]feedback.CoreMergeRecipient, error) {
	if len(contributorIDs) == 0 {
		return []feedback.CoreMergeRecipient{}, nil
	}
	var rows []struct {
		ContributorID uuid.UUID  `db:"contributor_id"`
		UserID        *uuid.UUID `db:"user_id"`
		Kind          string     `db:"kind"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT contributor.id AS contributor_id, contributor.user_id, contributor.kind
		FROM feedback_contributors contributor
		INNER JOIN feedback_item_followers target_follower
			ON target_follower.item_id = $2 AND target_follower.contributor_id = contributor.id
		LEFT JOIN users account ON account.user_id = contributor.user_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
		WHERE contributor.portal_id = $1 AND contributor.id = ANY($3)
			AND target_follower.unsubscribed_at IS NULL
			AND contributor.blocked_at IS NULL
			AND (
				(contributor.kind = 'account' AND contributor.user_id IS NOT NULL AND account.is_active = true)
				OR (
					contributor.kind IN ('verified_guest', 'external')
					AND contributor.email IS NOT NULL
					AND preference.email_unsubscribed_at IS NULL
				)
			)
		ORDER BY contributor.id
	`, portalID, targetItemID, pq.Array(contributorIDs)); err != nil {
		return nil, err
	}
	result := make([]feedback.CoreMergeRecipient, 0, len(rows))
	for _, row := range rows {
		userID := uuid.Nil
		if row.UserID != nil {
			userID = *row.UserID
		}
		result = append(result, feedback.CoreMergeRecipient{
			ContributorID: row.ContributorID, UserID: userID, Kind: row.Kind,
		})
	}
	return result, nil
}

func (r *Repo) CompleteMergeOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE feedback_item_merge_outbox
		SET status = 'completed', next_attempt_at = NULL, claim_token = NULL, claimed_at = NULL,
			completed_at = NOW(), last_error = NULL, updated_at = NOW()
		WHERE merge_event_id = $1 AND claim_token = $2 AND status = 'processing'
	`, eventID, claimToken)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *Repo) RetryMergeOutboxEvent(ctx context.Context, eventID, claimToken uuid.UUID, failure string, retryAt time.Time, terminal bool) error {
	status := "retrying"
	var nextAttempt *time.Time
	if terminal {
		status = "failed"
	} else {
		nextAttempt = &retryAt
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE feedback_item_merge_outbox
		SET status = $3, next_attempt_at = $4, claim_token = NULL, claimed_at = NULL,
			completed_at = NULL, last_error = LEFT($5, 4000), updated_at = NOW()
		WHERE merge_event_id = $1 AND claim_token = $2 AND status = 'processing'
	`, eventID, claimToken, status, nextAttempt, strings.TrimSpace(failure))
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func intervalLiteral(duration time.Duration) string {
	seconds := int64(duration.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%d seconds", seconds)
}

func (r *Repo) CreateContributorDelivery(ctx context.Context, input feedback.CoreCreateDeliveryInput) (feedback.CoreDelivery, bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreDelivery{}, false, err
	}
	defer tx.Rollback()
	payload, err := json.Marshal(map[string]any{
		"itemId": input.ItemID, "updateId": input.UpdateID, "eventType": input.EventType,
	})
	if err != nil {
		return feedback.CoreDelivery{}, false, err
	}
	var row deliveryRow
	err = tx.GetContext(ctx, &row, `
		WITH inserted AS (
			INSERT INTO feedback_contributor_deliveries (
				id, portal_id, contributor_id, item_id, update_id, event_type, dedupe_key,
				subject, message, destination_url, recipient_email, event_payload
			)
			SELECT $1, $2, contributor.id,
				CASE WHEN EXISTS (
					SELECT 1 FROM feedback_items item WHERE item.id = $4 AND item.portal_id = $2
				) THEN $4 ELSE NULL END,
				CASE WHEN EXISTS (
					SELECT 1 FROM feedback_updates update_record WHERE update_record.id = $5 AND update_record.portal_id = $2
				) THEN $5 ELSE NULL END,
				$6, $7, $8, $9, $10, contributor.email, CAST($11 AS jsonb)
			FROM feedback_contributors contributor
			LEFT JOIN feedback_contributor_preferences preference
				ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
			WHERE contributor.portal_id = $2 AND contributor.id = $3
				AND contributor.kind IN ('verified_guest', 'external')
				AND contributor.email IS NOT NULL AND contributor.blocked_at IS NULL
				AND preference.email_unsubscribed_at IS NULL
			ON CONFLICT (portal_id, contributor_id, channel, dedupe_key) DO NOTHING
			RETURNING *
		), selected AS (
			SELECT inserted.*, true AS was_created FROM inserted
			UNION ALL
			SELECT existing.*, false AS was_created
			FROM feedback_contributor_deliveries existing
			WHERE existing.portal_id = $2 AND existing.contributor_id = $3
				AND existing.channel = 'email' AND existing.dedupe_key = $7
				AND NOT EXISTS (SELECT 1 FROM inserted)
			LIMIT 1
		)
		SELECT selected.id, selected.portal_id, selected.contributor_id, selected.recipient_email,
			COALESCE(NULLIF(trim(contributor.display_name), ''), 'there') AS display_name,
			workspace.name AS portal_name, workspace.slug AS portal_slug,
			selected.item_id, selected.update_id, selected.event_type, selected.dedupe_key,
			selected.subject, selected.message, selected.destination_url, selected.status,
			selected.attempt_count, selected.final_failure_reason, selected.created_at,
			selected.was_created
		FROM selected
		INNER JOIN feedback_contributors contributor ON contributor.id = selected.contributor_id
		LEFT JOIN feedback_contributor_preferences preference
			ON preference.portal_id = contributor.portal_id AND preference.contributor_id = contributor.id
		INNER JOIN feedback_portals portal ON portal.id = selected.portal_id
		INNER JOIN workspaces workspace ON workspace.workspace_id = portal.workspace_id
		WHERE contributor.kind IN ('verified_guest', 'external')
			AND contributor.email IS NOT NULL AND contributor.blocked_at IS NULL
			AND preference.email_unsubscribed_at IS NULL
	`, input.DeliveryID, input.PortalID, input.ContributorID, input.ItemID, input.UpdateID, input.EventType, input.DedupeKey,
		input.Subject, input.Message, input.DestinationURL, payload)
	if errors.Is(err, sql.ErrNoRows) {
		return feedback.CoreDelivery{}, false, nil
	}
	if err != nil {
		return feedback.CoreDelivery{}, false, err
	}
	if row.WasCreated {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO feedback_contributor_unsubscribe_tokens (
				portal_id, contributor_id, delivery_id, purpose, token_hash, expires_at
			) VALUES ($1, $2, $3, 'all_email', $4, NOW() + INTERVAL '30 days')
		`, input.PortalID, input.ContributorID, row.ID, input.TokenHash); err != nil {
			return feedback.CoreDelivery{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreDelivery{}, false, err
	}
	return toCoreDelivery(row), row.WasCreated, nil
}

func toCoreDelivery(row deliveryRow) feedback.CoreDelivery {
	return feedback.CoreDelivery{
		ID: row.ID, PortalID: row.PortalID, ContributorID: row.ContributorID,
		Email: row.Email, DisplayName: row.DisplayName, PortalName: row.PortalName, PortalSlug: row.PortalSlug,
		ItemID: row.ItemID, UpdateID: row.UpdateID, EventType: row.EventType, DedupeKey: row.DedupeKey,
		Subject: row.Subject, Message: row.Message, DestinationURL: row.DestinationURL,
		Status: row.Status, AttemptCount: row.AttemptCount,
		FinalFailureReason: pointerString(row.FinalFailureReason), CreatedAt: row.CreatedAt,
	}
}

func (r *Repo) ListWorkspaceUpdates(ctx context.Context, workspaceID uuid.UUID, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	return r.listUpdates(ctx, `fu.workspace_id = $1`, []any{workspaceID}, page, pageSize)
}

func (r *Repo) GetWorkspaceUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	return r.getUpdate(ctx, `fu.workspace_id = $1 AND fu.id = $2`, workspaceID, updateID)
}

func (r *Repo) CreateUpdate(ctx context.Context, input feedback.CoreUpdateInput) (feedback.CoreFeedbackUpdate, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	defer tx.Rollback()
	var updateID uuid.UUID
	err = tx.GetContext(ctx, &updateID, `
		INSERT INTO feedback_updates (
			workspace_id, portal_id, author_id, title, body, status, slug, summary, cover_image_url
		)
		SELECT portal.workspace_id, portal.id, $3, $4, $5, 'draft', $6, $7, $8
		FROM feedback_portals portal
		WHERE portal.workspace_id = $1 AND portal.id = $2
		RETURNING id
	`, input.WorkspaceID, input.PortalID, input.ActorID, input.Title, input.Body, input.Slug, input.Summary, input.CoverImageURL)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, normalizeUpdateWriteError(err)
	}
	if err := replaceUpdateItems(ctx, tx, input.WorkspaceID, input.PortalID, updateID, input.ItemIDs); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return r.GetWorkspaceUpdate(ctx, input.WorkspaceID, updateID)
}

func (r *Repo) UpdateUpdate(ctx context.Context, input feedback.CoreUpdateInput) (feedback.CoreFeedbackUpdate, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		UPDATE feedback_updates
		SET portal_id = $3, title = $4, body = $5, slug = $6, summary = $7,
			cover_image_url = $8, updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2 AND status = 'draft'
	`, input.WorkspaceID, input.UpdateID, input.PortalID, input.Title, input.Body, input.Slug, input.Summary, input.CoverImageURL)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, normalizeUpdateWriteError(err)
	}
	if err := requireAffected(result); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	if err := replaceUpdateItems(ctx, tx, input.WorkspaceID, input.PortalID, input.UpdateID, input.ItemIDs); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return r.GetWorkspaceUpdate(ctx, input.WorkspaceID, input.UpdateID)
}

func (r *Repo) DeleteUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM feedback_updates WHERE workspace_id = $1 AND id = $2 AND status = 'draft'
	`, workspaceID, updateID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *Repo) PublishUpdate(ctx context.Context, workspaceID, updateID, actorID uuid.UUID) (feedback.CoreFeedbackUpdate, bool, error) {
	return r.publishUpdateWithOutbox(ctx, workspaceID, updateID, actorID)
}

func (r *Repo) UnpublishUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (feedback.CoreFeedbackUpdate, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE feedback_updates
		SET status = 'draft', published_at = NULL, published_by_user_id = NULL, updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2 AND status = 'published'
	`, workspaceID, updateID)
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	if err := requireAffected(result); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return r.GetWorkspaceUpdate(ctx, workspaceID, updateID)
}

func (r *Repo) ListPublicUpdates(ctx context.Context, portalID uuid.UUID, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	return r.listUpdates(ctx, `fu.portal_id = $1 AND fu.status = 'published' AND fu.published_at IS NOT NULL`, []any{portalID}, page, pageSize)
}

func (r *Repo) GetPublicUpdate(ctx context.Context, portalID uuid.UUID, slug string) (feedback.CoreFeedbackUpdate, error) {
	return r.getUpdate(ctx, `fu.portal_id = $1 AND fu.slug = $2 AND fu.status = 'published' AND fu.published_at IS NOT NULL`, portalID, slug)
}

func (r *Repo) listUpdates(ctx context.Context, where string, args []any, page, pageSize int) (feedback.CoreUpdatesPage, error) {
	limit := pageSize + 1
	offset := (page - 1) * pageSize
	query := `
		SELECT fu.id, fu.workspace_id, fu.portal_id, fu.slug, fu.title,
			fu.summary, fu.body, fu.cover_image_url, fu.status, fu.published_at,
			fu.published_by_user_id, fu.created_at, fu.updated_at
		FROM feedback_updates fu
		WHERE ` + where + `
		ORDER BY COALESCE(fu.published_at, fu.updated_at) DESC, fu.id DESC
		LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)
	var rows []feedbackUpdateRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return feedback.CoreUpdatesPage{}, err
	}
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

func (r *Repo) getUpdate(ctx context.Context, where string, args ...any) (feedback.CoreFeedbackUpdate, error) {
	var row feedbackUpdateRow
	query := `
		SELECT fu.id, fu.workspace_id, fu.portal_id, fu.slug, fu.title,
			fu.summary, fu.body, fu.cover_image_url, fu.status, fu.published_at,
			fu.published_by_user_id, fu.created_at, fu.updated_at
		FROM feedback_updates fu
		WHERE ` + where + `
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	updates, err := r.hydrateUpdates(ctx, []feedbackUpdateRow{row})
	if err != nil {
		return feedback.CoreFeedbackUpdate{}, err
	}
	return updates[0], nil
}

func (r *Repo) hydrateUpdates(ctx context.Context, rows []feedbackUpdateRow) ([]feedback.CoreFeedbackUpdate, error) {
	if len(rows) == 0 {
		return []feedback.CoreFeedbackUpdate{}, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	var itemRows []updateItemRow
	if err := r.db.SelectContext(ctx, &itemRows, `
		SELECT link.update_id, item.id, item.slug, item.title, `+projectedFeedbackStatus+` AS status
		FROM feedback_update_items link
		INNER JOIN feedback_items item ON item.id = link.item_id
		LEFT JOIN LATERAL (
			SELECT story_link.id, story_link.story_id
			FROM feedback_story_links story_link
			WHERE story_link.item_id = item.id AND story_link.is_primary = true
			LIMIT 1
		) primary_link ON true
		LEFT JOIN stories projected_story ON projected_story.id = primary_link.story_id
		LEFT JOIN statuses projected_state ON projected_state.status_id = projected_story.status_id
		WHERE link.update_id = ANY($1)
			AND item.deleted_at IS NULL
			AND item.merged_into_item_id IS NULL
		ORDER BY item.title ASC, item.id ASC
	`, pq.Array(ids)); err != nil {
		return nil, err
	}
	itemsByUpdate := make(map[uuid.UUID][]feedback.CoreUpdateItem, len(rows))
	for _, row := range itemRows {
		itemsByUpdate[row.UpdateID] = append(itemsByUpdate[row.UpdateID], feedback.CoreUpdateItem{ID: row.ID, Slug: row.Slug, Title: row.Title, Status: row.Status})
	}
	result := make([]feedback.CoreFeedbackUpdate, 0, len(rows))
	for _, row := range rows {
		result = append(result, feedback.CoreFeedbackUpdate{
			ID: row.ID, WorkspaceID: row.WorkspaceID, PortalID: row.PortalID, Slug: row.Slug,
			Title: row.Title, Summary: row.Summary, Body: row.Body, CoverImageURL: row.CoverImageURL,
			Status: row.Status, PublishedAt: row.PublishedAt, PublishedByUserID: row.PublishedByUserID,
			LinkedItems: itemsByUpdate[row.ID], CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}
	return result, nil
}

func replaceUpdateItems(ctx context.Context, tx *sqlx.Tx, workspaceID, portalID, updateID uuid.UUID, itemIDs []uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM feedback_update_items WHERE update_id = $1`, updateID); err != nil {
		return err
	}
	if len(itemIDs) == 0 {
		return nil
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_update_items (update_id, item_id)
		SELECT $1, item.id
		FROM feedback_items item
		WHERE item.workspace_id = $2
			AND item.portal_id = $3
			AND item.id = ANY($4)
			AND item.deleted_at IS NULL
			AND item.merged_into_item_id IS NULL
	`, updateID, workspaceID, portalID, pq.Array(itemIDs))
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != int64(len(itemIDs)) {
		return feedback.ErrNotFound
	}
	return nil
}

func normalizeUpdateWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "feedback_updates_portal_slug_unique" {
		return fmt.Errorf("%w: update slug already exists", feedback.ErrInvalidInput)
	}
	return err
}

func (r *Repo) GetWidgetSettings(ctx context.Context, workspaceID, portalID uuid.UUID) (feedback.CoreWidgetSettings, error) {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback_widget_settings (portal_id, enabled, allowed_origins)
		SELECT id, false, '{}' FROM feedback_portals WHERE workspace_id = $1 AND id = $2
		ON CONFLICT (portal_id) DO NOTHING
	`, workspaceID, portalID); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	var row widgetSettingsRow
	err := r.db.GetContext(ctx, &row, widgetSettingsSelect()+`
		WHERE portal.workspace_id = $1 AND settings.portal_id = $2
	`, workspaceID, portalID)
	return toCoreWidgetSettings(row), err
}

func (r *Repo) GetPublicWidgetSettings(ctx context.Context, portalID uuid.UUID) (feedback.CoreWidgetSettings, error) {
	var row widgetSettingsRow
	err := r.db.GetContext(ctx, &row, widgetSettingsSelect()+`
		WHERE settings.portal_id = $1
	`, portalID)
	return toCoreWidgetSettings(row), err
}

const upsertWidgetSettingsQuery = `
	WITH target AS (
		SELECT portal.id, settings.widget_key_id, settings.signing_secret_encrypted,
			settings.signing_secret_version
		FROM feedback_portals portal
		LEFT JOIN feedback_widget_settings settings ON settings.portal_id = portal.id
		WHERE portal.workspace_id = $1 AND portal.id = $2
	), updated AS (
		INSERT INTO feedback_widget_settings (
			portal_id, enabled, widget_key_id, allowed_origins,
			signing_secret_encrypted, signing_secret_version
		)
		SELECT id, $3, COALESCE(widget_key_id, gen_random_uuid()), $4,
			signing_secret_encrypted, COALESCE(signing_secret_version, 0)
		FROM target
		ON CONFLICT (portal_id) DO UPDATE
		SET enabled = EXCLUDED.enabled,
			allowed_origins = EXCLUDED.allowed_origins,
			updated_at = NOW()
		RETURNING *
	)
	SELECT updated.portal_id, updated.enabled, updated.widget_key_id, updated.allowed_origins,
		updated.signing_secret_encrypted, updated.signing_secret_version,
		(SELECT MAX(grace_expires_at) FROM feedback_widget_signing_secret_rotations rotation
			WHERE rotation.portal_id = updated.portal_id AND rotation.retired_at IS NULL AND rotation.grace_expires_at > NOW()) AS previous_version_expires_at,
		updated.created_at, updated.updated_at
	FROM updated
`

func (r *Repo) UpsertWidgetSettings(ctx context.Context, input feedback.CoreWidgetSettingsInput) (feedback.CoreWidgetSettings, error) {
	var row widgetSettingsRow
	err := r.db.GetContext(ctx, &row, upsertWidgetSettingsQuery, input.WorkspaceID, input.PortalID, input.Enabled, pq.Array(input.AllowedOrigins))
	return toCoreWidgetSettings(row), err
}

func (r *Repo) SetInitialWidgetSecret(ctx context.Context, workspaceID, portalID, keyID uuid.UUID, encrypted string, version int) (feedback.CoreWidgetSettings, error) {
	var row widgetSettingsRow
	err := r.db.GetContext(ctx, &row, `
		WITH target AS (
			SELECT id FROM feedback_portals WHERE workspace_id = $1 AND id = $2
		), updated AS (
			INSERT INTO feedback_widget_settings (
				portal_id, enabled, widget_key_id, allowed_origins, signing_secret_encrypted, signing_secret_version
			)
			SELECT id, false, $3, '{}', $4, $5 FROM target
			ON CONFLICT (portal_id)
			DO UPDATE SET widget_key_id = EXCLUDED.widget_key_id,
				signing_secret_encrypted = EXCLUDED.signing_secret_encrypted,
				signing_secret_version = EXCLUDED.signing_secret_version,
				updated_at = NOW()
			WHERE feedback_widget_settings.signing_secret_encrypted IS NULL
			RETURNING *
		)
		SELECT updated.portal_id, updated.enabled, updated.widget_key_id, updated.allowed_origins,
			updated.signing_secret_encrypted, updated.signing_secret_version,
			CAST(NULL AS timestamptz) AS previous_version_expires_at, updated.created_at, updated.updated_at
		FROM updated
	`, workspaceID, portalID, keyID, encrypted, version)
	if errors.Is(err, sql.ErrNoRows) {
		return feedback.CoreWidgetSettings{}, fmt.Errorf("%w: widget signing secret already exists", feedback.ErrInvalidInput)
	}
	return toCoreWidgetSettings(row), err
}

func (r *Repo) RotateWidgetSecret(ctx context.Context, workspaceID, portalID uuid.UUID, encrypted string, version int, graceExpiresAt time.Time) (feedback.CoreWidgetSettings, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	defer tx.Rollback()
	var current widgetSettingsRow
	if err := tx.GetContext(ctx, &current, `
		SELECT settings.portal_id, settings.enabled, settings.widget_key_id, settings.allowed_origins,
			settings.signing_secret_encrypted, settings.signing_secret_version,
			CAST(NULL AS timestamptz) AS previous_version_expires_at, settings.created_at, settings.updated_at
		FROM feedback_widget_settings settings
		INNER JOIN feedback_portals portal ON portal.id = settings.portal_id
		WHERE portal.workspace_id = $1 AND settings.portal_id = $2
		FOR UPDATE
	`, workspaceID, portalID); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	if current.SigningSecretEncrypted == nil || current.SigningSecretVersion <= 0 || version != current.SigningSecretVersion+1 {
		return feedback.CoreWidgetSettings{}, feedback.ErrInvalidInput
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_widget_signing_secret_rotations (
			portal_id, signing_secret_version, signing_secret_encrypted, activated_at, grace_expires_at
		) VALUES ($1, $2, $3, $4, $5)
	`, portalID, current.SigningSecretVersion, *current.SigningSecretEncrypted, current.UpdatedAt, graceExpiresAt); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE feedback_widget_settings
		SET signing_secret_encrypted = $2, signing_secret_version = $3, updated_at = NOW()
		WHERE portal_id = $1
	`, portalID, encrypted, version); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	return r.GetWidgetSettings(ctx, workspaceID, portalID)
}

func (r *Repo) GetWidgetSigningSecret(ctx context.Context, portalID, widgetKeyID uuid.UUID, version int) (string, error) {
	var encrypted string
	err := r.db.GetContext(ctx, &encrypted, `
		SELECT secret
		FROM (
			SELECT settings.signing_secret_encrypted AS secret, settings.signing_secret_version AS version
			FROM feedback_widget_settings settings
			WHERE settings.portal_id = $1 AND settings.widget_key_id = $2 AND settings.enabled = true
			UNION ALL
			SELECT rotation.signing_secret_encrypted AS secret, rotation.signing_secret_version AS version
			FROM feedback_widget_signing_secret_rotations rotation
			INNER JOIN feedback_widget_settings settings ON settings.portal_id = rotation.portal_id
			WHERE rotation.portal_id = $1 AND settings.widget_key_id = $2 AND settings.enabled = true
				AND rotation.retired_at IS NULL AND rotation.grace_expires_at > NOW()
		) versions
		WHERE version = $3 AND secret IS NOT NULL
		LIMIT 1
	`, portalID, widgetKeyID, version)
	return encrypted, err
}

func (r *Repo) ConsumeWidgetAssertionNonce(ctx context.Context, portalID, widgetKeyID uuid.UUID, version int, nonce, parentOrigin string, expiresAt time.Time) error {
	nonceHash := sha256.Sum256([]byte(nonce))
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback_widget_assertion_nonces (
			portal_id, widget_key_id, signing_secret_version, nonce_hash, parent_origin, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, portalID, widgetKeyID, version, nonceHash[:], parentOrigin, expiresAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return feedback.ErrWidgetAssertionReplayed
	}
	return err
}

func (r *Repo) CreateExternalContributorSession(ctx context.Context, portalID uuid.UUID, externalID, email, displayName string, avatarURL *string, tokenHash []byte, expiresAt time.Time) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	var row participantRow
	err = tx.GetContext(ctx, &row, `
		WITH existing AS (
			SELECT id
			FROM feedback_contributors
			WHERE portal_id = $1 AND external_id = $2 AND kind = 'external'
			LIMIT 1
			FOR UPDATE
		), updated AS (
			UPDATE feedback_contributors contributor
			SET kind = 'external', user_id = NULL, external_id = $2, email = $3,
				email_verified_at = $6, display_name = $4, avatar_url = $5,
				public_masked = false, last_seen_at = $6, updated_at = $6
			FROM existing WHERE contributor.id = existing.id
			RETURNING contributor.*
		), inserted AS (
			INSERT INTO feedback_contributors (
				portal_id, kind, external_id, email, email_verified_at, display_name, avatar_url, last_seen_at
			)
			SELECT $1, 'external', $2, $3, $6, $4, $5, $6
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING *
		)
		SELECT id, portal_id, user_id, kind, email, email_verified_at, display_name, avatar_url,
			external_id, public_masked, blocked_at, blocked_reason, last_seen_at, created_at, updated_at
		FROM updated
		UNION ALL
		SELECT id, portal_id, user_id, kind, email, email_verified_at, display_name, avatar_url,
			external_id, public_masked, blocked_at, blocked_reason, last_seen_at, created_at, updated_at
		FROM inserted
	`, portalID, externalID, email, displayName, avatarURL, now)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	participant := toCoreParticipant(row)
	if participant.BlockedAt != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, feedback.ErrContributorBlocked
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO feedback_contributor_preferences (portal_id, contributor_id)
		VALUES ($1, $2) ON CONFLICT (portal_id, contributor_id) DO NOTHING
	`, portalID, participant.ID); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	session, err := insertContributorSession(ctx, tx, portalID, participant.ID, tokenHash, feedback.ContributorSessionSourceWidget, expiresAt)
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, err
	}
	return participant, session, nil
}

func widgetSettingsSelect() string {
	return `
		SELECT settings.portal_id, settings.enabled, settings.widget_key_id, settings.allowed_origins,
			settings.signing_secret_encrypted, settings.signing_secret_version,
			(SELECT MAX(grace_expires_at) FROM feedback_widget_signing_secret_rotations rotation
				WHERE rotation.portal_id = settings.portal_id AND rotation.retired_at IS NULL AND rotation.grace_expires_at > NOW()) AS previous_version_expires_at,
			settings.created_at, settings.updated_at
		FROM feedback_widget_settings settings
		INNER JOIN feedback_portals portal ON portal.id = settings.portal_id
	`
}

func toCoreWidgetSettings(row widgetSettingsRow) feedback.CoreWidgetSettings {
	settings := feedback.CoreWidgetSettings{
		PortalID: row.PortalID, Enabled: row.Enabled, WidgetKeyID: row.WidgetKeyID,
		AllowedOrigins: []string(row.AllowedOrigins), SigningSecretVersion: row.SigningSecretVersion,
		PreviousVersionExpiresAt: row.PreviousExpiresAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.SigningSecretEncrypted != nil {
		settings.SigningSecretEncrypted = *row.SigningSecretEncrypted
	}
	return settings
}
