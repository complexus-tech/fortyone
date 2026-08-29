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

type CalendarWatchRenewer interface {
	RenewCalendarWatch(ctx context.Context, token ProviderToken, channel CalendarWatchChannel, input CalendarWatchInput) (CalendarWatchChannel, error)
}

type CalendarTokenRefresher interface {
	RefreshToken(ctx context.Context, token ProviderToken) (ProviderToken, error)
}

type CalendarEventWriter interface {
	UpsertScheduleEvent(ctx context.Context, token ProviderToken, input ExternalScheduleEventInput) (ExternalScheduleEventResult, error)
	DeleteScheduleEvent(ctx context.Context, token ProviderToken, calendarID, eventID string) error
}
