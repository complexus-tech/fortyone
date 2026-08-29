package usersrepository

import (
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
)

func mapActiveUserByID(row usersql.GetActiveUserByIDRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapActiveUserByEmail(row usersql.GetActiveUserByEmailRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapUserByEmailAnyStatus(row usersql.GetUserByEmailAnyStatusRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapUsersByIDs(rows []usersql.ListUsersByIDsRow) []usersdomain.User {
	result := make([]usersdomain.User, len(rows))
	for index, row := range rows {
		result[index] = toCoreUser(userRow{
			id: row.UserID, username: row.Username, email: row.Email,
			fullName: row.FullName, avatarURL: row.AvatarURL,
			isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
			hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
			workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
			workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
			lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
			createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		})
	}
	return result
}

func mapCreatedUser(row usersql.CreateUserRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapUpdatedUser(row usersql.UpdateActiveUserRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}

func mapVerifiedSignInReactivatedUser(row usersql.ReactivateUserForVerifiedSignInRow) usersdomain.User {
	return toCoreUser(userRow{
		id: row.UserID, username: row.Username, email: row.Email,
		fullName: row.FullName, avatarURL: row.AvatarURL,
		isActive: row.IsActive, isSystem: row.IsSystem, isInternal: row.IsInternal,
		hasSeenWalkthrough: row.HasSeenWalkthrough, timezone: row.Timezone,
		workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute,
		workingEndMinute: row.WorkingEndMinute, lastLoginAt: row.LastLoginAt,
		lastUsedWorkspaceID: row.LastUsedWorkspaceID, githubUsername: row.GithubUsername,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
	})
}
