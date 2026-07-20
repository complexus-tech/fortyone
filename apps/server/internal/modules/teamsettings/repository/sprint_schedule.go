package teamsettingsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type scheduledSprint struct {
	ID        uuid.UUID `db:"sprint_id"`
	Name      string    `db:"name"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
}

func (r *repo) ReconcileSprintSchedule(
	ctx context.Context,
	settings teamsettings.CoreTeamSprintSettings,
	actorID *uuid.UUID,
) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin sprint schedule reconciliation: %w", err)
	}
	defer tx.Rollback()

	updated, err := r.reconcileSprintScheduleTx(ctx, tx, settings, actorID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sprint schedule reconciliation: %w", err)
	}

	return updated, nil
}

func (r *repo) reconcileSprintScheduleTx(
	ctx context.Context,
	tx *sqlx.Tx,
	settings teamsettings.CoreTeamSprintSettings,
	actorID *uuid.UUID,
) (int, error) {
	var futureSprints []scheduledSprint
	if err := tx.SelectContext(ctx, &futureSprints, `
		SELECT sprint_id, name, start_date, end_date
		FROM sprints
		WHERE team_id = $1
			AND workspace_id = $2
			AND schedule_managed_by_automation = true
			AND start_date > CURRENT_DATE
		ORDER BY start_date, sprint_id
		FOR UPDATE
	`, settings.TeamID, settings.WorkspaceID); err != nil {
		return 0, fmt.Errorf("list automation-managed future sprints: %w", err)
	}
	if len(futureSprints) == 0 {
		return 0, nil
	}

	var currentDate time.Time
	if err := tx.GetContext(ctx, &currentDate, "SELECT CURRENT_DATE"); err != nil {
		return 0, fmt.Errorf("get current schedule date: %w", err)
	}

	anchorDate := currentDate
	var activeSprintEnd time.Time
	err := tx.GetContext(ctx, &activeSprintEnd, `
		SELECT end_date
		FROM sprints
		WHERE team_id = $1
			AND workspace_id = $2
			AND start_date <= CURRENT_DATE
			AND end_date >= CURRENT_DATE
		ORDER BY end_date DESC
		LIMIT 1
	`, settings.TeamID, settings.WorkspaceID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("get active sprint schedule boundary: %w", err)
	}
	if err == nil && activeSprintEnd.After(anchorDate) {
		anchorDate = activeSprintEnd
	}

	var customFutureSprints []scheduledSprint
	if err := tx.SelectContext(ctx, &customFutureSprints, `
		SELECT sprint_id, name, start_date, end_date
		FROM sprints
		WHERE team_id = $1
			AND workspace_id = $2
			AND schedule_managed_by_automation = false
			AND start_date > CURRENT_DATE
		ORDER BY start_date, sprint_id
	`, settings.TeamID, settings.WorkspaceID); err != nil {
		return 0, fmt.Errorf("list custom future sprints: %w", err)
	}

	nextStart := nextSprintStartAfter(anchorDate, settings.SprintStartDay)
	updatedCount := 0
	for _, sprint := range futureSprints {
		nextEnd := nextStart.AddDate(0, 0, settings.SprintDurationWeeks*7-1)
		if conflictingSprint, ok := findScheduleConflict(nextStart, nextEnd, customFutureSprints); ok {
			return 0, fmt.Errorf("%w: %s", teamsettings.ErrSprintScheduleConflict, conflictingSprint.Name)
		}

		if sprint.StartDate.Equal(nextStart) && sprint.EndDate.Equal(nextEnd) {
			nextStart = nextEnd.AddDate(0, 0, 1)
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE sprints
			SET start_date = $1, end_date = $2, updated_at = NOW()
			WHERE sprint_id = $3 AND workspace_id = $4
		`, nextStart, nextEnd, sprint.ID, settings.WorkspaceID); err != nil {
			return 0, fmt.Errorf("reschedule sprint %s: %w", sprint.ID, err)
		}

		if err := recordSprintScheduleAuditEvent(ctx, tx, settings, sprint, nextStart, nextEnd, actorID); err != nil {
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
	target, ok := weekdays[startDay]
	if !ok {
		target = time.Monday
	}

	daysUntilTarget := (int(target) - int(anchor.Weekday()) + 7) % 7
	if daysUntilTarget == 0 {
		daysUntilTarget = 7
	}
	return anchor.AddDate(0, 0, daysUntilTarget)
}

func findScheduleConflict(startDate, endDate time.Time, customSprints []scheduledSprint) (scheduledSprint, bool) {
	for _, sprint := range customSprints {
		if !startDate.After(sprint.EndDate) && !endDate.Before(sprint.StartDate) {
			return sprint, true
		}
	}
	return scheduledSprint{}, false
}

func recordSprintScheduleAuditEvent(
	ctx context.Context,
	tx *sqlx.Tx,
	settings teamsettings.CoreTeamSprintSettings,
	sprint scheduledSprint,
	startDate, endDate time.Time,
	actorID *uuid.UUID,
) error {
	metadata, err := json.Marshal(map[string]any{
		"old_start_date":        sprint.StartDate.Format("2006-01-02"),
		"old_end_date":          sprint.EndDate.Format("2006-01-02"),
		"new_start_date":        startDate.Format("2006-01-02"),
		"new_end_date":          endDate.Format("2006-01-02"),
		"sprint_duration_weeks": settings.SprintDurationWeeks,
		"sprint_start_day":      settings.SprintStartDay,
	})
	if err != nil {
		return fmt.Errorf("marshal sprint schedule audit metadata: %w", err)
	}

	actorType := "automation"
	if actorID != nil {
		actorType = "user"
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO audit_events (
			workspace_id, team_id, actor_type, actor_id,
			entity_type, entity_id, event_type, metadata
		) VALUES ($1, $2, $3, $4, 'sprint', $5, 'sprint.auto_rescheduled', $6::jsonb)
	`, settings.WorkspaceID, settings.TeamID, actorType, actorID, sprint.ID, string(metadata)); err != nil {
		return fmt.Errorf("record sprint schedule audit event: %w", err)
	}

	return nil
}
