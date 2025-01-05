package labels

import (
	"time"

	"github.com/google/uuid"
)

type CoreLabel struct {
	LabelID     uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	ProjectID   *uuid.UUID `json:"projectId"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CoreNewLabel struct {
	Name        string     `json:"name"`
	ProjectID   *uuid.UUID `json:"projectId"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Color       string     `json:"color"`
}
