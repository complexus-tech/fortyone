package reportdomain

import (
	"time"

	"github.com/google/uuid"
)

// Workload Analysis Models
type CoreWorkloadAnalysis struct {
	Summary    CoreWorkloadSummary       `json:"summary"`
	Members    []CoreMemberWorkload      `json:"members"`
	Teams      []CoreTeamWorkloadSummary `json:"teams"`
	Unassigned CoreUnassignedWorkload    `json:"unassigned"`
	Risks      CoreWorkloadRisks         `json:"risks"`
}

type CoreWorkloadSummary struct {
	TotalOpenStories    int `json:"totalOpenStories" db:"total_open_stories"`
	TotalEstimate       int `json:"totalEstimate" db:"total_estimate"`
	OverdueStories      int `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories  int `json:"unestimatedStories" db:"unestimated_stories"`
	UnassignedStories   int `json:"unassignedStories" db:"unassigned_stories"`
}

type CoreMemberWorkload struct {
	UserID                uuid.UUID  `json:"userId" db:"user_id"`
	FullName              string     `json:"fullName" db:"full_name"`
	Username              string     `json:"username" db:"username"`
	AvatarURL             string     `json:"avatarUrl" db:"avatar_url"`
	TeamAIRoleTitle       string     `json:"teamAiRoleTitle" db:"team_ai_role_title"`
	TeamAIRoleDescription string     `json:"teamAiRoleDescription" db:"team_ai_role_description"`
	OpenStories           int        `json:"openStories" db:"open_stories"`
	StartedStories        int        `json:"startedStories" db:"started_stories"`
	PausedStories         int        `json:"pausedStories" db:"paused_stories"`
	CompletedStories      int        `json:"completedStories" db:"completed_stories"`
	OverdueStories        int        `json:"overdueStories" db:"overdue_stories"`
	UrgentStories         int        `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories   int        `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories    int        `json:"unestimatedStories" db:"unestimated_stories"`
	EstimateTotal         int        `json:"estimateTotal" db:"estimate_total"`
	LastStoryActivityAt   *time.Time `json:"lastStoryActivityAt,omitempty" db:"last_story_activity_at"`
}

type CoreTeamWorkloadSummary struct {
	TeamID             uuid.UUID `json:"teamId" db:"team_id"`
	TeamName           string    `json:"teamName" db:"team_name"`
	TeamCode           string    `json:"teamCode" db:"team_code"`
	OpenStories        int       `json:"openStories" db:"open_stories"`
	EstimateTotal      int       `json:"estimateTotal" db:"estimate_total"`
	OverdueStories     int       `json:"overdueStories" db:"overdue_stories"`
	UnassignedStories  int       `json:"unassignedStories" db:"unassigned_stories"`
	UnestimatedStories int       `json:"unestimatedStories" db:"unestimated_stories"`
}

type CoreUnassignedWorkload struct {
	Stories             int `json:"stories" db:"stories"`
	EstimateTotal       int `json:"estimateTotal" db:"estimate_total"`
	OverdueStories      int `json:"overdueStories" db:"overdue_stories"`
	UrgentStories       int `json:"urgentStories" db:"urgent_stories"`
	HighPriorityStories int `json:"highPriorityStories" db:"high_priority_stories"`
	UnestimatedStories  int `json:"unestimatedStories" db:"unestimated_stories"`
}

type CoreWorkloadRisks struct {
	OverloadedMembers   []CoreMemberWorkload `json:"overloadedMembers"`
	OverdueMembers      []CoreMemberWorkload `json:"overdueMembers"`
	UnassignedStories   int                  `json:"unassignedStories"`
	UnestimatedStories  int                  `json:"unestimatedStories"`
	HighPriorityStories int                  `json:"highPriorityStories"`
}
