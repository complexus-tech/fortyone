package emailreply

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrActionUnauthorized = errors.New("email action is no longer authorized")
	ErrActionConflict     = errors.New("email action target changed after its preview")
)

const maximumAgentChoices = 300

// DBContextLoader rebuilds an email conversation's model-visible scope from
// current membership and entity rows. It never trusts IDs or versions supplied
// by email text or a model.
type DBContextLoader struct {
	db *sqlx.DB
}

func NewDBContextLoader(db *sqlx.DB) (*DBContextLoader, error) {
	if db == nil {
		return nil, errors.New("email reply context database is required")
	}
	return &DBContextLoader{db: db}, nil
}

func (loader *DBContextLoader) Load(ctx context.Context, thread messaging.EmailThreadRecord) (AuthorizedContext, error) {
	if loader == nil || loader.db == nil {
		return AuthorizedContext{}, errors.New("email reply context loader is not configured")
	}
	var threadContext emailthread.ThreadContext
	if err := jsonUnmarshalObject(thread.Context, &threadContext); err != nil {
		return AuthorizedContext{}, fmt.Errorf("decode email thread context: %w", err)
	}
	if threadContext.Version != 1 || strings.TrimSpace(threadContext.Source) == "" || strings.TrimSpace(threadContext.WorkspaceSlug) == "" {
		return AuthorizedContext{}, messaging.ErrInvalidEmailConversation
	}
	allowedTeams, workspaceSlug, role, err := loader.currentTeamScope(ctx, thread.WorkspaceID, thread.UserID)
	if err != nil {
		return AuthorizedContext{}, err
	}
	if workspaceSlug != threadContext.WorkspaceSlug {
		return AuthorizedContext{}, messaging.ErrInvalidEmailConversation
	}
	allowed := make(map[uuid.UUID]struct{}, len(allowedTeams))
	for _, teamID := range allowedTeams {
		allowed[teamID] = struct{}{}
	}

	targets := make([]emailagent.AuthorizedTarget, 0, len(threadContext.Targets))
	facts := make([]emailagent.GroundedFact, 0)
	storyTeams := make(map[uuid.UUID]struct{})
	seenTargets := make(map[string]struct{}, len(threadContext.Targets))
	for _, targetContext := range threadContext.Targets {
		target, found, err := loader.loadTarget(ctx, thread.WorkspaceID, targetContext)
		if err != nil {
			return AuthorizedContext{}, err
		}
		if !found {
			return AuthorizedContext{}, ErrActionUnauthorized
		}
		if _, ok := allowed[target.TeamID]; !ok {
			return AuthorizedContext{}, ErrActionUnauthorized
		}
		if role == "guest" && target.Kind != emailagent.TargetStory {
			facts = append(facts, emailagent.GroundedFact{
				Reference:       "read_" + targetReference(target.Kind, target.ID),
				Text:            target.DisplayName + ". " + target.CurrentState,
				ProtectedTokens: []string{target.DisplayName},
			})
			continue
		}
		key := string(target.Kind) + ":" + target.ID.String()
		if _, duplicate := seenTargets[key]; duplicate {
			continue
		}
		seenTargets[key] = struct{}{}
		target.Reference = targetReference(target.Kind, target.ID)
		targets = append(targets, target)
		if target.Kind == emailagent.TargetStory {
			storyTeams[target.TeamID] = struct{}{}
		}
	}
	sort.Slice(targets, func(left, right int) bool {
		if targets[left].Kind == targets[right].Kind {
			return targets[left].Reference < targets[right].Reference
		}
		return targets[left].Kind < targets[right].Kind
	})
	choices, err := loader.storyChoices(ctx, thread.WorkspaceID, storyTeams)
	if err != nil {
		return AuthorizedContext{}, err
	}
	return AuthorizedContext{
		AllowedTeamIDs: allowedTeams,
		Facts:          facts,
		Targets:        targets,
		Choices:        choices,
	}, nil
}

