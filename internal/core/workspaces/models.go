package workspaces

import (
	"time"

	"github.com/google/uuid"
)

// CoreWorkspace represents a workspace in the application layer.
type CoreWorkspace struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	Color     string
	TeamSize  string
	IsActive  bool
	UserRole  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
