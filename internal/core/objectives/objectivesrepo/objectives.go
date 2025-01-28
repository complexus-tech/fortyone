package objectivesrepo

import (
	"context"
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