func (loader *DBContextLoader) AuthorizeProposal(ctx context.Context, proposal emailagent.ActionProposal) error {
	if proposal.WorkspaceID == uuid.Nil || proposal.ActorID == uuid.Nil {
		return ErrActionUnauthorized
	}
	allowedTeams, _, role, err := loader.currentTeamScope(ctx, proposal.WorkspaceID, proposal.ActorID)
	if err != nil {
		return err
	}
	if role == "guest" && proposal.Kind != emailagent.ActionStoryUpdate {
		return ErrActionUnauthorized
	}
	allowed := make(map[uuid.UUID]struct{}, len(allowedTeams))
	for _, teamID := range allowedTeams {
		allowed[teamID] = struct{}{}
	}
	target, err := proposalTarget(proposal)
	if err != nil {
		return err
	}
	actualTeamID, found, err := loader.currentTargetTeam(ctx, proposal.WorkspaceID, proposal.Kind, target.ID)
	if err != nil {
		return err
	}
	if !found {
		return ErrActionUnauthorized
	}
	if actualTeamID != target.TeamID {
		return ErrActionConflict
	}
	if _, ok := allowed[actualTeamID]; !ok {
		return ErrActionUnauthorized
	}
	if proposal.Story != nil {
		if proposal.Story.Status != nil {
			var exists bool
			if err := loader.db.GetContext(ctx, &exists, `
				SELECT EXISTS (
					SELECT 1 FROM statuses
					WHERE workspace_id = $1 AND team_id = $2 AND status_id = $3
				)
			`, proposal.WorkspaceID, actualTeamID, proposal.Story.Status.StatusID); err != nil {
				return fmt.Errorf("reauthorize email story status: %w", err)
			}
			if !exists {
				return ErrActionConflict
			}
		}
		if proposal.Story.Assignee != nil && proposal.Story.Assignee.Operation == emailagent.AssigneeAssign {
			if proposal.Story.Assignee.AssigneeID == nil {
				return ErrActionConflict
			}
			var exists bool
			if err := loader.db.GetContext(ctx, &exists, `
				SELECT EXISTS (
					SELECT 1
					FROM team_members membership
					INNER JOIN users actor ON actor.user_id = membership.user_id
					WHERE membership.team_id = $1 AND membership.user_id = $2
					  AND actor.is_active = true AND actor.is_system = false
				)
			`, actualTeamID, *proposal.Story.Assignee.AssigneeID); err != nil {
				return fmt.Errorf("reauthorize email story assignee: %w", err)
			}
			if !exists {
				return ErrActionConflict
			}
		}
	}
	return nil
}

// CurrentVersion returns the target's current database version after verifying
// that it still belongs to the proposal workspace and has not been deleted.
func (loader *DBContextLoader) CurrentVersion(ctx context.Context, proposal emailagent.ActionProposal) (time.Time, error) {
	if loader == nil || loader.db == nil {
		return time.Time{}, errors.New("email reply context loader is not configured")
	}
	target, err := proposalTarget(proposal)
	if err != nil {
		return time.Time{}, err
	}
	query := ""
	switch proposal.Kind {
	case emailagent.ActionObjectiveUpdate:
		query = `SELECT updated_at FROM objectives WHERE workspace_id = $1 AND objective_id = $2`
	case emailagent.ActionKeyResultUpdate:
		query = `
			SELECT result.updated_at FROM key_results result
			INNER JOIN objectives objective ON objective.objective_id = result.objective_id
			WHERE objective.workspace_id = $1 AND result.id = $2`
	case emailagent.ActionStoryUpdate:
		query = `SELECT updated_at FROM stories WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL`
	case emailagent.ActionFeedbackStatus:
		query = `SELECT updated_at FROM feedback_items WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL`
	default:
		return time.Time{}, errors.New("unsupported email action kind")
	}
	var version time.Time
	if err := loader.db.GetContext(ctx, &version, query, proposal.WorkspaceID, target.ID); errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrActionUnauthorized
	} else if err != nil {
		return time.Time{}, fmt.Errorf("read current email action version: %w", err)
	}
	return version.UTC(), nil
}

