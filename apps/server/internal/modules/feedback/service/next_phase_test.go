package feedback

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// nextPhaseRepoStub embeds the full next-phase contract so individual tests
// only need to implement the repository calls they exercise.
type nextPhaseRepoStub struct {
	*repoStub
	NextPhaseRepository

	verificationInput    CoreVerificationRequest
	itemInput            CoreContributorItemInput
	commentInput         CoreContributorCommentInput
	widgetSettings       CoreWidgetSettings
	widgetSettingsInput  CoreWidgetSettingsInput
	widgetSecret         string
	nonceErr             error
	nonceCalls           int
	externalEmail        string
	sessionParticipant   CoreParticipant
	session              CoreParticipantSession
	publishedUpdate      CoreFeedbackUpdate
	newlyPublished       bool
	accountRecipients    []CoreAccountUpdateRecipient
	accountItemFollowers []uuid.UUID
	primaryStoryItems    []CoreItem
	itemCandidates       CoreMergeCandidatesPage
	candidateWorkspaceID uuid.UUID
	candidatePortalID    uuid.UUID
	candidateExcludedID  uuid.UUID
	candidateSearch      string
	candidateLimit       int
}

func (r *nextPhaseRepoStub) PublishUpdate(_ context.Context, workspaceID, updateID, actorID uuid.UUID) (CoreFeedbackUpdate, bool, error) {
	return r.publishedUpdate, r.newlyPublished, nil
}

func (r *nextPhaseRepoStub) ListAccountUpdateRecipients(_ context.Context, portalID, updateID uuid.UUID) ([]CoreAccountUpdateRecipient, error) {
	return r.accountRecipients, nil
}

func (r *nextPhaseRepoStub) ListAccountItemFollowers(_ context.Context, portalID, itemID uuid.UUID) ([]uuid.UUID, error) {
	return r.accountItemFollowers, nil
}

func (r *nextPhaseRepoStub) ListPrimaryStoryItems(_ context.Context, workspaceID, storyID uuid.UUID) ([]CoreItem, error) {
	return r.primaryStoryItems, nil
}

func (r *nextPhaseRepoStub) ListItemCandidates(_ context.Context, workspaceID, portalID, excludedItemID uuid.UUID, search string, limit int) (CoreMergeCandidatesPage, error) {
	r.candidateWorkspaceID = workspaceID
	r.candidatePortalID = portalID
	r.candidateExcludedID = excludedItemID
	r.candidateSearch = search
	r.candidateLimit = limit
	return r.itemCandidates, nil
}

func (r *nextPhaseRepoStub) GetParticipantByUser(_ context.Context, portalID, userID uuid.UUID) (CoreParticipant, error) {
	return CoreParticipant{}, ErrNotFound
}

func (r *nextPhaseRepoStub) GetContributorSession(_ context.Context, portalID uuid.UUID, tokenHash []byte, source string) (CoreParticipant, CoreParticipantSession, error) {
	if r.session.PortalID != portalID {
		return CoreParticipant{}, CoreParticipantSession{}, ErrNotFound
	}
	return r.sessionParticipant, r.session, nil
}

func (r *nextPhaseRepoStub) CreateContributorVerification(_ context.Context, input CoreVerificationRequest) (CoreVerificationChallenge, error) {
	r.verificationInput = input
	return CoreVerificationChallenge{ID: uuid.New(), ExpiresAt: input.ExpiresAt}, nil
}

func (r *nextPhaseRepoStub) CreateContributorItemAndFollow(_ context.Context, input CoreContributorItemInput) (CoreItem, error) {
	r.itemInput = input
	return CoreItem{
		ID:              uuid.New(),
		WorkspaceID:     input.Item.WorkspaceID,
		PortalID:        input.Item.PortalID,
		BoardID:         input.Item.BoardID,
		ContributorID:   input.Participant.ID,
		ParticipantKind: input.Participant.Kind,
		Following:       true,
		Title:           input.Item.Title,
		Description:     input.Item.Description,
		Slug:            input.Item.Slug,
	}, nil
}

