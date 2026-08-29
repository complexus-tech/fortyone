package maya

import (
	"time"

	"github.com/google/uuid"
)

type ReconcileScheduleInput struct {
	WorkspaceID *uuid.UUID
	UserID      *uuid.UUID
	StoryID     *uuid.UUID
}

const (
	scheduleRecoveryRetryDelay      = 5 * time.Minute
	interruptedScheduleRunStaleness = 10 * time.Minute
)
