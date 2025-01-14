package reports

import (
	"time"
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
