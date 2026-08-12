package keyresults

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Set of error variables for key result operations
var (
	ErrNotFound        = errors.New("key result not found")
	ErrVersionConflict = errors.New("key result changed since it was reviewed")
)

// Repository defines the storage contract for key results.
type Repository interface {
	Create(ctx context.Context, kr *CoreKeyResult, workspaceID uuid.UUID) (uuid.UUID, int, error)
	CreateBatch(ctx context.Context, keyResults []CoreKeyResult, workspaceID uuid.UUID) ([]CoreKeyResult, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	UpdateIfUnchanged(ctx context.Context, id, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) (bool, error)
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error)
	List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error)
	ListPaginated(ctx context.Context, filters CoreKeyResultFilters) (CoreKeyResultListResponse, error)
	AddContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	UpdateContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	GetContributors(ctx context.Context, keyResultID uuid.UUID) ([]uuid.UUID, error)
}

// Service manages the key result operations
type Service struct {
	repo          Repository
	okrActivities *okractivities.Service
	log           *logger.Logger
	publisher     *publisher.Publisher
}

// Option configures optional key result service dependencies.
type Option func(*Service)

// WithPublisher publishes key result changes for notification and integration consumers.
func WithPublisher(eventPublisher *publisher.Publisher) Option {
	return func(service *Service) {
		service.publisher = eventPublisher
	}
}

