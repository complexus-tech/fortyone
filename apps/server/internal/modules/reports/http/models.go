package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
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
	AssigneeIDs  []uuid.UUID `json:"assigneeIds" query:"assigneeIds"`
	StartDate    *time.Time  `json:"startDate" query:"startDate"`
	EndDate      *time.Time  `json:"endDate" query:"endDate"`
	SprintIDs    []uuid.UUID `json:"sprintIds" query:"sprintIds"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds" query:"objectiveIds"`
}

type AppTrackWorkspaceAnalyticsEventRequest struct {
	EventName   string         `json:"eventName"`
	Surface     string         `json:"surface"`
	TeamID      *uuid.UUID     `json:"teamId,omitempty"`
	StoryID     *uuid.UUID     `json:"storyId,omitempty"`
	ObjectiveID *uuid.UUID     `json:"objectiveId,omitempty"`
	SprintID    *uuid.UUID     `json:"sprintId,omitempty"`
	KeyResultID *uuid.UUID     `json:"keyResultId,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	OccurredAt  *time.Time     `json:"occurredAt,omitempty"`
}

type AppTrackWorkspaceAnalyticsEventResponse struct {
	EventName  string    `json:"eventName"`
	Surface    string    `json:"surface"`
	OccurredAt time.Time `json:"occurredAt"`
}

// Workload Analysis App Models
type AppWorkloadAnalysis struct {
	Summary    AppWorkloadSummary       `json:"summary"`
	Members    []AppMemberWorkload      `json:"members"`
	Teams      []AppTeamWorkloadSummary `json:"teams"`
	Unassigned AppUnassignedWorkload    `json:"unassigned"`
	Risks      AppWorkloadRisks         `json:"risks"`
}

type AppWorkloadSummary struct {
	TotalOpenStories    int `json:"totalOpenStories"`
	TotalEstimate       int `json:"totalEstimate"`
	OverdueStories      int `json:"overdueStories"`
	UrgentStories       int `json:"urgentStories"`
	HighPriorityStories int `json:"highPriorityStories"`
	UnestimatedStories  int `json:"unestimatedStories"`
	UnassignedStories   int `json:"unassignedStories"`
}

type AppMemberWorkload struct {
	UserID              uuid.UUID `json:"userId"`
	FullName            string    `json:"fullName"`
	Username            string    `json:"username"`
	AvatarURL           string    `json:"avatarUrl"`
	OpenStories         int       `json:"openStories"`
	StartedStories      int       `json:"startedStories"`
	PausedStories       int       `json:"pausedStories"`
	CompletedStories    int       `json:"completedStories"`
	OverdueStories      int       `json:"overdueStories"`
	UrgentStories       int       `json:"urgentStories"`
	HighPriorityStories int       `json:"highPriorityStories"`
	UnestimatedStories  int       `json:"unestimatedStories"`
	EstimateTotal       int       `json:"estimateTotal"`
}

type AppTeamWorkloadSummary struct {
	TeamID             uuid.UUID `json:"teamId"`
	TeamName           string    `json:"teamName"`
	TeamCode           string    `json:"teamCode"`
	OpenStories        int       `json:"openStories"`
	EstimateTotal      int       `json:"estimateTotal"`
	OverdueStories     int       `json:"overdueStories"`
	UnassignedStories  int       `json:"unassignedStories"`
	UnestimatedStories int       `json:"unestimatedStories"`
}

type AppUnassignedWorkload struct {
	Stories             int `json:"stories"`
	EstimateTotal       int `json:"estimateTotal"`
	OverdueStories      int `json:"overdueStories"`
	UrgentStories       int `json:"urgentStories"`
	HighPriorityStories int `json:"highPriorityStories"`
	UnestimatedStories  int `json:"unestimatedStories"`
}

type AppWorkloadRisks struct {
	OverloadedMembers   []AppMemberWorkload `json:"overloadedMembers"`
	OverdueMembers      []AppMemberWorkload `json:"overdueMembers"`
	UnassignedStories   int                 `json:"unassignedStories"`
	UnestimatedStories  int                 `json:"unestimatedStories"`
	HighPriorityStories int                 `json:"highPriorityStories"`
}

type AppPulseReport struct {
	WorkspaceID uuid.UUID               `json:"workspaceId"`
	ReportDate  time.Time               `json:"reportDate"`
	Filters     AppReportFilters        `json:"filters"`
	Summary     AppPulseSummary         `json:"summary"`
	Stories     AppPulseStoryHealth     `json:"stories"`
	Sprints     AppPulseSprintHealth    `json:"sprints"`
	Objectives  AppPulseObjectiveHealth `json:"objectives"`
	Requests    AppPulseRequestHealth   `json:"requests"`
	Workload    AppWorkloadAnalysis     `json:"workload"`
	Risks       []AppPulseRisk          `json:"risks"`
}

type AppPulseSummary struct {
	OpenStories       int `json:"openStories"`
	OverdueStories    int `json:"overdueStories"`
	BlockedStories    int `json:"blockedStories"`
	AtRiskSprints     int `json:"atRiskSprints"`
	AtRiskObjectives  int `json:"atRiskObjectives"`
	PendingRequests   int `json:"pendingRequests"`
	OverloadedMembers int `json:"overloadedMembers"`
}

type AppPulseStoryHealth struct {
	OpenStories         int `json:"openStories"`
	StartedStories      int `json:"startedStories"`
	PausedStories       int `json:"pausedStories"`
	CompletedStories    int `json:"completedStories"`
	CancelledStories    int `json:"cancelledStories"`
	BlockedStories      int `json:"blockedStories"`
	OverdueStories      int `json:"overdueStories"`
	UrgentStories       int `json:"urgentStories"`
	HighPriorityStories int `json:"highPriorityStories"`
	UnassignedStories   int `json:"unassignedStories"`
	UnestimatedStories  int `json:"unestimatedStories"`
}

type AppPulseSprintHealth struct {
	ActiveSprints      int `json:"activeSprints"`
	UpcomingSprints    int `json:"upcomingSprints"`
	CompletedSprints   int `json:"completedSprints"`
	AtRiskSprints      int `json:"atRiskSprints"`
	OverdueSprints     int `json:"overdueSprints"`
	UnestimatedStories int `json:"unestimatedStories"`
}

type AppPulseObjectiveHealth struct {
	ActiveObjectives   int `json:"activeObjectives"`
	AtRiskObjectives   int `json:"atRiskObjectives"`
	OffTrackObjectives int `json:"offTrackObjectives"`
	OverdueObjectives  int `json:"overdueObjectives"`
	ObjectivesDueSoon  int `json:"objectivesDueSoon"`
}

type AppPulseRequestHealth struct {
	PendingRequests  int `json:"pendingRequests"`
	UrgentRequests   int `json:"urgentRequests"`
	HighRequests     int `json:"highRequests"`
	GitHubRequests   int `json:"gitHubRequests"`
	SlackRequests    int `json:"slackRequests"`
	IntercomRequests int `json:"intercomRequests"`
	StaleRequests    int `json:"staleRequests"`
}

type AppPulseRisk struct {
	Kind        reports.PulseRiskKind     `json:"kind"`
	Severity    reports.PulseRiskSeverity `json:"severity"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Count       int                       `json:"count"`
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
		AssigneeIDs:  filters.AssigneeIDs,
		StartDate:    filters.StartDate,
		EndDate:      filters.EndDate,
		SprintIDs:    filters.SprintIDs,
		ObjectiveIDs: filters.ObjectiveIDs,
	}
}

