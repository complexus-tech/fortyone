package notifications

import (
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	"github.com/google/uuid"
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
	ID                    uuid.UUID
	Scope                 DeliveryScope
	NotificationIDs       []uuid.UUID
	NotificationSnapshots []EmailSnapshot
	FeedbackDigest        *feedback.CoreDigestDeliveryCompletion
	GuidanceDate          *time.Time
	Sent                  bool
	Now                   time.Time
}
