package calendarrepository

import (
	"context"
	"encoding/json"
	"fmt"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

type transactionScheduleEventOutboxStore struct {
	queries calendarsql.Querier
}

func (r *Repo) WithScheduleEventDispatchLock(ctx context.Context, userID uuid.UUID, dispatch func(calendar.ScheduleEventOutboxStore) error) error {
	if userID == uuid.Nil || dispatch == nil {
		return calendar.ErrInvalidScheduleBlock
	}
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, userID); err != nil {
			return err
		}
		return dispatch(transactionScheduleEventOutboxStore{queries: queries})
	}); err != nil {
		return fmt.Errorf("calendar schedule event dispatch transaction: %w", err)
	}
	return nil
}

func enqueueScheduleEventOutbox(ctx context.Context, queries calendarsql.Querier, workspaceID, userID uuid.UUID, blockID *uuid.UUID, provider calendar.Provider, operation calendar.ScheduleEventOperation, event calendar.ExternalScheduleEventInput, syncHash string, reactivateTerminal bool) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode calendar schedule event outbox payload: %w", err)
	}
	dedupeKey := fmt.Sprintf("%s:%s:%s:%s", provider, operation, event.BlockID, syncHash)
	if blockID != nil {
		if err := queries.SupersedeStaleScheduleEventOutbox(ctx, calendarsql.SupersedeStaleScheduleEventOutboxParams{
			ScheduleBlockID: blockID,
			DedupeKey:       dedupeKey,
			Provider:        string(provider),
		}); err != nil {
			return fmt.Errorf("supersede stale calendar schedule event outbox: %w", err)
		}
	}
	if err := queries.EnqueueScheduleEventOutbox(ctx, calendarsql.EnqueueScheduleEventOutboxParams{
		WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: blockID,
		Operation: string(operation), Provider: string(provider), CalendarID: event.CalendarID,
		ProviderEventID: event.EventID, Payload: payload, DedupeKey: dedupeKey,
		ReactivateTerminal: reactivateTerminal,
	}); err != nil {
		return fmt.Errorf("enqueue calendar schedule event outbox: %w", err)
	}
	return nil
}

func (r *Repo) ListReadyScheduleEventOutboxUsers(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	limit = boundedOutboxLimit(limit)
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("validate calendar schedule event outbox limit: %w", err)
	}
	userIDs, err := r.queries.ListReadyScheduleEventOutboxUsers(ctx, calendarsql.ListReadyScheduleEventOutboxUsersParams{
		GoogleOwnedScope:    calendar.GoogleCalendarEventsOwnedScope,
		GoogleReadScope:     calendar.GoogleCalendarEventsReadonlyScope,
		MicrosoftWriteScope: calendar.MicrosoftCalendarReadWriteScope,
		RowLimit:            rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list users with ready calendar schedule events: %w", err)
	}
	return userIDs, nil
}

func (r *Repo) ListPendingScheduleEventOutbox(ctx context.Context, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	return listPendingScheduleEventOutbox(ctx, r.queries, userID, provider, limit)
}

func (store transactionScheduleEventOutboxStore) ListPendingScheduleEventOutbox(ctx context.Context, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	return listPendingScheduleEventOutbox(ctx, store.queries, userID, provider, limit)
}

func listPendingScheduleEventOutbox(ctx context.Context, queries calendarsql.Querier, userID uuid.UUID, provider calendar.Provider, limit int) ([]calendar.CoreScheduleEventOutbox, error) {
	rowLimit, err := safecast.Int32(boundedOutboxLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("validate pending calendar schedule event outbox limit: %w", err)
	}
	rows, err := queries.ClaimPendingScheduleEventOutbox(ctx, calendarsql.ClaimPendingScheduleEventOutboxParams{
		UserID:   userID,
		Provider: string(provider),
		RowLimit: rowLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("claim calendar schedule event outbox: %w", err)
	}
	items := make([]calendar.CoreScheduleEventOutbox, len(rows))
	for index, row := range rows {
		items[index] = calendar.CoreScheduleEventOutbox{
			ID: row.OutboxID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
			ScheduleBlockID: row.ScheduleBlockID, Operation: calendar.ScheduleEventOperation(row.Operation),
			Provider: calendar.Provider(row.Provider), CalendarID: row.CalendarID,
			ProviderEventID: row.ProviderEventID, Payload: json.RawMessage(row.Payload),
			DedupeKey: row.DedupeKey, AttemptCount: int(row.AttemptCount),
		}
	}
	return items, nil
}

func boundedOutboxLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 100
	}
	return limit
}

