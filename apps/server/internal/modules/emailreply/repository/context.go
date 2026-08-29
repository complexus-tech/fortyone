package emailreplyrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	emailreplydomain "github.com/complexus-tech/projects-api/internal/modules/emailreply/domain"
	emailreplysql "github.com/complexus-tech/projects-api/internal/modules/emailreply/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries emailreplysql.Querier
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return nil
	}
	return newWithQueries(emailreplysql.New(pool))
}

func newWithQueries(queries emailreplysql.Querier) *Repository {
	return &Repository{queries: queries}
}

func (repository *Repository) ActorScope(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (emailreplydomain.ActorScope, bool, error) {
	actor, err := repository.queries.GetEmailReplyActorWorkspace(ctx, emailreplysql.GetEmailReplyActorWorkspaceParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return emailreplydomain.ActorScope{}, false, nil
	}
	if err != nil {
		return emailreplydomain.ActorScope{}, false, fmt.Errorf("read email reply actor workspace: %w", err)
	}

	teamIDs, err := repository.queries.ListEmailReplyActorTeams(ctx, emailreplysql.ListEmailReplyActorTeamsParams{
		WorkspaceID: workspaceID,
		ActorRole:   string(actor.Role),
		UserID:      userID,
	})
	if err != nil {
		return emailreplydomain.ActorScope{}, false, fmt.Errorf("read email reply actor teams: %w", err)
	}
	return emailreplydomain.ActorScope{
		WorkspaceSlug: actor.Slug,
		Role:          string(actor.Role),
		TeamIDs:       append([]uuid.UUID(nil), teamIDs...),
	}, true, nil
}

func (repository *Repository) LoadTarget(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind emailreplydomain.TargetKind,
	entityID uuid.UUID,
) (emailreplydomain.TargetSnapshot, bool, error) {
	switch kind {
	case emailreplydomain.TargetObjective:
		row, err := repository.queries.GetEmailReplyObjectiveTarget(ctx, emailreplysql.GetEmailReplyObjectiveTargetParams{
			WorkspaceID: workspaceID,
			EntityID:    entityID,
		})
		if err != nil {
			return noTarget("objective", err)
		}
		teamID, err := requiredTeamID(row.TeamID)
		if err != nil {
			return emailreplydomain.TargetSnapshot{}, false, fmt.Errorf("map email objective target: %w", err)
		}
		return emailreplydomain.TargetSnapshot{
			Kind: emailreplydomain.TargetObjective, ID: row.ID, TeamID: teamID, Name: row.Name,
			Health: row.Health, StartDate: row.StartDate, EndDate: row.EndDate, UpdatedAt: row.UpdatedAt,
		}, true, nil

	case emailreplydomain.TargetKeyResult:
		row, err := repository.queries.GetEmailReplyKeyResultTarget(ctx, emailreplysql.GetEmailReplyKeyResultTargetParams{
			WorkspaceID: workspaceID,
			EntityID:    entityID,
		})
		if err != nil {
			return noTarget("key result", err)
		}
		teamID, err := requiredTeamID(row.TeamID)
		if err != nil {
			return emailreplydomain.TargetSnapshot{}, false, fmt.Errorf("map email key result target: %w", err)
		}
		startDate, endDate := row.StartDate, row.EndDate
		return emailreplydomain.TargetSnapshot{
			Kind: emailreplydomain.TargetKeyResult, ID: row.ID, TeamID: teamID, Name: row.Name,
			MeasurementType: string(row.MeasurementType), CurrentValue: row.CurrentValue, TargetValue: row.TargetValue,
			StartDate: &startDate, EndDate: &endDate, UpdatedAt: row.UpdatedAt,
		}, true, nil

	case emailreplydomain.TargetStory:
		row, err := repository.queries.GetEmailReplyStoryTarget(ctx, emailreplysql.GetEmailReplyStoryTargetParams{
			WorkspaceID: workspaceID,
			EntityID:    entityID,
		})
		if err != nil {
			return noTarget("story", err)
		}
		return emailreplydomain.TargetSnapshot{
			Kind: emailreplydomain.TargetStory, ID: row.ID, TeamID: row.TeamID, Name: row.Title,
			StatusName: valueOrEmpty(row.StatusName), AssigneeName: row.AssigneeName,
			EndDate: row.EndDate, UpdatedAt: row.UpdatedAt,
		}, true, nil

	case emailreplydomain.TargetFeedback:
		row, err := repository.queries.GetEmailReplyFeedbackTarget(ctx, emailreplysql.GetEmailReplyFeedbackTargetParams{
			WorkspaceID: workspaceID,
			EntityID:    entityID,
		})
		if err != nil {
			return noTarget("feedback", err)
		}
		return emailreplydomain.TargetSnapshot{
			Kind: emailreplydomain.TargetFeedback, ID: row.ID, TeamID: row.TeamID, Name: row.Title,
			Status: row.Status, UpdatedAt: row.UpdatedAt,
		}, true, nil
	default:
		return emailreplydomain.TargetSnapshot{}, false, nil
	}
}

func noTarget(kind string, err error) (emailreplydomain.TargetSnapshot, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return emailreplydomain.TargetSnapshot{}, false, nil
	}
	return emailreplydomain.TargetSnapshot{}, false, fmt.Errorf("read email %s target: %w", kind, err)
}

