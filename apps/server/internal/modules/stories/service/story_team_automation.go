package stories

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// UpdateTeamAutomationIfUnchanged preserves the inspected version and the
// current human owner's authorization. The durable event remains internal so
// a reusable rule cannot trigger itself or another rule in a feedback loop.
func (s *Service) UpdateTeamAutomationIfUnchanged(ctx context.Context, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error {
	actorID, err := auth.GetUserID(ctx)
	if err != nil || expectedUpdatedAt.IsZero() {
		return ErrStoryMutationForbidden
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		expectedUpdatedAt: &expectedUpdatedAt, activityReason: reason, recordDescriptionUpdates: true,
		publishEvents: true, actorKind: auth.PrincipalHumanUser, mutationEventDelivery: mutationEventDeliveryInternalOnly,
	})
}
