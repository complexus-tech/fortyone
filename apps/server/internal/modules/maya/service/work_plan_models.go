package maya

import (
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/google/uuid"
)

type CreateWorkPlanInput struct {
	WorkspaceID      uuid.UUID   `json:"workspaceId"`
	StoryID          uuid.UUID   `json:"storyId"`
	TriggeredBy      uuid.UUID   `json:"triggeredBy"`
	Trigger          RunTrigger  `json:"trigger"`
	WindowStart      time.Time   `json:"windowStart"`
	WindowEnd        time.Time   `json:"windowEnd"`
	DurationMinutes  int         `json:"durationMinutes"`
	CandidateUserIDs []uuid.UUID `json:"candidateUserIds"`
	AutoApply        bool        `json:"autoApply"`
	AssignmentReason string      `json:"-"`
}

type ApplyWorkPlanInput struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	RunID       uuid.UUID `json:"runId"`
	TriggeredBy uuid.UUID `json:"triggeredBy"`
}

type persistedWorkPlanContext struct {
	DurationMinutes    int               `json:"durationMinutes"`
	StoryUpdatedAt     time.Time         `json:"storyUpdatedAt"`
	CandidateTimezones map[string]string `json:"candidateTimezones"`
}

type WorkPlan = mayadomain.WorkPlan

type ProcessAssignmentBatchInput struct {
	WorkspaceID     uuid.UUID
	TeamID          uuid.UUID
	StoryIDs        []uuid.UUID
	TriggeredBy     uuid.UUID
	WindowStart     time.Time
	WindowEnd       time.Time
	DurationMinutes int
	AutoApply       bool
}

type ProcessAssignmentBatchResult struct {
	Processed int
	Skipped   int
	Plans     []WorkPlan
}
