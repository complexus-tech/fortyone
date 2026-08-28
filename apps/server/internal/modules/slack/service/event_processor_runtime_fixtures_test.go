package slack

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

type eventStoreStub struct {
	processInbound  bool
	processOutbound bool
	inboundErr      error
	outboundErr     error
	deliveryStatus  string
	history         []messageRecord
	conversation    conversationRecord
	conversationErr error

	startedEventIDs     []string
	completions         []inboundCompletion
	nonces              []nonceInput
	conversations       []conversationInput
	conversationLookups []conversationInput
	appendedMessages    []appendedMessage
	outboundInputs      []outboundDeliveryInput
	setDeliveryContents []string
	destinationUpdates  []deliveryDestination
	completedDeliveries []completedDelivery
	failedDeliveries    []failedDelivery
	deliveryContent     *string
	deliveryMessageID   *string
	deliveryExpiresAt   *time.Time
	deliveryChannelID   string
	deliveryThreadID    *string
	recoverableOutbound []outboundDeliveryRecord
	inboundRecords      map[string]webhooks.Record
	getInboundErr       error
	installGeneration   *uuid.UUID
	cancelledDeliveries []uuid.UUID
}

func newEventStoreStub() *eventStoreStub {
	generation := testInstallGeneration
	return &eventStoreStub{
		processInbound:    true,
		processOutbound:   true,
		deliveryStatus:    "delivering",
		inboundRecords:    make(map[string]webhooks.Record),
		installGeneration: &generation,
	}
}

func (s *eventStoreStub) CreateNonce(_ context.Context, input nonceInput) error {
	s.nonces = append(s.nonces, input)
	return nil
}

func (s *eventStoreStub) ListRecoverableOutboundDeliveries(_ context.Context, _ string, limit int) ([]outboundDeliveryRecord, error) {
	if limit > len(s.recoverableOutbound) {
		limit = len(s.recoverableOutbound)
	}
	return append([]outboundDeliveryRecord(nil), s.recoverableOutbound[:limit]...), nil
}

func (s *eventStoreStub) UpsertConversation(_ context.Context, input conversationInput) (uuid.UUID, error) {
	s.conversations = append(s.conversations, input)
	return testConversationID, nil
}

func (s *eventStoreStub) FindConversation(_ context.Context, input conversationInput) (conversationRecord, error) {
	s.conversationLookups = append(s.conversationLookups, input)
	if s.conversationErr != nil {
		return conversationRecord{}, s.conversationErr
	}
	if s.conversation.ID == uuid.Nil {
		return conversationRecord{}, errMessagingRecordNotFound
	}
	return s.conversation, nil
}

func (s *eventStoreStub) FindChannelConversation(ctx context.Context, input conversationInput) (conversationRecord, error) {
	input.AudienceScope = conversationAudienceChannel
	return s.FindConversation(ctx, input)
}

func (s *eventStoreStub) AppendMessage(_ context.Context, conversationID uuid.UUID, externalMessageID, role, content string) error {
	s.appendedMessages = append(s.appendedMessages, appendedMessage{
		conversationID:    conversationID,
		externalMessageID: externalMessageID,
		role:              role,
		content:           content,
	})
	return nil
}

func (s *eventStoreStub) ListRecentMessages(_ context.Context, _ uuid.UUID, _ int) ([]messageRecord, error) {
	return s.history, nil
}

func (s *eventStoreStub) StartOutboundDelivery(_ context.Context, input outboundDeliveryInput) (outboundDeliveryRecord, bool, error) {
	s.outboundInputs = append(s.outboundInputs, input)
	if s.deliveryExpiresAt == nil {
		s.deliveryExpiresAt = input.ExpiresAt
	}
	externalChannelID := input.ExternalChannelID
	if s.deliveryChannelID != "" {
		externalChannelID = s.deliveryChannelID
	}
	externalThreadID := s.deliveryThreadID
	if s.deliveryChannelID == "" && externalThreadID == nil && input.ExternalThreadID != "" {
		externalThreadID = stringPointer(input.ExternalThreadID)
	}
	return outboundDeliveryRecord{
		ID:                      testOutboundDeliveryID,
		WorkspaceID:             input.WorkspaceID,
		UserID:                  input.UserID,
		InstallGeneration:       input.InstallGeneration,
		ExternalWorkspaceID:     input.ExternalWorkspaceID,
		ExternalRecipientUserID: stringPointer(input.ExternalRecipientUserID),
		ExternalChannelID:       externalChannelID,
		ExternalThreadID:        externalThreadID,
		Status:                  s.deliveryStatus,
		Content:                 s.deliveryContent,
		ProviderPayload:         append([]byte(nil), input.ProviderPayload...),
		ExternalMessageID:       s.deliveryMessageID,
		AttemptCount:            len(s.outboundInputs),
		Purpose:                 input.Purpose,
		ExpiresAt:               s.deliveryExpiresAt,
	}, s.processOutbound, s.outboundErr
}

