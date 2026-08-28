package maya

import (
	"errors"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
)

var (
	ErrInvalidPlanInput = mayadomain.ErrInvalidPlanInput
	ErrMissingDuration  = errors.New("story duration is required for scheduling")
)

const (
	automaticMinimumFocusBlockMinutes = 15
	planningStartGranularityMinutes   = 5
	maxFocusBlockMinutes              = 120
	recentActivityDays                = 30
)

type Planner struct {
	advisor CandidateAdvisor
}

func NewPlanner() Planner {
	return Planner{}
}

func NewPlannerWithAdvisor(advisor CandidateAdvisor) Planner {
	return Planner{advisor: advisor}
}
