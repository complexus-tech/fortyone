package reportsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/reports"
)

type dbStoryStats struct {
	Closed     int `db:"closed"`
	Overdue    int `db:"overdue"`
	InProgress int `db:"in_progress"`
	Created    int `db:"created"`
	Assigned   int `db:"assigned"`
}

type dbContributionStats struct {
	Date          time.Time `db:"date"`
	Contributions int       `db:"contributions"`
}

type dbUserStats struct {
	AssignedToMe int `db:"assigned_to_me"`
	CreatedByMe  int `db:"created_by_me"`
}

type dbStatusStats struct {
	Name  string `db:"name"`
	Count int    `db:"count"`
}

type dbPriorityStats struct {
	Priority string `db:"priority"`
	Count    int    `db:"count"`
}

func toCoreStoryStats(s dbStoryStats) reports.CoreStoryStats {
	return reports.CoreStoryStats{
		Closed:     s.Closed,
		Overdue:    s.Overdue,
		InProgress: s.InProgress,
		Created:    s.Created,
		Assigned:   s.Assigned,
	}
}

func toCoreContributionStats(c dbContributionStats) reports.CoreContributionStats {
	return reports.CoreContributionStats{
		Date:          c.Date,
		Contributions: c.Contributions,
	}
}

func toCoreContributionsStats(cs []dbContributionStats) []reports.CoreContributionStats {
	stats := make([]reports.CoreContributionStats, len(cs))
	for i, c := range cs {
		stats[i] = toCoreContributionStats(c)
	}
	return stats
}

func toCoreUserStats(s dbUserStats) reports.CoreUserStats {
	return reports.CoreUserStats{
		AssignedToMe: s.AssignedToMe,
		CreatedByMe:  s.CreatedByMe,
	}
}

func toCoreStatusStats(s []dbStatusStats) []reports.CoreStatusStats {
	stats := make([]reports.CoreStatusStats, len(s))
	for i, stat := range s {
		stats[i] = reports.CoreStatusStats{
			Name:  stat.Name,
			Count: stat.Count,
		}
	}
	return stats
}

func toCorePriorityStats(s []dbPriorityStats) []reports.CorePriorityStats {
	stats := make([]reports.CorePriorityStats, len(s))
	for i, stat := range s {
		stats[i] = reports.CorePriorityStats{
			Priority: stat.Priority,
			Count:    stat.Count,
		}
	}
	return stats
}
