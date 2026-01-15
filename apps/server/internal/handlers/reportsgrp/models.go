package reportsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/reports"
	"github.com/google/uuid"
)

type AppStoryStats struct {
	Closed     int `json:"closed"`
	Overdue    int `json:"overdue"`
	InProgress int `json:"inProgress"`
	Created    int `json:"created"`
	Assigned   int `json:"assigned"`
}

type AppContributionStats struct {
	Date          time.Time `json:"date"`
	Contributions int       `json:"contributions"`
}

type AppUserStats struct {
	AssignedToMe int `json:"assignedToMe"`
	CreatedByMe  int `json:"createdByMe"`
}

type AppFilters struct {
	Days string `json:"days"`
}

type AppStatusStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AppPriorityStats struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

type AppStatsFilters struct {
	TeamID      *uuid.UUID `json:"teamId" db:"team_id"`
	SprintID    *uuid.UUID `json:"sprintId" db:"sprint_id"`
	ObjectiveID *uuid.UUID `json:"objectiveId" db:"objective_id"`
}

// Workspace Reports App Models

type AppReportFilters struct {
	TeamIDs      []uuid.UUID `json:"teamIds" query:"teamIds"`
	StartDate    *time.Time  `json:"startDate" query:"startDate"`
	EndDate      *time.Time  `json:"endDate" query:"endDate"`
	SprintIDs    []uuid.UUID `json:"sprintIds" query:"sprintIds"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds" query:"objectiveIds"`
}

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

// 4. Team Performance App Models
type AppTeamPerformance struct {
	TeamWorkload        []AppTeamWorkloadItem       `json:"teamWorkload"`
	MemberContributions []AppMemberContributionItem `json:"memberContributions"`
	VelocityByTeam      []AppTeamVelocityItem       `json:"velocityByTeam"`
	WorkloadTrend       []AppWorkloadTrendPoint     `json:"workloadTrend"`
}

type AppTeamWorkloadItem struct {
	TeamID    uuid.UUID `json:"teamId"`
	TeamName  string    `json:"teamName"`
	Assigned  int       `json:"assigned"`
	Completed int       `json:"completed"`
	Capacity  int       `json:"capacity"`
}

type AppMemberContributionItem struct {
	UserID    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatarUrl"`
	TeamID    uuid.UUID `json:"teamId"`
	Completed int       `json:"completed"`
	Assigned  int       `json:"assigned"`
}

type AppTeamVelocityItem struct {
	TeamID   uuid.UUID `json:"teamId"`
	TeamName string    `json:"teamName"`
	Week1    int       `json:"week1"`
	Week2    int       `json:"week2"`
	Week3    int       `json:"week3"`
	Average  float64   `json:"average"`
}

type AppWorkloadTrendPoint struct {
	Date      time.Time `json:"date"`
	Assigned  int       `json:"assigned"`
	Completed int       `json:"completed"`
}

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

// 6. Timeline Trends App Models
type AppTimelineTrends struct {
	StoryCompletion   []AppStoryCompletionPoint   `json:"storyCompletion"`
	ObjectiveProgress []AppObjectiveProgressPoint `json:"objectiveProgress"`
	TeamVelocity      []AppTeamVelocityPoint      `json:"teamVelocity"`
	KeyMetricsTrend   []AppKeyMetricsTrendPoint   `json:"keyMetricsTrend"`
}

type AppStoryCompletionPoint struct {
	Date      time.Time `json:"date"`
	Completed int       `json:"completed"`
	Created   int       `json:"created"`
}

type AppObjectiveProgressPoint struct {
	Date                time.Time `json:"date"`
	TotalObjectives     int       `json:"totalObjectives"`
	CompletedObjectives int       `json:"completedObjectives"`
}

type AppTeamVelocityPoint struct {
	Date     time.Time `json:"date"`
	TeamID   uuid.UUID `json:"teamId"`
	Velocity int       `json:"velocity"`
}

type AppKeyMetricsTrendPoint struct {
	Date          time.Time `json:"date"`
	ActiveUsers   int       `json:"activeUsers"`
	StoriesPerDay float64   `json:"storiesPerDay"`
	AvgCycleTime  float64   `json:"avgCycleTime"`
}

// Conversion functions for existing models
func toAppStoryStats(s reports.CoreStoryStats) AppStoryStats {
	return AppStoryStats{
		Closed:     s.Closed,
		Overdue:    s.Overdue,
		InProgress: s.InProgress,
		Created:    s.Created,
		Assigned:   s.Assigned,
	}
}

func toAppContributionStats(c reports.CoreContributionStats) AppContributionStats {
	return AppContributionStats{
		Date:          c.Date,
		Contributions: c.Contributions,
	}
}

func toAppContributionsStats(cs []reports.CoreContributionStats) []AppContributionStats {
	stats := make([]AppContributionStats, len(cs))
	for i, c := range cs {
		stats[i] = toAppContributionStats(c)
	}
	return stats
}

func toAppUserStats(s reports.CoreUserStats) AppUserStats {
	return AppUserStats{
		AssignedToMe: s.AssignedToMe,
		CreatedByMe:  s.CreatedByMe,
	}
}

func toAppStatusStats(s []reports.CoreStatusStats) []AppStatusStats {
	stats := make([]AppStatusStats, len(s))
	for i, stat := range s {
		stats[i] = AppStatusStats{
			Name:  stat.Name,
			Count: stat.Count,
		}
	}
	return stats
}

