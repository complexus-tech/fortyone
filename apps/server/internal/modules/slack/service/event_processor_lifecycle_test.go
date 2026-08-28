package slack

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventProcessorIgnoresBotAndSubtypeMessages(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "bot message",
			raw:  `{"type":"event_callback","team_id":"T1","event_id":"Ev-bot","event":{"type":"message","channel_type":"im","user":"U-bot","bot_id":"B1","channel":"D1","ts":"10.1","text":"loop"}}`,
		},
		{
			name: "message subtype",
			raw:  `{"type":"event_callback","team_id":"T1","event_id":"Ev-edit","event":{"type":"message","subtype":"message_changed","channel_type":"im","user":"U1","channel":"D1","ts":"10.2","text":"edited"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			store := newEventStoreStub()
			assistant := &assistantStub{}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{}
			processor := newTestEventProcessor(t, repo, store, assistant, access, sender)

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			assertSingleInboundStatus(t, store, "ignored")
			if repo.getInstallationCalls != 0 || len(sender.messages) != 0 || len(assistant.requests) != 0 {
				t.Fatalf("ignored event performed side effects: repo calls=%d sends=%d assistant calls=%d", repo.getInstallationCalls, len(sender.messages), len(assistant.requests))
			}
		})
	}
}

func TestEventProcessorDeactivatesInstallationForLifecycleEvents(t *testing.T) {
	tests := []struct {
		name                 string
		raw                  string
		installationLookups  int
		wantDeactivatedTeams []string
	}{
		{
			name:                 "app uninstalled",
			raw:                  `{"type":"event_callback","team_id":"T1","event_id":"Ev-uninstall","event_time":1700000001,"event":{"type":"app_uninstalled"}}`,
			installationLookups:  1,
			wantDeactivatedTeams: []string{"T1"},
		},
		{
			name:                 "installed bot token revoked",
			raw:                  `{"type":"event_callback","team_id":"T1","event_id":"Ev-revoked","event_time":1700000001,"event":{"type":"tokens_revoked","tokens":{"oauth":[],"bot":["B1"]}}}`,
			installationLookups:  1,
			wantDeactivatedTeams: []string{"T1"},
		},
		{
			name:                "oauth user token revoked",
			raw:                 `{"type":"event_callback","team_id":"T1","event_id":"Ev-user-revoked","event_time":1700000001,"event":{"type":"tokens_revoked","tokens":{"oauth":["U1"],"bot":[]}}}`,
			installationLookups: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			store := newEventStoreStub()
			processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			assertSingleInboundStatus(t, store, "completed")
			if strings.Join(repo.deactivatedTeamIDs, ",") != strings.Join(test.wantDeactivatedTeams, ",") {
				t.Fatalf("deactivated team IDs = %v, want %v", repo.deactivatedTeamIDs, test.wantDeactivatedTeams)
			}
			if repo.getInstallationCalls != test.installationLookups {
				t.Fatalf("GetSlackWorkspaceByTeamID() calls = %d, want %d", repo.getInstallationCalls, test.installationLookups)
			}
			for _, generation := range repo.deactivatedGenerations {
				if generation != testInstallGeneration {
					t.Fatalf("deactivated generation = %s, want %s", generation, testInstallGeneration)
				}
			}
		})
	}
}

func TestEventProcessorIgnoresLifecycleEventsFromPriorInstallation(t *testing.T) {
	tests := []struct {
		name              string
		raw               string
		receiptGeneration uuid.UUID
	}{
		{
			name:              "missing provider event time",
			raw:               `{"type":"event_callback","team_id":"T1","event_id":"Ev-missing-time","event":{"type":"app_uninstalled"}}`,
			receiptGeneration: testInstallGeneration,
		},
		{
			name:              "event predates current authorization",
			raw:               `{"type":"event_callback","team_id":"T1","event_id":"Ev-stale-time","event_time":1699999999,"event":{"type":"app_uninstalled"}}`,
			receiptGeneration: testInstallGeneration,
		},
		{
			name:              "receipt belongs to prior generation",
			raw:               `{"type":"event_callback","team_id":"T1","event_id":"Ev-stale-generation","event_time":1700000001,"event":{"type":"app_uninstalled"}}`,
			receiptGeneration: uuid.MustParse("88888888-8888-4888-8888-888888888888"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			store := newEventStoreStub()
			store.installGeneration = uuidPointer(test.receiptGeneration)
			processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			assertSingleInboundStatus(t, store, "ignored")
			if len(repo.deactivatedTeamIDs) != 0 {
				t.Fatalf("stale lifecycle event deactivated teams = %v", repo.deactivatedTeamIDs)
			}
		})
	}
}

func TestEventProcessorSendsPrivateAccountLinkForUnlinkedMention(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
	processor.website = "https://fortyone.app"

	err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-link", "<@B1> show my work")))
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(store.nonces) != 1 {
		t.Fatalf("nonce count = %d, want 1", len(store.nonces))
	}
	nonce := store.nonces[0]
	if nonce.Provider != "slack" || nonce.Purpose != "account_link" || nonce.WorkspaceID != testWorkspaceID || nonce.ExternalWorkspaceID != "T1" || nonce.ExternalUserID != "U1" {
		t.Fatalf("nonce binding = %+v", nonce)
	}
	if len(nonce.NonceHash) != 32 || !nonce.ExpiresAt.After(time.Now()) {
		t.Fatalf("nonce hash length/expires = %d/%v", len(nonce.NonceHash), nonce.ExpiresAt)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("send count = %d, want 1", len(sender.messages))
	}
	message := sender.messages[0]
	if message.Ephemeral || message.UserID != "U1" || message.ChannelID != "U1" || message.ThreadTS != "" {
		t.Fatalf("account-link message routing = %+v", message)
	}
	if len(store.outboundInputs) != 1 || store.outboundInputs[0].ExternalWorkspaceID != "T1" || store.outboundInputs[0].ExternalChannelID != "U1" || store.outboundInputs[0].ExternalThreadID != "" {
		t.Fatalf("persisted account-link routing = %+v, want private user destination", store.outboundInputs)
	}
	if !strings.Contains(message.Text, "https://acme.fortyone.app/settings/integrations/slack?slack_link_token=") {
		t.Fatalf("account-link message = %q", message.Text)
	}
	if strings.Contains(message.Text, testSlackBotAccessToken) {
		t.Fatal("account-link message leaked the bot token")
	}
	if len(assistant.requests) != 0 || len(access.workspaceIDs) != 0 {
		t.Fatalf("unlinked request reached access/assistant: access=%d assistant=%d", len(access.workspaceIDs), len(assistant.requests))
	}
}
