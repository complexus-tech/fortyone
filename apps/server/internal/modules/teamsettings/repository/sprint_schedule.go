package teamsettingsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduledSprint struct {
	ID        uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

const maxManagedSprintScheduleRows = 100

type sprintScheduleAuditWriter func(
	context.Context,
	teamsettingssql.Querier,
	teamsettings.CoreTeamSprintSettings,
	scheduledSprint,
	time.Time,
	time.Time,
) error

func (r *repo) ReconcileSprintSchedule(
	ctx context.Context,
	settings teamsettings.CoreTeamSprintSettings,
	actor teamsettings.AuditActor,
) (int, error) {
	updated := 0
	err := r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		locked, err := queries.LockSprintSettings(ctx, teamsettingssql.LockSprintSettingsParams{
			TeamID: settings.TeamID, WorkspaceID: settings.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("lock sprint settings: %w", err)
		}
		updated, err = reconcileSprintSchedule(ctx, queries, mapLockedSprintSettings(locked), actor)
		return err
	})
	return updated, err
}

func reconcileSprintSchedule(
	ctx context.Context,
	queries teamsettingssql.Querier,
	settings teamsettings.CoreTeamSprintSettings,
	actor teamsettings.AuditActor,
) (int, error) {
	scheduleDate, err := queries.GetDatabaseDate(ctx)
	if err != nil {
		return 0, fmt.Errorf("get database schedule date: %w", err)
	}
	changedAt := time.Now().UTC()
	return reconcileSprintScheduleAt(
		ctx,
		queries,
		settings,
		scheduleDate,
		changedAt,
		func(
			ctx context.Context,
			queries teamsettingssql.Querier,
			settings teamsettings.CoreTeamSprintSettings,
			sprint scheduledSprint,
			startDate, endDate time.Time,
		) error {
			return insertAuditEvent(
				ctx, queries, settings.WorkspaceID, settings.TeamID, actor,
				"sprint", sprint.ID, "sprint.auto_rescheduled",
				scheduleAuditMetadata(settings, sprint, startDate, endDate),
			)
		},
	)
}

func reconcileSprintScheduleAt(
	ctx context.Context,
	queries teamsettingssql.Querier,
	settings teamsettings.CoreTeamSprintSettings,
	scheduleDate, changedAt time.Time,
	writeAudit sprintScheduleAuditWriter,
) (int, error) {
	managedRows, err := queries.ListManagedFutureSprintsForUpdate(ctx, teamsettingssql.ListManagedFutureSprintsForUpdateParams{
		TeamID: settings.TeamID, WorkspaceID: settings.WorkspaceID, ScheduleDate: scheduleDate,
		RowLimit: maxManagedSprintScheduleRows + 1,
	})
	if err != nil {
		return 0, fmt.Errorf("list automation-managed future sprints: %w", err)
	}
	if len(managedRows) > maxManagedSprintScheduleRows {
		return 0, teamsettings.ErrSprintScheduleTooLarge
	}
	if len(managedRows) == 0 {
		return 0, nil
	}
	managed := make([]scheduledSprint, len(managedRows))
	for index, row := range managedRows {
		managed[index] = scheduledSprint{ID: row.SprintID, Name: row.Name, StartDate: row.StartDate, EndDate: row.EndDate}
	}

	anchorDate := scheduleDate
	activeEnd, err := queries.GetActiveSprintEnd(ctx, teamsettingssql.GetActiveSprintEndParams{
		TeamID: settings.TeamID, WorkspaceID: settings.WorkspaceID, ScheduleDate: scheduleDate,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("get active sprint schedule boundary: %w", err)
	}
	if err == nil && activeEnd.After(anchorDate) {
		anchorDate = activeEnd
	}

	nextStart := nextSprintStartAfter(anchorDate, settings.SprintStartDay)
	updatedCount := 0
	for _, sprint := range managed {
		nextEnd := nextStart.AddDate(0, 0, settings.SprintDurationWeeks*7-1)
		conflict, err := queries.FindCustomSprintScheduleConflict(ctx, teamsettingssql.FindCustomSprintScheduleConflictParams{
			TeamID: settings.TeamID, WorkspaceID: settings.WorkspaceID,
			ScheduleDate: scheduleDate, ProposedStartDate: nextStart, ProposedEndDate: nextEnd,
		})
		if err == nil {
			return 0, fmt.Errorf("%w: %s", teamsettings.ErrSprintScheduleConflict, conflict.Name)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("find custom sprint schedule conflict: %w", err)
		}
		if sprint.StartDate.Equal(nextStart) && sprint.EndDate.Equal(nextEnd) {
			nextStart = nextEnd.AddDate(0, 0, 1)
			continue
		}
		rowsAffected, err := queries.UpdateManagedSprintSchedule(ctx, teamsettingssql.UpdateManagedSprintScheduleParams{
			StartDate: nextStart, EndDate: nextEnd, SprintID: sprint.ID,
			TeamID: settings.TeamID, WorkspaceID: settings.WorkspaceID, UpdatedAt: changedAt,
		})
		if err != nil {
			return 0, fmt.Errorf("reschedule managed sprint: %w", err)
		}
		if rowsAffected != 1 {
			return 0, teamsettings.ErrConcurrentUpdate
		}
		if writeAudit == nil {
			return 0, errors.New("sprint schedule audit writer is required")
		}
		if err := writeAudit(ctx, queries, settings, sprint, nextStart, nextEnd); err != nil {
			return 0, err
		}
		updatedCount++
		nextStart = nextEnd.AddDate(0, 0, 1)
	}
	return updatedCount, nil
}

func nextSprintStartAfter(anchor time.Time, startDay string) time.Time {
	weekdays := map[string]time.Weekday{
		"Monday": time.Monday, "Tuesday": time.Tuesday, "Wednesday": time.Wednesday,
		"Thursday": time.Thursday, "Friday": time.Friday, "Saturday": time.Saturday,
		"Sunday": time.Sunday,
	}
	target, valid := weekdays[startDay]
	if !valid {
		target = time.Monday
	}
	daysUntilTarget := (int(target) - int(anchor.Weekday()) + 7) % 7
	if daysUntilTarget == 0 {
		daysUntilTarget = 7
	}
	return anchor.AddDate(0, 0, daysUntilTarget)
}

func findScheduleConflict(startDate, endDate time.Time, custom []scheduledSprint) (scheduledSprint, bool) {
	for _, sprint := range custom {
		if !startDate.After(sprint.EndDate) && !endDate.Before(sprint.StartDate) {
			return sprint, true
		}
	}
	return scheduledSprint{}, false
}
