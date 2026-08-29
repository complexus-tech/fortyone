package teamsettingsrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/google/uuid"
)

type settingsAuditMetadata struct {
	ChangedFields []string `json:"changed_fields"`
}

type sprintScheduleAuditMetadata struct {
	OldStartDate        string `json:"old_start_date"`
	OldEndDate          string `json:"old_end_date"`
	NewStartDate        string `json:"new_start_date"`
	NewEndDate          string `json:"new_end_date"`
	SprintDurationWeeks int    `json:"sprint_duration_weeks"`
	SprintStartDay      string `json:"sprint_start_day"`
}

func insertAuditEvent(
	ctx context.Context,
	queries teamsettingssql.Querier,
	workspaceID, teamID uuid.UUID,
	actor teamsettings.AuditActor,
	entityType string,
	entityID uuid.UUID,
	eventType string,
	metadata any,
) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal team settings audit metadata: %w", err)
	}
	if err := queries.InsertTeamSettingsAuditEvent(ctx, teamsettingssql.InsertTeamSettingsAuditEventParams{
		WorkspaceID: workspaceID,
		TeamID:      &teamID,
		ActorType:   actor.Type,
		ActorID:     actor.ID,
		EntityType:  entityType,
		EntityID:    &entityID,
		EventType:   eventType,
		Metadata:    payload,
	}); err != nil {
		return fmt.Errorf("insert team settings audit event: %w", mapDatabaseError(err))
	}
	return nil
}

func scheduleAuditMetadata(
	settings teamsettings.CoreTeamSprintSettings,
	sprint scheduledSprint,
	startDate, endDate time.Time,
) sprintScheduleAuditMetadata {
	return sprintScheduleAuditMetadata{
		OldStartDate:        sprint.StartDate.Format(time.DateOnly),
		OldEndDate:          sprint.EndDate.Format(time.DateOnly),
		NewStartDate:        startDate.Format(time.DateOnly),
		NewEndDate:          endDate.Format(time.DateOnly),
		SprintDurationWeeks: settings.SprintDurationWeeks,
		SprintStartDay:      settings.SprintStartDay,
	}
}
