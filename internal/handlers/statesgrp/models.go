package statesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/google/uuid"
)

// AppStatesList represents a state in the application layer.
type AppStatesList struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Category   string     `json:"category" validate:"oneof=backlog unstarted started paused completed cancelled"`
	OrderIndex int        `json:"orderIndex"`
	Workflow   uuid.UUID  `json:"workflowId"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
}

type NewState struct {
	Name     string    `json:"name" validate:"required"`
	Category string    `json:"category" validate:"required,oneof=backlog unstarted started paused completed cancelled"`
	Workflow uuid.UUID `json:"workflowId" validate:"required"`
}

type UpdateState struct {
	Name       *string `json:"name,omitempty"`
	OrderIndex *int    `json:"orderIndex,omitempty"`
}

func toAppState(s states.CoreState) AppStatesList {
	return AppStatesList{
		ID:         s.ID,
		Name:       s.Name,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Workflow:   s.Workflow,
		Workspace:  s.Workspace,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toAppStates(ss []states.CoreState) []AppStatesList {
	states := make([]AppStatesList, len(ss))
	for i, s := range ss {
		states[i] = toAppState(s)
	}
	return states
}

func toCoreNewState(ns NewState) states.CoreNewState {
	return states.CoreNewState{
		Name:     ns.Name,
		Category: ns.Category,
		Workflow: ns.Workflow,
	}
}

func toCoreUpdateState(us UpdateState) states.CoreUpdateState {
	return states.CoreUpdateState{
		Name:       us.Name,
		OrderIndex: us.OrderIndex,
	}
}
