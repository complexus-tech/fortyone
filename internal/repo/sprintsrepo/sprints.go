package sprintsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error) {

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
		INNER JOIN team_members tm ON tm.team_id = s.team_id AND tm.user_id = :user_id
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
	`

	var setClauses []string
	filters["workspace_id"] = workspaceId
	filters["user_id"] = userID

	for field := range filters {
		if field != "user_id" { // Skip user_id since it's used in the JOIN
			setClauses = append(setClauses, fmt.Sprintf("s.%s = :%s", field, field))
		}
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

	params := map[string]any{
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

func (r *repo) Running(ctx context.Context, workspaceId, userID uuid.UUID) ([]sprints.CoreSprint, error) {

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
		INNER JOIN team_members tm ON tm.team_id = s.team_id AND tm.user_id = :user_id
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
		WHERE s.workspace_id = :workspace_id
		AND s.start_date <= NOW() AND s.end_date >= NOW() 
		ORDER BY s.end_date DESC;
	`

	var filters = make(map[string]any)
	filters["workspace_id"] = workspaceId
	filters["user_id"] = userID

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

	params := map[string]any{
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

	params := map[string]any{
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

	// First, get the sprint basic info (needed for other queries)
	sprint, err := r.GetByID(ctx, sprintID, workspaceID)
	if err != nil {
		return sprints.CoreSprintAnalytics{}, err
	}

	// Run independent queries in parallel for better performance
	var (
		storyBreakdown sprints.CoreStoryBreakdown
		burndownData   []sprints.CoreBurndownDataPoint
		teamAllocation []sprints.CoreTeamMemberAllocation
		wg             sync.WaitGroup
		mu             sync.Mutex
		analyticsErr   error
	)

	wg.Add(3)

	// Parallel query 1: Story breakdown
	go func() {
		defer wg.Done()
		breakdown, err := r.getStoryBreakdown(ctx, sprintID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		storyBreakdown = breakdown
	}()

	// Parallel query 2: Burndown data (simplified for performance)
	go func() {
		defer wg.Done()
		burndown, err := r.getBurndownData(ctx, sprintID, sprint.StartDate, sprint.EndDate)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		burndownData = burndown
	}()

	// Parallel query 3: Team allocation
	go func() {
		defer wg.Done()
		allocation, err := r.getTeamAllocation(ctx, sprintID, sprint.Team)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		teamAllocation = allocation
	}()

	// Wait for all queries to complete
	wg.Wait()

	if analyticsErr != nil {
		r.log.Error(ctx, "Failed to get analytics data", "error", analyticsErr)
		span.RecordError(analyticsErr)
		return sprints.CoreSprintAnalytics{}, analyticsErr
	}

	// Calculate overview metrics
	overview := r.calculateOverview(sprint, storyBreakdown)

	analytics := sprints.CoreSprintAnalytics{
		SprintID:       sprintID,
		Overview:       overview,
		StoryBreakdown: storyBreakdown,
		Burndown:       burndownData,
		TeamAllocation: teamAllocation,
	}

	r.log.Info(ctx, "Sprint analytics retrieved successfully.")
	span.AddEvent("sprint analytics retrieved.", trace.WithAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.Int("total_stories", storyBreakdown.Total),
	))

	return analytics, nil
}

func (r *repo) getStoryBreakdown(ctx context.Context, sprintID uuid.UUID) (sprints.CoreStoryBreakdown, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN st.category = 'started' THEN 1 END) as in_progress,
			COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as todo,
			COUNT(CASE WHEN st.category = 'paused' THEN 1 END) as blocked,
			COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled
		FROM stories s
		LEFT JOIN statuses st ON s.status_id = st.status_id
		WHERE s.sprint_id = :sprint_id 
		AND s.deleted_at IS NULL 
		AND s.archived_at IS NULL
	`

	params := map[string]any{
		"sprint_id": sprintID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return sprints.CoreStoryBreakdown{}, err
	}
	defer stmt.Close()

	var result sprints.CoreStoryBreakdown
	err = stmt.GetContext(ctx, &result, params)
	if err != nil {
		return sprints.CoreStoryBreakdown{}, err
	}

	return result, nil
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

	// Check if sprint has ended
	if now.After(sprint.EndDate) {
		status = "completed"
	} else if now.Before(sprint.StartDate) {
		status = "not_started"
	} else if totalDays > 0 {
		// Sprint is active, check progress
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
	// Enhanced query using story_activities to track accurate scope changes and completions
	query := `
		WITH date_series AS (
			SELECT DATE(generate_series(
				CAST(:start_date AS date),
				CAST(:end_date AS date),
				INTERVAL '1 day'
			)) as burn_date
		),
		sprint_activity AS (
			SELECT 
				DATE(sa.created_at) as activity_date,
				sa.story_id,
				sa.field_changed,
				sa.current_value,
				st.category as status_category,
				ROW_NUMBER() OVER (
					PARTITION BY sa.story_id, DATE(sa.created_at), sa.field_changed
					ORDER BY sa.created_at DESC
				) as rn
			FROM story_activities sa
			JOIN stories s ON sa.story_id = s.id
			LEFT JOIN statuses st ON sa.field_changed = 'status_id' AND CAST(sa.current_value AS uuid) = st.status_id
			WHERE s.sprint_id = :sprint_id
			  AND sa.activity_type = 'update'
			  AND sa.field_changed IN ('sprint_id', 'status_id')
			  AND sa.created_at >= :start_date
			  AND sa.created_at <= CAST(:end_date AS timestamp) + INTERVAL '1 day'
			  AND s.deleted_at IS NULL
			  AND s.archived_at IS NULL
		),
		initial_scope AS (
			SELECT COUNT(*) as initial_stories
			FROM stories s
			WHERE s.sprint_id = :sprint_id 
			AND s.deleted_at IS NULL 
			AND s.archived_at IS NULL
			AND s.created_at < :start_date
		),
		daily_changes AS (
			SELECT 
				ds.burn_date,
				-- Stories added to sprint on this date
				COUNT(CASE 
					WHEN spa.field_changed = 'sprint_id' 
					AND CAST(spa.current_value AS uuid) = :sprint_id 
					AND spa.rn = 1 
					THEN 1 
				END) as stories_added,
				-- Stories completed on this date  
				COUNT(CASE 
					WHEN spa.field_changed = 'status_id' 
					AND spa.status_category = 'completed' 
					AND spa.rn = 1 
					THEN 1 
				END) as stories_completed
			FROM date_series ds
			LEFT JOIN sprint_activity spa ON ds.burn_date = spa.activity_date
			GROUP BY ds.burn_date
		)
		SELECT 
			dc.burn_date as event_date,
			dc.stories_added,
			dc.stories_completed,
			ins.initial_stories
		FROM daily_changes dc
		CROSS JOIN initial_scope ins
		ORDER BY dc.burn_date
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	type queryResult struct {
		EventDate        time.Time `db:"event_date"`
		StoriesAdded     int       `db:"stories_added"`
		StoriesCompleted int       `db:"stories_completed"`
		InitialStories   int       `db:"initial_stories"`
	}

	var results []queryResult
	err = stmt.SelectContext(ctx, &results, map[string]any{
		"sprint_id":  sprintID,
		"start_date": startDate,
		"end_date":   endDate,
	})
	if err != nil {
		return nil, err
	}

	// Build daily changes maps
	scopeChanges := make(map[string]int)
	completions := make(map[string]int)
	initialStories := 0

	for _, result := range results {
		dateKey := result.EventDate.Format("2006-01-02")
		if result.StoriesAdded > 0 {
			scopeChanges[dateKey] = result.StoriesAdded
		}
		if result.StoriesCompleted > 0 {
			completions[dateKey] = result.StoriesCompleted
		}
		if result.InitialStories > 0 {
			initialStories = result.InitialStories
		}
	}

	// Generate burndown data for all days
	var burndownData []sprints.CoreBurndownDataPoint
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1 // +1 to include end date

	currentTotal := initialStories
	cumulativeCompleted := 0

	for i := 0; i < totalDays; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dateKey := currentDate.Format("2006-01-02")

		// Add any new stories added on this day
		if added, exists := scopeChanges[dateKey]; exists {
			currentTotal += added
		}

		// Calculate ideal remaining based on current scope and days remaining
		daysRemaining := totalDays - i - 1
		idealRemaining := 0
		if daysRemaining > 0 && currentTotal > cumulativeCompleted {
			// Ideal burndown from current scope over remaining days
			storiesRemaining := currentTotal - cumulativeCompleted
			idealRemaining = storiesRemaining - (storiesRemaining * (i) / (totalDays - 1))
		}
		if idealRemaining < 0 {
			idealRemaining = 0
		}

		// Calculate actual remaining
		actualRemaining := currentTotal
		if !currentDate.After(time.Now()) {
			// Add any completions for this day
			if completed, exists := completions[dateKey]; exists {
				cumulativeCompleted += completed
			}
			actualRemaining = currentTotal - cumulativeCompleted
		}

		burndownData = append(burndownData, sprints.CoreBurndownDataPoint{
			Date:      currentDate,
			Remaining: actualRemaining,
			Ideal:     idealRemaining,
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

	params := map[string]any{
		"sprint_id": sprintID,
		"team_id":   teamID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var allocation []sprints.CoreTeamMemberAllocation
	if err := stmt.SelectContext(ctx, &allocation, params); err != nil {
		return nil, err
	}

	return allocation, nil
}