func (s *eventStoreStub) SetOutboundDeliveryContent(_ context.Context, _ uuid.UUID, content string) error {
	s.setDeliveryContents = append(s.setDeliveryContents, content)
	s.deliveryContent = stringPointer(content)
	return nil
}

func (s *eventStoreStub) SetOutboundDeliveryContentAndProviderPayload(_ context.Context, _ uuid.UUID, content string, providerPayload []byte) error {
	s.setDeliveryContents = append(s.setDeliveryContents, content)
	s.deliveryContent = stringPointer(content)
	if len(s.outboundInputs) > 0 {
		s.outboundInputs[len(s.outboundInputs)-1].ProviderPayload = append([]byte(nil), providerPayload...)
	}
	return nil
}

func (s *eventStoreStub) SetOutboundDeliveryContentAndDestination(_ context.Context, id uuid.UUID, content, externalChannelID, externalThreadID string) error {
	s.setDeliveryContents = append(s.setDeliveryContents, content)
	s.deliveryContent = stringPointer(content)
	s.destinationUpdates = append(s.destinationUpdates, deliveryDestination{
		id:        id,
		channelID: externalChannelID,
		threadID:  externalThreadID,
	})
	return nil
}

func (s *eventStoreStub) CompleteOutboundDelivery(_ context.Context, id uuid.UUID, externalMessageID string) error {
	s.completedDeliveries = append(s.completedDeliveries, completedDelivery{id: id, externalMessageID: externalMessageID})
	return nil
}

func (s *eventStoreStub) FailOutboundDelivery(_ context.Context, id uuid.UUID, message string) error {
	s.failedDeliveries = append(s.failedDeliveries, failedDelivery{id: id, message: message})
	return nil
}

func (s *eventStoreStub) CancelOutboundDelivery(_ context.Context, id uuid.UUID, _ string) error {
	s.cancelledDeliveries = append(s.cancelledDeliveries, id)
	return nil
}

type assistantStub struct {
	response  AssistantResponse
	err       error
	requests  []assistantRequest
	onRespond func()
}

func (a *assistantStub) Respond(_ context.Context, request assistantRequest) (AssistantResponse, error) {
	a.requests = append(a.requests, request)
	if a.onRespond != nil {
		a.onRespond()
	}
	return a.response, a.err
}

type accessCheckerStub struct {
	allowed      bool
	err          error
	workspaceIDs []uuid.UUID
}

type callLimiterStub struct {
	decision AssistantAdmissionDecision
	err      error
	inputs   []AssistantAdmissionInput
}

func (l *callLimiterStub) Admit(_ context.Context, input AssistantAdmissionInput) (AssistantAdmissionDecision, error) {
	l.inputs = append(l.inputs, input)
	return l.decision, l.err
}

type usageBudgetStub struct {
	checkSnapshot   dailyUsageSnapshot
	checkErr        error
	recordSnapshot  dailyUsageSnapshot
	recordErr       error
	checkLimits     []int64
	checkWorkspaces []uuid.UUID
	recordInputs    []dailyUsageRecordInput
	recordLimits    []int64
}

type assistantContextCall struct {
	workspaceID    uuid.UUID
	userID         uuid.UUID
	allowedTeamIDs []uuid.UUID
	surface        assistantRuntimeSurface
	now            time.Time
}

type assistantContextProviderStub struct {
	runtime *assistantRuntimeContext
	err     error
	calls   []assistantContextCall
}

