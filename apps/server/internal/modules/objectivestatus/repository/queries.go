package objectivestatusrepository

import (
	"context"
	"errors"
	"fmt"

	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.List")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceId,
	}

	var statuses []dbObjectiveStatus
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			workspace_id,
			is_default,
			color,
			created_at,
			updated_at
		FROM
			objective_statuses
		WHERE workspace_id = :workspace_id
		ORDER BY order_index ASC;
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objective statuses.")
	if err := stmt.SelectContext(ctx, &statuses, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objective statuses from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("statuses not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Objective statuses retrieved successfully.")
	span.AddEvent("Objective statuses retrieved.", trace.WithAttributes(
		attribute.Int("statuses.count", len(statuses)),
		attribute.String("query", q),
	))

	return toCoreObjectiveStatuses(statuses), nil
}

func (r *repo) CountObjectivesWithStatus(ctx context.Context, statusID uuid.UUID, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.CountObjectivesWithStatus")
	defer span.End()

	params := map[string]any{
		"status_id":    statusID,
		"workspace_id": workspaceID,
	}

	q := `
		SELECT COUNT(*)
		FROM objectives
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count objectives statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count objectives with status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

func (r *repo) CountStatusesInCategory(ctx context.Context, workspaceID uuid.UUID, category string) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.CountStatusesInCategory")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceID,
		"category":     category,
	}

	q := `
		SELECT COUNT(*)
		FROM objective_statuses
		WHERE workspace_id = :workspace_id
		AND category = :category
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count statuses statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count statuses in category: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

func (r *repo) Get(ctx context.Context, workspaceId, statusId uuid.UUID) (objectivestatus.CoreObjectiveStatus, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.objectivestatus.Get")
	defer span.End()

	params := map[string]any{
		"status_id":    statusId,
		"workspace_id": workspaceId,
	}

	var status dbObjectiveStatus
	q := `
		SELECT
			status_id,
			name,
			category,
			order_index,
			workspace_id,
			is_default,
			color,
			created_at,
			updated_at
		FROM
			objective_statuses
		WHERE status_id = :status_id
		AND workspace_id = :workspace_id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching objective status.")
	if err := stmt.GetContext(ctx, &status, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve objective status from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("status not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return objectivestatus.CoreObjectiveStatus{}, err
	}

	r.log.Info(ctx, "Objective status retrieved successfully.")
	span.AddEvent("Objective status retrieved.", trace.WithAttributes(
		attribute.String("status.id", statusId.String()),
		attribute.String("query", q),
	))

	return toCoreObjectiveStatus(status), nil
}
