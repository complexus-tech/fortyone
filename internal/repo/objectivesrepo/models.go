package objectivesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/keyresults"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

type dbObjective struct {
	ID               uuid.UUID                   `db:"objective_id"`
	Name             string                      `db:"name"`
	Description      *string                     `db:"description"`
	LeadUser         *uuid.UUID                  `db:"lead_user_id"`
	Team             uuid.UUID                   `db:"team_id"`
	Workspace        uuid.UUID                   `db:"workspace_id"`
	StartDate        *time.Time                  `db:"start_date"`
	EndDate          *time.Time                  `db:"end_date"`
	IsPrivate        bool                        `db:"is_private"`
	Status           uuid.UUID                   `db:"status_id"`
	Priority         *string                     `db:"priority"`
	Health           *objectives.ObjectiveHealth `db:"health"`
	CreatedAt        time.Time                   `db:"created_at"`
	UpdatedAt        time.Time                   `db:"updated_at"`
	CreatedBy        uuid.UUID                   `db:"created_by"`
	TotalStories     int                         `db:"total_stories"`
	CancelledStories int                         `db:"cancelled_stories"`
	CompletedStories int                         `db:"completed_stories"`
	StartedStories   int                         `db:"started_stories"`
	UnstartedStories int                         `db:"unstarted_stories"`
	BacklogStories   int                         `db:"backlog_stories"`
}

type dbKeyResult struct {
	ID              uuid.UUID `db:"id"`
	ObjectiveID     uuid.UUID `db:"objective_id"`
	Name            string    `db:"name"`
	MeasurementType string    `db:"measurement_type"`
	StartValue      float64   `db:"start_value"`
	CurrentValue    float64   `db:"current_value"`
	TargetValue     float64   `db:"target_value"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	CreatedBy       uuid.UUID `db:"created_by"`
	LastUpdatedBy   uuid.UUID `db:"last_updated_by"`
}

func toDBObjective(co objectives.CoreNewObjective, workspaceID uuid.UUID) dbObjective {
	return dbObjective{
		Name:        co.Name,
		Description: co.Description,
		LeadUser:    co.LeadUser,
		Team:        co.Team,
		Workspace:   workspaceID,
		StartDate:   co.StartDate,
		EndDate:     co.EndDate,
		IsPrivate:   co.IsPrivate,
		Status:      co.Status,
		Priority:    co.Priority,
		CreatedBy:   co.CreatedBy,
	}
}

func toCoreObjective(dbo dbObjective) objectives.CoreObjective {
	return objectives.CoreObjective{
		ID:               dbo.ID,
		Name:             dbo.Name,
		Description:      dbo.Description,
		LeadUser:         dbo.LeadUser,
		Team:             dbo.Team,
		Workspace:        dbo.Workspace,
		StartDate:        dbo.StartDate,
		EndDate:          dbo.EndDate,
		IsPrivate:        dbo.IsPrivate,
		CreatedAt:        dbo.CreatedAt,
		UpdatedAt:        dbo.UpdatedAt,
		CreatedBy:        dbo.CreatedBy,
		Status:           dbo.Status,
		Priority:         dbo.Priority,
		Health:           dbo.Health,
		TotalStories:     dbo.TotalStories,
		CancelledStories: dbo.CancelledStories,
		CompletedStories: dbo.CompletedStories,
		StartedStories:   dbo.StartedStories,
		UnstartedStories: dbo.UnstartedStories,
		BacklogStories:   dbo.BacklogStories,
	}
}

func toCoreObjectives(do []dbObjective) []objectives.CoreObjective {
	objectives := make([]objectives.CoreObjective, len(do))
	for i, o := range do {
		objectives[i] = toCoreObjective(o)
	}
	return objectives
}

func toDBKeyResult(kr keyresults.CoreNewKeyResult, createdBy uuid.UUID, lastUpdatedBy uuid.UUID) dbKeyResult {
	return dbKeyResult{
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		CreatedBy:       createdBy,
		LastUpdatedBy:   lastUpdatedBy,
	}
}

func toCoreKeyResult(dbkr dbKeyResult) keyresults.CoreKeyResult {
	return keyresults.CoreKeyResult{
		ID:              dbkr.ID,
		ObjectiveID:     dbkr.ObjectiveID,
		Name:            dbkr.Name,
		MeasurementType: dbkr.MeasurementType,
		StartValue:      dbkr.StartValue,
		CurrentValue:    dbkr.CurrentValue,
		TargetValue:     dbkr.TargetValue,
		CreatedAt:       dbkr.CreatedAt,
		UpdatedAt:       dbkr.UpdatedAt,
		CreatedBy:       dbkr.CreatedBy,
	}
}
