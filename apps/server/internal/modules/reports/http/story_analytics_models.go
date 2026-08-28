package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// 2. Story Analytics App Models
type AppStoryAnalytics struct {
	StatusBreakdown      []AppStatusBreakdownItem      `json:"statusBreakdown"`
	PriorityDistribution []AppPriorityDistributionItem `json:"priorityDistribution"`
	CompletionByTeam     []AppTeamCompletionItem       `json:"completionByTeam"`
	Burndown             []AppBurndownPoint            `json:"burndown"`
}

type AppStatusBreakdownItem struct {
	StatusName string     `json:"statusName"`
	Count      int        `json:"count"`
	TeamID     *uuid.UUID `json:"teamId"`
}

type AppPriorityDistributionItem struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

type AppTeamCompletionItem struct {
	TeamID    uuid.UUID `json:"teamId"`
	TeamName  string    `json:"teamName"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
}

type AppBurndownPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
}

func toAppStoryAnalytics(analytics reports.CoreStoryAnalytics) AppStoryAnalytics {
	return AppStoryAnalytics{
		StatusBreakdown:      toAppStatusBreakdownItems(analytics.StatusBreakdown),
		PriorityDistribution: toAppPriorityDistributionItems(analytics.PriorityDistribution),
		CompletionByTeam:     toAppTeamCompletionItems(analytics.CompletionByTeam),
		Burndown:             toAppBurndownPoints(analytics.Burndown),
	}
}

func toAppStatusBreakdownItems(items []reports.CoreStatusBreakdownItem) []AppStatusBreakdownItem {
	result := make([]AppStatusBreakdownItem, len(items))
	for i, item := range items {
		result[i] = AppStatusBreakdownItem{
			StatusName: item.StatusName,
			Count:      item.Count,
			TeamID:     item.TeamID,
		}
	}
	return result
}

func toAppPriorityDistributionItems(items []reports.CorePriorityDistributionItem) []AppPriorityDistributionItem {
	result := make([]AppPriorityDistributionItem, len(items))
	for i, item := range items {
		result[i] = AppPriorityDistributionItem{
			Priority: item.Priority,
			Count:    item.Count,
		}
	}
	return result
}

func toAppTeamCompletionItems(items []reports.CoreTeamCompletionItem) []AppTeamCompletionItem {
	result := make([]AppTeamCompletionItem, len(items))
	for i, item := range items {
		result[i] = AppTeamCompletionItem{
			TeamID:    item.TeamID,
			TeamName:  item.TeamName,
			Completed: item.Completed,
			Total:     item.Total,
		}
	}
	return result
}

func toAppBurndownPoints(points []reports.CoreBurndownPoint) []AppBurndownPoint {
	result := make([]AppBurndownPoint, len(points))
	for i, point := range points {
		result[i] = AppBurndownPoint{
			Date:      point.Date,
			Remaining: point.Remaining,
		}
	}
	return result
}
