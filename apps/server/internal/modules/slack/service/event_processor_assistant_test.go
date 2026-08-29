package slack

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventProcessorPersistsAndRoutesLinkedAssistantResponse(t *testing.T) {
	tests := []struct {
		name                      string
		raw                       string
		wantPrompt                string
		wantChannelID             string
		wantConversationChannelID string
		wantConversationID        string
		wantReplyTS               string
		wantEphemeral             bool
		wantConversationTurn      int
	}{
		{
			name:                      "direct message",
			raw:                       directMessageEvent("Ev-dm", "What is due?"),
			wantPrompt:                "What is due?",
			wantChannelID:             "D1",
			wantConversationChannelID: "D1",
			wantConversationID:        "dm:D1",
			wantReplyTS:               "",
			wantEphemeral:             false,
			wantConversationTurn:      2,
		},
		{
			name:                      "public mention",
			raw:                       mentionEvent("Ev-mention", "<@B1> What is due?"),
			wantPrompt:                "What is due?",
			wantChannelID:             "C1",
			wantConversationChannelID: "C1",
			wantConversationID:        "10.1",
			wantReplyTS:               "10.1",
			wantEphemeral:             false,
			wantConversationTurn:      2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			store := newEventStoreStub()
			oldExternalID := "9.1"
			currentExternalID := "10.1"
			store.history = []messageRecord{
				{ExternalMessageID: &oldExternalID, Role: "user", Content: "Earlier question"},
				{Role: "assistant", Content: "Earlier answer"},
				// The repository returns the current message because the processor
				// persists it before loading history. It must not also be sent as a
				// conversation turn alongside Request.Prompt.
				{ExternalMessageID: &currentExternalID, Role: "user", Content: test.wantPrompt},
			}
			assistant := &assistantStub{response: AssistantResponse{
				Text: "Two stories are due today.",
				Usage: AssistantUsage{
					InputTokens: 120, OutputTokens: 30, TotalTokens: 150,
				},
			}}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{externalMessageID: "10.2"}
			limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
			contextProvider := processor.contextProvider.(*assistantContextProviderStub)
			contextProvider.runtime = &assistantRuntimeContext{
				Actor:     AssistantActorContext{DisplayName: "Joseph Mukorivo", Username: "joseph"},
				LocalTime: time.Date(2026, time.August, 9, 7, 36, 0, 0, time.FixedZone("Africa/Harare", 2*60*60)),
			}

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			assertSingleInboundStatus(t, store, "completed")
			if len(access.workspaceIDs) != 1 || access.workspaceIDs[0] != testWorkspaceID {
				t.Fatalf("access workspace IDs = %v, want [%s]", access.workspaceIDs, testWorkspaceID)
			}
			if len(store.conversations) != 1 {
				t.Fatalf("conversation upserts = %d, want 1", len(store.conversations))
			}
			conversation := store.conversations[0]
			if conversation.WorkspaceID != testWorkspaceID || conversation.UserID != testLinkedUserID || conversation.ExternalWorkspaceID != "T1" || conversation.ExternalChannelID != test.wantConversationChannelID || conversation.ExternalThreadID != test.wantConversationID {
				t.Fatalf("conversation input = %+v", conversation)
			}
			if len(assistant.requests) != 1 {
				t.Fatalf("assistant request count = %d, want 1", len(assistant.requests))
			}
			if len(limiter.inputs) != 1 || limiter.inputs[0].Provider != "slack" || limiter.inputs[0].WorkspaceID != testWorkspaceID || limiter.inputs[0].UserID != testLinkedUserID || limiter.inputs[0].ExternalWorkspaceID != "T1" {
				t.Fatalf("assistant admission input = %+v", limiter.inputs)
			}
			if len(usage.checkWorkspaces) != 1 || usage.checkWorkspaces[0] != testWorkspaceID || len(usage.checkLimits) != 1 || usage.checkLimits[0] != 1_000_000 {
				t.Fatalf("assistant usage checks = %v / %v", usage.checkWorkspaces, usage.checkLimits)
			}
			if len(usage.recordInputs) != 1 {
				t.Fatalf("assistant usage records = %+v, want one", usage.recordInputs)
			}
			recorded := usage.recordInputs[0]
			if recorded.InboundEventID != testInboundReceiptID || recorded.AttemptCount != 1 || recorded.WorkspaceID != testWorkspaceID || recorded.ExternalWorkspaceID != "T1" || recorded.Usage.TotalTokens != 150 {
				t.Fatalf("assistant usage record = %+v", recorded)
			}
			request := assistant.requests[0]
			if request.WorkspaceID != testWorkspaceID || request.UserID != testLinkedUserID || request.Prompt != test.wantPrompt {
				t.Fatalf("assistant request identity/prompt = %+v", request)
			}
			if request.Guidance != "Keep answers concise." || !request.AllowMutations {
				t.Fatalf("assistant settings = guidance %q / mutations %t", request.Guidance, request.AllowMutations)
			}
			if request.RuntimeContext == nil || request.RuntimeContext.Actor.DisplayName != "Joseph Mukorivo" || request.RuntimeContext.LocalTime.Hour() != 7 {
				t.Fatalf("assistant runtime context = %+v", request.RuntimeContext)
			}
			if len(contextProvider.calls) != 1 {
				t.Fatalf("assistant context calls = %+v, want one", contextProvider.calls)
			}
			contextCall := contextProvider.calls[0]
			if contextCall.workspaceID != testWorkspaceID || contextCall.userID != testLinkedUserID || len(contextCall.allowedTeamIDs) != 1 || contextCall.allowedTeamIDs[0] != testAllowedTeamID || contextCall.surface.Provider != "slack" {
				t.Fatalf("assistant context input = %+v", contextCall)
			}
			if test.name == "direct message" && contextCall.surface.Kind != assistantSurfaceDirect {
				t.Fatalf("direct message surface = %q", contextCall.surface.Kind)
			}
			if test.name != "direct message" && contextCall.surface.Kind != assistantSurfaceThread {
				t.Fatalf("channel surface = %q", contextCall.surface.Kind)
			}
			if test.name == "direct message" {
				if len(request.AllowedTeamIDs) != 1 || request.AllowedTeamIDs[0] != testAllowedTeamID || len(request.SharedTeamIDs) != 1 || request.SharedTeamIDs[0] != testAllowedTeamID || conversation.AudienceScope != conversationAudienceActor {
					t.Fatalf("direct assistant scope = allowed %v / shared %v / audience %q", request.AllowedTeamIDs, request.SharedTeamIDs, conversation.AudienceScope)
				}
			} else if len(request.AllowedTeamIDs) != 1 || request.AllowedTeamIDs[0] != testAllowedTeamID || len(request.SharedTeamIDs) != 1 || request.SharedTeamIDs[0] != testAllowedTeamID || conversation.AudienceScope != conversationAudienceChannel {
				t.Fatalf("channel assistant scope = allowed %v / shared %v / audience %q", request.AllowedTeamIDs, request.SharedTeamIDs, conversation.AudienceScope)
			}
			if len(request.Conversation) != test.wantConversationTurn || request.Conversation[0].Role != assistantRoleUser || request.Conversation[1].Role != assistantRoleAssistant {
				t.Fatalf("assistant conversation = %+v", request.Conversation)
			}
			if len(store.appendedMessages) != 2 {
				t.Fatalf("appended message count = %d, want 2", len(store.appendedMessages))
			}
			if got := store.appendedMessages[0]; got.role != "user" || got.content != test.wantPrompt || got.externalMessageID != "10.1" {
				t.Fatalf("persisted user message = %+v", got)
			}
			if got := store.appendedMessages[1]; got.role != "assistant" || got.content != "Two stories are due today." || got.externalMessageID != "10.2" {
				t.Fatalf("persisted assistant message = %+v", got)
			}
			if len(sender.messages) != 1 {
				t.Fatalf("send count = %d, want 1", len(sender.messages))
			}
			message := sender.messages[0]
			if message.ChannelID != test.wantChannelID || message.ThreadTS != test.wantReplyTS || message.Ephemeral != test.wantEphemeral || message.Text != "Two stories are due today." || message.ClientMessageID == "" {
				t.Fatalf("assistant response routing = %+v", message)
			}
			if len(store.completedDeliveries) != 1 || len(store.failedDeliveries) != 0 {
				t.Fatalf("delivery completion/failure counts = %d/%d, want 1/0", len(store.completedDeliveries), len(store.failedDeliveries))
			}
			if len(store.outboundInputs) != 1 || store.outboundInputs[0].ExternalWorkspaceID != "T1" || store.outboundInputs[0].ExternalChannelID != test.wantChannelID || store.outboundInputs[0].ExternalRecipientUserID != "U1" || store.outboundInputs[0].Purpose != "assistant" || store.outboundInputs[0].UserID == nil || *store.outboundInputs[0].UserID != testLinkedUserID {
				t.Fatalf("persisted assistant delivery binding = %+v, want Slack team T1", store.outboundInputs)
			}
			payload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
			if err != nil {
				t.Fatalf("DecodeSlackProviderPayload() error = %v", err)
			}
			if payload.Authorization == nil || payload.Authorization.ActorUserID == nil || *payload.Authorization.ActorUserID != testLinkedUserID || len(payload.Authorization.AllowedTeamIDs) != 1 || payload.Authorization.AllowedTeamIDs[0] != testAllowedTeamID || len(payload.Authorization.SharedTeamIDs) != 1 || payload.Authorization.SharedTeamIDs[0] != testAllowedTeamID {
				t.Fatalf("persisted assistant authorization = %+v", payload.Authorization)
			}
		})
	}
}

