package keyresultsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func contributorsJSON(contributorIDs []uuid.UUID) *json.RawMessage {
	encoded, err := json.Marshal(contributorIDs)
	if err != nil {
		encoded = []byte("[]")
	}
	raw := json.RawMessage(encoded)
	return &raw
}

// Create inserts a new key result into the database
func (r *repo) Create(ctx context.Context, kr *CoreKeyResult, workspaceID uuid.UUID) (uuid.UUID, int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Create")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("begin key result transaction: %w", err)
	}
	defer tx.Rollback()

	var scope struct {
		TeamID uuid.UUID `db:"team_id"`
	}
	if err := tx.GetContext(ctx, &scope, `
		SELECT team_id
		FROM objectives
		WHERE objective_id = $1
			AND workspace_id = $2
	`, kr.ObjectiveID, workspaceID); err != nil {
		return uuid.Nil, 0, fmt.Errorf("resolve key result team: %w", err)
	}

	var sequenceID int
	if err := tx.GetContext(ctx, &sequenceID, `
		INSERT INTO team_key_result_sequences (
			workspace_id,
			team_id,
			current_sequence
		) VALUES ($1, $2, 1)
		ON CONFLICT (workspace_id, team_id)
		DO UPDATE SET current_sequence =
			team_key_result_sequences.current_sequence + 1
		RETURNING current_sequence
	`, workspaceID, scope.TeamID); err != nil {
		return uuid.Nil, 0, fmt.Errorf("allocate key result sequence: %w", err)
	}

	const q = `
		INSERT INTO key_results (
			objective_id, team_id, sequence_id, name, measurement_type,
			start_value, current_value, target_value,
			lead, start_date, end_date, created_by
		) VALUES (
			:objective_id, :team_id, :sequence_id, :name, :measurement_type,
			:start_value, :current_value, :target_value,
			:lead, :start_date, :end_date, :created_by
		) RETURNING id
	`

	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return uuid.Nil, 0, err
	}
	defer stmt.Close()

	params := map[string]any{
		"objective_id":     kr.ObjectiveID,
		"team_id":          scope.TeamID,
		"sequence_id":      sequenceID,
		"name":             kr.Name,
		"measurement_type": kr.MeasurementType,
		"start_value":      kr.StartValue,
		"current_value":    kr.CurrentValue,
		"target_value":     kr.TargetValue,
		"lead":             kr.Lead,
		"start_date":       kr.StartDate,
		"end_date":         kr.EndDate,
		"created_by":       kr.CreatedBy,
	}

	r.log.Info(ctx, "creating key result")
	var id uuid.UUID
	if err := stmt.GetContext(ctx, &id, params); err != nil {
		errMsg := fmt.Sprintf("failed to create key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return uuid.Nil, 0, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, 0, fmt.Errorf("commit key result transaction: %w", err)
	}

	r.log.Info(ctx, "key result created successfully")
	span.AddEvent("key result created", trace.WithAttributes(
		attribute.String("key_result.name", kr.Name),
	))

	return id, sequenceID, nil
}

// CreateBatch inserts key results for one objective in a single transaction.
func (r *repo) CreateBatch(ctx context.Context, keyResults []CoreKeyResult, workspaceID uuid.UUID) ([]CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.CreateBatch")
	defer span.End()

	if len(keyResults) == 0 {
		return []CoreKeyResult{}, nil
	}

	objectiveID := keyResults[0].ObjectiveID
	for _, keyResult := range keyResults[1:] {
		if keyResult.ObjectiveID != objectiveID {
			return nil, errors.New("all key results in a batch must belong to the same objective")
		}
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin key result batch transaction: %w", err)
	}
	defer tx.Rollback()

	var teamID uuid.UUID
	if err := tx.GetContext(ctx, &teamID, `
		SELECT team_id
		FROM objectives
		WHERE objective_id = $1
			AND workspace_id = $2
	`, objectiveID, workspaceID); err != nil {
		return nil, fmt.Errorf("resolve key result batch objective: %w", err)
	}

	var finalSequenceID int
	if err := tx.GetContext(ctx, &finalSequenceID, `
		INSERT INTO team_key_result_sequences (
			workspace_id,
			team_id,
			current_sequence
		) VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, team_id)
		DO UPDATE SET current_sequence =
			team_key_result_sequences.current_sequence + EXCLUDED.current_sequence
		RETURNING current_sequence
	`, workspaceID, teamID, len(keyResults)); err != nil {
		return nil, fmt.Errorf("allocate key result batch sequences: %w", err)
	}
	firstSequenceID := finalSequenceID - len(keyResults) + 1

	const insertKeyResult = `
		INSERT INTO key_results (
			objective_id, team_id, sequence_id, name, measurement_type,
			start_value, current_value, target_value,
			lead, start_date, end_date, created_by
		) VALUES (
			:objective_id, :team_id, :sequence_id, :name, :measurement_type,
			:start_value, :current_value, :target_value,
			:lead, :start_date, :end_date, :created_by
		)
		RETURNING id, sequence_id, objective_id, name, measurement_type,
			start_value, current_value, target_value, lead, start_date, end_date,
			created_at, updated_at, created_by
	`
	keyResultStatement, err := tx.PrepareNamedContext(ctx, insertKeyResult)
	if err != nil {
		return nil, fmt.Errorf("prepare key result batch insert: %w", err)
	}
	defer keyResultStatement.Close()

	const insertContributor = `
		INSERT INTO key_result_contributors (key_result_id, user_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`
	contributorStatement, err := tx.PrepareContext(ctx, insertContributor)
	if err != nil {
		return nil, fmt.Errorf("prepare key result batch contributor insert: %w", err)
	}
	defer contributorStatement.Close()

	createdKeyResults := make([]CoreKeyResult, 0, len(keyResults))
	for i, keyResult := range keyResults {
		var created dbKeyResult
		params := map[string]any{
			"objective_id":     objectiveID,
			"team_id":          teamID,
			"sequence_id":      firstSequenceID + i,
			"name":             keyResult.Name,
			"measurement_type": keyResult.MeasurementType,
			"start_value":      keyResult.StartValue,
			"current_value":    keyResult.CurrentValue,
			"target_value":     keyResult.TargetValue,
			"lead":             keyResult.Lead,
			"start_date":       keyResult.StartDate,
			"end_date":         keyResult.EndDate,
			"created_by":       keyResult.CreatedBy,
		}
		if err := keyResultStatement.GetContext(ctx, &created, params); err != nil {
			return nil, fmt.Errorf("insert key result batch item %d: %w", i, err)
		}

		for _, contributorID := range keyResult.Contributors {
			if _, err := contributorStatement.ExecContext(ctx, created.ID, contributorID); err != nil {
				return nil, fmt.Errorf("insert key result batch contributor: %w", err)
			}
		}

		created.Contributors = contributorsJSON(keyResult.Contributors)
		createdKeyResults = append(createdKeyResults, toCoreKeyResult(created))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit key result batch transaction: %w", err)
	}

	return createdKeyResults, nil
}

