package notificationsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) ListRoutineRecipients(ctx context.Context, cursor *notifications.WeeklyDigestCursor, limit int) ([]notifications.RoutineRecipient, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("%w: routine recipient limit must be 1-100", notifications.ErrInvalid)
	}
	params := notificationssql.ListRoutineEmailRecipientsParams{RowLimit: int32(limit)}
	if cursor != nil {
		params.HasCursor = true
		params.AfterWorkspaceID = cursor.WorkspaceID
		params.AfterUserID = cursor.UserID
	}
	rows, err := r.queries.ListRoutineEmailRecipients(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list routine email recipients: %w", err)
	}
	recipients := make([]notifications.RoutineRecipient, 0, len(rows))
	for _, row := range rows {
		recipients = append(recipients, notifications.RoutineRecipient{UserID: row.UserID, WorkspaceID: row.WorkspaceID, Email: row.Email, Name: row.Name, WorkspaceName: row.WorkspaceName, WorkspaceSlug: row.WorkspaceSlug, Timezone: row.Timezone, WeeklyEnabled: row.WeeklyEnabled})
	}
	return recipients, nil
}

// GetRoutineRecipient returns only active members of an active workspace.
// External feedback contributors have no routine guidance audience.
func (r *Repository) GetRoutineRecipient(ctx context.Context, scope notifications.DeliveryScope) (*notifications.RoutineRecipient, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	row, err := r.queries.GetRoutineEmailRecipient(ctx, notificationssql.GetRoutineEmailRecipientParams{RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get routine email recipient: %w", err)
	}
	return &notifications.RoutineRecipient{UserID: row.UserID, WorkspaceID: row.WorkspaceID, Email: row.Email, Name: row.Name, WorkspaceName: row.WorkspaceName, WorkspaceSlug: row.WorkspaceSlug, Timezone: row.Timezone, WeeklyEnabled: row.WeeklyEnabled}, nil
}

func (r *Repository) HasRoutineGuidance(ctx context.Context, scope notifications.DeliveryScope, date time.Time) (bool, error) {
	if err := scope.Validate(); err != nil {
		return false, err
	}
	return r.queries.HasRoutineEmailGuidance(ctx, notificationssql.HasRoutineEmailGuidanceParams{RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID, DeliveryKey: "briefing:" + date.Format("2006-01-02")})
}

// ClaimRoutine serializes the read-then-send window shared by activity and
// briefings. The advisory lock is held only during the claim transaction.
// Each reclaimed attempt gets a new ID, fencing out an expired owner.
func (r *Repository) ClaimRoutine(ctx context.Context, claim notifications.RoutineClaim) (uuid.UUID, error) {
	scope := notifications.DeliveryScope{RecipientID: claim.RecipientID, WorkspaceID: claim.WorkspaceID}
	if err := scope.Validate(); err != nil {
		return uuid.Nil, err
	}
	if claim.Key == "" || len(claim.Key) > 250 || (claim.Kind != "activity" && claim.Kind != "briefing") || claim.Now.IsZero() || claim.LocalDate.IsZero() {
		return uuid.Nil, fmt.Errorf("%w: invalid routine claim", notifications.ErrInvalid)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	q := notificationssql.New(tx)
	if err := q.LockRoutineEmailRecipient(ctx, notificationssql.LockRoutineEmailRecipientParams{RecipientID: claim.RecipientID.String()}); err != nil {
		return uuid.Nil, err
	}
	staleBefore := claim.Now.Add(-10 * time.Minute)
	active, err := q.HasActiveRoutineEmailClaim(ctx, notificationssql.HasActiveRoutineEmailClaimParams{RecipientID: claim.RecipientID, StaleBefore: staleBefore})
	if err != nil {
		return uuid.Nil, err
	}
	if active {
		return uuid.Nil, errors.New("another routine email is being delivered; retry this batch")
	}
	id, err := q.ClaimRoutineEmail(ctx, notificationssql.ClaimRoutineEmailParams{RecipientID: claim.RecipientID, WorkspaceID: claim.WorkspaceID, DeliveryKey: claim.Key, Kind: claim.Kind, LocalDate: claim.LocalDate, Now: claim.Now, StaleBefore: staleBefore})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("claim routine email: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// CompleteRoutine records the delivery and covers every notification in its
// snapshot, including overflow represented by the summary link, atomically.
func (r *Repository) CompleteRoutine(ctx context.Context, completion notifications.RoutineCompletion) error {
	if err := completion.Scope.Validate(); err != nil {
		return err
	}
	if completion.ID == uuid.Nil || completion.Now.IsZero() || (completion.GuidanceDate != nil && completion.GuidanceDate.IsZero()) {
		return fmt.Errorf("%w: routine completion ID and time required", notifications.ErrInvalid)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	q := notificationssql.New(tx)
	status := "skipped"
	if completion.Sent {
		status = "sent"
	}
	count, err := q.CompleteRoutineEmail(ctx, notificationssql.CompleteRoutineEmailParams{ID: completion.ID, RecipientID: completion.Scope.RecipientID, WorkspaceID: completion.Scope.WorkspaceID, Status: status, Now: &completion.Now})
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("routine email claim is no longer owned")
	}
	if len(completion.NotificationIDs) > 0 {
		_, err = q.MarkNotificationEmailsSent(ctx, notificationssql.MarkNotificationEmailsSentParams{RecipientID: completion.Scope.RecipientID, WorkspaceID: completion.Scope.WorkspaceID, SentAt: completion.Now, NotificationIds: completion.NotificationIDs})
		if err != nil {
			return err
		}
	}
	if completion.Sent && completion.GuidanceDate != nil {
		date := *completion.GuidanceDate
		if err := q.RecordRoutineEmailGuidance(ctx, notificationssql.RecordRoutineEmailGuidanceParams{RecipientID: completion.Scope.RecipientID, WorkspaceID: completion.Scope.WorkspaceID, DeliveryKey: "briefing:" + date.Format("2006-01-02"), LocalDate: date, Now: completion.Now}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) FailRoutine(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: routine claim ID required", notifications.ErrInvalid)
	}
	_, err := r.queries.FailRoutineEmail(ctx, notificationssql.FailRoutineEmailParams{ID: id})
	return err
}
