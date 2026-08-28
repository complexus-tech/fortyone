package reportsrepository

import (
	"context"
	"fmt"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetStoryStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StoryStatsFilters) (reports.CoreStoryStats, error) {
	if err := r.authorize(ctx, filters.ActorID, workspaceID); err != nil {
		return reports.CoreStoryStats{}, err
	}

	row, err := r.queries.GetStoryStats(ctx, reportssql.GetStoryStatsParams{
		ActorID:     filters.ActorID,
		WorkspaceID: workspaceID,
		StartDate:   filters.StartDate,
		EndDate:     filters.EndDate,
	})
	if err != nil {
		return reports.CoreStoryStats{}, fmt.Errorf("selecting story statistics: %w", err)
	}

	return reports.CoreStoryStats{
		Closed:     int(row.Closed),
		Overdue:    int(row.Overdue),
		InProgress: int(row.InProgress),
		Created:    int(row.Created),
		Assigned:   int(row.Assigned),
	}, nil
}

func (r *repo) GetContributionStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, startDate time.Time, endDate time.Time) ([]reports.CoreContributionStats, error) {
	if err := r.authorize(ctx, userID, workspaceID); err != nil {
		return nil, err
	}

	rows, err := r.queries.GetContributionStats(ctx, reportssql.GetContributionStatsParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("selecting contribution statistics: %w", err)
	}

	result := make([]reports.CoreContributionStats, len(rows))
	for i, row := range rows {
		result[i] = reports.CoreContributionStats{
			Date:          row.Date,
			Contributions: int(row.Contributions),
		}
	}

	return result, nil
}

func (r *repo) GetUserStats(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (reports.CoreUserStats, error) {
	if err := r.authorize(ctx, userID, workspaceID); err != nil {
		return reports.CoreUserStats{}, err
	}

	row, err := r.queries.GetUserStats(ctx, reportssql.GetUserStatsParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return reports.CoreUserStats{}, fmt.Errorf("selecting user statistics: %w", err)
	}

	return reports.CoreUserStats{
		AssignedToMe: int(row.AssignedToMe),
		CreatedByMe:  int(row.CreatedByMe),
	}, nil
}

func (r *repo) GetStatusStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StatsFilters) ([]reports.CoreStatusStats, error) {
	if err := r.validateRequestedTeam(ctx, filters.ActorID, workspaceID, filters.TeamID); err != nil {
		return nil, err
	}

	rows, err := r.queries.GetStatusStats(ctx, reportssql.GetStatusStatsParams{
		WorkspaceID: workspaceID,
		ActorID:     filters.ActorID,
		TeamID:      filters.TeamID,
		SprintID:    filters.SprintID,
		ObjectiveID: filters.ObjectiveID,
		StartDate:   filters.StartDate,
		EndDate:     filters.EndDate,
	})
	if err != nil {
		return nil, fmt.Errorf("selecting status statistics: %w", err)
	}

	result := make([]reports.CoreStatusStats, len(rows))
	for i, row := range rows {
		result[i] = reports.CoreStatusStats{Name: row.Name, Count: int(row.Count)}
	}

	return result, nil
}

func (r *repo) GetPriorityStats(ctx context.Context, workspaceID uuid.UUID, filters reports.StatsFilters) ([]reports.CorePriorityStats, error) {
	if err := r.validateRequestedTeam(ctx, filters.ActorID, workspaceID, filters.TeamID); err != nil {
		return nil, err
	}

	rows, err := r.queries.GetPriorityStats(ctx, reportssql.GetPriorityStatsParams{
		WorkspaceID: workspaceID,
		ActorID:     filters.ActorID,
		TeamID:      filters.TeamID,
		SprintID:    filters.SprintID,
		ObjectiveID: filters.ObjectiveID,
		StartDate:   filters.StartDate,
		EndDate:     filters.EndDate,
	})
	if err != nil {
		return nil, fmt.Errorf("selecting priority statistics: %w", err)
	}

	result := make([]reports.CorePriorityStats, len(rows))
	for i, row := range rows {
		result[i] = reports.CorePriorityStats{Priority: row.Priority, Count: int(row.Count)}
	}

	return result, nil
}
