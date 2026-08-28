package slack

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestEventProcessorDeniesAssistantWithoutEntitlement(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: false}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)

	if err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-access", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(assistant.requests) != 0 || len(store.conversations) != 0 {
		t.Fatalf("denied request reached conversation/assistant: conversations=%d assistant=%d", len(store.conversations), len(assistant.requests))
	}
	if len(sender.messages) != 1 || sender.messages[0].Ephemeral || sender.messages[0].Text != "Maya is available on FortyOne paid plans and active trials." {
		t.Fatalf("access response = %+v", sender.messages)
	}
}

func TestEventProcessorRejectsOversizedPromptBeforePersistenceOrBudgetUse(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
	text := "<@B1> " + strings.Repeat("x", messaging.MaximumMessageBytes+1)

	if err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-oversized", text))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	assertSingleInboundStatus(t, store, "completed")
	if len(store.conversations) != 0 || len(store.appendedMessages) != 0 || len(assistant.requests) != 0 {
		t.Fatalf("oversized prompt was persisted or sent to the model: conversations=%d messages=%d requests=%d", len(store.conversations), len(store.appendedMessages), len(assistant.requests))
	}
	if len(usage.checkWorkspaces) != 0 || len(usage.recordInputs) != 0 || len(limiter.inputs) != 0 {
		t.Fatalf("oversized prompt consumed budget: checks=%d records=%d admissions=%d", len(usage.checkWorkspaces), len(usage.recordInputs), len(limiter.inputs))
	}
	if len(sender.messages) != 1 || sender.messages[0].Text != assistantMessageTooLargeReply || sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "" {
		t.Fatalf("oversized prompt response = %+v", sender.messages)
	}
	if len(store.outboundInputs) != 1 || store.outboundInputs[0].Purpose != "assistant" || store.outboundInputs[0].Content != assistantMessageTooLargeReply {
		t.Fatalf("oversized prompt durable delivery = %+v", store.outboundInputs)
	}
}

func TestEventProcessorReturnsDurableScopedRateLimitReplies(t *testing.T) {
	tests := []struct {
		name      string
		scope     string
		raw       string
		wantReply string
		wantDM    bool
	}{
		{
			name:      "user",
			scope:     "user",
			raw:       directMessageEvent("Ev-user-rate", "show my work"),
			wantReply: assistantUserRateLimitReply,
		},
		{
			name:      "workspace mention remains private",
			scope:     "workspace",
			raw:       mentionEvent("Ev-workspace-rate", "<@B1> show my work"),
			wantReply: assistantWorkspaceRateReply,
			wantDM:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			store := newEventStoreStub()
			assistant := &assistantStub{}
			sender := &messageSenderStub{externalMessageID: "10.2"}
			limiter := &callLimiterStub{decision: AssistantAdmissionDecision{LimitedScope: test.scope}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

			if err := processSlackRaw(t, processor, []byte(test.raw)); err != nil {
				t.Fatalf("Process() error = %v", err)
			}

			assertSingleInboundStatus(t, store, "completed")
			if len(limiter.inputs) != 1 || len(usage.checkWorkspaces) != 1 {
				t.Fatalf("budget checks = admissions %d, daily %d; want one each", len(limiter.inputs), len(usage.checkWorkspaces))
			}
			if len(assistant.requests) != 0 || len(store.conversations) != 0 || len(store.appendedMessages) != 0 || len(usage.recordInputs) != 0 {
				t.Fatalf("rate-limited request reached persistence/model/usage: conversations=%d messages=%d requests=%d records=%d", len(store.conversations), len(store.appendedMessages), len(assistant.requests), len(usage.recordInputs))
			}
			if len(sender.messages) != 1 || sender.messages[0].Text != test.wantReply {
				t.Fatalf("rate-limit response = %+v", sender.messages)
			}
			if test.wantDM && (sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "") {
				t.Fatalf("mention rate-limit response was not private = %+v", sender.messages[0])
			}
			if test.wantDM && (len(store.destinationUpdates) != 1 || store.destinationUpdates[0].channelID != "U1" || store.destinationUpdates[0].threadID != "") {
				t.Fatalf("mention rate-limit destination update = %+v", store.destinationUpdates)
			}
			if len(store.outboundInputs) != 1 || store.outboundInputs[0].Purpose != "assistant" || len(store.setDeliveryContents) != 1 || store.setDeliveryContents[0] != test.wantReply {
				t.Fatalf("rate-limit durable delivery = inputs %+v, contents %v", store.outboundInputs, store.setDeliveryContents)
			}
		})
	}
}

