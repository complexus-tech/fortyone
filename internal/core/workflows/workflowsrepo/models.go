package workflowsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workflows"
	"github.com/google/uuid"
)

type dbWorkflow struct {
	ID        uuid.UUID  `db:"workflow_id"`
	Name      string     `db:"name"`
	Team      uuid.UUID  `db:"team_id"`
	Workspace uuid.UUID  `db:"workspace_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func toCoreWorkflow(w dbWorkflow) workflows.CoreWorkflow {
	return workflows.CoreWorkflow{
		ID:        w.ID,
		Name:      w.Name,
		Team:      w.Team,
		Workspace: w.Workspace,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func toCoreWorkflows(ws []dbWorkflow) []workflows.CoreWorkflow {
	result := make([]workflows.CoreWorkflow, len(ws))
	for i, w := range ws {
		result[i] = toCoreWorkflow(w)
	}
	return result
}
