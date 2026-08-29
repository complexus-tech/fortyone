package slack

import (
	"errors"
	"testing"
)

func TestEventProcessorSetsNativeThinkingStatusBeforeAssistantResponse(t *testing.T) {
	repository := newEventRepositoryStub()
	repository.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistantText := "That's **WEB-545**.\n\n- **Status:** Todo"
	assistant := &assistantStub{response: AssistantResponse{Text: assistantText}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repository, store, assistant, &accessCheckerStub{allowed: true}, sender)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)
	assistant.onRespond = func() {
		if len(statusSetter.calls) != 1 || statusSetter.calls[0].status != slackAssistantThinkingStatus {
			t.Fatalf("status calls before Respond() = %+v", statusSetter.calls)
		}
	}

	if err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-thinking", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(statusSetter.calls) != 2 {
		t.Fatalf("status calls = %+v, want thinking then explicit clear", statusSetter.calls)
	}
	call := statusSetter.calls[0]
	if call.botToken != "xoxb-test-token" || call.channel != "D1" || call.threadTS != "10.1" ||
		call.status != slackAssistantThinkingStatus {
		t.Fatalf("thinking status call = %+v", call)
	}
	if clearCall := statusSetter.calls[1]; clearCall.channel != call.channel ||
		clearCall.threadTS != call.threadTS || clearCall.status != "" {
		t.Fatalf("clear status call = %+v, thinking call = %+v", clearCall, call)
	}
	if len(sender.messages) != 1 || !sender.messages[0].StandardMarkdown || sender.messages[0].Text != assistantText {
		t.Fatalf("assistant Markdown message = %+v", sender.messages)
	}
}

func TestEventProcessorRetriesNativeThinkingStatusClearOnce(t *testing.T) {
	repository := newEventRepositoryStub()
	repository.linkedUserID = uuidPointer(testLinkedUserID)
	processor := newTestEventProcessor(
		t,
		repository,
		newEventStoreStub(),
		&assistantStub{response: AssistantResponse{Text: "Done."}},
		&accessCheckerStub{allowed: true},
		&messageSenderStub{externalMessageID: "10.2"},
	)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)
	statusSetter.errors = []error{nil, errors.New("temporary clear failure"), nil}

	if err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-thinking-clear-retry", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(statusSetter.calls) != 3 {
		t.Fatalf("status calls = %+v, want thinking, failed clear, deferred clear retry", statusSetter.calls)
	}
	if statusSetter.calls[0].status != slackAssistantThinkingStatus ||
		statusSetter.calls[1].status != "" || statusSetter.calls[2].status != "" {
		t.Fatalf("status calls = %+v", statusSetter.calls)
	}
}

func TestEventProcessorClearsNativeThinkingStatusWhenAssistantFails(t *testing.T) {
	repository := newEventRepositoryStub()
	repository.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	responseErr := errors.New("temporary assistant failure")
	assistant := &assistantStub{err: responseErr}
	processor := newTestEventProcessor(
		t,
		repository,
		store,
		assistant,
		&accessCheckerStub{allowed: true},
		&messageSenderStub{},
	)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-thinking-error", "show my work")))
	if !errors.Is(err, responseErr) {
		t.Fatalf("Process() error = %v, want %v", err, responseErr)
	}
	if len(statusSetter.calls) != 2 || statusSetter.calls[0].status != slackAssistantThinkingStatus ||
		statusSetter.calls[1].status != "" {
		t.Fatalf("status calls = %+v, want thinking then explicit clear", statusSetter.calls)
	}
}
