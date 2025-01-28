package objectivesrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository errors
var (
	ErrNotFound = errors.New("objective not found")
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error) {
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
			o.end_date,
			o.is_private,
			o.created_at,
			o.updated_at,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM
			objectives o
		LEFT JOIN story_stats ss ON o.objective_id = ss.objective_id
	`
	var setClauses []string
	filters["workspace_id"] = workspaceId

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("o.%s = :%s", field, field))
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

	params := map[string]interface{}{
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

// Update updates an objective
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.Update")
	defer span.End()

	// Verify the objective exists and belongs to the workspace
	if _, err := r.Get(ctx, id, workspaceId); err != nil {
		return err
	}

	query := "UPDATE objectives SET "
	var setClauses []string
	params := map[string]any{"id": id}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE objective_id = :id"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating objective #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to update objective: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update objective"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Objective #%s updated successfully", id), "id", id)
	span.AddEvent("objective updated", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

// Delete deletes an objective
func (r *repo) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.Delete")
	defer span.End()

	query := `DELETE FROM objectives WHERE objective_id = :id AND workspace_id = :workspace_id`
	params := map[string]interface{}{
		"id":           id,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting objective #%s", id), "id", id)
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete objective: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete objective"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get rows affected: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get rows affected"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info(ctx, fmt.Sprintf("Objective #%s deleted successfully", id), "id", id)
	span.AddEvent("objective deleted", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}
