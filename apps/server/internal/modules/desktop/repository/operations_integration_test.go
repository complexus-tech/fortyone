//go:build integration

package desktoprepository

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentsrepository "github.com/complexus-tech/projects-api/internal/modules/documents/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Simulate a real COMMIT whose acknowledgement never reaches the transport.
type lostCommitPool struct {
	pool *pgxpool.Pool
	lose atomic.Bool
}
type lostCommitTx struct {
	pgx.Tx
	owner *lostCommitPool
}

func (p *lostCommitPool) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := p.pool.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &lostCommitTx{Tx: tx, owner: p}, nil
}
func (tx *lostCommitTx) Commit(ctx context.Context) error {
	if err := tx.Tx.Commit(ctx); err != nil {
		return err
	}
	if tx.owner.lose.Swap(false) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func TestDesktopReceiptAndDocumentMutationCommitOrRollbackTogether(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace, actor := uuid.New(), uuid.New()
	_, err := postgres.Pool.Exec(ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Offline', $2)`, workspace, "offline-"+workspace.String())
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO users(user_id,username,email,full_name,is_active,is_system) VALUES($1,$2,$3,'Offline member',TRUE,FALSE)`, actor, "offline-"+actor.String(), actor.String()+"@example.com")
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,CAST('member' AS public.user_role))`, workspace, actor)
	require.NoError(t, err)
	service := documents.New(documentsrepository.New(postgres.Pool))
	pool := &lostCommitPool{pool: postgres.Pool}
	store := New(pool)
	operationID := uuid.New()
	calls := 0
	mutation := func(ctx context.Context) (Result, error) {
		calls++
		document, err := service.Create(ctx, documents.CoreCreateInput{WorkspaceID: workspace, UserID: actor, Title: "Offline draft", Visibility: documentdomain.VisibilityPrivate})
		version := "1"
		return Result{ResourceID: document.ID, Version: &version}, err
	}
	pool.lose.Store(true)
	first, _, err := store.Execute(ctx, actor, workspace, operationID, []byte("create document"), mutation)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	otherTitle := "Another actor's online edit"
	_, err = service.Update(ctx, documents.CoreUpdateInput{WorkspaceID: workspace, UserID: actor, DocumentID: first.ResourceID, ExpectedRevision: 1, Title: &otherTitle})
	require.NoError(t, err)
	second, replayed, err := store.Execute(ctx, actor, workspace, operationID, []byte("create document"), mutation)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, first, second)
	require.Equal(t, "1", *second.Version, "replay must preserve committed version before the online edit")
	require.Equal(t, 1, calls)
	var documentsCount, receiptsCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM documents WHERE workspace_id=$1`, workspace).Scan(&documentsCount))
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM desktop_operation_receipts WHERE workspace_id=$1`, workspace).Scan(&receiptsCount))
	require.Equal(t, 1, documentsCount)
	require.Equal(t, 1, receiptsCount)

	// A domain write followed by a rejected operation must not survive merely
	// because the repository released its inner savepoint successfully.
	rejected := errors.New("receipt rejected after domain write")
	_, _, err = store.Execute(ctx, actor, workspace, uuid.New(), []byte("rolled back"), func(ctx context.Context) (Result, error) {
		id, err := mutation(ctx)
		if err != nil {
			return id, err
		}
		return id, rejected
	})
	require.ErrorIs(t, err, rejected)
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM documents WHERE workspace_id=$1`, workspace).Scan(&documentsCount))
	require.Equal(t, 1, documentsCount)
}
