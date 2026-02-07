package teamsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID) ([]teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.List")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceId,
		"user_id":      userID,
	}
	var teams []dbTeam
	q := `
		SELECT
			t.team_id,
			t.name,
			t.code,
			t.color,
			t.is_private,
			t.workspace_id,
			t.created_at,
			t.updated_at,
			COALESCE(
				(
					SELECT COUNT(*)
					FROM team_members tm
					WHERE tm.team_id = t.team_id
				), 0
			) as member_count,
			COALESCE(tss.auto_create_sprints, false) as sprints_enabled
		FROM
			teams t
		LEFT JOIN user_team_orders uto ON
			t.team_id = uto.team_id
			AND uto.user_id = :user_id
			AND uto.workspace_id = :workspace_id
		LEFT JOIN team_sprint_settings tss ON
			t.team_id = tss.team_id
			AND t.workspace_id = tss.workspace_id
		WHERE
			t.workspace_id = :workspace_id
			AND EXISTS (
				SELECT 1
				FROM team_members tm
				WHERE tm.team_id = t.team_id
				AND tm.user_id = :user_id
			)
		ORDER BY
			CASE
				WHEN uto.order_index IS NOT NULL THEN 0
				ELSE 1
			END,
			uto.order_index ASC NULLS LAST,
			t.created_at DESC;
	`
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching teams with user ordering.")
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

func (r *repo) GetByID(ctx context.Context, teamID uuid.UUID, workspaceID uuid.UUID, userID uuid.UUID) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.GetByID")
	defer span.End()

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
		"user_id":      userID,
	}

	var team dbTeam
	q := `
		SELECT
			t.team_id,
			t.name,
			t.code,
			t.color,
			t.is_private,
			t.workspace_id,
			t.created_at,
			t.updated_at,
			COALESCE(
				(
					SELECT COUNT(*)
					FROM team_members tm
					WHERE tm.team_id = t.team_id
				), 0
			) as member_count,
			COALESCE(tss.auto_create_sprints, false) as sprints_enabled
		FROM
			teams t
		LEFT JOIN team_sprint_settings tss ON
			t.team_id = tss.team_id
			AND t.workspace_id = tss.workspace_id
		WHERE
			t.team_id = :team_id
			AND t.workspace_id = :workspace_id
			AND EXISTS (
				SELECT 1
				FROM team_members tm
				WHERE tm.team_id = t.team_id
				AND tm.user_id = :user_id
			)
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching team by ID with membership check.")
	if err := stmt.GetContext(ctx, &team, params); err != nil {
		if err == sql.ErrNoRows {
			errMsg := "team not found or user is not a member"
			r.log.Info(ctx, errMsg, "team_id", teamID, "user_id", userID)
			span.RecordError(errors.New("team not found"), trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.CoreTeam{}, errors.New("team not found")
		}
		errMsg := fmt.Sprintf("Failed to retrieve team from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("team not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	r.log.Info(ctx, "Team retrieved successfully.")
	span.AddEvent("team retrieved.", trace.WithAttributes(
		attribute.String("team_id", team.ID.String()),
		attribute.String("query", q),
	))

	return toCoreTeam(team), nil
}

func (r *repo) ListPublicTeams(ctx context.Context, workspaceId uuid.UUID, userID uuid.UUID) ([]teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.ListPublicTeams")
	defer span.End()

	params := map[string]any{
		"workspace_id": workspaceId,
		"user_id":      userID,
	}

	var teams []dbTeam
	q := `
		SELECT
			t.team_id,
			t.name,
			t.code,
			t.color,
			t.is_private,
			t.workspace_id,
			t.created_at,
			t.updated_at,
			COALESCE(
				(
					SELECT COUNT(*)
					FROM team_members tm
					WHERE tm.team_id = t.team_id
				), 0
			) as member_count,
			COALESCE(tss.auto_create_sprints, false) as sprints_enabled
		FROM
			teams t
		LEFT JOIN team_sprint_settings tss ON
			t.team_id = tss.team_id
			AND t.workspace_id = tss.workspace_id
		WHERE
			t.workspace_id = :workspace_id
			AND t.is_private = false
			AND NOT EXISTS (
				SELECT 1
				FROM team_members tm
				WHERE tm.team_id = t.team_id
				AND tm.user_id = :user_id
			)
		ORDER BY t.created_at DESC;
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching public teams.")
	if err := stmt.SelectContext(ctx, &teams, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve public teams from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("teams not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Public teams retrieved successfully.")
	span.AddEvent("public teams retrieved.", trace.WithAttributes(
		attribute.Int("teams.count", len(teams)),
		attribute.String("query", q),
	))

	return toCoreTeams(teams), nil
}
