package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

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