func (r *nextPhaseRepoStub) CreateContributorComment(_ context.Context, input CoreContributorCommentInput) (CoreComment, error) {
	r.commentInput = input
	return CoreComment{
		ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, ContributorID: input.Participant.ID,
		ParticipantKind: input.Participant.Kind, ParentID: input.ParentID, Body: input.Body,
	}, nil
}

func (r *nextPhaseRepoStub) GetPublicWidgetSettings(_ context.Context, portalID uuid.UUID) (CoreWidgetSettings, error) {
	if r.widgetSettings.PortalID != portalID {
		return CoreWidgetSettings{}, ErrNotFound
	}
	return r.widgetSettings, nil
}

func (r *nextPhaseRepoStub) UpsertWidgetSettings(_ context.Context, input CoreWidgetSettingsInput) (CoreWidgetSettings, error) {
	r.widgetSettingsInput = input
	r.widgetSettings = CoreWidgetSettings{
		PortalID:       input.PortalID,
		Enabled:        input.Enabled,
		AllowedOrigins: append([]string(nil), input.AllowedOrigins...),
	}
	return r.widgetSettings, nil
}

func (r *nextPhaseRepoStub) GetWidgetSigningSecret(_ context.Context, portalID, keyID uuid.UUID, version int) (string, error) {
	if portalID != r.widgetSettings.PortalID || keyID != r.widgetSettings.WidgetKeyID || version != r.widgetSettings.SigningSecretVersion {
		return "", ErrNotFound
	}
	return r.widgetSecret, nil
}

func (r *nextPhaseRepoStub) ConsumeWidgetAssertionNonce(_ context.Context, portalID, keyID uuid.UUID, version int, nonce, origin string, expiresAt time.Time) error {
	r.nonceCalls++
	return r.nonceErr
}

func (r *nextPhaseRepoStub) CreateExternalContributorSession(_ context.Context, portalID uuid.UUID, externalID, email, displayName string, avatarURL *string, tokenHash []byte, expiresAt time.Time) (CoreParticipant, CoreParticipantSession, error) {
	r.externalEmail = email
	participant := CoreParticipant{
		ID:          uuid.New(),
		PortalID:    portalID,
		Kind:        ContributorKindExternal,
		Email:       email,
		ExternalID:  externalID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}
	return participant, CoreParticipantSession{ID: uuid.New(), PortalID: portalID, ContributorID: participant.ID, Source: ContributorSessionSourceWidget, ExpiresAt: expiresAt}, nil
}

func TestRequestContributorVerificationMasksAndPublishesOneTimeCredentials(t *testing.T) {
	portalID := uuid.New()
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	repo := &nextPhaseRepoStub{repoStub: &repoStub{portals: []CorePortal{{
		ID: portalID, Name: "Roads", Slug: "roads", IsPublic: true,
		ParticipationMode: ParticipationModeVerifiedGuest, GuestIdentityPolicy: GuestIdentityPolicyAllowPublicMasking,
	}}}}
	publisher := &eventPublisherStub{}
	service := New(repo, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("test-auth-secret", "https://app.example.com/", nil),
	)
	service.now = func() time.Time { return now }
	service.random = bytes.NewReader(make([]byte, 128))

	challenge, err := service.RequestContributorVerification(context.Background(), "roads", "Guest@Example.com", "  Guest  ", true, ContributorSessionSourcePortal)

	require.NoError(t, err)
	require.Equal(t, now.Add(verificationLifetime), challenge.ExpiresAt)
	require.Equal(t, portalID, repo.verificationInput.PortalID)
	require.Equal(t, "guest@example.com", repo.verificationInput.Email)
	require.Equal(t, "Guest", repo.verificationInput.DisplayName)
	require.True(t, repo.verificationInput.PublicMasked)
	require.Len(t, repo.verificationInput.TokenHash, 32)
	require.Len(t, repo.verificationInput.CodeHash, 32)
	require.Len(t, publisher.events, 1)
	require.Equal(t, events.FeedbackContributorVerification, publisher.events[0].Type)
	payload := publisher.events[0].Payload.(events.FeedbackContributorVerificationPayload)
	require.Equal(t, "000000", payload.Code)
	require.Contains(t, payload.VerificationURL, "/portal/roads/feedback/verify?token=")
	require.NotContains(t, payload.VerificationURL, "Guest@Example.com")
}

