package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/idempotency/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidLease   = errors.New("invalid idempotency receipt lease")
	ErrLeaseLost      = repository.ErrLeaseLost
	ErrInvalidPurge   = errors.New("invalid idempotency purge request")
	ErrInvalidService = errors.New("invalid idempotency service")
)

type BeginState string

const (
	BeginStateNew        BeginState = "new"
	BeginStateInProgress BeginState = "in_progress"
	BeginStateCompleted  BeginState = "completed"
	BeginStateConflict   BeginState = "conflict"
)

type Lease struct {
	ReceiptID  uuid.UUID
	Generation int64
	ExpiresAt  time.Time
}

func (l Lease) validate() error {
	if l.ReceiptID == uuid.Nil || l.Generation <= 0 || l.ExpiresAt.IsZero() {
		return ErrInvalidLease
	}
	return nil
}

// BeginResult uses state-specific fields: New includes Lease; InProgress
// includes RetryAt; Completed includes Replay; Conflict includes neither.
type BeginResult struct {
	State     BeginState
	Lease     Lease
	Replay    Response
	RetryAt   time.Time
	Reclaimed bool
}

type Service struct {
	store        *repository.Store
	config       Config
	now          func() time.Time
	newReceiptID func() uuid.UUID
}

func New(pool *pgxpool.Pool, config Config) (*Service, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	store, err := repository.New(pool)
	if err != nil {
		return nil, err
	}
	return newService(store, config, time.Now, uuid.New), nil
}

func newService(
	store *repository.Store,
	config Config,
	now func() time.Time,
	newReceiptID func() uuid.UUID,
) *Service {
	return &Service{
		store:        store,
		config:       config,
		now:          now,
		newReceiptID: newReceiptID,
	}
}

func (s *Service) Begin(
	ctx context.Context,
	scope Scope,
	key Key,
	requestBody []byte,
) (BeginResult, error) {
	if err := s.validate(); err != nil {
		return BeginResult{}, err
	}
	if err := scope.validate(); err != nil {
		return BeginResult{}, err
	}
	if err := key.validate(); err != nil {
		return BeginResult{}, err
	}
	if err := validateRequestBody(requestBody); err != nil {
		return BeginResult{}, err
	}

	now := s.now().UTC()
	keyDigest := key.digest()
	requestHash := HashRequest(requestBody)
	result, err := s.store.Begin(ctx, repository.BeginParams{
		Scope: repository.Scope{
			PrincipalKind:  string(scope.principalKind),
			PrincipalID:    scope.principalID,
			WorkspaceID:    optionalUUID(scope.workspaceID),
			HTTPMethod:     string(scope.method),
			RouteOperation: scope.operation.value,
			KeyDigest:      keyDigest[:],
		},
		ReceiptID:      s.newReceiptID(),
		RequestHash:    requestHash[:],
		Now:            now,
		LeaseExpiresAt: now.Add(s.config.LeaseDuration),
		ExpiresAt:      now.Add(s.config.RetentionDuration),
	})
	if err != nil {
		return BeginResult{}, err
	}
	return beginResultFromRepository(result)
}

func (s *Service) Complete(ctx context.Context, lease Lease, response Response) error {
	if err := s.validate(); err != nil {
		return err
	}
	if err := lease.validate(); err != nil {
		return err
	}
	if err := response.validate(); err != nil {
		return err
	}

	now := s.now().UTC()
	return s.store.Complete(ctx, repository.CompleteParams{
		Lease: repository.Lease{
			ReceiptID:  lease.ReceiptID,
			Generation: lease.Generation,
			ExpiresAt:  lease.ExpiresAt.UTC(),
		},
		Response: repository.Replay{
			StatusCode:  response.statusCode,
			Body:        cloneBytes(response.body),
			ContentType: response.contentType,
		},
		CompletedAt: now,
		ExpiresAt:   now.Add(s.config.RetentionDuration),
	})
}

func (s *Service) PurgeExpired(ctx context.Context, batchSize int) (int64, error) {
	if err := s.validate(); err != nil {
		return 0, err
	}
	if batchSize < 1 || batchSize > MaxPurgeBatchSize {
		return 0, fmt.Errorf("%w: batch size must be between 1 and %d", ErrInvalidPurge, MaxPurgeBatchSize)
	}
	return s.store.PurgeExpired(ctx, s.now().UTC(), int32(batchSize))
}

func (s *Service) validate() error {
	if s == nil || s.store == nil || s.now == nil || s.newReceiptID == nil {
		return ErrInvalidService
	}
	return s.config.validate()
}

func beginResultFromRepository(result repository.BeginResult) (BeginResult, error) {
	switch result.State {
	case repository.BeginNew:
		lease := Lease{
			ReceiptID:  result.Lease.ReceiptID,
			Generation: result.Lease.Generation,
			ExpiresAt:  result.Lease.ExpiresAt.UTC(),
		}
		if err := lease.validate(); err != nil {
			return BeginResult{}, err
		}
		return BeginResult{
			State:     BeginStateNew,
			Lease:     lease,
			Reclaimed: result.Reclaimed,
		}, nil
	case repository.BeginInProgress:
		if result.RetryAt.IsZero() {
			return BeginResult{}, repository.ErrInvalidStoredReceipt
		}
		return BeginResult{State: BeginStateInProgress, RetryAt: result.RetryAt.UTC()}, nil
	case repository.BeginCompleted:
		replay, err := NewResponse(result.Replay.StatusCode, result.Replay.Body, result.Replay.ContentType)
		if err != nil {
			return BeginResult{}, fmt.Errorf("%w: replay metadata", repository.ErrInvalidStoredReceipt)
		}
		return BeginResult{State: BeginStateCompleted, Replay: replay}, nil
	case repository.BeginConflict:
		return BeginResult{State: BeginStateConflict}, nil
	default:
		return BeginResult{}, repository.ErrInvalidStoredReceipt
	}
}

func optionalUUID(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	return &value
}
