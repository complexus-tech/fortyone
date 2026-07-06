package labelsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) GetLabels(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]labels.CoreLabel, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.GetLabels")
	defer span.End()
	var labels []dbLabel
	query := `
		SELECT 
			label_id,
			name,
			team_id,
			workspace_id,
			color,
			created_at,
			updated_at
		FROM labels
	`
	var setClauses []string
	filters["workspace_id"] = workspaceId
	search, hasSearch := filters["search"].(string)
	if hasSearch {
		search = strings.TrimSpace(search)
		if search == "" {
			delete(filters, "search")
			hasSearch = false
		} else {
			filters["search"] = search
		}
	}

	for field := range filters {
		switch field {
		case "search", "page", "pageSize", "limit", "offset":
			continue
		case "team_id":
			setClauses = append(setClauses, "(team_id = :team_id OR team_id IS NULL)")
		default:
			setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		}
	}
	if hasSearch {
		setClauses = append(setClauses, "name ILIKE '%' || :search || '%'")
	}

	query += " WHERE " + strings.Join(setClauses, " AND ") + " ORDER BY created_at DESC"
	if limit, ok := positiveIntFilter(filters, "limit"); ok {
		offset, _ := nonNegativeIntFilter(filters, "offset")
		filters["limit"] = limit
		filters["offset"] = offset
		query += " LIMIT :limit OFFSET :offset"
	}
	query += ";"

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

func (r *repo) GetLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) (labels.CoreLabel, error) {
	var label dbLabel
	query := `
		SELECT
			label_id,
			name,
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