func TestCreatePublicItemUsesVerifiedContributorAndFollowsAtomically(t *testing.T) {
	workspaceID, portalID, boardID := uuid.New(), uuid.New(), uuid.New()
	participant := CoreParticipant{ID: uuid.New(), PortalID: portalID, Kind: ContributorKindVerifiedGuest}
	repo := &nextPhaseRepoStub{repoStub: &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "roads", IsPublic: true, ParticipationMode: ParticipationModeVerifiedGuest}},
		boards:  []CoreBoard{{ID: boardID, WorkspaceID: workspaceID, PortalID: portalID}},
	}}
	service := New(repo, nil, WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))

	result, err := service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug: "roads", BoardID: boardID, Title: "Add a dark mode", Description: "Please", Source: SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentVerifiedGuest, Participant: &participant,
	})

	require.NoError(t, err)
	require.Equal(t, ContributorKindVerifiedGuest, result.ParticipantKind)
	require.True(t, result.Following)
	require.Equal(t, participant.ID, repo.itemInput.Participant.ID)
	require.Equal(t, participant.ID, repo.itemInput.Item.ContributorID)
	require.Equal(t, uuid.Nil, repo.itemInput.Item.AuthorID)

	_, err = service.CreatePublicItem(context.Background(), CorePublicItemInput{
		PortalSlug: "roads", BoardID: boardID, Title: "Anonymous attempt", Source: SubmissionSourcePortal,
		ParticipationIntent: ParticipationIntentAnonymous,
	})
	require.ErrorIs(t, err, ErrParticipationNotAllowed)
}

func TestPreferenceSessionCannotResolveAsParticipationSession(t *testing.T) {
	portalID := uuid.New()
	participant := CoreParticipant{ID: uuid.New(), PortalID: portalID, Kind: ContributorKindVerifiedGuest}
	repo := &nextPhaseRepoStub{
		repoStub:           &repoStub{portals: []CorePortal{{ID: portalID, Slug: "roads", IsPublic: true, ParticipationMode: ParticipationModeVerifiedGuest}}},
		sessionParticipant: participant,
		session: CoreParticipantSession{
			ID: uuid.New(), PortalID: portalID, ContributorID: participant.ID,
			Source: ContributorSessionSourcePreferences, ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	service := New(repo, nil, WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))

	_, err := service.ResolveContributorSession(context.Background(), "roads", "FeedbackSession preference-token", "")
	require.ErrorIs(t, err, ErrContributorSessionInvalid)
	_, err = service.ResolveContributorRateLimitIdentity(context.Background(), "roads", "FeedbackSession preference-token")
	require.ErrorIs(t, err, ErrContributorSessionInvalid)
}

func TestGuestReplyPublishesAccountNotificationsWithSafeAttribution(t *testing.T) {
	workspaceID, portalID, itemID := uuid.New(), uuid.New(), uuid.New()
	itemAuthorID, parentAuthorID, parentID := uuid.New(), uuid.New(), uuid.New()
	followerID := uuid.New()
	systemActorID := uuid.New()
	participant := CoreParticipant{
		ID: uuid.New(), PortalID: portalID, Kind: ContributorKindVerifiedGuest,
		DisplayName: "Private Guest", PublicMasked: true,
	}
	repo := &nextPhaseRepoStub{repoStub: &repoStub{
		portals:  []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "roads", IsPublic: true, ParticipationMode: ParticipationModeVerifiedGuest}},
		items:    []CoreItem{{ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, AuthorID: itemAuthorID, Title: "Dark mode", Slug: "dark-mode"}},
		comments: []CoreComment{{ID: parentID, WorkspaceID: workspaceID, ItemID: itemID, AuthorID: parentAuthorID}},
	}, accountItemFollowers: []uuid.UUID{itemAuthorID, parentAuthorID, followerID}}
	publisher := &eventPublisherStub{}
	service := New(repo, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("test-auth-secret", "https://app.example.com", nil),
		WithGuestNotificationActor(systemActorID),
	)

	comment, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "roads", ItemID: itemID, Participant: &participant, ParentID: &parentID, Body: "It is ready",
	})

	require.NoError(t, err)
	require.Equal(t, participant.ID, comment.ContributorID)
	require.Len(t, publisher.events, 3)
	recipients := make(map[uuid.UUID]events.FeedbackCommentCreatedPayload, 3)
	for _, event := range publisher.events {
		require.Equal(t, events.FeedbackCommentCreated, event.Type)
		require.Equal(t, systemActorID, event.ActorID)
		payload := event.Payload.(events.FeedbackCommentCreatedPayload)
		require.Equal(t, participant.ID, payload.ActorContributorID)
		require.Equal(t, "Anonymous", payload.ActorName)
		recipients[payload.RecipientID] = payload
	}
	require.False(t, recipients[itemAuthorID].IsReply)
	require.True(t, recipients[parentAuthorID].IsReply)
	require.False(t, recipients[followerID].IsReply)
}

