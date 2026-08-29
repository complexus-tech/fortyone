package feedback

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type publicationOutboxRepoStub struct {
	*nextPhaseRepoStub
	claimed                        []CorePublicationOutboxEvent
	deliveryRecipients             []CoreDeliveryRecipient
	accountRecipients              []CoreAccountUpdateRecipient
	resolvedActorID                uuid.UUID
	deliveryInputs                 []CoreCreateDeliveryInput
	completedEventID               uuid.UUID
	completedClaimToken            uuid.UUID
	completeErr                    error
	retriedEventID                 uuid.UUID
	retriedClaimToken              uuid.UUID
	retryFailure                   string
	retryAt                        time.Time
	retryTerminal                  bool
	publicationClaimed             bool
	deliveryWasCreated             bool
	mergeClaimCalls                int
	mergeClaimed                   bool
	claimedMerges                  []CoreMergeOutboxEvent
	mergeRecipients                []CoreMergeRecipient
	completedMergeID               uuid.UUID
	completedMergeToken            uuid.UUID
	retriedMergeID                 uuid.UUID
	retriedMergeToken              uuid.UUID
	mergeRetryFailure              string
	mergeRetryAt                   time.Time
	mergeRetryTerminal             bool
	mergeTargetItemID              uuid.UUID
	mergeFollowerIDs               []uuid.UUID
	publicationContributorAudience []uuid.UUID
	publicationItemIDs             []uuid.UUID
	publicationAccountAudience     []CoreAccountUpdateRecipient
}

func (r *publicationOutboxRepoStub) ClaimPublicationOutboxEvents(_ context.Context, _ int, _ time.Duration) ([]CorePublicationOutboxEvent, error) {
	if r.publicationClaimed {
		return []CorePublicationOutboxEvent{}, nil
	}
	r.publicationClaimed = true
	return r.claimed, nil
}

func (r *publicationOutboxRepoStub) CompletePublicationOutboxEvent(_ context.Context, eventID, claimToken uuid.UUID) error {
	r.completedEventID = eventID
	r.completedClaimToken = claimToken
	return r.completeErr
}

func (r *publicationOutboxRepoStub) RetryPublicationOutboxEvent(_ context.Context, eventID, claimToken uuid.UUID, failure string, retryAt time.Time, terminal bool) error {
	r.retriedEventID = eventID
	r.retriedClaimToken = claimToken
	r.retryFailure = failure
	r.retryAt = retryAt
	r.retryTerminal = terminal
	return nil
}

func (r *publicationOutboxRepoStub) ListPublicationDeliveryRecipients(_ context.Context, _ uuid.UUID, contributorIDs, itemIDs []uuid.UUID) ([]CoreDeliveryRecipient, error) {
	r.publicationContributorAudience = append([]uuid.UUID(nil), contributorIDs...)
	r.publicationItemIDs = append([]uuid.UUID(nil), itemIDs...)
	return r.deliveryRecipients, nil
}

func (r *publicationOutboxRepoStub) ListAccountPublicationRecipients(_ context.Context, _ uuid.UUID, audience []CoreAccountUpdateRecipient) ([]CoreAccountUpdateRecipient, error) {
	r.publicationAccountAudience = append([]CoreAccountUpdateRecipient(nil), audience...)
	return r.accountRecipients, nil
}

func (r *publicationOutboxRepoStub) ResolveNotificationActor(_ context.Context, actorID, fallbackID uuid.UUID) (uuid.UUID, error) {
	if r.resolvedActorID != uuid.Nil {
		return r.resolvedActorID, nil
	}
	return actorID, nil
}

func (r *publicationOutboxRepoStub) CreateContributorDelivery(_ context.Context, input CoreCreateDeliveryInput) (CoreDelivery, bool, error) {
	r.deliveryInputs = append(r.deliveryInputs, input)
	return CoreDelivery{ID: input.DeliveryID}, r.deliveryWasCreated, nil
}

func (r *publicationOutboxRepoStub) ClaimMergeOutboxEvents(_ context.Context, _ int, _ time.Duration) ([]CoreMergeOutboxEvent, error) {
	r.mergeClaimCalls++
	if r.mergeClaimed {
		return []CoreMergeOutboxEvent{}, nil
	}
	r.mergeClaimed = true
	return r.claimedMerges, nil
}

func (r *publicationOutboxRepoStub) ListMergeRecipients(_ context.Context, _ uuid.UUID, targetItemID uuid.UUID, followerIDs []uuid.UUID) ([]CoreMergeRecipient, error) {
	r.mergeTargetItemID = targetItemID
	r.mergeFollowerIDs = append([]uuid.UUID(nil), followerIDs...)
	return r.mergeRecipients, nil
}

func (r *publicationOutboxRepoStub) CompleteMergeOutboxEvent(_ context.Context, eventID, claimToken uuid.UUID) error {
	r.completedMergeID = eventID
	r.completedMergeToken = claimToken
	return nil
}

