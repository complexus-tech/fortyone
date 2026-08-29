package feedbackrepository

import (
	"context"
	"crypto/hmac"
	"errors"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxVerificationCodeAttempts = 5
	verificationRateLimit       = 3
	verificationRateWindow      = 10 * time.Minute
)

func (r *Repo) CreateContributorVerification(ctx context.Context, input feedback.CoreVerificationRequest) (feedback.CoreVerificationChallenge, error) {
	now := time.Now().UTC()
	row, err := r.queries.CreateContributorVerification(ctx, feedbacksql.CreateContributorVerificationParams{
		Email: input.Email, DisplayName: input.DisplayName, PublicMasked: input.PublicMasked,
		TokenHash: input.TokenHash, CodeHash: input.CodeHash, Source: input.Source, ExpiresAt: input.ExpiresAt,
		PortalID: input.PortalID, RateLimitSince: now.Add(-verificationRateWindow), RateLimit: verificationRateLimit,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return feedback.CoreVerificationChallenge{}, feedback.ErrVerificationAttempts
	}
	if err != nil {
		return feedback.CoreVerificationChallenge{}, normalizeError(err)
	}
	return feedback.CoreVerificationChallenge{ID: row.ID, ExpiresAt: row.ExpiresAt}, nil
}

func (r *Repo) ConfirmContributorVerification(ctx context.Context, input feedback.CoreVerificationConfirmation) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	var participant feedback.CoreParticipant
	var session feedback.CoreParticipantSession
	var semanticErr error
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		verification, err := lockVerification(ctx, q, input)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if verification.ConsumedAt != nil {
			return feedback.ErrVerificationConsumed
		}
		if !verification.ExpiresAt.After(now) {
			return feedback.ErrVerificationExpired
		}
		if verification.AttemptCount >= maxVerificationCodeAttempts {
			return feedback.ErrVerificationAttempts
		}
		if len(input.CodeHash) > 0 && !hmac.Equal(verification.CodeHash, input.CodeHash) {
			attempts := verification.AttemptCount + 1
			count, updateErr := q.IncrementContributorVerificationAttempts(ctx, feedbacksql.IncrementContributorVerificationAttemptsParams{
				AttemptCount: attempts, VerificationID: verification.ID,
			})
			if updateErr != nil {
				return updateErr
			}
			if err = requireRowsAffected(count); err != nil {
				return err
			}
			if attempts >= maxVerificationCodeAttempts {
				semanticErr = feedback.ErrVerificationAttempts
			} else {
				semanticErr = feedback.ErrContributorSessionInvalid
			}
			return nil
		}
		participantRow, err := q.UpsertVerifiedFeedbackContributor(ctx, feedbacksql.UpsertVerifiedFeedbackContributorParams{
			PortalID: verification.PortalID, Email: verification.Email, Now: &now,
			DisplayName: valueOrZero(verification.DisplayName), PublicMasked: verification.PublicMasked,
		})
		if err != nil {
			return err
		}
		participant = verifiedParticipantProjection(participantRow).core()
		if participant.BlockedAt != nil {
			return feedback.ErrContributorBlocked
		}
		count, err := q.ConsumeContributorVerification(ctx, feedbacksql.ConsumeContributorVerificationParams{ConsumedAt: &now, VerificationID: verification.ID})
		if err != nil {
			return err
		}
		if err = requireRowsAffected(count); err != nil {
			return feedback.ErrVerificationConsumed
		}
		if err = q.EnsureFeedbackContributorPreferences(ctx, feedbacksql.EnsureFeedbackContributorPreferencesParams{PortalID: input.PortalID, ContributorID: participant.ID}); err != nil {
			return err
		}
		sessionRow, err := q.CreateFeedbackContributorSession(ctx, feedbacksql.CreateFeedbackContributorSessionParams{
			TokenHash: input.SessionTokenHash, Source: input.Source, ExpiresAt: input.SessionExpiresAt,
			PortalID: input.PortalID, ContributorID: participant.ID,
		})
		if err != nil {
			return err
		}
		session = coreContributorSession(sessionRow.ID, sessionRow.PortalID, sessionRow.ContributorID,
			sessionRow.Source, sessionRow.ExpiresAt, sessionRow.RevokedAt, sessionRow.LastUsedAt, sessionRow.CreatedAt)
		return nil
	})
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, normalizeError(err)
	}
	if semanticErr != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, semanticErr
	}
	return participant, session, nil
}

