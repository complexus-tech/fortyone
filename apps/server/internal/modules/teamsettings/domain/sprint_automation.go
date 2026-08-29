package teamsettingsdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSprintAutomationQuery = errors.New("invalid sprint automation query")
	ErrSprintScheduleTooLarge       = errors.New("sprint automation schedule exceeds the supported bound")
)

// SprintAutomationCursor is the stable workspace/team keyset used by worker
// scans. Valid distinguishes the zero cursor from the first page.
type SprintAutomationCursor struct {
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
	Valid       bool
}

// SprintAutomationQuery requests one bounded page of teams whose durable
// settings currently enable sprint creation.
type SprintAutomationQuery struct {
	Cursor    SprintAutomationCursor
	BatchSize int
}

type SprintAutomationTeamRef struct {
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
}

// SprintAutomationRunResult describes changes committed atomically for one
// team. A retry after commit returns zero creations once the target is full.
type SprintAutomationRunResult struct {
	Rescheduled int
	Created     int
}

// SprintAutomationInactivityQuery pages old, still-enabled settings before
// the repository performs the more expensive team-scoped activity recheck.
type SprintAutomationInactivityQuery struct {
	TeamCreatedBefore     time.Time
	SettingsUpdatedBefore time.Time
	Cursor                SprintAutomationCursor
	BatchSize             int
}

// SprintAutomationInactivityEligibility carries every application-captured
// cutoff used to conditionally disable one team.
type SprintAutomationInactivityEligibility struct {
	WorkspaceID           uuid.UUID
	TeamID                uuid.UUID
	TeamCreatedBefore     time.Time
	SettingsUpdatedBefore time.Time
	ActivityBefore        time.Time
	DisabledAt            time.Time
	Reason                string
	InactivityDays        int
	GraceDays             int
}
