package calendar

import "context"

type CalendarProvider interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (ProviderToken, error)
	ListBusyWindows(ctx context.Context, token ProviderToken, input BusyWindowInput) ([]CoreBusyWindow, error)
}
