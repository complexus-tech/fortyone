package domain

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaximumBatchSize = 20
	DefaultPageSize  = 20
	MaximumPageSize  = 100
)

type MeasurementType string

const (
	MeasurementPercentage MeasurementType = "percentage"
	MeasurementNumber     MeasurementType = "number"
	MeasurementBoolean    MeasurementType = "boolean"
)

func (measurement MeasurementType) Valid() bool {
	switch measurement {
	case MeasurementPercentage, MeasurementNumber, MeasurementBoolean:
		return true
	default:
		return false
	}
}

type NewKeyResult struct {
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

func (draft NewKeyResult) Normalize() (NewKeyResult, error) {
	draft.Name = strings.TrimSpace(draft.Name)
	measurement := MeasurementType(strings.TrimSpace(draft.MeasurementType))
	if draft.ObjectiveID == uuid.Nil || draft.CreatedBy == uuid.Nil || draft.Name == "" {
		return NewKeyResult{}, fmt.Errorf("%w: objective, creator, and name are required", ErrInvalid)
	}
	if len([]rune(draft.Name)) > 255 {
		return NewKeyResult{}, fmt.Errorf("%w: name cannot exceed 255 characters", ErrInvalid)
	}
	if !measurement.Valid() {
		return NewKeyResult{}, fmt.Errorf("%w: unsupported measurement type", ErrInvalid)
	}
	if !finite(draft.StartValue) || !finite(draft.CurrentValue) || !finite(draft.TargetValue) {
		return NewKeyResult{}, fmt.Errorf("%w: measurement values must be finite", ErrInvalid)
	}
	if draft.StartDate == nil || draft.EndDate == nil {
		return NewKeyResult{}, fmt.Errorf("%w: start and end dates are required", ErrInvalid)
	}
	startDate := normalizeDate(*draft.StartDate)
	endDate := normalizeDate(*draft.EndDate)
	if endDate.Before(startDate) {
		return NewKeyResult{}, fmt.Errorf("%w: end date cannot be before start date", ErrInvalid)
	}
	if draft.Lead != nil && *draft.Lead == uuid.Nil {
		return NewKeyResult{}, fmt.Errorf("%w: lead cannot be a zero id", ErrInvalid)
	}
	contributors, err := normalizeUUIDs(draft.Contributors)
	if err != nil {
		return NewKeyResult{}, err
	}
	draft.MeasurementType = string(measurement)
	draft.StartDate = &startDate
	draft.EndDate = &endDate
	draft.Contributors = contributors
	return draft, nil
}

type KeyResult struct {
	ID              uuid.UUID
	SequenceID      int
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

type KeyResultWithObjective struct {
	KeyResult
	ObjectiveName string
	ObjectiveID   uuid.UUID
	TeamID        uuid.UUID
	TeamName      string
	TeamCode      string
	WorkspaceID   uuid.UUID
}

type ListResponse struct {
	KeyResults []KeyResultWithObjective
	TotalCount int
	Page       int
	PageSize   int
	HasMore    bool
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func normalizeDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func normalizeUUIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return nil, fmt.Errorf("%w: contributor cannot be a zero id", ErrInvalid)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}
