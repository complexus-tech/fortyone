package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// 1. Workspace Overview App Models
type AppWorkspaceOverview struct {
	WorkspaceID     uuid.UUID                 `json:"workspaceId"`
	ReportDate      time.Time                 `json:"reportDate"`
	Filters         AppReportFilters          `json:"filters"`
	Metrics         AppWorkspaceMetrics       `json:"metrics"`
	CompletionTrend []AppCompletionTrendPoint `json:"completionTrend"`
	VelocityTrend   []AppVelocityTrendPoint   `json:"velocityTrend"`
}

type AppWorkspaceMetrics struct {
	TotalStories     int `json:"totalStories"`
	CompletedStories int `json:"completedStories"`
	ActiveObjectives int `json:"activeObjectives"`
	ActiveSprints    int `json:"activeSprints"`
	TotalTeamMembers int `json:"totalTeamMembers"`
}

type AppCompletionTrendPoint struct {
	Date      time.Time `json:"date"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
}

type AppVelocityTrendPoint struct {
	Period   string `json:"period"`
	Velocity int    `json:"velocity"`
}

func toAppWorkspaceOverview(overview reports.CoreWorkspaceOverview) AppWorkspaceOverview {
	return AppWorkspaceOverview{
		WorkspaceID:     overview.WorkspaceID,
		ReportDate:      overview.ReportDate,
		Filters:         toAppReportFilters(overview.Filters),
		Metrics:         toAppWorkspaceMetrics(overview.Metrics),
		CompletionTrend: toAppCompletionTrendPoints(overview.CompletionTrend),
		VelocityTrend:   toAppVelocityTrendPoints(overview.VelocityTrend),
	}
}

func toAppWorkspaceMetrics(metrics reports.CoreWorkspaceMetrics) AppWorkspaceMetrics {
	return AppWorkspaceMetrics{
		TotalStories:     metrics.TotalStories,
		CompletedStories: metrics.CompletedStories,
		ActiveObjectives: metrics.ActiveObjectives,
		ActiveSprints:    metrics.ActiveSprints,
		TotalTeamMembers: metrics.TotalTeamMembers,
	}
}

func toAppCompletionTrendPoints(points []reports.CoreCompletionTrendPoint) []AppCompletionTrendPoint {
	result := make([]AppCompletionTrendPoint, len(points))
	for i, point := range points {
		result[i] = AppCompletionTrendPoint{
			Date:      point.Date,
			Completed: point.Completed,
			Total:     point.Total,
		}
	}
	return result
}

func toAppVelocityTrendPoints(points []reports.CoreVelocityTrendPoint) []AppVelocityTrendPoint {
	result := make([]AppVelocityTrendPoint, len(points))
	for i, point := range points {
		result[i] = AppVelocityTrendPoint{
			Period:   point.Period,
			Velocity: point.Velocity,
		}
	}
	return result
}
