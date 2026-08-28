package workspacesrepository

import (
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	"github.com/google/uuid"
)

type workspaceRecord struct {
	id          uuid.UUID
	slug        string
	name        string
	color       string
	teamSize    string
	avatarURL   *string
	createdBy   *uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
	trialEndsOn *time.Time
	deletedAt   *time.Time
	deletedBy   *uuid.UUID
}

type membershipWorkspaceRecord struct {
	workspaceRecord
	isActive bool
	userRole string
}

func workspaceFromRecord(record workspaceRecord) workspacedomain.Workspace {
	return workspacedomain.Workspace{
		ID:          record.id,
		Slug:        record.slug,
		Name:        record.name,
		Color:       record.color,
		TeamSize:    record.teamSize,
		AvatarURL:   record.avatarURL,
		CreatedBy:   record.createdBy,
		CreatedAt:   record.createdAt,
		UpdatedAt:   record.updatedAt,
		TrialEndsOn: record.trialEndsOn,
		DeletedAt:   record.deletedAt,
		DeletedBy:   record.deletedBy,
	}
}

func workspaceFromMembershipRecord(record membershipWorkspaceRecord) workspacedomain.Workspace {
	workspace := workspaceFromRecord(record.workspaceRecord)
	workspace.IsActive = record.isActive
	workspace.UserRole = record.userRole
	return workspace
}

func workspaceRecordFromCreateRow(row workspacesql.CreateWorkspaceRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromUpdateRow(row workspacesql.UpdateWorkspaceRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromIDRow(row workspacesql.GetWorkspaceByIDRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromPublicRow(row workspacesql.GetPublicWorkspaceBySlugRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromListRow(row workspacesql.ListWorkspacesForUserRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromMemberRow(row workspacesql.GetWorkspaceForMemberRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

func workspaceRecordFromMemberSlugRow(row workspacesql.GetWorkspaceForMemberBySlugRow) workspaceRecord {
	return workspaceRecord{id: row.WorkspaceID, slug: row.Slug, name: row.Name, color: row.Color, teamSize: row.TeamSize, avatarURL: row.AvatarURL, createdBy: row.CreatedBy, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, trialEndsOn: row.TrialEndsOn, deletedAt: row.DeletedAt, deletedBy: row.DeletedBy}
}

type workspaceSettingsRecord struct {
	workspaceID        uuid.UUID
	storyTerm          string
	sprintTerm         string
	objectiveTerm      string
	keyResultTerm      string
	objectiveEnabled   bool
	keyResultEnabled   bool
	workingDays        []int16
	workingStartMinute int16
	workingEndMinute   int16
	createdAt          time.Time
	updatedAt          time.Time
}

func workspaceSettingsFromRecord(record workspaceSettingsRecord) workspacedomain.WorkspaceSettings {
	workingDays := make([]int, len(record.workingDays))
	for index, day := range record.workingDays {
		workingDays[index] = int(day)
	}
	return workspacedomain.WorkspaceSettings{
		WorkspaceID:        record.workspaceID,
		StoryTerm:          record.storyTerm,
		SprintTerm:         record.sprintTerm,
		ObjectiveTerm:      record.objectiveTerm,
		KeyResultTerm:      record.keyResultTerm,
		ObjectiveEnabled:   record.objectiveEnabled,
		KeyResultEnabled:   record.keyResultEnabled,
		WorkingDays:        workingDays,
		WorkingStartMinute: int(record.workingStartMinute),
		WorkingEndMinute:   int(record.workingEndMinute),
		CreatedAt:          record.createdAt,
		UpdatedAt:          record.updatedAt,
	}
}

func workspaceSettingsRecordFromGet(row workspacesql.GetWorkspaceSettingsRow) workspaceSettingsRecord {
	return workspaceSettingsRecord{workspaceID: row.WorkspaceID, storyTerm: row.StoryTerm, sprintTerm: row.SprintTerm, objectiveTerm: row.ObjectiveTerm, keyResultTerm: row.KeyResultTerm, objectiveEnabled: row.ObjectiveEnabled, keyResultEnabled: row.KeyResultEnabled, workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute, workingEndMinute: row.WorkingEndMinute, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt}
}

func workspaceSettingsRecordFromGetOrCreate(row workspacesql.GetOrCreateWorkspaceSettingsRow) workspaceSettingsRecord {
	return workspaceSettingsRecord{workspaceID: row.WorkspaceID, storyTerm: row.StoryTerm, sprintTerm: row.SprintTerm, objectiveTerm: row.ObjectiveTerm, keyResultTerm: row.KeyResultTerm, objectiveEnabled: row.ObjectiveEnabled, keyResultEnabled: row.KeyResultEnabled, workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute, workingEndMinute: row.WorkingEndMinute, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt}
}

func workspaceSettingsRecordFromUpdate(row workspacesql.UpdateWorkspaceSettingsRow) workspaceSettingsRecord {
	return workspaceSettingsRecord{workspaceID: row.WorkspaceID, storyTerm: row.StoryTerm, sprintTerm: row.SprintTerm, objectiveTerm: row.ObjectiveTerm, keyResultTerm: row.KeyResultTerm, objectiveEnabled: row.ObjectiveEnabled, keyResultEnabled: row.KeyResultEnabled, workingDays: row.WorkingDays, workingStartMinute: row.WorkingStartMinute, workingEndMinute: row.WorkingEndMinute, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt}
}
