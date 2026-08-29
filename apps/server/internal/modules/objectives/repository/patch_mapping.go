package objectivesrepository

import (
	"fmt"
	"strconv"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/google/uuid"
)

func objectivePatchParams(
	command objectivesdomain.UpdateCommand,
) objectivessql.UpdateObjectiveParams {
	params := objectivessql.UpdateObjectiveParams{
		ObjectiveID: command.ObjectiveID,
		WorkspaceID: uuidPointer(command.WorkspaceID),
		ActorID:     command.ActorID,
	}
	if value, specified := command.Patch.Name.Value(); specified {
		params.SetName = true
		params.Name = valueOrZero(value)
	}
	if value, specified := command.Patch.Description.Value(); specified {
		params.SetDescription, params.Description = true, value
	}
	if value, specified := command.Patch.ShortSummary.Value(); specified {
		params.SetShortSummary, params.ShortSummary = true, value
	}
	if value, specified := command.Patch.LeadUser.Value(); specified {
		params.SetLeadUserID, params.LeadUserID = true, value
	}
	if value, specified := command.Patch.StartDate.Value(); specified {
		params.SetStartDate, params.StartDate = true, utcPointer(value)
	}
	if value, specified := command.Patch.EndDate.Value(); specified {
		params.SetEndDate, params.EndDate = true, utcPointer(value)
	}
	if value, specified := command.Patch.IsPrivate.Value(); specified {
		params.SetIsPrivate = true
		params.IsPrivate = valueOrZero(value)
	}
	if value, specified := command.Patch.Status.Value(); specified {
		params.SetStatusID, params.StatusID = true, value
	}
	if value, specified := command.Patch.Priority.Value(); specified {
		params.SetPriority, params.Priority = true, value
	}
	if value, specified := command.Patch.Health.Value(); specified {
		params.SetHealth = true
		if value != nil {
			health := objectivessql.ObjectiveHealthStatus(*value)
			params.Health = &health
		}
	}
	if value, specified := command.Patch.Color.Value(); specified {
		params.SetColor = true
		params.Color = valueOrZero(value)
	}
	return params
}

func patchReferenceParams(
	command objectivesdomain.UpdateCommand,
	teamID uuid.UUID,
) objectivessql.ValidateObjectivePatchReferencesParams {
	status, statusSpecified := command.Patch.Status.Value()
	lead, leadSpecified := command.Patch.LeadUser.Value()
	return objectivessql.ValidateObjectivePatchReferencesParams{
		StatusSpecified: statusSpecified, StatusID: status,
		LeadSpecified: leadSpecified, LeadUserID: lead,
		WorkspaceID: command.WorkspaceID, TeamID: teamID,
	}
}

type objectiveChange struct {
	field string
	value string
}

func objectivePatchChanges(patch objectivesdomain.ObjectivePatch) []objectiveChange {
	changes := make([]objectiveChange, 0, len(patch.Fields()))
	for _, field := range patch.Fields() {
		// These rich-text autosaves were deliberately absent from the legacy
		// activity feed. Preserve that product behavior.
		if field == "description" || field == "short_summary" {
			continue
		}
		changes = append(changes, objectiveChange{field: field, value: objectivePatchValue(patch, field)})
	}
	return changes
}

func objectivePatchValue(patch objectivesdomain.ObjectivePatch, field string) string {
	switch field {
	case "name":
		return formatPatchValue(patch.Name.Value())
	case "lead_user_id":
		return formatPatchValue(patch.LeadUser.Value())
	case "start_date":
		return formatPatchValue(patch.StartDate.Value())
	case "end_date":
		return formatPatchValue(patch.EndDate.Value())
	case "is_private":
		return formatPatchValue(patch.IsPrivate.Value())
	case "status_id":
		return formatPatchValue(patch.Status.Value())
	case "priority":
		return formatPatchValue(patch.Priority.Value())
	case "health":
		return formatPatchValue(patch.Health.Value())
	case "color":
		return formatPatchValue(patch.Color.Value())
	default:
		return "nil"
	}
}

func formatPatchValue[T any](value *T, specified bool) string {
	if !specified || value == nil {
		return "nil"
	}
	switch typed := any(*value).(type) {
	case time.Time:
		return typed.UTC().Format(time.RFC3339)
	case uuid.UUID:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func valueOrZero[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