func (store transactionScheduleEventOutboxStore) ScheduleEventUpsertIsCurrent(ctx context.Context, item calendar.CoreScheduleEventOutbox, event calendar.ExternalScheduleEventInput) (bool, error) {
	if item.ScheduleBlockID == nil {
		return false, nil
	}
	current, err := store.queries.ScheduleEventUpsertIsCurrent(ctx, calendarsql.ScheduleEventUpsertIsCurrentParams{
		BlockID: *item.ScheduleBlockID, WorkspaceID: item.WorkspaceID, UserID: item.UserID,
		Provider: string(item.Provider), CalendarID: item.CalendarID, ProviderEventID: item.ProviderEventID,
		Title: event.Title, StartAt: event.StartAt, EndAt: event.EndAt,
	})
	if err != nil {
		return false, fmt.Errorf("validate current calendar schedule upsert: %w", err)
	}
	return current, nil
}

func (r *Repo) MarkScheduleEventOutboxProcessed(ctx context.Context, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		return markScheduleEventOutboxProcessed(ctx, queries, item, syncHash)
	}); err != nil {
		return fmt.Errorf("complete calendar schedule event outbox transaction: %w", err)
	}
	return nil
}

func (store transactionScheduleEventOutboxStore) MarkScheduleEventOutboxProcessed(ctx context.Context, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	return markScheduleEventOutboxProcessed(ctx, store.queries, item, syncHash)
}

func markScheduleEventOutboxProcessed(ctx context.Context, queries calendarsql.Querier, item calendar.CoreScheduleEventOutbox, syncHash string) error {
	if err := queries.MarkScheduleEventOutboxProcessed(ctx, calendarsql.MarkScheduleEventOutboxProcessedParams{OutboxID: item.ID}); err != nil {
		return fmt.Errorf("complete calendar schedule event outbox: %w", err)
	}
	if item.Operation == calendar.ScheduleEventOperationUpsert && item.ScheduleBlockID != nil {
		if err := queries.MarkScheduleBlockMirrored(ctx, calendarsql.MarkScheduleBlockMirroredParams{
			Provider: string(item.Provider), CalendarID: item.CalendarID, ProviderEventID: item.ProviderEventID,
			SyncHash: syncHash, BlockID: *item.ScheduleBlockID,
		}); err != nil {
			return fmt.Errorf("mark calendar schedule block mirrored: %w", err)
		}
	}
	return nil
}

func (r *Repo) MarkScheduleEventOutboxFailed(ctx context.Context, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	if err := r.configured(); err != nil {
		return err
	}
	return markScheduleEventOutboxFailed(ctx, r.queries, item, message, permanent)
}

func (store transactionScheduleEventOutboxStore) MarkScheduleEventOutboxFailed(ctx context.Context, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	return markScheduleEventOutboxFailed(ctx, store.queries, item, message, permanent)
}

func markScheduleEventOutboxFailed(ctx context.Context, queries calendarsql.Querier, item calendar.CoreScheduleEventOutbox, message string, permanent bool) error {
	if err := queries.MarkScheduleEventOutboxFailed(ctx, calendarsql.MarkScheduleEventOutboxFailedParams{
		LastError: &message,
		Permanent: permanent,
		OutboxID:  item.ID,
	}); err != nil {
		return fmt.Errorf("mark calendar schedule event outbox failed: %w", err)
	}
	return nil
}

func (store transactionScheduleEventOutboxStore) ReleaseScheduleEventOutbox(ctx context.Context, outboxIDs []uuid.UUID) error {
	if len(outboxIDs) == 0 {
		return nil
	}
	if err := store.queries.ReleaseScheduleEventOutbox(ctx, calendarsql.ReleaseScheduleEventOutboxParams{OutboxIds: outboxIDs}); err != nil {
		return fmt.Errorf("release unprocessed calendar schedule event outbox: %w", err)
	}
	return nil
}

func (store transactionScheduleEventOutboxStore) DeleteCleanupPendingConnectionIfDrained(ctx context.Context, userID uuid.UUID, provider calendar.Provider) error {
	if err := store.queries.DeleteDrainedCleanupPendingCalendarConnection(ctx, calendarsql.DeleteDrainedCleanupPendingCalendarConnectionParams{
		UserID:   userID,
		Provider: string(provider),
	}); err != nil {
		return fmt.Errorf("delete drained calendar cleanup connection: %w", err)
	}
	return nil
}

var _ calendar.ScheduleEventOutboxStore = transactionScheduleEventOutboxStore{}