// ProposalAlreadyApplied compares only the explicitly proposed fields against
// current database state. It closes the crash window between a committed
// domain CAS write and the durable proposal receipt without treating unrelated
// later edits as proof that Maya applied the proposal.
func (loader *DBContextLoader) ProposalAlreadyApplied(ctx context.Context, proposal emailagent.ActionProposal) (bool, error) {
	if loader == nil || loader.db == nil {
		return false, errors.New("email reply context loader is not configured")
	}
	switch proposal.Kind {
	case emailagent.ActionObjectiveUpdate:
		if proposal.Objective == nil || proposal.Objective.Health == nil {
			return false, errors.New("objective proposal is incomplete")
		}
		var health sql.NullString
		err := loader.db.GetContext(ctx, &health, `
			SELECT health FROM objectives WHERE workspace_id = $1 AND objective_id = $2
		`, proposal.WorkspaceID, proposal.Objective.Target.ID)
		return health.Valid && health.String == string(*proposal.Objective.Health), normalizeReconciliationRead(err)
	case emailagent.ActionKeyResultUpdate:
		if proposal.KeyResult == nil || proposal.KeyResult.CurrentValue == nil {
			return false, errors.New("key result proposal is incomplete")
		}
		var value float64
		err := loader.db.GetContext(ctx, &value, `
			SELECT result.current_value FROM key_results result
			INNER JOIN objectives objective ON objective.objective_id = result.objective_id
			WHERE objective.workspace_id = $1 AND result.id = $2
		`, proposal.WorkspaceID, proposal.KeyResult.Target.ID)
		return value == *proposal.KeyResult.CurrentValue, normalizeReconciliationRead(err)
	case emailagent.ActionFeedbackStatus:
		if proposal.Feedback == nil {
			return false, errors.New("feedback proposal is incomplete")
		}
		var status string
		err := loader.db.GetContext(ctx, &status, `
			SELECT status FROM feedback_items WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL
		`, proposal.WorkspaceID, proposal.Feedback.Target.ID)
		return status == string(proposal.Feedback.Status), normalizeReconciliationRead(err)
	case emailagent.ActionStoryUpdate:
		if proposal.Story == nil {
			return false, errors.New("story proposal is incomplete")
		}
		return loader.storyProposalAlreadyApplied(ctx, proposal.WorkspaceID, *proposal.Story)
	default:
		return false, errors.New("unsupported email action kind")
	}
}

func (loader *DBContextLoader) storyProposalAlreadyApplied(ctx context.Context, workspaceID uuid.UUID, action emailagent.StoryAction) (bool, error) {
	type currentStory struct {
		StatusID   *uuid.UUID `db:"status_id"`
		AssigneeID *uuid.UUID `db:"assignee_id"`
		EndDate    *time.Time `db:"end_date"`
	}
	var current currentStory
	err := loader.db.GetContext(ctx, &current, `
		SELECT status_id, assignee_id, end_date
		FROM stories WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL
	`, workspaceID, action.Target.ID)
	if err := normalizeReconciliationRead(err); err != nil {
		return false, err
	}
	if action.Status != nil && (current.StatusID == nil || *current.StatusID != action.Status.StatusID) {
		return false, nil
	}
	if action.Assignee != nil {
		switch action.Assignee.Operation {
		case emailagent.AssigneeUnassign:
			if current.AssigneeID != nil {
				return false, nil
			}
		case emailagent.AssigneeAssign:
			if action.Assignee.AssigneeID == nil || current.AssigneeID == nil || *current.AssigneeID != *action.Assignee.AssigneeID {
				return false, nil
			}
		}
	}
	if action.DueDate != nil {
		switch action.DueDate.Operation {
		case emailagent.DateClear:
			if current.EndDate != nil {
				return false, nil
			}
		case emailagent.DateSet:
			expected, parseErr := time.Parse("2006-01-02", action.DueDate.Date)
			if parseErr != nil {
				return false, parseErr
			}
			if current.EndDate == nil || !sameCalendarDate(current.EndDate.UTC(), expected.UTC()) {
				return false, nil
			}
		}
	}
	return true, nil
}

func normalizeReconciliationRead(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrActionUnauthorized
	}
	return err
}

func sameCalendarDate(left, right time.Time) bool {
	leftYear, leftMonth, leftDay := left.Date()
	rightYear, rightMonth, rightDay := right.Date()
	return leftYear == rightYear && leftMonth == rightMonth && leftDay == rightDay
}