// Update modifies data about a key result
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Update")
	defer span.End()

	// Verify the key result exists and belongs to the workspace
	if _, err := r.getKeyResultById(ctx, id, workspaceId); err != nil {
		return err
	}

	query := "UPDATE key_results SET "
	var setClauses []string
	params := map[string]any{"id": id}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating key result #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to update key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Key result #%s updated successfully", id), "id", id)
	span.AddEvent("key result updated", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

func (r *repo) UpdateIfUnchanged(
	ctx context.Context,
	id, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) (bool, error) {
	if len(updates) == 0 {
		return false, errors.New("key result compare-and-swap update requires at least one field")
	}
	fields := make([]string, 0, len(updates))
	for field := range updates {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	setClauses := make([]string, 0, len(fields)+1)
	params := map[string]any{
		"id":                  id,
		"workspace_id":        workspaceID,
		"expected_updated_at": expectedUpdatedAt.UTC(),
	}
	for _, field := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = updates[field]
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	query := "UPDATE key_results SET " + strings.Join(setClauses, ", ")
	query += " WHERE id = :id AND workspace_id = :workspace_id AND updated_at = :expected_updated_at"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare key result compare-and-swap update: %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return false, fmt.Errorf("update key result if unchanged: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read key result compare-and-swap result: %w", err)
	}
	return rowsAffected == 1, nil
}

// Delete deletes a key result
func (r *repo) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Delete")
	defer span.End()

	// Verify the key result exists and belongs to the workspace
	if _, err := r.getKeyResultById(ctx, id, workspaceId); err != nil {
		return err
	}

	const q = `
		DELETE FROM key_results
		WHERE id = :id
	`

	params := map[string]any{
		"id": id,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "deleting key result")
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to delete key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "key result deleted successfully")
	span.AddEvent("key result deleted", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

// AddContributors adds contributors to a key result
func (r *repo) AddContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.AddContributors")
	defer span.End()

	if len(contributorIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO key_result_contributors (
			key_result_id,
			user_id,
			created_at,
			updated_at
		) VALUES (
			:key_result_id,
			:user_id,
			NOW(),
			NOW()
		)
	`

	for _, contributorID := range contributorIDs {
		params := map[string]any{
			"key_result_id": keyResultID,
			"user_id":       contributorID,
		}

		if _, err := r.db.NamedExecContext(ctx, query, params); err != nil {
			errMsg := fmt.Sprintf("failed to add contributor: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to add contributor"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
	}

	span.AddEvent("contributors added", trace.WithAttributes(
		attribute.Int("contributors.count", len(contributorIDs)),
		attribute.String("key_result.id", keyResultID.String()),
	))

	return nil
}

// UpdateContributors replaces all contributors for a key result
func (r *repo) UpdateContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.UpdateContributors")
	defer span.End()

	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing contributors
	deleteQuery := `DELETE FROM key_result_contributors WHERE key_result_id = :key_result_id`
	deleteParams := map[string]any{"key_result_id": keyResultID}

	if _, err := tx.NamedExecContext(ctx, deleteQuery, deleteParams); err != nil {
		errMsg := fmt.Sprintf("failed to delete existing contributors: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete contributors"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	// Add new contributors if any
	if len(contributorIDs) > 0 {
		insertQuery := `
			INSERT INTO key_result_contributors (
				key_result_id,
				user_id,
				created_at,
				updated_at
			) VALUES (
				:key_result_id,
				:user_id,
				NOW(),
				NOW()
			)
		`

		for _, contributorID := range contributorIDs {
			params := map[string]any{
				"key_result_id": keyResultID,
				"user_id":       contributorID,
			}

			if _, err := tx.NamedExecContext(ctx, insertQuery, params); err != nil {
				errMsg := fmt.Sprintf("failed to add contributor: %s", err)
				r.log.Error(ctx, errMsg)
				span.RecordError(errors.New("failed to add contributor"), trace.WithAttributes(attribute.String("error", errMsg)))
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	span.AddEvent("contributors updated", trace.WithAttributes(
		attribute.Int("contributors.count", len(contributorIDs)),
		attribute.String("key_result.id", keyResultID.String()),
	))

	return nil
}
