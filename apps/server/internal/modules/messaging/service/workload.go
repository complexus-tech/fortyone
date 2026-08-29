package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	"github.com/google/uuid"
)

// WorkloadService is the bounded reports surface used by the assistant to
// answer team workload questions.
type WorkloadService interface {
	GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreWorkloadAnalysis, error)
}

const (
	overloadedOpenStoriesThreshold = 8
	overloadedEstimateThreshold    = 20
)

func workloadToolDefinitions() []ToolDefinition {
	return []ToolDefinition{{
		Type:        "function",
		Name:        toolGetWorkloadSummary,
		Description: "Assess workload across shared teams and identify members who meet FortyOne's overloaded thresholds. Optionally check one member by name.",
		Strict:      true,
		Parameters: strictObjectSchema(map[string]any{
			"team_name": map[string]any{
				"type":        []string{"string", "null"},
				"description": "Optional accessible team name or code, or null for all shared teams.",
				"maxLength":   maxSearchRunes,
			},
			"member_name": map[string]any{
				"type":        []string{"string", "null"},
				"description": "Optional member display name or username to check, or null for the whole team.",
				"maxLength":   maxSearchRunes,
			},
			"limit": map[string]any{
				"type":        []string{"integer", "null"},
				"description": "Maximum overloaded members to return, or null for the default.",
				"minimum":     1,
				"maximum":     maxToolLimit,
			},
		}, []string{"team_name", "member_name", "limit"}),
	}}
}

func (e *FortyOneToolExecutor) getWorkloadSummary(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamName   *string `json:"team_name"`
		MemberName *string `json:"member_name"`
		Limit      *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_name", "member_name", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	joined, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	selectedTeam, err := teamByName(args.TeamName, joined)
	if err != nil {
		return nil, err
	}
	sharedTeamIDs := sharedAccessibleTeamIDs(scope.SharedTeamIDs, joinedByID)
	if selectedTeam != nil {
		if !teamWorkSharedTeamAllowed(scope.SharedTeamIDs, selectedTeam.ID) {
			return marshalToolResult(workloadSummaryResult{
				Access:       "denied",
				AccessReason: "shared_team_scope_required",
				TeamName:     selectedTeam.Name,
				Thresholds:   workloadThresholdResult{OpenStories: overloadedOpenStoriesThreshold, EstimateUnits: overloadedEstimateThreshold},
			})
		}
		sharedTeamIDs = []uuid.UUID{selectedTeam.ID}
	}
	if len(sharedTeamIDs) == 0 {
		return marshalToolResult(workloadSummaryResult{
			Access:       "denied",
			AccessReason: "shared_team_scope_required",
			Thresholds:   workloadThresholdResult{OpenStories: overloadedOpenStoriesThreshold, EstimateUnits: overloadedEstimateThreshold},
		})
	}

	analysis, err := e.workload.GetWorkloadAnalysis(ctx, scope.WorkspaceID, reportdomain.ReportFilters{TeamIDs: sharedTeamIDs})
	if err != nil {
		return nil, fmt.Errorf("get workload summary: %w", err)
	}
	overloaded := make([]workloadMemberResult, 0, min(len(analysis.Risks.OverloadedMembers), limit))
	for _, member := range analysis.Risks.OverloadedMembers {
		if len(overloaded) == limit {
			break
		}
		overloaded = append(overloaded, newWorkloadMemberResult(member))
	}

	result := workloadSummaryResult{
		Access:          "granted",
		TeamName:        workloadTeamName(selectedTeam, len(sharedTeamIDs)),
		TotalMembers:    len(analysis.Members),
		OverloadedCount: len(analysis.Risks.OverloadedMembers),
		Truncated:       len(analysis.Risks.OverloadedMembers) > len(overloaded),
		Overloaded:      overloaded,
		Summary: workloadSummaryMetrics{
			OpenStories:        analysis.Summary.TotalOpenStories,
			EstimateUnits:      analysis.Summary.TotalEstimate,
			OverdueStories:     analysis.Summary.OverdueStories,
			UnassignedStories:  analysis.Summary.UnassignedStories,
			UnestimatedStories: analysis.Summary.UnestimatedStories,
		},
		Thresholds: workloadThresholdResult{
			OpenStories:   overloadedOpenStoriesThreshold,
			EstimateUnits: overloadedEstimateThreshold,
		},
	}
	if args.MemberName != nil && strings.TrimSpace(*args.MemberName) != "" {
		member, found, err := resolveWorkloadMember(analysis.Members, *args.MemberName)
		if err != nil {
			return nil, err
		}
		result.MemberQuery = strings.TrimSpace(*args.MemberName)
		result.MemberFound = found
		if found {
			memberResult := newWorkloadMemberResult(member)
			memberResult.Overloaded = member.OpenStories >= overloadedOpenStoriesThreshold || member.EstimateTotal >= overloadedEstimateThreshold
			result.Member = &memberResult
		}
	}
	return marshalToolResult(result)
}

