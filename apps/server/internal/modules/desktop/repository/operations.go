package desktoprepository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrOperationConflict = errors.New("operation ID was already used for different changes")

// Result records the exact version produced by this committed operation.
// A later resource refresh must never become the base of dependent offline edits.
type Result struct {
	ResourceID uuid.UUID
	Version    *string
}

type Store struct{ beginner platformdatabase.Beginner }

func New(beginner platformdatabase.Beginner) *Store { return &Store{beginner: beginner} }

// Execute never separates the receipt from its mutation. A lost HTTP response
// can safely be retried with the same actor/workspace/ID and exact command.
// Request bodies and credentials are not retained in the receipt table.
func (s *Store) Execute(ctx context.Context, actorID, workspaceID, operationID uuid.UUID, command []byte, mutate func(context.Context) (Result, error)) (Result, bool, error) {
	if s == nil || s.beginner == nil || mutate == nil || actorID == uuid.Nil || workspaceID == uuid.Nil || operationID == uuid.Nil {
		return Result{}, false, errors.New("desktop operation store is not configured")
	}
	hash := sha256.Sum256(command)
	var result Result
	var replayed bool
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		result, replayed = Result{}, false
		err = platformdatabase.WithinUnitOfWork(ctx, s.beginner, func(ctx context.Context, tx pgx.Tx) error {
			// Serialize the same operation across API processes; unrelated operations
			// retain concurrency. Serializable retries handle a snapshot taken before
			// an earlier holder committed its receipt.
			key := fmt.Sprintf("desktop:%s:%s:%s", actorID, workspaceID, operationID)
			if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", key); err != nil {
				return err
			}
			var storedHash []byte
			err := tx.QueryRow(ctx, `SELECT request_hash, resource_id, operation_version FROM desktop_operation_receipts WHERE actor_id=$1 AND workspace_id=$2 AND operation_id=$3`, actorID, workspaceID, operationID).Scan(&storedHash, &result.ResourceID, &result.Version)
			if err == nil {
				if !bytes.Equal(storedHash, hash[:]) {
					return ErrOperationConflict
				}
				replayed = true
				return nil
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			result, err = mutate(ctx)
			if err != nil {
				return err
			}
			if result.ResourceID == uuid.Nil {
				return errors.New("desktop mutation returned no resource identity")
			}
			_, err = tx.Exec(ctx, `INSERT INTO desktop_operation_receipts (actor_id, workspace_id, operation_id, request_hash, resource_id, operation_version) VALUES ($1,$2,$3,$4,$5,$6)`, actorID, workspaceID, operationID, hash[:], result.ResourceID, result.Version)
			return err
		})
		if err == nil || !retryableReceiptError(err) {
			break
		}
		if ctx.Err() != nil {
			return Result{}, false, ctx.Err()
		}
	}
	return result, replayed, err
}

func retryableReceiptError(err error) bool {
	if platformdatabase.IsRetryableTransactionError(err) {
		return true
	}
	// A waiter can acquire its serializable snapshot before the earlier holder
	// commits. Its receipt insert can then report a uniqueness conflict instead
	// of serialization_failure. Restart the whole unit to read the winner.
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == "desktop_operation_receipts_pkey"
}
