package notificationsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	queries "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

// DispatchDocumentNotifications owns a bounded row lock through the Redis
// fanout. A crash retries the same notification ID; the existing push/digest
// tasks and clients deduplicate it. Revoked recipients are retired silently.
func (r *Repository) DispatchDocumentNotifications(ctx context.Context, limit int, deliver func(context.Context, domain.Notification) error) (int, error) {
	if r.pool == nil || deliver == nil || limit < 1 || limit > 100 {
		return 0, domain.ErrInvalid
	}
	count := 0
	for count < limit {
		found, err := r.dispatchDocumentNotification(ctx, deliver)
		if err != nil {
			return count, err
		}
		if !found {
			break
		}
		count++
	}
	return count, nil
}

func (r *Repository) dispatchDocumentNotification(ctx context.Context, deliver func(context.Context, domain.Notification) error) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer stop()
		_ = tx.Rollback(cleanup)
	}()
	q := queries.New(tx)
	row, err := q.NextDocumentNotificationDelivery(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim document notification: %w", err)
	}
	if row.Authorized {
		notification, err := toNotification(notificationRecord{ID: row.NotificationID, RecipientID: row.RecipientID, WorkspaceID: row.WorkspaceID, Type: row.Type, EntityType: row.EntityType, EntityID: row.EntityID, ActorID: row.ActorID, Title: row.Title, Message: row.Message, InAppEnabled: row.InAppEnabled, CreatedAt: row.CreatedAt, ReadAt: row.ReadAt})
		if err != nil {
			return false, err
		}
		if err := deliver(ctx, notification); err != nil {
			return false, err
		}
	}
	if err := q.CompleteDocumentNotificationDelivery(ctx, queries.CompleteDocumentNotificationDeliveryParams{NotificationID: row.NotificationID}); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