func toAppWorkloadAnalysis(analysis reports.CoreWorkloadAnalysis) AppWorkloadAnalysis {
	return AppWorkloadAnalysis{
		Summary:    toAppWorkloadSummary(analysis.Summary),
		Members:    toAppMemberWorkloads(analysis.Members),
		Teams:      toAppTeamWorkloadSummaries(analysis.Teams),
		Unassigned: toAppUnassignedWorkload(analysis.Unassigned),
		Risks:      toAppWorkloadRisks(analysis.Risks),
	}
}

func toAppWorkloadSummary(summary reports.CoreWorkloadSummary) AppWorkloadSummary {
	return AppWorkloadSummary{
		TotalOpenStories:    summary.TotalOpenStories,
		TotalEstimate:       summary.TotalEstimate,
		OverdueStories:      summary.OverdueStories,
		UrgentStories:       summary.UrgentStories,
		HighPriorityStories: summary.HighPriorityStories,
		UnestimatedStories:  summary.UnestimatedStories,
		UnassignedStories:   summary.UnassignedStories,
	}
}

func toAppMemberWorkloads(members []reports.CoreMemberWorkload) []AppMemberWorkload {
	result := make([]AppMemberWorkload, len(members))
	for i, member := range members {
		result[i] = toAppMemberWorkload(member)
	}
	return result
}

func toAppMemberWorkload(member reports.CoreMemberWorkload) AppMemberWorkload {
	return AppMemberWorkload{
		UserID:              member.UserID,
		FullName:            member.FullName,
		Username:            member.Username,
		AvatarURL:           member.AvatarURL,
		OpenStories:         member.OpenStories,
		StartedStories:      member.StartedStories,
		PausedStories:       member.PausedStories,
		CompletedStories:    member.CompletedStories,
		OverdueStories:      member.OverdueStories,
		UrgentStories:       member.UrgentStories,
		HighPriorityStories: member.HighPriorityStories,
		UnestimatedStories:  member.UnestimatedStories,
		EstimateTotal:       member.EstimateTotal,
	}
}

func toAppTeamWorkloadSummaries(teams []reports.CoreTeamWorkloadSummary) []AppTeamWorkloadSummary {
	result := make([]AppTeamWorkloadSummary, len(teams))
	for i, team := range teams {
		result[i] = AppTeamWorkloadSummary{
			TeamID:             team.TeamID,
			TeamName:           team.TeamName,
			TeamCode:           team.TeamCode,
			OpenStories:        team.OpenStories,
			EstimateTotal:      team.EstimateTotal,
			OverdueStories:     team.OverdueStories,
			UnassignedStories:  team.UnassignedStories,
			UnestimatedStories: team.UnestimatedStories,
		}
	}
	return result
}

