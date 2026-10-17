package notificationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	queries "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) DeferMayaReply(ctx context.Context, input domain.NewNotification, commentID uuid.UUID) error {
	payload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode deferred Maya reply: %w", err)
	}
	matches, err := r.queries.EnqueueMayaReplyNotification(ctx, queries.EnqueueMayaReplyNotificationParams{
		DedupeKey: input.DedupeKey, CommentID: commentID, RecipientID: input.RecipientID,
		WorkspaceID: input.WorkspaceID, StoryID: input.EntityID, ActorID: input.ActorID,
		Payload: payload, OccurredAt: input.OccurredAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// A deleted reply no longer needs an announcement.
		return nil
	}
	if err != nil {
		return fmt.Errorf("defer Maya reply notification: %w", err)
	}
	if matches == nil || !*matches {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) DispatchMayaReplies(ctx context.Context, limit int, deliver func(context.Context, domain.NewNotification) error) (int, error) {
	if r.pool == nil || deliver == nil || limit < 1 || limit > 100 {
		return 0, domain.ErrInvalid
	}
	count := 0
	var failures error
	for count < limit {
		found, err := r.dispatchMayaReply(ctx, deliver)
		if !found {
			return count, errors.Join(failures, err)
		}
		count++
		failures = errors.Join(failures, err)
	}
	return count, failures
}

// A bounded row lock serializes competing workers. The notification's existing
// event receipt makes a crash after persistence safe to replay. Transient errors
// retain the decision with bounded backoff; revoked access retires it quietly.
func (r *Repository) dispatchMayaReply(ctx context.Context, deliver func(context.Context, domain.NewNotification) error) (bool, error) {
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
	row, err := q.NextMayaReplyNotification(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim Maya reply notification: %w", err)
	}
	var deliveryErr error
	if !row.Seen && row.Authorized {
		var input domain.NewNotification
		if err := json.Unmarshal(row.Payload, &input); err != nil {
			return false, fmt.Errorf("decode deferred Maya reply: %w", err)
		}
		input.DedupeKey, input.OccurredAt = row.DedupeKey, row.OccurredAt
		input.InAppEnabled = &row.InAppEnabled
		deliveryErr = deliver(ctx, input)
	}
	if deliveryErr == nil || errors.Is(deliveryErr, domain.ErrForbidden) || errors.Is(deliveryErr, domain.ErrNotFound) {
		err = q.CompleteMayaReplyNotification(ctx, queries.CompleteMayaReplyNotificationParams{DedupeKey: row.DedupeKey})
		deliveryErr = nil
	} else {
		err = q.RetryMayaReplyNotification(ctx, queries.RetryMayaReplyNotificationParams{DedupeKey: row.DedupeKey})
	}
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, deliveryErr
}
