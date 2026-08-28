package notifications

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WeeklyDigestCursor identifies the last workspace-recipient pair processed by
// the weekly-digest worker. The repository uses the same pair as its stable
// ordering key.
type WeeklyDigestCursor struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

// WeeklyDigestRecipient contains the current delivery identity for one active
// workspace member. It deliberately excludes notification and work details;
// those are loaded again through the scoped stats query immediately before
// delivery.
type WeeklyDigestRecipient struct {
	UserID        uuid.UUID
	UserEmail     string
	UserName      string
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceSlug string
}

// WeeklyDigestStats is the bounded aggregate used to render one workspace
// digest. All values are evaluated against the same explicit UTC as-of time.
type WeeklyDigestStats struct {
	UnreadNotifications         int
	UnreadPriorityNotifications int
	OverdueStories              int
	DueThisWeekStories          int
	ObjectiveRisks              int
	TeamComments                int
}

// WeeklyDigestStatsQuery carries the complete tenant and time scope for one
// aggregate read.
type WeeklyDigestStatsQuery struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	AsOf        time.Time
}

func (query WeeklyDigestStatsQuery) Validate() error {
	if query.UserID == uuid.Nil || query.WorkspaceID == uuid.Nil {
		return fmt.Errorf("%w: weekly digest user and workspace are required", ErrInvalid)
	}
	if query.AsOf.IsZero() {
		return fmt.Errorf("%w: weekly digest as-of time is required", ErrInvalid)
	}
	return nil
}
