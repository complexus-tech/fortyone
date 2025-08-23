package teamsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
			) as member_count
		FROM
			teams t
		LEFT JOIN user_team_orders uto ON
			t.team_id = uto.team_id
			AND uto.user_id = :user_id
			AND uto.workspace_id = :workspace_id
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
			) as member_count
		FROM
			teams t
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

	params := map[string]any{
		"name":         team.Name,
		"code":         team.Code,
		"color":        team.Color,
		"is_private":   team.IsPrivate,
		"workspace_id": team.Workspace,
	}

	query := `
		INSERT INTO teams (name, code, color, is_private, workspace_id)
		VALUES (:name, :code, :color, :is_private, :workspace_id)
		RETURNING team_id, name, code, color, is_private, workspace_id, created_at, updated_at, 1 as member_count
	`

	defaultStoryAutomationSettingsQuery := `
	INSERT INTO team_story_automation_settings (
		team_id,
		workspace_id,
		auto_close_inactive_enabled,
		auto_close_inactive_months,
		auto_archive_enabled,
		auto_archive_months
	) VALUES (
		:team_id,
		:workspace_id,
		true,
		3,
		true,
		3
	)
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("team code %s already exists", team.Code)
			r.log.Error(ctx, errMsg)
			span.RecordError(teams.ErrTeamCodeExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.CoreTeam{}, teams.ErrTeamCodeExists
		}

		errMsg := fmt.Sprintf("failed to execute query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to execute query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	if err := r.createDefaultStoryStatuses(ctx, tx, dbTeam.ID, dbTeam.Workspace); err != nil {
		errMsg := fmt.Sprintf("failed to create default statuses: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create default statuses"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	defaultStoryAutomationSettingsParams := map[string]any{
		"team_id":      dbTeam.ID,
		"workspace_id": dbTeam.Workspace,
	}
	if _, err := tx.NamedExecContext(ctx, defaultStoryAutomationSettingsQuery, defaultStoryAutomationSettingsParams); err != nil {
		errMsg := fmt.Sprintf("failed to create default story automation settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create default story automation settings"), trace.WithAttributes(attribute.String("error", errMsg)))
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	return toCoreTeam(dbTeam), nil
}

// createDefaultStoryStatuses creates default story statuses for a team using an existing transaction.
func (r *repo) createDefaultStoryStatuses(ctx context.Context, tx *sqlx.Tx, teamID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "teamsrepo.createDefaultStoryStatuses")
	defer span.End()

	// Build values for story statuses batch insert
	storyValues := make([]string, len(teams.DefaultStoryStatuses))
	storyParams := make(map[string]any)
	for i, status := range teams.DefaultStoryStatuses {
		paramPrefix := fmt.Sprintf("s%d_", i)
		storyValues[i] = fmt.Sprintf("(:%sname, :%scategory, :%sorder_index, :%scolor, :team_id, :workspace_id)", paramPrefix, paramPrefix, paramPrefix, paramPrefix)
		storyParams[paramPrefix+"name"] = status.Name
		storyParams[paramPrefix+"category"] = status.Category
		storyParams[paramPrefix+"order_index"] = status.OrderIndex
		storyParams[paramPrefix+"color"] = status.Color
	}
	storyParams["team_id"] = teamID
	storyParams["workspace_id"] = workspaceID

	// Batch insert story statuses
	storyQuery := fmt.Sprintf(`
		INSERT INTO statuses (name, category, order_index, color, team_id, workspace_id)
		VALUES %s
	`, strings.Join(storyValues, ","))

	if _, err := tx.NamedExecContext(ctx, storyQuery, storyParams); err != nil {
		return fmt.Errorf("failed to create story statuses: %w", err)
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
			code = CASE WHEN :code = '' THEN code ELSE :code END,
			color = CASE WHEN :color = '' THEN color ELSE :color END,
			is_private = :is_private,
			updated_at = NOW()
		WHERE 
			team_id = :team_id
			AND workspace_id = :workspace_id
		RETURNING
			team_id,
			name,
			code,
			color,
			is_private,
			workspace_id,
			created_at,
			updated_at
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": updates.Workspace,
		"name":         updates.Name,
		"code":         updates.Code,
		"color":        updates.Color,
		"is_private":   updates.IsPrivate,
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("team code %s already used by another team", updates.Code)
			r.log.Error(ctx, errMsg)
			span.RecordError(teams.ErrTeamCodeExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.CoreTeam{}, teams.ErrTeamCodeExists
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

	params := map[string]any{
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

func (r *repo) AddMember(ctx context.Context, teamID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.AddMember")
	defer span.End()

	query := `
		INSERT INTO team_members (
			team_id,
			user_id,
			created_at,
			updated_at
		)
		VALUES (
			:team_id,
			:user_id,
			NOW(),
			NOW()
		)
	`

	params := map[string]any{
		"team_id": teamID,
		"user_id": userID,
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("user %s is already a member of team %s", userID, teamID)
			r.log.Error(ctx, errMsg)
			span.RecordError(teams.ErrTeamMemberExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.ErrTeamMemberExists
		}
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

	params := map[string]any{
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
		return errors.New("user is not a member of this team")
	}

	return nil
}

// CreateTx creates a new team using an existing transaction.
func (r *repo) CreateTx(ctx context.Context, tx *sqlx.Tx, team teams.CoreTeam) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "teamsrepo.CreateTx")
	defer span.End()

	params := map[string]any{
		"name":         team.Name,
		"code":         team.Code,
		"color":        team.Color,
		"is_private":   team.IsPrivate,
		"workspace_id": team.Workspace,
	}

	query := `
		INSERT INTO teams (name, code, color, is_private, workspace_id)
		VALUES (:name, :code, :color, :is_private, :workspace_id)
		RETURNING team_id, name, code, color, is_private, workspace_id, created_at, updated_at, 1 as member_count
	`

	defaultStoryAutomationSettingsQuery := `
	INSERT INTO team_story_automation_settings (
		team_id,
		workspace_id,
		auto_close_inactive_enabled,
		auto_close_inactive_months,
		auto_archive_enabled,
		auto_archive_months
	) VALUES (
		:team_id,
		:workspace_id,
		true,
		3,
		true,
		3
	)
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("team code %s already exists", team.Code)
			r.log.Error(ctx, errMsg)
			span.RecordError(teams.ErrTeamCodeExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.CoreTeam{}, teams.ErrTeamCodeExists
		}

		errMsg := fmt.Sprintf("failed to execute query: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to execute query"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	defaultStoryAutomationSettingsParams := map[string]any{
		"team_id":      dbTeam.ID,
		"workspace_id": dbTeam.Workspace,
	}

	if _, err := tx.NamedExecContext(ctx, defaultStoryAutomationSettingsQuery, defaultStoryAutomationSettingsParams); err != nil {
		errMsg := fmt.Sprintf("failed to create default story automation settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create default story automation settings"), trace.WithAttributes(attribute.String("error", errMsg)))
	}

	if err := r.createDefaultStoryStatuses(ctx, tx, dbTeam.ID, dbTeam.Workspace); err != nil {
		errMsg := fmt.Sprintf("failed to create default statuses: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create default statuses"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teams.CoreTeam{}, err
	}

	return toCoreTeam(dbTeam), nil
}

