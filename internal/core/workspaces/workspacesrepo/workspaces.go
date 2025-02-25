package workspacesrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math/rand"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
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

var colors = []string{
	"#FFE066", "#FF6B6B", "#C0392B", "#FFA07A", "#FFB6C1",
	"#E056FD", "#686DE0", "#E67E22", "#A8E6CF", "#9B59B6", "#8E44AD",
	"#6BCB77", "#4ECDC4", "#4A90E2", "#95A5A6", "#27AE60", "#2ECC71",
	"#30336B", "#B4A6AB", "#636E72", "#34495E", "#2C3E50",
}

// generateRandomColor returns a random color from the Colors slice
func generateRandomColor() string {
	return colors[rand.Intn(len(colors))]
}

// List returns a list of workspaces a user is a member of.
func (r *repo) List(ctx context.Context, userID uuid.UUID) ([]workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.List")
	defer span.End()

	var workspaces []dbWorkspaceWithRole

	params := map[string]interface{}{
		"user_id": userID,
	}

	query := `
		SELECT DISTINCT
			w.workspace_id,
			w.slug,
			w.name,
			CASE 
				WHEN u.last_used_workspace_id = w.workspace_id THEN TRUE
				WHEN u.last_used_workspace_id IS NULL AND w.workspace_id = (
					SELECT w2.workspace_id 
					FROM workspaces w2 
					INNER JOIN workspace_members wm2 ON w2.workspace_id = wm2.workspace_id 
					WHERE wm2.user_id = :user_id
					ORDER BY w2.created_at ASC
					LIMIT 1
				) THEN TRUE
				ELSE FALSE
			END as is_active,
			wm.role as user_role,
			w.created_at,
			w.color,
			w.updated_at
		FROM
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		INNER JOIN
			users u ON wm.user_id = u.user_id
		WHERE
			wm.user_id = :user_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg, userID)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &workspaces, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve workspaces from the database: %s", err)
		span.RecordError(errors.New("failed to retrieve workspaces"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to retrieve workspaces %w", err)
	}

	return toCoreWorkspacesWithRole(workspaces), nil
}

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, workspace workspaces.CoreWorkspace) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Create")
	defer span.End()

	var result dbWorkspace
	query := `
		INSERT INTO workspaces (
			name,
			slug,
			color,
			team_size
		)
		VALUES (
			:name,
			:slug,
			:color,
			:team_size
		)
		RETURNING
			workspace_id,
			name,
			slug,
			color,
			created_at,
			updated_at
	`

	params := map[string]any{
		"name":      workspace.Name,
		"slug":      workspace.Slug,
		"team_size": workspace.TeamSize,
		"color":     generateRandomColor(),
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return workspaces.CoreWorkspace{}, workspaces.ErrSlugTaken
		}
		errMsg := fmt.Sprintf("failed to create workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	// Create default objective statuses for the workspace
	if err := r.createDefaultObjectiveStatuses(ctx, tx, result.ID); err != nil {
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspace(result), nil
}

func (r *repo) Update(ctx context.Context, workspaceID uuid.UUID, updates workspaces.CoreWorkspace) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Update")
	defer span.End()

	var result dbWorkspace
	query := `
		UPDATE workspaces
		SET 
			name = CASE WHEN :name = '' THEN name ELSE :name END,
			updated_at = NOW()
		WHERE 
			workspace_id = :workspace_id
		RETURNING
			workspace_id,
			name,
			color,
			slug,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"name":         updates.Name,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to update workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspace(result), nil
}

