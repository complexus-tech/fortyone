package objectivesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) Create(ctx context.Context, objective objectives.CoreNewObjective, workspaceID uuid.UUID, keyResults []keyresults.CoreNewKeyResult) (objectives.CoreObjective, []keyresults.CoreKeyResult, error) {
	r.log.Info(ctx, "business.repository.objectives.Create")
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.Create")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return objectives.CoreObjective{}, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert objective
	const objQuery = `
		INSERT INTO objectives (
			name, description, lead_user_id, team_id,
			workspace_id, start_date, end_date, is_private,
			status_id, priority, created_by
		) VALUES (
			:name, :description, :lead_user_id, :team_id,
			:workspace_id, :start_date, :end_date, :is_private,
			:status_id, :priority, :created_by
		) RETURNING objectives.objective_id, objectives.name, objectives.description, objectives.lead_user_id, objectives.team_id, objectives.workspace_id, objectives.start_date, objectives.end_date, objectives.is_private, objectives.status_id, objectives.priority, objectives.created_at, objectives.updated_at, objectives.created_by, objectives.health;
	`

	var createdObj dbObjective
	stmt, err := tx.PrepareNamedContext(ctx, objQuery)
	if err != nil {
		return objectives.CoreObjective{}, nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &createdObj, toDBObjective(objective, workspaceID)); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("objective name %s already exists", objective.Name)
			r.log.Error(ctx, errMsg)
			span.RecordError(objectives.ErrNameExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return objectives.CoreObjective{}, nil, objectives.ErrNameExists
		}
		return objectives.CoreObjective{}, nil, err
	}

	var createdKRs []keyresults.CoreKeyResult
	if len(keyResults) > 0 {
		// Bulk insert key results
		const krQuery = `
			INSERT INTO key_results (
				objective_id, name, measurement_type,
				start_value, current_value, target_value, created_by, lead, start_date, end_date
			) VALUES (
				:objective_id, :name, :measurement_type,
				:start_value, :current_value, :target_value, :created_by, :lead, :start_date, :end_date
			) RETURNING *;
		`

		collaboratorsQuery := `
        INSERT INTO key_result_contributors (
            key_result_id, user_id, created_at, updated_at
        ) VALUES (
            :key_result_id, :user_id, NOW(), NOW()
        )
    `

		krstmt, err := tx.PrepareNamedContext(ctx, krQuery)
		if err != nil {
			return objectives.CoreObjective{}, nil, err
		}
		defer krstmt.Close()

		for _, kr := range keyResults {
			kr.ObjectiveID = createdObj.ID
			var dbKR dbKeyResult
			if err := krstmt.GetContext(ctx, &dbKR, toDBKeyResult(kr, kr.CreatedBy)); err != nil {
				return objectives.CoreObjective{}, nil, err
			}

			createdKRs = append(createdKRs, toCoreKeyResult(dbKR))

			collaboratorsStmt, err := tx.PrepareNamedContext(ctx, collaboratorsQuery)
			if err != nil {
				return objectives.CoreObjective{}, nil, err
			}
			defer collaboratorsStmt.Close()

			for _, contributor := range kr.Contributors {
				if err := collaboratorsStmt.GetContext(ctx, nil, map[string]any{
					"key_result_id": dbKR.ID,
					"user_id":       contributor,
				}); err != nil {
					return objectives.CoreObjective{}, nil, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return objectives.CoreObjective{}, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return toCoreObjective(createdObj), createdKRs, nil
}

// Update updates an objective
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.objectives.Update")
	defer span.End()

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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			// Get the name from updates if it exists
			nameValue, hasName := updates["name"]
			name := ""
			if hasName {
				name, _ = nameValue.(string)
			}
			errMsg := fmt.Sprintf("objective name %s already exists", name)
			r.log.Error(ctx, errMsg)
			span.RecordError(objectives.ErrNameExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return objectives.ErrNameExists
		}
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

	query := `
		DELETE FROM objectives
		WHERE objective_id = :objective_id
		AND workspace_id = :workspace_id
	`

	params := map[string]any{
		"objective_id": id,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete objective: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete objective"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
