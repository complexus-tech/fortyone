package usersrepository

import (
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
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