func TestAccountCommentNotifiesFollowersOnceAndSuppressesActor(t *testing.T) {
	workspaceID, portalID, itemID := uuid.New(), uuid.New(), uuid.New()
	authorID, followerID, actorID := uuid.New(), uuid.New(), uuid.New()
	repo := &nextPhaseRepoStub{
		repoStub: &repoStub{
			portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "roads", IsPublic: true}},
			items: []CoreItem{{
				ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, AuthorID: authorID,
				Title: "Dark mode", Slug: "dark-mode",
			}},
		},
		accountItemFollowers: []uuid.UUID{authorID, followerID, actorID, followerID},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher), WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))

	_, err := service.CreatePublicComment(context.Background(), CorePublicCommentInput{
		PortalSlug: "roads", ItemID: itemID, AuthorID: actorID, Body: "I can reproduce this",
	})

	require.NoError(t, err)
	require.Len(t, publisher.events, 2)
	recipients := map[uuid.UUID]struct{}{}
	for _, event := range publisher.events {
		payload := event.Payload.(events.FeedbackCommentCreatedPayload)
		require.NotEqual(t, actorID, payload.RecipientID)
		recipients[payload.RecipientID] = struct{}{}
	}
	require.Contains(t, recipients, authorID)
	require.Contains(t, recipients, followerID)
}

func TestDirectStatusNotifiesFollowersOnceAndSuppressesActor(t *testing.T) {
	workspaceID, portalID, itemID := uuid.New(), uuid.New(), uuid.New()
	authorID, followerID, actorID := uuid.New(), uuid.New(), uuid.New()
	base := &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "roads"}},
		items: []CoreItem{{
			ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, AuthorID: authorID,
			Title: "Dark mode", Slug: "dark-mode", Status: StatusPending,
		}},
	}
	repo := &nextPhaseRepoStub{repoStub: base, accountItemFollowers: []uuid.UUID{authorID, followerID, actorID, followerID}}
	publisher := &eventPublisherStub{}
	service := New(repo, nil, WithEventPublisher(nil, publisher), WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))

	_, err := service.UpdateItemStatus(context.Background(), workspaceID, itemID, CoreUpdateItemStatusInput{
		Status: StatusPlanned, ActorID: actorID,
	})

	require.NoError(t, err)
	require.Len(t, publisher.events, 2)
	recipients := map[uuid.UUID]struct{}{}
	for _, event := range publisher.events {
		payload := event.Payload.(events.FeedbackStatusUpdatedPayload)
		require.NotEqual(t, actorID, payload.RecipientID)
		recipients[payload.RecipientID] = struct{}{}
	}
	require.Contains(t, recipients, authorID)
	require.Contains(t, recipients, followerID)
}

