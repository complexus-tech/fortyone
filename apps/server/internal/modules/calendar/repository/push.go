package calendarrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type managedScheduleBlockState struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	StoryID     *uuid.UUID
	Title       string
	StartAt     time.Time
	EndAt       time.Time
}

func (r *Repo) SetNotificationChannel(ctx context.Context, connection calendar.CoreConnection, channel calendar.CalendarWatchChannel) error {
	if err := r.configured(); err != nil {
		return err
	}
	rows, err := r.queries.SetCalendarNotificationChannel(ctx, calendarsql.SetCalendarNotificationChannelParams{
		ChannelID:            &channel.ChannelID,
		ResourceID:           &channel.ResourceID,
		ExpiresAt:            &channel.ExpiresAt,
		ConnectionID:         connection.ID,
		WorkspaceID:          connection.WorkspaceID,
		CredentialGeneration: connection.CredentialGeneration,
	})
	if err != nil {
		return fmt.Errorf("store calendar notification channel: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}

func (r *Repo) ClearNotificationChannel(ctx context.Context, connectionID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.queries.ClearCalendarNotificationChannel(ctx, calendarsql.ClearCalendarNotificationChannelParams{ConnectionID: connectionID}); err != nil {
		return fmt.Errorf("clear calendar notification channel: %w", err)
	}
	return nil
}

func (r *Repo) ApplyCalendarChanges(ctx context.Context, connection calendar.CoreConnection, delta calendar.CalendarSyncDelta) error {
	if err := r.configured(); err != nil {
		return err
	}
	if err := r.withinTransaction(ctx, func(queries calendarsql.Querier) error {
		if err := lockCalendarUser(ctx, queries, connection.UserID); err != nil {
			return err
		}
		_, err := queries.LockCalendarConnectionForIncrementalSync(ctx, calendarsql.LockCalendarConnectionForIncrementalSyncParams{
			ConnectionID:         connection.ID,
			WorkspaceID:          connection.WorkspaceID,
			CredentialGeneration: connection.CredentialGeneration,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return calendar.ErrCalendarSyncSuperseded
		}
		if err != nil {
			return fmt.Errorf("lock calendar connection for incremental sync: %w", err)
		}

		changedIDs := append([]string{}, delta.DeletedEventIDs...)
		for _, event := range delta.Events {
			changedIDs = append(changedIDs, event.ProviderEventID)
		}
		for _, eventID := range changedIDs {
			eventID = strings.TrimSpace(eventID)
			if eventID == "" {
				continue
			}
			if err := queries.DeleteChangedCalendarEvent(ctx, calendarsql.DeleteChangedCalendarEventParams{
				ConnectionID:    connection.ID,
				ProviderEventID: eventID,
			}); err != nil {
				return fmt.Errorf("delete changed calendar event: %w", err)
			}
			if err := queries.DeleteChangedCalendarBusyWindow(ctx, calendarsql.DeleteChangedCalendarBusyWindowParams{
				ConnectionID:    connection.ID,
				ProviderEventID: eventID,
			}); err != nil {
				return fmt.Errorf("delete changed calendar busy window: %w", err)
			}
		}

		for _, event := range delta.Events {
			params, err := calendarEventUpsertParams(connection, event, uuid.New())
			if err != nil {
				return err
			}
			if err := queries.UpsertCalendarEvent(ctx, params); err != nil {
				return fmt.Errorf("insert incremental calendar event: %w", err)
			}
		}
		for _, window := range delta.BusyWindows {
			if err := queries.InsertCalendarBusyWindow(ctx, calendarBusyWindowParams(connection, window)); err != nil {
				return fmt.Errorf("insert incremental calendar busy window: %w", err)
			}
		}

		for _, change := range delta.ManagedScheduleEventChanges {
			if strings.TrimSpace(change.EventID) == "" {
				continue
			}
			if change.Deleted {
				if err := queries.InvalidateDeletedManagedScheduleEvent(ctx, calendarsql.InvalidateDeletedManagedScheduleEventParams{
					UserID:   connection.UserID,
					Provider: string(connection.Provider),
					EventID:  change.EventID,
				}); err != nil {
					return fmt.Errorf("invalidate deleted managed calendar event: %w", err)
				}
				continue
			}
			row, err := queries.LockChangedManagedScheduleEvent(ctx, calendarsql.LockChangedManagedScheduleEventParams{
				UserID:   connection.UserID,
				Provider: string(connection.Provider),
				EventID:  change.EventID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return fmt.Errorf("lock changed managed calendar event: %w", err)
			}
			block := managedScheduleBlockState{
				ID: row.BlockID, WorkspaceID: row.WorkspaceID, StoryID: row.StoryID,
				Title: row.Title, StartAt: row.StartAt, EndAt: row.EndAt,
			}
			if managedScheduleEventMatchesBlock(change, block) {
				continue
			}
			if err := queries.InvalidateChangedManagedScheduleEvent(ctx, calendarsql.InvalidateChangedManagedScheduleEventParams{BlockID: block.ID}); err != nil {
				return fmt.Errorf("invalidate changed managed calendar event: %w", err)
			}
		}
		if err := queries.StoreIncrementalCalendarSyncToken(ctx, calendarsql.StoreIncrementalCalendarSyncTokenParams{
			SyncToken:    strings.TrimSpace(delta.NextSyncToken),
			ConnectionID: connection.ID,
		}); err != nil {
			return fmt.Errorf("store incremental calendar sync token: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("apply calendar changes transaction: %w", err)
	}
	return nil
}

func managedScheduleEventMatchesBlock(change calendar.ManagedScheduleEventChange, block managedScheduleBlockState) bool {
	if change.Deleted || change.Visibility != "private" || change.Transparency != "opaque" || change.Status != "confirmed" || change.HasAttendees || change.Recurring {
		return false
	}
	if change.Source != "maya_schedule" || change.BlockID != block.ID.String() || change.WorkspaceID != block.WorkspaceID.String() {
		return false
	}
	storyID := ""
	if block.StoryID != nil {
		storyID = block.StoryID.String()
	}
	return change.StoryID == storyID && change.Title == block.Title && change.StartAt.Equal(block.StartAt) && change.EndAt.Equal(block.EndAt)
}
