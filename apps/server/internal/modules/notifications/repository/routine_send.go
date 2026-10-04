package notificationsrepository

import (
	"context"
	"fmt"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
)

// BeginRoutineSend fences the owned attempt before the external send. A begun
// attempt counts toward the local weekly limit until an explicit failure;
// recovery cannot infer that SMTP rejected a message from an expired claim.
func (r *Repository) BeginRoutineSend(ctx context.Context, id uuid.UUID, scope notifications.DeliveryScope, now time.Time) error {
	if err := scope.Validate(); err != nil {
		return err
	}
	if id == uuid.Nil || now.IsZero() {
		return fmt.Errorf("%w: routine send ID and time required", notifications.ErrInvalid)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin routine send transaction: %w", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	q := notificationssql.New(tx)
	if err := q.LockRoutineEmailRecipient(ctx, notificationssql.LockRoutineEmailRecipientParams{RecipientID: scope.RecipientID.String()}); err != nil {
		return fmt.Errorf("lock routine send recipient: %w", err)
	}
	count, err := q.BeginRoutineEmailSend(ctx, notificationssql.BeginRoutineEmailSendParams{
		ID: id, RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID, Now: now,
	})
	if err != nil {
		return fmt.Errorf("record routine send start: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("routine email claim can no longer start a send")
	}
	return tx.Commit(ctx)
}
