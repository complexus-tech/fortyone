package objectivesrepository

import (
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/google/uuid"
)

type objectiveValues struct {
	id                  uuid.UUID
	sequenceID          int32
	name                string
	description         *string
	shortSummary        *string
	leadUserID          *uuid.UUID
	teamID              *uuid.UUID
	workspaceID         *uuid.UUID
	startDate           *time.Time
	endDate             *time.Time
	isPrivate           bool
	createdAt           time.Time
	updatedAt           time.Time
	statusID            *uuid.UUID
	priority            *string
	health              string
	color               string
	createdBy           *uuid.UUID
	keyResultCount      int32
	totalStories        int32
	cancelledStories    int32
	completedStories    int32
	startedStories      int32
	unstartedStories    int32
	backlogStories      int32
	forecastStartDate   time.Time
	hasForecastStart    bool
	forecastEndDate     time.Time
	hasForecastEnd      bool
	forecastCauseID     *uuid.UUID
	forecastCauseSeqID  *int32
	forecastCauseTitle  *string
	forecastCauseSource *string
}

func objectiveFromValues(values objectiveValues) objectivesdomain.Objective {
	objective := objectivesdomain.Objective{
		ID: values.id, SequenceID: int(values.sequenceID), Name: values.name,
		Description: values.description, ShortSummary: values.shortSummary,
		LeadUser: values.leadUserID, Team: uuidOrNil(values.teamID),
		Workspace: uuidOrNil(values.workspaceID), StartDate: utcPointer(values.startDate),
		EndDate: utcPointer(values.endDate), IsPrivate: values.isPrivate,
		CreatedAt: values.createdAt.UTC(), UpdatedAt: values.updatedAt.UTC(),
		Status: uuidOrNil(values.statusID), Priority: values.priority,
		Health: objectiveHealth(values.health), Color: values.color,
		CreatedBy: uuidOrNil(values.createdBy), KeyResultCount: int(values.keyResultCount),
		TotalStories: int(values.totalStories), CancelledStories: int(values.cancelledStories),
		CompletedStories: int(values.completedStories), StartedStories: int(values.startedStories),
		UnstartedStories: int(values.unstartedStories), BacklogStories: int(values.backlogStories),
	}
	if values.hasForecastStart {
		forecast := values.forecastStartDate.UTC()
		objective.ForecastStartDate = &forecast
	}
	if values.hasForecastEnd {
		forecast := values.forecastEndDate.UTC()
		objective.ForecastEndDate = &forecast
	}
	if values.forecastCauseID != nil && values.forecastCauseSeqID != nil && values.forecastCauseTitle != nil {
		source := "planning"
		if values.forecastCauseSource != nil {
			source = *values.forecastCauseSource
		}
		objective.ForecastCauseStory = &objectivesdomain.ForecastCauseStory{
			ID: *values.forecastCauseID, SequenceID: int(*values.forecastCauseSeqID),
			Title: *values.forecastCauseTitle, Source: source,
		}
	}
	objective.ApplyScheduleForecast()
	return objective
}

func objectiveFromListRow(row objectivessql.ListObjectivesRow) objectivesdomain.Objective {
	return objectiveFromValues(objectiveValues{
		id: row.ObjectiveID, sequenceID: row.SequenceID, name: row.Name,
		description: row.Description, shortSummary: row.ShortSummary, leadUserID: row.LeadUserID,
		teamID: row.TeamID, workspaceID: row.WorkspaceID, startDate: row.StartDate, endDate: row.EndDate,
		isPrivate: row.IsPrivate, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		statusID: row.StatusID, priority: row.Priority, health: row.Health, color: row.Color,
		createdBy: row.CreatedBy, keyResultCount: row.KeyResultCount, totalStories: row.TotalStories,
		cancelledStories: row.CancelledStories, completedStories: row.CompletedStories,
		startedStories: row.StartedStories, unstartedStories: row.UnstartedStories,
		backlogStories: row.BacklogStories, forecastStartDate: row.ForecastStartDate,
		hasForecastStart: row.HasForecastStartDate, forecastEndDate: row.ForecastEndDate,
		hasForecastEnd: row.HasForecastEndDate, forecastCauseID: row.ForecastCauseID,
		forecastCauseSeqID: row.ForecastCauseSequenceID, forecastCauseTitle: row.ForecastCauseTitle,
		forecastCauseSource: row.ForecastCauseSource,
	})
}