func TestEventProcessorUnmappedChannelKeepsPersonalScopeWithoutSharedAccess(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.sharedTeamIDs = []uuid.UUID{}
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "Your work is ready."}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-unmapped-personal-scope", "<@B1> show my work")))

	require.NoError(t, err)
	require.Len(t, assistant.requests, 1)
	require.Equal(t, []uuid.UUID{testAllowedTeamID}, assistant.requests[0].AllowedTeamIDs)
	require.Empty(t, assistant.requests[0].SharedTeamIDs)
	require.Len(t, store.conversations, 1)
	require.Equal(
		t,
		assistantAudienceFingerprint([]uuid.UUID{testAllowedTeamID}, nil),
		store.conversations[0].AudienceFingerprint,
	)
	require.Len(t, store.outboundInputs, 1)
	payload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.NotNil(t, payload.Authorization)
	require.Equal(t, []uuid.UUID{testAllowedTeamID}, payload.Authorization.AllowedTeamIDs)
	require.Empty(t, payload.Authorization.SharedTeamIDs)
	require.Len(t, sender.messages, 1)
}

func TestEventProcessorCancelsChannelResponseWhenSharedAudienceNarrowsDuringModelCall(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "Team-wide answer"}}
	assistant.onRespond = func() {
		repo.sharedTeamIDs = []uuid.UUID{}
	}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-shared-scope-narrowed", "<@B1> what did everyone finish?")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "ignored")
	require.Len(t, assistant.requests, 1)
	require.Equal(t, []uuid.UUID{testAllowedTeamID}, assistant.requests[0].SharedTeamIDs)
	require.Empty(t, sender.messages)
	require.Len(t, store.cancelledDeliveries, 1)
}

