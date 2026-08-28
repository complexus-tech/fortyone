package slack

import (
	"errors"
	"strings"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	"github.com/google/uuid"
)

func TestEventProcessorPersistedReplyBypassesNewBudgetDenials(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	persistedReply := "The already-accounted answer."
	store.deliveryContent = &persistedReply
	assistant := &assistantStub{}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{LimitedScope: "workspace"}}
	usage := &usageBudgetStub{checkErr: errors.Join(ErrDailyWorkspaceTokenLimit, &messagingrepository.DailyTokenLimitError{
		WorkspaceID: testWorkspaceID,
		Used:        1_000_000,
		Limit:       1_000_000,
	})}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-persisted-budget", "show my work")))

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(usage.checkWorkspaces) != 0 || len(limiter.inputs) != 0 || len(assistant.requests) != 0 || len(usage.recordInputs) != 0 {
		t.Fatalf("persisted reply re-entered budget/model path: checks=%d admissions=%d assistant=%d records=%d", len(usage.checkWorkspaces), len(limiter.inputs), len(assistant.requests), len(usage.recordInputs))
	}
	if len(sender.messages) != 1 || sender.messages[0].Text != persistedReply {
		t.Fatalf("persisted reply sends = %+v", sender.messages)
	}
	if len(store.setDeliveryContents) != 0 || len(store.completedDeliveries) != 1 {
		t.Fatalf("persisted reply content/completion = %v / %+v", store.setDeliveryContents, store.completedDeliveries)
	}
}

func TestEventProcessorPersistedPrivateBudgetNoticeUsesPersistedDestination(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	store.deliveryContent = stringPointer(assistantDailyLimitReply)
	store.deliveryChannelID = "U1"
	assistant := &assistantStub{}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	if err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-persisted-private-budget", "<@B1> show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	assertSingleInboundStatus(t, store, "completed")
	if len(sender.messages) != 1 || sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "" || sender.messages[0].Text != assistantDailyLimitReply {
		t.Fatalf("persisted private budget delivery = %+v", sender.messages)
	}
	if len(usage.checkWorkspaces) != 0 || len(limiter.inputs) != 0 || len(assistant.requests) != 0 || len(store.destinationUpdates) != 0 {
		t.Fatalf("persisted private budget notice re-entered work: checks=%d admissions=%d assistant=%d destinations=%d", len(usage.checkWorkspaces), len(limiter.inputs), len(assistant.requests), len(store.destinationUpdates))
	}
}

func TestEventProcessorDeliveredBudgetNoticeReplayDoesNotPersistConversation(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	store.processOutbound = false
	store.deliveryStatus = "delivered"
	store.deliveryContent = stringPointer(assistantDailyLimitReply)
	store.deliveryMessageID = stringPointer("10.2")
	assistant := &assistantStub{}
	sender := &messageSenderStub{}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-delivered-budget", "show my work")))

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(store.conversations) != 0 || len(store.appendedMessages) != 0 || len(sender.messages) != 0 {
		t.Fatalf("delivered budget notice replay produced side effects: conversations=%d messages=%d sends=%d", len(store.conversations), len(store.appendedMessages), len(sender.messages))
	}
	if len(usage.checkWorkspaces) != 0 || len(limiter.inputs) != 0 || len(assistant.requests) != 0 {
		t.Fatalf("delivered budget notice replay entered budget/model path: checks=%d admissions=%d assistant=%d", len(usage.checkWorkspaces), len(limiter.inputs), len(assistant.requests))
	}
}

func TestEventProcessorPersistedReplyUsesPersistedExpiry(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	persistedReply := "This answer is stale."
	store.deliveryContent = &persistedReply
	expiresAt := time.Unix(1_700_000_000, 0).UTC()
	store.deliveryExpiresAt = &expiresAt
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.clock = fixedClock{now: expiresAt.Add(time.Second)}

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-expired-answer", "show my work")))

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "ignored")
	if len(store.cancelledDeliveries) != 1 || len(processor.callLimiter.(*callLimiterStub).inputs) != 0 || len(processor.usageBudget.(*usageBudgetStub).checkWorkspaces) != 0 {
		t.Fatalf("expired persisted reply cancellation/budgets = %d/%d/%d", len(store.cancelledDeliveries), len(processor.callLimiter.(*callLimiterStub).inputs), len(processor.usageBudget.(*usageBudgetStub).checkWorkspaces))
	}
}

