package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

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
