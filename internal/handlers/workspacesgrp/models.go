package workspacesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type AppWorkspace struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toAppWorkspace(workspace workspaces.CoreWorkspace) AppWorkspace {
	return AppWorkspace{
		ID:          workspace.ID,
		Slug:        workspace.Slug,
		Name:        workspace.Name,
		Description: workspace.Description,
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
