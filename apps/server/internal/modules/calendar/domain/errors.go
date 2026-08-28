// Package domain defines calendar values and persistence contracts without
// coupling adapters to the application service or HTTP transport.
package domain

import "errors"

var (
	ErrCalendarNotConfigured           = errors.New("calendar integration is not configured")
	ErrInvalidCalendarState            = errors.New("invalid calendar setup state")
	ErrCalendarNotFound                = errors.New("calendar connection not found")
	ErrCalendarAccessDenied            = errors.New("calendar access denied")
	ErrCalendarCredentialsIncomplete   = errors.New("calendar credentials are incomplete")
	ErrCalendarEventNotFound           = errors.New("calendar event not found")
	ErrCalendarSyncSuperseded          = errors.New("calendar sync was superseded by newer credentials")
	ErrInvalidScheduleRange            = errors.New("calendar schedule range is invalid")
	ErrInvalidScheduleBlock            = errors.New("calendar schedule block is invalid")
	ErrCalendarScheduleConflict        = errors.New("calendar time conflicts with an existing meeting or schedule block")
	ErrCalendarScheduleStalePlan       = errors.New("calendar schedule plan is stale")
	ErrCalendarScheduleBlockNotFound   = errors.New("calendar schedule block not found")
	ErrManagedScheduleBlock            = errors.New("Maya-managed schedule blocks can only be changed by automatic scheduling") //lint:ignore ST1005 Maya is a product name.
	ErrCalendarCleanupPending          = errors.New("calendar cleanup is still in progress")
	ErrCalendarAccountChangePending    = errors.New("disconnect the existing calendar before connecting a different provider account")
	ErrCalendarFullSyncRequired        = errors.New("calendar full sync is required")
	ErrInvalidCalendarNotification     = errors.New("invalid calendar notification")
	ErrCalendarWebhookNotConfigured    = errors.New("calendar webhook is not configured")
	ErrCalendarReauthorizationRequired = errors.New("calendar write access requires reauthorization")
)
