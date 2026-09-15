package useruow

import (
	"context"
	"errors"
	"fmt"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProviderCleanup interface {
	StageAccountDeletion(context.Context, pgx.Tx, uuid.UUID) error
	PendingAccountDeletion(context.Context, pgx.Tx, uuid.UUID) (bool, error)
}

// Manager commits identity erasure, credential revocation, provider cleanup
// staging and object deletion queues together. It performs no remote I/O.
type Manager struct {
	transactions platformdatabase.Transactor
	accounts     *usersrepository.AccountDeletionRepository
	cleanup      ProviderCleanup
}

func New(beginner platformdatabase.Beginner, accounts *usersrepository.AccountDeletionRepository, cleanup ProviderCleanup) (*Manager, error) {
	if beginner == nil || accounts == nil || cleanup == nil {
		return nil, errors.New("account deletion requires transaction, account and provider repositories")
	}
	return &Manager{transactions: platformdatabase.NewTransactor(beginner), accounts: accounts, cleanup: cleanup}, nil
}

func (manager *Manager) Delete(ctx context.Context, command usersdomain.AccountDeletion) (bool, error) {
	if err := command.Validate(); err != nil {
		return false, err
	}
	var pending bool
	err := manager.transact(ctx, func(tx pgx.Tx) error {
		snapshot, err := manager.accounts.Lock(ctx, tx, command.UserID)
		if err != nil {
			return err
		}
		if err := manager.accounts.Deactivate(ctx, tx, command); err != nil {
			return err
		}
		if err := manager.cleanup.StageAccountDeletion(ctx, tx, command.UserID); err != nil {
			return err
		}
		if err := manager.accounts.Erase(ctx, tx, snapshot, command.RequestedAt); err != nil {
			return err
		}
		pending, err = manager.cleanup.PendingAccountDeletion(ctx, tx, command.UserID)
		if err != nil {
			return err
		}
		if pending {
			return manager.accounts.Queue(ctx, tx, command)
		}
		return manager.accounts.Delete(ctx, tx, command.UserID)
	})
	// Subscriber cleanup is always queued. Its worker removes the last retained
	// provider contact outside this transaction, even when the user FK is gone.
	return err == nil, err
}

// FinalizePending is safe under overlapping deliveries. SKIP LOCKED avoids
// duplicate finalization; touching undrained requests prevents batch starvation.
func (manager *Manager) FinalizePending(ctx context.Context, limit int) (int, error) {
	if limit < 1 || limit > 500 {
		return 0, errors.New("account deletion batch must be between 1 and 500")
	}
	ids, err := manager.accounts.Pending(ctx, int32(limit))
	if err != nil {
		return 0, err
	}
	completed := 0
	for _, userID := range ids {
		finalized := false
		err := manager.transact(ctx, func(tx pgx.Tx) error {
			finalized = false
			locked, err := manager.accounts.LockPending(ctx, tx, userID)
			if err != nil || !locked {
				return err
			}
			pending, err := manager.cleanup.PendingAccountDeletion(ctx, tx, userID)
			if err != nil {
				return err
			}
			if pending {
				return manager.accounts.TouchPending(ctx, tx, userID, time.Now().UTC())
			}
			if err := manager.accounts.Delete(ctx, tx, userID); err != nil {
				return err
			}
			finalized = true
			return nil
		})
		if err != nil {
			return completed, fmt.Errorf("finalize account deletion: %w", err)
		}
		if finalized {
			completed++
		}
	}
	return completed, nil
}

func (manager *Manager) transact(ctx context.Context, operation func(pgx.Tx) error) error {
	for attempt := 0; ; attempt++ {
		err := manager.transactions.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, operation)
		var postgres *pgconn.PgError
		if attempt >= 2 || !errors.As(err, &postgres) || (postgres.Code != "40001" && postgres.Code != "40P01") || ctx.Err() != nil {
			return err
		}
	}
}
