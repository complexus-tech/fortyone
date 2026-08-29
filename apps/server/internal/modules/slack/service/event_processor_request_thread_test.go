package slack

import (
	"context"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type threadSyncStub struct {
	input     InboundProviderCommentInput
	bindInput BindProviderThreadInput
}

func (s *threadSyncStub) IngestInboundProviderComment(_ context.Context, input InboundProviderCommentInput) (bool, error) {
	s.input = input
	return true, nil
}

func (s *threadSyncStub) BindProviderThread(_ context.Context, input BindProviderThreadInput) (ProviderThread, error) {
	s.bindInput = input
	return ProviderThread{ID: uuid.New()}, nil
}

func TestSyncIntegrationRequestThreadReplyPreservesCanonicalThreadAndActor(t *testing.T) {
	syncer := &threadSyncStub{}
	processor := &EventProcessor{threadSync: syncer}
	event := normalizedSlackEvent{
		TeamID: "T1", UserID: "U1", ChannelID: "C1",
		MessageTS: "1710000000.002", ThreadTS: "1710000000.001", Text: "Customer confirmed",
	}

	handled, err := processor.syncIntegrationRequestThreadReply(
		context.Background(),
		slackrepository.SlackWorkspaceRecord{SlackTeamID: "T1", InstallGeneration: testInstallGeneration},
		&testLinkedUserID,
		event,
	)

	if err != nil || !handled {
		t.Fatalf("sync reply handled = %v, error = %v", handled, err)
	}
	if syncer.input.ExternalThreadID != event.ThreadTS || syncer.input.ExternalMessageID != event.MessageTS {
		t.Fatalf("thread binding = %#v, want thread %q message %q", syncer.input, event.ThreadTS, event.MessageTS)
	}
	if syncer.input.AuthorUserID == nil || *syncer.input.AuthorUserID != testLinkedUserID || syncer.input.InstallationGeneration != testInstallGeneration {
		t.Fatalf("actor/install binding = %#v", syncer.input)
	}
	if want := time.Unix(1710000000, 0).UTC(); !syncer.input.CreatedAt.Equal(want) {
		t.Fatalf("created at = %s, want %s", syncer.input.CreatedAt, want)
	}
}

func TestEventProcessorIngestsBoundRequestReplyFromUnlinkedActorWithoutMaya(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = nil
	store := newEventStoreStub()
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-unlinked-request-reply", "U-external", "Customer confirmed")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "completed")
	require.Nil(t, syncer.input.AuthorUserID)
	require.Equal(t, "U-external", syncer.input.ExternalAuthorID)
	require.Equal(t, "10.1", syncer.input.ExternalThreadID)
	require.Equal(t, "10.2", syncer.input.ExternalMessageID)
	require.Empty(t, store.conversationLookups)
	require.Empty(t, assistant.requests)
	require.Empty(t, store.conversations)
	require.Empty(t, store.outboundInputs)
	require.Empty(t, sender.messages)
}

func TestEventProcessorComposesBoundRequestReplyWithSubscribedMayaThread(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.installation.AuthorizedAt = now.Add(-time.Hour)
	store := newEventStoreStub()
	store.conversation = conversationRecord{ID: testConversationID, UpdatedAt: now}
	assistant := &assistantStub{response: AssistantResponse{Text: "The customer confirmation is now part of this thread."}}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	processor.clock = fixedClock{now: now}
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-composed-request-reply", "U1", "Customer confirmed")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "completed")
	require.Equal(t, &testLinkedUserID, syncer.input.AuthorUserID)
	require.Len(t, store.conversationLookups, 1)
	lookup := store.conversationLookups[0]
	require.Equal(t, testLinkedUserID, lookup.UserID)
	require.Equal(t, "C1", lookup.ExternalChannelID)
	require.Equal(t, "10.1", lookup.ExternalThreadID)
	require.Equal(t, conversationAudienceChannel, lookup.AudienceScope)
	require.Equal(t, assistantAudienceFingerprint([]uuid.UUID{testAllowedTeamID}, []uuid.UUID{testAllowedTeamID}), lookup.AudienceFingerprint)
	require.Len(t, assistant.requests, 1)
	require.Equal(t, "Customer confirmed", assistant.requests[0].Prompt)
	require.Len(t, sender.messages, 1)
	require.Equal(t, "10.1", sender.messages[0].ThreadTS)
}

func TestEventProcessorKeepsBoundRequestReplyCommentOnlyWithoutMayaSubscription(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.installation.AuthorizedAt = now.Add(-time.Hour)
	store := newEventStoreStub()
	assistant := &assistantStub{}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	processor.clock = fixedClock{now: now}
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-request-comment-only", "U1", "Customer confirmed")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "completed")
	require.Equal(t, &testLinkedUserID, syncer.input.AuthorUserID)
	require.Len(t, store.conversationLookups, 1)
	require.Equal(t, assistantAudienceFingerprint([]uuid.UUID{testAllowedTeamID}, []uuid.UUID{testAllowedTeamID}), store.conversationLookups[0].AudienceFingerprint)
	require.Empty(t, assistant.requests)
	require.Empty(t, store.conversations)
	require.Empty(t, store.outboundInputs)
	require.Empty(t, sender.messages)
}

func TestEventProcessorSuppressesBroadMentionCopyAfterBoundRequestSync(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	repo.installation.AuthorizedAt = now.Add(-time.Hour)
	store := newEventStoreStub()
	store.conversation = conversationRecord{ID: testConversationID, UpdatedAt: now}
	assistant := &assistantStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.clock = fixedClock{now: now}
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processSlackRaw(t, processor, []byte(channelThreadEvent("Ev-request-mention-copy", "U1", "<@B1> summarize this")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "ignored")
	require.Equal(t, "10.2", syncer.input.ExternalMessageID)
	require.Empty(t, store.conversationLookups)
	require.Empty(t, assistant.requests)
	require.Empty(t, store.outboundInputs)
}
