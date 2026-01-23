package reports

import (
	"time"

	"github.com/google/uuid"
)

// CoreStoryStats represents story statistics
type CoreStoryStats struct {
	Closed     int `json:"closed"`
	Overdue    int `json:"overdue"`
	InProgress int `json:"inProgress"`
	Created    int `json:"created"`
	Assigned   int `json:"assigned"`
}

type StoryStatsFilters struct {
	StartDate time.Time
	EndDate   time.Time
}

// CoreContributionStats represents contribution statistics
type CoreContributionStats struct {
	Date          time.Time `db:"date"`
	Contributions int       `db:"contributions"`
}

// CoreUserStats represents user-specific statistics
type CoreUserStats struct {
	AssignedToMe int `db:"assigned_to_me"`
	CreatedByMe  int `db:"created_by_me"`
}

type CoreStatusStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type CorePriorityStats struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

type StatsFilters struct {
	TeamID      *uuid.UUID
	SprintID    *uuid.UUID
	ObjectiveID *uuid.UUID
}

// Workspace Reports Models

// ReportFilters represents common filters for workspace reports
type ReportFilters struct {
	TeamIDs      []uuid.UUID `json:"teamIds"`
	StartDate    *time.Time  `json:"startDate"`
	EndDate      *time.Time  `json:"endDate"`
	SprintIDs    []uuid.UUID `json:"sprintIds"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds"`
}

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

// 4. Team Performance Models
type CoreTeamPerformance struct {
	TeamWorkload        []CoreTeamWorkloadItem       `json:"teamWorkload"`
	MemberContributions []CoreMemberContributionItem `json:"memberContributions"`
	VelocityByTeam      []CoreTeamVelocityItem       `json:"velocityByTeam"`
	WorkloadTrend       []CoreWorkloadTrendPoint     `json:"workloadTrend"`
}

type CoreTeamWorkloadItem struct {
	TeamID    uuid.UUID `json:"teamId" db:"team_id"`
	TeamName  string    `json:"teamName" db:"team_name"`
	Assigned  int       `json:"assigned" db:"assigned"`
	Completed int       `json:"completed" db:"completed"`
	Capacity  int       `json:"capacity" db:"capacity"`
}

type CoreMemberContributionItem struct {
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	AvatarURL string    `json:"avatarUrl" db:"avatar_url"`
	TeamID    uuid.UUID `json:"teamId" db:"team_id"`
	Completed int       `json:"completed" db:"completed"`
	Assigned  int       `json:"assigned" db:"assigned"`
}

type CoreTeamVelocityItem struct {
	TeamID   uuid.UUID `json:"teamId" db:"team_id"`
	TeamName string    `json:"teamName" db:"team_name"`
	Week1    int       `json:"week1" db:"week1"`
	Week2    int       `json:"week2" db:"week2"`
	Week3    int       `json:"week3" db:"week3"`
	Average  float64   `json:"average" db:"average"`
}

type CoreWorkloadTrendPoint struct {
	Date      time.Time `json:"date" db:"date"`
	Assigned  int       `json:"assigned" db:"assigned"`
	Completed int       `json:"completed" db:"completed"`
}

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
