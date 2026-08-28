package objectives

import (
	"context"
	"reflect"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (service *Service) publishUpdate(
	ctx context.Context,
	objective CoreObjective,
	actorID uuid.UUID,
	patch objectivesdomain.ObjectivePatch,
) {
	updates := objectivePatchMap(patch)
	if service.publisher == nil || !hasNotifiableObjectiveUpdate(updates) {
		return
	}
	event := events.Event{
		Type: events.ObjectiveUpdated,
		Payload: events.ObjectiveUpdatedPayload{
			ObjectiveID: objective.ID, WorkspaceID: objective.Workspace,
			Updates: updates, LeadID: objective.LeadUser,
		},
		Timestamp: service.now().UTC(), ActorID: actorID,
	}
	if err := service.publisher.Publish(ctx, event); err != nil && service.log != nil {
		service.log.Error(ctx, "failed to publish objective update event", "error", err, "objective_id", objective.ID)
	}
}

func objectivePatchMap(patch objectivesdomain.ObjectivePatch) map[string]any {
	updates := make(map[string]any, len(patch.Fields()))
	for _, field := range patch.Fields() {
		switch field {
		case "name":
			updates[field] = fieldValue(patch.Name.Value())
		case "description":
			updates[field] = fieldValue(patch.Description.Value())
		case "short_summary":
			updates[field] = fieldValue(patch.ShortSummary.Value())
		case "lead_user_id":
			updates[field] = fieldValue(patch.LeadUser.Value())
		case "start_date":
			updates[field] = fieldValue(patch.StartDate.Value())
		case "end_date":
			updates[field] = fieldValue(patch.EndDate.Value())
		case "is_private":
			updates[field] = fieldValue(patch.IsPrivate.Value())
		case "status_id":
			updates[field] = fieldValue(patch.Status.Value())
		case "priority":
			updates[field] = fieldValue(patch.Priority.Value())
		case "health":
			value, _ := patch.Health.Value()
			if value == nil {
				updates[field] = nil
			} else {
				updates[field] = string(*value)
			}
		case "color":
			updates[field] = fieldValue(patch.Color.Value())
		}
	}
	return updates
}

func fieldValue[T any](value *T, specified bool) any {
	if !specified || value == nil {
		return nil
	}
	return *value
}

func hasNotifiableObjectiveUpdate(updates map[string]any) bool {
	for _, field := range []string{"lead_user_id", "status_id", "health", "end_date"} {
		if _, exists := updates[field]; exists {
			return true
		}
	}
	return false
}

func objectiveExternalUpdatesAlreadyApplied(objective CoreObjective, updates map[string]any) bool {
	patch, err := objectivePatchFromCompatibilityMap(updates)
	return err == nil && objectivePatchAlreadyApplied(objective, patch)
}

func objectivePatchAlreadyApplied(objective CoreObjective, patch objectivesdomain.ObjectivePatch) bool {
	for _, field := range patch.Fields() {
		switch field {
		case "name":
			if !scalarFieldMatches(objective.Name, patch.Name) {
				return false
			}
		case "description":
			if !nullableFieldMatches(objective.Description, patch.Description) {
				return false
			}
		case "short_summary":
			if !nullableFieldMatches(objective.ShortSummary, patch.ShortSummary) {
				return false
			}
		case "lead_user_id":
			if !nullableFieldMatches(objective.LeadUser, patch.LeadUser) {
				return false
			}
		case "start_date":
			if !timeFieldMatches(objective.StartDate, patch.StartDate) {
				return false
			}
		case "end_date":
			if !timeFieldMatches(objective.EndDate, patch.EndDate) {
				return false
			}
		case "is_private":
			if !scalarFieldMatches(objective.IsPrivate, patch.IsPrivate) {
				return false
			}
		case "status_id":
			if !scalarFieldMatches(objective.Status, patch.Status) {
				return false
			}
		case "priority":
			if !nullableFieldMatches(objective.Priority, patch.Priority) {
				return false
			}
		case "health":
			if !nullableFieldMatches(objective.Health, patch.Health) {
				return false
			}
		case "color":
			if !scalarFieldMatches(objective.Color, patch.Color) {
				return false
			}
		}
	}
	return true
}

func scalarFieldMatches[T comparable](current T, field objectivesdomain.Field[T]) bool {
	desired, specified := field.Value()
	return specified && desired != nil && current == *desired
}

func nullableFieldMatches[T comparable](current *T, field objectivesdomain.Field[T]) bool {
	desired, specified := field.Value()
	return specified && reflect.DeepEqual(current, desired)
}

func timeFieldMatches(current *time.Time, field objectivesdomain.Field[time.Time]) bool {
	desired, specified := field.Value()
	if !specified || current == nil || desired == nil {
		return specified && current == nil && desired == nil
	}
	return current.UTC().Equal(desired.UTC())
}