func (r *publicationOutboxRepoStub) RetryMergeOutboxEvent(_ context.Context, eventID, claimToken uuid.UUID, failure string, retryAt time.Time, terminal bool) error {
	r.retriedMergeID = eventID
	r.retriedMergeToken = claimToken
	r.mergeRetryFailure = failure
	r.mergeRetryAt = retryAt
	r.mergeRetryTerminal = terminal
	return nil
}

type publicationTaskStub struct {
	payloads []tasks.FeedbackContributorDeliveryPayload
	err      error
}

func (t *publicationTaskStub) EnqueueFeedbackContributorDelivery(payload tasks.FeedbackContributorDeliveryPayload) error {
	t.payloads = append(t.payloads, payload)
	return t.err
}

type publicationPublisherStub struct {
	events []events.Event
	err    error
}

func (p *publicationPublisherStub) Publish(_ context.Context, event events.Event) error {
	p.events = append(p.events, event)
	return p.err
}

func TestDispatchReadyPublicationEventsPersistsDedupedSideEffectsBeforeCompleting(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, snapshot := publicationEventFixture(t, now.Add(-time.Minute), 2)
	guestID, accountID, itemID := snapshot.ContributorAudience[0], snapshot.AccountAudience[0].UserID, snapshot.LinkedItems[0].ID
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub:  &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:            []CorePublicationOutboxEvent{publication},
		deliveryRecipients: []CoreDeliveryRecipient{{ContributorID: guestID, Kind: ContributorKindVerifiedGuest}},
		accountRecipients:  []CoreAccountUpdateRecipient{{UserID: accountID, ItemID: itemID}},
		deliveryWasCreated: true,
	}
	taskService := &publicationTaskStub{}
	publisher := &publicationPublisherStub{}
	service := New(repository, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", taskService),
		WithGuestNotificationActor(uuid.New()),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyPublicationEvents(context.Background()))
	require.Equal(t, snapshot.ContributorAudience, repository.publicationContributorAudience)
	require.Equal(t, []uuid.UUID{itemID}, repository.publicationItemIDs)
	require.Equal(t, []CoreAccountUpdateRecipient{{UserID: accountID, ItemID: itemID}}, repository.publicationAccountAudience)
	require.Len(t, repository.deliveryInputs, 1)
	delivery := repository.deliveryInputs[0]
	require.Equal(t, stablePublicationDeliveryID(publication.EventID, guestID), delivery.DeliveryID)
	require.Equal(t, "update-publication:"+publication.EventID.String()+":"+guestID.String(), delivery.DedupeKey)
	require.Equal(t, "https://app.example.com/portal/roads/updates/dark-mode-shipped", delivery.DestinationURL)
	require.NotNil(t, delivery.ItemID)
	require.Equal(t, itemID, *delivery.ItemID)
	require.Len(t, taskService.payloads, 1)
	require.Equal(t, delivery.DeliveryID, taskService.payloads[0].DeliveryID)

	require.Len(t, publisher.events, 1)
	accountEvent := publisher.events[0]
	require.Equal(t, events.FeedbackUpdatePublished, accountEvent.Type)
	require.Equal(t, publication.PublishedAt, accountEvent.Timestamp)
	payload := accountEvent.Payload.(events.FeedbackUpdatePublishedPayload)
	require.Equal(t, publication.EventID, payload.PublicationEventID)
	require.Equal(t, publication.PublicationSequence, payload.PublicationSequence)
	require.Equal(t, accountID, payload.RecipientID)
	require.Equal(t, itemID, payload.LinkedItemID)

	require.Equal(t, publication.EventID, repository.completedEventID)
	require.Equal(t, publication.ClaimToken, repository.completedClaimToken)
	require.Equal(t, uuid.Nil, repository.retriedEventID)
}

func TestDispatchReadyOutboxEventsAttemptsPublicationAndMerge(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, _ := publicationEventFixture(t, now.Add(-time.Minute), 1)
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:           []CorePublicationOutboxEvent{publication},
		completeErr:       errors.New("publication completion uncertain"),
	}
	service := New(repository, nil,
		WithEventPublisher(nil, &publicationPublisherStub{}),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", &publicationTaskStub{}),
	)
	service.now = func() time.Time { return now }

	err := service.DispatchReadyOutboxEvents(context.Background())
	require.ErrorContains(t, err, "publication completion uncertain")
	require.Equal(t, 1, repository.mergeClaimCalls, "publication failures must not starve the merge outbox")
}

func TestDispatchReadyPublicationEventsRetriesAfterDownstreamFailure(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, snapshot := publicationEventFixture(t, now.Add(-time.Minute), 2)
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:           []CorePublicationOutboxEvent{publication},
		accountRecipients: []CoreAccountUpdateRecipient{{UserID: uuid.New(), ItemID: snapshot.LinkedItems[0].ID}},
	}
	publisher := &publicationPublisherStub{err: errors.New("redis unavailable")}
	service := New(repository, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", &publicationTaskStub{}),
		WithGuestNotificationActor(uuid.New()),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyPublicationEvents(context.Background()))
	require.Equal(t, uuid.Nil, repository.completedEventID)
	require.Equal(t, publication.EventID, repository.retriedEventID)
	require.Equal(t, publication.ClaimToken, repository.retriedClaimToken)
	require.Contains(t, repository.retryFailure, "redis unavailable")
	require.Equal(t, now.Add(time.Minute), repository.retryAt)
	require.False(t, repository.retryTerminal)
}

