package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestTransactorCommitsSuccessfulOperation(t *testing.T) {
	t.Parallel()

	tx := &transactionSpy{}
	beginner := &beginnerStub{tx: tx}
	transactor := NewTransactor(beginner)

	err := transactor.WithinTransaction(context.Background(), pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}, func(actual pgx.Tx) error {
		if actual != tx {
			t.Fatalf("operation transaction = %T, want transaction spy", actual)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WithinTransaction() error = %v", err)
	}
	if tx.commits != 1 {
		t.Fatalf("commit count = %d, want 1", tx.commits)
	}
	if tx.effectiveRollbacks != 0 {
		t.Fatalf("effective rollback count = %d, want 0", tx.effectiveRollbacks)
	}
	if beginner.options.IsoLevel != pgx.Serializable || beginner.options.AccessMode != pgx.ReadWrite {
		t.Fatalf("transaction options = %#v, want serializable read-write", beginner.options)
	}
}

func TestTransactorRollsBackFailedOperation(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("operation failed")
	tx := &transactionSpy{}
	transactor := NewTransactor(&beginnerStub{tx: tx})

	err := transactor.WithinTransaction(context.Background(), pgx.TxOptions{}, func(pgx.Tx) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("WithinTransaction() error = %v, want %v", err, wantErr)
	}
	if tx.commits != 0 {
		t.Fatalf("commit count = %d, want 0", tx.commits)
	}
	if tx.effectiveRollbacks != 1 {
		t.Fatalf("effective rollback count = %d, want 1", tx.effectiveRollbacks)
	}
}

func TestTransactorRollsBackAndPropagatesPanic(t *testing.T) {
	t.Parallel()

	tx := &transactionSpy{}
	transactor := NewTransactor(&beginnerStub{tx: tx})

	defer func() {
		if recovered := recover(); recovered != "boom" {
			t.Fatalf("recovered panic = %v, want boom", recovered)
		}
		if tx.commits != 0 {
			t.Fatalf("commit count = %d, want 0", tx.commits)
		}
		if tx.effectiveRollbacks != 1 {
			t.Fatalf("effective rollback count = %d, want 1", tx.effectiveRollbacks)
		}
	}()

	_ = transactor.WithinTransaction(context.Background(), pgx.TxOptions{}, func(pgx.Tx) error {
		panic("boom")
	})
}

func TestTransactorRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	if err := (Transactor{}).WithinTransaction(context.Background(), pgx.TxOptions{}, func(pgx.Tx) error {
		return nil
	}); err == nil {
		t.Fatal("WithinTransaction() error = nil, want missing beginner error")
	}

	transactor := NewTransactor(&beginnerStub{tx: &transactionSpy{}})
	if err := transactor.WithinTransaction(context.Background(), pgx.TxOptions{}, nil); !errors.Is(err, ErrNilTransactionOperation) {
		t.Fatalf("WithinTransaction() error = %v, want ErrNilTransactionOperation", err)
	}
}

type beginnerStub struct {
	options pgx.TxOptions
	tx      pgx.Tx
	err     error
}

func (b *beginnerStub) BeginTx(_ context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	b.options = options
	return b.tx, b.err
}

type transactionSpy struct {
	pgx.Tx
	commits            int
	rollbackCalls      int
	effectiveRollbacks int
	closed             bool
}

func (tx *transactionSpy) Commit(context.Context) error {
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.commits++
	tx.closed = true
	return nil
}

func (tx *transactionSpy) Rollback(context.Context) error {
	tx.rollbackCalls++
	if tx.closed {
		return pgx.ErrTxClosed
	}
	tx.effectiveRollbacks++
	tx.closed = true
	return nil
}
