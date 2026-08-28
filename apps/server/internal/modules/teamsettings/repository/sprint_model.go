package teamsettingsrepository

import (
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func newSprintSettings(
	teamID uuid.UUID,
	workspaceID uuid.UUID,
	autoCreateSprints bool,
	upcomingSprintsCount int32,
	sprintDurationWeeks int32,
	sprintStartDay string,
	workingDays []int16,
	moveIncompleteStoriesEnabled bool,
	lastAutoSprintNumber int32,
	nextAutoSprintNumber int32,
	autoCreateDisabledAt *time.Time,
	autoCreateDisabledReason *string,
	createdAt time.Time,
	updatedAt time.Time,
) teamsettings.CoreTeamSprintSettings {
	return teamsettings.CoreTeamSprintSettings{
		TeamID:                       teamID,
		WorkspaceID:                  workspaceID,
		AutoCreateSprints:            autoCreateSprints,
		UpcomingSprintsCount:         int(upcomingSprintsCount),
		SprintDurationWeeks:          int(sprintDurationWeeks),
		SprintStartDay:               sprintStartDay,
		WorkingDays:                  smallIntsToInts(workingDays),
		MoveIncompleteStoriesEnabled: moveIncompleteStoriesEnabled,
		LastAutoSprintNumber:         int(lastAutoSprintNumber),
		NextAutoSprintNumber:         int(nextAutoSprintNumber),
		AutoCreateDisabledAt:         autoCreateDisabledAt,
		AutoCreateDisabledReason:     autoCreateDisabledReason,
		CreatedAt:                    createdAt,
		UpdatedAt:                    updatedAt,
	}
}

func smallIntsToInts(values []int16) []int {
	result := make([]int, len(values))
	for index, value := range values {
		result[index] = int(value)
	}
	return result
}

func intsToSmallInts(values []int) ([]int16, error) {
	result := make([]int16, len(values))
	for index, value := range values {
		converted, err := safecast.Int16(value)
		if err != nil {
			return nil, teamsettings.ErrInvalidWorkingDays
		}
		result[index] = converted
	}
	return result, nil
}
