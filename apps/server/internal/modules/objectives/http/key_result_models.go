package objectiveshttp

import (
	"errors"
	"fmt"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

type AppNewObjective struct {
	Name         string            `json:"name" validate:"required"`
	Description  *string           `json:"description"`
	ShortSummary *string           `json:"shortSummary" validate:"omitempty,max=500"`
	LeadUser     *uuid.UUID        `json:"leadUser"`
	Team         uuid.UUID         `json:"teamId" validate:"required"`
	StartDate    *date.Date        `json:"startDate"`
	EndDate      *date.Date        `json:"endDate"`
	IsPrivate    bool              `json:"isPrivate"`
	Status       uuid.UUID         `json:"statusId" validate:"required"`
	Priority     *string           `json:"priority"`
	Color        string            `json:"color" validate:"omitempty,hexcolor"`
	KeyResults   []AppNewKeyResult `json:"keyResults,omitempty" validate:"max=20,dive"`
}

type AppNewKeyResult struct {
	Name            string      `json:"name" validate:"required"`
	MeasurementType string      `json:"measurementType" validate:"required,oneof=percentage number boolean"`
	StartValue      float64     `json:"startValue"`
	CurrentValue    float64     `json:"currentValue"`
	TargetValue     float64     `json:"targetValue"`
	Lead            *uuid.UUID  `json:"lead"`
	Contributors    []uuid.UUID `json:"contributors" validate:"max=100,dive,required"`
	StartDate       *date.Date  `json:"startDate" validate:"required"`
	EndDate         *date.Date  `json:"endDate" validate:"required"`
	CreatedBy       uuid.UUID   `json:"createdBy"`
}

const maxKeyResultBatchSize = 20

type AppCreateKeyResultsRequest struct {
	KeyResults []AppNewKeyResult `json:"keyResults" validate:"required,max=20,dive"`
}

func (request AppCreateKeyResultsRequest) Validate() error {
	if len(request.KeyResults) == 0 {
		return errors.New("at least one key result is required")
	}
	if len(request.KeyResults) > maxKeyResultBatchSize {
		return fmt.Errorf("a maximum of %d key results can be created at once", maxKeyResultBatchSize)
	}
	for index, keyResult := range request.KeyResults {
		if keyResult.StartDate == nil || keyResult.EndDate == nil {
			return fmt.Errorf("keyResults[%d] requires startDate and endDate", index)
		}
		if keyResult.EndDate.Time().Before(keyResult.StartDate.Time()) {
			return fmt.Errorf("keyResults[%d].endDate cannot be before startDate", index)
		}
	}
	return nil
}

func toCoreNewObjective(value AppNewObjective, createdBy uuid.UUID) objectives.CoreNewObjective {
	return objectives.CoreNewObjective{
		Name: value.Name, Description: value.Description, ShortSummary: value.ShortSummary,
		LeadUser: value.LeadUser, Team: value.Team, StartDate: value.StartDate.TimePtr(),
		EndDate: value.EndDate.TimePtr(), IsPrivate: value.IsPrivate, Status: value.Status,
		Priority: value.Priority, Color: value.Color, CreatedBy: createdBy,
	}
}

type AppKeyResult struct {
	ID              uuid.UUID   `json:"id"`
	ObjectiveID     uuid.UUID   `json:"objectiveId"`
	Name            string      `json:"name"`
	MeasurementType string      `json:"measurementType"`
	StartValue      float64     `json:"startValue"`
	CurrentValue    float64     `json:"currentValue"`
	TargetValue     float64     `json:"targetValue"`
	Lead            *uuid.UUID  `json:"lead"`
	Contributors    []uuid.UUID `json:"contributors"`
	StartDate       *time.Time  `json:"startDate"`
	EndDate         *time.Time  `json:"endDate"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	CreatedBy       uuid.UUID   `json:"createdBy"`
}

func toAppKeyResult(value keyresults.CoreKeyResult) AppKeyResult {
	return AppKeyResult{
		ID: value.ID, ObjectiveID: value.ObjectiveID, Name: value.Name,
		MeasurementType: value.MeasurementType, StartValue: value.StartValue,
		CurrentValue: value.CurrentValue, TargetValue: value.TargetValue,
		Lead: value.Lead, Contributors: value.Contributors, StartDate: value.StartDate,
		EndDate: value.EndDate, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
		CreatedBy: value.CreatedBy,
	}
}

func toAppKeyResults(values []keyresults.CoreKeyResult) []AppKeyResult {
	result := make([]AppKeyResult, len(values))
	for index, value := range values {
		result[index] = toAppKeyResult(value)
	}
	return result
}

func toAppObjectiveKeyResults(values []objectives.CoreKeyResult) []AppKeyResult {
	result := make([]AppKeyResult, len(values))
	for index, value := range values {
		result[index] = AppKeyResult{
			ID: value.ID, ObjectiveID: value.ObjectiveID, Name: value.Name,
			MeasurementType: value.MeasurementType, StartValue: value.StartValue,
			CurrentValue: value.CurrentValue, TargetValue: value.TargetValue,
			Lead: value.Lead, Contributors: value.Contributors, StartDate: value.StartDate,
			EndDate: value.EndDate, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
			CreatedBy: value.CreatedBy,
		}
	}
	return result
}