func (loader *DBContextLoader) currentTeamScope(ctx context.Context, workspaceID, userID uuid.UUID) ([]uuid.UUID, string, string, error) {
	type actorWorkspace struct {
		Slug string `db:"slug"`
		Role string `db:"role"`
	}
	var actor actorWorkspace
	err := loader.db.GetContext(ctx, &actor, `
		SELECT workspace.slug, member.role
		FROM workspaces workspace
		INNER JOIN workspace_members member
			ON member.workspace_id = workspace.workspace_id AND member.user_id = $2
		INNER JOIN users actor ON actor.user_id = member.user_id
		WHERE workspace.workspace_id = $1
		  AND workspace.deleted_at IS NULL
		  AND actor.is_active = true
		  AND actor.is_system = false
		  AND member.role IN ('admin', 'member', 'guest')
	`, workspaceID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", "", ErrActionUnauthorized
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("authorize email conversation actor: %w", err)
	}
	teamIDs := make([]uuid.UUID, 0)
	err = loader.db.SelectContext(ctx, &teamIDs, `
		SELECT team.team_id
		FROM teams team
		WHERE team.workspace_id = $1
		  AND ($3 = 'admin' OR EXISTS (
			SELECT 1 FROM team_members membership
			WHERE membership.team_id = team.team_id AND membership.user_id = $2
		  ))
		ORDER BY team.team_id
	`, workspaceID, userID, actor.Role)
	if err != nil {
		return nil, "", "", fmt.Errorf("load email conversation team scope: %w", err)
	}
	return teamIDs, actor.Slug, actor.Role, nil
}

