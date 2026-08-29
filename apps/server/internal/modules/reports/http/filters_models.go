package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

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
