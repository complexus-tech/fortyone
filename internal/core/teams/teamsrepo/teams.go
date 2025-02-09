package teamsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/teams"
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

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID) ([]teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.List")
	defer span.End()

	params := map[string]interface{}{
		"workspace_id": workspaceId,
	}
	var teams []dbTeam
	q := `
		SELECT
			team_id,
			name,
			description,
			code,
			color,
			icon,
			workspace_id,
			created_at,
			updated_at
		FROM
			teams
		WHERE
			workspace_id = :workspace_id
		ORDER BY created_at DESC;
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching teams.")
	if err := stmt.SelectContext(ctx, &teams, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve teams from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("teams not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "teams retrieved successfully.")
	span.AddEvent("teams retrieved.", trace.WithAttributes(
		attribute.Int("teams.count", len(teams)),
		attribute.String("query", q),
	))

	return toCoreTeams(teams), nil
}

func (r *repo) Create(ctx context.Context, team teams.CoreTeam) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "teamsrepo.Create")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to begin transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}
	defer tx.Rollback()

	params := map[string]interface{}{
		"name":         team.Name,
		"description":  team.Description,
		"code":         team.Code,
		"color":        team.Color,
		"icon":         team.Icon,
		"workspace_id": team.Workspace,
	}

	query := `
		INSERT INTO teams (name, description, code, color, icon, workspace_id)
		VALUES (:name, :description, :code, :color, :icon, :workspace_id)
		RETURNING team_id, name, description, code, color, icon, workspace_id, created_at, updated_at
	`

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}
	defer stmt.Close()

	var dbTeam dbTeam
	if err := stmt.GetContext(ctx, &dbTeam, params); err != nil {
		errMsg := fmt.Sprintf("failed to execute query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to execute query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	if err := r.createDefaultStatuses(ctx, tx, dbTeam.ID, dbTeam.Workspace); err != nil {
		errMsg := fmt.Sprintf("failed to create default statuses: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create default statuses"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	return toCoreTeam(dbTeam), nil
}

func (r *repo) createDefaultStatuses(ctx context.Context, tx *sqlx.Tx, teamID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "teamsrepo.createDefaultStatuses")
	defer span.End()

	// Create story statuses
	storyQuery := `
		INSERT INTO statuses (name, category, order_index, team_id, workspace_id)
		VALUES (:name, :category, :order_index, :team_id, :workspace_id)
	`
	storyStmt, err := tx.PrepareNamedContext(ctx, storyQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare story status statement: %w", err)
	}
	defer storyStmt.Close()

	for _, status := range teams.DefaultStoryStatuses {
		params := map[string]interface{}{
			"name":         status.Name,
			"category":     status.Category,
			"order_index":  status.OrderIndex,
			"team_id":      teamID,
			"workspace_id": workspaceID,
		}
		if _, err := storyStmt.ExecContext(ctx, params); err != nil {
			return fmt.Errorf("failed to create story status: %w", err)
		}
	}

	// Create objective statuses
	objectiveQuery := `
		INSERT INTO objective_statuses (name, category, order_index, team_id, workspace_id)
		VALUES (:name, :category, :order_index, :team_id, :workspace_id)
	`
	objectiveStmt, err := tx.PrepareNamedContext(ctx, objectiveQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare objective status statement: %w", err)
	}
	defer objectiveStmt.Close()

	for _, status := range teams.DefaultObjectiveStatuses {
		params := map[string]interface{}{
			"name":         status.Name,
			"category":     status.Category,
			"order_index":  status.OrderIndex,
			"team_id":      teamID,
			"workspace_id": workspaceID,
		}
		if _, err := objectiveStmt.ExecContext(ctx, params); err != nil {
			return fmt.Errorf("failed to create objective status: %w", err)
		}
	}

	return nil
}

func (r *repo) Update(ctx context.Context, teamID uuid.UUID, updates teams.CoreTeam) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.Update")
	defer span.End()

	var result dbTeam
	query := `
		UPDATE teams
		SET 
			name = CASE WHEN :name = '' THEN name ELSE :name END,
			description = CASE WHEN :description IS NULL THEN description ELSE :description END,
			code = CASE WHEN :code = '' THEN code ELSE :code END,
			color = CASE WHEN :color = '' THEN color ELSE :color END,
			icon = CASE WHEN :icon = '' THEN icon ELSE :icon END,
			updated_at = NOW()
		WHERE 
			team_id = :team_id
			AND workspace_id = :workspace_id
		RETURNING
			team_id,
			name,
			description,
			code,
			color,
			icon,
			workspace_id,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"team_id":      teamID,
		"workspace_id": updates.Workspace,
		"name":         updates.Name,
		"description":  updates.Description,
		"code":         updates.Code,
		"color":        updates.Color,
		"icon":         updates.Icon,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return teams.CoreTeam{}, errors.New("team not found")
		}
		errMsg := fmt.Sprintf("failed to update team: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update team"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	return toCoreTeam(result), nil
}

func (r *repo) Delete(ctx context.Context, teamID uuid.UUID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.Delete")
	defer span.End()

	query := `
		DELETE FROM teams
		WHERE 
			team_id = :team_id
			AND workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"team_id":      teamID,
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
		errMsg := fmt.Sprintf("failed to delete team: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete team"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("team not found")
	}

	return nil
}

func (r *repo) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.AddMember")
	defer span.End()

	query := `
		INSERT INTO team_members (
			team_id,
			user_id,
			role,
			created_at,
			updated_at
		)
		VALUES (
			:team_id,
			:user_id,
			:role,
			NOW(),
			NOW()
		)
	`

	params := map[string]interface{}{
		"team_id": teamID,
		"user_id": userID,
		"role":    role,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to add team member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to add team member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

func (r *repo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.RemoveMember")
	defer span.End()

	query := `
		DELETE FROM team_members tm
		USING teams t
		WHERE 
			tm.team_id = t.team_id
			AND tm.team_id = :team_id
			AND tm.user_id = :user_id
			AND t.workspace_id = :workspace_id
	`

	params := map[string]interface{}{
		"team_id":      teamID,
		"user_id":      userID,
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
		errMsg := fmt.Sprintf("failed to remove team member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to remove team member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("member not found")
	}

	return nil
}
