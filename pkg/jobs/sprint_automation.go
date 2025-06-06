package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teamsettings"
	"github.com/complexus-tech/projects-api/internal/repo/teamsettingsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessSprintAutoCreation creates sprints for teams that have auto-creation enabled
func ProcessSprintAutoCreation(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessSprintAutoCreation")
	defer span.End()

	log.Info(ctx, "Processing sprint auto-creation for teams")

	// Initialize the team settings service
	teamsettingsRepo := teamsettingsrepo.New(log, db)
	teamsettingsService := teamsettings.New(log, teamsettingsRepo)

	// Get teams that have auto sprint creation enabled
	teams, err := teamsettingsService.GetTeamsWithAutoSprintCreation(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get teams with auto sprint creation: %w", err)
	}

	if len(teams) == 0 {
		log.Info(ctx, "No teams found with auto sprint creation enabled")
		span.AddEvent("no teams found")
		return nil
	}

	log.Info(ctx, fmt.Sprintf("Found %d teams with auto sprint creation enabled", len(teams)))
	span.AddEvent("teams found", trace.WithAttributes(
		attribute.Int("teams.count", len(teams)),
	))

	teamsProcessed := 0
	sprintsCreated := 0
	teamsWithNewSprints := 0
	errors := 0

	// Track teams and their created sprint counts for bulk update
	teamSprintCounts := make(map[uuid.UUID]int)

	// Process each team
	for _, team := range teams {
		sprintsCreatedForTeam, err := createSprintsForTeam(ctx, db, log, team)
		if err != nil {
			log.Error(ctx, "Failed to create sprints for team",
				"error", err,
				"team_id", team.TeamID,
				"workspace_id", team.WorkspaceID)
			errors++
			continue
		}

		// Track actual sprints created for this team
		if sprintsCreatedForTeam > 0 {
			teamSprintCounts[team.TeamID] = sprintsCreatedForTeam
			teamsWithNewSprints++
		}

		sprintsCreated += sprintsCreatedForTeam
		teamsProcessed++
	}

	// Increment auto sprint numbers with actual counts
	if len(teamSprintCounts) > 0 {
		if err := incrementAutoSprintNumbers(ctx, db, log, teamSprintCounts); err != nil {
			log.Error(ctx, "Failed to increment auto sprint numbers", "error", err)
			// Don't fail the entire job for this - sprints were created successfully
		}
	}

	span.AddEvent("sprint auto-creation completed", trace.WithAttributes(
		attribute.Int("teams.processed", teamsProcessed),
		attribute.Int("teams.with.new.sprints", teamsWithNewSprints),
		attribute.Int("sprints.created", sprintsCreated),
		attribute.Int("errors", errors),
	))

	log.Info(ctx, fmt.Sprintf("Sprint auto-creation completed: %d teams processed, %d teams created new sprints, %d total sprints created, %d errors",
		teamsProcessed, teamsWithNewSprints, sprintsCreated, errors))

	return nil
}

