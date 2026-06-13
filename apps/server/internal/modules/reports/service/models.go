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
	StartDate   time.Time
	EndDate     time.Time
}

// Workspace Reports Models

// ReportFilters represents common filters for workspace reports
type ReportFilters struct {
	TeamIDs      []uuid.UUID `json:"teamIds"`
	AssigneeIDs  []uuid.UUID `json:"assigneeIds"`
	StartDate    *time.Time  `json:"startDate"`
	EndDate      *time.Time  `json:"endDate"`
	SprintIDs    []uuid.UUID `json:"sprintIds"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds"`
}

// Workload Analysis Models
type CoreWorkloadAnalysis struct {
	Summary    CoreWorkloadSummary       `json:"summary"`
	Members    []CoreMemberWorkload      `json:"members"`
	Teams      []CoreTeamWorkloadSummary `json:"teams"`
	Unassigned CoreUnassignedWorkload    `json:"unassigned"`
	Risks      CoreWorkloadRisks         `json:"risks"`
}

type CoreWorkloadSummary struct {
	TotalOpenStories    int `json:"totalOpenStories" db:"total_open_stories"`
	TotalEstimate       int `json:"totalEstimate" db:"total_estimate"`
	OverdueStories      int `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories  int `json:"unestimatedStories" db:"unestimated_stories"`
	UnassignedStories   int `json:"unassignedStories" db:"unassigned_stories"`
}

type CoreMemberWorkload struct {
	UserID              uuid.UUID `json:"userId" db:"user_id"`
	FullName            string    `json:"fullName" db:"full_name"`
	Username            string    `json:"username" db:"username"`
	AvatarURL           string    `json:"avatarUrl" db:"avatar_url"`
	OpenStories         int       `json:"openStories" db:"open_stories"`
	StartedStories      int       `json:"startedStories" db:"started_stories"`
	PausedStories       int       `json:"pausedStories" db:"paused_stories"`
	CompletedStories    int       `json:"completedStories" db:"completed_stories"`
	OverdueStories      int       `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int       `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int       `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories  int       `json:"unestimatedStories" db:"unestimated_stories"`
	EstimateTotal       int       `json:"estimateTotal" db:"estimate_total"`
}

type CoreTeamWorkloadSummary struct {
	TeamID             uuid.UUID `json:"teamId" db:"team_id"`
	TeamName           string    `json:"teamName" db:"team_name"`
	TeamCode           string    `json:"teamCode" db:"team_code"`
	OpenStories        int       `json:"openStories" db:"open_stories"`
	EstimateTotal      int       `json:"estimateTotal" db:"estimate_total"`
	OverdueStories     int       `json:"overdueStories" db:"overdue_stories"`
	UnassignedStories  int       `json:"unassignedStories" db:"unassigned_stories"`
	UnestimatedStories int       `json:"unestimatedStories" db:"unestimated_stories"`
}

