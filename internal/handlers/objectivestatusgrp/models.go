package objectivestatusgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/google/uuid"
)

// AppObjectiveStatusList represents an objective status in the application layer.
type AppObjectiveStatusList struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Category   string     `json:"category" validate:"oneof=backlog unstarted started paused completed cancelled"`
	OrderIndex int        `json:"orderIndex"`
	Team       uuid.UUID  `json:"teamId"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}

type NewObjectiveStatus struct {
	Name     string    `json:"name" validate:"required"`
	Category string    `json:"category" validate:"required,oneof=backlog unstarted started paused completed cancelled"`
	Team     uuid.UUID `json:"teamId" validate:"required"`
}

type UpdateObjectiveStatus struct {
	Name       *string `json:"name,omitempty"`
	OrderIndex *int    `json:"orderIndex,omitempty"`
}

func toAppObjectiveStatus(s objectivestatus.CoreObjectiveStatus) AppObjectiveStatusList {
	return AppObjectiveStatusList{
		ID:         s.ID,
		Name:       s.Name,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Team:       s.Team,
		Workspace:  s.Workspace,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toAppObjectiveStatuses(ss []objectivestatus.CoreObjectiveStatus) []AppObjectiveStatusList {
	statuses := make([]AppObjectiveStatusList, len(ss))
	for i, s := range ss {
		statuses[i] = toAppObjectiveStatus(s)
	}
	return statuses
}

func toCoreNewObjectiveStatus(ns NewObjectiveStatus) objectivestatus.CoreNewObjectiveStatus {
	return objectivestatus.CoreNewObjectiveStatus{
		Name:     ns.Name,
		Category: ns.Category,
		Team:     ns.Team,
	}
}

func toCoreUpdateObjectiveStatus(us UpdateObjectiveStatus) objectivestatus.CoreUpdateObjectiveStatus {
	return objectivestatus.CoreUpdateObjectiveStatus{
		Name:       us.Name,
		OrderIndex: us.OrderIndex,
	}
}
