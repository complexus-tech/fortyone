package projectsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/projects"
	"github.com/google/uuid"
)

type dbProject struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Owner       *uuid.UUID `db:"owner"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	IsPublic    bool       `db:"is_public"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func toCoreProject(p dbProject) projects.CoreProject {
	return projects.CoreProject{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Owner:       p.Owner,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		IsPublic:    p.IsPublic,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toCoreProjects(ps []dbProject) []projects.CoreProject {
	projects := make([]projects.CoreProject, len(ps))
	for i, p := range ps {
		projects[i] = toCoreProject(p)
	}
	return projects
}
