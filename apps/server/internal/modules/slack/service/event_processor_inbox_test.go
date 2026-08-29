package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventProcessorLoadsAndDecryptsCanonicalInboxPayload(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-canonical","event":{"type":"unsupported"}}`)
	record := webhooks.Record{
		ID: testInboundReceiptID,
		Envelope: webhooks.Envelope{
			Provider:               slackWebhookProvider,
			ExternalAccountID:      "T1",
			DeliveryID:             "Ev-canonical",
			WorkspaceID:            testWorkspaceID,
			InstallationID:         testSlackWorkspaceID,
			InstallationGeneration: testInstallGeneration,
		},
		Status: webhooks.StatusPending,
	}
	encrypted, err := processor.webhookPayloads.Seal(context.Background(), slackWebhookPayloadBinding(record), body)
	if err != nil {
		t.Fatalf("seal Slack webhook payload: %v", err)
	}
	record.EncryptedPayload = &encrypted
	store.inboundRecords["Ev-canonical"] = record

	err = processor.ProcessEvent(context.Background(), "T1", "Ev-canonical")

	if err != nil {
		t.Fatalf("ProcessEvent() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "ignored")
}

func TestEventProcessorQueuedLinkSharedPublishesWorkObjectUnfurl(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	storyID := uuid.MustParse("99999999-9999-4999-8999-999999999999")
	createdAt := time.Date(2026, time.August, 9, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(2 * time.Hour)
	storyReader := &eventStoryReaderStub{story: singleStory{
		ID:         storyID,
		SequenceID: 123,
		Title:      "Fix workspace login",
		Priority:   "High",
		Team:       testAllowedTeamID,
		TeamCode:   "WEB",
		Workspace:  testWorkspaceID,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}}
	processor, err := NewEventProcessor(nil, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, EventProcessorConfig{
		WebsiteURL:               "https://app.fortyone.com",
		WebhookPayloadSecret:     testSlackWebhookPayloadSecret,
		CredentialVault:          newTestCredentialVault(t),
		CallLimiter:              &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
		UsageBudget:              &usageBudgetStub{},
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
		StoryReader:              storyReader,
		WebhookInbox:             store,
		WebhookRecovery:          &webhookRecoveryStub{},
	})
	require.NoError(t, err)
	sealTestEventInstallation(t, processor, repo)

	type unfurlCapture struct {
		path          string
		authorization string
		payload       SlackChatUnfurlRequest
		decodeErr     error
	}
	captures := make(chan unfurlCapture, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var payload SlackChatUnfurlRequest
		decodeErr := json.NewDecoder(request.Body).Decode(&payload)
		captures <- unfurlCapture{
			path:          request.URL.Path,
			authorization: request.Header.Get("Authorization"),
			payload:       payload,
			decodeErr:     decodeErr,
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	processor.webClient.client = provider.Client()
	processor.webClient.baseURL = provider.URL

	const eventID = "Ev-link-work-object"
	rawBody := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-link-work-object","event":{"type":"link_shared","user":"U1","channel":"C1","message_ts":"1754700000.123","links":[{"domain":"fortyone.app","url":"https://acme.fortyone.app/work/WEB-123"}]}}`)
	record := webhooks.Record{
		ID: testInboundReceiptID,
		Envelope: webhooks.Envelope{
			Provider:               slackWebhookProvider,
			ExternalAccountID:      "T1",
			DeliveryID:             eventID,
			WorkspaceID:            testWorkspaceID,
			InstallationID:         testSlackWorkspaceID,
			InstallationGeneration: testInstallGeneration,
		},
		Status: webhooks.StatusPending,
	}
	encrypted, err := processor.webhookPayloads.Seal(context.Background(), slackWebhookPayloadBinding(record), rawBody)
	require.NoError(t, err)
	record.EncryptedPayload = &encrypted
	store.inboundRecords[eventID] = record

	require.NoError(t, processor.ProcessEvent(context.Background(), "T1", eventID))

	var capture unfurlCapture
	select {
	case capture = <-captures:
	case <-time.After(time.Second):
		t.Fatal("queued link_shared event did not reach chat.unfurl")
	}
	require.NoError(t, capture.decodeErr)
	require.Equal(t, "/chat.unfurl", capture.path)
	require.Equal(t, "Bearer "+testSlackBotAccessToken, capture.authorization)
	require.Equal(t, "C1", capture.payload.Channel)
	require.Equal(t, "1754700000.123", capture.payload.TS)
	require.False(t, capture.payload.UserAuthRequired)
	require.NotNil(t, capture.payload.Metadata)
	require.Len(t, capture.payload.Metadata.Entities, 1)
	entity := capture.payload.Metadata.Entities[0]
	require.Equal(t, slackTaskEntityType, entity.EntityType)
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", entity.AppUnfurlURL)
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", entity.URL)
	require.Equal(t, "acme:"+storyID.String(), entity.ExternalRef.ID)
	require.Equal(t, slackStoryExternalRefType, entity.ExternalRef.Type)
	require.Equal(t, "Fix workspace login", entity.EntityPayload.Attributes.Title.Text)
	require.Equal(t, "WEB-123", entity.EntityPayload.Attributes.DisplayID)
	require.Equal(t, updatedAt.Unix(), entity.EntityPayload.Attributes.MetadataLastModified)
	require.Equal(t, "High", entity.EntityPayload.Fields["priority"].Value)
	require.Equal(t, []uuid.UUID{testWorkspaceID}, storyReader.workspaceIDs)
	require.Equal(t, []string{"WEB-123"}, storyReader.references)
	assertSingleInboundStatus(t, store, "completed")
}

func TestEventProcessorLinkSharedPublishesAuthorizedRequestWorkObject(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	requestID := uuid.MustParse("99999999-9999-4999-8999-999999999998")
	createdAt := time.Date(2026, time.August, 10, 6, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	requestReader := &eventRequestReaderStub{request: integrationRequest{
		ID:              requestID,
		WorkspaceID:     testWorkspaceID,
		TeamID:          testAllowedTeamID,
		Title:           "Investigate mobile login",
		Status:          integrationRequestStatusPending,
		Priority:        "High",
		CreatedByUserID: uuidPointer(testLinkedUserID),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}}
	processor, err := NewEventProcessor(nil, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, EventProcessorConfig{
		WebsiteURL:               "https://fortyone.app",
		WebhookPayloadSecret:     testSlackWebhookPayloadSecret,
		CredentialVault:          newTestCredentialVault(t),
		CallLimiter:              &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
		UsageBudget:              &usageBudgetStub{},
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
		RequestReader:            requestReader,
		WebhookInbox:             store,
		WebhookRecovery:          &webhookRecoveryStub{},
	})
	require.NoError(t, err)
	sealTestEventInstallation(t, processor, repo)

	type requestUnfurlCapture struct {
		path          string
		authorization string
		payload       SlackChatUnfurlRequest
		decodeErr     error
	}
	captures := make(chan requestUnfurlCapture, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var payload SlackChatUnfurlRequest
		decodeErr := json.NewDecoder(request.Body).Decode(&payload)
		captures <- requestUnfurlCapture{
			path:          request.URL.Path,
			authorization: request.Header.Get("Authorization"),
			payload:       payload,
			decodeErr:     decodeErr,
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	processor.webClient.client = provider.Client()
	processor.webClient.baseURL = provider.URL

	requestURL := fmt.Sprintf("https://acme.fortyone.app/teams/%s/requests/%s", testAllowedTeamID, requestID)
	body := fmt.Sprintf(
		`{"type":"event_callback","team_id":"T1","event_id":"Ev-request-unfurl","event":{"type":"link_shared","user":"U1","channel":"C1","message_ts":"1754786400.123","links":[{"domain":"fortyone.app","url":%q}]}}`,
		requestURL,
	)
	require.NoError(t, processSlackRaw(t, processor, []byte(body)))

	select {
	case capture := <-captures:
		require.NoError(t, capture.decodeErr)
		require.Equal(t, "/chat.unfurl", capture.path)
		require.Equal(t, "Bearer "+testSlackBotAccessToken, capture.authorization)
		payload := capture.payload
		require.Equal(t, "C1", payload.Channel)
		require.Equal(t, "1754786400.123", payload.TS)
		require.NotNil(t, payload.Metadata)
		require.Len(t, payload.Metadata.Entities, 1)
		entity := payload.Metadata.Entities[0]
		require.Equal(t, requestURL, entity.AppUnfurlURL)
		require.Equal(t, requestURL, entity.URL)
		require.Equal(t, slackRequestExternalRefType, entity.ExternalRef.Type)
		require.Equal(t, fmt.Sprintf("acme:%s:%s", testAllowedTeamID, requestID), entity.ExternalRef.ID)
		require.Equal(t, "Investigate mobile login", entity.EntityPayload.Attributes.Title.Text)
		require.Equal(t, updatedAt.Unix(), entity.EntityPayload.Attributes.MetadataLastModified)
		require.Equal(t, "Pending", entity.EntityPayload.Fields["status"].Value)
		require.Equal(t, "High", entity.EntityPayload.Fields["priority"].Value)
		require.Equal(t, "U1", entity.EntityPayload.Fields["created_by"].User.UserID)
		require.Nil(t, entity.EntityPayload.Attributes.Title.Edit)
	case <-time.After(time.Second):
		t.Fatal("request link_shared event did not reach chat.unfurl")
	}
	require.Equal(t, []uuid.UUID{testWorkspaceID}, requestReader.workspaceIDs)
	require.Equal(t, []uuid.UUID{requestID}, requestReader.requestIDs)
	require.Equal(t, []uuid.UUID{testLinkedUserID}, requestReader.userIDs)
	assertSingleInboundStatus(t, store, "completed")
}

func TestEventProcessorFailsUnreadableCanonicalInboxPayload(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	invalid := "v1.invalid"
	store.inboundRecords["Ev-poison"] = webhooks.Record{
		ID: testInboundReceiptID,
		Envelope: webhooks.Envelope{
			Provider:               slackWebhookProvider,
			ExternalAccountID:      "T1",
			DeliveryID:             "Ev-poison",
			WorkspaceID:            testWorkspaceID,
			InstallationID:         testSlackWorkspaceID,
			InstallationGeneration: testInstallGeneration,
		},
		Status:           webhooks.StatusPending,
		EncryptedPayload: &invalid,
	}

	err := processor.ProcessEvent(context.Background(), "T1", "Ev-poison")

	if err == nil {
		t.Fatal("ProcessEvent() error = nil, want decrypt failure")
	}
	assertSingleInboundStatus(t, store, "failed")
}

func TestEventProcessorRetriesBusyInboundLease(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	store.processInbound = false
	store.inboundErr = &messagingrepository.LeaseBusyError{
		Resource:   "messaging inbound event",
		RetryAfter: 41 * time.Second,
	}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-busy", "show my work")))

	if !errors.Is(err, messagingrepository.ErrLeaseBusy) {
		t.Fatalf("Process() error = %v, want busy lease", err)
	}
	if retryAfter, ok := messagingrepository.LeaseRetryAfter(err); !ok || retryAfter != 41*time.Second {
		t.Fatalf("LeaseRetryAfter() = %s, %v; want 41s, true", retryAfter, ok)
	}
	if repo.getInstallationCalls != 0 || len(store.completions) != 0 {
		t.Fatalf("busy event performed work: repo calls=%d completions=%d", repo.getInstallationCalls, len(store.completions))
	}
}

func TestEventProcessorSkipsDeliveredOutboundMessage(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	store.processOutbound = false
	store.deliveryStatus = "delivered"
	store.deliveryContent = stringPointer("Here is your work.")
	store.deliveryMessageID = stringPointer("171.200")
	assistant := &assistantStub{}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-delivered", "show my work")))

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(assistant.requests) != 0 || len(sender.messages) != 0 || len(store.completedDeliveries) != 0 {
		t.Fatalf("delivered outbound message performed work: assistant=%d sends=%d completions=%d", len(assistant.requests), len(sender.messages), len(store.completedDeliveries))
	}
	lastMessage := store.appendedMessages[len(store.appendedMessages)-1]
	if lastMessage.externalMessageID != "171.200" || lastMessage.role != "assistant" || lastMessage.content != "Here is your work." {
		t.Fatalf("repaired assistant history message = %+v", lastMessage)
	}
}

func TestEventProcessorRetriesBusyOutboundLease(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	store.processOutbound = false
	store.outboundErr = &messagingrepository.LeaseBusyError{
		Resource:   "messaging outbound delivery",
		RetryAfter: 43 * time.Second,
	}
	assistant := &assistantStub{}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-delivery-busy", "show my work")))

	if !errors.Is(err, messagingrepository.ErrLeaseBusy) {
		t.Fatalf("Process() error = %v, want busy lease", err)
	}
	if len(store.completions) != 1 || store.completions[0].status != "failed" {
		t.Fatalf("inbound completions = %+v, want one failed completion", store.completions)
	}
	if len(assistant.requests) != 0 || len(sender.messages) != 0 || len(store.completedDeliveries) != 0 {
		t.Fatalf("busy outbound lease performed work: assistant=%d sends=%d completions=%d", len(assistant.requests), len(sender.messages), len(store.completedDeliveries))
	}
}
