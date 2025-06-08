package sprintsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error) {

	ctx, span := web.AddSpan(ctx, "business.repository.sprints.List")
	defer span.End()

	var sprints []dbSprint
	query := `
		WITH story_stats AS (
			SELECT 
				s.sprint_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM sprints s
			LEFT JOIN stories str ON s.sprint_id = str.sprint_id
			LEFT JOIN statuses st ON str.status_id = st.status_id
			WHERE str.deleted_at IS NULL AND str.archived_at IS NULL
			GROUP BY s.sprint_id
		)
		SELECT
			s.sprint_id,
			s.name,
			s.goal,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.start_date,
			s.end_date,
			s.created_at,
			s.updated_at,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM
			sprints s
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
	`

	var setClauses []string
	filters["workspace_id"] = workspaceId

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("s.%s = :%s", field, field))
	}

	query += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY s.end_date DESC;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching sprints.")
	if err := stmt.SelectContext(ctx, &sprints, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve sprints from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("sprints not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "sprints retrieved successfully.")
	span.AddEvent("sprints retrieved.", trace.WithAttributes(
		attribute.Int("sprints.count", len(sprints)),
		attribute.String("query", query),
	))

	return toCoreSprints(sprints), nil
}

func (r *repo) GetByID(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) (sprints.CoreSprint, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.GetByID")
	defer span.End()

	var sprint dbSprint
	query := `
		SELECT
			sprint_id,
			name,
			goal,
			objective_id,
			team_id,
			workspace_id,
			start_date,
			end_date,
			created_at,
			updated_at,
			0 as total_stories,
			0 as cancelled_stories,
			0 as completed_stories,
			0 as started_stories,
			0 as unstarted_stories,
			0 as backlog_stories
		FROM sprints
		WHERE sprint_id = :sprint_id AND workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"sprint_id":    sprintID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching sprint by ID.")
	if err := stmt.GetContext(ctx, &sprint, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve sprint from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("sprint not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	r.log.Info(ctx, "sprint retrieved successfully.")
	span.AddEvent("sprint retrieved.", trace.WithAttributes(
		attribute.String("sprint.id", sprint.ID.String()),
		attribute.String("workspace.id", sprint.Workspace.String()),
	))

	return toCoreSprint(sprint), nil
}

func (r *repo) Running(ctx context.Context, workspaceId uuid.UUID) ([]sprints.CoreSprint, error) {

	ctx, span := web.AddSpan(ctx, "business.repository.sprints.List")
	defer span.End()

	var sprints []dbSprint
	query := `
		WITH story_stats AS (
			SELECT 
				s.sprint_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM sprints s
			LEFT JOIN stories str ON s.sprint_id = str.sprint_id
			LEFT JOIN statuses st ON str.status_id = st.status_id
			WHERE str.deleted_at IS NULL AND str.archived_at IS NULL
			GROUP BY s.sprint_id
		)
		SELECT
			s.sprint_id,
			s.name,
			s.goal,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.start_date,
			s.end_date,
			s.created_at,
			s.updated_at,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM
			sprints s
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
		WHERE s.workspace_id = :workspace_id
		AND s.start_date <= NOW() AND s.end_date >= NOW() 
		ORDER BY s.end_date DESC;
	`

	var filters = make(map[string]any)
	filters["workspace_id"] = workspaceId

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching sprints.")
	if err := stmt.SelectContext(ctx, &sprints, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve sprints from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("sprints not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "sprints retrieved successfully.")
	span.AddEvent("sprints retrieved.", trace.WithAttributes(
		attribute.Int("sprints.count", len(sprints)),
		attribute.String("query", query),
	))

	return toCoreSprints(sprints), nil
}

func (r *repo) Create(ctx context.Context, sprint sprints.CoreNewSprint) (sprints.CoreSprint, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Create")
	defer span.End()

	var s dbSprint
	query := `
		INSERT INTO sprints (
			name,
			goal,
			objective_id,
			team_id,
			workspace_id,
			start_date,
			end_date
		) VALUES (
			:name,
			:goal,
			:objective_id,
			:team_id,
			:workspace_id,
			:start_date,
			:end_date
		)
		RETURNING
			sprint_id,
			name,
			goal,
			objective_id,
			team_id,
			workspace_id,
			start_date,
			end_date,
			created_at,
			updated_at,
			0 as total_stories,
			0 as cancelled_stories,
			0 as completed_stories,
			0 as started_stories,
			0 as unstarted_stories,
			0 as backlog_stories
	`

	params := map[string]interface{}{
		"name":         sprint.Name,
		"goal":         sprint.Goal,
		"objective_id": sprint.Objective,
		"team_id":      sprint.Team,
		"workspace_id": sprint.Workspace,
		"start_date":   sprint.StartDate,
		"end_date":     sprint.EndDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &s, params); err != nil {
		errMsg := fmt.Sprintf("Failed to create sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	r.log.Info(ctx, "sprint created successfully.")
	span.AddEvent("sprint created.", trace.WithAttributes(
		attribute.String("sprint.id", s.ID.String()),
		attribute.String("workspace.id", s.Workspace.String()),
	))

	return toCoreSprint(s), nil
}

func (r *repo) Update(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID, updates sprints.CoreUpdateSprint) (sprints.CoreSprint, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Update")
	defer span.End()

	query := "UPDATE sprints SET "
	var setClauses []string
	params := map[string]any{
		"sprint_id":    sprintID,
		"workspace_id": workspaceID,
	}

	if updates.Name != nil {
		setClauses = append(setClauses, "name = :name")
		params["name"] = *updates.Name
	}
	if updates.Goal != nil {
		setClauses = append(setClauses, "goal = :goal")
		params["goal"] = *updates.Goal
	}
	if updates.Objective != nil {
		setClauses = append(setClauses, "objective_id = :objective_id")
		params["objective_id"] = *updates.Objective
	}
	if updates.StartDate != nil {
		setClauses = append(setClauses, "start_date = :start_date")
		params["start_date"] = *updates.StartDate
	}
	if updates.EndDate != nil {
		setClauses = append(setClauses, "end_date = :end_date")
		params["end_date"] = *updates.EndDate
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE sprint_id = :sprint_id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("Failed to update sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	// Get the updated sprint
	var s dbSprint
	getQuery := `
		WITH story_stats AS (
			SELECT 
				s.sprint_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM sprints s
			LEFT JOIN stories str ON s.sprint_id = str.sprint_id
			LEFT JOIN statuses st ON str.status_id = st.status_id
			WHERE str.deleted_at IS NULL AND str.archived_at IS NULL
			GROUP BY s.sprint_id
		)
		SELECT
			s.sprint_id,
			s.name,
			s.goal,
			s.objective_id,
			s.team_id,
			s.workspace_id,
			s.start_date,
			s.end_date,
			s.created_at,
			s.updated_at,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM sprints s
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
		WHERE s.sprint_id = :sprint_id AND s.workspace_id = :workspace_id
	`

	stmt, err = r.db.PrepareNamedContext(ctx, getQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare get statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare get statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &s, params); err != nil {
		errMsg := fmt.Sprintf("Failed to get updated sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get updated sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	r.log.Info(ctx, "sprint updated successfully.")
	span.AddEvent("sprint updated.", trace.WithAttributes(
		attribute.String("sprint.id", s.ID.String()),
		attribute.String("workspace.id", s.Workspace.String()),
	))

	return toCoreSprint(s), nil
}

func (r *repo) Delete(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Delete")
	defer span.End()

	query := `
		DELETE FROM sprints
		WHERE sprint_id = :sprint_id
		AND workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"sprint_id":    sprintID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to delete sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("sprint not found")
	}

	r.log.Info(ctx, "sprint deleted successfully.")
	span.AddEvent("sprint deleted.", trace.WithAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

func (r *repo) GetAnalytics(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) (sprints.CoreSprintAnalytics, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.GetAnalytics")
	defer span.End()

	// First, get the sprint basic info
	sprint, err := r.GetByID(ctx, sprintID, workspaceID)
	if err != nil {
		return sprints.CoreSprintAnalytics{}, err
	}

	// Calculate story breakdown and health indicators in one query (optimization)
	storyBreakdown, healthIndicators, err := r.getStoryBreakdownAndHealth(ctx, sprintID, sprint.StartDate)
	if err != nil {
		r.log.Error(ctx, "Failed to get story breakdown and health", "error", err)
		span.RecordError(err)
		return sprints.CoreSprintAnalytics{}, err
	}

	// Calculate overview metrics
	overview := r.calculateOverview(sprint, storyBreakdown)

	// Get burndown data
	burndownData, err := r.getBurndownData(ctx, sprintID, sprint.StartDate, sprint.EndDate)
	if err != nil {
		r.log.Error(ctx, "Failed to get burndown data", "error", err)
		span.RecordError(err)
		return sprints.CoreSprintAnalytics{}, err
	}

	// Get team allocation
	teamAllocation, err := r.getTeamAllocation(ctx, sprintID, sprint.Team)
	if err != nil {
		r.log.Error(ctx, "Failed to get team allocation", "error", err)
		span.RecordError(err)
		return sprints.CoreSprintAnalytics{}, err
	}

	analytics := sprints.CoreSprintAnalytics{
		SprintID:         sprintID,
		Overview:         overview,
		StoryBreakdown:   storyBreakdown,
		Burndown:         burndownData,
		TeamAllocation:   teamAllocation,
		HealthIndicators: healthIndicators,
	}

	r.log.Info(ctx, "Sprint analytics retrieved successfully.")
	span.AddEvent("sprint analytics retrieved.", trace.WithAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.Int("total_stories", storyBreakdown.Total),
	))

	return analytics, nil
}

func (r *repo) getStoryBreakdownAndHealth(ctx context.Context, sprintID uuid.UUID, sprintStartDate time.Time) (sprints.CoreStoryBreakdown, sprints.CoreSprintHealthIndicators, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN st.category = 'started' THEN 1 END) as in_progress,
			COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as todo,
			COUNT(CASE WHEN st.category = 'paused' THEN 1 END) as blocked,
			COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
			COUNT(CASE WHEN st.category = 'paused' THEN 1 END) as blocked_count,
			COUNT(CASE WHEN s.end_date < NOW() AND st.category NOT IN ('completed', 'cancelled') THEN 1 END) as overdue_count,
			COUNT(CASE WHEN s.created_at > :sprint_start_date THEN 1 END) as added_mid_sprint
		FROM stories s
		LEFT JOIN statuses st ON s.status_id = st.status_id
		WHERE s.sprint_id = :sprint_id 
		AND s.deleted_at IS NULL 
		AND s.archived_at IS NULL
	`

	params := map[string]interface{}{
		"sprint_id":         sprintID,
		"sprint_start_date": sprintStartDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return sprints.CoreStoryBreakdown{}, sprints.CoreSprintHealthIndicators{}, err
	}
	defer stmt.Close()

	var result struct {
		sprints.CoreStoryBreakdown
		sprints.CoreSprintHealthIndicators
	}

	err = stmt.GetContext(ctx, &result, params)
	if err != nil {
		return sprints.CoreStoryBreakdown{}, sprints.CoreSprintHealthIndicators{}, err
	}

	return result.CoreStoryBreakdown, result.CoreSprintHealthIndicators, nil
}

func (r *repo) calculateOverview(sprint sprints.CoreSprint, breakdown sprints.CoreStoryBreakdown) sprints.CoreSprintOverview {
	now := time.Now()

	// Calculate completion percentage
	completionPercentage := 0
	if breakdown.Total > 0 {
		completionPercentage = (breakdown.Completed * 100) / breakdown.Total
	}

	// Calculate days elapsed and remaining
	totalDays := int(sprint.EndDate.Sub(sprint.StartDate).Hours() / 24)
	daysElapsed := int(now.Sub(sprint.StartDate).Hours() / 24)
	daysRemaining := int(sprint.EndDate.Sub(now).Hours() / 24)

	// Ensure values are not negative
	if daysElapsed < 0 {
		daysElapsed = 0
	}
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Calculate sprint status
	status := "on_track"
	if totalDays > 0 {
		timeProgress := float64(daysElapsed) / float64(totalDays)
		workProgress := float64(completionPercentage) / 100.0

		if workProgress < timeProgress-0.2 {
			status = "behind"
		} else if workProgress < timeProgress-0.1 {
			status = "at_risk"
		}
	}

	return sprints.CoreSprintOverview{
		CompletionPercentage: completionPercentage,
		DaysElapsed:          daysElapsed,
		DaysRemaining:        daysRemaining,
		Status:               status,
	}
}

func (r *repo) getBurndownData(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]sprints.CoreBurndownDataPoint, error) {
	// Optimized burndown query - generate date series and calculate remaining stories efficiently
	query := `
		WITH RECURSIVE date_series AS (
			SELECT CAST(:start_date AS date) as date
			UNION ALL
			SELECT CAST(date + INTERVAL '1 day' AS date)
			FROM date_series
			WHERE date < CAST(:end_date AS date)
		),
		story_completion AS (
			SELECT 
				DATE(s.updated_at) as completion_date,
				COUNT(*) as completed_count
			FROM stories s
			INNER JOIN statuses st ON s.status_id = st.status_id
			WHERE s.sprint_id = :sprint_id 
			AND st.category = 'completed'
			AND s.deleted_at IS NULL 
			AND s.archived_at IS NULL
			AND s.updated_at BETWEEN :start_date AND :end_date
			GROUP BY DATE(s.updated_at)
		),
		total_count AS (
			SELECT COUNT(*) as total
			FROM stories s
			WHERE s.sprint_id = :sprint_id 
			AND s.deleted_at IS NULL 
			AND s.archived_at IS NULL
		)
		SELECT 
			ds.date,
			tc.total - COALESCE(
				(SELECT SUM(sc.completed_count) 
				 FROM story_completion sc 
				 WHERE sc.completion_date <= ds.date), 0
			) as remaining
		FROM date_series ds
		CROSS JOIN total_count tc
		ORDER BY ds.date
	`

	params := map[string]interface{}{
		"sprint_id":  sprintID,
		"start_date": startDate,
		"end_date":   endDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var burndownData []sprints.CoreBurndownDataPoint
	for rows.Next() {
		var date time.Time
		var remaining int
		if err := rows.Scan(&date, &remaining); err != nil {
			return nil, err
		}
		burndownData = append(burndownData, sprints.CoreBurndownDataPoint{
			Date:      date,
			Remaining: remaining,
		})
	}

	return burndownData, nil
}

func (r *repo) getTeamAllocation(ctx context.Context, sprintID uuid.UUID, teamID uuid.UUID) ([]sprints.CoreTeamMemberAllocation, error) {
	query := `
		SELECT 
			u.user_id,
			u.username,
			u.avatar_url,
			COUNT(s.id) as assigned,
			COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed
		FROM users u
		INNER JOIN team_members tm ON u.user_id = tm.user_id
		LEFT JOIN stories s ON u.user_id = s.assignee_id 
			AND s.sprint_id = :sprint_id 
			AND s.deleted_at IS NULL 
			AND s.archived_at IS NULL
		LEFT JOIN statuses st ON s.status_id = st.status_id
		WHERE tm.team_id = :team_id
		AND u.is_active = true
		GROUP BY u.user_id, u.username, u.avatar_url
		ORDER BY assigned DESC, u.username
	`

	params := map[string]interface{}{
		"sprint_id": sprintID,
		"team_id":   teamID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocation []sprints.CoreTeamMemberAllocation
	for rows.Next() {
		var member sprints.CoreTeamMemberAllocation
		if err := rows.Scan(&member.MemberID, &member.Username, &member.AvatarURL, &member.Assigned, &member.Completed); err != nil {
			return nil, err
		}
		allocation = append(allocation, member)
	}

	return allocation, nil
}