func TestDispatchReadyPublicationEventsFailsMalformedSnapshotPermanently(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, _ := publicationEventFixture(t, now.Add(-time.Minute), 1)
	publication.Payload = json.RawMessage(`{"schemaVersion":1}`)
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:           []CorePublicationOutboxEvent{publication},
	}
	service := New(repository, nil,
		WithEventPublisher(nil, &publicationPublisherStub{}),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", &publicationTaskStub{}),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyPublicationEvents(context.Background()))
	require.Equal(t, publication.EventID, repository.retriedEventID)
	require.True(t, repository.retryTerminal)
	require.Empty(t, repository.deliveryInputs)
	require.Equal(t, uuid.Nil, repository.completedEventID)
}

func TestDispatchReadyPublicationEventsFailsSnapshotWithoutImmutableAudience(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, snapshot := publicationEventFixture(t, now.Add(-time.Minute), 1)
	snapshot.ContributorAudience = nil
	payload, err := json.Marshal(snapshot)
	require.NoError(t, err)
	publication.Payload = payload
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:           []CorePublicationOutboxEvent{publication},
	}
	service := New(repository, nil,
		WithEventPublisher(nil, &publicationPublisherStub{}),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", &publicationTaskStub{}),
	)
	service.now = func() time.Time { return now }

	require.NoError(t, service.dispatchReadyPublicationEvents(context.Background()))
	require.Equal(t, publication.EventID, repository.retriedEventID)
	require.True(t, repository.retryTerminal)
	require.Contains(t, repository.retryFailure, "immutable audience")
	require.Empty(t, repository.deliveryInputs)
}

func TestDispatchReadyPublicationEventsDoesNotOverwriteUncertainCompletion(t *testing.T) {
	now := time.Date(2026, time.August, 13, 14, 0, 0, 0, time.UTC)
	publication, _ := publicationEventFixture(t, now.Add(-time.Minute), 1)
	repository := &publicationOutboxRepoStub{
		nextPhaseRepoStub: &nextPhaseRepoStub{repoStub: &repoStub{}},
		claimed:           []CorePublicationOutboxEvent{publication},
		completeErr:       errors.New("commit outcome unknown"),
	}
	service := New(repository, nil,
		WithEventPublisher(nil, &publicationPublisherStub{}),
		WithContributorFeatures("publication-test-secret", "https://app.example.com", &publicationTaskStub{}),
	)
	service.now = func() time.Time { return now }

	err := service.dispatchReadyPublicationEvents(context.Background())
	require.ErrorContains(t, err, "commit outcome unknown")
	require.Equal(t, publication.EventID, repository.completedEventID)
	require.Equal(t, uuid.Nil, repository.retriedEventID, "uncertain completion must be recovered as a stale claim")
}

func TestPublicationRetryDelayIsBounded(t *testing.T) {
	t.Parallel()
	require.Equal(t, 30*time.Second, publicationRetryDelay(1))
	require.Equal(t, time.Minute, publicationRetryDelay(2))
	require.Equal(t, publicationOutboxMaxDelay, publicationRetryDelay(100))
}

func publicationEventFixture(t *testing.T, publishedAt time.Time, attempt int) (CorePublicationOutboxEvent, publicationSnapshot) {
	t.Helper()
	event := CorePublicationOutboxEvent{
		EventID: uuid.New(), UpdateID: uuid.New(), WorkspaceID: uuid.New(), PortalID: uuid.New(),
		ActorID: uuid.New(), PublicationSequence: 3, PublishedAt: publishedAt,
		ClaimToken: uuid.New(), AttemptCount: attempt,
	}
	summary := "The requested dark mode is now available."
	itemID, contributorID, accountID := uuid.New(), uuid.New(), uuid.New()
	snapshot := publicationSnapshot{
		SchemaVersion: 1, PublicationEventID: event.EventID, UpdateID: event.UpdateID,
		WorkspaceID: event.WorkspaceID, PortalID: event.PortalID, PortalSlug: "roads",
		PublishedByUserID: event.ActorID, PublicationSequence: event.PublicationSequence,
		PublishedAt: event.PublishedAt, Slug: "dark-mode-shipped", Title: "Dark mode shipped",
		Summary: &summary, Body: "Dark mode is live.",
		LinkedItems:         []CoreUpdateItem{{ID: itemID, Slug: "dark-mode", Title: "Dark mode", Status: StatusCompleted}},
		ContributorAudience: []uuid.UUID{contributorID},
		AccountAudience:     []publicationAccountPair{{UserID: accountID, ItemID: itemID}},
	}
	payload, err := json.Marshal(snapshot)
	require.NoError(t, err)
	event.Payload = payload
	return event, snapshot
}