func (r *repo) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Delete")
	defer span.End()

	query := `
		DELETE FROM workspaces
		WHERE 
			workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
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
		errMsg := fmt.Sprintf("failed to delete workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return workspaces.ErrNotFound
	}

	return nil
}

// createDefaultObjectiveStatuses creates the default objective statuses for a workspace
func (r *repo) createDefaultObjectiveStatuses(ctx context.Context, tx *sqlx.Tx, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.createDefaultObjectiveStatuses")
	defer span.End()

	// Build values for objective statuses batch insert
	objectiveValues := make([]string, len(workspaces.DefaultObjectiveStatuses))
	objectiveParams := make(map[string]interface{})

	for i, status := range workspaces.DefaultObjectiveStatuses {
		paramPrefix := fmt.Sprintf("o%d_", i)
		objectiveValues[i] = fmt.Sprintf("(:%sname, :%scategory, :%sorder_index, :workspace_id)", paramPrefix, paramPrefix, paramPrefix)
		objectiveParams[paramPrefix+"name"] = status.Name
		objectiveParams[paramPrefix+"category"] = status.Category
		objectiveParams[paramPrefix+"order_index"] = status.OrderIndex
	}
	objectiveParams["workspace_id"] = workspaceID

	// Batch insert objective statuses
	objectiveQuery := fmt.Sprintf(`
		INSERT INTO objective_statuses (name, category, order_index, workspace_id)
		VALUES %s
	`, strings.Join(objectiveValues, ","))

	if _, err := tx.NamedExecContext(ctx, objectiveQuery, objectiveParams); err != nil {
		errMsg := fmt.Sprintf("failed to create objective statuses: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create objective statuses"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	return r.addMemberImpl(ctx, r.db, workspaceID, userID, role)
}

func (r *repo) AddMemberTx(ctx context.Context, tx *sqlx.Tx, workspaceID, userID uuid.UUID, role string) error {
	return r.addMemberImpl(ctx, tx, workspaceID, userID, role)
}

func (r *repo) addMemberImpl(ctx context.Context, executor sqlx.ExtContext, workspaceID, userID uuid.UUID, role string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.addMemberImpl")
	defer span.End()

	query := `
		INSERT INTO workspace_members (
			workspace_id,
			user_id,
			role
		)
		VALUES (
			:workspace_id,
			:user_id,
			:role
		)
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"role":         role,
	}

	// Use type assertion to get the correct type for PrepareNamedContext
	var stmt *sqlx.NamedStmt
	var err error
	switch e := executor.(type) {
	case *sqlx.DB:
		stmt, err = e.PrepareNamedContext(ctx, query)
	case *sqlx.Tx:
		stmt, err = e.PrepareNamedContext(ctx, query)
	default:
		return fmt.Errorf("unsupported executor type: %T", executor)
	}

	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to add workspace member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to add workspace member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) Get(ctx context.Context, workspaceID, userID uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Get")
	defer span.End()

	var workspace dbWorkspaceWithRole
	query := `
		SELECT 
			w.workspace_id,
			w.slug,
			w.name,
			CASE 
				WHEN u.last_used_workspace_id = w.workspace_id THEN TRUE
				WHEN u.last_used_workspace_id IS NULL AND w.workspace_id = (
					SELECT w2.workspace_id 
					FROM workspaces w2 
					INNER JOIN workspace_members wm2 ON w2.workspace_id = wm2.workspace_id 
					WHERE wm2.user_id = :user_id
					ORDER BY w2.created_at ASC
					LIMIT 1
				) THEN TRUE
				ELSE FALSE
			END as is_active,
			wm.role as user_role,
			w.color,
			w.created_at,
			w.updated_at
		FROM 
			workspaces w
		INNER JOIN
			workspace_members wm ON w.workspace_id = wm.workspace_id
		INNER JOIN
			users u ON wm.user_id = u.user_id
		WHERE 
			w.workspace_id = :workspace_id
			AND wm.user_id = :user_id
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &workspace, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to get workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	return toCoreWorkspaceWithRole(workspace), nil
}

func (r *repo) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.RemoveMember")
	defer span.End()

	query := `
		DELETE FROM workspace_members
		WHERE 
			workspace_id = :workspace_id
			AND user_id = :user_id
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
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
		errMsg := fmt.Sprintf("failed to remove workspace member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to remove workspace member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return workspaces.ErrNotFound
	}

	return nil
}

func (r *repo) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.CheckSlugAvailability")
	defer span.End()

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM workspaces 
			WHERE slug = :slug
		)
	`

	params := map[string]interface{}{
		"slug": slug,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		errMsg := fmt.Sprintf("failed to check slug availability: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to check slug"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}

	return !exists, nil
}
