package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// Workspace Reports Models

// ReportFilters represents common filters for workspace reports
type ReportFilters struct {
	ActorID      uuid.UUID   `json:"-"`
	TeamIDs      []uuid.UUID `json:"teamIds"`
	AssigneeIDs  []uuid.UUID `json:"assigneeIds"`
	StartDate    *time.Time  `json:"startDate"`
	EndDate      *time.Time  `json:"endDate"`
	SprintIDs    []uuid.UUID `json:"sprintIds"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds"`
}

type CoreWorkspaceAnalyticsEventInput struct {
	WorkspaceID uuid.UUID      `json:"workspaceId"`
	UserID      uuid.UUID      `json:"userId"`
	EventName   string         `json:"eventName"`
	Surface     string         `json:"surface"`
	TeamID      *uuid.UUID     `json:"teamId,omitempty"`
	StoryID     *uuid.UUID     `json:"storyId,omitempty"`
	ObjectiveID *uuid.UUID     `json:"objectiveId,omitempty"`
	SprintID    *uuid.UUID     `json:"sprintId,omitempty"`
	KeyResultID *uuid.UUID     `json:"keyResultId,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	OccurredAt  time.Time      `json:"occurredAt"`
}
