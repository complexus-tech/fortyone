package calendarrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func (r *Repo) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]calendar.CoreConnection, error) {
	query := `
		SELECT connection_id, workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, sync_token,
		       notification_channel_id, notification_resource_id, notification_expires_at,
		       revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE revoked_at IS NULL
			AND cleanup_pending_at IS NULL
	`
	params := map[string]any{"workspace_id": workspaceID}
	if userID != nil {
		query += " AND user_id = :user_id"
		params["user_id"] = *userID
	} else {
		query += " AND workspace_id = :workspace_id"
	}
	query += " ORDER BY created_at DESC"

	rows := []dbConnection{}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare list calendar connections: %w", err)
	}
	defer stmt.Close()
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		return nil, fmt.Errorf("list calendar connections: %w", err)
	}
	return toCoreConnections(rows), nil
}

func (r *Repo) GetOwnedConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	const query = `
		SELECT connection_id, workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, sync_token,
		       notification_channel_id, notification_resource_id, notification_expires_at,
		       revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE user_id = $1
			AND connection_id = $2
			AND revoked_at IS NULL
			AND cleanup_pending_at IS NULL
		LIMIT 1
	`
	var row dbConnection
	if err := r.db.GetContext(ctx, &row, query, userID, connectionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, fmt.Errorf("get owned calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM workspace_members wm
			INNER JOIN users u ON u.user_id = wm.user_id
			WHERE wm.workspace_id = $1
				AND wm.user_id = $2
				AND u.is_active = true
		)
	`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, workspaceID, userID); err != nil {
		return false, fmt.Errorf("check calendar workspace membership: %w", err)
	}
	return exists, nil
}

func (r *Repo) GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider calendar.Provider) (calendar.CoreConnection, error) {
	const query = `
		SELECT connection_id, workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, sync_token,
		       notification_channel_id, notification_resource_id, notification_expires_at,
		       revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE user_id = $1
			AND provider = $2
			AND revoked_at IS NULL
			AND cleanup_pending_at IS NULL
		LIMIT 1
	`
	var row dbConnection
	if err := r.db.GetContext(ctx, &row, query, userID, string(provider)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, fmt.Errorf("get active calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) GetScheduleEventDispatchConnection(ctx context.Context, userID uuid.UUID) (calendar.CoreConnection, bool, error) {
	const query = `
		SELECT connection_id, workspace_id, user_id, credential_generation, provider_account_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, sync_token,
		       notification_channel_id, notification_resource_id, notification_expires_at,
		       revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE user_id = $1
			AND provider = 'google'
			AND revoked_at IS NULL
		ORDER BY cleanup_pending_at NULLS FIRST
		LIMIT 1
	`
	var row dbConnection
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, false, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, false, fmt.Errorf("get calendar schedule dispatch connection: %w", err)
	}
	var cleanupPending bool
	if err := r.db.GetContext(ctx, &cleanupPending, `
		SELECT cleanup_pending_at IS NOT NULL
		FROM calendar_connections
		WHERE connection_id = $1
	`, row.ID); err != nil {
		return calendar.CoreConnection{}, false, fmt.Errorf("read calendar schedule dispatch cleanup state: %w", err)
	}
	return toCoreConnection(row), cleanupPending, nil
}
