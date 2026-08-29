package keyresultsrepository

import (
	"context"
	"fmt"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresultssql "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Delete(ctx context.Context, command keyresultsdomain.DeleteCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	err := repository.withinTransaction(ctx, func(queries keyresultssql.Querier) error {
		if _, err := queries.GetKeyResultForMutation(ctx, keyresultssql.GetKeyResultForMutationParams{
			ActorID: command.Access.ActorID, KeyResultID: command.KeyResultID,
			WorkspaceID: uuidPointer(command.Access.WorkspaceID), AllTeams: command.Access.AllTeams,
			AllowedTeamIds: command.Access.TeamIDs,
		}); err != nil {
			if err == pgx.ErrNoRows {
				return keyresultsdomain.ErrNotFound
			}
			return fmt.Errorf("lock key result for delete: %w", err)
		}
		deleted, err := queries.DeleteKeyResult(ctx, keyresultssql.DeleteKeyResultParams{
			KeyResultID: command.KeyResultID, WorkspaceID: uuidPointer(command.Access.WorkspaceID),
			ActorID: command.Access.ActorID, AllTeams: command.Access.AllTeams,
			AllowedTeamIds: command.Access.TeamIDs,
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return keyresultsdomain.ErrNotFound
			}
			return fmt.Errorf("delete key result: %w", err)
		}
		value := deleted.Name
		rows, err := queries.CreateDeletedKeyResultActivity(ctx, keyresultssql.CreateDeletedKeyResultActivityParams{
			CurrentValue: &value, ActorID: command.Access.ActorID,
			ObjectiveID: deleted.ObjectiveID, WorkspaceID: uuidPointer(command.Access.WorkspaceID),
		})
		if err != nil {
			return fmt.Errorf("record key result delete activity: %w", err)
		}
		if rows != 1 {
			return keyresultsdomain.ErrForbidden
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete key result: %w", err)
	}
	return nil
}