func (repository *Repository) ListStoryChoices(
	ctx context.Context,
	workspaceID uuid.UUID,
	teamID uuid.UUID,
	limit int,
) ([]emailreplydomain.Choice, error) {
	if limit <= 0 {
		return []emailreplydomain.Choice{}, nil
	}
	remaining, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("validate email story choice limit: %w", err)
	}
	statuses, err := repository.queries.ListEmailReplyStoryStatuses(ctx, emailreplysql.ListEmailReplyStoryStatusesParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("read email story status choices: %w", err)
	}
	choices := make([]emailreplydomain.Choice, 0, min(len(statuses), limit))
	for _, status := range statuses {
		if remaining == 0 {
			return choices, nil
		}
		choices = append(choices, emailreplydomain.Choice{
			Kind: emailreplydomain.ChoiceStoryStatus, ID: status.ID, TeamID: teamID, Name: status.Name,
		})
		remaining--
	}

	assignees, err := repository.queries.ListEmailReplyStoryAssignees(ctx, emailreplysql.ListEmailReplyStoryAssigneesParams{
		TeamID:   teamID,
		RowLimit: remaining,
	})
	if err != nil {
		return nil, fmt.Errorf("read email story assignee choices: %w", err)
	}
	for _, assignee := range assignees {
		choices = append(choices, emailreplydomain.Choice{
			Kind: emailreplydomain.ChoiceStoryAssignee, ID: assignee.ID, TeamID: teamID, Name: assignee.Name,
		})
	}
	return choices, nil
}

func (repository *Repository) StoryStatusExists(
	ctx context.Context,
	workspaceID uuid.UUID,
	teamID uuid.UUID,
	statusID uuid.UUID,
) (bool, error) {
	exists, err := repository.queries.StoryStatusExists(ctx, emailreplysql.StoryStatusExistsParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		StatusID:    statusID,
	})
	if err != nil {
		return false, fmt.Errorf("read email story status authorization: %w", err)
	}
	return exists, nil
}

func (repository *Repository) StoryAssigneeExists(
	ctx context.Context,
	teamID uuid.UUID,
	userID uuid.UUID,
) (bool, error) {
	exists, err := repository.queries.StoryAssigneeExists(ctx, emailreplysql.StoryAssigneeExistsParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("read email story assignee authorization: %w", err)
	}
	return exists, nil
}