func lockVerification(ctx context.Context, q feedbacksql.Querier, input feedback.CoreVerificationConfirmation) (feedbacksql.FeedbackContributorVerification, error) {
	var row feedbacksql.FeedbackContributorVerification
	var err error
	if len(input.CodeHash) > 0 {
		row, err = q.LockContributorVerificationByCode(ctx, feedbacksql.LockContributorVerificationByCodeParams{
			PortalID: input.PortalID, Email: input.Email, Source: input.Source,
		})
	} else {
		row, err = q.LockContributorVerificationByToken(ctx, feedbacksql.LockContributorVerificationByTokenParams{
			PortalID: input.PortalID, TokenHash: input.TokenHash,
		})
	}
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && row.Source != input.Source) {
		return feedbacksql.FeedbackContributorVerification{}, feedback.ErrContributorSessionInvalid
	}
	return row, err
}

func (r *Repo) GetContributorSession(ctx context.Context, portalID uuid.UUID, tokenHash []byte, source string) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	row, err := r.queries.GetFeedbackContributorSession(ctx, feedbacksql.GetFeedbackContributorSessionParams{PortalID: portalID, TokenHash: tokenHash, Source: source})
	if err != nil {
		return feedback.CoreParticipant{}, feedback.CoreParticipantSession{}, normalizeError(err)
	}
	participant := sessionParticipantProjection(row).core()
	session := coreContributorSession(row.SessionID, row.PortalID, row.ID, row.SessionSource,
		row.SessionExpiresAt, row.SessionRevokedAt, row.SessionLastUsedAt, row.SessionCreatedAt)
	return participant, session, nil
}

func (r *Repo) RevokeContributorSession(ctx context.Context, portalID uuid.UUID, tokenHash []byte) error {
	count, err := r.queries.RevokeFeedbackContributorSession(ctx, feedbacksql.RevokeFeedbackContributorSessionParams{PortalID: portalID, TokenHash: tokenHash})
	if err != nil {
		return err
	}
	return requireRowsAffected(count)
}

func (r *Repo) GetParticipantByUser(ctx context.Context, portalID, userID uuid.UUID) (feedback.CoreParticipant, error) {
	row, err := r.queries.GetOrCreateAccountParticipant(ctx, feedbacksql.GetOrCreateAccountParticipantParams{UserID: userID, PortalID: portalID})
	if err != nil {
		return feedback.CoreParticipant{}, normalizeError(err)
	}
	participant := accountParticipantProjection(row).core()
	if participant.BlockedAt != nil {
		return feedback.CoreParticipant{}, feedback.ErrContributorBlocked
	}
	return participant, nil
}