func (p *assistantContextProviderStub) Load(
	_ context.Context,
	workspaceID, userID uuid.UUID,
	allowedTeamIDs []uuid.UUID,
	surface assistantRuntimeSurface,
	now time.Time,
) (*assistantRuntimeContext, error) {
	p.calls = append(p.calls, assistantContextCall{
		workspaceID:    workspaceID,
		userID:         userID,
		allowedTeamIDs: append([]uuid.UUID(nil), allowedTeamIDs...),
		surface:        surface,
		now:            now,
	})
	return p.runtime, p.err
}

func (b *usageBudgetStub) Check(_ context.Context, workspaceID uuid.UUID, limit int64) (dailyUsageSnapshot, error) {
	b.checkWorkspaces = append(b.checkWorkspaces, workspaceID)
	b.checkLimits = append(b.checkLimits, limit)
	return b.checkSnapshot, b.checkErr
}

func (b *usageBudgetStub) Record(_ context.Context, input dailyUsageRecordInput, limit int64) (dailyUsageSnapshot, error) {
	b.recordInputs = append(b.recordInputs, input)
	b.recordLimits = append(b.recordLimits, limit)
	return b.recordSnapshot, b.recordErr
}

func (a *accessCheckerStub) CanUseAssistant(_ context.Context, workspaceID uuid.UUID) (bool, error) {
	a.workspaceIDs = append(a.workspaceIDs, workspaceID)
	return a.allowed, a.err
}

type messageSenderStub struct {
	errors            []error
	externalMessageID string
	messages          []SlackOutboundMessage
	botTokens         []string
}

type assistantStatusCall struct {
	botToken string
	channel  string
	threadTS string
	status   string
}

type assistantStatusSetterStub struct {
	calls  []assistantStatusCall
	errors []error
}

func (s *assistantStatusSetterStub) SetStatus(_ context.Context, botToken, channelID, threadTS, status string) error {
	s.calls = append(s.calls, assistantStatusCall{
		botToken: botToken,
		channel:  channelID,
		threadTS: threadTS,
		status:   status,
	})
	index := len(s.calls) - 1
	if index < len(s.errors) {
		return s.errors[index]
	}
	return nil
}

func (s *messageSenderStub) Send(_ context.Context, botToken string, message SlackOutboundMessage) (string, error) {
	s.botTokens = append(s.botTokens, botToken)
	s.messages = append(s.messages, message)
	index := len(s.messages) - 1
	if index < len(s.errors) && s.errors[index] != nil {
		return "", s.errors[index]
	}
	return s.externalMessageID, nil
}

func newTestEventProcessor(
	t *testing.T,
	repo *eventRepositoryStub,
	store *eventStoreStub,
	assistant *assistantStub,
	access *accessCheckerStub,
	sender *messageSenderStub,
) *EventProcessor {
	return newTestEventProcessorWithBudgets(
		t,
		repo,
		store,
		assistant,
		access,
		sender,
		&callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
		&usageBudgetStub{},
	)
}

func newTestEventProcessorWithBudgets(
	t *testing.T,
	repo *eventRepositoryStub,
	store *eventStoreStub,
	assistant *assistantStub,
	access *accessCheckerStub,
	sender *messageSenderStub,
	limiter *callLimiterStub,
	usage *usageBudgetStub,
) *EventProcessor {
	t.Helper()
	processor, err := NewEventProcessor(nil, repo, store, assistant, access, EventProcessorConfig{
		WebsiteURL:               "https://app.fortyone.com",
		WebhookPayloadSecret:     testSlackWebhookPayloadSecret,
		CredentialVault:          newTestCredentialVault(t),
		CallLimiter:              limiter,
		UsageBudget:              usage,
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
		WebhookInbox:             store,
		WebhookRecovery:          &webhookRecoveryStub{},
	})
	if err != nil {
		t.Fatalf("NewEventProcessor() error = %v", err)
	}
	processor.sender = sender
	processor.statusSetter = &assistantStatusSetterStub{}
	sealTestEventInstallation(t, processor, repo)
	processor.random = bytes.NewReader(make([]byte, 64))
	return processor
}
