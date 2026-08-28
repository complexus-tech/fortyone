package teamsettingsrepository

import (
	"errors"
	"math"
	"testing"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/google/uuid"
)

func TestSprintUpdateParamsRejectsOverflow(t *testing.T) {
	t.Parallel()

	overflow := int(int64(math.MaxInt32) + 1)
	updates := teamsettings.CoreUpdateTeamSprintSettings{
		UpcomingSprintsCount: teamsettings.PatchField[int]{Value: overflow, Present: true},
	}
	_, err := sprintUpdateParams(uuid.New(), uuid.New(), updates)
	if !errors.Is(err, teamsettings.ErrInvalidUpcomingCount) {
		t.Fatalf("sprintUpdateParams() error = %v, want ErrInvalidUpcomingCount", err)
	}
}

func TestStoryAutomationUpdateParamsRejectsOverflow(t *testing.T) {
	t.Parallel()

	updates := teamsettings.CoreUpdateTeamStoryAutomationSettings{
		AutoArchiveMonths: teamsettings.PatchField[int]{
			Value: int(int64(math.MaxInt32) + 1), Present: true,
		},
	}
	_, err := storyAutomationUpdateParams(uuid.New(), uuid.New(), updates)
	if !errors.Is(err, teamsettings.ErrInvalidArchiveMonths) {
		t.Fatalf("storyAutomationUpdateParams() error = %v, want ErrInvalidArchiveMonths", err)
	}
}

func TestWorkingDaysConversionRejectsSmallIntOverflow(t *testing.T) {
	t.Parallel()

	_, err := intsToSmallInts([]int{math.MaxInt16 + 1})
	if !errors.Is(err, teamsettings.ErrInvalidWorkingDays) {
		t.Fatalf("intsToSmallInts() error = %v, want ErrInvalidWorkingDays", err)
	}
}
