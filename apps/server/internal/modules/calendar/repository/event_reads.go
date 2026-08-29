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

// Calendar event data belongs to the account user and is intentionally visible
// across every workspace in which that user is active.
func (r *Repo) ListCalendarEvents(ctx context.Context, _ uuid.UUID, userID uuid.UUID, startAt, endAt time.Time) ([]calendar.CoreCalendarEventSummary, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListCalendarEvents(ctx, calendarsql.ListCalendarEventsParams{
		GoogleProvider:     string(calendar.ProviderGoogle),
		GoogleReadScope:    calendar.GoogleCalendarEventsReadonlyScope,
		MicrosoftProvider:  string(calendar.ProviderMicrosoft),
		MicrosoftReadScope: calendar.MicrosoftCalendarReadWriteScope,
		UserID:             userID,
		StartAt:            startAt,
		EndAt:              endAt,
	})
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	return toCoreCalendarEventSummaries(rows), nil
}

func (r *Repo) GetCalendarEvent(ctx context.Context, _ uuid.UUID, userID, eventID uuid.UUID) (calendar.CoreCalendarEvent, error) {
	if err := r.configured(); err != nil {
		return calendar.CoreCalendarEvent{}, err
	}
	row, err := r.queries.GetCalendarEvent(ctx, calendarsql.GetCalendarEventParams{
		GoogleProvider:     string(calendar.ProviderGoogle),
		GoogleReadScope:    calendar.GoogleCalendarEventsReadonlyScope,
		MicrosoftProvider:  string(calendar.ProviderMicrosoft),
		MicrosoftReadScope: calendar.MicrosoftCalendarReadWriteScope,
		UserID:             userID,
		EventID:            eventID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return calendar.CoreCalendarEvent{}, calendar.ErrCalendarEventNotFound
	}
	if err != nil {
		return calendar.CoreCalendarEvent{}, fmt.Errorf("get calendar event: %w", err)
	}
	event, err := toCoreCalendarEvent(row)
	if err != nil {
		return calendar.CoreCalendarEvent{}, err
	}
	return event, nil
}
