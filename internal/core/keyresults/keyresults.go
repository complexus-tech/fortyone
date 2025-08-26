package keyresults

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Set of error variables for key result operations
var (
	ErrNotFound = errors.New("key result not found")
)

// Service manages the key result operations
type Service struct {
	repo          keyresultsrepo.Repository
	okrActivities *okractivities.Service
	log           *logger.Logger
}

// New creates a new key result service
func New(log *logger.Logger, repo keyresultsrepo.Repository, okrActivities *okractivities.Service) *Service {
	return &Service{
		repo:          repo,
		okrActivities: okrActivities,
		log:           log,
	}
}

// Create inserts a new key result into the system
func (s *Service) Create(ctx context.Context, nkr CoreNewKeyResult, workspaceID uuid.UUID) (CoreKeyResult, error) {

	kr := keyresultsrepo.CoreKeyResult{
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
		CreatedBy:       nkr.CreatedBy,
	}

	id, err := s.repo.Create(ctx, &kr)
	if err != nil {
		return CoreKeyResult{}, fmt.Errorf("create: %w", err)
	}

	// Add contributors if provided
	if len(nkr.Contributors) > 0 {
		if err := s.repo.AddContributors(ctx, id, nkr.Contributors); err != nil {
			return CoreKeyResult{}, fmt.Errorf("failed to add contributors: %w", err)
		}
	}

	// Record the create activity
	activity := okractivities.CoreNewActivity{
		ObjectiveID:  kr.ObjectiveID,
		KeyResultID:  &id,
		UserID:       nkr.CreatedBy,
		Type:         okractivities.ActivityTypeCreate,
		UpdateType:   okractivities.UpdateTypeKeyResult,
		Field:        "all",
		CurrentValue: kr.Name,
		Comment:      "",
		WorkspaceID:  workspaceID,
	}

	if err := s.okrActivities.Create(ctx, activity); err != nil {
		s.log.Error(ctx, "failed to record key result create activity", "error", err, "keyResultID", kr.ID)
		// Don't fail the create operation if activity recording fails
	}

	return CoreKeyResult{
		ID:              id,
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
	}, nil
}

// Update updates a key result in the system
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID, updates map[string]any, comment string) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Update")
	defer span.End()

	// Get the current key result before updating to capture its details
	previousKR, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Handle contributors separately
	var contributors []uuid.UUID
	var contributorsChanged bool
	if contribVal, exists := updates["contributors"]; exists {
		if contribSlice, ok := contribVal.([]uuid.UUID); ok {
			contributors = contribSlice
			contributorsChanged = haveContributorsChanged(contributors, previousKR.Contributors)
		}
		delete(updates, "contributors") // Remove from updates map
	}

	// Filter updates to only include fields that have actually changed
	changedUpdates := make(map[string]any)
	currentKR := CoreKeyResult(previousKR) // Convert to core type for comparison
	for field, value := range updates {
		if s.hasFieldChanged(field, value, currentKR) {
			changedUpdates[field] = value
		}
	}

	if len(changedUpdates) == 0 && !contributorsChanged {
		span.AddEvent("no changes detected", trace.WithAttributes(
			attribute.String("key_result.id", id.String()),
		))
		return nil
	}

	// Only update if there are actual changes
	if len(changedUpdates) > 0 {
		if err := s.repo.Update(ctx, id, workspaceId, changedUpdates); err != nil {
			if errors.Is(err, keyresultsrepo.ErrNotFound) {
				return ErrNotFound
			}
			return err
		}
	}
	ca := []okractivities.CoreNewActivity{}
	// Update contributors if they changed
	if contributorsChanged {
		if err := s.repo.UpdateContributors(ctx, id, contributors); err != nil {
			return fmt.Errorf("failed to update contributors: %w", err)
		}
		if len(contributors) > 0 {
			// record the update activity
			strs := make([]string, len(contributors))
			for i, u := range contributors {
				strs[i] = u.String()
			}
			activity := okractivities.CoreNewActivity{
				ObjectiveID:  previousKR.ObjectiveID,
				KeyResultID:  &id,
				UserID:       userID,
				Type:         okractivities.ActivityTypeUpdate,
				UpdateType:   okractivities.UpdateTypeKeyResult,
				Field:        "contributors",
				CurrentValue: strings.Join(strs, ","),
				Comment:      "",
				WorkspaceID:  workspaceId,
			}
			ca = append(ca, activity)
		}
	}

	// Record activity only for fields that actually changed
	if len(changedUpdates) > 0 {
		fieldCount := 0
		totalFields := len(changedUpdates)
		for field, value := range changedUpdates {
			fieldCount++
			// Only add comment to the last activity
			activityComment := ""
			if fieldCount == totalFields {
				activityComment = comment
			}

			activity := okractivities.CoreNewActivity{
				ObjectiveID:  previousKR.ObjectiveID,
				KeyResultID:  &id,
				UserID:       userID,
				Type:         okractivities.ActivityTypeUpdate,
				UpdateType:   okractivities.UpdateTypeKeyResult,
				Field:        field,
				CurrentValue: s.formatValue(value),
				Comment:      activityComment,
				WorkspaceID:  workspaceId,
			}
			ca = append(ca, activity)
		}

	}
	if len(ca) > 0 {
		if err := s.okrActivities.CreateBatch(ctx, ca); err != nil {
			s.log.Error(ctx, "failed to record key result update activity", "error", err, "keyResultID", id)
			// Don't fail the update operation if activity recording fails
		}
	}

	span.AddEvent("key result updated", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
		attribute.Int("fields_changed", len(changedUpdates)),
		attribute.Bool("contributors_changed", contributorsChanged),
	))

	return nil
}

