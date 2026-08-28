package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// 6. Timeline Trends Models
type CoreTimelineTrends struct {
	StoryCompletion   []CoreStoryCompletionPoint   `json:"storyCompletion"`
	ObjectiveProgress []CoreObjectiveProgressPoint `json:"objectiveProgress"`
	TeamVelocity      []CoreTeamVelocityPoint      `json:"teamVelocity"`
	KeyMetricsTrend   []CoreKeyMetricsTrendPoint   `json:"keyMetricsTrend"`
}

type CoreStoryCompletionPoint struct {
	Date      time.Time `json:"date" db:"date"`
	Completed int       `json:"completed" db:"completed"`
	Created   int       `json:"created" db:"created"`
}

type CoreObjectiveProgressPoint struct {
	Date                time.Time `json:"date" db:"date"`
	TotalObjectives     int       `json:"totalObjectives" db:"total_objectives"`
	CompletedObjectives int       `json:"completedObjectives" db:"completed_objectives"`
}

type CoreTeamVelocityPoint struct {
	Date     time.Time `json:"date" db:"date"`
	TeamID   uuid.UUID `json:"teamId" db:"team_id"`
	Velocity int       `json:"velocity" db:"velocity"`
}

type CoreKeyMetricsTrendPoint struct {
	Date          time.Time `json:"date" db:"date"`
	ActiveUsers   int       `json:"activeUsers" db:"active_users"`
	StoriesPerDay float64   `json:"storiesPerDay" db:"stories_per_day"`
	AvgCycleTime  float64   `json:"avgCycleTime" db:"avg_cycle_time"`
}
