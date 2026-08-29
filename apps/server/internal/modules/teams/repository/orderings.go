package teamsrepository

import (
	"context"
	"fmt"
	"math"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) UpdateUserTeamOrdering(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	teamIDs []uuid.UUID,
) error {
	if len(teamIDs) > math.MaxInt32 {
		return fmt.Errorf("team ordering contains too many teams: %d", len(teamIDs))
	}

	return r.withinTransaction(ctx, func(queries teamsql.Querier) error {
		allowed, err := queries.ActorCanOrderWorkspaceTeams(
			ctx,
			teamsql.ActorCanOrderWorkspaceTeamsParams{
				WorkspaceID: workspaceID,
				ActorID:     userID,
			},
		)
		if err != nil {
			return fmt.Errorf("authorize team ordering: %w", err)
		}
		if !allowed {
			return teamsdomain.ErrNotFound
		}

		if err := queries.DeleteActorTeamOrdering(ctx, teamsql.DeleteActorTeamOrderingParams{
			ActorID:     userID,
			WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("clear existing team ordering: %w", err)
		}

		seen := make(map[uuid.UUID]struct{}, len(teamIDs))
		for index, teamID := range teamIDs {
			if _, duplicate := seen[teamID]; duplicate {
				return fmt.Errorf("team ordering contains duplicate team %s", teamID)
			}
			seen[teamID] = struct{}{}

			rowsAffected, err := queries.InsertActorTeamOrder(ctx, teamsql.InsertActorTeamOrderParams{
				ActorID:     userID,
				OrderIndex:  int32(index),
				TeamID:      teamID,
				WorkspaceID: workspaceID,
			})
			if err != nil {
				return fmt.Errorf("insert team ordering position %d: %w", index, err)
			}
			if rowsAffected != 1 {
				return teamsdomain.ErrNotFound
			}
		}
		return nil
	})
}
