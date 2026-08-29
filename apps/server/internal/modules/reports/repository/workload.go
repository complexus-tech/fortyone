package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetWorkloadAnalysis(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreWorkloadAnalysis, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreWorkloadAnalysis{}, err
	}
	summary, err := r.queries.GetWorkloadSummary(ctx, reportssql.GetWorkloadSummaryParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("selecting workload summary: %w", err)
	}

	memberRows, err := r.queries.ListMemberWorkload(ctx, reportssql.ListMemberWorkloadParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("selecting member workload: %w", err)
	}

	teamRows, err := r.queries.ListTeamWorkloadSummary(ctx, reportssql.ListTeamWorkloadSummaryParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("selecting team workload summary: %w", err)
	}

	unassigned, err := r.queries.GetUnassignedWorkload(ctx, reportssql.GetUnassignedWorkloadParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("selecting unassigned workload: %w", err)
	}

	members := make([]reportdomain.CoreMemberWorkload, len(memberRows))
	for i, row := range memberRows {
		if row.UserID == uuid.Nil {
			return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("mapping member workload: %w", ErrInvalidProjection)
		}
		lastActivity := row.LastStoryActivityAt
		members[i] = reportdomain.CoreMemberWorkload{
			UserID:                row.UserID,
			FullName:              row.FullName,
			Username:              row.Username,
			AvatarURL:             row.AvatarURL,
			TeamAIRoleTitle:       row.TeamAiRoleTitle,
			TeamAIRoleDescription: row.TeamAiRoleDescription,
			OpenStories:           int(row.OpenStories),
			StartedStories:        int(row.StartedStories),
			PausedStories:         int(row.PausedStories),
			CompletedStories:      int(row.CompletedStories),
			OverdueStories:        int(row.OverdueStories),
			UrgentStories:         int(row.UrgentStories),
			HighPriorityStories:   int(row.HighPriorityStories),
			UnestimatedStories:    int(row.UnestimatedStories),
			EstimateTotal:         int(row.EstimateTotal),
			LastStoryActivityAt:   &lastActivity,
		}
	}

	teams := make([]reportdomain.CoreTeamWorkloadSummary, len(teamRows))
	for i, row := range teamRows {
		if row.TeamID == uuid.Nil {
			return reportdomain.CoreWorkloadAnalysis{}, fmt.Errorf("mapping team workload: %w", ErrInvalidProjection)
		}
		teams[i] = reportdomain.CoreTeamWorkloadSummary{
			TeamID:             row.TeamID,
			TeamName:           row.TeamName,
			TeamCode:           row.TeamCode,
			OpenStories:        int(row.OpenStories),
			EstimateTotal:      int(row.EstimateTotal),
			OverdueStories:     int(row.OverdueStories),
			UnassignedStories:  int(row.UnassignedStories),
			UnestimatedStories: int(row.UnestimatedStories),
		}
	}

	return reportdomain.CoreWorkloadAnalysis{
		Summary: reportdomain.CoreWorkloadSummary{
			TotalOpenStories:    int(summary.TotalOpenStories),
			TotalEstimate:       int(summary.TotalEstimate),
			OverdueStories:      int(summary.OverdueStories),
			UrgentStories:       int(summary.UrgentStories),
			HighPriorityStories: int(summary.HighPriorityStories),
			UnestimatedStories:  int(summary.UnestimatedStories),
			UnassignedStories:   int(summary.UnassignedStories),
		},
		Members: members,
		Teams:   teams,
		Unassigned: reportdomain.CoreUnassignedWorkload{
			Stories:             int(unassigned.Stories),
			EstimateTotal:       int(unassigned.EstimateTotal),
			OverdueStories:      int(unassigned.OverdueStories),
			UrgentStories:       int(unassigned.UrgentStories),
			HighPriorityStories: int(unassigned.HighPriorityStories),
			UnestimatedStories:  int(unassigned.UnestimatedStories),
		},
	}, nil
}
