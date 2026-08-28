package teamsettingsdomain

import "errors"

var (
	ErrInvalidSprintStartDay  = errors.New("sprint start day must be Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, or Sunday")
	ErrInvalidSprintDuration  = errors.New("sprint duration must be between 1 and 8 weeks")
	ErrInvalidWorkingDays     = errors.New("working days must contain unique weekday numbers between 1 and 7")
	ErrInvalidUpcomingCount   = errors.New("upcoming sprints count must be between 0 and 10")
	ErrInvalidNextAutoNumber  = errors.New("next auto sprint number must be between 1 and 10000")
	ErrInvalidCloseMonths     = errors.New("auto-close inactive months must be between 1 and 24")
	ErrInvalidArchiveMonths   = errors.New("auto-archive months must be between 1 and 24")
	ErrInvalidEstimateScheme  = errors.New("estimate scheme must be one of: points, tshirt")
	ErrSprintScheduleConflict = errors.New("an upcoming sprint with custom dates conflicts with the automated schedule")
	ErrTeamSettingsNotFound   = errors.New("team settings not found")
	ErrTeamMembershipRequired = errors.New("active team membership is required")
	ErrNoSettingsChanges      = errors.New("at least one team setting must be provided")
	ErrConcurrentUpdate       = errors.New("team settings changed concurrently; retry the operation")
)
