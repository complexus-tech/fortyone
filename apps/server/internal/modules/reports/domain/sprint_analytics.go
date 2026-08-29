package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// 5. Sprint Analytics Models
type CoreSprintAnalyticsWorkspace struct {
	SprintProgress   []CoreSprintProgressItem    `json:"sprintProgress"`
	CombinedBurndown []CoreCombinedBurndownPoint `json:"combinedBurndown"`
	TeamAllocation   []CoreSprintTeamAllocation  `json:"teamAllocation"`
	SprintHealth     []CoreSprintHealthItem      `json:"sprintHealth"`
}

type CoreSprintProgressItem struct {
	SprintID   uuid.UUID `json:"sprintId" db:"sprint_id"`
	SprintName string    `json:"sprintName" db:"sprint_name"`
	TeamID     uuid.UUID `json:"teamId" db:"team_id"`
	Completed  int       `json:"completed" db:"completed"`
	Total      int       `json:"total" db:"total"`
	Status     string    `json:"status" db:"status"`
}

type CoreCombinedBurndownPoint struct {
	Date    time.Time `json:"date" db:"date"`
	Planned int       `json:"planned" db:"planned"`
	Actual  int       `json:"actual" db:"actual"`
}

type CoreSprintTeamAllocation struct {
	TeamID           uuid.UUID `json:"teamId" db:"team_id"`
	TeamName         string    `json:"teamName" db:"team_name"`
	ActiveSprints    int       `json:"activeSprints" db:"active_sprints"`
	TotalStories     int       `json:"totalStories" db:"total_stories"`
	CompletedStories int       `json:"completedStories" db:"completed_stories"`
}

type CoreSprintHealthItem struct {
	Status string `json:"status" db:"status"`
	Count  int    `json:"count" db:"count"`
}
