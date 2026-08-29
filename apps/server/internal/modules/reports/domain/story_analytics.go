package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// 2. Story Analytics Models
type CoreStoryAnalytics struct {
	StatusBreakdown      []CoreStatusBreakdownItem      `json:"statusBreakdown"`
	PriorityDistribution []CorePriorityDistributionItem `json:"priorityDistribution"`
	CompletionByTeam     []CoreTeamCompletionItem       `json:"completionByTeam"`
	Burndown             []CoreBurndownPoint            `json:"burndown"`
}

type CoreStatusBreakdownItem struct {
	StatusName string     `json:"statusName" db:"status_name"`
	Count      int        `json:"count" db:"count"`
	TeamID     *uuid.UUID `json:"teamId" db:"team_id"`
}

type CorePriorityDistributionItem struct {
	Priority string `json:"priority" db:"priority"`
	Count    int    `json:"count" db:"count"`
}

type CoreTeamCompletionItem struct {
	TeamID    uuid.UUID `json:"teamId" db:"team_id"`
	TeamName  string    `json:"teamName" db:"team_name"`
	Completed int       `json:"completed" db:"completed"`
	Total     int       `json:"total" db:"total"`
}

type CoreBurndownPoint struct {
	Date      time.Time `json:"date" db:"date"`
	Remaining int       `json:"remaining" db:"remaining"`
}
