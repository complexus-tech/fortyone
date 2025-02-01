package keyresults

import (
	"time"

	"github.com/google/uuid"
)

// CoreNewKeyResult represents the data needed to create a new key result
type CoreNewKeyResult struct {
	ObjectiveID     uuid.UUID
	Name            string
	MeasurementType string
	StartValue      float64
	CurrentValue    float64
	TargetValue     float64
}

// CoreKeyResult represents a key result in the system
type CoreKeyResult struct {
	ID              uuid.UUID
	ObjectiveID     uuid.UUID
	Name            string
	MeasurementType string
	StartValue      float64
	CurrentValue    float64
	TargetValue     float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
