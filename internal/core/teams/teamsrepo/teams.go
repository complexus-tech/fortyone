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
	ctx, span := web.AddSpan(ctx, "business.repository.teams.Create")
	defer span.End()

	var result dbTeam
	query := `
		INSERT INTO teams (
			name,
			description,
			code,
			color,
			icon,
			workspace_id,
			created_at,
			updated_at
		)
		VALUES (
			:name,
			:description,
			:code,
			:color,
			:icon,
			:workspace_id,
			NOW(),
			NOW()
		)
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
		"name":         team.Name,
		"description":  team.Description,
		"code":         team.Code,
		"color":        team.Color,
		"icon":         team.Icon,
		"workspace_id": team.Workspace,
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
		errMsg := fmt.Sprintf("failed to create team: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create team"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	return toCoreTeam(result), nil
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
