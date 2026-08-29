package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventProcessorDeduplicatesCompletedInboundEvent(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	store.processInbound = false
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-duplicate", "show my work")))
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if repo.getInstallationCalls != 0 || len(sender.messages) != 0 || len(assistant.requests) != 0 {
		t.Fatalf("duplicate event performed side effects: repo calls=%d sends=%d assistant calls=%d", repo.getInstallationCalls, len(sender.messages), len(assistant.requests))
	}
	if len(store.completions) != 0 {
		t.Fatalf("duplicate event completion calls = %d, want 0", len(store.completions))
	}
}

func TestEventProcessorRecoversInboxReceiptWithoutQueueingMessageContent(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	recovery := &webhookRecoveryStub{report: webhooks.RecoveryReport{Claimed: 1, Dispatched: 1}}
	processor.webhookRecovery = recovery

	recovered, err := processor.RecoverPendingEvents(context.Background())
	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 1 {
		t.Fatalf("recovered = %d, want 1", recovered)
	}
	if recovery.provider != slackWebhookProvider || recovery.policy.ClaimLimit != 500 || recovery.policy.ProcessingLease != slackWebhookProcessingLease {
		t.Fatalf("recovery request = provider %q policy %+v", recovery.provider, recovery.policy)
	}
}

func TestEventProcessorRecoveryContinuesPastQueueFailure(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	queueFailure := errors.New("queue unavailable")
	processor.webhookRecovery = &webhookRecoveryStub{
		report: webhooks.RecoveryReport{Claimed: 2, Dispatched: 1, Released: 1},
		err:    queueFailure,
	}

	recovered, err := processor.RecoverPendingEvents(context.Background())

	if !errors.Is(err, queueFailure) {
		t.Fatalf("RecoverPendingEvents() error = %v, want queue failure", err)
	}
	if recovered != 1 {
		t.Fatalf("recovered = %d, want one successful dispatch", recovered)
	}
}

func TestEventProcessorRetriesTransientUninstallWithoutTargetingReconnect(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.installationErr = slackdomain.ErrNotFound
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.clientID = "client-id"
	processor.clientSecret = "client-secret"
	credentialPayload, credentialVersion, err := processor.codec.seal(slackCredentialBinding{
		WorkspaceID:       testWorkspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: testInstallGeneration,
	}, slackCredential{AccessToken: "xoxb-disconnected"})
	if err != nil {
		t.Fatalf("seal uninstall credential: %v", err)
	}
	uninstallID := uuid.New()
	repo.recoverableUninstalls = []slackrepository.SlackUninstallRecord{
		{
			ID:                   uninstallID,
			SlackWorkspaceID:     testSlackWorkspaceID,
			WorkspaceID:          testWorkspaceID,
			InstallGeneration:    testInstallGeneration,
			SlackTeamID:          "T1",
			UninstallKind:        "disconnect",
			CredentialPayload:    credentialPayload,
			CredentialKeyVersion: credentialVersion,
			Status:               "processing",
			AttemptCount:         1,
		},
	}
	providerCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		providerCalls++
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	processor.webClient = newSlackWebClient(server.Client())
	processor.webClient.baseURL = server.URL

	recovered, err := processor.RecoverPendingEvents(context.Background())
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("RecoverPendingEvents() error = %v, want rate limit", err)
	}
	if recovered != 0 || len(repo.failedUninstalls) != 1 || len(repo.completedUninstalls) != 0 || providerCalls != 1 {
		t.Fatalf("first recovery recovered/failed/completed/provider = %d/%d/%d/%d", recovered, len(repo.failedUninstalls), len(repo.completedUninstalls), providerCalls)
	}

	// A new same-team generation wins before the retry. Recovery must complete
	// the old outbox without calling apps.uninstall through the replacement.
	repo.installationErr = nil
	repo.installation.InstallGeneration = uuid.New()
	recovered, err = processor.RecoverPendingEvents(context.Background())
	if err != nil {
		t.Fatalf("RecoverPendingEvents() after reconnect error = %v", err)
	}
	if recovered != 1 || len(repo.completedUninstalls) != 1 || providerCalls != 1 {
		t.Fatalf("reconnect recovery recovered/completed/provider = %d/%d/%d, want 1/1/1", recovered, len(repo.completedUninstalls), providerCalls)
	}
}

