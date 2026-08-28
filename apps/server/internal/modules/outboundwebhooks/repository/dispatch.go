package outboundwebhooksrepository

import (
	"context"
	"fmt"
	"math"
	"net/netip"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhookssql "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) ClaimNextDelivery(
	ctx context.Context,
	leaseToken uuid.UUID,
	claimedAt, leaseExpiresAt time.Time,
) (outboundwebhooksdomain.ClaimedDelivery, error) {
	if err := repository.configured(); err != nil {
		return outboundwebhooksdomain.ClaimedDelivery{}, err
	}
	if leaseToken == uuid.Nil || claimedAt.IsZero() || !leaseExpiresAt.After(claimedAt) {
		return outboundwebhooksdomain.ClaimedDelivery{}, outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	var delivery outboundwebhooksdomain.ClaimedDelivery
	err := repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		endpointID, err := queries.LockNextOutboundWebhookEndpointForDelivery(ctx, outboundwebhookssql.LockNextOutboundWebhookEndpointForDeliveryParams{
			ClaimedAt: claimedAt.UTC(),
		})
		if err != nil {
			return mapReadError(err, outboundwebhooksdomain.ErrDeliveryNotFound)
		}
		lease := leaseToken
		expires := leaseExpiresAt.UTC()
		row, err := queries.ClaimNextOutboundWebhookDelivery(ctx, outboundwebhookssql.ClaimNextOutboundWebhookDeliveryParams{
			EndpointID: endpointID, ClaimedAt: claimedAt.UTC(), LeaseToken: &lease, LeaseExpiresAt: &expires,
		})
		if err != nil {
			return mapReadError(err, outboundwebhooksdomain.ErrDeliveryNotFound)
		}
		claimed := claimedAt.UTC()
		if rows, err := queries.TouchOutboundWebhookEndpointClaim(ctx, outboundwebhookssql.TouchOutboundWebhookEndpointClaimParams{
			ClaimedAt: &claimed, EndpointID: endpointID,
		}); err != nil {
			return fmt.Errorf("touch outbound webhook endpoint claim: %w", err)
		} else if rows != 1 {
			return outboundwebhooksdomain.ErrDeliveryLeaseLost
		}
		delivery, err = mapClaimedDelivery(row)
		return err
	})
	if err != nil {
		return outboundwebhooksdomain.ClaimedDelivery{}, err
	}
	return delivery, nil
}

func (repository *Repository) CompleteAttempt(ctx context.Context, attempt outboundwebhooksdomain.DeliveryAttempt, workspaceID, endpointID uuid.UUID) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if err := validateAttempt(attempt, workspaceID, endpointID); err != nil {
		return err
	}
	attemptNumber, err := safecast.Int32(attempt.AttemptNumber)
	if err != nil {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	disableAfterFailures, err := safecast.Int32(attempt.DisableAfterFailures)
	if err != nil {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	durationMilliseconds, err := safecast.Int64ToInt32(attempt.Duration.Milliseconds())
	if err != nil {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		var resolvedIP *netip.Addr
		if attempt.ResolvedIP != nil {
			address := attempt.ResolvedIP.Unmap()
			resolvedIP = &address
		}
		if err := queries.RecordOutboundWebhookDeliveryAttempt(ctx, outboundwebhookssql.RecordOutboundWebhookDeliveryAttemptParams{
			AttemptID: attempt.ID, DeliveryID: attempt.DeliveryID, AttemptNumber: attemptNumber,
			Outcome: string(attempt.Outcome), ResolvedIp: resolvedIP,
			HttpStatus: optionalInt32(attempt.HTTPStatus), ResponseBytes: optionalInt32(attempt.ResponseBytes),
			ResponseDigest: optionalBytes(attempt.ResponseDigest), ErrorCode: optionalString(attempt.ErrorCode),
			DurationMs: durationMilliseconds, StartedAt: attempt.StartedAt.UTC(), FinishedAt: attempt.FinishedAt.UTC(),
		}); err != nil {
			return fmt.Errorf("record outbound webhook delivery attempt: %w", err)
		}

		status, completedAt, nextAttemptAt, err := completionState(attempt)
		if err != nil {
			return err
		}
		leaseToken := attempt.LeaseToken
		rows, err := queries.CompleteOutboundWebhookDelivery(ctx, outboundwebhookssql.CompleteOutboundWebhookDeliveryParams{
			Status: string(status), HttpStatus: optionalInt32(attempt.HTTPStatus), ErrorCode: optionalString(attempt.ErrorCode),
			NextAttemptAt: nextAttemptAt, CompletedAt: completedAt, FinishedAt: attempt.FinishedAt.UTC(),
			DeliveryID: attempt.DeliveryID, LeaseToken: &leaseToken, AttemptNumber: attemptNumber,
		})
		if err != nil {
			return fmt.Errorf("complete outbound webhook delivery: %w", err)
		}
		if rows != 1 {
			return outboundwebhooksdomain.ErrDeliveryLeaseLost
		}
		if attempt.Outcome == outboundwebhooksdomain.AttemptSucceeded {
			succeededAt := attempt.FinishedAt.UTC()
			rows, err = queries.RecordOutboundWebhookEndpointSuccess(ctx, outboundwebhookssql.RecordOutboundWebhookEndpointSuccessParams{
				SucceededAt: &succeededAt, EndpointID: endpointID, WorkspaceID: workspaceID,
			})
		} else if attempt.CountEndpointFailure {
			reason := attempt.ErrorCode
			if reason == "" {
				reason = "delivery_failed"
			}
			failedAt := attempt.FinishedAt.UTC()
			rows, err = queries.RecordOutboundWebhookEndpointFailure(ctx, outboundwebhookssql.RecordOutboundWebhookEndpointFailureParams{
				DisableEndpoint: attempt.DisableEndpoint, DisableAfterFailures: disableAfterFailures,
				FailedAt: &failedAt, DisabledReason: &reason,
				EndpointID: endpointID, WorkspaceID: workspaceID,
			})
		} else {
			return nil
		}
		if err != nil {
			return fmt.Errorf("update outbound webhook endpoint delivery state: %w", err)
		}
		if rows != 1 {
			return outboundwebhooksdomain.ErrEndpointNotFound
		}
		return nil
	})
}

