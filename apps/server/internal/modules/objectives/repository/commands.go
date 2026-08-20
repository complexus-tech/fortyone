package objectivesrepository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
	defer tx.Rollback()

	var sequenceID int
	if err := tx.GetContext(ctx, &sequenceID, `
		INSERT INTO team_objective_sequences (
			workspace_id,
			team_id,
			current_sequence
		) VALUES ($1, $2, 1)
		ON CONFLICT (workspace_id, team_id)
		DO UPDATE SET current_sequence =
			team_objective_sequences.current_sequence + 1
		RETURNING current_sequence
	`, workspaceID, objective.Team); err != nil {
		return objectives.CoreObjective{}, nil, fmt.Errorf("allocate objective sequence: %w", err)
	}

	// Insert objective
	const objQuery = `
		INSERT INTO objectives (
			sequence_id, name, description, short_summary, lead_user_id, team_id,
			workspace_id, start_date, end_date, is_private,
			status_id, priority, color, created_by
		) VALUES (
			:sequence_id, :name, :description, :short_summary, :lead_user_id, :team_id,
			:workspace_id, :start_date, :end_date, :is_private,
			:status_id, :priority, :color, :created_by
		) RETURNING objectives.objective_id, objectives.sequence_id, objectives.name, objectives.description, objectives.short_summary, objectives.lead_user_id, objectives.team_id, objectives.workspace_id, objectives.start_date, objectives.end_date, objectives.is_private, objectives.status_id, objectives.priority, objectives.color, objectives.created_at, objectives.updated_at, objectives.created_by, objectives.health;
	`

	var createdObj dbObjective
	stmt, err := tx.PrepareNamedContext(ctx, objQuery)
	if err != nil {
		return objectives.CoreObjective{}, nil, err
	}
	defer stmt.Close()

	objectiveParams := toDBObjective(objective, workspaceID)
	objectiveParams.SequenceID = sequenceID
	if err := stmt.GetContext(ctx, &createdObj, objectiveParams); err != nil {
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
		writer, err := newSQLObjectiveKeyResultWriter(ctx, tx)
		if err != nil {
			return objectives.CoreObjective{}, nil, err
		}
		defer writer.Close()

		createdKRs, err = createObjectiveKeyResults(
			ctx,
			writer,
			createdObj.ID,
			workspaceID,
			objective.Team,
			keyResults,
		)
		if err != nil {
			return objectives.CoreObjective{}, nil, err
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

	query, params := buildObjectiveUpdateStatement(id, workspaceId, updates)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating objective #%s", id), "id", id)
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting updated objective row count: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	r.log.Info(ctx, fmt.Sprintf("Objective #%s updated successfully", id), "id", id)
	span.AddEvent("objective updated", trace.WithAttributes(
		attribute.String("objective.id", id.String()),
	))

	return nil
}

func (r *repo) UpdateIfUnchanged(
	ctx context.Context,
	id, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) (bool, error) {
	query, params := buildObjectiveUpdateStatement(id, workspaceID, updates)
	query += " AND updated_at = :expected_updated_at"
	params["expected_updated_at"] = expectedUpdatedAt.UTC()

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare objective compare-and-swap update: %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return false, fmt.Errorf("update objective if unchanged: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read objective compare-and-swap result: %w", err)
	}
	return rowsAffected == 1, nil
}

func buildObjectiveUpdateStatement(id, workspaceID uuid.UUID, updates map[string]any) (string, map[string]any) {
	fields := make([]string, 0, len(updates))
	for field := range updates {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	setClauses := make([]string, 0, len(fields)+1)
	params := map[string]any{
		"id":           id,
		"workspace_id": workspaceID,
	}
	for _, field := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = updates[field]
	}
	setClauses = append(setClauses, "updated_at = NOW()")

	query := "UPDATE objectives SET " + strings.Join(setClauses, ", ")
	query += " WHERE objective_id = :id AND workspace_id = :workspace_id"
	return query, params
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
