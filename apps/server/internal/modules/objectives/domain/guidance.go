package domain

import (
	"time"

	"github.com/google/uuid"
)

// OverdueGuidanceCursor identifies the last recipient processed by the
// objective-guidance worker. The repository orders recipients by lead and
// workspace, making this cursor stable while the eligible set changes.
type OverdueGuidanceCursor struct {
	LeadUserID  uuid.UUID
	WorkspaceID uuid.UUID
}

// OverdueGuidanceRecipient is the persistence-neutral recipient context needed
// to load and deliver objective guidance.
type OverdueGuidanceRecipient struct {
	LeadUserID    uuid.UUID
	LeadEmail     string
	LeadName      string
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceSlug string
	EmailEnabled  bool
}

// OverdueGuidanceObjective is an objective or key-result deadline signal sent
// to an eligible objective lead.
type OverdueGuidanceObjective struct {
	ID             uuid.UUID
	Name           string
	EndDate        time.Time
	LeadUserID     uuid.UUID
	LeadEmail      string
	LeadName       string
	WorkspaceID    uuid.UUID
	WorkspaceName  string
	WorkspaceSlug  string
	TeamID         uuid.UUID
	DeadlineStatus string
	DaysDifference int
	KeyResults     string
}

// OverdueGuidanceKeyResult is the typed representation embedded in an
// objective-guidance signal's JSON key-result collection.
type OverdueGuidanceKeyResult struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	EndDate         string    `json:"end_date"`
	MeasurementType string    `json:"measurement_type"`
	StartValue      float64   `json:"start_value"`
	CurrentValue    float64   `json:"current_value"`
	TargetValue     float64   `json:"target_value"`
	IsCompleted     bool      `json:"is_completed"`
	DeadlineStatus  string    `json:"deadline_status"`
	DaysDifference  int       `json:"days_difference"`
}
