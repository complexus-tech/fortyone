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

func TestStartOutboundDeliveryRejectsMalformedProviderPayload(t *testing.T) {
	t.Parallel()

	repo := &Repository{}
	_, _, err := repo.StartOutboundDelivery(context.Background(), OutboundDeliveryInput{
		Provider:            "slack",
		WorkspaceID:         uuid.New(),
		ExternalWorkspaceID: "T1",
		IdempotencyKey:      "slack:test",
		ProviderPayload:     []byte(`{"blocks":`),
	})

	require.ErrorContains(t, err, "provider payload must be valid JSON")
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

func TestUpsertChannelConversationRequiresAudienceFingerprint(t *testing.T) {
	t.Parallel()

	repo := &Repository{}
	_, err := repo.UpsertConversation(context.Background(), ConversationInput{
		Provider:            "slack",
		WorkspaceID:         uuid.New(),
		ExternalWorkspaceID: "T1",
		ExternalChannelID:   "C1",
		ExternalThreadID:    "10.1",
		UserID:              uuid.New(),
		AudienceScope:       ConversationAudienceChannel,
	})

	require.ErrorContains(t, err, "audience fingerprint is required")
}

func TestUpsertActorConversationRejectsAudienceFingerprint(t *testing.T) {
	t.Parallel()

	repo := &Repository{}
	_, err := repo.UpsertConversation(context.Background(), ConversationInput{
		Provider:            "slack",
		WorkspaceID:         uuid.New(),
		ExternalWorkspaceID: "T1",
		ExternalChannelID:   "D1",
		ExternalThreadID:    "dm:D1",
		UserID:              uuid.New(),
		AudienceScope:       ConversationAudienceActor,
		AudienceFingerprint: "v1:not-valid-for-an-actor",
	})

	require.ErrorContains(t, err, "actor conversation cannot have an audience fingerprint")
}
