package integrationrequestsrepository

import (
	"context"
	"fmt"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	"github.com/google/uuid"
)

func validatePendingRequestProperties(
	ctx context.Context,
	queries integrationrequestssql.Querier,
	workspaceID, teamID uuid.UUID,
	statusID, assigneeID, objectiveID, keyResultID, sprintID *uuid.UUID,
	labelIDs []uuid.UUID,
	startDate, endDate *time.Time,
) error {
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return invalidProperty("start date must not be after deadline")
	}
	if statusID != nil {
		available, err := queries.IntegrationRequestStatusAvailable(ctx, integrationrequestssql.IntegrationRequestStatusAvailableParams{
			StatusID: *statusID, WorkspaceID: workspaceID, TeamID: teamID,
		})
		if err := requireReference("status", available, err); err != nil {
			return err
		}
	}
	if assigneeID != nil {
		available, err := queries.IntegrationRequestAssigneeAvailable(ctx, integrationrequestssql.IntegrationRequestAssigneeAvailableParams{
			WorkspaceID: workspaceID, TeamID: teamID, AssigneeID: *assigneeID,
		})
		if err := requireReference("assignee", available, err); err != nil {
			return err
		}
	}
	if objectiveID != nil {
		available, err := queries.IntegrationRequestObjectiveAvailable(ctx, integrationrequestssql.IntegrationRequestObjectiveAvailableParams{
			ObjectiveID: *objectiveID, WorkspaceID: uuidPointer(workspaceID), TeamID: uuidPointer(teamID),
		})
		if err := requireReference("objective", available, err); err != nil {
			return err
		}
	}
	if keyResultID != nil {
		if objectiveID == nil {
			return invalidProperty("key result requires an objective")
		}
		available, err := queries.IntegrationRequestKeyResultAvailable(ctx, integrationrequestssql.IntegrationRequestKeyResultAvailableParams{
			KeyResultID: *keyResultID, ObjectiveID: *objectiveID,
			WorkspaceID: uuidPointer(workspaceID), TeamID: uuidPointer(teamID),
		})
		if err := requireReference("key result", available, err); err != nil {
			return err
		}
	}
	if sprintID != nil {
		available, err := queries.IntegrationRequestSprintAvailable(ctx, integrationrequestssql.IntegrationRequestSprintAvailableParams{
			SprintID: *sprintID, WorkspaceID: workspaceID, TeamID: teamID,
		})
		if err := requireReference("sprint", available, err); err != nil {
			return err
		}
	}
	if len(labelIDs) == 0 {
		return nil
	}
	count, err := queries.CountAvailableIntegrationRequestLabels(ctx, integrationrequestssql.CountAvailableIntegrationRequestLabelsParams{
		WorkspaceID: uuidPointer(workspaceID), TeamID: uuidPointer(teamID), LabelIds: labelIDs,
	})
	if err != nil {
		return fmt.Errorf("validate labels: %w", err)
	}
	if count != int64(len(labelIDs)) {
		return invalidProperty("one or more labels are not available to the team")
	}
	return nil
}

func requireReference(name string, available bool, err error) error {
	if err != nil {
		return fmt.Errorf("validate %s: %w", name, err)
	}
	if !available {
		return invalidProperty(name + " is not available to the team")
	}
	return nil
}

func requestLabelIDs(current []uuid.UUID, patch *[]uuid.UUID) ([]uuid.UUID, error) {
	values := current
	if patch != nil {
		values = *patch
	}
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, labelID := range values {
		if labelID == uuid.Nil {
			if patch == nil {
				return nil, invalidProperty("stored label id is invalid")
			}
			return nil, invalidProperty("label id is required")
		}
		if _, exists := seen[labelID]; exists {
			continue
		}
		seen[labelID] = struct{}{}
		result = append(result, labelID)
	}
	return result, nil
}

func invalidProperty(message string) error {
	return fmt.Errorf("%w: %s", integrationrequestdomain.ErrInvalidRequestProperty, message)
}
