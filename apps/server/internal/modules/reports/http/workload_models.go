package reportshttp

import (
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

// Workload Analysis App Models
type AppWorkloadAnalysis struct {
	Summary    AppWorkloadSummary       `json:"summary"`
	Members    []AppMemberWorkload      `json:"members"`
	Teams      []AppTeamWorkloadSummary `json:"teams"`
	Unassigned AppUnassignedWorkload    `json:"unassigned"`
	Risks      AppWorkloadRisks         `json:"risks"`
}

type AppWorkloadSummary struct {
	TotalOpenStories    int `json:"totalOpenStories"`
	TotalEstimate       int `json:"totalEstimate"`
	OverdueStories      int `json:"overdueStories"`
	UrgentStories       int `json:"urgentStories"`
	HighPriorityStories int `json:"highPriorityStories"`
	UnestimatedStories  int `json:"unestimatedStories"`
	UnassignedStories   int `json:"unassignedStories"`
}

type AppMemberWorkload struct {
	UserID              uuid.UUID `json:"userId"`
	FullName            string    `json:"fullName"`
	Username            string    `json:"username"`
	AvatarURL           string    `json:"avatarUrl"`
	OpenStories         int       `json:"openStories"`
	StartedStories      int       `json:"startedStories"`
	PausedStories       int       `json:"pausedStories"`
	CompletedStories    int       `json:"completedStories"`
	OverdueStories      int       `json:"overdueStories"`
	UrgentStories       int       `json:"urgentStories"`
	HighPriorityStories int       `json:"highPriorityStories"`
	UnestimatedStories  int       `json:"unestimatedStories"`
	EstimateTotal       int       `json:"estimateTotal"`
}

type AppTeamWorkloadSummary struct {
	TeamID             uuid.UUID `json:"teamId"`
	TeamName           string    `json:"teamName"`
	TeamCode           string    `json:"teamCode"`
	OpenStories        int       `json:"openStories"`
	EstimateTotal      int       `json:"estimateTotal"`
	OverdueStories     int       `json:"overdueStories"`
	UnassignedStories  int       `json:"unassignedStories"`
	UnestimatedStories int       `json:"unestimatedStories"`
}

type AppUnassignedWorkload struct {
	Stories             int `json:"stories"`
	EstimateTotal       int `json:"estimateTotal"`
	OverdueStories      int `json:"overdueStories"`
	UrgentStories       int `json:"urgentStories"`
	HighPriorityStories int `json:"highPriorityStories"`
	UnestimatedStories  int `json:"unestimatedStories"`
}

type AppWorkloadRisks struct {
	OverloadedMembers   []AppMemberWorkload `json:"overloadedMembers"`
	OverdueMembers      []AppMemberWorkload `json:"overdueMembers"`
	UnassignedStories   int                 `json:"unassignedStories"`
	UnestimatedStories  int                 `json:"unestimatedStories"`
	HighPriorityStories int                 `json:"highPriorityStories"`
}

func toAppWorkloadAnalysis(analysis reports.CoreWorkloadAnalysis) AppWorkloadAnalysis {
	return AppWorkloadAnalysis{
		Summary:    toAppWorkloadSummary(analysis.Summary),
		Members:    toAppMemberWorkloads(analysis.Members),
		Teams:      toAppTeamWorkloadSummaries(analysis.Teams),
		Unassigned: toAppUnassignedWorkload(analysis.Unassigned),
		Risks:      toAppWorkloadRisks(analysis.Risks),
	}
}

func toAppWorkloadSummary(summary reports.CoreWorkloadSummary) AppWorkloadSummary {
	return AppWorkloadSummary{
		TotalOpenStories:    summary.TotalOpenStories,
		TotalEstimate:       summary.TotalEstimate,
		OverdueStories:      summary.OverdueStories,
		UrgentStories:       summary.UrgentStories,
		HighPriorityStories: summary.HighPriorityStories,
		UnestimatedStories:  summary.UnestimatedStories,
		UnassignedStories:   summary.UnassignedStories,
	}
}

func toAppMemberWorkloads(members []reports.CoreMemberWorkload) []AppMemberWorkload {
	result := make([]AppMemberWorkload, len(members))
	for i, member := range members {
		result[i] = toAppMemberWorkload(member)
	}
	return result
}

func toAppMemberWorkload(member reports.CoreMemberWorkload) AppMemberWorkload {
	return AppMemberWorkload{
		UserID:              member.UserID,
		FullName:            member.FullName,
		Username:            member.Username,
		AvatarURL:           member.AvatarURL,
		OpenStories:         member.OpenStories,
		StartedStories:      member.StartedStories,
		PausedStories:       member.PausedStories,
		CompletedStories:    member.CompletedStories,
		OverdueStories:      member.OverdueStories,
		UrgentStories:       member.UrgentStories,
		HighPriorityStories: member.HighPriorityStories,
		UnestimatedStories:  member.UnestimatedStories,
		EstimateTotal:       member.EstimateTotal,
	}
}

func toAppTeamWorkloadSummaries(teams []reports.CoreTeamWorkloadSummary) []AppTeamWorkloadSummary {
	result := make([]AppTeamWorkloadSummary, len(teams))
	for i, team := range teams {
		result[i] = AppTeamWorkloadSummary{
			TeamID:             team.TeamID,
			TeamName:           team.TeamName,
			TeamCode:           team.TeamCode,
			OpenStories:        team.OpenStories,
			EstimateTotal:      team.EstimateTotal,
			OverdueStories:     team.OverdueStories,
			UnassignedStories:  team.UnassignedStories,
			UnestimatedStories: team.UnestimatedStories,
		}
	}
	return result
}

func toAppUnassignedWorkload(unassigned reports.CoreUnassignedWorkload) AppUnassignedWorkload {
	return AppUnassignedWorkload{
		Stories:             unassigned.Stories,
		EstimateTotal:       unassigned.EstimateTotal,
		OverdueStories:      unassigned.OverdueStories,
		UrgentStories:       unassigned.UrgentStories,
		HighPriorityStories: unassigned.HighPriorityStories,
		UnestimatedStories:  unassigned.UnestimatedStories,
	}
}

func toAppWorkloadRisks(risks reports.CoreWorkloadRisks) AppWorkloadRisks {
	return AppWorkloadRisks{
		OverloadedMembers:   toAppMemberWorkloads(risks.OverloadedMembers),
		OverdueMembers:      toAppMemberWorkloads(risks.OverdueMembers),
		UnassignedStories:   risks.UnassignedStories,
		UnestimatedStories:  risks.UnestimatedStories,
		HighPriorityStories: risks.HighPriorityStories,
	}
}
