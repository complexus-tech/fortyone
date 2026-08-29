package calendarrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	var (
		rows []calendarsql.CalendarConnection
		err  error
	)
	if userID != nil {
		rows, err = r.queries.ListCalendarConnectionsByUser(ctx, calendarsql.ListCalendarConnectionsByUserParams{UserID: *userID})
	} else {
		rows, err = r.queries.ListCalendarConnectionsByWorkspace(ctx, calendarsql.ListCalendarConnectionsByWorkspaceParams{WorkspaceID: workspaceID})
	}
	if err != nil {
		return nil, fmt.Errorf("list calendar connections: %w", err)
	}
	return toCoreConnections(rows), nil
}

func (r *Repo) GetOwnedConnection(ctx context.Context, _ uuid.UUID, userID, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}
	row, err := r.queries.GetOwnedCalendarConnection(ctx, calendarsql.GetOwnedCalendarConnectionParams{
		UserID:       userID,
		ConnectionID: connectionID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
	}
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("get owned calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	exists, err := r.queries.WorkspaceCalendarMemberExists(ctx, calendarsql.WorkspaceCalendarMemberExistsParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return false, fmt.Errorf("check calendar workspace membership: %w", err)
	}
	return exists, nil
}

func (r *Repo) GetActiveConnection(ctx context.Context, _ uuid.UUID, userID uuid.UUID, provider calendar.Provider) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}
	row, err := r.queries.GetActiveCalendarConnection(ctx, calendarsql.GetActiveCalendarConnectionParams{
		UserID:   userID,
		Provider: string(provider),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
	}
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("get active calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) GetConnection(ctx context.Context, connectionID uuid.UUID) (calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, err
	}
	row, err := r.queries.GetCalendarConnection(ctx, calendarsql.GetCalendarConnectionParams{ConnectionID: connectionID})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreConnection{}, calendar.ErrCalendarNotFound
	}
	if err != nil {
		return calendar.CoreConnection{}, fmt.Errorf("get calendar connection: %w", err)
	}
	return toCoreConnection(row), nil
}

func (r *Repo) GetScheduleEventDispatchConnection(ctx context.Context, userID uuid.UUID) (calendar.CoreConnection, bool, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreConnection{}, false, err
	}
	row, err := r.queries.GetScheduleEventDispatchConnection(ctx, calendarsql.GetScheduleEventDispatchConnectionParams{UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreConnection{}, false, calendar.ErrCalendarNotFound
	}
	if err != nil {
		return calendar.CoreConnection{}, false, fmt.Errorf("get calendar schedule dispatch connection: %w", err)
	}
	return toCoreConnection(dispatchConnection(row)), row.CleanupPending, nil
}

func (r *Repo) ListConnectionsNeedingWatch(ctx context.Context, renewBefore time.Time) ([]calendar.CoreConnection, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCalendarConnectionsNeedingWatch(ctx, calendarsql.ListCalendarConnectionsNeedingWatchParams{RenewBefore: &renewBefore})
	if err != nil {
		return nil, fmt.Errorf("list calendar connections needing watch: %w", err)
	}
	return toCoreConnections(rows), nil
}

func dispatchConnection(row calendarsql.GetScheduleEventDispatchConnectionRow) calendarsql.CalendarConnection {
	return calendarsql.CalendarConnection{
		ConnectionID:           row.ConnectionID,
		WorkspaceID:            row.WorkspaceID,
		UserID:                 row.UserID,
		Provider:               row.Provider,
		ConnectedEmail:         row.ConnectedEmail,
		Timezone:               row.Timezone,
		TokenPayload:           row.TokenPayload,
		Scopes:                 row.Scopes,
		SyncStatus:             row.SyncStatus,
		SyncError:              row.SyncError,
		LastSyncedAt:           row.LastSyncedAt,
		RevokedAt:              row.RevokedAt,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
		ProviderAccountID:      row.ProviderAccountID,
		CredentialGeneration:   row.CredentialGeneration,
		SyncToken:              row.SyncToken,
		NotificationChannelID:  row.NotificationChannelID,
		NotificationResourceID: row.NotificationResourceID,
		NotificationExpiresAt:  row.NotificationExpiresAt,
		CleanupPendingAt:       row.CleanupPendingAt,
		IsPrimary:              row.IsPrimary,
	}
}
