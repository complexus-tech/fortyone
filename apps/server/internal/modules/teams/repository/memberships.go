package teamsrepository

import (
	"context"
	"fmt"
	"strings"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) AddMember(ctx context.Context, teamID, userID, workspaceID uuid.UUID) error {
	outcome, err := r.queries.AddTeamMemberForWorkspace(ctx, teamsql.AddTeamMemberForWorkspaceParams{
		UserID:      userID,
		TeamID:      teamID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("add scoped team member: %w", err)
	}
	return validateMembershipAdd(outcome.Eligible, outcome.Added)
}

func (transaction *workspaceTransaction) AddMember(ctx context.Context, teamID, userID, workspaceID uuid.UUID) error {
	outcome, err := transaction.queries.AddTeamMemberForWorkspace(
		ctx,
		teamsql.AddTeamMemberForWorkspaceParams{
			UserID:      userID,
			TeamID:      teamID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		return fmt.Errorf("add team member in workspace transaction: %w", err)
	}
	return validateMembershipAdd(outcome.Eligible, outcome.Added)
}

func validateMembershipAdd(eligible, added bool) error {
	if !eligible {
		return teamsdomain.ErrNotFound
	}
	if !added {
		return teamsdomain.ErrMemberExists
	}
	return nil
}

// JoinPublicTeam atomically authorizes and adds the authenticated actor.
func (r *repo) JoinPublicTeam(ctx context.Context, input teamsdomain.PublicTeamJoin) error {
	outcome, err := r.queries.JoinPublicTeamForActor(ctx, teamsql.JoinPublicTeamForActorParams{
		ActorID:     input.ActorID,
		TeamID:      input.TeamID,
		WorkspaceID: input.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("join scoped public team: %w", err)
	}
	return validatePublicTeamJoin(outcome.Eligible, outcome.Joined)
}

func validatePublicTeamJoin(eligible, joined bool) error {
	if !eligible {
		return teamsdomain.ErrNotFound
	}
	if !joined {
		return teamsdomain.ErrMemberExists
	}
	return nil
}

func (r *repo) RemoveMember(ctx context.Context, teamID, userID, workspaceID uuid.UUID) error {
	rowsAffected, err := r.queries.RemoveTeamMemberForWorkspace(
		ctx,
		teamsql.RemoveTeamMemberForWorkspaceParams{
			TeamID:      teamID,
			UserID:      userID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		return fmt.Errorf("remove scoped team member: %w", err)
	}
	if rowsAffected == 0 {
		return teamsdomain.ErrMemberNotFound
	}
	return nil
}

// LeaveTeam deletes only the authenticated actor's membership in the scoped workspace.
func (r *repo) LeaveTeam(ctx context.Context, input teamsdomain.TeamSelfLeave) error {
	rowsAffected, err := r.queries.LeaveTeamForActor(ctx, teamsql.LeaveTeamForActorParams{
		TeamID:      input.TeamID,
		ActorID:     input.ActorID,
		WorkspaceID: input.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("leave scoped team: %w", err)
	}
	if rowsAffected == 0 {
		return teamsdomain.ErrMemberNotFound
	}
	return nil
}

func (r *repo) UpdateMemberAIContext(
	ctx context.Context,
	teamID, userID, workspaceID uuid.UUID,
	input teamsdomain.MemberAIContext,
) error {
	rowsAffected, err := r.queries.UpdateTeamMemberAIContextForWorkspace(
		ctx,
		teamsql.UpdateTeamMemberAIContextForWorkspaceParams{
			AiRoleTitle:       strings.TrimSpace(input.RoleTitle),
			AiRoleDescription: strings.TrimSpace(input.RoleDescription),
			TeamID:            teamID,
			UserID:            userID,
			WorkspaceID:       workspaceID,
		},
	)
	if err != nil {
		return fmt.Errorf("update scoped team member AI context: %w", err)
	}
	if rowsAffected == 0 {
		return teamsdomain.ErrMemberNotFound
	}
	return nil
}
