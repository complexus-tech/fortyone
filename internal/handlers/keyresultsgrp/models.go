package keyresultsgrp

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AppKeyResult represents a key result in the application
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

// AppNewKeyResult represents the data needed to create a new key result
type AppNewKeyResult struct {
	ObjectiveID     uuid.UUID   `json:"objectiveId" validate:"required"`
	Name            string      `json:"name" validate:"required"`
	MeasurementType string      `json:"measurementType" validate:"required,oneof=percentage number boolean"`
	StartValue      float64     `json:"startValue"`
	CurrentValue    float64     `json:"currentValue"`
	TargetValue     float64     `json:"targetValue"`
	Lead            *uuid.UUID  `json:"lead,omitempty"`
	Contributors    []uuid.UUID `json:"contributors,omitempty"`
	StartDate       *time.Time  `json:"startDate" validate:"required"`
	EndDate         *time.Time  `json:"endDate" validate:"required"`
}

// AppUpdateKeyResult represents the data needed to update a key result
type AppUpdateKeyResult struct {
	Name            string       `json:"name" db:"name"`
	MeasurementType string       `json:"measurementType" db:"measurement_type" validate:"omitempty,oneof=percentage number boolean"`
	StartValue      *float64     `json:"startValue" db:"start_value"`
	CurrentValue    *float64     `json:"currentValue" db:"current_value"`
	TargetValue     *float64     `json:"targetValue" db:"target_value"`
	Lead            *uuid.UUID   `json:"lead,omitempty" db:"lead"`
	Contributors    *[]uuid.UUID `json:"contributors,omitempty" db:"-"` // Not directly updatable via this struct
	StartDate       *time.Time   `json:"startDate" db:"start_date" validate:"required"`
	EndDate         *time.Time   `json:"endDate" db:"end_date" validate:"required"`
}

// AppKeyResultWithObjective extends AppKeyResult with objective info
type AppKeyResultWithObjective struct {
	AppKeyResult
	ObjectiveName string    `json:"objectiveName"`
	ObjectiveID   uuid.UUID `json:"objectiveId"`
	TeamID        uuid.UUID `json:"teamId"`
	TeamName      string    `json:"teamName"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
}

// AppKeyResultListResponse represents paginated response
type AppKeyResultListResponse struct {
	KeyResults []AppKeyResultWithObjective `json:"keyResults"`
	TotalCount int                         `json:"totalCount"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"pageSize"`
	HasMore    bool                        `json:"hasMore"`
}

// AppKeyResultFilters represents filtering options
type AppKeyResultFilters struct {
	ObjectiveIDs     []uuid.UUID `json:"objectiveIds"`
	TeamIDs          []uuid.UUID `json:"teamIds"`
	MeasurementTypes []string    `json:"measurementTypes"`
	CreatedAfter     *time.Time  `json:"createdAfter"`
	CreatedBefore    *time.Time  `json:"createdBefore"`
	UpdatedAfter     *time.Time  `json:"updatedAfter"`
	UpdatedBefore    *time.Time  `json:"updatedBefore"`
	Page             int         `json:"page"`
	PageSize         int         `json:"pageSize"`
	OrderBy          string      `json:"orderBy"`
	OrderDirection   string      `json:"orderDirection"`
}

// Validate validates the AppNewKeyResult struct
func (a AppNewKeyResult) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(a)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				fieldName := getJSONTagName(reflect.TypeOf(a), e.Field())
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
				case "oneof":
					options := strings.Split(e.Param(), " ")
					formattedOptions := formatOptions(options)
					errorMessages = append(errorMessages, fmt.Sprintf("%s should be one of: %s", fieldName, formattedOptions))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", fieldName, e.Tag()))
				}
			}
			return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
		}
	}

	return nil
}

// formatOptions formats the options for error messages
func formatOptions(options []string) string {
	if len(options) == 0 {
		return ""
	}
	if len(options) == 1 {
		return options[0]
	}
	return fmt.Sprintf("%s or %s", strings.Join(options[:len(options)-1], ", "), options[len(options)-1])
}

// getJSONTagName gets the JSON tag name for a field
func getJSONTagName(t reflect.Type, fieldName string) string {
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	tag := field.Tag.Get("json")
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return fieldName
	}
	return name
}

// Validate validates the AppUpdateKeyResult struct
func (a AppUpdateKeyResult) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(a)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				fieldName := getJSONTagName(reflect.TypeOf(a), e.Field())
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
				case "oneof":
					options := strings.Split(e.Param(), " ")
					formattedOptions := formatOptions(options)
					errorMessages = append(errorMessages, fmt.Sprintf("%s should be one of: %s", fieldName, formattedOptions))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", fieldName, e.Tag()))
				}
			}
			return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
		}
	}

	return nil
}

// toAppKeyResult converts a CoreKeyResult to an AppKeyResult
func toAppKeyResult(kr keyresults.CoreKeyResult) AppKeyResult {
	return AppKeyResult{
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

// toAppKeyResults converts a slice of CoreKeyResult to a slice of AppKeyResult
func toAppKeyResults(krs []keyresults.CoreKeyResult) []AppKeyResult {
	result := make([]AppKeyResult, len(krs))
	for i, kr := range krs {
		result[i] = toAppKeyResult(kr)
	}
	return result
}

// toAppKeyResultWithObjective converts a CoreKeyResultWithObjective to an AppKeyResultWithObjective
func toAppKeyResultWithObjective(kr keyresultsrepo.CoreKeyResultWithObjective) AppKeyResultWithObjective {
	return AppKeyResultWithObjective{
		AppKeyResult:  toAppKeyResult(keyresults.CoreKeyResult(kr.CoreKeyResult)),
		ObjectiveName: kr.ObjectiveName,
		ObjectiveID:   kr.ObjectiveID,
		TeamID:        kr.TeamID,
		TeamName:      kr.TeamName,
		WorkspaceID:   kr.WorkspaceID,
	}
}

// toAppKeyResultListResponse converts a CoreKeyResultListResponse to an AppKeyResultListResponse
func toAppKeyResultListResponse(response keyresultsrepo.CoreKeyResultListResponse) AppKeyResultListResponse {
	keyResults := make([]AppKeyResultWithObjective, len(response.KeyResults))
	for i, kr := range response.KeyResults {
		keyResults[i] = toAppKeyResultWithObjective(kr)
	}

	return AppKeyResultListResponse{
		KeyResults: keyResults,
		TotalCount: response.TotalCount,
		Page:       response.Page,
		PageSize:   response.PageSize,
		HasMore:    response.HasMore,
	}
}

// toCoreNewKeyResult converts an AppNewKeyResult to a CoreNewKeyResult
func toCoreNewKeyResult(nkr AppNewKeyResult, userID uuid.UUID) keyresults.CoreNewKeyResult {
	return keyresults.CoreNewKeyResult{
		ObjectiveID:     nkr.ObjectiveID,
		Name:            nkr.Name,
		MeasurementType: nkr.MeasurementType,
		StartValue:      nkr.StartValue,
		CurrentValue:    nkr.CurrentValue,
		TargetValue:     nkr.TargetValue,
		Lead:            nkr.Lead,
		Contributors:    nkr.Contributors,
		StartDate:       nkr.StartDate,
		EndDate:         nkr.EndDate,
		CreatedBy:       userID,
	}
}
