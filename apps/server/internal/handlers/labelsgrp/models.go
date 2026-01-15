package labelsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/labels"
	"github.com/google/uuid"
)

type AppLabel struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AppNewLabel struct {
	Name   string     `json:"name"`
	TeamID *uuid.UUID `json:"teamId"`
	Color  string     `json:"color"`
}

type AppUpdateLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AppFilters struct {
	Team *uuid.UUID `json:"teamId" db:"team_id"`
}

func toAppLabel(label labels.CoreLabel) AppLabel {
	return AppLabel{
		ID:          label.LabelID,
		Name:        label.Name,
		TeamID:      label.TeamID,
		WorkspaceID: label.WorkspaceID,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}

func toAppLabels(labels []labels.CoreLabel) []AppLabel {
	appLabels := make([]AppLabel, len(labels))
	for i, label := range labels {
		appLabels[i] = toAppLabel(label)
	}
	return appLabels
}