func TestPublishUpdateDefersNotificationsToDurableOutbox(t *testing.T) {
	workspaceID, portalID, updateID, itemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	actorID, recipientID := uuid.New(), uuid.New()
	repo := &nextPhaseRepoStub{
		repoStub: &repoStub{},
		publishedUpdate: CoreFeedbackUpdate{
			ID: updateID, WorkspaceID: workspaceID, PortalID: portalID, Slug: "dark-mode", Title: "Dark mode shipped",
			Status: FeedbackUpdateStatusPublished, LinkedItems: []CoreUpdateItem{{ID: itemID}},
		},
		newlyPublished: true,
		accountRecipients: []CoreAccountUpdateRecipient{
			{UserID: recipientID, ItemID: itemID},
			{UserID: actorID, ItemID: itemID},
		},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("test-auth-secret", "https://app.example.com", nil),
	)

	_, err := service.PublishUpdate(context.Background(), workspaceID, updateID, actorID)

	require.NoError(t, err)
	require.Empty(t, publisher.events, "publication side effects belong to the durable outbox worker")
}

func TestListMergeCandidatesUsesCanonicalSourcePortalAndCapsLimit(t *testing.T) {
	workspaceID, portalID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &nextPhaseRepoStub{
		repoStub: &repoStub{items: []CoreItem{{
			ID: sourceID, WorkspaceID: workspaceID, PortalID: portalID,
		}}},
		itemCandidates: CoreMergeCandidatesPage{Candidates: []CoreMergeCandidate{{
			ID: targetID, Slug: "canonical", Title: "Canonical", Status: StatusCompleted,
		}}},
	}
	service := New(repository, nil)

	page, err := service.ListMergeCandidates(context.Background(), workspaceID, sourceID, "  dark mode  ", 500)

	require.NoError(t, err)
	require.Len(t, page.Candidates, 1)
	require.Equal(t, workspaceID, repository.candidateWorkspaceID)
	require.Equal(t, portalID, repository.candidatePortalID)
	require.Equal(t, sourceID, repository.candidateExcludedID)
	require.Equal(t, "dark mode", repository.candidateSearch)
	require.Equal(t, maxMergeCandidatesLimit, repository.candidateLimit)
}

func TestListMergeCandidatesRejectsMergedSource(t *testing.T) {
	workspaceID, portalID, sourceID, targetID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repository := &nextPhaseRepoStub{repoStub: &repoStub{items: []CoreItem{{
		ID: sourceID, WorkspaceID: workspaceID, PortalID: portalID, MergedIntoItemID: &targetID,
	}}}}
	service := New(repository, nil)

	_, err := service.ListMergeCandidates(context.Background(), workspaceID, sourceID, "", 30)

	require.ErrorIs(t, err, ErrMergeConflict)
	require.Equal(t, uuid.Nil, repository.candidatePortalID)
}

func TestListPortalItemCandidatesIncludesAllCanonicalStatuses(t *testing.T) {
	workspaceID, portalID := uuid.New(), uuid.New()
	repository := &nextPhaseRepoStub{
		repoStub: &repoStub{portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID}}},
		itemCandidates: CoreMergeCandidatesPage{Candidates: []CoreMergeCandidate{
			{ID: uuid.New(), Status: StatusCompleted},
			{ID: uuid.New(), Status: StatusClosed},
		}},
	}
	service := New(repository, nil)

	page, err := service.ListPortalItemCandidates(context.Background(), workspaceID, portalID, "  shipped  ", 0)

	require.NoError(t, err)
	require.Len(t, page.Candidates, 2)
	require.Equal(t, portalID, repository.candidatePortalID)
	require.Equal(t, uuid.Nil, repository.candidateExcludedID)
	require.Equal(t, "shipped", repository.candidateSearch)
	require.Equal(t, 30, repository.candidateLimit)
}

