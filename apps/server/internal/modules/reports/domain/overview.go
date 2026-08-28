package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// 1. Workspace Overview Models
type CoreWorkspaceOverview struct {
	WorkspaceID     uuid.UUID                  `json:"workspaceId"`
	ReportDate      time.Time                  `json:"reportDate"`
	Filters         ReportFilters              `json:"filters"`
	Metrics         CoreWorkspaceMetrics       `json:"metrics"`
	CompletionTrend []CoreCompletionTrendPoint `json:"completionTrend"`
	VelocityTrend   []CoreVelocityTrendPoint   `json:"velocityTrend"`
}

type CoreWorkspaceMetrics struct {
	TotalStories     int `json:"totalStories" db:"total_stories"`
	CompletedStories int `json:"completedStories" db:"completed_stories"`
	ActiveObjectives int `json:"activeObjectives" db:"active_objectives"`
	ActiveSprints    int `json:"activeSprints" db:"active_sprints"`
	TotalTeamMembers int `json:"totalTeamMembers" db:"total_team_members"`
}

type CoreCompletionTrendPoint struct {
	Date      time.Time `json:"date" db:"date"`
	Completed int       `json:"completed" db:"completed"`
	Total     int       `json:"total" db:"total"`
}

type CoreVelocityTrendPoint struct {
	Period   string `json:"period" db:"period"`
	Velocity int    `json:"velocity" db:"velocity"`
}
