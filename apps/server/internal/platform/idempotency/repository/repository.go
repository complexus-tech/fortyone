// Package repository owns the PostgreSQL receipt state machine and keeps sqlc
// generated types inside the shared idempotency persistence boundary.
package repository

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	idempotencysql "github.com/complexus-tech/projects-api/internal/platform/idempotency/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	databaseStateInProgress = "in_progress"
	databaseStateCompleted  = "completed"
)

var (
	ErrLeaseLost            = errors.New("idempotency receipt lease was lost")
	ErrInvalidStoredReceipt = errors.New("invalid stored idempotency receipt")
)

type BeginState uint8

const (
	BeginNew BeginState = iota + 1
	BeginInProgress
	BeginCompleted
	BeginConflict
)

type Scope struct {
	PrincipalKind  string
	PrincipalID    uuid.UUID
	WorkspaceID    *uuid.UUID
	HTTPMethod     string
	RouteOperation string
	KeyDigest      []byte
}

type BeginParams struct {
	Scope          Scope
	ReceiptID      uuid.UUID
	RequestHash    []byte
	Now            time.Time
	LeaseExpiresAt time.Time
	ExpiresAt      time.Time
}

type Lease struct {
	ReceiptID  uuid.UUID
	Generation int64
	ExpiresAt  time.Time
}

type Replay struct {
	StatusCode  int
	Body        []byte
	ContentType string
}

type BeginResult struct {
	State     BeginState
	Lease     Lease
	Replay    Replay
	RetryAt   time.Time
	Reclaimed bool
}

type CompleteParams struct {
	Lease       Lease
	Response    Replay
	CompletedAt time.Time
	ExpiresAt   time.Time
}

type Store struct {
	queries    *idempotencysql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) (*Store, error) {
	if pool == nil {
		return nil, errors.New("idempotency repository pool is required")
	}
	return &Store{
		queries:    idempotencysql.New(pool),
		transactor: platformdatabase.NewTransactor(pool),
	}, nil
}

func (s *Store) Begin(ctx context.Context, params BeginParams) (BeginResult, error) {
	var result BeginResult
	err := s.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := s.queries.WithTx(tx)
		if err := queries.LockReceiptScope(ctx, lockParams(params.Scope)); err != nil {
			return fmt.Errorf("lock idempotency receipt scope: %w", err)
		}

		stored, err := queries.GetReceiptForUpdate(ctx, getParams(params.Scope))
		if errors.Is(err, pgx.ErrNoRows) {
			created, createErr := queries.CreateReceipt(ctx, createParams(params))
			if createErr != nil {
				return fmt.Errorf("create idempotency receipt: %w", createErr)
			}
			result, createErr = newLeaseResult(created.ReceiptID, created.LeaseGeneration, created.LeaseExpiresAt, false)
			return createErr
		}
		if err != nil {
			return fmt.Errorf("get idempotency receipt: %w", err)
		}

		if !stored.ExpiresAt.After(params.Now) {
			restarted, restartErr := queries.RestartExpiredReceipt(ctx, idempotencysql.RestartExpiredReceiptParams{
				NewReceiptID:   params.ReceiptID,
				RequestHash:    cloneBytes(params.RequestHash),
				LeaseExpiresAt: timePointer(params.LeaseExpiresAt),
				ExpiresAt:      params.ExpiresAt,
				RestartedAt:    params.Now,
				ReceiptID:      stored.ReceiptID,
			})
			if restartErr != nil {
				return fmt.Errorf("restart expired idempotency receipt: %w", restartErr)
			}
			result, restartErr = newLeaseResult(restarted.ReceiptID, restarted.LeaseGeneration, restarted.LeaseExpiresAt, false)
			return restartErr
		}

		if subtle.ConstantTimeCompare(stored.RequestHash, params.RequestHash) != 1 {
			result = BeginResult{State: BeginConflict}
			return nil
		}

		switch stored.State {
		case databaseStateCompleted:
			replay, replayErr := replayFromStored(stored)
			if replayErr != nil {
				return replayErr
			}
			result = BeginResult{State: BeginCompleted, Replay: replay}
			return nil
		case databaseStateInProgress:
			if stored.LeaseExpiresAt == nil {
				return ErrInvalidStoredReceipt
			}
			if stored.LeaseExpiresAt.After(params.Now) {
				result = BeginResult{State: BeginInProgress, RetryAt: stored.LeaseExpiresAt.UTC()}
				return nil
			}
			takenOver, takeoverErr := queries.TakeOverStaleReceipt(ctx, idempotencysql.TakeOverStaleReceiptParams{
				LeaseExpiresAt: timePointer(params.LeaseExpiresAt),
				ExpiresAt:      params.ExpiresAt,
				TakenOverAt:    params.Now,
				ReceiptID:      stored.ReceiptID,
				RequestHash:    cloneBytes(params.RequestHash),
			})
			if takeoverErr != nil {
				return fmt.Errorf("take over stale idempotency receipt: %w", takeoverErr)
			}
			result, takeoverErr = newLeaseResult(
				takenOver.ReceiptID,
				takenOver.LeaseGeneration,
				takenOver.LeaseExpiresAt,
				true,
			)
			return takeoverErr
		default:
			return fmt.Errorf("%w: unknown lifecycle state", ErrInvalidStoredReceipt)
		}
	})
	if err != nil {
		return BeginResult{}, err
	}
	return result, nil
}