func createSprintsForTeam(ctx context.Context, db *sqlx.DB, log *logger.Logger, team teamsettings.CoreTeamSprintSettings) (int, error) {
	ctx, span := web.AddSpan(ctx, "jobs.createSprintsForTeam")
	defer span.End()

	span.SetAttributes(
		attribute.String("team.id", team.TeamID.String()),
		attribute.String("workspace.id", team.WorkspaceID.String()),
		attribute.Int("target.upcoming.count", team.UpcomingSprintsCount),
		attribute.Int("duration.weeks", team.SprintDurationWeeks),
		attribute.String("start.day", team.SprintStartDay),
	)

	// Check how many upcoming sprints currently exist
	existingCount, err := getExistingUpcomingSprintsCount(ctx, db, team.TeamID, team.WorkspaceID)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to check existing sprints: %w", err)
	}

	// Calculate how many sprints to create to reach the target count
	sprintsToCreate := team.UpcomingSprintsCount - existingCount
	if sprintsToCreate <= 0 {
		if existingCount >= team.UpcomingSprintsCount {
			log.Info(ctx, "Team has enough upcoming sprints",
				"team_id", team.TeamID,
				"existing_count", existingCount,
				"target_count", team.UpcomingSprintsCount,
				"action", "skipping")
		}
		span.AddEvent("no sprints needed", trace.WithAttributes(
			attribute.Int("existing.count", existingCount),
			attribute.Int("target.count", team.UpcomingSprintsCount),
		))
		return 0, nil
	}

	log.Info(ctx, "Creating sprints to reach target count",
		"team_id", team.TeamID,
		"existing_count", existingCount,
		"target_count", team.UpcomingSprintsCount,
		"sprints_to_create", sprintsToCreate)

	// Calculate start date for the first sprint
	startDate := calculateNextSprintStartDate(team.SprintStartDay)

	// If there are existing sprints, start after the last one
	if existingCount > 0 {
		lastSprintEndDate, err := getLastSprintEndDate(ctx, db, team.TeamID, team.WorkspaceID)
		if err != nil {
			span.RecordError(err)
			return 0, fmt.Errorf("failed to get last sprint end date: %w", err)
		}
		// Start the day after the last sprint ends
		startDate = lastSprintEndDate.AddDate(0, 0, 1)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create sprints to reach the target upcoming count
	createdSprints := 0
	for i := 0; i < sprintsToCreate; i++ {
		sprintNumber := team.LastAutoSprintNumber + i + 1
		sprintStartDate := startDate.AddDate(0, 0, i*team.SprintDurationWeeks*7)
		sprintEndDate := sprintStartDate.AddDate(0, 0, team.SprintDurationWeeks*7-1)

		sprintName := fmt.Sprintf("Sprint %d", sprintNumber)

		query := `
			INSERT INTO sprints (name, team_id, workspace_id, start_date, end_date, created_at, updated_at)
			VALUES (:name, :team_id, :workspace_id, :start_date, :end_date, NOW(), NOW())
		`

		params := map[string]any{
			"name":         sprintName,
			"team_id":      team.TeamID,
			"workspace_id": team.WorkspaceID,
			"start_date":   sprintStartDate,
			"end_date":     sprintEndDate,
		}

		stmt, err := tx.PrepareNamedContext(ctx, query)
		if err != nil {
			span.RecordError(err)
			return 0, fmt.Errorf("failed to prepare statement: %w", err)
		}

		_, err = stmt.ExecContext(ctx, params)
		stmt.Close()
		if err != nil {
			span.RecordError(err)
			return 0, fmt.Errorf("failed to create sprint %d for team %s: %w", sprintNumber, team.TeamID, err)
		}

		log.Info(ctx, "Created sprint",
			"sprint_name", sprintName,
			"team_id", team.TeamID,
			"start_date", sprintStartDate.Format("2006-01-02"),
			"end_date", sprintEndDate.Format("2006-01-02"))

		createdSprints++
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	span.AddEvent("sprints created", trace.WithAttributes(
		attribute.Int("sprints.created", createdSprints),
		attribute.Int("existing.count", existingCount),
		attribute.Int("target.count", team.UpcomingSprintsCount),
	))

	return createdSprints, nil
}

func calculateNextSprintStartDate(startDay string) time.Time {
	now := time.Now()

	// Map day names to time.Weekday
	dayMap := map[string]time.Weekday{
		"Monday":    time.Monday,
		"Tuesday":   time.Tuesday,
		"Wednesday": time.Wednesday,
		"Thursday":  time.Thursday,
		"Friday":    time.Friday,
		"Saturday":  time.Saturday,
		"Sunday":    time.Sunday,
	}

	targetDay, exists := dayMap[startDay]
	if !exists {
		targetDay = time.Monday // Default to Monday if invalid day
	}

	// Calculate days until next occurrence of target day
	daysUntilTarget := (int(targetDay) - int(now.Weekday()) + 7) % 7
	if daysUntilTarget == 0 {
		// If today is the target day, start next week
		daysUntilTarget = 7
	}

	return now.AddDate(0, 0, daysUntilTarget)
}

// getExistingUpcomingSprintsCount returns the number of sprints that haven't started yet
func getExistingUpcomingSprintsCount(ctx context.Context, db *sqlx.DB, teamID, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getExistingUpcomingSprintsCount")
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM sprints
		WHERE 
			team_id = :team_id
			AND workspace_id = :workspace_id
			AND start_date > NOW()
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get upcoming sprints count: %w", err)
	}

	span.AddEvent("upcoming sprints counted", trace.WithAttributes(
		attribute.Int("count", count),
	))

	return count, nil
}

// getLastSprintEndDate returns the end date of the last sprint for the team
func getLastSprintEndDate(ctx context.Context, db *sqlx.DB, teamID, workspaceID uuid.UUID) (time.Time, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getLastSprintEndDate")
	defer span.End()

	query := `
		SELECT end_date
		FROM sprints
		WHERE 
			team_id = :team_id
			AND workspace_id = :workspace_id
		ORDER BY end_date DESC
		LIMIT 1
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return time.Time{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var endDate time.Time
	if err := stmt.GetContext(ctx, &endDate, params); err != nil {
		span.RecordError(err)
		return time.Time{}, fmt.Errorf("failed to get last sprint end date: %w", err)
	}

	span.AddEvent("last sprint end date retrieved", trace.WithAttributes(
		attribute.String("end_date", endDate.Format("2006-01-02")),
	))

	return endDate, nil
}

// incrementAutoSprintNumbers increments sprint numbers for multiple teams using individual updates
func incrementAutoSprintNumbers(ctx context.Context, db *sqlx.DB, log *logger.Logger, teamSprintCounts map[uuid.UUID]int) error {
	ctx, span := web.AddSpan(ctx, "jobs.incrementAutoSprintNumbers")
	defer span.End()

	if len(teamSprintCounts) == 0 {
		return nil
	}

	// Use a simpler approach with individual updates within a transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE team_sprint_settings
		SET 
			last_auto_sprint_number = last_auto_sprint_number + :sprints_created,
			updated_at = NOW()
		WHERE team_id = :team_id
	`

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to prepare increment statement: %w", err)
	}
	defer stmt.Close()

	totalUpdated := 0
	for teamID, sprintsCreated := range teamSprintCounts {
		params := map[string]any{
			"team_id":         teamID,
			"sprints_created": sprintsCreated,
		}

		result, err := stmt.ExecContext(ctx, params)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to increment auto sprint number for team %s: %w", teamID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for team %s: %w", teamID, err)
		}

		if rowsAffected > 0 {
			totalUpdated++
		}
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info(ctx, "Incremented auto sprint numbers",
		"teams_updated", totalUpdated,
		"expected_teams", len(teamSprintCounts))

	span.AddEvent("increment completed", trace.WithAttributes(
		attribute.Int("teams.updated", totalUpdated),
		attribute.Int("teams.count", len(teamSprintCounts)),
	))

	return nil
}
