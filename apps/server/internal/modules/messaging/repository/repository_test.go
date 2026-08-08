package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStartOutboundDeliveryRequiresExternalWorkspaceBinding(t *testing.T) {
	t.Parallel()

	repo := &Repository{}
	_, _, err := repo.StartOutboundDelivery(context.Background(), OutboundDeliveryInput{
		Provider:       "slack",
		WorkspaceID:    uuid.New(),
		IdempotencyKey: "slack:test",
	})

	require.ErrorContains(t, err, "external workspace id is required")
}

func TestLeaseBusyErrorSupportsRetryThroughWrapping(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("process provider event: %w", newLeaseBusyError("messaging inbound event"))

	require.ErrorIs(t, err, ErrLeaseBusy)
	retryAfter, ok := LeaseRetryAfter(err)
	require.True(t, ok)
	require.Equal(t, messagingLeaseDuration+messagingLeaseRetryMargin, retryAfter)
}

func TestLeaseRetryAfterUsesSafeDefault(t *testing.T) {
	t.Parallel()

	retryAfter, ok := LeaseRetryAfter(&LeaseBusyError{Resource: "messaging outbound delivery"})

	require.True(t, ok)
	require.Equal(t, messagingLeaseDuration+messagingLeaseRetryMargin, retryAfter)
}

func TestLeaseRetryAfterRejectsUnrelatedError(t *testing.T) {
	t.Parallel()

	_, ok := LeaseRetryAfter(errors.New("provider unavailable"))

	require.False(t, ok)
}