func toAppUnassignedWorkload(unassigned reports.CoreUnassignedWorkload) AppUnassignedWorkload {
	return AppUnassignedWorkload{
		Stories:             unassigned.Stories,
		EstimateTotal:       unassigned.EstimateTotal,
		OverdueStories:      unassigned.OverdueStories,
		UrgentStories:       unassigned.UrgentStories,
		HighPriorityStories: unassigned.HighPriorityStories,
		UnestimatedStories:  unassigned.UnestimatedStories,
	}
}

func toAppWorkloadRisks(risks reports.CoreWorkloadRisks) AppWorkloadRisks {
	return AppWorkloadRisks{
		OverloadedMembers:   toAppMemberWorkloads(risks.OverloadedMembers),
		OverdueMembers:      toAppMemberWorkloads(risks.OverdueMembers),
		UnassignedStories:   risks.UnassignedStories,
		UnestimatedStories:  risks.UnestimatedStories,
		HighPriorityStories: risks.HighPriorityStories,
	}
}

func toAppPulseReport(report reports.CorePulseReport) AppPulseReport {
	return AppPulseReport{
		WorkspaceID: report.WorkspaceID,
		ReportDate:  report.ReportDate,
		Filters:     toAppReportFilters(report.Filters),
		Summary:     toAppPulseSummary(report.Summary),
		Stories:     toAppPulseStoryHealth(report.Stories),
		Sprints:     toAppPulseSprintHealth(report.Sprints),
		Objectives:  toAppPulseObjectiveHealth(report.Objectives),
		Requests:    toAppPulseRequestHealth(report.Requests),
		Workload:    toAppWorkloadAnalysis(report.Workload),
		Risks:       toAppPulseRisks(report.Risks),
	}
}

func toAppPulseSummary(summary reports.CorePulseSummary) AppPulseSummary {
	return AppPulseSummary{
		OpenStories:       summary.OpenStories,
		OverdueStories:    summary.OverdueStories,
		BlockedStories:    summary.BlockedStories,
		AtRiskSprints:     summary.AtRiskSprints,
		AtRiskObjectives:  summary.AtRiskObjectives,
		PendingRequests:   summary.PendingRequests,
		OverloadedMembers: summary.OverloadedMembers,
	}
}

func toAppPulseStoryHealth(health reports.CorePulseStoryHealth) AppPulseStoryHealth {
	return AppPulseStoryHealth{
		OpenStories:         health.OpenStories,
		StartedStories:      health.StartedStories,
		PausedStories:       health.PausedStories,
		CompletedStories:    health.CompletedStories,
		CancelledStories:    health.CancelledStories,
		BlockedStories:      health.BlockedStories,
		OverdueStories:      health.OverdueStories,
		UrgentStories:       health.UrgentStories,
		HighPriorityStories: health.HighPriorityStories,
		UnassignedStories:   health.UnassignedStories,
		UnestimatedStories:  health.UnestimatedStories,
	}
}

func toAppPulseSprintHealth(health reports.CorePulseSprintHealth) AppPulseSprintHealth {
	return AppPulseSprintHealth{
		ActiveSprints:      health.ActiveSprints,
		UpcomingSprints:    health.UpcomingSprints,
		CompletedSprints:   health.CompletedSprints,
		AtRiskSprints:      health.AtRiskSprints,
		OverdueSprints:     health.OverdueSprints,
		UnestimatedStories: health.UnestimatedStories,
	}
}

func toAppPulseObjectiveHealth(health reports.CorePulseObjectiveHealth) AppPulseObjectiveHealth {
	return AppPulseObjectiveHealth{
		ActiveObjectives:   health.ActiveObjectives,
		AtRiskObjectives:   health.AtRiskObjectives,
		OffTrackObjectives: health.OffTrackObjectives,
		OverdueObjectives:  health.OverdueObjectives,
		ObjectivesDueSoon:  health.ObjectivesDueSoon,
	}
}

func toAppPulseRequestHealth(health reports.CorePulseRequestHealth) AppPulseRequestHealth {
	return AppPulseRequestHealth{
		PendingRequests:  health.PendingRequests,
		UrgentRequests:   health.UrgentRequests,
		HighRequests:     health.HighRequests,
		GitHubRequests:   health.GitHubRequests,
		SlackRequests:    health.SlackRequests,
		IntercomRequests: health.IntercomRequests,
		StaleRequests:    health.StaleRequests,
	}
}

func toAppPulseRisks(risks []reports.CorePulseRisk) []AppPulseRisk {
	result := make([]AppPulseRisk, len(risks))
	for i, risk := range risks {
		result[i] = AppPulseRisk{
			Kind:        risk.Kind,
			Severity:    risk.Severity,
			Title:       risk.Title,
			Description: risk.Description,
			Count:       risk.Count,
		}
	}
	return result
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
