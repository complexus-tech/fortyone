package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type unitOfWorkKey struct{}
type unitOfWork struct {
	tx            pgx.Tx
	active        bool
	afterCommit   []func()
	afterRollback []func()
}

// WithinUnitOfWork composes existing transactional repositories with a durable
// operation receipt. Work is sequential: callers must not share its context
// across goroutines. Repositories use savepoints and cannot commit this unit.
func WithinUnitOfWork(ctx context.Context, beginner Beginner, operation func(context.Context, pgx.Tx) error) error {
	if beginner == nil {
		return errors.New("transaction beginner is required")
	}
	if operation == nil {
		return ErrNilTransactionOperation
	}
	if _, ok := UnitTransaction(ctx); ok {
		return errors.New("unit of work is already active")
	}
	tx, err := beginner.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	unit := &unitOfWork{tx: tx, active: true}
	commitAttempted := false
	defer func() {
		unit.active = false
		// Cancellation must not leave the connection open or prevent blob cleanup.
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
		if !commitAttempted {
			for _, callback := range unit.afterRollback {
				callback()
			}
		}
	}()
	if err := operation(context.WithValue(ctx, unitOfWorkKey{}, unit), tx); err != nil {
		return err
	}
	commitAttempted = true
	if err := tx.Commit(ctx); err != nil {
		// A transport error can mean COMMIT succeeded. Never delete uploaded
		// objects in that case: retrying the same operation resolves its receipt.
		var postgresError *pgconn.PgError
		if errors.Is(err, pgx.ErrTxCommitRollback) || errors.As(err, &postgresError) {
			for _, callback := range unit.afterRollback {
				callback()
			}
		}
		return err
	}
	unit.active = false
	for _, callback := range unit.afterCommit {
		callback()
	}
	return nil
}

// UnitTransaction binds single-statement repositories to the caller's unit.
// The caller owns its lifecycle; repositories must not commit or roll it back.
func UnitTransaction(ctx context.Context) (pgx.Tx, bool) {
	unit, ok := ctx.Value(unitOfWorkKey{}).(*unitOfWork)
	if !ok || !unit.active {
		return nil, false
	}
	return unit.tx, true
}

// AfterCommit defers best-effort side effects when a caller owns a larger
// transaction. Existing standalone service calls retain their original timing.
// Durable effects must still be written to the domain outbox in the transaction.
func AfterCommit(ctx context.Context, callback func()) {
	if unit, ok := ctx.Value(unitOfWorkKey{}).(*unitOfWork); ok && unit.active {
		unit.afterCommit = append(unit.afterCommit, callback)
		return
	}
	callback()
}

// AfterRollback registers cleanup for objects created outside PostgreSQL.
// It runs only when the transaction is known not to have committed.
func AfterRollback(ctx context.Context, callback func()) {
	if unit, ok := ctx.Value(unitOfWorkKey{}).(*unitOfWork); ok && unit.active {
		unit.afterRollback = append(unit.afterRollback, callback)
	}
}
