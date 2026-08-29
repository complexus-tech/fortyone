package sprintsrepository

import (
	"context"
	"fmt"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprintssql "github.com/complexus-tech/projects-api/internal/modules/sprints/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/internal/platform/workweek"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (repository *Repository) GetAnalytics(
	ctx context.Context,
	sprintID, workspaceID, actorID uuid.UUID,
	now time.Time,
) (sprintdomain.Analytics, error) {
	sprint, err := repository.GetByID(ctx, sprintID, workspaceID, actorID)
	if err != nil {
		return sprintdomain.Analytics{}, err
	}

	var (
		workingDays []int
		breakdown   sprintdomain.StoryBreakdown
		changes     []sprintdomain.BurndownChange
		allocation  []sprintdomain.TeamMemberAllocation
	)
	group, groupContext := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		workingDays, err = repository.getWorkingDays(groupContext, workspaceID, actorID)
		return err
	})
	group.Go(func() error {
		var err error
		breakdown, err = repository.getStoryBreakdown(groupContext, sprintID, workspaceID, actorID)
		return err
	})
	group.Go(func() error {
		var err error
		changes, err = repository.getBurndownChanges(groupContext, sprint, actorID)
		return err
	})
	group.Go(func() error {
		var err error
		allocation, err = repository.getTeamAllocation(groupContext, sprint, actorID)
		return err
	})
	if err := group.Wait(); err != nil {
		return sprintdomain.Analytics{}, err
	}

	return sprintdomain.Analytics{
		SprintID: sprintID, WorkingDays: workingDays,
		Overview:       sprintdomain.CalculateOverview(sprint, breakdown, workingDays, now),
		StoryBreakdown: breakdown,
		Burndown:       sprintdomain.BuildBurndown(changes, sprint.StartDate, sprint.EndDate, workingDays, now),
		TeamAllocation: allocation,
	}, nil
}

func (repository *Repository) getWorkingDays(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
) ([]int, error) {
	stored, err := repository.queries.GetWorkspaceWorkingDays(ctx, sprintssql.GetWorkspaceWorkingDaysParams{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return nil, fmt.Errorf("get workspace working days: %w", err)
	}
	days := make([]int, len(stored))
	for index, day := range stored {
		days[index] = int(day)
	}
	return workweek.Normalize(days), nil
}

func (repository *Repository) getStoryBreakdown(
	ctx context.Context,
	sprintID, workspaceID, actorID uuid.UUID,
) (sprintdomain.StoryBreakdown, error) {
	row, err := repository.queries.GetSprintStoryBreakdown(ctx, sprintssql.GetSprintStoryBreakdownParams{
		SprintID: sprintID, WorkspaceID: workspaceID, ActorID: actorID,
	})
	if err != nil {
		return sprintdomain.StoryBreakdown{}, fmt.Errorf("get sprint story breakdown: %w", err)
	}
	values := []int64{row.Total, row.Completed, row.InProgress, row.Todo, row.Blocked, row.Cancelled}
	counts := make([]int, len(values))
	for index, value := range values {
		counts[index], err = safecast.Int64(value)
		if err != nil {
			return sprintdomain.StoryBreakdown{}, fmt.Errorf("map story breakdown count: %w", err)
		}
	}
	return sprintdomain.StoryBreakdown{
		Total: counts[0], Completed: counts[1], InProgress: counts[2],
		Todo: counts[3], Blocked: counts[4], Cancelled: counts[5],
	}, nil
}

func (repository *Repository) getBurndownChanges(
	ctx context.Context,
	sprint sprintdomain.Sprint,
	actorID uuid.UUID,
) ([]sprintdomain.BurndownChange, error) {
	rows, err := repository.queries.ListSprintBurndownChanges(ctx, sprintssql.ListSprintBurndownChangesParams{
		WorkspaceID: sprint.WorkspaceID, SprintID: sprint.ID,
		StartDate: sprint.StartDate, EndDate: sprint.EndDate,
		ActorID: actorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list sprint burndown changes: %w", err)
	}
	changes := make([]sprintdomain.BurndownChange, 0, len(rows))
	for _, row := range rows {
		values := []int64{row.ScopeDelta, row.CompletionDelta, row.InitialStories, row.InitialCompleted}
		counts := [4]int{}
		for index, value := range values {
			counts[index], err = safecast.Int64(value)
			if err != nil {
				return nil, fmt.Errorf("map burndown count: %w", err)
			}
		}
		changes = append(changes, sprintdomain.BurndownChange{
			Date: row.EventDate, ScopeDelta: counts[0], CompletionDelta: counts[1],
			InitialStories: counts[2], InitialCompleted: counts[3],
		})
	}
	return changes, nil
}

func (repository *Repository) getTeamAllocation(
	ctx context.Context,
	sprint sprintdomain.Sprint,
	actorID uuid.UUID,
) ([]sprintdomain.TeamMemberAllocation, error) {
	rows, err := repository.queries.ListSprintTeamAllocation(ctx, sprintssql.ListSprintTeamAllocationParams{
		SprintID: sprint.ID, WorkspaceID: sprint.WorkspaceID, TeamID: sprint.TeamID,
		ActorID: actorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list sprint team allocation: %w", err)
	}
	allocation := make([]sprintdomain.TeamMemberAllocation, 0, len(rows))
	for _, row := range rows {
		assigned, err := safecast.Int64(row.Assigned)
		if err != nil {
			return nil, fmt.Errorf("map assigned count: %w", err)
		}
		completed, err := safecast.Int64(row.Completed)
		if err != nil {
			return nil, fmt.Errorf("map completed count: %w", err)
		}
		allocation = append(allocation, sprintdomain.TeamMemberAllocation{
			MemberID: row.UserID, Username: row.Username, AvatarURL: row.AvatarURL,
			Assigned: assigned, Completed: completed,
		})
	}
	return allocation, nil
}
