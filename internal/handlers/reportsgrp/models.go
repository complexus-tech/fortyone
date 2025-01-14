package reportsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/reports"
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

type AppFilters struct {
	Days string `json:"days"`
}

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
