package workspacesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type dbWorkspace struct {
	ID          uuid.UUID `db:"workspace_id"`
	Slug        string    `db:"slug"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreWorkspace(p dbWorkspace) workspaces.CoreWorkspace {
	return workspaces.CoreWorkspace{
		ID:          p.ID,
		Slug:        p.Slug,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toCoreWorkspaces(du []dbWorkspace) []workspaces.CoreWorkspace {
	coreWorkspaces := make([]workspaces.CoreWorkspace, len(du))
	for i, workspace := range du {
		coreWorkspaces[i] = toCoreWorkspace(workspace)
	}
	return coreWorkspaces
}