func TestMergedSourceRejectsContributorFollowReadsAndWrites(t *testing.T) {
	workspaceID, portalID, itemID, targetID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	participant := CoreParticipant{ID: uuid.New(), PortalID: portalID, Kind: ContributorKindVerifiedGuest}
	repository := &nextPhaseRepoStub{repoStub: &repoStub{
		portals: []CorePortal{{ID: portalID, WorkspaceID: workspaceID, Slug: "roads"}},
		items: []CoreItem{{
			ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, MergedIntoItemID: &targetID,
		}},
	}}
	service := New(repository, nil)

	_, getErr := service.GetItemFollow(context.Background(), "roads", itemID, participant)
	_, setErr := service.FollowItem(context.Background(), "roads", itemID, participant, true)
	_, unsetErr := service.FollowItem(context.Background(), "roads", itemID, participant, false)

	require.ErrorIs(t, getErr, ErrMergeConflict)
	require.ErrorIs(t, setErr, ErrMergeConflict)
	require.ErrorIs(t, unsetErr, ErrMergeConflict)
}

func TestLinkedStoryStatusBridgePublishesStableAccountEvent(t *testing.T) {
	workspaceID, storyID, itemID := uuid.New(), uuid.New(), uuid.New()
	actorID, recipientID := uuid.New(), uuid.New()
	occurredAt := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	followerID := uuid.New()
	repo := &nextPhaseRepoStub{
		repoStub:             &repoStub{},
		accountItemFollowers: []uuid.UUID{recipientID, followerID, actorID, followerID},
		primaryStoryItems: []CoreItem{{
			ID: itemID, WorkspaceID: workspaceID, PortalID: uuid.New(), AuthorID: recipientID,
			Title: "Dark mode", Slug: "dark-mode", Status: StatusInProgress,
		}},
	}
	publisher := &eventPublisherStub{}
	service := New(repo, nil,
		WithEventPublisher(nil, publisher),
		WithContributorFeatures("test-auth-secret", "https://app.example.com", nil),
	)

	require.NoError(t, service.NotifyLinkedStoryStatusTransition(context.Background(), workspaceID, storyID, actorID, occurredAt))
	require.Len(t, publisher.events, 2)
	payload := publisher.events[0].Payload.(events.FeedbackStatusUpdatedPayload)
	require.Equal(t, StatusInProgress, payload.Status)
	require.Equal(t, occurredAt, publisher.events[0].Timestamp)
	expectedEventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(storyID.String()+":"+occurredAt.Format(time.RFC3339Nano)+":"+itemID.String()+":"+StatusInProgress))
	require.Equal(t, expectedEventID, payload.EventID)
	recipients := map[uuid.UUID]struct{}{}
	for _, event := range publisher.events {
		statusPayload := event.Payload.(events.FeedbackStatusUpdatedPayload)
		require.Equal(t, expectedEventID, statusPayload.EventID)
		require.NotEqual(t, actorID, statusPayload.RecipientID)
		recipients[statusPayload.RecipientID] = struct{}{}
	}
	require.Contains(t, recipients, recipientID)
	require.Contains(t, recipients, followerID)

	publisher.events = nil
	require.NoError(t, service.NotifyLinkedStoryStatusTransition(context.Background(), workspaceID, storyID, actorID, occurredAt))
	require.Len(t, publisher.events, 2)
	require.Equal(t, expectedEventID, publisher.events[0].Payload.(events.FeedbackStatusUpdatedPayload).EventID)
}

