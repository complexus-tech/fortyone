package calendarrepository

import (
	"context"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (r *Repo) UpsertConnection(ctx context.Context, input calendar.CoreConnectionUpsert) (calendar.CoreConnection, error) {
	const query = `
		INSERT INTO calendar_connections (
			workspace_id, user_id, provider, connected_email, timezone,
			token_payload, scopes, sync_status, sync_error, revoked_at, updated_at
		) VALUES (
			:workspace_id, :user_id, :provider, :connected_email, :timezone,
			:token_payload, :scopes, 'connected', NULL, NULL, NOW()
		)
		ON CONFLICT (workspace_id, user_id, provider)
		WHERE revoked_at IS NULL
		DO UPDATE SET
			connected_email = EXCLUDED.connected_email,
			timezone = EXCLUDED.timezone,
			token_payload = EXCLUDED.token_payload,
			scopes = EXCLUDED.scopes,
			sync_status = 'connected',
			sync_error = NULL,
			updated_at = NOW()
		RETURNING connection_id, workspace_id, user_id, provider, connected_email, timezone,
		          token_payload, scopes, sync_status, sync_error, last_synced_at, revoked_at, created_at, updated_at
	`
	params := map[string]any{
		"workspace_id":    input.WorkspaceID,
		"user_id":         input.UserID,
		"provider":        string(input.Provider),
		"connected_email": input.ConnectedEmail,
		"timezone":        input.Timezone,
		"token_payload":   input.TokenPayload,
		"scopes":          pq.StringArray(input.Scopes),
	}
	var row dbConnection
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("prepare upsert calendar connection: %w", err)
	}
	defer stmt.Close()
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("upsert calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	const query = `
		UPDATE calendar_connections
		SET revoked_at = NOW(),
			sync_status = 'revoked',
			updated_at = NOW()
		WHERE workspace_id = $1
			AND user_id = $2
			AND connection_id = $3
			AND revoked_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, workspaceID, userID, connectionID)
	if err != nil {
		return fmt.Errorf("revoke calendar connection: %w", err)
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return calendar.ErrCalendarNotFound
	}
	return nil
}

func (r *Repo) ReplaceBusyWindows(ctx context.Context, connection calendar.CoreConnection, windows []calendar.CoreBusyWindow) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace calendar busy windows: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM calendar_busy_windows WHERE connection_id = $1`, connection.ID); err != nil {
		return fmt.Errorf("delete existing calendar busy windows: %w", err)
	}

	if len(windows) > 0 {
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
		for _, window := range windows {
			params := map[string]any{
				"connection_id":     connection.ID,
				"workspace_id":      connection.WorkspaceID,
				"user_id":           connection.UserID,
				"provider":          string(connection.Provider),
				"provider_event_id": window.ProviderEventID,
				"calendar_id":       window.CalendarID,
				"title":             window.Title,
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

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit replace calendar busy windows: %w", err)
	}
	return nil
}

func (r *Repo) MarkConnectionSynced(ctx context.Context, workspaceID, connectionID uuid.UUID, syncedAt time.Time) error {
	const query = `
		UPDATE calendar_connections
		SET sync_status = 'synced',
			sync_error = NULL,
			last_synced_at = $3,
			updated_at = NOW()
		WHERE workspace_id = $1
			AND connection_id = $2
			AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, workspaceID, connectionID, syncedAt)
	return err
}

func (r *Repo) MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID uuid.UUID, message string) error {
	const query = `
		UPDATE calendar_connections
		SET sync_status = 'failed',
			sync_error = $3,
			updated_at = NOW()
		WHERE workspace_id = $1
			AND connection_id = $2
			AND revoked_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, workspaceID, connectionID, message)
	return err
}
