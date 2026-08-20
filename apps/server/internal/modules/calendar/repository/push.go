package calendarrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

const connectionColumns = `
	connection_id, workspace_id, user_id, credential_generation, provider_account_id,
	provider, connected_email, timezone, token_payload, scopes, sync_status, sync_error,
	last_synced_at, sync_token, notification_channel_id, notification_resource_id,
	notification_expires_at, revoked_at, created_at, updated_at`

type managedScheduleBlockState struct {
	ID          uuid.UUID  `db:"block_id"`
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	StoryID     *uuid.UUID `db:"story_id"`
	Title       string     `db:"title"`
	StartAt     time.Time  `db:"start_at"`
	EndAt       time.Time  `db:"end_at"`
}

// calendar_schedule_blocks.updated_at is the concurrency token for user-visible
// schedule content. Provider mirror bookkeeping must preserve it.
const invalidateDeletedManagedScheduleEventQuery = `
	UPDATE calendar_schedule_blocks
	SET external_sync_hash = NULL, external_synced_at = NULL
	WHERE user_id = $1 AND external_provider = 'google' AND external_event_id = $2
`

const invalidateChangedManagedScheduleEventQuery = `
	UPDATE calendar_schedule_blocks
	SET external_sync_hash = NULL, external_synced_at = NULL
	WHERE block_id = $1
`

func (r *Repo) GetConnection(ctx context.Context, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	var row dbConnection
	query := `SELECT ` + connectionColumns + ` FROM calendar_connections WHERE connection_id = $1 AND revoked_at IS NULL AND cleanup_pending_at IS NULL`
	if err := r.db.GetContext(ctx, &row, query, connectionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, fmt.Errorf("get calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) ListConnectionsNeedingWatch(ctx context.Context, renewBefore time.Time) ([]calendar.CoreConnection, error) {
	rows := []dbConnection{}
	query := `SELECT ` + connectionColumns + `
		FROM calendar_connections
		WHERE revoked_at IS NULL
		  AND cleanup_pending_at IS NULL
		  AND provider = 'google'
		  AND (notification_channel_id IS NULL OR notification_expires_at IS NULL OR notification_expires_at <= $1)
		ORDER BY notification_expires_at NULLS FIRST, connection_id`
	if err := r.db.SelectContext(ctx, &rows, query, renewBefore); err != nil {
		return nil, fmt.Errorf("list calendar connections needing watch: %w", err)
	}
	return toCoreConnections(rows), nil
}

func (r *Repo) SetNotificationChannel(ctx context.Context, connection calendar.CoreConnection, channel calendar.CalendarWatchChannel) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE calendar_connections
		SET notification_channel_id = $4,
			notification_resource_id = $5,
			notification_expires_at = $6,
			updated_at = NOW()
		WHERE connection_id = $1 AND workspace_id = $2 AND credential_generation = $3 AND revoked_at IS NULL AND cleanup_pending_at IS NULL`,
		connection.ID, connection.WorkspaceID, connection.CredentialGeneration,
		channel.ChannelID, channel.ResourceID, channel.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("store calendar notification channel: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read calendar notification channel update: %w", err)
	}
	if count == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}

func (r *Repo) ClearNotificationChannel(ctx context.Context, connectionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE calendar_connections
		SET notification_channel_id = NULL, notification_resource_id = NULL,
			notification_expires_at = NULL, updated_at = NOW()
		WHERE connection_id = $1`, connectionID)
	if err != nil {
		return fmt.Errorf("clear calendar notification channel: %w", err)
	}
	return nil
}