func sharedAccessibleTeamIDs(sharedTeamIDs []uuid.UUID, joinedByID map[uuid.UUID]teamsdomain.Team) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(sharedTeamIDs))
	seen := make(map[uuid.UUID]struct{}, len(sharedTeamIDs))
	for _, teamID := range sharedTeamIDs {
		if _, joined := joinedByID[teamID]; !joined {
			continue
		}
		if _, exists := seen[teamID]; exists {
			continue
		}
		seen[teamID] = struct{}{}
		result = append(result, teamID)
	}
	return result
}

func resolveWorkloadMember(members []reportdomain.CoreMemberWorkload, query string) (reportdomain.CoreMemberWorkload, bool, error) {
	wanted := normalizePlanningName(query)
	exact := make([]reportdomain.CoreMemberWorkload, 0, 1)
	contains := make([]reportdomain.CoreMemberWorkload, 0, 1)
	for _, member := range members {
		name := normalizePlanningName(member.FullName)
		username := normalizePlanningName(member.Username)
		if name == wanted || username == wanted {
			exact = append(exact, member)
			continue
		}
		if strings.Contains(name, wanted) || strings.Contains(username, wanted) {
			contains = append(contains, member)
		}
	}
	if len(exact) == 1 {
		return exact[0], true, nil
	}
	if len(exact) > 1 || len(contains) > 1 {
		return reportdomain.CoreMemberWorkload{}, false, fmt.Errorf("member %q is ambiguous; include a more specific name", strings.TrimSpace(query))
	}
	if len(contains) == 1 {
		return contains[0], true, nil
	}
	return reportdomain.CoreMemberWorkload{}, false, nil
}

func workloadTeamName(team *teamsdomain.Team, sharedTeamCount int) string {
	if team != nil {
		return team.Name
	}
	if sharedTeamCount == 1 {
		return "shared team"
	}
	return "shared teams"
}

func newWorkloadMemberResult(member reportdomain.CoreMemberWorkload) workloadMemberResult {
	return workloadMemberResult{
		Name:                member.FullName,
		Username:            member.Username,
		OpenStories:         member.OpenStories,
		StartedStories:      member.StartedStories,
		PausedStories:       member.PausedStories,
		EstimateUnits:       member.EstimateTotal,
		OverdueStories:      member.OverdueStories,
		UrgentStories:       member.UrgentStories,
		HighPriorityStories: member.HighPriorityStories,
		UnestimatedStories:  member.UnestimatedStories,
	}
}

type workloadSummaryResult struct {
	Access          string                  `json:"access"`
	AccessReason    string                  `json:"access_reason,omitempty"`
	TeamName        string                  `json:"team_name,omitempty"`
	TotalMembers    int                     `json:"total_members,omitempty"`
	OverloadedCount int                     `json:"overloaded_count,omitempty"`
	Truncated       bool                    `json:"truncated,omitempty"`
	Overloaded      []workloadMemberResult  `json:"overloaded,omitempty"`
	MemberQuery     string                  `json:"member_query,omitempty"`
	MemberFound     bool                    `json:"member_found,omitempty"`
	Member          *workloadMemberResult   `json:"member,omitempty"`
	Summary         workloadSummaryMetrics  `json:"summary,omitempty"`
	Thresholds      workloadThresholdResult `json:"thresholds"`
}

type workloadSummaryMetrics struct {
	OpenStories        int `json:"open_stories"`
	EstimateUnits      int `json:"estimate_units"`
	OverdueStories     int `json:"overdue_stories"`
	UnassignedStories  int `json:"unassigned_stories"`
	UnestimatedStories int `json:"unestimated_stories"`
}

type workloadThresholdResult struct {
	OpenStories   int `json:"open_stories"`
	EstimateUnits int `json:"estimate_units"`
}

type workloadMemberResult struct {
	Name                string `json:"name"`
	Username            string `json:"username"`
	Overloaded          bool   `json:"overloaded"`
	OpenStories         int    `json:"open_stories"`
	StartedStories      int    `json:"started_stories"`
	PausedStories       int    `json:"paused_stories"`
	EstimateUnits       int    `json:"estimate_units"`
	OverdueStories      int    `json:"overdue_stories"`
	UrgentStories       int    `json:"urgent_stories"`
	HighPriorityStories int    `json:"high_priority_stories"`
	UnestimatedStories  int    `json:"unestimated_stories"`
}