func objectiveFromGetRow(row objectivessql.GetObjectiveRow) objectivesdomain.Objective {
	return objectiveFromValues(objectiveValues{
		id: row.ObjectiveID, sequenceID: row.SequenceID, name: row.Name,
		description: row.Description, shortSummary: row.ShortSummary, leadUserID: row.LeadUserID,
		teamID: row.TeamID, workspaceID: row.WorkspaceID, startDate: row.StartDate, endDate: row.EndDate,
		isPrivate: row.IsPrivate, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		statusID: row.StatusID, priority: row.Priority, health: row.Health, color: row.Color,
		createdBy: row.CreatedBy, keyResultCount: row.KeyResultCount, totalStories: row.TotalStories,
		cancelledStories: row.CancelledStories, completedStories: row.CompletedStories,
		startedStories: row.StartedStories, unstartedStories: row.UnstartedStories,
		backlogStories: row.BacklogStories, forecastStartDate: row.ForecastStartDate,
		hasForecastStart: row.HasForecastStartDate, forecastEndDate: row.ForecastEndDate,
		hasForecastEnd: row.HasForecastEndDate, forecastCauseID: row.ForecastCauseID,
		forecastCauseSeqID: row.ForecastCauseSequenceID, forecastCauseTitle: row.ForecastCauseTitle,
		forecastCauseSource: row.ForecastCauseSource,
	})
}

func objectiveFromCreateRow(row objectivessql.CreateObjectiveRow) objectivesdomain.Objective {
	return objectiveFromValues(objectiveValues{
		id: row.ObjectiveID, sequenceID: row.SequenceID, name: row.Name,
		description: row.Description, shortSummary: row.ShortSummary, leadUserID: row.LeadUserID,
		teamID: row.TeamID, workspaceID: row.WorkspaceID, startDate: row.StartDate, endDate: row.EndDate,
		isPrivate: row.IsPrivate, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		statusID: row.StatusID, priority: row.Priority, health: row.Health, color: row.Color,
		createdBy: row.CreatedBy,
	})
}

func objectiveFromUpdateRow(row objectivessql.UpdateObjectiveRow) objectivesdomain.Objective {
	return objectiveFromValues(objectiveValues{
		id: row.ObjectiveID, sequenceID: row.SequenceID, name: row.Name,
		description: row.Description, shortSummary: row.ShortSummary, leadUserID: row.LeadUserID,
		teamID: row.TeamID, workspaceID: row.WorkspaceID, startDate: row.StartDate, endDate: row.EndDate,
		isPrivate: row.IsPrivate, createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		statusID: row.StatusID, priority: row.Priority, health: row.Health, color: row.Color,
		createdBy: row.CreatedBy,
	})
}

func keyResultFromCreateRow(row objectivessql.CreateObjectiveKeyResultRow, contributors []uuid.UUID) objectivesdomain.KeyResult {
	return objectivesdomain.KeyResult{
		ID: row.ID, SequenceID: int(row.SequenceID), ObjectiveID: row.ObjectiveID,
		Name: row.Name, MeasurementType: row.MeasurementType, StartValue: row.StartValue,
		CurrentValue: row.CurrentValue, TargetValue: row.TargetValue, Lead: row.Lead,
		Contributors: append([]uuid.UUID(nil), contributors...), StartDate: utcPointer(&row.StartDate),
		EndDate: utcPointer(&row.EndDate), CreatedAt: row.CreatedAt.UTC(), UpdatedAt: row.UpdatedAt.UTC(),
		CreatedBy: uuidOrNil(row.CreatedBy),
	}
}

func objectiveHealth(value string) *objectivesdomain.ObjectiveHealth {
	if value == "" {
		return nil
	}
	health := objectivesdomain.ObjectiveHealth(value)
	return &health
}

func uuidOrNil(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func uuidPointer(value uuid.UUID) *uuid.UUID { return &value }
