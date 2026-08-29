package usersrepository

import (
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
)

type userRow struct {
	id                  uuid.UUID
	username            string
	email               string
	fullName            *string
	avatarURL           *string
	isActive            bool
	isSystem            bool
	isInternal          bool
	hasSeenWalkthrough  bool
	timezone            string
	workingDays         []int16
	workingStartMinute  *int16
	workingEndMinute    *int16
	lastLoginAt         *time.Time
	lastUsedWorkspaceID *uuid.UUID
	githubUsername      *string
	createdAt           time.Time
	updatedAt           time.Time
}

func toCoreUser(row userRow) usersdomain.User {
	return usersdomain.User{
		ID:                  row.id,
		Username:            row.username,
		Email:               row.email,
		FullName:            derefString(row.fullName),
		AvatarURL:           derefString(row.avatarURL),
		IsActive:            row.isActive,
		IsSystem:            row.isSystem,
		IsInternal:          row.isInternal,
		HasSeenWalkthrough:  row.hasSeenWalkthrough,
		Timezone:            row.timezone,
		WorkingDays:         int16sToInts(row.workingDays),
		WorkingStartMinute:  int16PointerToInt(row.workingStartMinute),
		WorkingEndMinute:    int16PointerToInt(row.workingEndMinute),
		LastLoginAt:         derefTime(row.lastLoginAt),
		LastUsedWorkspaceID: row.lastUsedWorkspaceID,
		GitHubUsername:      row.githubUsername,
		CreatedAt:           row.createdAt,
		UpdatedAt:           row.updatedAt,
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func int16sToInts(values []int16) []int {
	result := make([]int, len(values))
	for index, value := range values {
		result[index] = int(value)
	}
	return result
}

func int16PointerToInt(value *int16) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func optionalTime(present bool, value time.Time) *time.Time {
	if !present {
		return nil
	}
	return &value
}