func (r *Repo) ApplyCalendarChanges(ctx context.Context, connection calendar.CoreConnection, delta calendar.CalendarSyncDelta) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin apply calendar changes: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, connection.WorkspaceID, connection.UserID); err != nil {
		return err
	}

	var locked uuid.UUID
	if err := tx.GetContext(ctx, &locked, `
		SELECT connection_id FROM calendar_connections
			WHERE connection_id = $1 AND workspace_id = $2 AND credential_generation = $3 AND revoked_at IS NULL AND cleanup_pending_at IS NULL
		FOR UPDATE`, connection.ID, connection.WorkspaceID, connection.CredentialGeneration); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.ErrCalendarSyncSuperseded
		}
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
		if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_events WHERE connection_id = $1 AND calendar_id = 'primary' AND provider_event_id = $2`, connection.ID, eventID); err != nil {
			return fmt.Errorf("delete changed calendar event: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_busy_windows WHERE connection_id = $1 AND provider_event_id = $2`, connection.ID, eventID); err != nil {
			return fmt.Errorf("delete changed calendar busy window: %w", err)
		}
	}

	for _, event := range delta.Events {
		organizer, err := marshalOptionalCalendarParticipant(event.Organizer)
		if err != nil {
			return err
		}
		attendees, err := json.Marshal(event.Attendees)
		if err != nil {
			return fmt.Errorf("encode incremental calendar attendees: %w", err)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO calendar_events (
				connection_id, workspace_id, user_id, provider, calendar_id, provider_event_id,
				title, description, location, meeting_url, html_link, organizer, attendees,
				attendees_omitted, is_all_day, start_date, end_date, start_at, end_at,
				visibility, is_private, source_hash, sync_generation, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,CAST($12 AS jsonb),CAST($13 AS jsonb),$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,NOW())`,
			connection.ID, connection.WorkspaceID, connection.UserID, string(connection.Provider), event.CalendarID,
			event.ProviderEventID, event.Title, event.Description, event.Location, event.MeetingURL, event.HTMLLink,
			organizer, string(attendees), event.AttendeesOmitted, event.IsAllDay, event.StartDate, event.EndDate,
			event.StartAt, event.EndAt, event.Visibility, event.IsPrivate, event.SourceHash, uuid.New(),
		)
		if err != nil {
			return fmt.Errorf("insert incremental calendar event: %w", err)
		}
	}

	for _, window := range delta.BusyWindows {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO calendar_busy_windows (
				connection_id, workspace_id, user_id, provider, provider_event_id, calendar_id,
				title, start_at, end_at, status, transparency, is_private, source_hash, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,NULL,$7,$8,$9,$10,$11,$12,NOW())`,
			connection.ID, connection.WorkspaceID, connection.UserID, string(connection.Provider), window.ProviderEventID,
			window.CalendarID, window.StartAt, window.EndAt, string(window.Status), string(window.Transparency), window.IsPrivate, window.SourceHash,
		)
		if err != nil {
			return fmt.Errorf("insert incremental calendar busy window: %w", err)
		}
	}

	for _, change := range delta.ManagedScheduleEventChanges {
		if strings.TrimSpace(change.EventID) == "" {
			continue
		}
		if change.Deleted {
			if _, err := tx.ExecContext(ctx, invalidateDeletedManagedScheduleEventQuery, connection.UserID, change.EventID); err != nil {
				return fmt.Errorf("invalidate deleted managed calendar event: %w", err)
			}
			continue
		}
		var block managedScheduleBlockState
		if err := tx.GetContext(ctx, &block, `
			SELECT block_id, workspace_id, story_id, title, start_at, end_at
			FROM calendar_schedule_blocks
			WHERE user_id = $1
				AND external_provider = 'google'
				AND external_event_id = $2
			FOR UPDATE
		`, connection.UserID, change.EventID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return fmt.Errorf("lock changed managed calendar event: %w", err)
		}
		if managedScheduleEventMatchesBlock(change, block) {
			continue
		}
		if _, err := tx.ExecContext(ctx, invalidateChangedManagedScheduleEventQuery, block.ID); err != nil {
			return fmt.Errorf("invalidate changed managed calendar event: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE calendar_connections SET sync_token = NULLIF($2, ''), updated_at = NOW() WHERE connection_id = $1`, connection.ID, strings.TrimSpace(delta.NextSyncToken)); err != nil {
		return fmt.Errorf("store incremental calendar sync token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit incremental calendar changes: %w", err)
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
