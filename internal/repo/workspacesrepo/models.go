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

// dbWorkspaceTerminology represents workspace terminology preferences in the database.
type dbWorkspaceTerminology struct {
	WorkspaceID   uuid.UUID `db:"workspace_id"`
	StoryTerm     string    `db:"story_term"`
	SprintTerm    string    `db:"sprint_term"`
	ObjectiveTerm string    `db:"objective_term"`
	KeyResultTerm string    `db:"key_result_term"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
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

func toCoreWorkspacesWithRole(du []dbWorkspaceWithRole) []workspaces.CoreWorkspace {
	coreWorkspaces := make([]workspaces.CoreWorkspace, len(du))
	for i, workspace := range du {
		coreWorkspaces[i] = toCoreWorkspaceWithRole(workspace)
	}
	return coreWorkspaces
}

// toCoreWorkspaceTerminology converts a database workspace terminology to a core workspace terminology.
func toCoreWorkspaceTerminology(t dbWorkspaceTerminology) workspaces.CoreWorkspaceTerminology {
	return workspaces.CoreWorkspaceTerminology{
		WorkspaceID:   t.WorkspaceID,
		StoryTerm:     t.StoryTerm,
		SprintTerm:    t.SprintTerm,
		ObjectiveTerm: t.ObjectiveTerm,
		KeyResultTerm: t.KeyResultTerm,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}
