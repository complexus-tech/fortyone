package workspacesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/google/uuid"
)

type dbWorkspace struct {
	ID          uuid.UUID  `db:"workspace_id"`
	Slug        string     `db:"slug"`
	Name        string     `db:"name"`
	Color       string     `db:"color"`
	AvatarURL   *string    `db:"avatar_url"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	TrialEndsOn *time.Time `db:"trial_ends_on"`
	DeletedAt   *time.Time `db:"deleted_at"`
	DeletedBy   *uuid.UUID `db:"deleted_by"`
}

// dbWorkspaceWithRole extends workspace data with user role information
type dbWorkspaceWithRole struct {
	ID          uuid.UUID  `db:"workspace_id"`
	Slug        string     `db:"slug"`
	Name        string     `db:"name"`
	Color       string     `db:"color"`
	AvatarURL   *string    `db:"avatar_url"`
	IsActive    bool       `db:"is_active"`
	UserRole    string     `db:"user_role"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	TrialEndsOn *time.Time `db:"trial_ends_on"`
	DeletedAt   *time.Time `db:"deleted_at"`
	DeletedBy   *uuid.UUID `db:"deleted_by"`
}

// dbWorkspaceSettings represents workspace settings in the database.
type dbWorkspaceSettings struct {
	WorkspaceID      uuid.UUID `db:"workspace_id"`
	StoryTerm        string    `db:"story_term"`
	SprintTerm       string    `db:"sprint_term"`
	ObjectiveTerm    string    `db:"objective_term"`
	KeyResultTerm    string    `db:"key_result_term"`
	ObjectiveEnabled bool      `db:"objective_enabled"`
	KeyResultEnabled bool      `db:"key_result_enabled"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func toCoreWorkspace(p dbWorkspace) workspaces.CoreWorkspace {
	return workspaces.CoreWorkspace{
		ID:          p.ID,
		Slug:        p.Slug,
		Name:        p.Name,
		Color:       p.Color,
		AvatarURL:   p.AvatarURL,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		TrialEndsOn: p.TrialEndsOn,
		DeletedAt:   p.DeletedAt,
		DeletedBy:   p.DeletedBy,
	}
}

func toCoreWorkspaceWithRole(p dbWorkspaceWithRole) workspaces.CoreWorkspace {
	return workspaces.CoreWorkspace{
		ID:          p.ID,
		Slug:        p.Slug,
		Name:        p.Name,
		Color:       p.Color,
		AvatarURL:   p.AvatarURL,
		IsActive:    p.IsActive,
		UserRole:    p.UserRole,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		TrialEndsOn: p.TrialEndsOn,
		DeletedAt:   p.DeletedAt,
		DeletedBy:   p.DeletedBy,
	}
}

func toCoreWorkspacesWithRole(du []dbWorkspaceWithRole) []workspaces.CoreWorkspace {
	coreWorkspaces := make([]workspaces.CoreWorkspace, len(du))
	for i, workspace := range du {
		coreWorkspaces[i] = toCoreWorkspaceWithRole(workspace)
	}
	return coreWorkspaces
}

// toCoreWorkspaceSettings converts a database workspace settings to core workspace settings.
func toCoreWorkspaceSettings(s dbWorkspaceSettings) workspaces.CoreWorkspaceSettings {
	return workspaces.CoreWorkspaceSettings{
		WorkspaceID:      s.WorkspaceID,
		StoryTerm:        s.StoryTerm,
		SprintTerm:       s.SprintTerm,
		ObjectiveTerm:    s.ObjectiveTerm,
		KeyResultTerm:    s.KeyResultTerm,
		ObjectiveEnabled: s.ObjectiveEnabled,
		KeyResultEnabled: s.KeyResultEnabled,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
