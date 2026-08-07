package calendarrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func (r *Repo) ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreCalendarEventSummary, error) {
	const query = `
		SELECT
			ce.event_id,
			ce.connection_id,
			ce.provider,
			ce.calendar_id,
			ce.provider_event_id,
			ce.title,
			ce.location,
			ce.meeting_url,
			ce.html_link,
			ce.start_at,
			ce.end_at,
			ce.is_all_day,
			ce.start_date,
			ce.end_date,
			ce.is_private,
			ce.created_at,
			ce.updated_at
		FROM calendar_events ce
		INNER JOIN calendar_connections cc ON
			cc.connection_id = ce.connection_id
			AND cc.workspace_id = ce.workspace_id
			AND cc.user_id = ce.user_id
			AND cc.revoked_at IS NULL
			AND $5 = ANY(cc.scopes)
		WHERE ce.workspace_id = $1
			AND ce.user_id = $2
			AND ce.start_at < $4
			AND ce.end_at > $3
		ORDER BY ce.start_at ASC, ce.event_id ASC
	`
	rows := []dbCalendarEventSummary{}
	if err := r.db.SelectContext(ctx, &rows, query, workspaceID, userID, startAt, endAt, calendar.GoogleCalendarEventsReadonlyScope); err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	return toCoreCalendarEventSummaries(rows), nil
}

func (r *Repo) GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (calendar.CoreCalendarEvent, error) {
	const query = `
		SELECT
			ce.event_id,
			ce.connection_id,
			ce.workspace_id,
			ce.user_id,
			ce.provider,
			ce.calendar_id,
			ce.provider_event_id,
			ce.title,
			ce.description,
			ce.location,
			ce.meeting_url,
			ce.html_link,
			ce.organizer,
			ce.attendees,
			ce.attendees_omitted,
			ce.is_all_day,
			ce.start_date,
			ce.end_date,
			ce.start_at,
			ce.end_at,
			ce.visibility,
			ce.is_private,
			ce.source_hash,
			ce.created_at,
			ce.updated_at
		FROM calendar_events ce
		INNER JOIN calendar_connections cc ON
			cc.connection_id = ce.connection_id
			AND cc.workspace_id = ce.workspace_id
			AND cc.user_id = ce.user_id
			AND cc.revoked_at IS NULL
			AND $4 = ANY(cc.scopes)
		WHERE ce.workspace_id = $1
			AND ce.user_id = $2
			AND ce.event_id = $3
		LIMIT 1
	`
	var row dbCalendarEvent
	if err := r.db.GetContext(ctx, &row, query, workspaceID, userID, eventID, calendar.GoogleCalendarEventsReadonlyScope); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calendar.CoreCalendarEvent{}, calendar.ErrCalendarEventNotFound
		}
		return calendar.CoreCalendarEvent{}, fmt.Errorf("get calendar event: %w", err)
	}
	event, err := toCoreCalendarEvent(row)
	if err != nil {
		return calendar.CoreCalendarEvent{}, err
	}
	return event, nil
}
