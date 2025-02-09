package workflowsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/workflows"
	"github.com/google/uuid"
)

type AppWorkflow struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Team      uuid.UUID `json:"teamId"`
	Workspace uuid.UUID `json:"workspaceId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NewWorkflow struct {
	Name string    `json:"name" validate:"required"`
	Team uuid.UUID `json:"teamId" validate:"required"`
}

type UpdateWorkflow struct {
	Name *string `json:"name,omitempty"`
}

func toAppWorkflow(w workflows.CoreWorkflow) AppWorkflow {
	return AppWorkflow{
		ID:        w.ID,
		Name:      w.Name,
		Team:      w.Team,
		Workspace: w.Workspace,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func toAppWorkflows(ws []workflows.CoreWorkflow) []AppWorkflow {
	result := make([]AppWorkflow, len(ws))
	for i, w := range ws {
		result[i] = toAppWorkflow(w)
	}
	return result
}

func toCoreNewWorkflow(nw NewWorkflow) workflows.CoreNewWorkflow {
	return workflows.CoreNewWorkflow{
		Name: nw.Name,
		Team: nw.Team,
	}
}

func toCoreUpdateWorkflow(uw UpdateWorkflow) workflows.CoreUpdateWorkflow {
	return workflows.CoreUpdateWorkflow{
		Name: uw.Name,
	}
}