func TestEventProcessorRecoversPersistedOutboundDelivery(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	content := "Task created in FortyOne."
	threadID := "171.100"
	store.deliveryContent = &content
	store.recoverableOutbound = []outboundDeliveryRecord{
		{
			ID:                  testOutboundDeliveryID,
			WorkspaceID:         testWorkspaceID,
			InstallGeneration:   uuidPointer(testInstallGeneration),
			ExternalWorkspaceID: "T1",
			IdempotencyKey:      "slack:task:confirmation",
			ExternalChannelID:   "C1",
			ExternalThreadID:    &threadID,
			Content:             &content,
			Status:              "failed",
			AttemptCount:        2,
		},
	}
	sender := &messageSenderStub{externalMessageID: "171.200"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

	recovered, err := processor.RecoverPendingEvents(context.Background())

	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 1 || len(sender.messages) != 1 || len(store.completedDeliveries) != 1 {
		t.Fatalf("recovered/sent/completed = %d/%d/%d, want 1/1/1", recovered, len(sender.messages), len(store.completedDeliveries))
	}
	if len(repo.requestedTeamIDs) != 2 || repo.requestedTeamIDs[0] != "T1" || repo.requestedTeamIDs[1] != "T1" {
		t.Fatalf("installation lookup team IDs = %v, want [T1 T1]", repo.requestedTeamIDs)
	}
	message := sender.messages[0]
	if message.ChannelID != "C1" || message.ThreadTS != threadID || message.Text != content || message.ClientMessageID == "" {
		t.Fatalf("recovered message = %+v", message)
	}
}

func TestEventProcessorRecoversFirstUseGuideWithoutLinkedUser(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	content := "*Welcome to FortyOne in Slack*"
	recipient := "U-first-use"
	store.deliveryContent = &content
	store.recoverableOutbound = []outboundDeliveryRecord{
		{
			ID:                      testOutboundDeliveryID,
			WorkspaceID:             testWorkspaceID,
			InstallGeneration:       uuidPointer(testInstallGeneration),
			ExternalWorkspaceID:     "T1",
			ExternalRecipientUserID: &recipient,
			IdempotencyKey:          "slack-onboarding:generation:U-first-use",
			ExternalChannelID:       recipient,
			Content:                 &content,
			Purpose:                 slackOnboardingPurpose,
			Status:                  "failed",
			AttemptCount:            2,
		},
	}
	sender := &messageSenderStub{externalMessageID: "171.200"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

	recovered, err := processor.RecoverPendingEvents(context.Background())

	require.NoError(t, err)
	require.Equal(t, 1, recovered)
	require.Len(t, sender.messages, 1)
	require.Equal(t, recipient, sender.messages[0].ChannelID)
	require.Equal(t, content, sender.messages[0].Text)
	require.Equal(
		t,
		deterministicSlackMessageID(slackFirstInteractionGuideProviderKey(testWorkspaceID, "T1", recipient)),
		sender.messages[0].ClientMessageID,
	)
	require.Len(t, store.completedDeliveries, 1)
	require.Len(t, store.outboundInputs, 1)
	require.Nil(t, store.outboundInputs[0].UserID)
}

func TestEventProcessorCancelsAssistantRecoveryAfterActorLosesAccess(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = nil
	store := newEventStoreStub()
	content := "Private workspace answer"
	recipient := "U1"
	expiresAt := time.Now().UTC().Add(time.Hour)
	store.deliveryContent = &content
	store.recoverableOutbound = []outboundDeliveryRecord{
		{
			ID:                      testOutboundDeliveryID,
			WorkspaceID:             testWorkspaceID,
			UserID:                  uuidPointer(testLinkedUserID),
			InstallGeneration:       uuidPointer(testInstallGeneration),
			ExternalWorkspaceID:     "T1",
			ExternalRecipientUserID: &recipient,
			IdempotencyKey:          "slack:private-assistant",
			ExternalChannelID:       "D1",
			Content:                 &content,
			Purpose:                 "assistant",
			ExpiresAt:               &expiresAt,
			Status:                  "failed",
			AttemptCount:            2,
		},
	}
	sender := &messageSenderStub{externalMessageID: "171.200"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

	recovered, err := processor.RecoverPendingEvents(context.Background())
	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 0 || len(sender.messages) != 0 || len(store.cancelledDeliveries) != 1 {
		t.Fatalf("recovered/sent/cancelled = %d/%d/%d, want 0/0/1", recovered, len(sender.messages), len(store.cancelledDeliveries))
	}
}

func TestEventProcessorCancelsAssistantRecoveryAfterSharedChannelAudienceNarrows(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.sharedTeamIDs = nil
	store := newEventStoreStub()
	content := "Private team answer"
	recipient := "U1"
	expiresAt := time.Now().UTC().Add(time.Hour)
	providerPayload, err := EncodeSlackProviderPayload(SlackProviderPayload{
		Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{testAllowedTeamID},
			SharedTeamIDs:  []uuid.UUID{testAllowedTeamID},
			ActorUserID:    uuidPointer(testLinkedUserID),
		},
	})
	if err != nil {
		t.Fatalf("EncodeSlackProviderPayload() error = %v", err)
	}
	store.deliveryContent = &content
	store.recoverableOutbound = []outboundDeliveryRecord{{
		ID:                      testOutboundDeliveryID,
		WorkspaceID:             testWorkspaceID,
		UserID:                  uuidPointer(testLinkedUserID),
		InstallGeneration:       uuidPointer(testInstallGeneration),
		ExternalWorkspaceID:     "T1",
		ExternalRecipientUserID: &recipient,
		IdempotencyKey:          "slack:channel-assistant",
		ExternalChannelID:       "C1",
		Content:                 &content,
		ProviderPayload:         providerPayload,
		Purpose:                 "assistant",
		ExpiresAt:               &expiresAt,
		Status:                  "failed",
		AttemptCount:            2,
	}}
	sender := &messageSenderStub{externalMessageID: "171.200"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

	recovered, err := processor.RecoverPendingEvents(context.Background())
	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 0 || len(sender.messages) != 0 || len(store.cancelledDeliveries) != 1 {
		t.Fatalf("recovered/sent/cancelled = %d/%d/%d, want 0/0/1", recovered, len(sender.messages), len(store.cancelledDeliveries))
	}
}

func TestEventProcessorCancelsRecoveredRequestCommentAfterDMRecipientLosesTeamAccess(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = nil
	store := newEventStoreStub()
	content := "Follow-up from FortyOne"
	recipient := "U1"
	providerPayload, err := EncodeSlackProviderPayload(SlackProviderPayload{
		Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{testAllowedTeamID},
			ActorUserID:    uuidPointer(testLinkedUserID),
		},
	})
	if err != nil {
		t.Fatalf("EncodeSlackProviderPayload() error = %v", err)
	}
	store.deliveryContent = &content
	store.recoverableOutbound = []outboundDeliveryRecord{{
		ID:                      testOutboundDeliveryID,
		WorkspaceID:             testWorkspaceID,
		UserID:                  uuidPointer(testLinkedUserID),
		InstallGeneration:       uuidPointer(testInstallGeneration),
		ExternalWorkspaceID:     "T1",
		ExternalRecipientUserID: &recipient,
		IdempotencyKey:          "integration-request-comment:recover-dm",
		ExternalChannelID:       "D1",
		ExternalThreadID:        stringPointer("171.100"),
		Content:                 &content,
		ProviderPayload:         providerPayload,
		Purpose:                 "provider_message",
		Status:                  "failed",
		AttemptCount:            2,
	}}
	sender := &messageSenderStub{externalMessageID: "171.200"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

	recovered, err := processor.RecoverPendingEvents(context.Background())

	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 0 || len(sender.messages) != 0 || len(store.cancelledDeliveries) != 1 {
		t.Fatalf("recovered/sent/cancelled = %d/%d/%d, want 0/0/1", recovered, len(sender.messages), len(store.cancelledDeliveries))
	}
}

func TestEventProcessorRecoveryRejectsMismatchedSlackInstallation(t *testing.T) {
	tests := []struct {
		name                  string
		installationTeamID    string
		installationWorkspace uuid.UUID
	}{
		{
			name:                  "workspace reinstalled to another Slack team",
			installationTeamID:    "T-reinstalled",
			installationWorkspace: testWorkspaceID,
		},
		{
			name:                  "Slack team belongs to another FortyOne workspace",
			installationTeamID:    "T-original",
			installationWorkspace: uuid.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.installation.SlackTeamID = tt.installationTeamID
			repo.installation.WorkspaceID = tt.installationWorkspace
			store := newEventStoreStub()
			content := "Do not send through a replacement installation."
			store.deliveryContent = &content
			store.recoverableOutbound = []outboundDeliveryRecord{
				{
					ID:                  testOutboundDeliveryID,
					WorkspaceID:         testWorkspaceID,
					InstallGeneration:   uuidPointer(testInstallGeneration),
					ExternalWorkspaceID: "T-original",
					IdempotencyKey:      "slack:bound-delivery",
					ExternalChannelID:   "C-original",
					Content:             &content,
					Status:              "failed",
					AttemptCount:        2,
				},
			}
			sender := &messageSenderStub{externalMessageID: "171.200"}
			processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender)

			recovered, err := processor.RecoverPendingEvents(context.Background())

			if err != nil {
				t.Fatalf("RecoverPendingEvents() error = %v, want stale delivery cancellation", err)
			}
			if recovered != 0 || len(sender.messages) != 0 || len(store.completedDeliveries) != 0 {
				t.Fatalf("recovered/sent/completed = %d/%d/%d, want 0/0/0", recovered, len(sender.messages), len(store.completedDeliveries))
			}
			if len(store.failedDeliveries) != 0 || len(store.cancelledDeliveries) != 1 {
				t.Fatalf("failed/cancelled deliveries = %d/%d, want 0/1", len(store.failedDeliveries), len(store.cancelledDeliveries))
			}
			if len(repo.requestedTeamIDs) != 1 || repo.requestedTeamIDs[0] != "T-original" {
				t.Fatalf("installation lookup team IDs = %v, want [T-original]", repo.requestedTeamIDs)
			}
		})
	}
}