func (s *Store) Complete(ctx context.Context, params CompleteParams) error {
	statusCode, err := safecast.Int32(params.Response.StatusCode)
	if err != nil {
		return fmt.Errorf("validate idempotency response status: %w", err)
	}
	contentType := params.Response.ContentType
	rows, err := s.queries.CompleteReceipt(ctx, idempotencysql.CompleteReceiptParams{
		ResponseStatus:      &statusCode,
		ResponseBody:        cloneBytes(params.Response.Body),
		ResponseContentType: &contentType,
		CompletedAt:         timePointer(params.CompletedAt),
		ExpiresAt:           params.ExpiresAt,
		ReceiptID:           params.Lease.ReceiptID,
		LeaseGeneration:     params.Lease.Generation,
	})
	if err != nil {
		return fmt.Errorf("complete idempotency receipt: %w", err)
	}
	if rows != 1 {
		return ErrLeaseLost
	}
	return nil
}

func (s *Store) PurgeExpired(ctx context.Context, expiredAt time.Time, batchSize int32) (int64, error) {
	rows, err := s.queries.DeleteExpiredReceipts(ctx, idempotencysql.DeleteExpiredReceiptsParams{
		ExpiredAt: expiredAt,
		BatchSize: batchSize,
	})
	if err != nil {
		return 0, fmt.Errorf("purge expired idempotency receipts: %w", err)
	}
	return rows, nil
}

func lockParams(scope Scope) idempotencysql.LockReceiptScopeParams {
	return idempotencysql.LockReceiptScopeParams{
		PrincipalKind:  scope.PrincipalKind,
		PrincipalID:    scope.PrincipalID,
		WorkspaceID:    scope.WorkspaceID,
		HttpMethod:     scope.HTTPMethod,
		RouteOperation: scope.RouteOperation,
		KeyDigest:      cloneBytes(scope.KeyDigest),
	}
}

func getParams(scope Scope) idempotencysql.GetReceiptForUpdateParams {
	return idempotencysql.GetReceiptForUpdateParams{
		PrincipalKind:  scope.PrincipalKind,
		PrincipalID:    scope.PrincipalID,
		WorkspaceID:    scope.WorkspaceID,
		HttpMethod:     scope.HTTPMethod,
		RouteOperation: scope.RouteOperation,
		KeyDigest:      cloneBytes(scope.KeyDigest),
	}
}

func createParams(params BeginParams) idempotencysql.CreateReceiptParams {
	return idempotencysql.CreateReceiptParams{
		ReceiptID:      params.ReceiptID,
		PrincipalKind:  params.Scope.PrincipalKind,
		PrincipalID:    params.Scope.PrincipalID,
		WorkspaceID:    params.Scope.WorkspaceID,
		HttpMethod:     params.Scope.HTTPMethod,
		RouteOperation: params.Scope.RouteOperation,
		KeyDigest:      cloneBytes(params.Scope.KeyDigest),
		RequestHash:    cloneBytes(params.RequestHash),
		LeaseExpiresAt: timePointer(params.LeaseExpiresAt),
		ExpiresAt:      params.ExpiresAt,
		CreatedAt:      params.Now,
	}
}

func newLeaseResult(
	receiptID uuid.UUID,
	generation int64,
	expiresAt *time.Time,
	reclaimed bool,
) (BeginResult, error) {
	if receiptID == uuid.Nil || generation <= 0 || expiresAt == nil {
		return BeginResult{}, ErrInvalidStoredReceipt
	}
	return BeginResult{
		State: BeginNew,
		Lease: Lease{
			ReceiptID:  receiptID,
			Generation: generation,
			ExpiresAt:  expiresAt.UTC(),
		},
		Reclaimed: reclaimed,
	}, nil
}

func replayFromStored(row idempotencysql.GetReceiptForUpdateRow) (Replay, error) {
	if row.ResponseStatus == nil || row.ResponseContentType == nil {
		return Replay{}, ErrInvalidStoredReceipt
	}
	return Replay{
		StatusCode:  int(*row.ResponseStatus),
		Body:        cloneBytes(row.ResponseBody),
		ContentType: *row.ResponseContentType,
	}, nil
}

func timePointer(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

func cloneBytes(value []byte) []byte {
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
