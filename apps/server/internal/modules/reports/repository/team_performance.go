package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetTeamPerformance(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CoreTeamPerformance, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, err
	}
	startDate, endDate, err := query.requiredDates()
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, err
	}

	workloadRows, err := r.queries.ListTeamWorkload(ctx, reportssql.ListTeamWorkloadParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, fmt.Errorf("selecting team workload: %w", err)
	}

	contributionRows, err := r.queries.ListMemberContributions(ctx, reportssql.ListMemberContributionsParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, fmt.Errorf("selecting member contributions: %w", err)
	}

	velocityRows, err := r.queries.ListTeamVelocity(ctx, reportssql.ListTeamVelocityParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, fmt.Errorf("selecting team velocity: %w", err)
	}

	trendRows, err := r.queries.ListWorkloadTrend(ctx, reportssql.ListWorkloadTrendParams{
		WorkspaceID: query.workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
		TeamIds:     query.teamIDs,
	})
	if err != nil {
		return reportdomain.CoreTeamPerformance{}, fmt.Errorf("selecting workload trend: %w", err)
	}

	result := reportdomain.CoreTeamPerformance{
		TeamWorkload:        make([]reportdomain.CoreTeamWorkloadItem, len(workloadRows)),
		MemberContributions: make([]reportdomain.CoreMemberContributionItem, len(contributionRows)),
		VelocityByTeam:      make([]reportdomain.CoreTeamVelocityItem, len(velocityRows)),
		WorkloadTrend:       make([]reportdomain.CoreWorkloadTrendPoint, len(trendRows)),
	}
	for i, row := range workloadRows {
		result.TeamWorkload[i] = reportdomain.CoreTeamWorkloadItem{
			TeamID:    row.TeamID,
			TeamName:  row.TeamName,
			Assigned:  int(row.Assigned),
			Completed: int(row.Completed),
			Capacity:  int(row.Capacity),
		}
	}
	for i, row := range contributionRows {
		result.MemberContributions[i] = reportdomain.CoreMemberContributionItem{
			UserID:    row.UserID,
			Username:  row.Username,
			AvatarURL: row.AvatarURL,
			TeamID:    row.TeamID,
			Assigned:  int(row.Assigned),
			Completed: int(row.Completed),
		}
	}
	for i, row := range velocityRows {
		result.VelocityByTeam[i] = reportdomain.CoreTeamVelocityItem{
			TeamID:   row.TeamID,
			TeamName: row.TeamName,
			Week1:    int(row.Week1),
			Week2:    int(row.Week2),
			Week3:    int(row.Week3),
			Average:  row.Average,
		}
	}
	for i, row := range trendRows {
		result.WorkloadTrend[i] = reportdomain.CoreWorkloadTrendPoint{
			Date:      row.Date,
			Assigned:  int(row.Assigned),
			Completed: int(row.Completed),
		}
	}

	return result, nil
}
