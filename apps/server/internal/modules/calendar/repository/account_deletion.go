package calendarrepository

import (
	"context"
	"errors"
	"fmt"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// StageAccountDeletion retains sealed credentials only while the existing
// dispatcher removes FortyOne-created provider events. The account deletion
// transaction must keep its inaccessible user row until PendingAccountDeletion
// returns false; otherwise the user FK would cascade away the cleanup work.
func (r *Repo) StageAccountDeletion(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	if tx == nil || userID == uuid.Nil {
		return errors.New("account calendar deletion transaction and user are required")
	}
	return stageAccountDeletion(ctx, calendarsql.New(tx), userID)
}

func (r *Repo) PendingAccountDeletion(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	if tx == nil || userID == uuid.Nil {
		return false, errors.New("account calendar deletion transaction and user are required")
	}
	return calendarsql.New(tx).AccountCalendarCleanupPending(ctx, calendarsql.AccountCalendarCleanupPendingParams{UserID: userID})
}

func stageAccountDeletion(ctx context.Context, queries calendarsql.Querier, userID uuid.UUID) error {
	if err := lockCalendarUser(ctx, queries, userID); err != nil {
		return err
	}
	connections, err := queries.ListCalendarConnectionsByUser(ctx, calendarsql.ListCalendarConnectionsByUserParams{UserID: userID})
	if err != nil {
		return fmt.Errorf("list account calendars for deletion: %w", err)
	}
	for _, connection := range connections {
		if _, err := queries.MarkCalendarConnectionCleanupPending(ctx, calendarsql.MarkCalendarConnectionCleanupPendingParams{UserID: userID, ConnectionID: connection.ConnectionID}); err != nil {
			return fmt.Errorf("stage account calendar cleanup: %w", err)
		}
		if err := queries.ReactivateCalendarOutboxForCleanup(ctx, calendarsql.ReactivateCalendarOutboxForCleanupParams{UserID: userID, Provider: connection.Provider}); err != nil {
			return fmt.Errorf("reactivate account calendar cleanup: %w", err)
		}
		mappings, err := queries.ListMayaScheduleMirrorsForCleanup(ctx, calendarsql.ListMayaScheduleMirrorsForCleanupParams{UserID: userID, Provider: connection.Provider})
		if err != nil {
			return fmt.Errorf("list account calendar mirrors: %w", err)
		}
		for _, mapping := range mappings {
			event := cleanupScheduleEvent(mapping, connection.Provider)
			if err := enqueueScheduleEventOutbox(ctx, queries, mapping.WorkspaceID, userID, &mapping.BlockID, calendar.Provider(connection.Provider), calendar.ScheduleEventOperationDelete, event, "", true); err != nil {
				return err
			}
		}
		if err := queries.DetachMayaScheduleMirrors(ctx, calendarsql.DetachMayaScheduleMirrorsParams{UserID: userID, Provider: connection.Provider}); err != nil {
			return fmt.Errorf("detach account calendar mirrors: %w", err)
		}
		if err := queries.DeleteCalendarConnectionEvents(ctx, calendarsql.DeleteCalendarConnectionEventsParams{ConnectionID: connection.ConnectionID}); err != nil {
			return fmt.Errorf("delete account calendar cache: %w", err)
		}
		if err := queries.DeleteCalendarConnectionBusyWindows(ctx, calendarsql.DeleteCalendarConnectionBusyWindowsParams{ConnectionID: connection.ConnectionID}); err != nil {
			return fmt.Errorf("delete account calendar availability: %w", err)
		}
		if err := queries.DeleteDrainedCalendarConnection(ctx, calendarsql.DeleteDrainedCalendarConnectionParams{ConnectionID: connection.ConnectionID}); err != nil {
			return fmt.Errorf("delete drained account calendar: %w", err)
		}
	}
	if err := queries.DeleteRevokedAccountCalendarConnections(ctx, calendarsql.DeleteRevokedAccountCalendarConnectionsParams{UserID: userID}); err != nil {
		return fmt.Errorf("delete revoked account calendars: %w", err)
	}
	if err := queries.ScrubAccountCalendarCleanup(ctx, calendarsql.ScrubAccountCalendarCleanupParams{UserID: userID}); err != nil {
		return fmt.Errorf("scrub account calendar metadata: %w", err)
	}
	if err := queries.ScrubAccountCalendarOutbox(ctx, calendarsql.ScrubAccountCalendarOutboxParams{UserID: userID}); err != nil {
		return fmt.Errorf("scrub account calendar cleanup payloads: %w", err)
	}
	return nil
}
