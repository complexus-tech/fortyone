package invitationsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (r *repo) ClaimInvitationOutboxEvents(
	ctx context.Context,
	batchSize int,
	claimedAt time.Time,
	staleBefore time.Time,
) ([]invitationsdomain.OutboxEvent, error) {
	if batchSize <= 0 {
		return nil, errors.New("invitation outbox batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return nil, fmt.Errorf("validate invitation outbox batch size: %w", err)
	}
	claimedAt, staleBefore = claimedAt.UTC(), staleBefore.UTC()
	rows, err := r.queries.ClaimInvitationOutboxEvents(ctx, invitationsql.ClaimInvitationOutboxEventsParams{
		ClaimedAt:   &claimedAt,
		StaleBefore: &staleBefore,
		BatchSize:   databaseBatchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("claim invitation outbox: %w", err)
	}

	result := make([]invitationsdomain.OutboxEvent, 0, len(rows))
	for _, row := range rows {
		if row.ClaimToken == nil {
			return nil, errors.New("claimed invitation outbox row is missing a claim token")
		}
		event := invitationsdomain.OutboxEvent{
			ID:                  row.OutboxID,
			InvitationID:        row.InvitationID,
			WorkspaceID:         row.WorkspaceID,
			ActorID:             row.ActorID,
			EventType:           row.EventType,
			EventPayload:        append([]byte(nil), row.EventPayload...),
			IdempotencyKey:      row.IdempotencyKey,
			ClaimToken:          *row.ClaimToken,
			AttemptCount:        int(row.AttemptCount),
			CreatedAt:           row.CreatedAt,
			InvitationExpiresAt: row.InvitationExpiresAt,
			InvitationUsedAt:    row.InvitationUsedAt,
		}
		if row.EventType == string(events.InvitationEmail) && row.TokenKeyID != nil && row.TokenVersion != nil {
			event.StoredToken = &invitationsdomain.StoredToken{
				Digest:  append([]byte(nil), row.TokenDigest...),
				Nonce:   append([]byte(nil), row.TokenNonce...),
				KeyID:   *row.TokenKeyID,
				Version: *row.TokenVersion,
			}
		}
		result = append(result, event)
	}
	return result, nil
}

func (r *repo) CompleteInvitationOutboxEvent(
	ctx context.Context,
	outboxID uuid.UUID,
	claimToken uuid.UUID,
	completedAt time.Time,
) error {
	completedAt = completedAt.UTC()
	rows, err := r.queries.CompleteInvitationOutboxEvent(ctx, invitationsql.CompleteInvitationOutboxEventParams{
		CompletedAt: &completedAt,
		OutboxID:    outboxID,
		ClaimToken:  &claimToken,
	})
	if err != nil {
		return fmt.Errorf("complete invitation outbox: %w", err)
	}
	if rows != 1 {
		return invitationsdomain.ErrOutboxClaimLost
	}
	return nil
}

func (r *repo) RetryInvitationOutboxEvent(
	ctx context.Context,
	outboxID uuid.UUID,
	claimToken uuid.UUID,
	lastError string,
	retryAt time.Time,
	updatedAt time.Time,
	terminal bool,
) error {
	updatedAt = updatedAt.UTC()
	var (
		rows int64
		err  error
	)
	if terminal {
		rows, err = r.queries.FailInvitationOutboxEvent(ctx, invitationsql.FailInvitationOutboxEventParams{
			LastError:  lastError,
			UpdatedAt:  updatedAt,
			OutboxID:   outboxID,
			ClaimToken: &claimToken,
		})
	} else {
		retryAt = retryAt.UTC()
		rows, err = r.queries.RetryInvitationOutboxEvent(ctx, invitationsql.RetryInvitationOutboxEventParams{
			RetryAt:    &retryAt,
			LastError:  lastError,
			UpdatedAt:  updatedAt,
			OutboxID:   outboxID,
			ClaimToken: &claimToken,
		})
	}
	if err != nil {
		return fmt.Errorf("release invitation outbox claim: %w", err)
	}
	if rows != 1 {
		return invitationsdomain.ErrOutboxClaimLost
	}
	return nil
}
