package keyresultsrepository

import (
	"encoding/json"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
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
	ID              uuid.UUID        `db:"id"`
	ObjectiveID     uuid.UUID        `db:"objective_id"`
	Name            string           `db:"name"`
	MeasurementType string           `db:"measurement_type"`
	StartValue      float64          `db:"start_value"`
	CurrentValue    float64          `db:"current_value"`
	TargetValue     float64          `db:"target_value"`
	Lead            *uuid.UUID       `db:"lead"`
	Contributors    *json.RawMessage `db:"contributors"`
	StartDate       *time.Time       `db:"start_date"`
	EndDate         *time.Time       `db:"end_date"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
	CreatedBy       uuid.UUID        `db:"created_by"`
}

// Core key result types are owned by the service layer.
type CoreKeyResult = keyresults.CoreKeyResult
type CoreKeyResultFilters = keyresults.CoreKeyResultFilters
type CoreKeyResultWithObjective = keyresults.CoreKeyResultWithObjective
type CoreKeyResultListResponse = keyresults.CoreKeyResultListResponse

// dbKeyResultWithObjective represents the database model with objective info
type dbKeyResultWithObjective struct {
	dbKeyResult
	ObjectiveName string    `db:"objective_name"`
	TeamID        uuid.UUID `db:"team_id"`
	TeamName      string    `db:"team_name"`
	WorkspaceID   uuid.UUID `db:"workspace_id"`
}

// toCoreKeyResult converts a dbKeyResult to a CoreKeyResult
func toCoreKeyResult(kr dbKeyResult) CoreKeyResult {
	var contributors []uuid.UUID
	if kr.Contributors != nil {
		if err := json.Unmarshal(*kr.Contributors, &contributors); err != nil {
			contributors = []uuid.UUID{}
		}
	} else {
		contributors = []uuid.UUID{}
	}

	return CoreKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		Lead:            kr.Lead,
		Contributors:    contributors,
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
	}
}

// toCoreKeyResultWithObjective converts a dbKeyResultWithObjective to a CoreKeyResultWithObjective
func toCoreKeyResultWithObjective(kr dbKeyResultWithObjective) CoreKeyResultWithObjective {
	coreKr := toCoreKeyResult(kr.dbKeyResult)

	return CoreKeyResultWithObjective{
		CoreKeyResult: coreKr,
		ObjectiveName: kr.ObjectiveName,
		ObjectiveID:   kr.ObjectiveID,
		TeamID:        kr.TeamID,
		TeamName:      kr.TeamName,
		WorkspaceID:   kr.WorkspaceID,
	}
}

// toDBKeyResult converts a CoreKeyResult to a dbKeyResult
func toDBKeyResult(kr CoreKeyResult) dbKeyResult {
	return dbKeyResult{
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		Lead:            kr.Lead,
		Contributors:    nil, // Not populated when converting to DB model
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
	}
}
