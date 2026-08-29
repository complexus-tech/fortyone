package feedback

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDispatchReadyMergeEventsUsesImmutableAudienceAndDedupesRecipients(t *testing.T) {
	now := time.Date(2026, time.August, 13, 16, 0, 0, 0, time.UTC)
	workspaceID, portalID := uuid.New(), uuid.New()
	merge, followerIDs := mergeEventFixture(t, now.Add(-time.Minute), workspaceID, portalID, 1)
	accountID, deletedActorFallbackID := uuid.New(), uuid.New()
	guestID, externalID := followerIDs[3], followerIDs[4]
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{portals: []CorePortal{{
			ID: portalID, WorkspaceID: workspaceID, Slug: "roads",
		}}}},
		claimedMerges: []CoreMergeOutboxEvent{merge},
		mergeRecipients: []CoreMergeRecipient{
			{ContributorID: followerIDs[0], UserID: accountID, Kind: ContributorKindAccount},
			{ContributorID: followerIDs[1], UserID: accountID, Kind: ContributorKindAccount},
			{ContributorID: followerIDs[2], UserID: merge.ActorID, Kind: ContributorKindAccount},
			{ContributorID: guestID, Kind: ContributorKindVerifiedGuest},
			{ContributorID: externalID, Kind: ContributorKindExternal},
		},
		resolvedActorID:    deletedActorFallbackID,
		deliveryWasCreated: true,
	}
	taskService := &publicationTaskStub{}
	publisher := &publicationPublisherStub{}
	service := New(repository, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("merge-test-secret", "https://app.example.com", taskService),
		WithGuestNotificationActor(deletedActorFallbackID),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyMergeEvents(context.Background()))
	require.Equal(t, merge.EventID, repository.completedMergeID)
	require.Equal(t, merge.ClaimToken, repository.completedMergeToken)
	require.Equal(t, uuid.Nil, repository.retriedMergeID)
	require.Equal(t, merge.TargetItemID, repository.mergeTargetItemID)
	require.Equal(t, followerIDs, repository.mergeFollowerIDs)

	require.Len(t, publisher.events, 1, "duplicate account contributors and the actor must be suppressed")
	accountEvent := publisher.events[0]
	require.Equal(t, events.FeedbackItemMerged, accountEvent.Type)
	require.Equal(t, deletedActorFallbackID, accountEvent.ActorID)
	require.Equal(t, merge.MergedAt, accountEvent.Timestamp)
	payload := accountEvent.Payload.(events.FeedbackItemMergedPayload)
	require.Equal(t, accountID, payload.RecipientID)
	require.Equal(t, merge.TargetItemID, payload.TargetItemID)
	require.Equal(t, "Canonical dark mode", payload.TargetItemTitle)

	require.Len(t, repository.deliveryInputs, 2)
	require.Len(t, taskService.payloads, 2)
	deliveryIDs := make(map[uuid.UUID]CoreCreateDeliveryInput, 2)
	for _, input := range repository.deliveryInputs {
		deliveryIDs[input.ContributorID] = input
		require.Equal(t, "item-merge:"+merge.EventID.String(), input.DedupeKey)
		require.Equal(t, "feedback.item.merged", input.EventType)
		require.Equal(t, "https://app.example.com/portal/roads/feedback/canonical-dark-mode", input.DestinationURL)
		require.NotNil(t, input.ItemID)
		require.Equal(t, merge.TargetItemID, *input.ItemID)
	}
	require.Equal(t, stableMergeDeliveryID(merge.EventID, guestID), deliveryIDs[guestID].DeliveryID)
	require.Equal(t, stableMergeDeliveryID(merge.EventID, externalID), deliveryIDs[externalID].DeliveryID)
}

func TestDispatchReadyMergeEventsPermanentlyFailsMalformedSnapshot(t *testing.T) {
	now := time.Date(2026, time.August, 13, 16, 0, 0, 0, time.UTC)
	workspaceID, portalID := uuid.New(), uuid.New()
	merge, _ := mergeEventFixture(t, now.Add(-time.Minute), workspaceID, portalID, 1)
	merge.Payload = json.RawMessage(`{"schemaVersion":1}`)
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimedMerges:     []CoreMergeOutboxEvent{merge},
	}
	service := New(repository, nil,
		WithEventPublisher(nil, &publicationPublisherStub{}),
		WithContributorFeatures("merge-test-secret", "https://app.example.com", &publicationTaskStub{}),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyMergeEvents(context.Background()))
	require.Equal(t, merge.EventID, repository.retriedMergeID)
	require.Equal(t, merge.ClaimToken, repository.retriedMergeToken)
	require.True(t, repository.mergeRetryTerminal)
	require.Contains(t, repository.mergeRetryFailure, "does not match")
	require.Equal(t, uuid.Nil, repository.completedMergeID)
}

func mergeEventFixture(t *testing.T, mergedAt time.Time, workspaceID, portalID uuid.UUID, attempt int) (CoreMergeOutboxEvent, []uuid.UUID) {
	t.Helper()
	followerIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	merge := CoreMergeOutboxEvent{
		EventID: uuid.New(), WorkspaceID: workspaceID, PortalID: portalID,
		SourceItemID: uuid.New(), TargetItemID: uuid.New(), ActorID: uuid.New(),
		MergedAt: mergedAt, ClaimToken: uuid.New(), AttemptCount: attempt,
	}
	payload, err := json.Marshal(mergeSnapshot{
		SchemaVersion: 1, MergeEventID: merge.EventID, WorkspaceID: workspaceID, PortalID: portalID,
		SourceItemID: merge.SourceItemID, TargetItemID: merge.TargetItemID,
		MergedByUserID: merge.ActorID, MergedAt: mergedAt,
		SourceTitle: "Duplicate dark mode", SourceSlug: "duplicate-dark-mode",
		TargetTitle: "Canonical dark mode", TargetSlug: "canonical-dark-mode",
		SourceFollowerIDs: followerIDs,
	})
	require.NoError(t, err)
	merge.Payload = payload
	return merge, followerIDs
}
