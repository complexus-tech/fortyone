package chatrepository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type retryTransactionStub struct {
	pgx.Tx
	commitError error
	rollbacks   int
}

func (t *retryTransactionStub) Commit(context.Context) error   { return t.commitError }
func (t *retryTransactionStub) Rollback(context.Context) error { t.rollbacks++; return nil }

type retryBeginner struct {
	transactions []*retryTransactionStub
	options      []pgx.TxOptions
	commitError  func(int) error
}

func (b *retryBeginner) BeginTx(_ context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx := &retryTransactionStub{}
	if b.commitError != nil {
		tx.commitError = b.commitError(len(b.transactions))
	}
	b.transactions = append(b.transactions, tx)
	b.options = append(b.options, options)
	return tx, nil
}

func TestChatTransactionRetriesOperationAndCommitFailuresWithoutDuplicateEffects(t *testing.T) {
	for _, code := range []string{"40001", "40P01"} {
		for _, phase := range []string{"operation", "commit"} {
			t.Run(code+"/"+phase, func(t *testing.T) {
				failure := fmt.Errorf("wrapped database error: %w", &pgconn.PgError{Code: code})
				beginner := &retryBeginner{}
				if phase == "commit" {
					beginner.commitError = func(attempt int) error {
						if attempt < 4 {
							return failure
						}
						return nil
					}
				}
				calls, committed, rolledBack := 0, 0, 0
				err := withinRetryableChatUnit(t.Context(), beginner, func(ctx context.Context) error {
					// The prior attempt has fully rolled back before another begins.
					for _, previous := range beginner.transactions[:len(beginner.transactions)-1] {
						require.Equal(t, 1, previous.rollbacks)
					}
					_, active := platformdatabase.UnitTransaction(ctx)
					require.True(t, active)
					calls++
					platformdatabase.AfterCommit(ctx, func() { committed++ })
					platformdatabase.AfterRollback(ctx, func() { rolledBack++ })
					if phase == "operation" && calls <= 4 {
						return failure
					}
					return nil
				})
				require.NoError(t, err)
				require.Equal(t, 5, calls)
				require.Equal(t, 1, committed)
				require.Equal(t, 4, rolledBack)
				for _, option := range beginner.options {
					require.Equal(t, pgx.Serializable, option.IsoLevel)
				}
			})
		}
	}
}

func TestChatTransactionDoesNotRetryNonRetryableOrAmbiguousFailures(t *testing.T) {
	for _, test := range []struct {
		name              string
		operation, commit error
	}{
		{name: "validation", operation: errors.New("invalid message")},
		{name: "constraint", operation: &pgconn.PgError{Code: "23505"}},
		{name: "ambiguous commit", commit: io.ErrUnexpectedEOF},
	} {
		t.Run(test.name, func(t *testing.T) {
			beginner := &retryBeginner{commitError: func(int) error { return test.commit }}
			calls, committed := 0, 0
			err := withinRetryableChatUnit(t.Context(), beginner, func(ctx context.Context) error {
				calls++
				platformdatabase.AfterCommit(ctx, func() { committed++ })
				return test.operation
			})
			if test.operation != nil {
				require.ErrorIs(t, err, test.operation)
			} else {
				require.ErrorIs(t, err, test.commit)
			}
			require.Equal(t, 1, calls)
			require.Zero(t, committed)
		})
	}
}

func TestChatTransactionRetryBudgetAndCancellation(t *testing.T) {
	failure := &pgconn.PgError{Code: "40001"}
	beginner := &retryBeginner{}
	err := withinRetryableChatUnit(t.Context(), beginner, func(context.Context) error { return failure })
	require.ErrorIs(t, err, failure)
	require.Len(t, beginner.transactions, chatTransactionAttempts)
	for _, tx := range beginner.transactions {
		require.Equal(t, 1, tx.rollbacks)
	}

	ctx, cancel := context.WithCancel(t.Context())
	beginner = &retryBeginner{}
	err = withinRetryableChatUnit(ctx, beginner, func(context.Context) error { cancel(); return failure })
	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, beginner.transactions, 1)
	err = withinRetryableChatUnit(ctx, beginner, func(context.Context) error { t.Fatal("must not begin after cancellation"); return nil })
	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, beginner.transactions, 1)

	waiting, cancelWait := context.WithCancel(t.Context())
	finished := make(chan error, 1)
	go func() { finished <- waitForChatRetry(waiting, time.Hour) }()
	cancelWait()
	select {
	case err := <-finished:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt backoff")
	}
}

func TestChatNestedUnitRetainsCallerRetryOwnership(t *testing.T) {
	beginner := &retryBeginner{}
	repository := &Repository{}
	failure := &pgconn.PgError{Code: "40001"}
	calls := 0
	err := platformdatabase.WithinUnitOfWork(t.Context(), beginner, func(ctx context.Context, _ pgx.Tx) error {
		return repository.WithinUnit(ctx, func(context.Context) error { calls++; return failure })
	})
	require.ErrorIs(t, err, failure)
	require.Equal(t, 1, calls)
	require.Len(t, beginner.transactions, 1)
}

func TestChatRetryJitterHasBoundedNonzeroDelays(t *testing.T) {
	var maximumTotal time.Duration
	for attempt := range chatTransactionAttempts - 1 {
		ceiling := min(chatRetryInitialDelay<<min(attempt, 4), chatRetryMaximumDelay)
		maximumTotal += ceiling
		for range 100 {
			delay := chatRetryDelay(attempt)
			require.GreaterOrEqual(t, delay, ceiling/2)
			require.LessOrEqual(t, delay, ceiling)
		}
	}
	require.Equal(t, 630*time.Millisecond, maximumTotal)
}
