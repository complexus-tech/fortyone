package reportdomain

import "github.com/google/uuid"

// 3. Objective Progress Models
type CoreObjectiveProgress struct {
	HealthDistribution []CoreHealthDistributionItem    `json:"healthDistribution"`
	StatusBreakdown    []CoreObjectiveStatusItem       `json:"statusBreakdown"`
	KeyResultsProgress []CoreKeyResultProgressItem     `json:"keyResultsProgress"`
	ProgressByTeam     []CoreObjectiveTeamProgressItem `json:"progressByTeam"`
}

type CoreHealthDistributionItem struct {
	Status string `json:"status" db:"status"`
	Count  int    `json:"count" db:"count"`
}

type CoreObjectiveStatusItem struct {
	StatusName string `json:"statusName" db:"status_name"`
	Count      int    `json:"count" db:"count"`
}

type CoreKeyResultProgressItem struct {
	ObjectiveID   uuid.UUID `json:"objectiveId" db:"objective_id"`
	ObjectiveName string    `json:"objectiveName" db:"objective_name"`
	Completed     int       `json:"completed" db:"completed"`
	Total         int       `json:"total" db:"total"`
	AvgProgress   float64   `json:"avgProgress" db:"avg_progress"`
}

type CoreObjectiveTeamProgressItem struct {
	TeamID     uuid.UUID `json:"teamId" db:"team_id"`
	TeamName   string    `json:"teamName" db:"team_name"`
	Objectives int       `json:"objectives" db:"objectives"`
	Completed  int       `json:"completed" db:"completed"`
}