type CoreUnassignedWorkload struct {
	Stories             int `json:"stories" db:"stories"`
	EstimateTotal       int `json:"estimateTotal" db:"estimate_total"`
	OverdueStories      int `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories  int `json:"unestimatedStories" db:"unestimated_stories"`
}

type CoreWorkloadRisks struct {
	OverloadedMembers   []CoreMemberWorkload `json:"overloadedMembers"`
	OverdueMembers      []CoreMemberWorkload `json:"overdueMembers"`
	UnassignedStories   int                  `json:"unassignedStories"`
	UnestimatedStories  int                  `json:"unestimatedStories"`
	HighPriorityStories int                  `json:"highPriorityStories"`
}

type PulseRiskSeverity string

const (
	PulseRiskSeverityHigh   PulseRiskSeverity = "high"
	PulseRiskSeverityMedium PulseRiskSeverity = "medium"
	PulseRiskSeverityLow    PulseRiskSeverity = "low"
)

type PulseRiskKind string

const (
	PulseRiskKindOverdueStories    PulseRiskKind = "overdue_stories"
	PulseRiskKindBlockedStories    PulseRiskKind = "blocked_stories"
	PulseRiskKindOverloadedMembers PulseRiskKind = "overloaded_members"
	PulseRiskKindAtRiskSprints     PulseRiskKind = "at_risk_sprints"
	PulseRiskKindAtRiskObjectives  PulseRiskKind = "at_risk_objectives"
	PulseRiskKindPendingRequests   PulseRiskKind = "pending_requests"
	PulseRiskKindUnassignedStories PulseRiskKind = "unassigned_stories"
)

type CorePulseReport struct {
	WorkspaceID uuid.UUID                `json:"workspaceId"`
	ReportDate  time.Time                `json:"reportDate"`
	Filters     ReportFilters            `json:"filters"`
	Summary     CorePulseSummary         `json:"summary"`
	Stories     CorePulseStoryHealth     `json:"stories"`
	Sprints     CorePulseSprintHealth    `json:"sprints"`
	Objectives  CorePulseObjectiveHealth `json:"objectives"`
	Requests    CorePulseRequestHealth   `json:"requests"`
	Workload    CoreWorkloadAnalysis     `json:"workload"`
	Risks       []CorePulseRisk          `json:"risks"`
}

type CorePulseSummary struct {
	OpenStories       int `json:"openStories"`
	OverdueStories    int `json:"overdueStories"`
	BlockedStories    int `json:"blockedStories"`
	AtRiskSprints     int `json:"atRiskSprints"`
	AtRiskObjectives  int `json:"atRiskObjectives"`
	PendingRequests   int `json:"pendingRequests"`
	OverloadedMembers int `json:"overloadedMembers"`
}

type CorePulseStoryHealth struct {
	OpenStories         int `json:"openStories" db:"open_stories"`
	StartedStories      int `json:"startedStories" db:"started_stories"`
	PausedStories       int `json:"pausedStories" db:"paused_stories"`
	CompletedStories    int `json:"completedStories" db:"completed_stories"`
	CancelledStories    int `json:"cancelledStories" db:"cancelled_stories"`
	BlockedStories      int `json:"blockedStories" db:"blocked_stories"`
	OverdueStories      int `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int `json:"highPriorityStories" db:"high_priority_stories"`
	UnassignedStories   int `json:"unassignedStories" db:"unassigned_stories"`
	UnestimatedStories  int `json:"unestimatedStories" db:"unestimated_stories"`
}

type CorePulseSprintHealth struct {
	ActiveSprints      int `json:"activeSprints" db:"active_sprints"`
	UpcomingSprints    int `json:"upcomingSprints" db:"upcoming_sprints"`
	CompletedSprints   int `json:"completedSprints" db:"completed_sprints"`
	AtRiskSprints      int `json:"atRiskSprints" db:"at_risk_sprints"`
	OverdueSprints     int `json:"overdueSprints" db:"overdue_sprints"`
	UnestimatedStories int `json:"unestimatedStories" db:"unestimated_stories"`
}

type CorePulseObjectiveHealth struct {
	ActiveObjectives   int `json:"activeObjectives" db:"active_objectives"`
	AtRiskObjectives   int `json:"atRiskObjectives" db:"at_risk_objectives"`
	OffTrackObjectives int `json:"offTrackObjectives" db:"off_track_objectives"`
	OverdueObjectives  int `json:"overdueObjectives" db:"overdue_objectives"`
	ObjectivesDueSoon  int `json:"objectivesDueSoon" db:"objectives_due_soon"`
}

type CorePulseRequestHealth struct {
	PendingRequests  int `json:"pendingRequests" db:"pending_requests"`
	UrgentRequests   int `json:"urgentRequests" db:"urgent_requests"`
	HighRequests     int `json:"highRequests" db:"high_requests"`
	GitHubRequests   int `json:"gitHubRequests" db:"github_requests"`
	SlackRequests    int `json:"slackRequests" db:"slack_requests"`
	IntercomRequests int `json:"intercomRequests" db:"intercom_requests"`
	StaleRequests    int `json:"staleRequests" db:"stale_requests"`
}

type CorePulseRisk struct {
	Kind        PulseRiskKind     `json:"kind"`
	Severity    PulseRiskSeverity `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Count       int               `json:"count"`
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
