package labelsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/labels"
	"github.com/google/uuid"
)

type dbLabel struct {
	ID          uuid.UUID  `db:"label_id"`
	Name        string     `db:"name"`
	TeamID      *uuid.UUID `db:"team_id"`
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	Color       string     `db:"color"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type dbNewLabel struct {
	Name        string     `db:"name"`
	TeamID      *uuid.UUID `db:"team_id"`
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	Color       string     `db:"color"`
}

func toCoreLabel(label dbLabel) labels.CoreLabel {
	return labels.CoreLabel{
		LabelID:     label.ID,
		Name:        label.Name,
		TeamID:      label.TeamID,
		WorkspaceID: label.WorkspaceID,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}

func toCoreLabels(lbs []dbLabel) []labels.CoreLabel {
	coreLabels := make([]labels.CoreLabel, len(lbs))
	for i, label := range lbs {
		coreLabels[i] = toCoreLabel(label)
	}
	return coreLabels
}

func toDbNewLabel(cnl labels.CoreNewLabel) dbNewLabel {
	return dbNewLabel{
		Name:        cnl.Name,
		TeamID:      cnl.TeamID,
		WorkspaceID: cnl.WorkspaceID,
		Color:       cnl.Color,
	}
}
