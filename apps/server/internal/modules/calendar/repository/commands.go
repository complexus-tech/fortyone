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
	"github.com/lib/pq"
)

func (r *Repo) UpsertConnection(ctx context.Context, input calendar.CoreConnectionUpsert) (calendar.CoreConnection, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("begin upsert calendar connection: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, input.WorkspaceID, input.UserID); err != nil {
		return calendar.CoreConnection{}, err
	}

	var existingProviderAccountID string
	const existingQuery = `
		SELECT provider_account_id
		FROM calendar_connections
		WHERE workspace_id = $1
			AND user_id = $2
			AND provider = $3
			AND revoked_at IS NULL
		FOR UPDATE
	`
	hasExisting := true
	if err := tx.GetContext(
		ctx,
		&existingProviderAccountID,
		existingQuery,
		input.WorkspaceID,
		input.UserID,
		string(input.Provider),
	); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, fmt.Errorf("lock existing calendar connection: %w", err)
		}
		hasExisting = false
	}

	credentialGeneration := uuid.New()
	const query = `
		INSERT INTO calendar_connections (
			workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
			token_payload, scopes, sync_status, sync_error, revoked_at, updated_at
		) VALUES (
			:workspace_id, :user_id, :credential_generation, :provider_account_id, :provider, :connected_email, :timezone,
			:token_payload, :scopes, 'connected', NULL, NULL, NOW()
		)
		ON CONFLICT (workspace_id, user_id, provider)
		WHERE revoked_at IS NULL
		DO UPDATE SET
			credential_generation = EXCLUDED.credential_generation,
			provider_account_id = EXCLUDED.provider_account_id,
			connected_email = EXCLUDED.connected_email,
			timezone = EXCLUDED.timezone,
			token_payload = EXCLUDED.token_payload,
			scopes = EXCLUDED.scopes,
			sync_status = 'connected',
			sync_error = NULL,
			sync_token = NULL,
			notification_channel_id = NULL,
			notification_resource_id = NULL,
			notification_expires_at = NULL,
			last_synced_at = CASE
				WHEN calendar_connections.provider_account_id <> ''
					AND calendar_connections.provider_account_id = EXCLUDED.provider_account_id
					THEN calendar_connections.last_synced_at
				ELSE NULL
			END,
			updated_at = NOW()
		RETURNING connection_id, workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
		          token_payload, scopes, sync_status, sync_error, last_synced_at, sync_token,
		          notification_channel_id, notification_resource_id, notification_expires_at,
		          revoked_at, created_at, updated_at
	`
	params := map[string]any{
		"workspace_id":          input.WorkspaceID,
		"user_id":               input.UserID,
		"credential_generation": credentialGeneration,
		"provider_account_id":   input.ProviderAccountID,
		"provider":              string(input.Provider),
		"connected_email":       input.ConnectedEmail,
		"timezone":              input.Timezone,
		"token_payload":         input.TokenPayload,
		"scopes":                pq.StringArray(input.Scopes),
	}
	var row dbConnection
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("prepare upsert calendar connection: %w", err)
	}
	defer stmt.Close()
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("upsert calendar connection: %w", err)
	}
	if hasExisting && existingProviderAccountID != input.ProviderAccountID {
		if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_events WHERE connection_id = $1`, row.ID); err != nil {
			return calendar.CoreConnection{}, fmt.Errorf("clear calendar events after account change: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_busy_windows WHERE connection_id = $1`, row.ID); err != nil {
			return calendar.CoreConnection{}, fmt.Errorf("clear calendar availability after account change: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("commit upsert calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

// BeginConnectionSync rotates the generation before provider I/O. A later sync
// attempt therefore supersedes every earlier in-flight attempt, preventing an
// older provider response from replacing a newer snapshot.
func (r *Repo) BeginConnectionSync(ctx context.Context, connection calendar.CoreConnection) (calendar.CoreConnection, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("begin calendar sync: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, connection.WorkspaceID, connection.UserID); err != nil {
		return calendar.CoreConnection{}, err
	}

	const query = `
		UPDATE calendar_connections
		SET credential_generation = $6,
			updated_at = NOW()
		WHERE connection_id = $1
			AND workspace_id = $2
			AND user_id = $3
			AND provider = $4
			AND credential_generation = $5
			AND revoked_at IS NULL
		RETURNING connection_id, workspace_id, user_id, credential_generation, provider_account_id,
		          provider, connected_email, timezone, token_payload, scopes, sync_status,
		          sync_error, last_synced_at, sync_token, notification_channel_id,
		          notification_resource_id, notification_expires_at, revoked_at, created_at, updated_at
	`
	var row dbConnection
	if err := tx.GetContext(
		ctx,
		&row,
		query,
		connection.ID,
		connection.WorkspaceID,
		connection.UserID,
		string(connection.Provider),
		connection.CredentialGeneration,
		uuid.New(),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarSyncSuperseded
		}
		return calendar.CoreConnection{}, fmt.Errorf("start calendar sync: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("commit calendar sync start: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin revoke calendar connection: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const revokeQuery = `
		UPDATE calendar_connections
		SET revoked_at = NOW(),
			sync_status = 'revoked',
			provider_account_id = '',
			token_payload = '',
			scopes = '{}',
			sync_error = NULL,
			sync_token = NULL,
			notification_channel_id = NULL,
			notification_resource_id = NULL,
			notification_expires_at = NULL,
			updated_at = NOW()
		WHERE workspace_id = $1
			AND user_id = $2
			AND connection_id = $3
			AND revoked_at IS NULL
	`
	result, err := tx.ExecContext(ctx, revokeQuery, workspaceID, userID, connectionID)
	if err != nil {
		return fmt.Errorf("revoke calendar connection: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read revoked calendar connection result: %w", err)
	}
	if affected == 0 {
		return calendar.ErrCalendarNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_events WHERE connection_id = $1`, connectionID); err != nil {
		return fmt.Errorf("delete calendar events for revoked connection: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_busy_windows WHERE connection_id = $1`, connectionID); err != nil {
		return fmt.Errorf("delete calendar busy windows for revoked connection: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit revoked calendar connection: %w", err)
	}
	return nil
}

func (r *Repo) ReplaceCalendarSnapshot(ctx context.Context, connection calendar.CoreConnection, snapshot calendar.CalendarSyncSnapshot) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace calendar snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := lockCalendarUser(ctx, tx, connection.WorkspaceID, connection.UserID); err != nil {
		return err
	}

	const lockConnectionQuery = `
		SELECT connection_id
		FROM calendar_connections
		WHERE connection_id = $1
			AND workspace_id = $2
			AND user_id = $3
			AND provider = $4
			AND credential_generation = $5
			AND revoked_at IS NULL
		FOR UPDATE
	`
	var lockedConnectionID uuid.UUID
	if err := tx.GetContext(
		ctx,
		&lockedConnectionID,
		lockConnectionQuery,
		connection.ID,
		connection.WorkspaceID,
		connection.UserID,
		string(connection.Provider),
		connection.CredentialGeneration,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.ErrCalendarSyncSuperseded
		}
		return fmt.Errorf("lock active calendar connection: %w", err)
	}
	if connection.CanReadEventDetails() && !snapshot.CanReadEventDetails {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE calendar_connections SET scopes = array_remove(scopes, $2), updated_at = NOW() WHERE connection_id = $1`,
			connection.ID,
			calendar.GoogleCalendarEventsReadonlyScope,
		); err != nil {
			return fmt.Errorf("downgrade calendar event detail access: %w", err)
		}
	}

	syncGeneration := uuid.New()
	if len(snapshot.Events) > 0 {
		const eventQuery = `
			INSERT INTO calendar_events (
				connection_id, workspace_id, user_id, provider, calendar_id,
				provider_event_id, title, description, location, meeting_url,
				html_link, organizer, attendees, attendees_omitted, is_all_day,
				start_date, end_date, start_at, end_at, visibility, is_private, source_hash,
				sync_generation, updated_at
			) VALUES (
				:connection_id, :workspace_id, :user_id, :provider, :calendar_id,
				:provider_event_id, :title, :description, :location, :meeting_url,
				:html_link, CAST(:organizer AS jsonb), CAST(:attendees AS jsonb),
				:attendees_omitted, :is_all_day, :start_date, :end_date,
				:start_at, :end_at, :visibility,
				:is_private, :source_hash, :sync_generation, NOW()
			)
			ON CONFLICT (connection_id, calendar_id, provider_event_id)
			DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				location = EXCLUDED.location,
				meeting_url = EXCLUDED.meeting_url,
				html_link = EXCLUDED.html_link,
				organizer = EXCLUDED.organizer,
				attendees = EXCLUDED.attendees,
				attendees_omitted = EXCLUDED.attendees_omitted,
				is_all_day = EXCLUDED.is_all_day,
				start_date = EXCLUDED.start_date,
				end_date = EXCLUDED.end_date,
				start_at = EXCLUDED.start_at,
				end_at = EXCLUDED.end_at,
				visibility = EXCLUDED.visibility,
				is_private = EXCLUDED.is_private,
				source_hash = EXCLUDED.source_hash,
				sync_generation = EXCLUDED.sync_generation,
				updated_at = NOW()
		`
		stmt, err := tx.PrepareNamedContext(ctx, eventQuery)
		if err != nil {
			return fmt.Errorf("prepare upsert calendar events: %w", err)
		}
		defer stmt.Close()
		for _, event := range snapshot.Events {
			organizer, err := marshalOptionalCalendarParticipant(event.Organizer)
			if err != nil {
				return err
			}
			attendees, err := json.Marshal(event.Attendees)
			if err != nil {
				return fmt.Errorf("encode calendar event attendees: %w", err)
			}
			params := map[string]any{
				"connection_id":     connection.ID,
				"workspace_id":      connection.WorkspaceID,
				"user_id":           connection.UserID,
				"provider":          string(connection.Provider),
				"calendar_id":       event.CalendarID,
				"provider_event_id": event.ProviderEventID,
				"title":             event.Title,
				"description":       event.Description,
				"location":          event.Location,
				"meeting_url":       event.MeetingURL,
				"html_link":         event.HTMLLink,
				"organizer":         organizer,
				"attendees":         string(attendees),
				"attendees_omitted": event.AttendeesOmitted,
				"is_all_day":        event.IsAllDay,
				"start_date":        event.StartDate,
				"end_date":          event.EndDate,
				"start_at":          event.StartAt,
				"end_at":            event.EndAt,
				"visibility":        event.Visibility,
				"is_private":        event.IsPrivate,
				"source_hash":       event.SourceHash,
				"sync_generation":   syncGeneration,
			}
			if _, err := stmt.ExecContext(ctx, params); err != nil {
				return fmt.Errorf("upsert calendar event: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM calendar_events WHERE connection_id = $1 AND sync_generation <> $2`,
		connection.ID,
		syncGeneration,
	); err != nil {
		return fmt.Errorf("delete stale calendar events: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM calendar_busy_windows WHERE connection_id = $1`, connection.ID); err != nil {
		return fmt.Errorf("delete existing calendar busy windows: %w", err)
	}

	if len(snapshot.BusyWindows) > 0 {
		const insertQuery = `
			INSERT INTO calendar_busy_windows (
				connection_id, workspace_id, user_id, provider, provider_event_id,
				calendar_id, title, start_at, end_at, status, transparency,
				is_private, source_hash, updated_at
			) VALUES (
				:connection_id, :workspace_id, :user_id, :provider, :provider_event_id,
				:calendar_id, :title, :start_at, :end_at, :status, :transparency,
				:is_private, :source_hash, NOW()
			)
		`
		stmt, prepareErr := tx.PrepareNamedContext(ctx, insertQuery)
		if prepareErr != nil {
			err = prepareErr
			return fmt.Errorf("prepare insert calendar busy windows: %w", err)
		}
		defer stmt.Close()
		for _, window := range snapshot.BusyWindows {
			params := map[string]any{
				"connection_id":     connection.ID,
				"workspace_id":      connection.WorkspaceID,
				"user_id":           connection.UserID,
				"provider":          string(connection.Provider),
				"provider_event_id": window.ProviderEventID,
				"calendar_id":       window.CalendarID,
				"title":             nil,
				"start_at":          window.StartAt,
				"end_at":            window.EndAt,
				"status":            string(window.Status),
				"transparency":      string(window.Transparency),
				"is_private":        window.IsPrivate,
				"source_hash":       window.SourceHash,
			}
			if _, execErr := stmt.ExecContext(ctx, params); execErr != nil {
				err = execErr
				return fmt.Errorf("insert calendar busy window: %w", err)
			}
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE calendar_connections SET sync_token = NULLIF($2, ''), updated_at = NOW() WHERE connection_id = $1`,
		connection.ID,
		strings.TrimSpace(snapshot.NextSyncToken),
	); err != nil {
		return fmt.Errorf("store calendar sync token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit replace calendar snapshot: %w", err)
	}
	return nil
}

func marshalOptionalCalendarParticipant(participant *calendar.CoreCalendarParticipant) (any, error) {
	if participant == nil {
		return nil, nil
	}
	value, err := json.Marshal(participant)
	if err != nil {
		return nil, fmt.Errorf("encode calendar event organizer: %w", err)
	}
	return string(value), nil
}

func (r *Repo) MarkConnectionSynced(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	syncedAt time.Time,
) error {
	const query = `
		UPDATE calendar_connections
		SET sync_status = 'synced',
			sync_error = NULL,
			last_synced_at = $4,
			updated_at = NOW()
		WHERE workspace_id = $1
			AND connection_id = $2
			AND credential_generation = $3
			AND revoked_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, workspaceID, connectionID, credentialGeneration, syncedAt)
	if err != nil {
		return fmt.Errorf("mark calendar connection synced: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read synced calendar connection count: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}

func (r *Repo) MarkConnectionSyncFailed(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	message string,
) error {
	const query = `
		UPDATE calendar_connections
		SET sync_status = 'failed',
			sync_error = $4,
			updated_at = NOW()
		WHERE workspace_id = $1
			AND connection_id = $2
			AND credential_generation = $3
			AND revoked_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, workspaceID, connectionID, credentialGeneration, message)
	if err != nil {
		return fmt.Errorf("mark calendar connection sync failed: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read failed calendar connection count: %w", err)
	}
	if rows == 0 {
		return calendar.ErrCalendarSyncSuperseded
	}
	return nil
}
