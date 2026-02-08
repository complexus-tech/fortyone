package objectivesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.List")
	defer span.End()

	var objectives []dbObjective
	q := `
		WITH story_stats AS (
			SELECT 
				o.objective_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM objectives o
			LEFT JOIN stories s ON o.objective_id = s.objective_id
			LEFT JOIN statuses st ON s.status_id = st.status_id
			WHERE s.deleted_at IS NULL AND s.archived_at IS NULL
			GROUP BY o.objective_id
		)
		SELECT
			o.objective_id,
			o.name,
			o.description,
			o.lead_user_id,
			o.team_id,
			o.workspace_id,
			o.start_date,
			o.status_id,
			o.priority,
			o.health,
			o.end_date,
			o.is_private,
			o.created_at,
			o.updated_at,
			o.created_by,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM
			objectives o
		INNER JOIN team_members tm ON tm.team_id = o.team_id AND tm.user_id = :user_id
		LEFT JOIN story_stats ss ON o.objective_id = ss.objective_id
	`
	var setClauses []string
	filters["workspace_id"] = workspaceId
	filters["user_id"] = userID

	for field := range filters {
		if field != "user_id" { // Skip user_id since it's used in the JOIN
			setClauses = append(setClauses, fmt.Sprintf("o.%s = :%s", field, field))
		}
	}

	q += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY o.created_at DESC;"

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objectives.")
	if err := stmt.SelectContext(ctx, &objectives, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objectives from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("objectives not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "objectives retrieved successfully.")
	span.AddEvent("objectives retrieved.", trace.WithAttributes(
		attribute.Int("objectives.count", len(objectives)),
		attribute.String("query", q),
	))

	return toCoreObjectives(objectives), nil
}

// Get retrieves an objective by ID
func (r *repo) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (objectives.CoreObjective, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.Get")
	defer span.End()

	var objective dbObjective
	q := `
		WITH story_stats AS (
			SELECT 
				o.objective_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM objectives o
			LEFT JOIN stories s ON o.objective_id = s.objective_id
			LEFT JOIN statuses st ON s.status_id = st.status_id
			WHERE s.deleted_at IS NULL AND s.archived_at IS NULL
			GROUP BY o.objective_id
		)
		SELECT
			o.objective_id,
			o.name,
			o.description,
			o.lead_user_id,
			o.team_id,
			o.workspace_id,
			o.start_date,
			o.end_date,
			o.is_private,
			o.created_at,
			o.updated_at,
			o.status_id,
			o.priority,
			o.health,
			o.created_by,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM
			objectives o
		LEFT JOIN story_stats ss ON o.objective_id = ss.objective_id
		WHERE o.objective_id = :id AND o.workspace_id = :workspace_id;
	`

	params := map[string]any{
		"id":           id,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectives.CoreObjective{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objective.")
	if err := stmt.GetContext(ctx, &objective, params); err != nil {
		if err == sql.ErrNoRows {
			return objectives.CoreObjective{}, ErrNotFound
		}
		errMsg := fmt.Sprintf("Failed to retrieve objective from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("objective not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectives.CoreObjective{}, err
	}

	r.log.Info(ctx, "Objective retrieved successfully.")
	span.AddEvent("objective retrieved.", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return toCoreObjective(objective), nil
}

func (r *repo) GetAnalytics(ctx context.Context, objectiveID uuid.UUID, workspaceID uuid.UUID) (objectives.CoreObjectiveAnalytics, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.GetAnalytics")
	defer span.End()

	// Run queries in parallel for better performance
	var (
		priorityBreakdown []objectives.CorePriorityBreakdown
		progressBreakdown objectives.CoreProgressBreakdown
		teamAllocation    []objectives.CoreTeamMemberAllocation
		progressChart     []objectives.CoreObjectiveProgressDataPoint
		wg                sync.WaitGroup
		mu                sync.Mutex
		analyticsErr      error
	)

	wg.Add(4)

	// Parallel query 1: Priority breakdown
	go func() {
		defer wg.Done()
		breakdown, err := r.getPriorityBreakdown(ctx, objectiveID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		priorityBreakdown = breakdown
	}()

	// Parallel query 2: Progress breakdown
	go func() {
		defer wg.Done()
		breakdown, err := r.getProgressBreakdown(ctx, objectiveID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		progressBreakdown = breakdown
	}()

	// Parallel query 3: Team allocation
	go func() {
		defer wg.Done()
		allocation, err := r.getTeamAllocation(ctx, objectiveID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		teamAllocation = allocation
	}()

	// Parallel query 4: Progress chart
	go func() {
		defer wg.Done()
		chart, err := r.getObjectiveProgressData(ctx, objectiveID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			analyticsErr = err
			return
		}
		progressChart = chart
	}()

	// Wait for all queries to complete
	wg.Wait()

	if analyticsErr != nil {
		r.log.Error(ctx, "Failed to get analytics data", "error", analyticsErr)
		span.RecordError(analyticsErr)
		return objectives.CoreObjectiveAnalytics{}, analyticsErr
	}

	analytics := objectives.CoreObjectiveAnalytics{
		ObjectiveID:       objectiveID,
		PriorityBreakdown: priorityBreakdown,
		ProgressBreakdown: progressBreakdown,
		TeamAllocation:    teamAllocation,
		ProgressChart:     progressChart,
	}

	r.log.Info(ctx, "Objective analytics retrieved successfully.")
	span.AddEvent("objective analytics retrieved.", trace.WithAttributes(
		attribute.String("objective.id", objectiveID.String()),
		attribute.Int("priority_breakdown.count", len(priorityBreakdown)),
		attribute.Int("team_allocation.count", len(teamAllocation)),
		attribute.Int("progress_chart.count", len(progressChart)),
	))

	return analytics, nil
}

func (r *repo) getPriorityBreakdown(ctx context.Context, objectiveID uuid.UUID) ([]objectives.CorePriorityBreakdown, error) {
	query := `
		SELECT 
			COALESCE(priority, 'No Priority') as priority,
			COUNT(*) as count
		FROM stories
		WHERE objective_id = :objective_id
			AND deleted_at IS NULL
			AND archived_at IS NULL
		GROUP BY priority
		ORDER BY 
			CASE priority
				WHEN 'Urgent' THEN 1
				WHEN 'High' THEN 2
				WHEN 'Medium' THEN 3
				WHEN 'Low' THEN 4
				WHEN 'No Priority' THEN 5
			END
	`

	params := map[string]any{
		"objective_id": objectiveID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var breakdown []objectives.CorePriorityBreakdown
	if err := stmt.SelectContext(ctx, &breakdown, params); err != nil {
		return nil, err
	}

	return breakdown, nil
}

func (r *repo) getProgressBreakdown(ctx context.Context, objectiveID uuid.UUID) (objectives.CoreProgressBreakdown, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN st.category = 'started' THEN 1 END) as in_progress,
			COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as todo,
			COUNT(CASE WHEN st.category = 'blocked' THEN 1 END) as blocked,
			COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled
		FROM stories s
		INNER JOIN statuses st ON s.status_id = st.status_id
		WHERE s.objective_id = :objective_id
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
	`

	params := map[string]any{
		"objective_id": objectiveID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return objectives.CoreProgressBreakdown{}, err
	}
	defer stmt.Close()

	var breakdown objectives.CoreProgressBreakdown
	if err := stmt.GetContext(ctx, &breakdown, params); err != nil {
		return objectives.CoreProgressBreakdown{}, err
	}

	return breakdown, nil
}

func (r *repo) getTeamAllocation(ctx context.Context, objectiveID uuid.UUID) ([]objectives.CoreTeamMemberAllocation, error) {
	query := `
		SELECT 
			u.user_id,
			u.username,
			u.avatar_url,
			COUNT(s.id) as assigned,
			COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed
		FROM stories s
		INNER JOIN users u ON s.assignee_id = u.user_id
		LEFT JOIN statuses st ON s.status_id = st.status_id
		WHERE s.objective_id = :objective_id
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND u.is_active = true
		GROUP BY u.user_id, u.username, u.avatar_url
		ORDER BY assigned DESC, u.username
	`

	params := map[string]any{
		"objective_id": objectiveID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var allocation []objectives.CoreTeamMemberAllocation
	if err := stmt.SelectContext(ctx, &allocation, params); err != nil {
		return nil, err
	}

	return allocation, nil
}

func (r *repo) getObjectiveProgressData(ctx context.Context, objectiveID uuid.UUID) ([]objectives.CoreObjectiveProgressDataPoint, error) {
	query := `
		WITH date_series AS (
			SELECT DATE(generate_series(
				NOW() - INTERVAL '30 days',
				NOW(),
				INTERVAL '1 day'
			)) as completion_date
		),
		story_status_changes AS (
			SELECT 
				s.id as story_id,
				DATE(sa.created_at) as change_date,
				st.category,
				ROW_NUMBER() OVER (
					PARTITION BY s.id, DATE(sa.created_at) 
					ORDER BY sa.created_at DESC
				) as rn
			FROM stories s
			JOIN story_activities sa ON sa.story_id = s.id
			JOIN statuses st ON CAST(sa.current_value AS uuid) = st.status_id
			WHERE s.objective_id = :objective_id
			  AND sa.activity_type = 'update'
			  AND sa.field_changed = 'status_id'
			  AND s.deleted_at IS NULL
			  AND s.archived_at IS NULL
		),
		latest_status_by_date AS (
			SELECT 
				ds.completion_date,
				s.id as story_id,
				COALESCE(
					(SELECT ssc.category 
					 FROM story_status_changes ssc 
					 WHERE ssc.story_id = s.id 
					   AND ssc.change_date <= ds.completion_date 
					   AND ssc.rn = 1
					 ORDER BY ssc.change_date DESC 
					 LIMIT 1),
					'unstarted'
				) as status_category
			FROM date_series ds
			CROSS JOIN stories s
			WHERE s.objective_id = :objective_id
			  AND s.created_at <= ds.completion_date + INTERVAL '1 day'
			  AND s.deleted_at IS NULL
			  AND s.archived_at IS NULL
		)
		SELECT 
			ds.completion_date,
			COALESCE(SUM(CASE WHEN lsbd.status_category = 'completed' THEN 1 ELSE 0 END), 0) as stories_completed,
			COALESCE(SUM(CASE WHEN lsbd.status_category = 'started' THEN 1 ELSE 0 END), 0) as stories_in_progress,
			COALESCE(COUNT(lsbd.story_id), 0) as total_stories
		FROM date_series ds
		LEFT JOIN latest_status_by_date lsbd ON ds.completion_date = lsbd.completion_date
		GROUP BY ds.completion_date
		ORDER BY ds.completion_date
	`

	params := map[string]any{
		"objective_id": objectiveID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var progressData []objectives.CoreObjectiveProgressDataPoint
	if err := stmt.SelectContext(ctx, &progressData, params); err != nil {
		return nil, err
	}

	return progressData, nil
}
