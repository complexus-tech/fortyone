package reportsrepository

import (
	"context"
	"fmt"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportssql "github.com/complexus-tech/projects-api/internal/modules/reports/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) GetPulseStoryHealth(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CorePulseStoryHealth, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CorePulseStoryHealth{}, err
	}
	row, err := r.queries.GetPulseStoryHealth(ctx, reportssql.GetPulseStoryHealthParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CorePulseStoryHealth{}, fmt.Errorf("selecting pulse story health: %w", err)
	}

	return reportdomain.CorePulseStoryHealth{
		OpenStories:         int(row.OpenStories),
		StartedStories:      int(row.StartedStories),
		PausedStories:       int(row.PausedStories),
		CompletedStories:    int(row.CompletedStories),
		CancelledStories:    int(row.CancelledStories),
		BlockedStories:      int(row.BlockedStories),
		OverdueStories:      int(row.OverdueStories),
		UrgentStories:       int(row.UrgentStories),
		HighPriorityStories: int(row.HighPriorityStories),
		UnassignedStories:   int(row.UnassignedStories),
		UnestimatedStories:  int(row.UnestimatedStories),
	}, nil
}

func (r *repo) GetPulseSprintHealth(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CorePulseSprintHealth, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CorePulseSprintHealth{}, err
	}
	row, err := r.queries.GetPulseSprintHealth(ctx, reportssql.GetPulseSprintHealthParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CorePulseSprintHealth{}, fmt.Errorf("selecting pulse sprint health: %w", err)
	}

	return reportdomain.CorePulseSprintHealth{
		ActiveSprints:      int(row.ActiveSprints),
		UpcomingSprints:    int(row.UpcomingSprints),
		CompletedSprints:   int(row.CompletedSprints),
		AtRiskSprints:      int(row.AtRiskSprints),
		OverdueSprints:     int(row.OverdueSprints),
		UnestimatedStories: int(row.UnestimatedStories),
	}, nil
}

func (r *repo) GetPulseObjectiveHealth(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CorePulseObjectiveHealth, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CorePulseObjectiveHealth{}, err
	}
	row, err := r.queries.GetPulseObjectiveHealth(ctx, reportssql.GetPulseObjectiveHealthParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CorePulseObjectiveHealth{}, fmt.Errorf("selecting pulse objective health: %w", err)
	}

	return reportdomain.CorePulseObjectiveHealth{
		ActiveObjectives:   int(row.ActiveObjectives),
		AtRiskObjectives:   int(row.AtRiskObjectives),
		OffTrackObjectives: int(row.OffTrackObjectives),
		OverdueObjectives:  int(row.OverdueObjectives),
		ObjectivesDueSoon:  int(row.ObjectivesDueSoon),
	}, nil
}

func (r *repo) GetPulseRequestHealth(ctx context.Context, workspaceID uuid.UUID, filters reportdomain.ReportFilters) (reportdomain.CorePulseRequestHealth, error) {
	query, err := r.scopedQueryFilters(ctx, workspaceID, filters)
	if err != nil {
		return reportdomain.CorePulseRequestHealth{}, err
	}
	row, err := r.queries.GetPulseRequestHealth(ctx, reportssql.GetPulseRequestHealthParams{
		WorkspaceID:  query.workspaceID,
		TeamIds:      query.teamIDs,
		AssigneeIds:  query.assigneeIDs,
		SprintIds:    query.sprintIDs,
		ObjectiveIds: query.objectiveIDs,
		StartDate:    query.startDate,
		EndDate:      query.endDate,
	})
	if err != nil {
		return reportdomain.CorePulseRequestHealth{}, fmt.Errorf("selecting pulse request health: %w", err)
	}

	return reportdomain.CorePulseRequestHealth{
		PendingRequests:  int(row.PendingRequests),
		UrgentRequests:   int(row.UrgentRequests),
		HighRequests:     int(row.HighRequests),
		GitHubRequests:   int(row.GithubRequests),
		SlackRequests:    int(row.SlackRequests),
		IntercomRequests: int(row.IntercomRequests),
		StaleRequests:    int(row.StaleRequests),
	}, nil
}