func toAppPriorityStats(s []reports.CorePriorityStats) []AppPriorityStats {
	stats := make([]AppPriorityStats, len(s))
	for i, stat := range s {
		stats[i] = AppPriorityStats{
			Priority: stat.Priority,
			Count:    stat.Count,
		}
	}
	return stats
}

// Conversion functions for workspace reports

func toAppReportFilters(filters reports.ReportFilters) AppReportFilters {
	return AppReportFilters{
		TeamIDs:      filters.TeamIDs,
		StartDate:    filters.StartDate,
		EndDate:      filters.EndDate,
		SprintIDs:    filters.SprintIDs,
		ObjectiveIDs: filters.ObjectiveIDs,
	}
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

func toAppTeamPerformance(performance reports.CoreTeamPerformance) AppTeamPerformance {
	return AppTeamPerformance{
		TeamWorkload:        toAppTeamWorkloadItems(performance.TeamWorkload),
		MemberContributions: toAppMemberContributionItems(performance.MemberContributions),
		VelocityByTeam:      toAppTeamVelocityItems(performance.VelocityByTeam),
		WorkloadTrend:       toAppWorkloadTrendPoints(performance.WorkloadTrend),
	}
}

func toAppTeamWorkloadItems(items []reports.CoreTeamWorkloadItem) []AppTeamWorkloadItem {
	result := make([]AppTeamWorkloadItem, len(items))
	for i, item := range items {
		result[i] = AppTeamWorkloadItem{
			TeamID:    item.TeamID,
			TeamName:  item.TeamName,
			Assigned:  item.Assigned,
			Completed: item.Completed,
			Capacity:  item.Capacity,
		}
	}
	return result
}

func toAppMemberContributionItems(items []reports.CoreMemberContributionItem) []AppMemberContributionItem {
	result := make([]AppMemberContributionItem, len(items))
	for i, item := range items {
		result[i] = AppMemberContributionItem{
			UserID:    item.UserID,
			Username:  item.Username,
			AvatarURL: item.AvatarURL,
			TeamID:    item.TeamID,
			Completed: item.Completed,
			Assigned:  item.Assigned,
		}
	}
	return result
}

func toAppTeamVelocityItems(items []reports.CoreTeamVelocityItem) []AppTeamVelocityItem {
	result := make([]AppTeamVelocityItem, len(items))
	for i, item := range items {
		result[i] = AppTeamVelocityItem{
			TeamID:   item.TeamID,
			TeamName: item.TeamName,
			Week1:    item.Week1,
			Week2:    item.Week2,
			Week3:    item.Week3,
			Average:  item.Average,
		}
	}
	return result
}

func toAppWorkloadTrendPoints(points []reports.CoreWorkloadTrendPoint) []AppWorkloadTrendPoint {
	result := make([]AppWorkloadTrendPoint, len(points))
	for i, point := range points {
		result[i] = AppWorkloadTrendPoint{
			Date:      point.Date,
			Assigned:  point.Assigned,
			Completed: point.Completed,
		}
	}
	return result
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

func toAppTimelineTrends(trends reports.CoreTimelineTrends) AppTimelineTrends {
	return AppTimelineTrends{
		StoryCompletion:   toAppStoryCompletionPoints(trends.StoryCompletion),
		ObjectiveProgress: toAppObjectiveProgressPoints(trends.ObjectiveProgress),
		TeamVelocity:      toAppTeamVelocityPoints(trends.TeamVelocity),
		KeyMetricsTrend:   toAppKeyMetricsTrendPoints(trends.KeyMetricsTrend),
	}
}

func toAppStoryCompletionPoints(points []reports.CoreStoryCompletionPoint) []AppStoryCompletionPoint {
	result := make([]AppStoryCompletionPoint, len(points))
	for i, point := range points {
		result[i] = AppStoryCompletionPoint{
			Date:      point.Date,
			Completed: point.Completed,
			Created:   point.Created,
		}
	}
	return result
}

func toAppObjectiveProgressPoints(points []reports.CoreObjectiveProgressPoint) []AppObjectiveProgressPoint {
	result := make([]AppObjectiveProgressPoint, len(points))
	for i, point := range points {
		result[i] = AppObjectiveProgressPoint{
			Date:                point.Date,
			TotalObjectives:     point.TotalObjectives,
			CompletedObjectives: point.CompletedObjectives,
		}
	}
	return result
}

func toAppTeamVelocityPoints(points []reports.CoreTeamVelocityPoint) []AppTeamVelocityPoint {
	result := make([]AppTeamVelocityPoint, len(points))
	for i, point := range points {
		result[i] = AppTeamVelocityPoint{
			Date:     point.Date,
			TeamID:   point.TeamID,
			Velocity: point.Velocity,
		}
	}
	return result
}

func toAppKeyMetricsTrendPoints(points []reports.CoreKeyMetricsTrendPoint) []AppKeyMetricsTrendPoint {
	result := make([]AppKeyMetricsTrendPoint, len(points))
	for i, point := range points {
		result[i] = AppKeyMetricsTrendPoint{
			Date:          point.Date,
			ActiveUsers:   point.ActiveUsers,
			StoriesPerDay: point.StoriesPerDay,
			AvgCycleTime:  point.AvgCycleTime,
		}
	}
	return result
}
