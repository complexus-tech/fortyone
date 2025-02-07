package workspacesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type dbWorkspace struct {
	ID        uuid.UUID `db:"workspace_id"`
	Slug      string    `db:"slug"`
	Name      string    `db:"name"`
	Color     string    `db:"color"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// dbWorkspaceWithRole extends workspace data with user role information
type dbWorkspaceWithRole struct {
	ID        uuid.UUID `db:"workspace_id"`
	Slug      string    `db:"slug"`
	Name      string    `db:"name"`
	Color     string    `db:"color"`
	IsActive  bool      `db:"is_active"`
	UserRole  string    `db:"user_role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func toCoreWorkspace(p dbWorkspace) workspaces.CoreWorkspace {
	return workspaces.CoreWorkspace{
		ID:        p.ID,
		Slug:      p.Slug,
		Name:      p.Name,
		Color:     p.Color,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toCoreWorkspaceWithRole(p dbWorkspaceWithRole) workspaces.CoreWorkspace {
	return workspaces.CoreWorkspace{
		ID:        p.ID,
		Slug:      p.Slug,
		Name:      p.Name,
		Color:     p.Color,
		IsActive:  p.IsActive,
		UserRole:  p.UserRole,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toCoreWorkspaces(du []dbWorkspace) []workspaces.CoreWorkspace {
	coreWorkspaces := make([]workspaces.CoreWorkspace, len(du))
	for i, workspace := range du {
		coreWorkspaces[i] = toCoreWorkspace(workspace)
	}
	return coreWorkspaces
}

func toCoreWorkspacesWithRole(du []dbWorkspaceWithRole) []workspaces.CoreWorkspace {
	coreWorkspaces := make([]workspaces.CoreWorkspace, len(du))
	for i, workspace := range du {
		coreWorkspaces[i] = toCoreWorkspaceWithRole(workspace)
	}
	return coreWorkspaces
}
