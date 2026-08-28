package reportdomain

import "errors"

var (
	ErrInvalidReportFilters           = errors.New("invalid report filters")
	ErrInvalidWorkspaceAnalyticsEvent = errors.New("invalid workspace analytics event")
	ErrReportsAccessDenied            = errors.New("reports access denied")
)