// New creates a new key result service
func New(log *logger.Logger, repo Repository, okrActivities *okractivities.Service, options ...Option) *Service {
	service := &Service{
		repo:          repo,
		okrActivities: okrActivities,
		log:           log,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

// Create inserts a new key result into the system
func (s *Service) Create(ctx context.Context, nkr CoreNewKeyResult, workspaceID uuid.UUID) (CoreKeyResult, error) {

	kr := CoreKeyResult{
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

	id, sequenceID, err := s.repo.Create(ctx, &kr, workspaceID)
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
		SequenceID:      sequenceID,
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

// CreateBatch creates key results atomically for one objective.
func (s *Service) CreateBatch(ctx context.Context, newKeyResults []CoreNewKeyResult, workspaceID uuid.UUID) ([]CoreKeyResult, error) {
	if len(newKeyResults) == 0 {
		return []CoreKeyResult{}, nil
	}

	keyResults := make([]CoreKeyResult, len(newKeyResults))
	for i, newKeyResult := range newKeyResults {
		keyResults[i] = CoreKeyResult{
			ObjectiveID:     newKeyResult.ObjectiveID,
			Name:            newKeyResult.Name,
			MeasurementType: newKeyResult.MeasurementType,
			StartValue:      newKeyResult.StartValue,
			CurrentValue:    newKeyResult.CurrentValue,
			TargetValue:     newKeyResult.TargetValue,
			Lead:            newKeyResult.Lead,
			Contributors:    newKeyResult.Contributors,
			StartDate:       newKeyResult.StartDate,
			EndDate:         newKeyResult.EndDate,
			CreatedBy:       newKeyResult.CreatedBy,
		}
	}

	createdKeyResults, err := s.repo.CreateBatch(ctx, keyResults, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("create batch: %w", err)
	}

	activities := make([]okractivities.CoreNewActivity, len(createdKeyResults))
	for i, keyResult := range createdKeyResults {
		keyResultID := keyResult.ID
		activities[i] = okractivities.CoreNewActivity{
			ObjectiveID:  keyResult.ObjectiveID,
			KeyResultID:  &keyResultID,
			UserID:       keyResult.CreatedBy,
			Type:         okractivities.ActivityTypeCreate,
			UpdateType:   okractivities.UpdateTypeKeyResult,
			Field:        "all",
			CurrentValue: keyResult.Name,
			WorkspaceID:  workspaceID,
		}
	}
	if err := s.okrActivities.CreateBatch(ctx, activities); err != nil {
		s.log.Error(ctx, "failed to record key result batch create activities", "error", err)
	}

	return createdKeyResults, nil
}

// Update updates a key result in the system
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID, updates map[string]any, comment string) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Update")
	defer span.End()

	// Get the current key result before updating to capture its details
	previousKR, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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
	currentKR := previousKR
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
			if errors.Is(err, ErrNotFound) {
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

	eventUpdates := make(map[string]any, len(changedUpdates)+1)
	for field, value := range changedUpdates {
		eventUpdates[field] = value
	}
	if contributorsChanged {
		eventUpdates["contributors"] = contributors
	}
	if s.publisher != nil && hasNotifiableKeyResultUpdate(eventUpdates) {
		event := events.Event{
			Type: events.KeyResultUpdated,
			Payload: events.KeyResultUpdatedPayload{
				KeyResultID: id,
				ObjectiveID: previousKR.ObjectiveID,
				WorkspaceID: workspaceId,
				Updates:     eventUpdates,
			},
			Timestamp: time.Now().UTC(),
			ActorID:   userID,
		}
		if publishErr := s.publisher.Publish(ctx, event); publishErr != nil {
			s.log.Error(ctx, "failed to publish key result update event", "error", publishErr, "keyResultID", id)
		}
	}

	span.AddEvent("key result updated", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
		attribute.Int("fields_changed", len(changedUpdates)),
		attribute.Bool("contributors_changed", contributorsChanged),
	))

	return nil
}

// UpdateExternalUserActionIfUnchanged applies a user-requested scalar update
// only while the key result still matches the version shown in email. Email
// actions intentionally cannot change contributor membership.
func (s *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	id, workspaceID, userID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
	comment string,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected key result update time is required")
	}
	if _, exists := updates["contributors"]; exists {
		return errors.New("email key result actions cannot change contributors")
	}
	previous, err := s.repo.Get(ctx, id, workspaceID)
	if err != nil {
		return err
	}
	changed := make(map[string]any, len(updates))
	for field, value := range updates {
		if s.hasFieldChanged(field, value, previous) {
			changed[field] = value
		}
	}
	if len(changed) == 0 {
		return nil
	}
	updated, err := s.repo.UpdateIfUnchanged(ctx, id, workspaceID, expectedUpdatedAt.UTC(), changed)
	if err != nil {
		return err
	}
	if !updated {
		return ErrVersionConflict
	}

	activities := make([]okractivities.CoreNewActivity, 0, len(changed))
	fieldIndex := 0
	for field, value := range changed {
		fieldIndex++
		activityComment := ""
		if fieldIndex == len(changed) {
			activityComment = comment
		}
		activities = append(activities, okractivities.CoreNewActivity{
			ObjectiveID:  previous.ObjectiveID,
			KeyResultID:  &id,
			UserID:       userID,
			Type:         okractivities.ActivityTypeUpdate,
			UpdateType:   okractivities.UpdateTypeKeyResult,
			Field:        field,
			CurrentValue: s.formatValue(value),
			Comment:      activityComment,
			WorkspaceID:  workspaceID,
		})
	}
	if err := s.okrActivities.CreateBatch(ctx, activities); err != nil {
		s.log.Error(ctx, "failed to record external key result update activity", "error", err, "keyResultID", id)
	}
	if s.publisher != nil && hasNotifiableKeyResultUpdate(changed) {
		if err := s.publisher.Publish(ctx, events.Event{
			Type: events.KeyResultUpdated,
			Payload: events.KeyResultUpdatedPayload{
				KeyResultID: id,
				ObjectiveID: previous.ObjectiveID,
				WorkspaceID: workspaceID,
				Updates:     changed,
			},
			Timestamp: time.Now().UTC(),
			ActorID:   userID,
		}); err != nil {
			s.log.Error(ctx, "failed to publish external key result update event", "error", err, "keyResultID", id)
		}
	}
	return nil
}

func hasNotifiableKeyResultUpdate(updates map[string]any) bool {
	for _, field := range []string{"lead", "contributors", "current_value", "target_value", "start_date", "end_date"} {
		if _, exists := updates[field]; exists {
			return true
		}
	}
	return false
}

// Delete removes a key result from the system
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Delete")
	defer span.End()

	// Get the current key result before deletion to capture its details
	currentKR, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	name := currentKR.Name
	objId := currentKR.ObjectiveID

	// Delete the key result
	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		if errors.Is(err, ErrNotFound) {
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
		if errors.Is(err, ErrNotFound) {
			return CoreKeyResult{}, ErrNotFound
		}
		return CoreKeyResult{}, err
	}

	return kr, nil
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
		results[i] = kr
	}

	return results, nil
}

// ListPaginated retrieves paginated key results with filters
func (s *Service) ListPaginated(ctx context.Context, filters CoreKeyResultFilters) (CoreKeyResultListResponse, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.ListPaginated")
	defer span.End()

	response, err := s.repo.ListPaginated(ctx, filters)
	if err != nil {
		return CoreKeyResultListResponse{}, fmt.Errorf("listing paginated key results: %w", err)
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
