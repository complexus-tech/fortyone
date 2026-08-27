package teamsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) Create(ctx context.Context, team teams.CoreTeam) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "teamsrepository.Create")
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
	ctx, span := web.AddSpan(ctx, "teamsrepository.createDefaultStoryStatuses")
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
			return teams.CoreTeam{}, teams.ErrTeamNotFound
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
		return teams.ErrTeamNotFound
	}

	return nil
}

const addMemberQuery = `
	WITH eligible_member AS (
		SELECT team.team_id, membership.user_id
		FROM teams team
		INNER JOIN workspace_members membership
			ON membership.workspace_id = team.workspace_id
			AND membership.user_id = :user_id
		INNER JOIN users member
			ON member.user_id = membership.user_id
			AND member.is_active = TRUE
		WHERE team.team_id = :team_id
			AND team.workspace_id = :workspace_id
	), inserted AS (
		INSERT INTO team_members (team_id, user_id, created_at, updated_at)
		SELECT team_id, user_id, NOW(), NOW()
		FROM eligible_member
		ON CONFLICT (team_id, user_id) DO NOTHING
		RETURNING team_id
	)
	SELECT
		EXISTS (SELECT 1 FROM eligible_member) AS eligible,
		EXISTS (SELECT 1 FROM inserted) AS added
`

func (r *repo) AddMember(ctx context.Context, teamID, userID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.AddMember")
	defer span.End()

	params := map[string]any{
		"team_id":      teamID,
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, addMemberQuery)
	if err != nil {
		return fmt.Errorf("preparing scoped team member add: %w", err)
	}
	defer stmt.Close()

	var outcome addMemberOutcome
	if err := stmt.GetContext(ctx, &outcome, params); err != nil {
		return fmt.Errorf("adding scoped team member: %w", err)
	}
	return validateAddMember(outcome)
}

type addMemberOutcome struct {
	Eligible bool `db:"eligible"`
	Added    bool `db:"added"`
}

func validateAddMember(outcome addMemberOutcome) error {
	if !outcome.Eligible {
		return teams.ErrTeamNotFound
	}
	if !outcome.Added {
		return teams.ErrTeamMemberExists
	}
	return nil
}

const joinPublicTeamQuery = `
	WITH eligible_team AS (
		SELECT team.team_id, membership.user_id
		FROM teams team
		INNER JOIN workspace_members membership
			ON membership.workspace_id = team.workspace_id
			AND membership.user_id = :actor_id
		INNER JOIN users actor
			ON actor.user_id = membership.user_id
			AND actor.is_active = TRUE
		WHERE team.team_id = :team_id
			AND team.workspace_id = :workspace_id
			AND team.is_private = FALSE
	), inserted AS (
		INSERT INTO team_members (team_id, user_id, created_at, updated_at)
		SELECT team_id, user_id, NOW(), NOW()
		FROM eligible_team
		ON CONFLICT (team_id, user_id) DO NOTHING
		RETURNING team_id
	)
	SELECT
		EXISTS (SELECT 1 FROM eligible_team) AS eligible,
		EXISTS (SELECT 1 FROM inserted) AS joined
`

// JoinPublicTeam atomically authorizes and adds the authenticated actor.
func (r *repo) JoinPublicTeam(ctx context.Context, input teams.CorePublicTeamJoin) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.JoinPublicTeam")
	defer span.End()

	params := map[string]any{
		"team_id":      input.TeamID,
		"actor_id":     input.ActorID,
		"workspace_id": input.WorkspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, joinPublicTeamQuery)
	if err != nil {
		return fmt.Errorf("preparing public team join: %w", err)
	}
	defer stmt.Close()

	var outcome publicTeamJoinOutcome
	if err := stmt.GetContext(ctx, &outcome, params); err != nil {
		return fmt.Errorf("joining public team: %w", err)
	}
	return validatePublicTeamJoin(outcome)
}

type publicTeamJoinOutcome struct {
	Eligible bool `db:"eligible"`
	Joined   bool `db:"joined"`
}

func validatePublicTeamJoin(outcome publicTeamJoinOutcome) error {
	if !outcome.Eligible {
		return teams.ErrTeamNotFound
	}
	if !outcome.Joined {
		return teams.ErrTeamMemberExists
	}
	return nil
}

const removeMemberQuery = `
	DELETE FROM team_members tm
	USING teams t
	WHERE
		tm.team_id = t.team_id
		AND tm.team_id = :team_id
		AND tm.user_id = :user_id
		AND t.workspace_id = :workspace_id
`

func (r *repo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.RemoveMember")
	defer span.End()

	params := map[string]any{
		"team_id":      teamID,
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, removeMemberQuery)
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
		return teams.ErrTeamMemberNotFound
	}

	return nil
}

const leaveTeamQuery = `
	DELETE FROM team_members membership
	USING teams team, workspace_members workspace_membership, users actor
	WHERE membership.team_id = team.team_id
		AND membership.team_id = :team_id
		AND membership.user_id = :actor_id
		AND team.workspace_id = :workspace_id
		AND workspace_membership.workspace_id = team.workspace_id
		AND workspace_membership.user_id = :actor_id
		AND actor.user_id = workspace_membership.user_id
		AND actor.is_active = TRUE
`

// LeaveTeam deletes only the authenticated actor's membership in the scoped workspace.
func (r *repo) LeaveTeam(ctx context.Context, input teams.CoreTeamSelfLeave) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.LeaveTeam")
	defer span.End()

	params := map[string]any{
		"team_id":      input.TeamID,
		"actor_id":     input.ActorID,
		"workspace_id": input.WorkspaceID,
	}

	result, err := r.db.NamedExecContext(ctx, leaveTeamQuery, params)
	if err != nil {
		return fmt.Errorf("leaving scoped team: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reading team leave result: %w", err)
	}
	if rowsAffected == 0 {
		return teams.ErrTeamMemberNotFound
	}
	return nil
}

func (r *repo) UpdateMemberAIContext(ctx context.Context, teamID, userID uuid.UUID, workspaceID uuid.UUID, input teams.CoreTeamMemberAIContext) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teams.UpdateMemberAIContext")
	defer span.End()

	query := `
		UPDATE team_members tm
		SET
			ai_role_title = :ai_role_title,
			ai_role_description = :ai_role_description,
			updated_at = NOW()
		FROM teams t
		WHERE tm.team_id = t.team_id
			AND tm.team_id = :team_id
			AND tm.user_id = :user_id
			AND t.workspace_id = :workspace_id
	`

	params := map[string]any{
		"team_id":             teamID,
		"user_id":             userID,
		"workspace_id":        workspaceID,
		"ai_role_title":       strings.TrimSpace(input.RoleTitle),
		"ai_role_description": strings.TrimSpace(input.RoleDescription),
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
		errMsg := fmt.Sprintf("failed to update team member ai context: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update team member ai context"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("team member not found")
	}
	return nil
}

// CreateTx creates a new team using an existing transaction.
func (r *repo) CreateTx(ctx context.Context, tx *sqlx.Tx, team teams.CoreTeam) (teams.CoreTeam, error) {
	ctx, span := web.AddSpan(ctx, "teamsrepository.CreateTx")
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