func (loader *DBContextLoader) loadTarget(
	ctx context.Context,
	workspaceID uuid.UUID,
	input emailthread.TargetContext,
) (emailagent.AuthorizedTarget, bool, error) {
	if input.ID == uuid.Nil {
		return emailagent.AuthorizedTarget{}, false, nil
	}
	switch strings.TrimSpace(input.Kind) {
	case string(emailagent.TargetObjective):
		type row struct {
			ID        uuid.UUID  `db:"id"`
			TeamID    uuid.UUID  `db:"team_id"`
			Name      string     `db:"name"`
			Health    *string    `db:"health"`
			StartDate *time.Time `db:"start_date"`
			EndDate   *time.Time `db:"end_date"`
			UpdatedAt time.Time  `db:"updated_at"`
		}
		var value row
		err := loader.db.GetContext(ctx, &value, `
			SELECT objective_id AS id, team_id, name, CAST(health AS text) AS health,
			       start_date, end_date, updated_at
			FROM objectives
			WHERE workspace_id = $1 AND objective_id = $2
		`, workspaceID, input.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return emailagent.AuthorizedTarget{}, false, nil
		}
		if err != nil {
			return emailagent.AuthorizedTarget{}, false, fmt.Errorf("load email objective target: %w", err)
		}
		health := "not set"
		if value.Health != nil && strings.TrimSpace(*value.Health) != "" {
			health = strings.TrimSpace(*value.Health)
		}
		state := "Health: " + health + dateState(value.StartDate, value.EndDate)
		return authorizedTarget(emailagent.TargetObjective, value.ID, value.TeamID, value.Name, state, value.UpdatedAt), true, nil

	case string(emailagent.TargetKeyResult):
		type row struct {
			ID              uuid.UUID  `db:"id"`
			TeamID          uuid.UUID  `db:"team_id"`
			Name            string     `db:"name"`
			MeasurementType string     `db:"measurement_type"`
			CurrentValue    float64    `db:"current_value"`
			TargetValue     float64    `db:"target_value"`
			StartDate       *time.Time `db:"start_date"`
			EndDate         *time.Time `db:"end_date"`
			UpdatedAt       time.Time  `db:"updated_at"`
		}
		var value row
		err := loader.db.GetContext(ctx, &value, `
			SELECT result.id, objective.team_id, result.name, result.measurement_type,
			       result.current_value, result.target_value, result.start_date, result.end_date, result.updated_at
			FROM key_results result
			INNER JOIN objectives objective ON objective.objective_id = result.objective_id
			WHERE objective.workspace_id = $1 AND result.id = $2
		`, workspaceID, input.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return emailagent.AuthorizedTarget{}, false, nil
		}
		if err != nil {
			return emailagent.AuthorizedTarget{}, false, fmt.Errorf("load email key result target: %w", err)
		}
		state := fmt.Sprintf("Current value: %s; target value: %s; measurement: %s%s",
			formatFloat(value.CurrentValue), formatFloat(value.TargetValue), value.MeasurementType, dateState(value.StartDate, value.EndDate))
		return authorizedTarget(emailagent.TargetKeyResult, value.ID, value.TeamID, value.Name, state, value.UpdatedAt), true, nil

	case string(emailagent.TargetStory):
		type row struct {
			ID           uuid.UUID  `db:"id"`
			TeamID       uuid.UUID  `db:"team_id"`
			Title        string     `db:"title"`
			StatusName   *string    `db:"status_name"`
			AssigneeName *string    `db:"assignee_name"`
			EndDate      *time.Time `db:"end_date"`
			UpdatedAt    time.Time  `db:"updated_at"`
		}
		var value row
		err := loader.db.GetContext(ctx, &value, `
			SELECT story.id, story.team_id, story.title, state.name AS status_name,
			       COALESCE(actor.full_name, actor.email) AS assignee_name,
			       story.end_date, story.updated_at
			FROM stories story
			LEFT JOIN statuses state ON state.status_id = story.status_id
			LEFT JOIN users actor ON actor.user_id = story.assignee_id
			WHERE story.workspace_id = $1 AND story.id = $2 AND story.deleted_at IS NULL
		`, workspaceID, input.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return emailagent.AuthorizedTarget{}, false, nil
		}
		if err != nil {
			return emailagent.AuthorizedTarget{}, false, fmt.Errorf("load email story target: %w", err)
		}
		status, assignee, due := "not set", "unassigned", "not set"
		if value.StatusName != nil && strings.TrimSpace(*value.StatusName) != "" {
			status = strings.TrimSpace(*value.StatusName)
		}
		if value.AssigneeName != nil && strings.TrimSpace(*value.AssigneeName) != "" {
			assignee = strings.TrimSpace(*value.AssigneeName)
		}
		if value.EndDate != nil {
			due = value.EndDate.UTC().Format("2006-01-02")
		}
		state := fmt.Sprintf("Status: %s; assignee: %s; due date: %s", status, assignee, due)
		return authorizedTarget(emailagent.TargetStory, value.ID, value.TeamID, value.Title, state, value.UpdatedAt), true, nil

	case string(emailagent.TargetFeedback):
		type row struct {
			ID        uuid.UUID `db:"id"`
			TeamID    uuid.UUID `db:"team_id"`
			Title     string    `db:"title"`
			Status    string    `db:"status"`
			UpdatedAt time.Time `db:"updated_at"`
		}
		var value row
		err := loader.db.GetContext(ctx, &value, `
			SELECT item.id, board.team_id, item.title, item.status, item.updated_at
			FROM feedback_items item
			INNER JOIN feedback_boards board ON board.id = item.board_id AND board.workspace_id = item.workspace_id
			WHERE item.workspace_id = $1 AND item.id = $2 AND item.deleted_at IS NULL
		`, workspaceID, input.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return emailagent.AuthorizedTarget{}, false, nil
		}
		if err != nil {
			return emailagent.AuthorizedTarget{}, false, fmt.Errorf("load email feedback target: %w", err)
		}
		return authorizedTarget(emailagent.TargetFeedback, value.ID, value.TeamID, value.Title, "Status: "+value.Status, value.UpdatedAt), true, nil
	default:
		return emailagent.AuthorizedTarget{}, false, nil
	}
}

