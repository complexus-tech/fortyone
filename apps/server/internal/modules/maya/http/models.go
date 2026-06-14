package mayahttp

import (
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

type AppCreateWorkPlanRequest struct {
	StoryID          uuid.UUID   `json:"storyId"`
	WindowStart      *time.Time  `json:"windowStart"`
	WindowEnd        *time.Time  `json:"windowEnd"`
	DurationMinutes  int         `json:"durationMinutes"`
	CandidateUserIDs []uuid.UUID `json:"candidateUserIds"`
	AutoApply        bool        `json:"autoApply"`
}

type AppWorkPlan struct {
	Run     maya.CoreRun      `json:"run"`
	Actions []maya.CoreAction `json:"actions"`
}

func toAppWorkPlan(plan maya.WorkPlan) AppWorkPlan {
	return AppWorkPlan{
		Run:     plan.Run,
		Actions: plan.Actions,
	}
}