func TestEventProcessorKeepsAgentAvailableWhenLegacySettingChangesDuringModelCall(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "Private team answer"}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-settings-race", "<@B1> show private work")))
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(sender.messages) != 1 || len(store.cancelledDeliveries) != 0 {
		t.Fatalf("sent/cancelled = %d/%d, want 1/0", len(sender.messages), len(store.cancelledDeliveries))
	}
}

func TestEventProcessorRetriesWhenAuthoritativeAssistantContextCannotLoad(t *testing.T) {
	t.Parallel()

	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "must not be used"}}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	contextProvider := processor.contextProvider.(*assistantContextProviderStub)
	contextProvider.err = errors.New("context database unavailable")

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-context-failed", "What time is it?")))

	if err == nil || !strings.Contains(err.Error(), "context database unavailable") {
		t.Fatalf("Process() error = %v, want context failure", err)
	}
	if len(assistant.requests) != 0 || len(sender.messages) != 0 {
		t.Fatalf("context failure reached model or delivery: requests=%d sends=%d", len(assistant.requests), len(sender.messages))
	}
	if len(store.failedDeliveries) != 1 {
		t.Fatalf("failed deliveries = %+v, want one retryable failure", store.failedDeliveries)
	}
}

func TestEventProcessorContinuesSubscribedChannelThread(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		channelID string
	}{
		{
			name:      "public channel",
			raw:       channelThreadEvent("Ev-channel-followup", "U1", "Only show urgent items"),
			channelID: "C1",
		},
		{
			name:      "private channel",
			raw:       privateChannelThreadEvent("Ev-private-followup", "U1", "Only show urgent items"),
			channelID: "G1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Unix(1_800_000_000, 0).UTC()
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			repo.installation.AuthorizedAt = now.Add(-time.Hour)
			store := newEventStoreStub()
			store.conversation = conversationRecord{
				ID:        testConversationID,
				UpdatedAt: now,
			}
			previousQuestionID := "10.1"
			currentQuestionID := "10.2"
			store.history = []messageRecord{
				{ExternalMessageID: &previousQuestionID, Role: "user", Content: "What work is pending?"},
				{Role: "assistant", Content: "You have four open stories."},
				{ExternalMessageID: &currentQuestionID, Role: "user", Content: "Only show urgent items"},
			}
			assistant := &assistantStub{response: AssistantResponse{Text: "ENG-3 is urgent."}}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{externalMessageID: "10.3"}
			processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
			processor.clock = fixedClock{now: now}

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			assertSingleInboundStatus(t, store, "completed")
			if len(store.conversationLookups) != 1 {
				t.Fatalf("conversation lookups = %d, want 1", len(store.conversationLookups))
			}
			lookup := store.conversationLookups[0]
			if lookup.WorkspaceID != testWorkspaceID || lookup.UserID != testLinkedUserID || lookup.ExternalWorkspaceID != "T1" || lookup.ExternalChannelID != test.channelID || lookup.ExternalThreadID != "10.1" {
				t.Fatalf("conversation lookup = %+v", lookup)
			}
			if len(assistant.requests) != 1 || assistant.requests[0].Prompt != "Only show urgent items" || len(assistant.requests[0].Conversation) != 2 {
				t.Fatalf("assistant requests = %+v", assistant.requests)
			}
			if len(sender.messages) != 1 {
				t.Fatalf("sent messages = %d, want 1", len(sender.messages))
			}
			message := sender.messages[0]
			if message.ChannelID != test.channelID || message.ThreadTS != "10.1" || message.Text != "ENG-3 is urgent." || message.Ephemeral {
				t.Fatalf("thread response = %+v", message)
			}
		})
	}
}

