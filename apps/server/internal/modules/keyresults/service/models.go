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
	Lead            *uuid.UUID
	Contributors    []uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedBy       uuid.UUID
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
	Lead            *uuid.UUID
	Contributors    []uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uuid.UUID
}

// CoreKeyResultFilters represents filtering options for key results.
type CoreKeyResultFilters struct {
	ObjectiveIDs     []uuid.UUID `json:"objectiveIds"`
	TeamIDs          []uuid.UUID `json:"teamIds"`
	MeasurementTypes []string    `json:"measurementTypes"`
	CreatedBy        []uuid.UUID `json:"createdBy"`
	WorkspaceID      uuid.UUID   `json:"workspaceId"`
	CurrentUserID    uuid.UUID   `json:"currentUserId"`
	CreatedAfter     *time.Time  `json:"createdAfter"`
	CreatedBefore    *time.Time  `json:"createdBefore"`
	UpdatedAfter     *time.Time  `json:"updatedAfter"`
	UpdatedBefore    *time.Time  `json:"updatedBefore"`
	Page             int         `json:"page"`
	PageSize         int         `json:"pageSize"`
	OrderBy          string      `json:"orderBy"`
	OrderDirection   string      `json:"orderDirection"`
}

// CoreKeyResultWithObjective extends CoreKeyResult with objective data.
type CoreKeyResultWithObjective struct {
	CoreKeyResult
	ObjectiveName string    `json:"objectiveName"`
	ObjectiveID   uuid.UUID `json:"objectiveId"`
	TeamID        uuid.UUID `json:"teamId"`
	TeamName      string    `json:"teamName"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
}

// CoreKeyResultListResponse represents paginated key result response.
type CoreKeyResultListResponse struct {
	KeyResults []CoreKeyResultWithObjective `json:"keyResults"`
	TotalCount int                          `json:"totalCount"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"pageSize"`
	HasMore    bool                         `json:"hasMore"`
}
