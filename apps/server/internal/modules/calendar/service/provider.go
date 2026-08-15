package calendar

import "context"

type CalendarProvider interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (ProviderToken, error)
	SyncCalendar(ctx context.Context, token ProviderToken, input BusyWindowInput) (CalendarSyncSnapshot, error)
	SyncCalendarChanges(ctx context.Context, token ProviderToken, syncToken string) (CalendarSyncDelta, error)
	WatchCalendar(ctx context.Context, token ProviderToken, input CalendarWatchInput) (CalendarWatchChannel, error)
	StopCalendarWatch(ctx context.Context, token ProviderToken, channel CalendarWatchChannel) error
}

type CalendarEventWriter interface {
	UpsertScheduleEvent(ctx context.Context, token ProviderToken, input ExternalScheduleEventInput) error
	DeleteScheduleEvent(ctx context.Context, token ProviderToken, calendarID, eventID string) error
}
