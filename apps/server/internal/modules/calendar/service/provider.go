package calendar

import "context"

type CalendarProvider interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (ProviderToken, error)
	SyncCalendar(ctx context.Context, token ProviderToken, input BusyWindowInput) (CalendarSyncSnapshot, error)
}
