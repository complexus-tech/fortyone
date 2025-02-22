package workspacesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type AppWorkspace struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	Color     string    `json:"color"`
	UserRole  string    `json:"userRole"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAppWorkspace(workspace workspaces.CoreWorkspace) AppWorkspace {
	return AppWorkspace{
		ID:        workspace.ID,
		Slug:      workspace.Slug,
		Name:      workspace.Name,
		IsActive:  workspace.IsActive,
		Color:     workspace.Color,
		UserRole:  workspace.UserRole,
		CreatedAt: workspace.CreatedAt,
		UpdatedAt: workspace.UpdatedAt,
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

type AppSlugAvailability struct {
	Available bool   `json:"available"`
	Slug      string `json:"slug"`
}