func TestEventProcessorReturnsDurableDailyUsageLimitReply(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{checkErr: errors.Join(ErrDailyWorkspaceTokenLimit, &messagingrepository.DailyTokenLimitError{
		WorkspaceID: testWorkspaceID,
		Used:        1_000_000,
		Limit:       1_000_000,
	})}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	if err := processSlackRaw(t, processor, []byte(mentionEvent("Ev-daily-limit", "<@B1> show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	assertSingleInboundStatus(t, store, "completed")
	if len(limiter.inputs) != 0 || len(assistant.requests) != 0 || len(store.conversations) != 0 {
		t.Fatalf("daily-limited request progressed: admissions=%d assistant=%d conversations=%d", len(limiter.inputs), len(assistant.requests), len(store.conversations))
	}
	if len(sender.messages) != 1 || sender.messages[0].Text != assistantDailyLimitReply || sender.messages[0].ChannelID != "U1" || sender.messages[0].ThreadTS != "" {
		t.Fatalf("daily-limit response = %+v", sender.messages)
	}
	if len(store.destinationUpdates) != 1 || store.destinationUpdates[0].channelID != "U1" || store.destinationUpdates[0].threadID != "" {
		t.Fatalf("daily-limit destination update = %+v", store.destinationUpdates)
	}
	if len(store.outboundInputs) != 1 || store.outboundInputs[0].Purpose != "assistant" || len(store.setDeliveryContents) != 1 || store.setDeliveryContents[0] != assistantDailyLimitReply {
		t.Fatalf("daily-limit durable delivery = inputs %+v, contents %v", store.outboundInputs, store.setDeliveryContents)
	}
}

func TestEventProcessorDoesNotSwallowBudgetInfrastructureErrors(t *testing.T) {
	tests := []struct {
		name       string
		limiterErr error
		usageErr   error
	}{
		{name: "Redis admission", limiterErr: errors.New("redis unavailable")},
		{name: "daily usage check", usageErr: errors.New("database unavailable")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			store := newEventStoreStub()
			assistant := &assistantStub{}
			sender := &messageSenderStub{}
			limiter := &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}, err: test.limiterErr}
			usage := &usageBudgetStub{checkErr: test.usageErr}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

			err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-budget-infra", "show my work")))

			if err == nil || (!errors.Is(err, test.limiterErr) && !errors.Is(err, test.usageErr)) {
				t.Fatalf("Process() error = %v, want budget infrastructure error", err)
			}
			assertSingleInboundStatus(t, store, "failed")
			if len(sender.messages) != 0 || len(assistant.requests) != 0 || len(store.conversations) != 0 {
				t.Fatalf("budget infrastructure failure progressed: sends=%d assistant=%d conversations=%d", len(sender.messages), len(assistant.requests), len(store.conversations))
			}
		})
	}
}

func TestEventProcessorReportsUnavailableAssistantWithoutRetry(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{err: errors.Join(errors.New("missing API key"), errAssistantNotConfigured)}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
	var logs bytes.Buffer
	processor.log = logger.NewWithJSON(&logs, slog.LevelError, "test")

	if err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-unavailable", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	assertSingleInboundStatus(t, store, "completed")
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].Text, "temporarily unavailable") {
		t.Fatalf("assistant-unavailable response = %+v", sender.messages)
	}
	if len(store.failedDeliveries) != 0 || len(store.completedDeliveries) != 1 {
		t.Fatalf("delivery failure/completion counts = %d/%d, want 0/1", len(store.failedDeliveries), len(store.completedDeliveries))
	}
	require.Contains(t, logs.String(), `"msg":"Slack Maya assistant response failed"`)
	require.Contains(t, logs.String(), `"classification":"not_configured"`)
	require.NotContains(t, logs.String(), "missing API key")
	require.Contains(t, logs.String(), `"slack_event_id":"Ev-unavailable"`)
}

