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
