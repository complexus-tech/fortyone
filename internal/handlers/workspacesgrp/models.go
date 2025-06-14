package workspacesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type AppWorkspace struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	AvatarURL   *string    `json:"avatarUrl"`
	IsActive    bool       `json:"isActive"`
	Color       string     `json:"color"`
	UserRole    string     `json:"userRole"`
	TrialEndsOn *time.Time `json:"trialEndsOn"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func toAppWorkspace(workspace workspaces.CoreWorkspace) AppWorkspace {
	return AppWorkspace{
		ID:          workspace.ID,
		Slug:        workspace.Slug,
		Name:        workspace.Name,
		AvatarURL:   workspace.AvatarURL,
		IsActive:    workspace.IsActive,
		Color:       workspace.Color,
		UserRole:    workspace.UserRole,
		TrialEndsOn: workspace.TrialEndsOn,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

func toAppWorkspaces(workspaces []workspaces.CoreWorkspace) []AppWorkspace {
	appWorkspaces := make([]AppWorkspace, len(workspaces))
	for i, workspace := range workspaces {
		appWorkspaces[i] = toAppWorkspace(workspace)
	}
	return appWorkspaces
}

type AppNewWorkspace struct {
	Name     string `json:"name" validate:"required"`
	Slug     string `json:"slug" validate:"required,min=3,max=255"`
	TeamSize string `json:"teamSize" validate:"required"`
}

type AppUpdateWorkspace struct {
	Name string `json:"name,omitempty"`
}

type AppNewWorkspaceMember struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=member guest admin"`
}

type AppUpdateWorkspaceMemberRole struct {
	Role string `json:"role" validate:"required,oneof=member guest admin"`
}

type AppSlugAvailability struct {
	Available bool   `json:"available"`
	Slug      string `json:"slug"`
}

// AppWorkspaceTerminologySettings represents workspace terminology preferences for the API.
type AppWorkspaceTerminologySettings struct {
	StoryTerm     string    `json:"storyTerm"`
	SprintTerm    string    `json:"sprintTerm"`
	ObjectiveTerm string    `json:"objectiveTerm"`
	KeyResultTerm string    `json:"keyResultTerm"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AppWorkspaceSettings represents workspace settings for the API.
type AppWorkspaceSettings struct {
	StoryTerm        string    `json:"storyTerm"`
	SprintTerm       string    `json:"sprintTerm"`
	ObjectiveTerm    string    `json:"objectiveTerm"`
	KeyResultTerm    string    `json:"keyResultTerm"`
	SprintEnabled    bool      `json:"sprintEnabled"`
	ObjectiveEnabled bool      `json:"objectiveEnabled"`
	KeyResultEnabled bool      `json:"keyResultEnabled"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// AppUpdateWorkspaceSettings represents the payload for updating workspace settings.
type AppUpdateWorkspaceSettings struct {
	StoryTerm        string `json:"storyTerm,omitempty"`
	SprintTerm       string `json:"sprintTerm,omitempty"`
	ObjectiveTerm    string `json:"objectiveTerm,omitempty"`
	KeyResultTerm    string `json:"keyResultTerm,omitempty"`
	SprintEnabled    *bool  `json:"sprintEnabled,omitempty"`
	ObjectiveEnabled *bool  `json:"objectiveEnabled,omitempty"`
	KeyResultEnabled *bool  `json:"keyResultEnabled,omitempty"`
}

// toCoreWorkspaceSettings converts an API update model to a core model
func toCoreWorkspaceSettings(settings AppUpdateWorkspaceSettings, workspaceID uuid.UUID, current workspaces.CoreWorkspaceSettings) workspaces.CoreWorkspaceSettings {
	result := workspaces.CoreWorkspaceSettings{
		WorkspaceID:      workspaceID,
		StoryTerm:        settings.StoryTerm,
		SprintTerm:       settings.SprintTerm,
		ObjectiveTerm:    settings.ObjectiveTerm,
		KeyResultTerm:    settings.KeyResultTerm,
		SprintEnabled:    current.SprintEnabled,
		ObjectiveEnabled: current.ObjectiveEnabled,
		KeyResultEnabled: current.KeyResultEnabled,
	}

	// Only update boolean fields if they are provided
	if settings.SprintEnabled != nil {
		result.SprintEnabled = *settings.SprintEnabled
	}
	if settings.ObjectiveEnabled != nil {
		result.ObjectiveEnabled = *settings.ObjectiveEnabled
	}
	if settings.KeyResultEnabled != nil {
		result.KeyResultEnabled = *settings.KeyResultEnabled
	}

	return result
}

// toAppWorkspaceSettings converts a core settings model to an API model
func toAppWorkspaceSettings(settings workspaces.CoreWorkspaceSettings) AppWorkspaceSettings {
	return AppWorkspaceSettings{
		StoryTerm:        settings.StoryTerm,
		SprintTerm:       settings.SprintTerm,
		ObjectiveTerm:    settings.ObjectiveTerm,
		KeyResultTerm:    settings.KeyResultTerm,
		SprintEnabled:    settings.SprintEnabled,
		ObjectiveEnabled: settings.ObjectiveEnabled,
		KeyResultEnabled: settings.KeyResultEnabled,
		CreatedAt:        settings.CreatedAt,
		UpdatedAt:        settings.UpdatedAt,
	}
}