func (repository *Repository) RecoverExpiredLeases(ctx context.Context, recoveredAt time.Time) (int64, error) {
	if err := repository.configured(); err != nil {
		return 0, err
	}
	rows, err := repository.queries.RecoverExpiredOutboundWebhookLeases(ctx, outboundwebhookssql.RecoverExpiredOutboundWebhookLeasesParams{
		RecoveredAt: recoveredAt.UTC(),
	})
	if err != nil {
		return 0, fmt.Errorf("recover expired outbound webhook leases: %w", err)
	}
	return rows, nil
}

func validateAttempt(attempt outboundwebhooksdomain.DeliveryAttempt, workspaceID, endpointID uuid.UUID) error {
	if attempt.ID == uuid.Nil || attempt.DeliveryID == uuid.Nil || attempt.LeaseToken == uuid.Nil ||
		workspaceID == uuid.Nil || endpointID == uuid.Nil || attempt.AttemptNumber < 1 || attempt.AttemptNumber > 32 ||
		attempt.DisableAfterFailures < 1 || attempt.DisableAfterFailures > 1000 ||
		attempt.StartedAt.IsZero() || attempt.FinishedAt.Before(attempt.StartedAt) || attempt.Duration < 0 || attempt.Duration > 30*time.Second {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	switch attempt.Outcome {
	case outboundwebhooksdomain.AttemptSucceeded, outboundwebhooksdomain.AttemptRetryScheduled,
		outboundwebhooksdomain.AttemptFailed, outboundwebhooksdomain.AttemptCancelled:
	default:
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	if attempt.Outcome == outboundwebhooksdomain.AttemptRetryScheduled && attempt.NextAttemptAt == nil {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	if attempt.Outcome != outboundwebhooksdomain.AttemptRetryScheduled && attempt.NextAttemptAt != nil {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	if attempt.DisableEndpoint && !attempt.CountEndpointFailure {
		return outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
	return nil
}

func completionState(attempt outboundwebhooksdomain.DeliveryAttempt) (outboundwebhooksdomain.DeliveryStatus, *time.Time, *time.Time, error) {
	finishedAt := attempt.FinishedAt.UTC()
	switch attempt.Outcome {
	case outboundwebhooksdomain.AttemptSucceeded:
		return outboundwebhooksdomain.DeliverySucceeded, &finishedAt, nil, nil
	case outboundwebhooksdomain.AttemptRetryScheduled:
		return outboundwebhooksdomain.DeliveryRetryScheduled, nil, attempt.NextAttemptAt, nil
	case outboundwebhooksdomain.AttemptFailed:
		return outboundwebhooksdomain.DeliveryFailed, &finishedAt, nil, nil
	case outboundwebhooksdomain.AttemptCancelled:
		return outboundwebhooksdomain.DeliveryCancelled, &finishedAt, nil, nil
	default:
		return "", nil, nil, outboundwebhooksdomain.ErrDeliveryLeaseLost
	}
}

func optionalInt32(value *int) *int32 {
	if value == nil {
		return nil
	}
	if *value < math.MinInt32 || *value > math.MaxInt32 {
		return nil
	}
	converted := int32(*value)
	return &converted
}

func optionalBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return append([]byte(nil), value...)
}
