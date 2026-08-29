package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// 5. Sprint Analytics App Models
type AppSprintAnalyticsWorkspace struct {
	SprintProgress   []AppSprintProgressItem    `json:"sprintProgress"`
	CombinedBurndown []AppCombinedBurndownPoint `json:"combinedBurndown"`
	TeamAllocation   []AppSprintTeamAllocation  `json:"teamAllocation"`
	SprintHealth     []AppSprintHealthItem      `json:"sprintHealth"`
}

type AppSprintProgressItem struct {
	SprintID   uuid.UUID `json:"sprintId"`
	SprintName string    `json:"sprintName"`
	TeamID     uuid.UUID `json:"teamId"`
	Completed  int       `json:"completed"`
	Total      int       `json:"total"`
	Status     string    `json:"status"`
}

type AppCombinedBurndownPoint struct {
	Date    time.Time `json:"date"`
	Planned int       `json:"planned"`
	Actual  int       `json:"actual"`
}

type AppSprintTeamAllocation struct {
	TeamID           uuid.UUID `json:"teamId"`
	TeamName         string    `json:"teamName"`
	ActiveSprints    int       `json:"activeSprints"`
	TotalStories     int       `json:"totalStories"`
	CompletedStories int       `json:"completedStories"`
}

type AppSprintHealthItem struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func toAppSprintAnalyticsWorkspace(analytics reports.CoreSprintAnalyticsWorkspace) AppSprintAnalyticsWorkspace {
	return AppSprintAnalyticsWorkspace{
		SprintProgress:   toAppSprintProgressItems(analytics.SprintProgress),
		CombinedBurndown: toAppCombinedBurndownPoints(analytics.CombinedBurndown),
		TeamAllocation:   toAppSprintTeamAllocations(analytics.TeamAllocation),
		SprintHealth:     toAppSprintHealthItems(analytics.SprintHealth),
	}
}

func toAppSprintProgressItems(items []reports.CoreSprintProgressItem) []AppSprintProgressItem {
	result := make([]AppSprintProgressItem, len(items))
	for i, item := range items {
		result[i] = AppSprintProgressItem{
			SprintID:   item.SprintID,
			SprintName: item.SprintName,
			TeamID:     item.TeamID,
			Completed:  item.Completed,
			Total:      item.Total,
			Status:     item.Status,
		}
	}
	return result
}

func toAppCombinedBurndownPoints(points []reports.CoreCombinedBurndownPoint) []AppCombinedBurndownPoint {
	result := make([]AppCombinedBurndownPoint, len(points))
	for i, point := range points {
		result[i] = AppCombinedBurndownPoint{
			Date:    point.Date,
			Planned: point.Planned,
			Actual:  point.Actual,
		}
	}
	return result
}

func toAppSprintTeamAllocations(allocations []reports.CoreSprintTeamAllocation) []AppSprintTeamAllocation {
	result := make([]AppSprintTeamAllocation, len(allocations))
	for i, allocation := range allocations {
		result[i] = AppSprintTeamAllocation{
			TeamID:           allocation.TeamID,
			TeamName:         allocation.TeamName,
			ActiveSprints:    allocation.ActiveSprints,
			TotalStories:     allocation.TotalStories,
			CompletedStories: allocation.CompletedStories,
		}
	}
	return result
}

func toAppSprintHealthItems(items []reports.CoreSprintHealthItem) []AppSprintHealthItem {
	result := make([]AppSprintHealthItem, len(items))
	for i, item := range items {
		result[i] = AppSprintHealthItem{
			Status: item.Status,
			Count:  item.Count,
		}
	}
	return result
}
