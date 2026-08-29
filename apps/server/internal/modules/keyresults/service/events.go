package keyresults

import (
	"context"
	"slices"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

var notifiableFields = []string{"lead", "contributors", "target_value", "end_date"}

func hasNotifiableKeyResultUpdate(fields []string) bool {
	for _, field := range fields {
		if slices.Contains(notifiableFields, field) {
			return true
		}
	}
	return false
}

func (service *Service) publishUpdate(
	ctx context.Context,
	result keyresultsdomain.MutationResult,
	actorID, workspaceID uuid.UUID,
) {
	if service.publisher == nil || !hasNotifiableKeyResultUpdate(result.ChangedFields) {
		return
	}
	updates := eventUpdates(result.After, result.ChangedFields)
	err := service.publisher.Publish(ctx, events.Event{
		Type: events.KeyResultUpdated,
		Payload: events.KeyResultUpdatedPayload{
			KeyResultID: result.After.ID, ObjectiveID: result.After.ObjectiveID,
			WorkspaceID: workspaceID, Updates: updates,
		},
		Timestamp: service.now().UTC(), ActorID: actorID,
	})
	if err != nil && service.log != nil {
		service.log.Error(ctx, "failed to publish committed key result update", "error", err, "keyResultID", result.After.ID)
	}
}

func eventUpdates(keyResult CoreKeyResult, fields []string) map[string]any {
	updates := make(map[string]any, len(fields))
	for _, field := range fields {
		switch field {
		case "name":
			updates[field] = keyResult.Name
		case "measurement_type":
			updates[field] = keyResult.MeasurementType
		case "start_value":
			updates[field] = keyResult.StartValue
		case "current_value":
			updates[field] = keyResult.CurrentValue
		case "target_value":
			updates[field] = keyResult.TargetValue
		case "lead":
			updates[field] = keyResult.Lead
		case "contributors":
			updates[field] = append([]uuid.UUID(nil), keyResult.Contributors...)
		case "start_date":
			updates[field] = keyResult.StartDate
		case "end_date":
			updates[field] = keyResult.EndDate
		}
	}
	return updates
}
