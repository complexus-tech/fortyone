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
	Category   string     `json:"category"`
	OrderIndex int        `json:"orderIndex"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	IsDefault  bool       `json:"isDefault"`
	Color      string     `json:"color"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}

type NewObjectiveStatus struct {
	Name      string `json:"name" validate:"required"`
	Category  string `json:"category" validate:"required,oneof=unstarted started paused completed cancelled"`
	IsDefault bool   `json:"isDefault"`
	Color     string `json:"color" validate:"required"`
}

type UpdateObjectiveStatus struct {
	Name       *string `json:"name,omitempty"`
	OrderIndex *int    `json:"orderIndex,omitempty"`
	IsDefault  *bool   `json:"isDefault,omitempty"`
	Color      *string `json:"color,omitempty"`
}

func toAppObjectiveStatus(s objectivestatus.CoreObjectiveStatus) AppObjectiveStatusList {
	return AppObjectiveStatusList{
		ID:         s.ID,
		Name:       s.Name,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Workspace:  s.Workspace,
		IsDefault:  s.IsDefault,
		Color:      s.Color,
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
		Name:      ns.Name,
		Category:  ns.Category,
		IsDefault: ns.IsDefault,
		Color:     ns.Color,
	}
}

func toCoreUpdateObjectiveStatus(us UpdateObjectiveStatus) objectivestatus.CoreUpdateObjectiveStatus {
	return objectivestatus.CoreUpdateObjectiveStatus{
		Name:       us.Name,
		OrderIndex: us.OrderIndex,
		IsDefault:  us.IsDefault,
		Color:      us.Color,
	}
}
