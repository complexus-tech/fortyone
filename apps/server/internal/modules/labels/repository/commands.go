package labelsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) CreateLabel(ctx context.Context, cnl labels.CoreNewLabel) (labels.CoreLabel, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.CreateLabel")
	defer span.End()
	var label dbLabel
	query := `
			INSERT INTO
				labels (name, team_id, workspace_id, color)
			VALUES
				(:name, :team_id, :workspace_id, :color)
			RETURNING
				label_id, name, team_id,
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

func (r *repo) UpdateLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID, name string, color string) (labels.CoreLabel, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.UpdateLabel")
	defer span.End()

	var label dbLabel
	query := `
		UPDATE labels
		SET 
			name = :name,
			color = :color,
			updated_at = NOW()
		WHERE 
			label_id = :label_id
			AND workspace_id = :workspace_id
		RETURNING
			label_id, name, team_id,
			workspace_id, color, created_at,
			updated_at
	`

	params := map[string]interface{}{
		"label_id":     labelID,
		"workspace_id": workspaceID,
		"name":         name,
		"color":        color,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return labels.CoreLabel{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Updating label.")
	if err := stmt.GetContext(ctx, &label, params); err != nil {
		if err == sql.ErrNoRows {
			return labels.CoreLabel{}, errors.New("label not found")
		}
		errMsg := fmt.Sprintf("Failed to update label: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update label"), trace.WithAttributes(attribute.String("error", errMsg)))
		return labels.CoreLabel{}, err
	}

	return toCoreLabel(label), nil
}

func (r *repo) DeleteLabel(ctx context.Context, labelID uuid.UUID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.labels.DeleteLabel")
	defer span.End()

	query := `
		DELETE FROM labels
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
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Deleting label.")
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to delete label: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete label"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("label not found")
	}

	return nil
}
