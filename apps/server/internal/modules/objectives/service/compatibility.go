package objectives

import (
	"fmt"
	"strings"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
)

func listQueryFromCompatibilityMap(
	workspaceID, actorID uuid.UUID,
	filters map[string]any,
) (objectivesdomain.ListQuery, error) {
	query := objectivesdomain.ListQuery{WorkspaceID: workspaceID, ActorID: actorID}
	for field, raw := range filters {
		switch field {
		case "objective_id":
			value, err := compatibilityUUID(raw)
			if err != nil {
				return objectivesdomain.ListQuery{}, fmt.Errorf("%w: objective_id: %v", ErrInvalid, err)
			}
			query.ObjectiveID = value
		case "team_id":
			value, err := compatibilityUUID(raw)
			if err != nil {
				return objectivesdomain.ListQuery{}, fmt.Errorf("%w: team_id: %v", ErrInvalid, err)
			}
			query.TeamID = value
		case "search":
			value, ok := raw.(string)
			if !ok {
				return objectivesdomain.ListQuery{}, fmt.Errorf("%w: search must be a string", ErrInvalid)
			}
			query.Search = value
		case "limit":
			value, ok := compatibilityInt(raw)
			if !ok {
				return objectivesdomain.ListQuery{}, fmt.Errorf("%w: limit must be an integer", ErrInvalid)
			}
			query.Limit = value
		case "offset":
			value, ok := compatibilityInt(raw)
			if !ok {
				return objectivesdomain.ListQuery{}, fmt.Errorf("%w: offset must be an integer", ErrInvalid)
			}
			query.Offset = value
		case "page", "pageSize":
			// HTTP pagination translates these into limit/offset before the
			// service boundary. They remain accepted for older in-process callers.
		default:
			return objectivesdomain.ListQuery{}, fmt.Errorf("%w: unsupported objective filter %q", ErrInvalid, field)
		}
	}
	return query.Normalize()
}

func objectivePatchFromCompatibilityMap(updates map[string]any) (objectivesdomain.ObjectivePatch, error) {
	var patch objectivesdomain.ObjectivePatch
	for field, raw := range updates {
		switch field {
		case "name":
			value, ok := raw.(string)
			if !ok {
				return patch, invalidCompatibilityField(field)
			}
			patch.Name = objectivesdomain.SetField(value)
		case "description":
			value, err := compatibilityNullableString(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.Description = nullableField(value)
		case "short_summary":
			value, err := compatibilityNullableString(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.ShortSummary = nullableField(value)
		case "lead_user_id":
			value, err := compatibilityUUID(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.LeadUser = nullableField(value)
		case "start_date":
			value, err := compatibilityTime(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.StartDate = nullableField(value)
		case "end_date":
			value, err := compatibilityTime(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.EndDate = nullableField(value)
		case "is_private":
			value, ok := raw.(bool)
			if !ok {
				return patch, invalidCompatibilityField(field)
			}
			patch.IsPrivate = objectivesdomain.SetField(value)
		case "status_id":
			value, err := compatibilityUUID(raw)
			if err != nil || value == nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.Status = objectivesdomain.SetField(*value)
		case "priority":
			value, err := compatibilityNullableString(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.Priority = nullableField(value)
		case "health":
			value, err := compatibilityHealth(raw)
			if err != nil {
				return patch, invalidCompatibilityField(field)
			}
			patch.Health = nullableField(value)
		case "color":
			value, ok := raw.(string)
			if !ok {
				return patch, invalidCompatibilityField(field)
			}
			patch.Color = objectivesdomain.SetField(value)
		default:
			return patch, fmt.Errorf("%w: unsupported objective update %q", ErrInvalid, field)
		}
	}
	return patch, patch.Validate()
}

func compatibilityUUID(raw any) (*uuid.UUID, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case uuid.UUID:
		return &value, nil
	case *uuid.UUID:
		return value, nil
	case string:
		parsed, err := uuid.Parse(strings.TrimSpace(value))
		return &parsed, err
	default:
		return nil, fmt.Errorf("must be a UUID")
	}
}

func compatibilityTime(raw any) (*time.Time, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case time.Time:
		value = value.UTC()
		return &value, nil
	case *time.Time:
		if value == nil {
			return nil, nil
		}
		utc := value.UTC()
		return &utc, nil
	default:
		return nil, fmt.Errorf("must be a time")
	}
}

func compatibilityNullableString(raw any) (*string, error) {
	switch value := raw.(type) {
	case nil:
		return nil, nil
	case string:
		return &value, nil
	case *string:
		return value, nil
	default:
		return nil, fmt.Errorf("must be a string")
	}
}

func compatibilityHealth(raw any) (*objectivesdomain.ObjectiveHealth, error) {
	if raw == nil {
		return nil, nil
	}
	var health objectivesdomain.ObjectiveHealth
	switch value := raw.(type) {
	case string:
		health = objectivesdomain.ObjectiveHealth(value)
	case objectivesdomain.ObjectiveHealth:
		health = value
	case *string:
		if value == nil {
			return nil, nil
		}
		health = objectivesdomain.ObjectiveHealth(*value)
	case *objectivesdomain.ObjectiveHealth:
		return value, nil
	default:
		return nil, fmt.Errorf("must be an objective health")
	}
	return &health, nil
}

func compatibilityInt(raw any) (int, bool) {
	switch value := raw.(type) {
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}

func nullableField[T any](value *T) objectivesdomain.Field[T] {
	if value == nil {
		return objectivesdomain.ClearField[T]()
	}
	return objectivesdomain.SetField(*value)
}

func invalidCompatibilityField(field string) error {
	return fmt.Errorf("%w: objective update %q has an invalid value", ErrInvalid, field)
}