// AddMemberTx adds a member to a team using an existing transaction.
func (r *repo) AddMemberTx(ctx context.Context, tx *sqlx.Tx, teamID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.AddMemberTx")
	defer span.End()

	query := `
		INSERT INTO team_members (
			team_id,
			user_id,
			created_at,
			updated_at
		)
		VALUES (
			:team_id,
			:user_id,
			NOW(),
			NOW()
		)
	`

	params := map[string]any{
		"team_id": teamID,
		"user_id": userID,
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
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errMsg := fmt.Sprintf("user %s is already a member of team %s", userID, teamID)
			r.log.Error(ctx, errMsg)
			span.RecordError(teams.ErrTeamMemberExists, trace.WithAttributes(attribute.String("error", errMsg)))
			return teams.ErrTeamMemberExists
		}
		errMsg := fmt.Sprintf("failed to add team member: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to add team member"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// UpdateUserTeamOrdering updates the user's custom team ordering for a workspace.
func (r *repo) UpdateUserTeamOrdering(ctx context.Context, userID, workspaceId uuid.UUID, teamIds []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.UpdateUserTeamOrdering")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to begin transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to begin transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer tx.Rollback()

	// Delete existing ordering
	deleteQuery := `
		DELETE FROM user_team_orders
		WHERE user_id = :user_id AND workspace_id = :workspace_id
	`

	deleteParams := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceId,
	}

	if _, err := tx.NamedExecContext(ctx, deleteQuery, deleteParams); err != nil {
		errMsg := fmt.Sprintf("failed to delete existing ordering: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete existing ordering"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	// Insert new ordering
	insertQuery := `
		INSERT INTO user_team_orders (user_id, team_id, workspace_id, order_index)
		VALUES (:user_id, :team_id, :workspace_id, :order_index)
	`

	for i, teamId := range teamIds {
		insertParams := map[string]any{
			"user_id":      userID,
			"team_id":      teamId,
			"workspace_id": workspaceId,
			"order_index":  i,
		}

		if _, err := tx.NamedExecContext(ctx, insertQuery, insertParams); err != nil {
			errMsg := fmt.Sprintf("failed to insert team ordering: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to insert team ordering"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		errMsg := fmt.Sprintf("failed to commit transaction: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to commit transaction"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	span.AddEvent("user team ordering updated.", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("workspace_id", workspaceId.String()),
		attribute.Int("teams_ordered", len(teamIds)),
	))

	return nil
}
