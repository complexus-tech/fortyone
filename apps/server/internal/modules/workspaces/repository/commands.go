package workspacesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) Create(ctx context.Context, tx *sqlx.Tx, workspace workspaces.CoreWorkspace, createdBy uuid.UUID) (workspaces.CoreWorkspace, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Create")
	defer span.End()

	// Set trial_ends_on to 14 days from now
	trialEndsOn := time.Now().AddDate(0, 0, 14)

	var result dbWorkspace
	query := `
		INSERT INTO workspaces (
			name,
			slug,
			color,
			team_size,
			trial_ends_on,
			created_by
		)
		VALUES (
			:name,
			:slug,
			:color,
			:team_size,
			:trial_ends_on,
			:created_by
		)
		RETURNING
			workspace_id,
			name,
			slug,
			color,
			created_at,
			updated_at,
			trial_ends_on,
			created_by
	`

	params := map[string]any{
		"name":          workspace.Name,
		"slug":          workspace.Slug,
		"team_size":     workspace.TeamSize,
		"color":         generateRandomColor(),
		"trial_ends_on": trialEndsOn,
		"created_by":    createdBy,
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

	// Build SET clauses dynamically based on provided fields
	var setClauses []string
	params := map[string]any{
		"workspace_id": workspaceID,
	}

	// Only update fields that are provided (not empty/nil)
	if updates.Name != "" {
		setClauses = append(setClauses, "name = :name")
		params["name"] = updates.Name
	}

	if updates.AvatarURL != nil {
		setClauses = append(setClauses, "avatar_url = :avatar_url")
		params["avatar_url"] = *updates.AvatarURL
	}

	if len(setClauses) == 0 {
		return workspaces.CoreWorkspace{}, fmt.Errorf("no fields to update")
	}

	// Always update the timestamp
	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE workspaces
		SET %s
		WHERE 
			workspace_id = :workspace_id
		RETURNING
			workspace_id,
			name,
			color,
			slug,
			avatar_url,
			created_at,
			updated_at,
			trial_ends_on
	`, strings.Join(setClauses, ", "))

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}
	defer stmt.Close()

	var result dbWorkspace
	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to update workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspace{}, err
	}

	span.AddEvent("workspace updated", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("fields_updated", len(setClauses)-1), // -1 for updated_at
	))

	return toCoreWorkspace(result), nil
}

func (r *repo) Delete(ctx context.Context, workspaceID, deletedBy uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Delete")
	defer span.End()

	query := `
		UPDATE workspaces
		SET 
			deleted_at = NOW(),
			deleted_by = :deleted_by,
			updated_at = NOW()
		WHERE 
			workspace_id = :workspace_id
			AND deleted_at IS NULL
	`

	params := map[string]any{
		"workspace_id": workspaceID,
		"deleted_by":   deletedBy,
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
	objectiveParams := make(map[string]any)

	for i, status := range workspaces.DefaultObjectiveStatuses {
		paramPrefix := fmt.Sprintf("o%d_", i)
		objectiveValues[i] = fmt.Sprintf("(:%sname, :%scategory, :%sorder_index, :%scolor, :workspace_id)", paramPrefix, paramPrefix, paramPrefix, paramPrefix)
		objectiveParams[paramPrefix+"name"] = status.Name
		objectiveParams[paramPrefix+"category"] = status.Category
		objectiveParams[paramPrefix+"order_index"] = status.OrderIndex
		objectiveParams[paramPrefix+"color"] = status.Color
	}
	objectiveParams["workspace_id"] = workspaceID

	// Batch insert objective statuses
	objectiveQuery := fmt.Sprintf(`
		INSERT INTO objective_statuses (name, category, order_index, color, workspace_id)
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

func (r *repo) Restore(ctx context.Context, workspaceID, restoredBy uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.Restore")
	defer span.End()

	query := `
		UPDATE workspaces
		SET 
			deleted_at = NULL,
			deleted_by = NULL,
			updated_at = NOW()
		WHERE 
			workspace_id = :workspace_id
			AND deleted_at IS NOT NULL
	`

	params := map[string]any{
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
		errMsg := fmt.Sprintf("failed to restore workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to restore workspace"), trace.WithAttributes(attribute.String("error", errMsg)))
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

	params := map[string]any{
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("user %s is already a member of workspace %s", userID, workspaceID)
			r.log.Error(ctx, errMsg)
			span.RecordError(workspaces.ErrAlreadyWorkspaceMember, trace.WithAttributes(attribute.String("error", errMsg)))
			return workspaces.ErrAlreadyWorkspaceMember
		}
		errMsg := fmt.Sprintf("failed to add workspace member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to add workspace member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.RemoveMember")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to begin transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer tx.Rollback() // Rollback is a no-op if Commit has been called

	// Remove from workspace_members
	queryWorkspaceMembers := `
		DELETE FROM workspace_members
		WHERE
			workspace_id = :workspace_id
			AND user_id = :user_id
	`

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
	}

	stmtWorkspaceMembers, err := tx.PrepareNamedContext(ctx, queryWorkspaceMembers)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement for workspace_members: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement for workspace_members"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmtWorkspaceMembers.Close()

	resultWorkspaceMembers, err := stmtWorkspaceMembers.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to remove workspace member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to remove workspace member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffectedWorkspaceMembers, err := resultWorkspaceMembers.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffectedWorkspaceMembers == 0 {
		return workspaces.ErrMemberNotFound
	}

	// Remove from team_members
	queryTeamMembers := `
		DELETE FROM team_members
		WHERE user_id = :user_id
		AND team_id IN (
			SELECT team_id
			FROM teams
			WHERE workspace_id = :workspace_id
		)
	`
	// Parameters are the same: workspaceID and userID

	stmtTeamMembers, err := tx.PrepareNamedContext(ctx, queryTeamMembers)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement for team_members: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement for team_members"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmtTeamMembers.Close()

	_, err = stmtTeamMembers.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to remove member from teams: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to remove member from teams"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.UpdateMemberRole")
	defer span.End()

	query := `
		UPDATE workspace_members
		SET role = :role
		WHERE workspace_id = :workspace_id AND user_id = :user_id
	`

	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"role":         role,
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
		errMsg := fmt.Sprintf("failed to update member role: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update member role"), trace.WithAttributes(attribute.String("error", errMsg)))
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