func (loader *DBContextLoader) storyChoices(ctx context.Context, workspaceID uuid.UUID, teams map[uuid.UUID]struct{}) ([]emailagent.AuthorizedChoice, error) {
	teamIDs := make([]uuid.UUID, 0, len(teams))
	for teamID := range teams {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(left, right int) bool { return teamIDs[left].String() < teamIDs[right].String() })
	choices := make([]emailagent.AuthorizedChoice, 0)
	for _, teamID := range teamIDs {
		type choiceRow struct {
			ID   uuid.UUID `db:"id"`
			Name string    `db:"name"`
		}
		statuses := make([]choiceRow, 0)
		if err := loader.db.SelectContext(ctx, &statuses, `
			SELECT status_id AS id, name
			FROM statuses
			WHERE workspace_id = $1 AND team_id = $2
			ORDER BY order_index, status_id
		`, workspaceID, teamID); err != nil {
			return nil, fmt.Errorf("load email story status choices: %w", err)
		}
		for _, status := range statuses {
			if len(choices) == maximumAgentChoices {
				return choices, nil
			}
			choices = append(choices, emailagent.AuthorizedChoice{
				Reference: choiceReference("status", status.ID), Kind: emailagent.ChoiceStoryStatus,
				DisplayName: status.Name, ID: status.ID, TeamID: teamID,
			})
		}
		members := make([]choiceRow, 0)
		if err := loader.db.SelectContext(ctx, &members, `
			SELECT actor.user_id AS id, COALESCE(NULLIF(actor.full_name, ''), actor.email) AS name
			FROM team_members member
			INNER JOIN users actor ON actor.user_id = member.user_id
			WHERE member.team_id = $1 AND actor.is_active = true AND actor.is_system = false
			ORDER BY name, actor.user_id
			LIMIT $2
		`, teamID, maximumAgentChoices-len(choices)); err != nil {
			return nil, fmt.Errorf("load email story assignee choices: %w", err)
		}
		for _, member := range members {
			choices = append(choices, emailagent.AuthorizedChoice{
				Reference: choiceReference("assignee", member.ID), Kind: emailagent.ChoiceStoryAssignee,
				DisplayName: member.Name, ID: member.ID, TeamID: teamID,
			})
		}
		if len(choices) == maximumAgentChoices {
			return choices, nil
		}
	}
	return choices, nil
}

func (loader *DBContextLoader) currentTargetTeam(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind emailagent.ActionKind,
	entityID uuid.UUID,
) (uuid.UUID, bool, error) {
	query := ""
	switch kind {
	case emailagent.ActionObjectiveUpdate:
		query = `SELECT team_id FROM objectives WHERE workspace_id = $1 AND objective_id = $2`
	case emailagent.ActionKeyResultUpdate:
		query = `
			SELECT objective.team_id FROM key_results result
			INNER JOIN objectives objective ON objective.objective_id = result.objective_id
			WHERE objective.workspace_id = $1 AND result.id = $2`
	case emailagent.ActionStoryUpdate:
		query = `SELECT team_id FROM stories WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL`
	case emailagent.ActionFeedbackStatus:
		query = `
			SELECT board.team_id FROM feedback_items item
			INNER JOIN feedback_boards board ON board.id = item.board_id AND board.workspace_id = item.workspace_id
			WHERE item.workspace_id = $1 AND item.id = $2 AND item.deleted_at IS NULL`
	default:
		return uuid.Nil, false, errors.New("unsupported email action kind")
	}
	var teamID uuid.UUID
	err := loader.db.GetContext(ctx, &teamID, query, workspaceID, entityID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("reauthorize email action target: %w", err)
	}
	return teamID, true, nil
}

func authorizedTarget(kind emailagent.TargetKind, id, teamID uuid.UUID, name, state string, updatedAt time.Time) emailagent.AuthorizedTarget {
	return emailagent.AuthorizedTarget{
		Kind: kind, ID: id, TeamID: teamID, DisplayName: strings.TrimSpace(name),
		CurrentState: strings.TrimSpace(state), ExpectedUpdatedAt: updatedAt.UTC(),
	}
}

func targetReference(kind emailagent.TargetKind, id uuid.UUID) string {
	prefix := strings.ReplaceAll(string(kind), "_", "")
	return prefix + "_" + strings.ReplaceAll(id.String(), "-", "")
}

func choiceReference(kind string, id uuid.UUID) string {
	return kind + "_" + strings.ReplaceAll(id.String(), "-", "")
}

func dateState(start, end *time.Time) string {
	parts := make([]string, 0, 2)
	if start != nil {
		parts = append(parts, "start date: "+start.UTC().Format("2006-01-02"))
	}
	if end != nil {
		parts = append(parts, "end date: "+end.UTC().Format("2006-01-02"))
	}
	if len(parts) == 0 {
		return ""
	}
	return "; " + strings.Join(parts, "; ")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func jsonUnmarshalObject(raw []byte, destination any) error {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return errors.New("JSON object is empty")
	}
	return json.Unmarshal(raw, destination)
}
