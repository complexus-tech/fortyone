package database

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type unitTransaction struct {
	transactionSpy
	commitError error
	child       *unitTransaction
}

func (tx *unitTransaction) Commit(ctx context.Context) error {
	_ = tx.transactionSpy.Commit(ctx)
	return tx.commitError
}
func (tx *unitTransaction) Begin(context.Context) (pgx.Tx, error) { return tx.child, nil }

func TestUnitOfWorkSideEffectsFollowCommitOutcome(t *testing.T) {
	operationError := errors.New("mutation rejected")
	for _, test := range []struct {
		name                        string
		operationError, commitError error
		committed, cleaned          bool
	}{
		{name: "committed", committed: true},
		{name: "operation rejected", operationError: operationError, cleaned: true},
		{name: "serialization failure", commitError: &pgconn.PgError{Code: "40001"}, cleaned: true},
		{name: "transaction aborted", commitError: pgx.ErrTxCommitRollback, cleaned: true},
		{name: "ambiguous commit", commitError: io.ErrUnexpectedEOF},
	} {
		t.Run(test.name, func(t *testing.T) {
			tx := &unitTransaction{commitError: test.commitError}
			beginner := &beginnerStub{tx: tx}
			committed, cleaned := false, false
			err := WithinUnitOfWork(t.Context(), beginner, func(ctx context.Context, actual pgx.Tx) error {
				require.Same(t, tx, actual)
				AfterCommit(ctx, func() {
					committed = true
					_, active := UnitTransaction(ctx)
					require.False(t, active, "post-commit callbacks must not reuse a committed transaction")
				})
				AfterRollback(ctx, func() { cleaned = true })
				require.False(t, committed)
				return test.operationError
			})
			if test.operationError != nil {
				require.ErrorIs(t, err, test.operationError)
			} else {
				require.ErrorIs(t, err, test.commitError)
			}
			require.Equal(t, test.committed, committed)
			require.Equal(t, test.cleaned, cleaned)
			require.Equal(t, pgx.Serializable, beginner.options.IsoLevel)
		})
	}
}

func TestUnitOfWorkRepositoriesUseSavepointsAndCannotCommitOuterUnit(t *testing.T) {
	child := &unitTransaction{}
	outer := &unitTransaction{child: child}
	beginner := &beginnerStub{tx: outer}
	err := WithinUnitOfWork(t.Context(), beginner, func(ctx context.Context, tx pgx.Tx) error {
		err := NewTransactor(beginner).WithinTransaction(ctx, pgx.TxOptions{}, func(actual pgx.Tx) error {
			require.Same(t, child, actual)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, child.commits)
		require.Zero(t, outer.commits)
		return errors.New("receipt write failed")
	})
	require.ErrorContains(t, err, "receipt write failed")
	require.Zero(t, outer.commits)
	require.Equal(t, 1, outer.effectiveRollbacks)
}

func TestUnitOfWorkCleansUpPanicWithoutSwallowingIt(t *testing.T) {
	tx := &unitTransaction{}
	cleaned := false
	require.PanicsWithValue(t, "interrupted", func() {
		_ = WithinUnitOfWork(t.Context(), &beginnerStub{tx: tx}, func(ctx context.Context, _ pgx.Tx) error {
			AfterRollback(ctx, func() { cleaned = true })
			panic("interrupted")
		})
	})
	require.True(t, cleaned)
	require.Equal(t, 1, tx.effectiveRollbacks)
}
