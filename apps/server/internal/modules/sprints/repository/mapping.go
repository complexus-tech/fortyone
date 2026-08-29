package sprintsrepository

import (
	"fmt"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprintssql "github.com/complexus-tech/projects-api/internal/modules/sprints/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

type sprintRecord struct {
	ID                          uuid.UUID
	Name                        string
	Goal                        *string
	ObjectiveID                 *uuid.UUID
	TeamID                      uuid.UUID
	WorkspaceID                 uuid.UUID
	StartDate                   time.Time
	EndDate                     time.Time
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	ScheduleManagedByAutomation bool
	Counts                      [6]int64
}

func sprintFromRecord(record sprintRecord) (sprintdomain.Sprint, error) {
	counts := [6]int{}
	for index, value := range record.Counts {
		converted, err := safecast.Int64(value)
		if err != nil {
			return sprintdomain.Sprint{}, fmt.Errorf("map sprint count %d: %w", index, err)
		}
		counts[index] = converted
	}
	return sprintdomain.Sprint{
		ID: record.ID, Name: record.Name, Goal: record.Goal, ObjectiveID: record.ObjectiveID,
		TeamID: record.TeamID, WorkspaceID: record.WorkspaceID,
		StartDate: record.StartDate, EndDate: record.EndDate,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		ScheduleManagedByAutomation: record.ScheduleManagedByAutomation,
		TotalStories:                counts[0], CancelledStories: counts[1], CompletedStories: counts[2],
		StartedStories: counts[3], UnstartedStories: counts[4], BacklogStories: counts[5],
	}, nil
}

func sprintFromListRow(row sprintssql.ListSprintsRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
		Counts:                      [6]int64{row.TotalStories, row.CancelledStories, row.CompletedStories, row.StartedStories, row.UnstartedStories, row.BacklogStories},
	})
}

func sprintFromRunningRow(row sprintssql.ListRunningSprintsRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
		Counts:                      [6]int64{row.TotalStories, row.CancelledStories, row.CompletedStories, row.StartedStories, row.UnstartedStories, row.BacklogStories},
	})
}

func sprintFromGetRow(row sprintssql.GetSprintByIDRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
	})
}

func sprintFromCreateRow(row sprintssql.CreateSprintRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
	})
}

func sprintFromLockRow(row sprintssql.LockSprintForMutationRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
	})
}

func sprintFromUpdateRow(row sprintssql.UpdateSprintRow) (sprintdomain.Sprint, error) {
	return sprintFromRecord(sprintRecord{
		ID: row.SprintID, Name: row.Name, Goal: row.Goal, ObjectiveID: row.ObjectiveID,
		TeamID: row.TeamID, WorkspaceID: row.WorkspaceID, StartDate: row.StartDate, EndDate: row.EndDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ScheduleManagedByAutomation: row.ScheduleManagedByAutomation,
	})
}

func mapListRows(rows []sprintssql.ListSprintsRow) ([]sprintdomain.Sprint, error) {
	items := make([]sprintdomain.Sprint, 0, len(rows))
	for _, row := range rows {
		item, err := sprintFromListRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func mapRunningRows(rows []sprintssql.ListRunningSprintsRow) ([]sprintdomain.Sprint, error) {
	items := make([]sprintdomain.Sprint, 0, len(rows))
	for _, row := range rows {
		item, err := sprintFromRunningRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