func TestLinkedStoryStatusNotificationPolicy(t *testing.T) {
	t.Parallel()

	text := func(value string) *string { return &value }
	tests := []struct {
		name             string
		status           string
		roadmapSummary   *string
		wantNotification bool
	}{
		{name: "pending is internal triage", status: StatusPending},
		{name: "reviewing is internal triage", status: StatusReviewing},
		{name: "planned is public", status: StatusPlanned, wantNotification: true},
		{name: "in progress is public", status: StatusInProgress, wantNotification: true},
		{name: "completed is public", status: StatusCompleted, wantNotification: true},
		{name: "closed without public explanation is silent", status: StatusClosed},
		{name: "closed with blank public explanation is silent", status: StatusClosed, roadmapSummary: text("  ")},
		{name: "closed with existing public explanation is public", status: StatusClosed, roadmapSummary: text("This is no longer aligned with the product direction."), wantNotification: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			workspaceID, portalID, storyID, itemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			actorID, recipientID := uuid.New(), uuid.New()
			repo := &nextPhaseRepoStub{
				repoStub: &repoStub{portals: []CorePortal{{
					ID: portalID, WorkspaceID: workspaceID, Slug: "roads",
				}}},
				accountItemFollowers: []uuid.UUID{recipientID},
				primaryStoryItems: []CoreItem{{
					ID: itemID, WorkspaceID: workspaceID, PortalID: portalID, AuthorID: recipientID,
					Title: "Dark mode", Slug: "dark-mode", Status: test.status, RoadmapSummary: test.roadmapSummary,
				}},
			}
			publisher := &eventPublisherStub{}
			service := New(repo, nil, WithEventPublisher(nil, publisher))

			err := service.NotifyLinkedStoryStatusTransition(
				context.Background(), workspaceID, storyID, actorID,
				time.Date(2026, time.August, 13, 15, 0, 0, 0, time.UTC),
			)

			require.NoError(t, err)
			if test.wantNotification {
				require.Len(t, publisher.events, 1)
				require.Equal(t, test.status, publisher.events[0].Payload.(events.FeedbackStatusUpdatedPayload).Status)
			} else {
				require.Empty(t, publisher.events)
			}
		})
	}
}

func TestWidgetAssertionIsPortalOriginExpiryAndReplayBound(t *testing.T) {
	now := time.Date(2026, time.August, 13, 8, 0, 0, 0, time.UTC)
	portalID, keyID := uuid.New(), uuid.New()
	repo := &nextPhaseRepoStub{repoStub: &repoStub{portals: []CorePortal{{
		ID: portalID, Slug: "roads", IsPublic: true, ParticipationMode: ParticipationModeVerifiedGuest,
	}}}}
	service := New(repo, nil, WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))
	service.now = func() time.Time { return now }
	service.random = bytes.NewReader(make([]byte, 256))
	encrypted, err := service.sealWidgetSecret(portalID, 1, "customer-hmac-secret")
	require.NoError(t, err)
	repo.widgetSettings = CoreWidgetSettings{
		PortalID: portalID, Enabled: true, WidgetKeyID: keyID, AllowedOrigins: []string{"https://customer.example"},
		SigningSecretVersion: 1, SigningSecretEncrypted: encrypted,
	}
	repo.widgetSecret = encrypted
	assertion, err := encodeWidgetAssertionForTest(CoreWidgetIdentityAssertion{
		Version: 1, KeyID: keyID.String(), ExternalID: "customer-user-42", Email: "person@example.com",
		DisplayName: "Person", IssuedAt: now.Unix(), ExpiresAt: now.Add(2 * time.Minute).Unix(),
		Nonce: "once-only", Origin: "https://customer.example",
	}, "customer-hmac-secret")
	require.NoError(t, err)

	result, err := service.CreateWidgetContributorSession(context.Background(), CoreWidgetSessionInput{
		PortalSlug: "roads", Assertion: assertion, ParentOrigin: "https://customer.example",
	})
	require.NoError(t, err)
	require.Equal(t, ContributorKindExternal, result.Participant.Kind)
	require.NotEmpty(t, result.Token)
	require.Equal(t, "person@example.com", repo.externalEmail)
	require.Equal(t, 1, repo.nonceCalls)

	_, err = service.CreateWidgetContributorSession(context.Background(), CoreWidgetSessionInput{
		PortalSlug: "roads", Assertion: assertion, ParentOrigin: "https://evil.example",
	})
	require.ErrorIs(t, err, ErrWidgetOriginNotAllowed)
	require.Equal(t, 1, repo.nonceCalls, "origin failures must not consume a nonce")

	repo.nonceErr = ErrWidgetAssertionReplayed
	_, err = service.CreateWidgetContributorSession(context.Background(), CoreWidgetSessionInput{
		PortalSlug: "roads", Assertion: assertion, ParentOrigin: "https://customer.example",
	})
	require.ErrorIs(t, err, ErrWidgetAssertionReplayed)
}

