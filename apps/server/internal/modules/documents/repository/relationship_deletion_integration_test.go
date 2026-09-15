//go:build integration

package documentsrepository

import (
	"context"
	"testing"
	"time"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestDocumentRelationshipInsertSerializesWithTargetDeletion(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	for _, entityType := range []string{"story", "objective"} {
		for _, deleteFirst := range []bool{true, false} {
			order := "insert holds target first"
			if deleteFirst {
				order = "deletion holds target first"
			}
			t.Run(entityType+"/"+order, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				defer cancel()
				fixture := newDocumentFixture(t, ctx, postgres.Pool)
				document := createDocument(t, ctx, New(postgres.Pool), fixture.workspaceA, fixture.ownerA, documentdomain.VisibilityWorkspace, "Surviving document")
				entityID := fixture.storyA
				lockTargetSQL := "SELECT id FROM stories WHERE id = $1 FOR UPDATE"
				deleteTargetSQL := "DELETE FROM stories WHERE id = $1"
				if entityType == "objective" {
					entityID = uuid.New()
					_, err := postgres.Pool.Exec(ctx, `
						INSERT INTO objectives (objective_id, workspace_id, team_id, name, sequence_id)
						VALUES ($1, $2, $3, 'Linked objective', 1)
					`, entityID, fixture.workspaceA, fixture.teamA)
					require.NoError(t, err)
					lockTargetSQL = "SELECT objective_id FROM objectives WHERE objective_id = $1 FOR UPDATE"
					deleteTargetSQL = "DELETE FROM objectives WHERE objective_id = $1"
				}
				params := documentssql.InsertEditableDocumentRelationshipParams{
					DocumentID: document.ID, WorkspaceID: fixture.workspaceA, ActorID: fixture.ownerA,
					EntityType: entityType, EntityID: entityID,
				}
				cleanupSQL := "DELETE FROM document_relationships WHERE workspace_id = $1 AND entity_type = $2 AND entity_id = $3"
				first, err := postgres.Pool.Begin(ctx)
				require.NoError(t, err)
				defer first.Rollback(ctx)

				if deleteFirst {
					_, err = first.Exec(ctx, lockTargetSQL, entityID)
					require.NoError(t, err)
					// Team deletion removes polymorphic links after locking their
					// targets and before the target rows disappear in its cascade.
					_, err = first.Exec(ctx, cleanupSQL, fixture.workspaceA, entityType, entityID)
					require.NoError(t, err)
					pid, completed := startDocumentDeletionRaceQuery(ctx, postgres.Pool, func(connection *pgxpool.Conn) error {
						_, insertErr := documentssql.New(connection).InsertEditableDocumentRelationship(ctx, params)
						return insertErr
					})
					assertDocumentDeletionQueryBlocked(t, ctx, postgres.Pool, pid)
					_, err = first.Exec(ctx, deleteTargetSQL, entityID)
					require.NoError(t, err)
					require.NoError(t, first.Commit(ctx))
					require.ErrorIs(t, awaitDocumentDeletionRaceQuery(t, ctx, completed), pgx.ErrNoRows)
				} else {
					_, err = documentssql.New(first).InsertEditableDocumentRelationship(ctx, params)
					require.NoError(t, err)
					pid, completed := startDocumentDeletionRaceQuery(ctx, postgres.Pool, func(connection *pgxpool.Conn) error {
						return pgx.BeginTxFunc(ctx, connection, pgx.TxOptions{}, func(deletion pgx.Tx) error {
							if _, err := deletion.Exec(ctx, lockTargetSQL, entityID); err != nil {
								return err
							}
							if _, err := deletion.Exec(ctx, cleanupSQL, fixture.workspaceA, entityType, entityID); err != nil {
								return err
							}
							_, err := deletion.Exec(ctx, deleteTargetSQL, entityID)
							return err
						})
					})
					assertDocumentDeletionQueryBlocked(t, ctx, postgres.Pool, pid)
					require.NoError(t, first.Commit(ctx))
					require.NoError(t, awaitDocumentDeletionRaceQuery(t, ctx, completed))
				}

				var relationships int
				require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT count(*) FROM document_relationships WHERE document_id = $1", document.ID).Scan(&relationships))
				require.Zero(t, relationships, "concurrent target deletion must never leave a dangling document relationship")
				_, err = New(postgres.Pool).Get(ctx, fixture.workspaceA, fixture.ownerA, document.ID)
				require.NoError(t, err, "target deletion must preserve the document")
			})
		}
	}
}

// Each asynchronous operation owns its connection until its query completes;
// cancellation releases blocked connections even when a test assertion fails.
func startDocumentDeletionRaceQuery(
	ctx context.Context,
	pool *pgxpool.Pool,
	query func(*pgxpool.Conn) error,
) (<-chan uint32, <-chan error) {
	pid := make(chan uint32, 1)
	completed := make(chan error, 1)
	go func() {
		connection, err := pool.Acquire(ctx)
		if err != nil {
			close(pid)
			completed <- err
			return
		}
		defer connection.Release()
		pid <- connection.Conn().PgConn().PID()
		completed <- query(connection)
	}()
	return pid, completed
}

func assertDocumentDeletionQueryBlocked(t *testing.T, ctx context.Context, pool *pgxpool.Pool, pids <-chan uint32) {
	t.Helper()
	var pid uint32
	select {
	case value, ok := <-pids:
		require.True(t, ok, "concurrent query must acquire a connection")
		pid = value
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	require.Eventually(t, func() bool {
		var blocked bool
		err := pool.QueryRow(ctx, "SELECT cardinality(pg_blocking_pids($1)) > 0", pid).Scan(&blocked)
		return err == nil && blocked
	}, 2*time.Second, 10*time.Millisecond, "query must wait on the opposing target row lock")
}

func awaitDocumentDeletionRaceQuery(t *testing.T, ctx context.Context, completed <-chan error) error {
	t.Helper()
	select {
	case err := <-completed:
		return err
	case <-ctx.Done():
		t.Fatal(ctx.Err())
		return ctx.Err()
	}
}
