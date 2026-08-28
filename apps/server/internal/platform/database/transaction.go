package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrNilTransactionOperation = errors.New("transaction operation is required")

// Beginner is the narrow pgx capability required to start a transaction. Both
// *pgxpool.Pool and *pgx.Conn satisfy this interface.
type Beginner interface {
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}

// Transactor owns transaction lifecycle. Repositories keep ownership of the
// business invariant and bind their generated query sets to the supplied tx.
type Transactor struct {
	beginner Beginner
}

func NewTransactor(beginner Beginner) Transactor {
	return Transactor{beginner: beginner}
}

// WithinTransaction executes operation once inside a transaction configured
// with options. It commits only after a nil operation result and rolls back on
// an error or panic. A panic is deliberately not recovered.
func (t Transactor) WithinTransaction(
	ctx context.Context,
	options pgx.TxOptions,
	operation func(pgx.Tx) error,
) error {
	if t.beginner == nil {
		return errors.New("transaction beginner is required")
	}
	if operation == nil {
		return ErrNilTransactionOperation
	}

	return pgx.BeginTxFunc(ctx, t.beginner, options, operation)
}