// Delete removes a key result from the system
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Delete")
	defer span.End()

	// Get the current key result before deletion to capture its details
	currentKR, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	name := currentKR.Name
	objId := currentKR.ObjectiveID

	// Delete the key result
	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Record the delete activity
	activity := okractivities.CoreNewActivity{
		ObjectiveID:  objId,
		UserID:       userID,
		Type:         okractivities.ActivityTypeDelete,
		UpdateType:   okractivities.UpdateTypeKeyResult,
		Field:        "all",
		CurrentValue: name,
		Comment:      "",
		WorkspaceID:  workspaceId,
	}

	if err := s.okrActivities.Create(ctx, activity); err != nil {
		s.log.Error(ctx, "failed to record key result delete activity", "error", err, "keyResultID", id)
		// Don't fail the delete operation if activity recording fails
	}

	span.AddEvent("key result deleted", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

// Get retrieves a key result from the system
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Get")
	defer span.End()

	kr, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return CoreKeyResult{}, ErrNotFound
		}
		return CoreKeyResult{}, err
	}

	return CoreKeyResult(kr), nil
}

// List retrieves all key results for an objective
func (s *Service) List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.List")
	defer span.End()

	krs, err := s.repo.List(ctx, objectiveId, workspaceId)
	if err != nil {
		return nil, err
	}

	results := make([]CoreKeyResult, len(krs))
	for i, kr := range krs {
		results[i] = CoreKeyResult(kr)
	}

	return results, nil
}

// ListPaginated retrieves paginated key results with filters
func (s *Service) ListPaginated(ctx context.Context, filters keyresultsrepo.CoreKeyResultFilters) (keyresultsrepo.CoreKeyResultListResponse, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.ListPaginated")
	defer span.End()

	response, err := s.repo.ListPaginated(ctx, filters)
	if err != nil {
		return keyresultsrepo.CoreKeyResultListResponse{}, fmt.Errorf("listing paginated key results: %w", err)
	}

	return response, nil
}

// hasFieldChanged compares a field value with the current key result value
func (s *Service) hasFieldChanged(fieldName string, newValue any, currentKR CoreKeyResult) bool {
	switch fieldName {
	case "name":
		if currentValue, ok := newValue.(string); ok {
			return currentValue != currentKR.Name
		}
	case "measurementType", "measurement_type":
		if currentValue, ok := newValue.(string); ok {
			return currentValue != currentKR.MeasurementType
		}
	case "startValue", "start_value":
		if currentValue, ok := newValue.(float64); ok {
			return currentValue != currentKR.StartValue
		}
	case "currentValue", "current_value":
		if currentValue, ok := newValue.(float64); ok {
			return currentValue != currentKR.CurrentValue
		}
	case "targetValue", "target_value":
		if currentValue, ok := newValue.(float64); ok {
			return currentValue != currentKR.TargetValue
		}
	case "lead":
		if currentValue, ok := newValue.(*uuid.UUID); ok {
			if currentValue == nil && currentKR.Lead == nil {
				return false
			}
			if currentValue == nil || currentKR.Lead == nil {
				return true
			}
			return *currentValue != *currentKR.Lead
		}
	case "startDate", "start_date":
		if currentValue, ok := newValue.(*time.Time); ok {
			if currentValue == nil && currentKR.StartDate == nil {
				return false
			}
			if currentValue == nil || currentKR.StartDate == nil {
				return true
			}
			return !currentValue.Equal(*currentKR.StartDate)
		}
	case "endDate", "end_date":
		if currentValue, ok := newValue.(*time.Time); ok {
			if currentValue == nil && currentKR.EndDate == nil {
				return false
			}
			if currentValue == nil || currentKR.EndDate == nil {
				return true
			}
			return !currentValue.Equal(*currentKR.EndDate)
		}
	}
	return true // If we can't compare, assume it changed
}

// haveContributorsChanged compares contributor UUID slices
func haveContributorsChanged(newContributors []uuid.UUID, currentContributors []uuid.UUID) bool {
	if len(newContributors) != len(currentContributors) {
		return true
	}
	for i, contributor := range newContributors {
		if i >= len(currentContributors) || contributor != currentContributors[i] {
			return true
		}
	}
	return false
}

func (s *Service) formatValue(value any) string {
	if value == nil {
		return "nil"
	}
	switch v := value.(type) {
	case *float64:
		if v != nil {
			return fmt.Sprintf("%.2f", *v)
		}
		return "nil"
	case *uuid.UUID:
		if v != nil {
			return v.String()
		}
		return "nil"
	case *time.Time:
		if v != nil {
			return v.Format(time.RFC3339)
		}
		return "nil"
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}
