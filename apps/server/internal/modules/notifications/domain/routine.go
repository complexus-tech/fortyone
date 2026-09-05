package notifications

import (
	"github.com/google/uuid"
	"time"
)

// Routine delivery claims serialize activity batches and briefings for a person,
// while keeping the content and reply authorization scoped to one workspace.
type RoutineRecipient struct {
	UserID        uuid.UUID
	WorkspaceID   uuid.UUID
	Email         string
	Name          string
	WorkspaceName string
	WorkspaceSlug string
	Timezone      string
	WeeklyEnabled bool
}
type RoutineClaim struct {
	RecipientID uuid.UUID
	WorkspaceID uuid.UUID
	Key         string
	Kind        string
	LocalDate   time.Time
	Now         time.Time
}
type RoutineCompletion struct {
	ID              uuid.UUID
	Scope           DeliveryScope
	NotificationIDs []uuid.UUID
	GuidanceDate    *time.Time
	Sent            bool
	Now             time.Time
}
