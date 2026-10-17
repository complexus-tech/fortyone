package chatadapter

import (
	"context"
	"errors"
	"time"

	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttachmentRetirer joins chat relation changes to the shared durable storage
// lifecycle. It never performs remote storage work inside a transaction.
type AttachmentRetirer struct {
	repository *attachmentsrepository.Repository
	pool       *pgxpool.Pool
	provider   string
	container  string
}

func NewAttachmentRetirer(pool *pgxpool.Pool, provider, container string) *AttachmentRetirer {
	return &AttachmentRetirer{repository: attachmentsrepository.New(pool), pool: pool, provider: provider, container: container}
}

func (r *AttachmentRetirer) Retire(ctx context.Context, workspaceID uuid.UUID, ids []uuid.UUID) error {
	tx, ok := platformdatabase.UnitTransaction(ctx)
	if !ok {
		return errors.New("chat attachment retirement requires the message transaction")
	}
	return r.repository.RetireUnreferencedAttachments(ctx, tx, ids, workspaceID, r.provider, r.container, time.Now().UTC())
}

// RetireUnlinkedUpload compensates for a server-created upload whose pending
// chat link failed. The all-consumer guard also protects an ambiguous commit.
func (r *AttachmentRetirer) RetireUnlinkedUpload(ctx context.Context, workspaceID, id uuid.UUID) error {
	return platformdatabase.WithinUnitOfWork(ctx, r.pool, func(unit context.Context, _ pgx.Tx) error {
		return r.Retire(unit, workspaceID, []uuid.UUID{id})
	})
}
