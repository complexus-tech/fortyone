package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	emailreplydomain "github.com/complexus-tech/projects-api/internal/modules/emailreply/domain"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/google/uuid"
)

var (
	ErrActionUnauthorized = errors.New("email action is no longer authorized")
	ErrActionConflict     = errors.New("email action target changed after its preview")
)

const (
	maximumAgentChoices       = 300
	maximumAuthorizedTeams    = 500
	maximumAuthorizedTargets  = 100
	maximumThreadContextBytes = 256 << 10
)

// DBContextLoader rebuilds an email conversation's model-visible scope from
// current membership and entity rows. It never trusts IDs or versions supplied
// by email text or a model.
type DBContextLoader struct {
	store ContextStore
}

func NewDBContextLoader(store ContextStore) (*DBContextLoader, error) {
	if store == nil {
		return nil, errors.New("email reply context store is required")
	}
	return &DBContextLoader{store: store}, nil
}

func (loader *DBContextLoader) Load(ctx context.Context, thread Thread) (AuthorizedContext, error) {
	if loader == nil || loader.store == nil {
		return AuthorizedContext{}, errors.New("email reply context loader is not configured")
	}
	if thread.ID == uuid.Nil || thread.WorkspaceID == uuid.Nil || thread.UserID == uuid.Nil ||
		len(thread.Context) == 0 || len(thread.Context) > maximumThreadContextBytes {
		return AuthorizedContext{}, ErrInvalidConversation
	}

	var threadContext emailthread.ThreadContext
	if err := jsonUnmarshalObject(thread.Context, &threadContext); err != nil {
		return AuthorizedContext{}, fmt.Errorf("decode email thread context: %w", err)
	}
	if threadContext.Version != 1 || strings.TrimSpace(threadContext.Source) == "" ||
		strings.TrimSpace(threadContext.WorkspaceSlug) == "" || len(threadContext.Targets) > maximumAuthorizedTargets {
		return AuthorizedContext{}, ErrInvalidConversation
	}

	allowedTeams, workspaceSlug, role, err := loader.currentTeamScope(ctx, thread.WorkspaceID, thread.UserID)
	if err != nil {
		return AuthorizedContext{}, err
	}
	if err := validateAuthorizedTeamIDs(allowedTeams); err != nil {
		return AuthorizedContext{}, err
	}
	if workspaceSlug != threadContext.WorkspaceSlug {
		return AuthorizedContext{}, ErrInvalidConversation
	}
	allowed := make(map[uuid.UUID]struct{}, len(allowedTeams))
	for _, teamID := range allowedTeams {
		allowed[teamID] = struct{}{}
	}

	targets := make([]AuthorizedTarget, 0, len(threadContext.Targets))
	facts := make([]GroundedFact, 0)
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
		if role == "guest" && target.Kind != TargetStory {
			facts = append(facts, GroundedFact{
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
		if target.Kind == TargetStory {
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

func (loader *DBContextLoader) AuthorizeProposal(ctx context.Context, proposal ActionProposal) error {
	if loader == nil || loader.store == nil || proposal.WorkspaceID == uuid.Nil || proposal.ActorID == uuid.Nil {
		return ErrActionUnauthorized
	}
	allowedTeams, _, role, err := loader.currentTeamScope(ctx, proposal.WorkspaceID, proposal.ActorID)
	if err != nil {
		return err
	}
	if err := validateAuthorizedTeamIDs(allowedTeams); err != nil {
		return err
	}
	if role == "guest" && proposal.Kind != ActionStoryUpdate {
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
	if proposal.Story == nil {
		return nil
	}
	if proposal.Story.Status != nil {
		exists, err := loader.store.StoryStatusExists(
			ctx,
			proposal.WorkspaceID,
			actualTeamID,
			proposal.Story.Status.StatusID,
		)
		if err != nil {
			return fmt.Errorf("reauthorize email story status: %w", err)
		}
		if !exists {
			return ErrActionConflict
		}
	}
	if proposal.Story.Assignee != nil && proposal.Story.Assignee.Operation == AssigneeAssign {
		if proposal.Story.Assignee.AssigneeID == nil {
			return ErrActionConflict
		}
		exists, err := loader.store.StoryAssigneeExists(ctx, actualTeamID, *proposal.Story.Assignee.AssigneeID)
		if err != nil {
			return fmt.Errorf("reauthorize email story assignee: %w", err)
		}
		if !exists {
			return ErrActionConflict
		}
	}
	return nil
}

func validateAuthorizedTeamIDs(teamIDs []uuid.UUID) error {
	if len(teamIDs) > maximumAuthorizedTeams {
		return ErrActionUnauthorized
	}
	seen := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		if teamID == uuid.Nil {
			return ErrActionUnauthorized
		}
		if _, duplicate := seen[teamID]; duplicate {
			return ErrActionUnauthorized
		}
		seen[teamID] = struct{}{}
	}
	return nil
}

// CurrentVersion returns the target's current database version after verifying
// that it still belongs to the proposal workspace and has not been deleted.
func (loader *DBContextLoader) CurrentVersion(ctx context.Context, proposal ActionProposal) (time.Time, error) {
	if loader == nil || loader.store == nil {
		return time.Time{}, errors.New("email reply context loader is not configured")
	}
	target, err := proposalTarget(proposal)
	if err != nil {
		return time.Time{}, err
	}
	version, found, err := loader.store.CurrentVersion(
		ctx,
		proposal.WorkspaceID,
		emailreplydomain.ActionKind(proposal.Kind),
		target.ID,
	)
	if err != nil {
		return time.Time{}, fmt.Errorf("read current email action version: %w", err)
	}
	if !found {
		return time.Time{}, ErrActionUnauthorized
	}
	return version.UTC(), nil
}

// ProposalAlreadyApplied compares only the explicitly proposed fields against
// current database state. It closes the crash window between a committed
// domain CAS write and the durable proposal receipt without treating unrelated
// later edits as proof that Maya applied the proposal.
func (loader *DBContextLoader) ProposalAlreadyApplied(ctx context.Context, proposal ActionProposal) (bool, error) {
	if loader == nil || loader.store == nil {
		return false, errors.New("email reply context loader is not configured")
	}
	target, err := proposalTarget(proposal)
	if err != nil {
		return false, err
	}
	state, found, err := loader.store.CurrentProposalState(
		ctx,
		proposal.WorkspaceID,
		emailreplydomain.ActionKind(proposal.Kind),
		target.ID,
	)
	if err != nil {
		return false, fmt.Errorf("read current email proposal state: %w", err)
	}
	if !found {
		return false, ErrActionUnauthorized
	}
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		if proposal.Objective == nil || proposal.Objective.Health == nil {
			return false, errors.New("objective proposal is incomplete")
		}
		return state.ObjectiveHealth == string(*proposal.Objective.Health), nil
	case ActionKeyResultUpdate:
		if proposal.KeyResult == nil || proposal.KeyResult.CurrentValue == nil {
			return false, errors.New("key result proposal is incomplete")
		}
		return state.KeyResultValue == *proposal.KeyResult.CurrentValue, nil
	case ActionFeedbackStatus:
		if proposal.Feedback == nil {
			return false, errors.New("feedback proposal is incomplete")
		}
		return state.FeedbackStatus == string(proposal.Feedback.Status), nil
	case ActionStoryUpdate:
		if proposal.Story == nil {
			return false, errors.New("story proposal is incomplete")
		}
		return storyProposalAlreadyApplied(state, *proposal.Story)
	default:
		return false, errors.New("unsupported email action kind")
	}
}

func storyProposalAlreadyApplied(state emailreplydomain.ProposalState, action StoryAction) (bool, error) {
	if action.Status != nil && (state.StoryStatusID == nil || *state.StoryStatusID != action.Status.StatusID) {
		return false, nil
	}
	if action.Assignee != nil {
		switch action.Assignee.Operation {
		case AssigneeUnassign:
			if state.StoryAssigneeID != nil {
				return false, nil
			}
		case AssigneeAssign:
			if action.Assignee.AssigneeID == nil || state.StoryAssigneeID == nil ||
				*state.StoryAssigneeID != *action.Assignee.AssigneeID {
				return false, nil
			}
		}
	}
	if action.DueDate != nil {
		switch action.DueDate.Operation {
		case DateClear:
			if state.StoryEndDate != nil {
				return false, nil
			}
		case DateSet:
			expected, err := time.Parse("2006-01-02", action.DueDate.Date)
			if err != nil {
				return false, err
			}
			if state.StoryEndDate == nil || !sameCalendarDate(state.StoryEndDate.UTC(), expected.UTC()) {
				return false, nil
			}
		}
	}
	return true, nil
}

func sameCalendarDate(left, right time.Time) bool {
	leftYear, leftMonth, leftDay := left.Date()
	rightYear, rightMonth, rightDay := right.Date()
	return leftYear == rightYear && leftMonth == rightMonth && leftDay == rightDay
}

func (loader *DBContextLoader) currentTeamScope(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) ([]uuid.UUID, string, string, error) {
	scope, found, err := loader.store.ActorScope(ctx, workspaceID, userID)
	if err != nil {
		return nil, "", "", fmt.Errorf("authorize email conversation actor: %w", err)
	}
	if !found {
		return nil, "", "", ErrActionUnauthorized
	}
	return append([]uuid.UUID(nil), scope.TeamIDs...), scope.WorkspaceSlug, scope.Role, nil
}

func (loader *DBContextLoader) loadTarget(
	ctx context.Context,
	workspaceID uuid.UUID,
	input emailthread.TargetContext,
) (AuthorizedTarget, bool, error) {
	if input.ID == uuid.Nil {
		return AuthorizedTarget{}, false, nil
	}
	kind := emailreplydomain.TargetKind(strings.TrimSpace(input.Kind))
	value, found, err := loader.store.LoadTarget(ctx, workspaceID, kind, input.ID)
	if err != nil {
		return AuthorizedTarget{}, false, fmt.Errorf("load email %s target: %w", kind, err)
	}
	if !found {
		return AuthorizedTarget{}, false, nil
	}

	switch value.Kind {
	case emailreplydomain.TargetObjective:
		health := strings.TrimSpace(value.Health)
		if health == "" {
			health = "not set"
		}
		state := "Health: " + health + dateState(value.StartDate, value.EndDate)
		return authorizedTarget(TargetObjective, value.ID, value.TeamID, value.Name, state, value.UpdatedAt), true, nil
	case emailreplydomain.TargetKeyResult:
		state := fmt.Sprintf(
			"Current value: %s; target value: %s; measurement: %s%s",
			formatFloat(value.CurrentValue),
			formatFloat(value.TargetValue),
			value.MeasurementType,
			dateState(value.StartDate, value.EndDate),
		)
		return authorizedTarget(TargetKeyResult, value.ID, value.TeamID, value.Name, state, value.UpdatedAt), true, nil
	case emailreplydomain.TargetStory:
		status, assignee, due := strings.TrimSpace(value.StatusName), strings.TrimSpace(value.AssigneeName), "not set"
		if status == "" {
			status = "not set"
		}
		if assignee == "" {
			assignee = "unassigned"
		}
		if value.EndDate != nil {
			due = value.EndDate.UTC().Format("2006-01-02")
		}
		state := fmt.Sprintf("Status: %s; assignee: %s; due date: %s", status, assignee, due)
		return authorizedTarget(TargetStory, value.ID, value.TeamID, value.Name, state, value.UpdatedAt), true, nil
	case emailreplydomain.TargetFeedback:
		return authorizedTarget(
			TargetFeedback,
			value.ID,
			value.TeamID,
			value.Name,
			"Status: "+value.Status,
			value.UpdatedAt,
		), true, nil
	default:
		return AuthorizedTarget{}, false, nil
	}
}

func (loader *DBContextLoader) storyChoices(
	ctx context.Context,
	workspaceID uuid.UUID,
	teams map[uuid.UUID]struct{},
) ([]AuthorizedChoice, error) {
	teamIDs := make([]uuid.UUID, 0, len(teams))
	for teamID := range teams {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(left, right int) bool { return teamIDs[left].String() < teamIDs[right].String() })
	choices := make([]AuthorizedChoice, 0)
	for _, teamID := range teamIDs {
		remaining := maximumAgentChoices - len(choices)
		rows, err := loader.store.ListStoryChoices(ctx, workspaceID, teamID, remaining)
		if err != nil {
			return nil, fmt.Errorf("load email story choices: %w", err)
		}
		for _, row := range rows {
			kind, prefix := ChoiceStoryAssignee, "assignee"
			if row.Kind == emailreplydomain.ChoiceStoryStatus {
				kind, prefix = ChoiceStoryStatus, "status"
			}
			choices = append(choices, AuthorizedChoice{
				Reference:   choiceReference(prefix, row.ID),
				Kind:        kind,
				DisplayName: row.Name,
				ID:          row.ID,
				TeamID:      row.TeamID,
			})
			if len(choices) == maximumAgentChoices {
				return choices, nil
			}
		}
	}
	return choices, nil
}

func (loader *DBContextLoader) currentTargetTeam(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind ActionKind,
	entityID uuid.UUID,
) (uuid.UUID, bool, error) {
	teamID, found, err := loader.store.TargetTeam(ctx, workspaceID, emailreplydomain.ActionKind(kind), entityID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("reauthorize email action target: %w", err)
	}
	return teamID, found, nil
}

func authorizedTarget(kind TargetKind, id, teamID uuid.UUID, name, state string, updatedAt time.Time) AuthorizedTarget {
	return AuthorizedTarget{
		Kind: kind, ID: id, TeamID: teamID, DisplayName: strings.TrimSpace(name),
		CurrentState: strings.TrimSpace(state), ExpectedUpdatedAt: updatedAt.UTC(),
	}
}

func targetReference(kind TargetKind, id uuid.UUID) string {
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
