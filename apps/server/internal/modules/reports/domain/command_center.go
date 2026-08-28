package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

type CoreWorkspaceCommandCenterReport struct {
	WorkspaceID   uuid.UUID                                `json:"workspaceId"`
	ReportDate    time.Time                                `json:"reportDate"`
	Filters       ReportFilters                            `json:"filters"`
	SectionErrors []CoreWorkspaceCommandCenterSectionError `json:"sectionErrors"`
	Overview      CoreWorkspaceOverview                    `json:"overview"`
	Pulse         CorePulseReport                          `json:"pulse"`
	Stories       CoreStoryAnalytics                       `json:"stories"`
	Objectives    CoreObjectiveProgress                    `json:"objectives"`
	Teams         CoreTeamPerformance                      `json:"teams"`
	Workload      CoreWorkloadAnalysis                     `json:"workload"`
	Sprints       CoreSprintAnalyticsWorkspace             `json:"sprints"`
	Trends        CoreTimelineTrends                       `json:"trends"`
	Requests      CoreRequestSourceAnalytics               `json:"requests"`
	Engagement    CoreWorkspaceEngagementAnalytics         `json:"engagement"`
}

type CoreWorkspaceCommandCenterSectionError struct {
	Section string `json:"section"`
	Message string `json:"message"`
}

type CoreRequestSourceAnalytics struct {
	Providers        []CoreRequestProviderPerformance `json:"providers"`
	TotalRequests    int                              `json:"totalRequests"`
	PendingRequests  int                              `json:"pendingRequests"`
	AcceptedRequests int                              `json:"acceptedRequests"`
	DeclinedRequests int                              `json:"declinedRequests"`
}

type CoreRequestProviderPerformance struct {
	Provider         string  `json:"provider"`
	TotalRequests    int     `json:"totalRequests"`
	PendingRequests  int     `json:"pendingRequests"`
	AcceptedRequests int     `json:"acceptedRequests"`
	DeclinedRequests int     `json:"declinedRequests"`
	UrgentRequests   int     `json:"urgentRequests"`
	HighRequests     int     `json:"highRequests"`
	StaleRequests    int     `json:"staleRequests"`
	AcceptanceRate   float64 `json:"acceptanceRate"`
}

type CoreWorkspaceEngagementAnalytics struct {
	TotalEvents     int                            `json:"totalEvents"`
	UniqueUsers     int                            `json:"uniqueUsers"`
	EventsByName    []CoreWorkspaceEngagementCount `json:"eventsByName"`
	EventsBySurface []CoreWorkspaceEngagementCount `json:"eventsBySurface"`
	TopUsers        []CoreWorkspaceEngagementUser  `json:"topUsers"`
}

type CoreWorkspaceEngagementCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type CoreWorkspaceEngagementUser struct {
	UserID    uuid.UUID `json:"userId"`
	FullName  string    `json:"fullName"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatarUrl"`
	Events    int       `json:"events"`
}
