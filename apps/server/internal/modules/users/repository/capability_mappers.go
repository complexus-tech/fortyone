package usersrepository

import (
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
)

func mapAutomationPreferences(
	row usersql.GetOrCreateAutomationPreferencesForMemberRow,
) usersdomain.AutomationPreferences {
	return usersdomain.AutomationPreferences{
		UserID:                     row.UserID,
		WorkspaceID:                row.WorkspaceID,
		AutoAssignSelf:             row.AutoAssignSelf,
		AutoScheduling:             row.AutoScheduling,
		AssignSelfOnBranchCopy:     row.AssignSelfOnBranchCopy,
		MoveStoryToStartedOnBranch: row.MoveStoryToStartedOnBranch,
		OpenStoryInDialog:          row.OpenStoryInDialog,
		CreatedAt:                  row.CreatedAt,
		UpdatedAt:                  row.UpdatedAt,
	}
}

func mapOnboardingTourProgress(
	userID, workspaceID uuid.UUID,
	tourKey, tourVersion string,
	completedStepIDs, completedActionIDs []string,
	status string,
	createdAt, updatedAt time.Time,
) usersdomain.OnboardingTourProgress {
	return usersdomain.OnboardingTourProgress{
		UserID:             userID,
		WorkspaceID:        workspaceID,
		TourKey:            tourKey,
		TourVersion:        tourVersion,
		CompletedStepIDs:   append([]string(nil), completedStepIDs...),
		CompletedActionIDs: append([]string(nil), completedActionIDs...),
		Status:             usersdomain.OnboardingTourStatus(status),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}

func mapCreatedMemory(row usersql.CreateUserMemoryForMemberRow) usersdomain.UserMemoryItem {
	return usersdomain.UserMemoryItem{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		UserID:      row.UserID,
		Content:     row.Content,
		CreatedAt:   derefTime(row.CreatedAt),
		UpdatedAt:   derefTime(row.UpdatedAt),
	}
}

func mapUserMemories(rows []usersql.ListUserMemoriesForOwnerRow) []usersdomain.UserMemoryItem {
	result := make([]usersdomain.UserMemoryItem, len(rows))
	for index, row := range rows {
		result[index] = usersdomain.UserMemoryItem{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			UserID:      row.UserID,
			Content:     row.Content,
			CreatedAt:   derefTime(row.CreatedAt),
			UpdatedAt:   derefTime(row.UpdatedAt),
		}
	}
	return result
}
