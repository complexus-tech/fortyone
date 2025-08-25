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
	ID              uuid.UUID   `db:"id"`
	ObjectiveID     uuid.UUID   `db:"objective_id"`
	Name            string      `db:"name"`
	MeasurementType string      `db:"measurement_type"`
	StartValue      float64     `db:"start_value"`
	CurrentValue    float64     `db:"current_value"`
	TargetValue     float64     `db:"target_value"`
	Lead            *uuid.UUID  `db:"lead"`
	Contributors    []uuid.UUID `db:"-"` // Not stored directly in key_results table
	StartDate       *time.Time  `db:"start_date"`
	EndDate         *time.Time  `db:"end_date"`
	CreatedAt       time.Time   `db:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at"`
	CreatedBy       uuid.UUID   `db:"created_by"`
}

// CoreKeyResult represents the core business model for a key result
type CoreKeyResult struct {
	ID              uuid.UUID
	ObjectiveID     uuid.UUID
	Name            string
	MeasurementType string
	StartValue      float64
	CurrentValue    float64
	TargetValue     float64
	Lead            *uuid.UUID
	Contributors    []uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uuid.UUID
}

// CoreKeyResultFilters represents filtering options for key results
type CoreKeyResultFilters struct {
	ObjectiveIDs     []uuid.UUID `json:"objectiveIds"`
	TeamIDs          []uuid.UUID `json:"teamIds"`          // Filter by teams
	MeasurementTypes []string    `json:"measurementTypes"` // "percentage", "number", "boolean"
	CreatedBy        []uuid.UUID `json:"createdBy"`
	WorkspaceID      uuid.UUID   `json:"workspaceId"`
	CurrentUserID    uuid.UUID   `json:"currentUserId"` // For team membership filtering
	// Date range filters
	CreatedAfter  *time.Time `json:"createdAfter"`
	CreatedBefore *time.Time `json:"createdBefore"`
	UpdatedAfter  *time.Time `json:"updatedAfter"`
	UpdatedBefore *time.Time `json:"updatedBefore"`
	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	// Sorting
	OrderBy        string `json:"orderBy"`        // "name", "created_at", "updated_at", "objective_name"
	OrderDirection string `json:"orderDirection"` // "asc", "desc"
}

// CoreKeyResultWithObjective extends CoreKeyResult with objective info
type CoreKeyResultWithObjective struct {
	CoreKeyResult
	ObjectiveName string    `json:"objectiveName"`
	ObjectiveID   uuid.UUID `json:"objectiveId"`
	TeamID        uuid.UUID `json:"teamId"`
	TeamName      string    `json:"teamName"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
}

// CoreKeyResultListResponse represents paginated response
type CoreKeyResultListResponse struct {
	KeyResults []CoreKeyResultWithObjective `json:"keyResults"`
	TotalCount int                          `json:"totalCount"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"pageSize"`
	HasMore    bool                         `json:"hasMore"`
}

// dbKeyResultWithObjective represents the database model with objective info
type dbKeyResultWithObjective struct {
	dbKeyResult
	ObjectiveName string      `db:"objective_name"`
	TeamID        uuid.UUID   `db:"team_id"`
	TeamName      string      `db:"team_name"`
	WorkspaceID   uuid.UUID   `db:"workspace_id"`
	Contributors  []uuid.UUID `db:"-"` // Not stored directly in query result
}

// toCoreKeyResult converts a dbKeyResult to a CoreKeyResult
func toCoreKeyResult(kr dbKeyResult) CoreKeyResult {
	return CoreKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		Lead:            kr.Lead,
		Contributors:    kr.Contributors,
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
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

// toCoreKeyResultWithObjective converts a dbKeyResultWithObjective to a CoreKeyResultWithObjective
func toCoreKeyResultWithObjective(kr dbKeyResultWithObjective) CoreKeyResultWithObjective {
	coreKr := toCoreKeyResult(kr.dbKeyResult)
	coreKr.Contributors = kr.Contributors

	return CoreKeyResultWithObjective{
		CoreKeyResult: coreKr,
		ObjectiveName: kr.ObjectiveName,
		ObjectiveID:   kr.ObjectiveID,
		TeamID:        kr.TeamID,
		TeamName:      kr.TeamName,
		WorkspaceID:   kr.WorkspaceID,
	}
}

// toCoreKeyResultsWithObjective converts a slice of dbKeyResultWithObjective to a slice of CoreKeyResultWithObjective
func toCoreKeyResultsWithObjective(krs []dbKeyResultWithObjective) []CoreKeyResultWithObjective {
	result := make([]CoreKeyResultWithObjective, len(krs))
	for i, kr := range krs {
		result[i] = toCoreKeyResultWithObjective(kr)
	}
	return result
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
		Contributors:    kr.Contributors,
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
	}
}
