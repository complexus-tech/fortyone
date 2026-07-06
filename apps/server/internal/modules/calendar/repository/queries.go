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
		SELECT connection_id, workspace_id, user_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE workspace_id = :workspace_id
			AND revoked_at IS NULL
	`
	params := map[string]any{"workspace_id": workspaceID}
	if userID != nil {
		query += " AND user_id = :user_id"
		params["user_id"] = *userID
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

func (r *Repo) GetConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	const query = `
		SELECT connection_id, workspace_id, user_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE workspace_id = $1
			AND connection_id = $2
			AND revoked_at IS NULL
		LIMIT 1
	`
	var row dbConnection
	if err := r.db.GetContext(ctx, &row, query, workspaceID, connectionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, fmt.Errorf("get calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider calendar.Provider) (calendar.CoreConnection, error) {
	const query = `
		SELECT connection_id, workspace_id, user_id, provider, connected_email, timezone,
		       token_payload, scopes, sync_status, sync_error, last_synced_at, revoked_at, created_at, updated_at
		FROM calendar_connections
		WHERE workspace_id = $1
			AND user_id = $2
			AND provider = $3
			AND revoked_at IS NULL
		LIMIT 1
	`
	var row dbConnection
	if err := r.db.GetContext(ctx, &row, query, workspaceID, userID, string(provider)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
		}
		return calendar.CoreConnection{}, fmt.Errorf("get active calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}