func (r *Repo) ConsumeUnsubscribeToken(ctx context.Context, portalID uuid.UUID, tokenHash, sessionTokenHash []byte, sessionExpiresAt time.Time) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	var participant feedback.CoreParticipant
	var session feedback.CoreParticipantSession
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		token, err := q.LockFeedbackUnsubscribeToken(ctx, feedbacksql.LockFeedbackUnsubscribeTokenParams{PortalID: portalID, TokenHash: tokenHash})
		if err != nil {
			return normalizeError(err)
		}
		if token.ConsumedAt != nil {
			return feedback.ErrVerificationConsumed
		}
		if !token.ExpiresAt.After(time.Now().UTC()) {
			return feedback.ErrVerificationExpired
		}
		count, err := q.ConsumeFeedbackUnsubscribeToken(ctx, feedbacksql.ConsumeFeedbackUnsubscribeTokenParams{TokenID: token.ID})
		if err != nil {
			return err
		}
		if err = requireRowsAffected(count); err != nil {
			return feedback.ErrVerificationConsumed
		}
		switch token.Purpose {
		case "unsubscribe_item":
			if token.ItemID != nil {
				err = q.UnsubscribeFeedbackItem(ctx, feedbacksql.UnsubscribeFeedbackItemParams{PortalID: portalID, ItemID: *token.ItemID, ContributorID: token.ContributorID})
			}
		case "unsubscribe_portal":
			err = q.UnsubscribeFeedbackPortal(ctx, feedbacksql.UnsubscribeFeedbackPortalParams{PortalID: portalID, ContributorID: token.ContributorID})
		case "all_email":
			err = q.UnsubscribeAllFeedbackEmail(ctx, feedbacksql.UnsubscribeAllFeedbackEmailParams{PortalID: portalID, ContributorID: token.ContributorID})
		}
		if err != nil {
			return err
		}
		participantRow, err := q.GetFeedbackParticipant(ctx, feedbacksql.GetFeedbackParticipantParams{PortalID: portalID, ContributorID: token.ContributorID})
		if err != nil {
			return normalizeError(err)
		}
		participant = participantProjectionFromGet(participantRow).core()
		sessionRow, err := q.CreateFeedbackContributorSession(ctx, feedbacksql.CreateFeedbackContributorSessionParams{
			TokenHash: sessionTokenHash, Source: feedback.ContributorSessionSourcePreferences, ExpiresAt: sessionExpiresAt,
			PortalID: portalID, ContributorID: token.ContributorID,
		})
		if err != nil {
			return err
		}
		session = coreContributorSession(sessionRow.ID, sessionRow.PortalID, sessionRow.ContributorID,
			sessionRow.Source, sessionRow.ExpiresAt, sessionRow.RevokedAt, sessionRow.LastUsedAt, sessionRow.CreatedAt)
		return nil
	})
	return participant, session, err
}

func (r *Repo) CreateExternalContributorSession(ctx context.Context, portalID uuid.UUID, externalID, email, displayName string, avatarURL *string, tokenHash []byte, expiresAt time.Time) (feedback.CoreParticipant, feedback.CoreParticipantSession, error) {
	var participant feedback.CoreParticipant
	var session feedback.CoreParticipantSession
	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		now := time.Now().UTC()
		row, err := q.UpsertExternalFeedbackContributor(ctx, feedbacksql.UpsertExternalFeedbackContributorParams{
			PortalID: portalID, ExternalID: &externalID, Email: &email, Now: &now, DisplayName: displayName, AvatarURL: avatarURL,
		})
		if err != nil {
			return normalizeError(err)
		}
		participant = externalParticipantProjection(row).core()
		if participant.BlockedAt != nil {
			return feedback.ErrContributorBlocked
		}
		if err = q.EnsureFeedbackContributorPreferences(ctx, feedbacksql.EnsureFeedbackContributorPreferencesParams{PortalID: portalID, ContributorID: participant.ID}); err != nil {
			return err
		}
		sessionRow, err := q.CreateFeedbackContributorSession(ctx, feedbacksql.CreateFeedbackContributorSessionParams{
			TokenHash: tokenHash, Source: feedback.ContributorSessionSourceWidget, ExpiresAt: expiresAt,
			PortalID: portalID, ContributorID: participant.ID,
		})
		if err != nil {
			return err
		}
		session = coreContributorSession(sessionRow.ID, sessionRow.PortalID, sessionRow.ContributorID,
			sessionRow.Source, sessionRow.ExpiresAt, sessionRow.RevokedAt, sessionRow.LastUsedAt, sessionRow.CreatedAt)
		return nil
	})
	return participant, session, err
}

