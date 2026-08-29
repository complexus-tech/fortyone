package reportshttp

import (
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

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
