package keyresultsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Create inserts a new key result into the database
func (r *repo) Create(ctx context.Context, kr *CoreKeyResult) (uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Create")
	defer span.End()

	const q = `
		INSERT INTO key_results (
			objective_id, name, measurement_type,
			start_value, current_value, target_value,
			lead, start_date, end_date, created_by
		) VALUES (
			:objective_id, :name, :measurement_type,
			:start_value, :current_value, :target_value,
			:lead, :start_date, :end_date, :created_by
		) RETURNING id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return uuid.Nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "creating key result")
	var id uuid.UUID
	if err := stmt.GetContext(ctx, &id, toDBKeyResult(*kr)); err != nil {
		errMsg := fmt.Sprintf("failed to create key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return uuid.Nil, err
	}

	r.log.Info(ctx, "key result created successfully")
	span.AddEvent("key result created", trace.WithAttributes(
		attribute.String("key_result.name", kr.Name),
	))

	return id, nil
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
		AND workspace_id = :workspace_id
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
