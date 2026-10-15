package desktoprepository

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type receiptKey struct{ actor, workspace, operation uuid.UUID }
type receipt struct {
	hash     []byte
	resource uuid.UUID
	version  *string
}
type receiptDatabase struct {
	receipts    map[receiptKey]receipt
	commitError error
	insertError error
	commits     int
}
type receiptTransaction struct {
	pgx.Tx
	db      *receiptDatabase
	key     receiptKey
	pending *receipt
}
type receiptRow struct {
	value receipt
	found bool
}

func (row receiptRow) Scan(values ...any) error {
	if !row.found {
		return pgx.ErrNoRows
	}
	*values[0].(*[]byte) = row.value.hash
	*values[1].(*uuid.UUID) = row.value.resource
	*values[2].(**string) = row.value.version
	return nil
}
func (db *receiptDatabase) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return &receiptTransaction{db: db}, nil
}
func (tx *receiptTransaction) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	tx.key = receiptKey{args[0].(uuid.UUID), args[1].(uuid.UUID), args[2].(uuid.UUID)}
	value, found := tx.db.receipts[tx.key]
	return receiptRow{value: value, found: found}
}
func (tx *receiptTransaction) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	if len(args) == 6 {
		if tx.db.insertError != nil {
			return pgconn.CommandTag{}, tx.db.insertError
		}
		tx.pending = &receipt{hash: append([]byte{}, args[3].([]byte)...), resource: args[4].(uuid.UUID), version: args[5].(*string)}
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}
func (tx *receiptTransaction) Commit(context.Context) error {
	if tx.pending != nil {
		tx.db.receipts[tx.key] = *tx.pending
	}
	tx.db.commits++
	err := tx.db.commitError
	tx.db.commitError = nil
	return err
}
func (tx *receiptTransaction) Rollback(context.Context) error { tx.pending = nil; return nil }

func TestOperationReceiptResolvesLostResponseWithoutRepeatingMutation(t *testing.T) {
	db := &receiptDatabase{receipts: make(map[receiptKey]receipt), commitError: io.ErrUnexpectedEOF}
	store := New(db)
	actor, workspace, operation, resource := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	calls := 0
	mutate := func(context.Context) (Result, error) { calls++; return Result{ResourceID: resource}, nil }
	_, _, err := store.Execute(t.Context(), actor, workspace, operation, []byte(`{"title":"Draft"}`), mutate)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	id, replayed, err := store.Execute(t.Context(), actor, workspace, operation, []byte(`{"title":"Draft"}`), mutate)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, resource, id.ResourceID)
	require.Equal(t, 1, calls)
}

func TestOperationReceiptRejectsChangedPayloadAndScopesIdentity(t *testing.T) {
	db := &receiptDatabase{receipts: make(map[receiptKey]receipt)}
	store := New(db)
	actor, workspace, operation := uuid.New(), uuid.New(), uuid.New()
	calls := 0
	mutate := func(context.Context) (Result, error) { calls++; return Result{ResourceID: uuid.New()}, nil }
	_, _, err := store.Execute(t.Context(), actor, workspace, operation, []byte("original"), mutate)
	require.NoError(t, err)
	_, _, err = store.Execute(t.Context(), actor, workspace, operation, []byte("changed"), mutate)
	require.ErrorIs(t, err, ErrOperationConflict)
	require.Equal(t, 1, calls)
	for _, identity := range []receiptKey{{uuid.New(), workspace, operation}, {actor, uuid.New(), operation}} {
		_, replayed, err := store.Execute(t.Context(), identity.actor, identity.workspace, operation, []byte("changed"), mutate)
		require.NoError(t, err)
		require.False(t, replayed)
	}
	require.Equal(t, 3, calls)
}

func TestOperationFailureNeverCommitsReceipt(t *testing.T) {
	for _, failInsert := range []bool{false, true} {
		db := &receiptDatabase{receipts: make(map[receiptKey]receipt)}
		failure := errors.New("persistence unavailable")
		if failInsert {
			db.insertError = failure
		}
		_, _, err := New(db).Execute(t.Context(), uuid.New(), uuid.New(), uuid.New(), []byte("command"), func(context.Context) (Result, error) {
			if !failInsert {
				return Result{}, failure
			}
			return Result{ResourceID: uuid.New()}, nil
		})
		require.ErrorIs(t, err, failure)
		require.Empty(t, db.receipts)
		require.Zero(t, db.commits)
	}
}

func TestReplayPreservesCommittedVersionAfterAnotherWriterChangesResource(t *testing.T) {
	db := &receiptDatabase{receipts: make(map[receiptKey]receipt)}
	store := New(db)
	actor, workspace, operation, resource := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	version := "2"
	calls := 0
	mutate := func(context.Context) (Result, error) {
		calls++
		committed := version
		return Result{ResourceID: resource, Version: &committed}, nil
	}
	result, _, err := store.Execute(t.Context(), actor, workspace, operation, []byte("offline A"), mutate)
	require.NoError(t, err)
	require.Equal(t, "2", *result.Version)
	version = "3" // Another actor edited after our successful commit.
	replayedResult, replayed, err := store.Execute(t.Context(), actor, workspace, operation, []byte("offline A"), mutate)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, "2", *replayedResult.Version, "dependent offline B must conflict with v3, never rebase over it")
	require.Equal(t, 1, calls)
}