func (repository *Repository) CurrentVersion(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind emailreplydomain.ActionKind,
	entityID uuid.UUID,
) (time.Time, bool, error) {
	var (
		version time.Time
		err     error
	)
	switch kind {
	case emailreplydomain.ActionObjectiveUpdate:
		version, err = repository.queries.GetObjectiveVersion(ctx, emailreplysql.GetObjectiveVersionParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionKeyResultUpdate:
		version, err = repository.queries.GetKeyResultVersion(ctx, emailreplysql.GetKeyResultVersionParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionStoryUpdate:
		version, err = repository.queries.GetStoryVersion(ctx, emailreplysql.GetStoryVersionParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionFeedbackStatus:
		version, err = repository.queries.GetFeedbackVersion(ctx, emailreplysql.GetFeedbackVersionParams{WorkspaceID: workspaceID, EntityID: entityID})
	default:
		return time.Time{}, false, fmt.Errorf("unsupported email action kind %q", kind)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read email action version: %w", err)
	}
	return version.UTC(), true, nil
}

func (repository *Repository) CurrentProposalState(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind emailreplydomain.ActionKind,
	entityID uuid.UUID,
) (emailreplydomain.ProposalState, bool, error) {
	state := emailreplydomain.ProposalState{Kind: kind}
	var err error
	switch kind {
	case emailreplydomain.ActionObjectiveUpdate:
		state.ObjectiveHealth, err = repository.queries.GetObjectiveHealth(ctx, emailreplysql.GetObjectiveHealthParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionKeyResultUpdate:
		state.KeyResultValue, err = repository.queries.GetKeyResultCurrentValue(ctx, emailreplysql.GetKeyResultCurrentValueParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionFeedbackStatus:
		state.FeedbackStatus, err = repository.queries.GetFeedbackStatus(ctx, emailreplysql.GetFeedbackStatusParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionStoryUpdate:
		var row emailreplysql.GetStoryReconciliationStateRow
		row, err = repository.queries.GetStoryReconciliationState(ctx, emailreplysql.GetStoryReconciliationStateParams{WorkspaceID: workspaceID, EntityID: entityID})
		state.StoryStatusID = row.StatusID
		state.StoryAssigneeID = row.AssigneeID
		state.StoryEndDate = row.EndDate
	default:
		return emailreplydomain.ProposalState{}, false, fmt.Errorf("unsupported email action kind %q", kind)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return emailreplydomain.ProposalState{}, false, nil
	}
	if err != nil {
		return emailreplydomain.ProposalState{}, false, fmt.Errorf("read email proposal state: %w", err)
	}
	return state, true, nil
}

func (repository *Repository) TargetTeam(
	ctx context.Context,
	workspaceID uuid.UUID,
	kind emailreplydomain.ActionKind,
	entityID uuid.UUID,
) (uuid.UUID, bool, error) {
	var (
		teamID *uuid.UUID
		err    error
	)
	switch kind {
	case emailreplydomain.ActionObjectiveUpdate:
		teamID, err = repository.queries.GetObjectiveTeam(ctx, emailreplysql.GetObjectiveTeamParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionKeyResultUpdate:
		teamID, err = repository.queries.GetKeyResultTeam(ctx, emailreplysql.GetKeyResultTeamParams{WorkspaceID: workspaceID, EntityID: entityID})
	case emailreplydomain.ActionStoryUpdate:
		var value uuid.UUID
		value, err = repository.queries.GetStoryTeam(ctx, emailreplysql.GetStoryTeamParams{WorkspaceID: workspaceID, EntityID: entityID})
		teamID = &value
	case emailreplydomain.ActionFeedbackStatus:
		var value uuid.UUID
		value, err = repository.queries.GetFeedbackTeam(ctx, emailreplysql.GetFeedbackTeamParams{WorkspaceID: workspaceID, EntityID: entityID})
		teamID = &value
	default:
		return uuid.Nil, false, fmt.Errorf("unsupported email action kind %q", kind)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("read email action target team: %w", err)
	}
	if teamID == nil || *teamID == uuid.Nil {
		return uuid.Nil, false, nil
	}
	return *teamID, true, nil
}

func requiredTeamID(teamID *uuid.UUID) (uuid.UUID, error) {
	if teamID == nil || *teamID == uuid.Nil {
		return uuid.Nil, errors.New("target has no team")
	}
	return *teamID, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
