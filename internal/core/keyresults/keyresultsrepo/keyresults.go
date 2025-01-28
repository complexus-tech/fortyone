package keyresultsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository errors
var (
	ErrNotFound = errors.New("key result not found")
)

// Repository defines the repository for key results
type Repository interface {
	Create(ctx context.Context, kr *CoreKeyResult) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error)
	List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error)
}

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

// New creates a new key results repository
func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// Create creates a new key result
func (r *repo) Create(ctx context.Context, kr *CoreKeyResult) error {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Create")
	defer span.End()

	const q = `
		INSERT INTO key_results (
			id, objective_id, name, measurement_type,
			start_value, target_value, created_at, updated_at
		) VALUES (
			:id, :objective_id, :name, :measurement_type,
			:start_value, :target_value, :created_at, :updated_at
		)
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, "creating key result")
	if _, err := stmt.ExecContext(ctx, toDBKeyResult(*kr)); err != nil {
		errMsg := fmt.Sprintf("failed to create key result: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create key result"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "key result created successfully")
	span.AddEvent("key result created", trace.WithAttributes(
		attribute.String("key_result.name", kr.Name),
	))

	return nil
}

// Update updates a key result
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
	`

	params := map[string]interface{}{
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

// Get retrieves a key result by ID
func (r *repo) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.Get")
	defer span.End()

	kr, err := r.getKeyResultById(ctx, id, workspaceId)
	if err != nil {
		return CoreKeyResult{}, err
	}

	return toCoreKeyResult(kr), nil
}

// List retrieves all key results for an objective
func (r *repo) List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.keyresults.List")
	defer span.End()

	const q = `
		SELECT kr.*
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		WHERE kr.objective_id = :objective_id
		AND o.workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"objective_id": objectiveId,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	var keyResults []dbKeyResult
	if err := stmt.SelectContext(ctx, &keyResults, params); err != nil {
		errMsg := fmt.Sprintf("failed to list key results: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list key results"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreKeyResults(keyResults), nil
}

// getKeyResultById retrieves a key result by ID and verifies it belongs to the workspace
func (r *repo) getKeyResultById(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (dbKeyResult, error) {
	const q = `
		SELECT kr.*
		FROM key_results kr
		INNER JOIN objectives o ON kr.objective_id = o.objective_id
		WHERE kr.id = :id
		AND o.workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"id":           id,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err))
		return dbKeyResult{}, err
	}
	defer stmt.Close()

	var kr dbKeyResult
	if err := stmt.GetContext(ctx, &kr, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get key result: %s", err))
		return dbKeyResult{}, ErrNotFound
	}

	return kr, nil
}
