package usersrepository

import (
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
)

func mapWorkspaceUsers(rows []usersql.ListWorkspaceUsersRow) []usersdomain.User {
	result := make([]usersdomain.User, len(rows))
	for index, row := range rows {
		role := row.Role
		user := toCoreUser(userRow{
			id: row.UserID, username: row.Username, email: row.Email,
			fullName: row.FullName, avatarURL: row.AvatarURL,
			isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
			hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
			workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
			workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
			lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
			createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		})
		user.Role = &role
		user.TeamAIRoleTitle = row.TeamAiRoleTitle
		user.TeamAIRoleDescription = row.TeamAiRoleDescription
		user.InferredTeamAIRoleTitle = row.InferredTeamAiRoleTitle
		user.InferredTeamAIRoleDescription = row.InferredTeamAiRoleDescription
		user.InferredTeamAIRoleStoryCount = int(row.InferredTeamAiRoleStoryCount)
		user.InferredTeamAIRoleConfidence = row.InferredTeamAiRoleConfidence
		user.InferredTeamAIRoleGeneratedAt = row.InferredAiRoleGeneratedAt
		user.LastStoryActivityAt = optionalTime(row.HasLastStoryActivity, row.LastStoryActivityAt)
		result[index] = user
	}
	return result
}