func TestEventProcessorKeepsSubscribedThreadOperationalNoticePrivate(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.installation.AuthorizedAt = now.Add(-time.Hour)
	store := newEventStoreStub()
	store.conversation = conversationRecord{
		ID:        testConversationID,
		UpdatedAt: now,
	}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: false}, sender)
	processor.clock = fixedClock{now: now}

	if err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-channel-access", "U1", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	assertSingleInboundStatus(t, store, "completed")
	if len(sender.messages) != 1 || sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "" || sender.messages[0].Text != "Maya is available on FortyOne paid plans and active trials." {
		t.Fatalf("private operational notice = %+v", sender.messages)
	}
}

func TestEventProcessorKeepsSubscribedThreadRateLimitPrivate(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.installation.AuthorizedAt = now.Add(-time.Hour)
	store := newEventStoreStub()
	store.conversation = conversationRecord{
		ID:        testConversationID,
		UpdatedAt: now,
	}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{LimitedScope: "user"}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender, limiter, &usageBudgetStub{})
	processor.clock = fixedClock{now: now}

	if err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-channel-rate", "U1", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	assertSingleInboundStatus(t, store, "completed")
	if len(sender.messages) != 1 || sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "" || sender.messages[0].Text != assistantUserRateLimitReply {
		t.Fatalf("private rate-limit notice = %+v", sender.messages)
	}
	if len(store.destinationUpdates) != 1 || store.destinationUpdates[0].channelID != "U1" || store.destinationUpdates[0].threadID != "" {
		t.Fatalf("private rate-limit destination update = %+v", store.destinationUpdates)
	}
}