// UpdateWorkspaceSettings updates the settings for a workspace
func (r *repo) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, settings workspaces.CoreWorkspaceSettings) (workspaces.CoreWorkspaceSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.UpdateWorkspaceSettings")
	defer span.End()

	var result dbWorkspaceSettings
	query := `
		UPDATE workspace_settings
		SET 
			story_term = CASE WHEN :story_term = '' THEN story_term ELSE :story_term END,
			sprint_term = CASE WHEN :sprint_term = '' THEN sprint_term ELSE :sprint_term END,
			objective_term = CASE WHEN :objective_term = '' THEN objective_term ELSE :objective_term END,
			key_result_term = CASE WHEN :key_result_term = '' THEN key_result_term ELSE :key_result_term END,
			objective_enabled = :objective_enabled,
			key_result_enabled = :key_result_enabled,
			updated_at = NOW()
		WHERE 
			workspace_id = :workspace_id
		RETURNING
			workspace_id,
			story_term,
			sprint_term,
			objective_term,
			key_result_term,
			objective_enabled,
			key_result_enabled,
			created_at,
			updated_at
	`

	params := map[string]any{
		"workspace_id":       workspaceID,
		"story_term":         settings.StoryTerm,
		"sprint_term":        settings.SprintTerm,
		"objective_term":     settings.ObjectiveTerm,
		"key_result_term":    settings.KeyResultTerm,
		"objective_enabled":  settings.ObjectiveEnabled,
		"key_result_enabled": settings.KeyResultEnabled,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspaceSettings{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return workspaces.CoreWorkspaceSettings{}, workspaces.ErrNotFound
		}
		errMsg := fmt.Sprintf("failed to update workspace settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update workspace settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return workspaces.CoreWorkspaceSettings{}, err
	}

	return toCoreWorkspaceSettings(result), nil
}

// InitializeWorkspaceSettings creates default settings for a new workspace
func (r *repo) InitializeWorkspaceSettings(ctx context.Context, tx *sqlx.Tx, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.workspaces.InitializeWorkspaceSettings")
	defer span.End()

	query := `
		INSERT INTO workspace_settings (
			workspace_id,
			story_term,
			sprint_term,
			objective_term,
			key_result_term,
			objective_enabled,
			key_result_enabled
		)
		VALUES (
			:workspace_id,
			'task',
			'sprint',
			'objective',
			'key result',
			true,
			true
		)
	`

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to initialize workspace settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to initialize workspace settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}
