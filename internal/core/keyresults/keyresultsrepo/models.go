package keyresultsrepo

import (
	"time"

	"github.com/google/uuid"
)

// MeasurementType represents the type of measurement for a key result
type MeasurementType string

const (
	MeasurementTypePercentage MeasurementType = "percentage"
	MeasurementTypeNumber     MeasurementType = "number"
	MeasurementTypeBoolean    MeasurementType = "boolean"
)

// dbKeyResult represents the database model for a key result
type dbKeyResult struct {
	ID              uuid.UUID `db:"id"`
	ObjectiveID     uuid.UUID `db:"objective_id"`
	Name            string    `db:"name"`
	MeasurementType string    `db:"measurement_type"`
	StartValue      float64   `db:"start_value"`
	TargetValue     float64   `db:"target_value"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// CoreKeyResult represents the core business model for a key result
type CoreKeyResult struct {
	ID              uuid.UUID
	ObjectiveID     uuid.UUID
	Name            string
	MeasurementType string
	StartValue      float64
	TargetValue     float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// toCoreKeyResult converts a dbKeyResult to a CoreKeyResult
func toCoreKeyResult(kr dbKeyResult) CoreKeyResult {
	return CoreKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		TargetValue:     kr.TargetValue,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
	}
}

// toCoreKeyResults converts a slice of dbKeyResult to a slice of CoreKeyResult
func toCoreKeyResults(krs []dbKeyResult) []CoreKeyResult {
	result := make([]CoreKeyResult, len(krs))
	for i, kr := range krs {
		result[i] = toCoreKeyResult(kr)
	}
	return result
}

// toDBKeyResult converts a CoreKeyResult to a dbKeyResult
func toDBKeyResult(kr CoreKeyResult) dbKeyResult {
	return dbKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		TargetValue:     kr.TargetValue,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
	}
}
