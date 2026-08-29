package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// 4. Team Performance Models
type CoreTeamPerformance struct {
	TeamWorkload        []CoreTeamWorkloadItem       `json:"teamWorkload"`
	MemberContributions []CoreMemberContributionItem `json:"memberContributions"`
	VelocityByTeam      []CoreTeamVelocityItem       `json:"velocityByTeam"`
	WorkloadTrend       []CoreWorkloadTrendPoint     `json:"workloadTrend"`
}

type CoreTeamWorkloadItem struct {
	TeamID    uuid.UUID `json:"teamId" db:"team_id"`
	TeamName  string    `json:"teamName" db:"team_name"`
	Assigned  int       `json:"assigned" db:"assigned"`
	Completed int       `json:"completed" db:"completed"`
	Capacity  int       `json:"capacity" db:"capacity"`
}

type CoreMemberContributionItem struct {
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	AvatarURL string    `json:"avatarUrl" db:"avatar_url"`
	TeamID    uuid.UUID `json:"teamId" db:"team_id"`
	Completed int       `json:"completed" db:"completed"`
	Assigned  int       `json:"assigned" db:"assigned"`
}

type CoreTeamVelocityItem struct {
	TeamID   uuid.UUID `json:"teamId" db:"team_id"`
	TeamName string    `json:"teamName" db:"team_name"`
	Week1    int       `json:"week1" db:"week1"`
	Week2    int       `json:"week2" db:"week2"`
	Week3    int       `json:"week3" db:"week3"`
	Average  float64   `json:"average" db:"average"`
}

type CoreWorkloadTrendPoint struct {
	Date      time.Time `json:"date" db:"date"`
	Assigned  int       `json:"assigned" db:"assigned"`
	Completed int       `json:"completed" db:"completed"`
}