func TestEventProcessorGenericDeliveryUsesPersistedExpiry(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	expiresAt := time.Unix(1_700_000_000, 0).UTC()
	store.deliveryExpiresAt = &expiresAt
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: false}, sender)
	processor.clock = fixedClock{now: expiresAt.Add(time.Second)}

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-expired-access", "show my work")))

	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(store.cancelledDeliveries) != 1 || len(sender.messages) != 0 || len(store.setDeliveryContents) != 0 {
		t.Fatalf("expired generic delivery cancellation/sends/content = %d/%d/%d", len(store.cancelledDeliveries), len(sender.messages), len(store.setDeliveryContents))
	}
}

func TestEventProcessorPropagatesRateLimitAndReusesReplyOnRetry(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{Text: "A persisted answer."}}
	access := &accessCheckerStub{allowed: true}
	rateLimit := &RateLimitError{Method: "chat.postMessage", RetryAfter: 7 * time.Second}
	sender := &messageSenderStub{
		errors:            []error{rateLimit, nil},
		externalMessageID: "10.2",
	}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
	raw := []byte(directMessageEvent("Ev-rate-limit", "What is due?"))

	err := processSlackRaw(t, processor, raw)
	if !errors.Is(err, rateLimit) {
		t.Fatalf("first Process() error = %v, want %v", err, rateLimit)
	}
	if retryAfter, ok := SlackRetryAfter(err); !ok || retryAfter != 7*time.Second {
		t.Fatalf("SlackRetryAfter() = %s, %v; want 7s, true", retryAfter, ok)
	}
	if len(store.failedDeliveries) != 1 || !strings.Contains(store.failedDeliveries[0].message, "rate limited") {
		t.Fatalf("failed deliveries = %+v", store.failedDeliveries)
	}
	if len(store.completions) != 1 || store.completions[0].status != "failed" || store.completions[0].message != "slack.processing_failed" {
		t.Fatalf("first inbound completion = %+v", store.completions)
	}

	if err := processSlackRaw(t, processor, raw); err != nil {
		t.Fatalf("retry Process() error = %v", err)
	}
	if len(assistant.requests) != 1 {
		t.Fatalf("assistant request count = %d, want 1; persisted reply should be reused", len(assistant.requests))
	}
	if len(usage.recordInputs) != 1 || usage.recordInputs[0].AttemptCount != 1 {
		t.Fatalf("usage records = %+v, want exactly first execution", usage.recordInputs)
	}
	if len(limiter.inputs) != 1 || len(usage.checkWorkspaces) != 1 {
		t.Fatalf("retry budget calls = admissions %d, daily checks %d; persisted reply must bypass both", len(limiter.inputs), len(usage.checkWorkspaces))
	}
	if len(store.setDeliveryContents) != 1 || store.setDeliveryContents[0] != "A persisted answer." {
		t.Fatalf("persisted delivery contents = %v", store.setDeliveryContents)
	}
	if len(sender.messages) != 2 || sender.messages[0].Text != sender.messages[1].Text || sender.messages[0].ClientMessageID != sender.messages[1].ClientMessageID {
		t.Fatalf("retry messages differ = %+v", sender.messages)
	}
	if !sender.messages[0].StandardMarkdown || !sender.messages[1].StandardMarkdown {
		t.Fatalf("retry messages lost standard Markdown mode = %+v", sender.messages)
	}
	if len(store.completedDeliveries) != 1 || len(store.completions) != 2 || store.completions[1].status != "completed" {
		t.Fatalf("retry delivery/inbound completion = %+v / %+v", store.completedDeliveries, store.completions)
	}
}

func assertSingleInboundStatus(t *testing.T, store *eventStoreStub, want string) {
	t.Helper()
	if len(store.completions) != 1 || store.completions[0].status != want {
		t.Fatalf("inbound completions = %+v, want one %q completion", store.completions, want)
	}
}

func mentionEvent(eventID, text string) string {
	return `{"type":"event_callback","team_id":"T1","event_id":"` + eventID + `","event":{"type":"app_mention","user":"U1","channel":"C1","ts":"10.1","text":"` + text + `"}}`
}

func directMessageEvent(eventID, text string) string {
	return `{"type":"event_callback","team_id":"T1","event_id":"` + eventID + `","event":{"type":"message","channel_type":"im","user":"U1","channel":"D1","ts":"10.1","text":"` + text + `"}}`
}

func channelThreadEvent(eventID, userID, text string) string {
	return `{"type":"event_callback","team_id":"T1","event_id":"` + eventID + `","event":{"type":"message","channel_type":"channel","user":"` + userID + `","channel":"C1","ts":"10.2","thread_ts":"10.1","text":"` + text + `"}}`
}

func privateChannelThreadEvent(eventID, userID, text string) string {
	return `{"type":"event_callback","team_id":"T1","event_id":"` + eventID + `","event":{"type":"message","channel_type":"group","user":"` + userID + `","channel":"G1","ts":"10.2","thread_ts":"10.1","text":"` + text + `"}}`
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