func TestEventProcessorIgnoresUnsubscribedChannelThreadMessages(t *testing.T) {
	tests := []struct {
		name            string
		raw             string
		linkedUserID    *uuid.UUID
		conversation    conversationRecord
		authorizedAt    time.Time
		wantLookupCount int
	}{
		{
			name:            "unlinked actor",
			raw:             channelThreadEvent("Ev-thread-unlinked", "U1", "show my work"),
			wantLookupCount: 0,
		},
		{
			name:            "thread was never subscribed",
			raw:             channelThreadEvent("Ev-thread-new", "U1", "show my work"),
			linkedUserID:    uuidPointer(testLinkedUserID),
			wantLookupCount: 1,
		},
		{
			name:            "thread belongs to a different actor",
			raw:             channelThreadEvent("Ev-thread-other-actor", "U2", "show my work"),
			linkedUserID:    uuidPointer(testLinkedUserID),
			wantLookupCount: 1,
		},
		{
			name:         "thread predates current installation",
			raw:          channelThreadEvent("Ev-thread-old-install", "U1", "show my work"),
			linkedUserID: uuidPointer(testLinkedUserID),
			conversation: conversationRecord{
				ID:        testConversationID,
				UpdatedAt: testSlackAuthorizedAt.Add(-time.Second),
			},
			authorizedAt:    testSlackAuthorizedAt,
			wantLookupCount: 1,
		},
		{
			name:         "thread subscription expired",
			raw:          channelThreadEvent("Ev-thread-expired", "U1", "show my work"),
			linkedUserID: uuidPointer(testLinkedUserID),
			conversation: conversationRecord{
				ID:        testConversationID,
				UpdatedAt: testSlackAuthorizedAt.Add(-31 * 24 * time.Hour),
			},
			authorizedAt:    testSlackAuthorizedAt.Add(-60 * 24 * time.Hour),
			wantLookupCount: 1,
		},
		{
			name:         "duplicate broad event for explicit mention",
			raw:          channelThreadEvent("Ev-thread-mention-copy", "U1", "<@B1> show my work"),
			linkedUserID: uuidPointer(testLinkedUserID),
			conversation: conversationRecord{
				ID:        testConversationID,
				UpdatedAt: testSlackAuthorizedAt,
			},
			wantLookupCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = test.linkedUserID
			if !test.authorizedAt.IsZero() {
				repo.installation.AuthorizedAt = test.authorizedAt
			}
			store := newEventStoreStub()
			store.conversation = test.conversation
			assistant := &assistantStub{}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{}
			limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			assertSingleInboundStatus(t, store, "ignored")
			if len(store.conversationLookups) != test.wantLookupCount {
				t.Fatalf("conversation lookups = %d, want %d", len(store.conversationLookups), test.wantLookupCount)
			}
			if len(store.nonces) != 0 || len(store.conversations) != 0 || len(store.outboundInputs) != 0 || len(access.workspaceIDs) != 0 || len(limiter.inputs) != 0 || len(usage.checkWorkspaces) != 0 || len(assistant.requests) != 0 || len(sender.messages) != 0 {
				t.Fatalf("ignored thread performed work: nonces=%d conversations=%d outbound=%d access=%d limits=%d usage=%d assistant=%d sends=%d", len(store.nonces), len(store.conversations), len(store.outboundInputs), len(access.workspaceIDs), len(limiter.inputs), len(usage.checkWorkspaces), len(assistant.requests), len(sender.messages))
			}
		})
	}
}
