package keyresultsrepository

import (
	"context"
	"fmt"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresultssql "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) CreateBatch(
	ctx context.Context,
	command keyresultsdomain.CreateCommand,
) ([]keyresultsdomain.KeyResult, error) {
	normalized, err := command.Normalize()
	if err != nil {
		return nil, err
	}
	created := make([]keyresultsdomain.KeyResult, 0, len(normalized.KeyResults))
	err = repository.withinTransaction(ctx, func(queries keyresultssql.Querier) error {
		objectiveID := normalized.KeyResults[0].ObjectiveID
		teamID, err := queries.GetKeyResultCreateScope(ctx, keyresultssql.GetKeyResultCreateScopeParams{
			ActorID: normalized.Access.ActorID, ObjectiveID: objectiveID,
			WorkspaceID: uuidPointer(normalized.Access.WorkspaceID), AllTeams: normalized.Access.AllTeams,
			AllowedTeamIds: normalized.Access.TeamIDs,
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return keyresultsdomain.ErrForbidden
			}
			return fmt.Errorf("resolve key result objective: %w", err)
		}
		if teamID == nil || *teamID == uuid.Nil {
			return keyresultsdomain.ErrInvalidReference
		}
		assignees := collectCreateAssignees(normalized.KeyResults)
		valid, err := queries.ValidateKeyResultAssignees(ctx, keyresultssql.ValidateKeyResultAssigneesParams{
			UserIds: assignees, WorkspaceID: normalized.Access.WorkspaceID, TeamID: *teamID,
		})
		if err != nil {
			return fmt.Errorf("validate key result assignees: %w", err)
		}
		if !valid {
			return keyresultsdomain.ErrInvalidReference
		}

		sequenceCount, err := safecast.Int32(len(normalized.KeyResults))
		if err != nil {
			return fmt.Errorf("validate key result sequence count: %w", err)
		}
		finalSequence, err := queries.AllocateKeyResultSequences(ctx, keyresultssql.AllocateKeyResultSequencesParams{
			SequenceCount: sequenceCount, WorkspaceID: normalized.Access.WorkspaceID, TeamID: *teamID,
		})
		if err != nil {
			return fmt.Errorf("allocate key result sequences: %w", err)
		}
		firstSequence, err := firstSequenceID(finalSequence, sequenceCount)
		if err != nil {
			return err
		}

		for index, draft := range normalized.KeyResults {
			sequenceOffset, err := safecast.Int32(index)
			if err != nil {
				return fmt.Errorf("calculate key result sequence offset: %w", err)
			}
			sequenceID, err := safecast.Int64ToInt32(int64(firstSequence) + int64(sequenceOffset))
			if err != nil || sequenceID <= 0 {
				return fmt.Errorf("calculate key result sequence: %w", keyresultsdomain.ErrInvalid)
			}
			row, err := queries.CreateKeyResult(ctx, keyresultssql.CreateKeyResultParams{
				SequenceID: sequenceID, Name: draft.Name,
				MeasurementType: keyresultssql.MeasurementType(draft.MeasurementType),
				StartValue:      draft.StartValue, CurrentValue: draft.CurrentValue, TargetValue: draft.TargetValue,
				LeadUserID: draft.Lead, StartDate: *draft.StartDate, EndDate: *draft.EndDate,
				ActorID: normalized.Access.ActorID, ObjectiveID: objectiveID,
				WorkspaceID: uuidPointer(normalized.Access.WorkspaceID), TeamID: teamID,
				AllTeams: normalized.Access.AllTeams, AllowedTeamIds: normalized.Access.TeamIDs,
			})
			if err != nil {
				return fmt.Errorf("create key result %d: %w", index, err)
			}
			for _, contributorID := range draft.Contributors {
				rows, err := queries.AddKeyResultContributor(ctx, keyresultssql.AddKeyResultContributorParams{
					UserID: contributorID, KeyResultID: row.ID,
					WorkspaceID: uuidPointer(normalized.Access.WorkspaceID),
				})
				if err != nil {
					return fmt.Errorf("add key result contributor: %w", err)
				}
				if rows != 1 {
					return keyresultsdomain.ErrInvalidReference
				}
			}
			field, value, comment := "all", draft.Name, ""
			rows, err := queries.CreateKeyResultActivity(ctx, keyresultssql.CreateKeyResultActivityParams{
				ActivityType: keyresultssql.OkrActivityTypeCreate,
				FieldChanged: &field, CurrentValue: &value, Comment: &comment,
				ActorID: normalized.Access.ActorID, KeyResultID: row.ID,
				WorkspaceID: uuidPointer(normalized.Access.WorkspaceID),
			})
			if err != nil {
				return fmt.Errorf("record key result create activity: %w", err)
			}
			if rows != 1 {
				return keyresultsdomain.ErrForbidden
			}
			created = append(created, keyResultFromCreateRow(row, draft.Contributors))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create key results: %w", err)
	}
	return created, nil
}

func collectCreateAssignees(drafts []keyresultsdomain.NewKeyResult) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	result := make([]uuid.UUID, 0)
	for _, draft := range drafts {
		if draft.Lead != nil {
			seen[*draft.Lead] = struct{}{}
		}
		for _, contributorID := range draft.Contributors {
			seen[contributorID] = struct{}{}
		}
	}
	for candidate := range seen {
		result = append(result, candidate)
	}
	return result
}

func firstSequenceID(finalSequence, count int32) (int32, error) {
	first, err := safecast.Int64ToInt32(int64(finalSequence) - int64(count) + 1)
	if err != nil || first <= 0 {
		return 0, fmt.Errorf("invalid key result sequence range: %w", keyresultsdomain.ErrInvalid)
	}
	return first, nil
}