func TestBasicWidgetCanBeEnabledWithoutCustomIdentity(t *testing.T) {
	workspaceID, portalID := uuid.New(), uuid.New()
	repo := &nextPhaseRepoStub{repoStub: &repoStub{}}
	service := New(repo, nil)

	settings, err := service.UpdateWidgetSettings(context.Background(), CoreWidgetSettingsInput{
		WorkspaceID: workspaceID,
		PortalID:    portalID,
		Enabled:     true,
		AllowedOrigins: []string{
			"http://localhost:5500",
			"https://app.example.com/",
		},
	})

	require.NoError(t, err)
	require.True(t, settings.Enabled)
	require.Equal(t, []string{"http://localhost:5500", "https://app.example.com"}, settings.AllowedOrigins)
	require.True(t, repo.widgetSettingsInput.Enabled)
	require.Empty(t, repo.widgetSettings.SigningSecretEncrypted)
}

func TestWidgetSecretEnvelopeCannotMoveAcrossPortalOrVersion(t *testing.T) {
	service := New(&repoStub{}, nil, WithContributorFeatures("test-auth-secret", "https://app.example.com", nil))
	service.random = bytes.NewReader(make([]byte, 128))
	portalID := uuid.New()

	envelope, err := service.sealWidgetSecret(portalID, 2, "secret")
	require.NoError(t, err)
	plaintext, err := service.openWidgetSecret(portalID, 2, envelope)
	require.NoError(t, err)
	require.Equal(t, "secret", plaintext)
	_, err = service.openWidgetSecret(uuid.New(), 2, envelope)
	require.Error(t, err)
	_, err = service.openWidgetSecret(portalID, 3, envelope)
	require.Error(t, err)
}

func TestNormalizeAllowedOriginsRequiresExactSecureOrigins(t *testing.T) {
	origins, err := normalizeAllowedOrigins([]string{"https://B.example", "https://a.example", "https://B.example/", "http://localhost:3000"})
	require.NoError(t, err)
	require.Equal(t, []string{"http://localhost:3000", "https://a.example", "https://b.example"}, origins)

	for _, invalid := range []string{"http://example.com", "https://*.example.com", "https://example.com/path", "https://user@example.com", "https://example.com?q=1"} {
		_, err := normalizeAllowedOrigins([]string{invalid})
		require.ErrorIs(t, err, ErrInvalidInput, invalid)
	}
}

func TestParseWidgetAssertionRejectsTrailingJSON(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"version":1} {"version":2}`))
	signature := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	_, _, _, err := parseWidgetAssertion(payload + "." + signature)
	require.True(t, errors.Is(err, ErrWidgetAssertionInvalid))
}

func TestMaskParticipantAppliesAlwaysMaskOnlyToGuestKinds(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{ContributorKindVerifiedGuest, ContributorKindExternal} {
		participant := maskParticipant(CoreParticipant{Kind: kind}, GuestIdentityPolicyAlwaysMaskGuests)
		require.True(t, participant.PublicMasked, kind)
	}

	account := maskParticipant(CoreParticipant{Kind: ContributorKindAccount}, GuestIdentityPolicyAlwaysMaskGuests)
	require.False(t, account.PublicMasked)
}
