package labelsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/labels"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		log: log,
		db:  db,
	}
}

func (r *repo) GetLabels(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]labels.CoreLabel, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.GetLabels")
	defer span.End()
	var labels []dbLabel
	query := `
		SELECT 
			label_id,
			name,
			project_id,
			team_id,
			workspace_id,
			color,
			created_at,
			updated_at
		FROM labels
	`
	var setClauses []string
	filters["workspace_id"] = workspaceId

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	query += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY created_at DESC;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching labels.")
	if err := stmt.SelectContext(ctx, &labels, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve labels from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("labels not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "labels retrieved successfully.")
	span.AddEvent("labels retrieved.", trace.WithAttributes(
		attribute.Int("labels.count", len(labels)),
	))

	return toCoreLabels(labels), nil
}

func (r *repo) CreateLabel(ctx context.Context, cnl labels.CoreNewLabel) (labels.CoreLabel, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.CreateLabel")
	defer span.End()
	var label dbLabel
	query := `
			INSERT INTO
				labels (name, project_id, team_id, workspace_id, color)
			VALUES
				(:name,:project_id,:team_id,:workspace_id,:color)
			RETURNING
				label_id, name, project_id, team_id,
				workspace_id, color, created_at,
				updated_at
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return labels.CoreLabel{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Creating label.")
	if err := stmt.GetContext(ctx, &label, toDbNewLabel(cnl)); err != nil {
		errMsg := fmt.Sprintf("Failed to create label: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create label"), trace.WithAttributes(attribute.String("error", errMsg)))
		return labels.CoreLabel{}, err
	}

	return toCoreLabel(label), nil
}

func (r *repo) GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (labels.CoreLabel, error) {
	var label dbLabel
	query := `
		SELECT
			label_id,
			name,
			project_id,
			team_id,
			workspace_id,
			color,
			created_at,
			updated_at
		FROM
			labels
		WHERE
			label_id = :label_id
			AND workspace_id = :workspace_id
			`
	params := map[string]interface{}{
		"label_id":     labelID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "id", labelID)
		return labels.CoreLabel{}, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &label, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return labels.CoreLabel{}, errors.New("label not found")
		}
		r.log.Error(ctx, fmt.Sprintf("failed to execute query: %s", err), "id", labelID)
		return labels.CoreLabel{}, err
	}

	return toCoreLabel(label), nil
}