func TestEventProcessorClassifiesOpenAIErrorsAfterRecordingPartialUsage(t *testing.T) {
	tests := []struct {
		name      string
		apiError  *AssistantAPIError
		permanent bool
	}{
		{
			name:      "deterministic bad request becomes durable safe reply",
			apiError:  &AssistantAPIError{StatusCode: http.StatusBadRequest, Code: "invalid_request_error", Message: "invalid model input", RequestID: "req_test_permanent", Permanent: true},
			permanent: true,
		},
		{
			name:     "ordinary rate limit remains retryable",
			apiError: &AssistantAPIError{StatusCode: http.StatusTooManyRequests, Code: "rate_limit_exceeded", Message: "slow down"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			store := newEventStoreStub()
			partialUsage := AssistantUsage{InputTokens: 80, OutputTokens: 20, TotalTokens: 100}
			assistant := &assistantStub{response: AssistantResponse{Usage: partialUsage}, err: test.apiError}
			sender := &messageSenderStub{externalMessageID: "10.2"}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(
				t,
				repo,
				store,
				assistant,
				&accessCheckerStub{allowed: true},
				sender,
				&callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
				usage,
			)
			var logs bytes.Buffer
			processor.log = logger.NewWithJSON(&logs, slog.LevelError, "test")

			err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-openai-error", "show my work")))
			require.Contains(t, logs.String(), `"msg":"Slack Maya assistant response failed"`)
			require.Contains(t, logs.String(), test.apiError.Code)

			if len(usage.recordInputs) != 1 {
				t.Fatalf("partial usage records = %+v, want one", usage.recordInputs)
			}
			recorded := usage.recordInputs[0]
			if recorded.InboundEventID != testInboundReceiptID || recorded.AttemptCount != 1 || recorded.Usage != partialUsage {
				t.Fatalf("partial usage record = %+v", recorded)
			}
			if test.permanent {
				require.Contains(t, logs.String(), `"classification":"permanent_provider_error"`)
				require.Contains(t, logs.String(), `"openai_status_code":400`)
				require.Contains(t, logs.String(), `"openai_request_id":"req_test_permanent"`)
				if err != nil {
					t.Fatalf("Process() error = %v", err)
				}
				assertSingleInboundStatus(t, store, "completed")
				if len(sender.messages) != 1 || sender.messages[0].Text != assistantConfigurationReply || len(store.completedDeliveries) != 1 || len(store.failedDeliveries) != 0 {
					t.Fatalf("permanent error delivery = messages %+v, completed %+v, failed %+v", sender.messages, store.completedDeliveries, store.failedDeliveries)
				}
				if len(store.setDeliveryContents) != 1 || store.setDeliveryContents[0] != assistantConfigurationReply {
					t.Fatalf("permanent error durable content = %v", store.setDeliveryContents)
				}
				return
			}
			require.Contains(t, logs.String(), `"classification":"retryable"`)

			if !errors.Is(err, test.apiError) {
				t.Fatalf("Process() error = %v, want %v", err, test.apiError)
			}
			assertSingleInboundStatus(t, store, "failed")
			if len(sender.messages) != 0 || len(store.completedDeliveries) != 0 || len(store.failedDeliveries) != 1 || len(store.setDeliveryContents) != 0 {
				t.Fatalf("retryable error delivery = messages %+v, completed %+v, failed %+v, content %v", sender.messages, store.completedDeliveries, store.failedDeliveries, store.setDeliveryContents)
			}
		})
	}
}

func TestEventProcessorFailsAndRetriesWhenUsageCannotBeRecorded(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{response: AssistantResponse{
		Text:  "This response must not be sent.",
		Usage: AssistantUsage{InputTokens: 40, OutputTokens: 10, TotalTokens: 50},
	}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	recordErr := errors.New("usage database unavailable")
	usage := &usageBudgetStub{recordErr: recordErr}
	processor := newTestEventProcessorWithBudgets(
		t,
		repo,
		store,
		assistant,
		&accessCheckerStub{allowed: true},
		sender,
		&callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
		usage,
	)

	err := processSlackRaw(t, processor, []byte(directMessageEvent("Ev-usage-write", "show my work")))

	if !errors.Is(err, recordErr) {
		t.Fatalf("Process() error = %v, want %v", err, recordErr)
	}
	assertSingleInboundStatus(t, store, "failed")
	if len(usage.recordInputs) != 1 || len(store.failedDeliveries) != 1 {
		t.Fatalf("usage records/failed deliveries = %d/%d, want 1/1", len(usage.recordInputs), len(store.failedDeliveries))
	}
	if len(store.setDeliveryContents) != 0 || len(sender.messages) != 0 || len(store.completedDeliveries) != 0 {
		t.Fatalf("unaccounted assistant response escaped: content=%v sends=%v completed=%v", store.setDeliveryContents, sender.messages, store.completedDeliveries)
	}
}