type participantProjection struct {
	ID, PortalID          uuid.UUID
	UserID                *uuid.UUID
	Kind, Email           string
	EmailVerifiedAt       *time.Time
	DisplayName           string
	AvatarURL, ExternalID *string
	PublicMasked          bool
	BlockedAt             *time.Time
	BlockedReason         *string
	LastSeenAt            *time.Time
	CreatedAt, UpdatedAt  time.Time
}

func (row participantProjection) core() feedback.CoreParticipant {
	return feedback.CoreParticipant{ID: row.ID, PortalID: row.PortalID, UserID: valueOrZero(row.UserID), Kind: row.Kind,
		Email: row.Email, EmailVerifiedAt: row.EmailVerifiedAt, DisplayName: row.DisplayName, AvatarURL: row.AvatarURL,
		ExternalID: valueOrZero(row.ExternalID), PublicMasked: row.PublicMasked, BlockedAt: row.BlockedAt,
		BlockedReason: valueOrZero(row.BlockedReason), LastSeenAt: row.LastSeenAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func verifiedParticipantProjection(row feedbacksql.UpsertVerifiedFeedbackContributorRow) participantProjection {
	return participantProjection{row.ID, row.PortalID, row.UserID, row.Kind, valueOrZero(row.Email), row.EmailVerifiedAt,
		row.DisplayName, row.AvatarURL, row.ExternalID, row.PublicMasked, row.BlockedAt, row.BlockedReason,
		row.LastSeenAt, row.CreatedAt, row.UpdatedAt}
}

func externalParticipantProjection(row feedbacksql.UpsertExternalFeedbackContributorRow) participantProjection {
	return participantProjection{row.ID, row.PortalID, row.UserID, row.Kind, valueOrZero(row.Email), row.EmailVerifiedAt,
		row.DisplayName, row.AvatarURL, row.ExternalID, row.PublicMasked, row.BlockedAt, row.BlockedReason,
		row.LastSeenAt, row.CreatedAt, row.UpdatedAt}
}

func accountParticipantProjection(row feedbacksql.GetOrCreateAccountParticipantRow) participantProjection {
	return participantProjection{row.ID, row.PortalID, row.UserID, row.Kind, row.Email, row.EmailVerifiedAt,
		row.DisplayName, row.AvatarURL, row.ExternalID, row.PublicMasked, row.BlockedAt, row.BlockedReason,
		row.LastSeenAt, row.CreatedAt, row.UpdatedAt}
}

func participantProjectionFromGet(row feedbacksql.GetFeedbackParticipantRow) participantProjection {
	return participantProjection{row.ID, row.PortalID, row.UserID, row.Kind, row.Email, row.EmailVerifiedAt,
		row.DisplayName, row.AvatarURL, row.ExternalID, row.PublicMasked, row.BlockedAt, row.BlockedReason,
		row.LastSeenAt, row.CreatedAt, row.UpdatedAt}
}

func sessionParticipantProjection(row feedbacksql.GetFeedbackContributorSessionRow) participantProjection {
	return participantProjection{row.ID, row.PortalID, row.UserID, row.Kind, row.Email, row.EmailVerifiedAt,
		row.DisplayName, row.AvatarURL, row.ExternalID, row.PublicMasked, row.BlockedAt, row.BlockedReason,
		row.LastSeenAt, row.CreatedAt, row.UpdatedAt}
}

func coreContributorSession(id, portalID, contributorID uuid.UUID, source string, expiresAt time.Time, revokedAt, lastUsedAt *time.Time, createdAt time.Time) feedback.CoreParticipantSession {
	return feedback.CoreParticipantSession{ID: id, PortalID: portalID, ContributorID: contributorID,
		Source: source, ExpiresAt: expiresAt, RevokedAt: revokedAt, LastUsedAt: lastUsedAt, CreatedAt: createdAt}
}
