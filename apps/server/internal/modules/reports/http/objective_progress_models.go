package reportshttp

import (
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// 3. Objective Progress App Models
type AppObjectiveProgress struct {
	HealthDistribution []AppHealthDistributionItem    `json:"healthDistribution"`
	StatusBreakdown    []AppObjectiveStatusItem       `json:"statusBreakdown"`
	KeyResultsProgress []AppKeyResultProgressItem     `json:"keyResultsProgress"`
	ProgressByTeam     []AppObjectiveTeamProgressItem `json:"progressByTeam"`
}

type AppHealthDistributionItem struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type AppObjectiveStatusItem struct {
	StatusName string `json:"statusName"`
	Count      int    `json:"count"`
}

type AppKeyResultProgressItem struct {
	ObjectiveID   uuid.UUID `json:"objectiveId"`
	ObjectiveName string    `json:"objectiveName"`
	Completed     int       `json:"completed"`
	Total         int       `json:"total"`
	AvgProgress   float64   `json:"avgProgress"`
}

type AppObjectiveTeamProgressItem struct {
	TeamID     uuid.UUID `json:"teamId"`
	TeamName   string    `json:"teamName"`
	Objectives int       `json:"objectives"`
	Completed  int       `json:"completed"`
}

func toAppObjectiveProgress(progress reports.CoreObjectiveProgress) AppObjectiveProgress {
	return AppObjectiveProgress{
		HealthDistribution: toAppHealthDistributionItems(progress.HealthDistribution),
		StatusBreakdown:    toAppObjectiveStatusItems(progress.StatusBreakdown),
		KeyResultsProgress: toAppKeyResultProgressItems(progress.KeyResultsProgress),
		ProgressByTeam:     toAppObjectiveTeamProgressItems(progress.ProgressByTeam),
	}
}

func toAppHealthDistributionItems(items []reports.CoreHealthDistributionItem) []AppHealthDistributionItem {
	result := make([]AppHealthDistributionItem, len(items))
	for i, item := range items {
		result[i] = AppHealthDistributionItem{
			Status: item.Status,
			Count:  item.Count,
		}
	}
	return result
}

func toAppObjectiveStatusItems(items []reports.CoreObjectiveStatusItem) []AppObjectiveStatusItem {
	result := make([]AppObjectiveStatusItem, len(items))
	for i, item := range items {
		result[i] = AppObjectiveStatusItem{
			StatusName: item.StatusName,
			Count:      item.Count,
		}
	}
	return result
}

func toAppKeyResultProgressItems(items []reports.CoreKeyResultProgressItem) []AppKeyResultProgressItem {
	result := make([]AppKeyResultProgressItem, len(items))
	for i, item := range items {
		result[i] = AppKeyResultProgressItem{
			ObjectiveID:   item.ObjectiveID,
			ObjectiveName: item.ObjectiveName,
			Completed:     item.Completed,
			Total:         item.Total,
			AvgProgress:   item.AvgProgress,
		}
	}
	return result
}

func toAppObjectiveTeamProgressItems(items []reports.CoreObjectiveTeamProgressItem) []AppObjectiveTeamProgressItem {
	result := make([]AppObjectiveTeamProgressItem, len(items))
	for i, item := range items {
		result[i] = AppObjectiveTeamProgressItem{
			TeamID:     item.TeamID,
			TeamName:   item.TeamName,
			Objectives: item.Objectives,
			Completed:  item.Completed,
		}
	}
	return result
}
