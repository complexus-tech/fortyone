package reports

import reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"

var (
	ErrInvalidReportFilters           = reportdomain.ErrInvalidReportFilters
	ErrInvalidWorkspaceAnalyticsEvent = reportdomain.ErrInvalidWorkspaceAnalyticsEvent
	ErrReportsAccessDenied            = reportdomain.ErrReportsAccessDenied
)
