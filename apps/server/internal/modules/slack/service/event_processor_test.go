package slack

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testSlackCredentialSecret = "test-slack-credential-secret"

var (
	testSlackWorkspaceID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testWorkspaceID         = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testLinkedUserID        = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testAllowedTeamID       = uuid.MustParse("88888888-8888-4888-8888-888888888888")
	testInboundReceiptID    = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	testConversationID      = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	testOutboundDeliveryID  = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	testInstallGeneration   = uuid.MustParse("77777777-7777-4777-8777-777777777777")
	testSlackBotUserID      = "B1"
	testSlackBotAccessToken = "xoxb-test-token"
	testSlackAuthorizedAt   = time.Unix(1_700_000_000, 0).UTC()
)

type eventRepositoryStub struct {
	installation      slackrepository.SlackWorkspaceRecord
	workspace         slackrepository.WorkspaceRecord
	linkedUserID      *uuid.UUID
	agentSettings     slackrepository.AgentSettingsRecord
	authorizedTeamIDs []uuid.UUID
	sharedTeamIDs     []uuid.UUID

	getInstallationCalls       int
	requestedTeamIDs           []string
	findWorkspaceCalls         int
	findLinkedUserCalls        int
	credentialUpgrades         int
	legacyCredentials          []slackrepository.LegacySlackCredentialRecord
	versionedLegacyCredentials int
	deactivatedTeamIDs         []string
	deactivatedGenerations     []uuid.UUID
	installationErr            error
	recoverableUninstalls      []slackrepository.SlackUninstallRecord
	completedUninstalls        []uuid.UUID
	failedUninstalls           []uuid.UUID
}

func (r *eventRepositoryStub) GetAgentSettings(_ context.Context, _ uuid.UUID) (slackrepository.AgentSettingsRecord, error) {
	return r.agentSettings, nil
}

func (r *eventRepositoryStub) ListAuthorizedChannelTeamIDs(_ context.Context, _, _ uuid.UUID, _ string, _ uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.authorizedTeamIDs...), nil
}

func (r *eventRepositoryStub) GetAuthorizedAssistantChannelTeamScope(
	_ context.Context,
	_, _ uuid.UUID,
	_ string,
	_ uuid.UUID,
) (slackrepository.AssistantChannelTeamScope, error) {
	return slackrepository.AssistantChannelTeamScope{
		AllowedTeamIDs: append([]uuid.UUID(nil), r.authorizedTeamIDs...),
		SharedTeamIDs:  append([]uuid.UUID(nil), r.sharedTeamIDs...),
	}, nil
}

func (r *eventRepositoryStub) ListTeamStatuses(_ context.Context, _ uuid.UUID) ([]slackrepository.StatusRecord, error) {
	return nil, nil
}

func (r *eventRepositoryStub) ListTeamMembers(_ context.Context, _ uuid.UUID) ([]slackrepository.TeamMemberRecord, error) {
	return nil, nil
}

func (r *eventRepositoryStub) FindTeamMemberByID(_ context.Context, _, _ uuid.UUID) (slackrepository.TeamMemberRecord, error) {
	return slackrepository.TeamMemberRecord{}, sql.ErrNoRows
}

func (r *eventRepositoryStub) ListWorkspaceTeamsForUser(_ context.Context, _, _ uuid.UUID) ([]slackrepository.TeamRecord, error) {
	teams := make([]slackrepository.TeamRecord, 0, len(r.authorizedTeamIDs))
	for _, teamID := range r.authorizedTeamIDs {
		teams = append(teams, slackrepository.TeamRecord{ID: teamID})
	}
	return teams, nil
}

func newEventRepositoryStub() *eventRepositoryStub {
	return &eventRepositoryStub{
		installation: slackrepository.SlackWorkspaceRecord{
			ID:                testSlackWorkspaceID,
			WorkspaceID:       testWorkspaceID,
			SlackTeamID:       "T1",
			BotUserID:         &testSlackBotUserID,
			BotAccessToken:    testSlackBotAccessToken,
			CredentialVersion: 0,
			InstallGeneration: testInstallGeneration,
			AuthorizedAt:      testSlackAuthorizedAt,
			IsActive:          true,
		},
		workspace: slackrepository.WorkspaceRecord{
			ID:   testWorkspaceID,
			Slug: "acme",
			Name: "Acme",
		},
		agentSettings: slackrepository.AgentSettingsRecord{
			Guidance: "Keep answers concise.",
		},
		authorizedTeamIDs: []uuid.UUID{testAllowedTeamID},
		sharedTeamIDs:     []uuid.UUID{testAllowedTeamID},
	}
}

func (r *eventRepositoryStub) GetSlackWorkspaceByTeamID(_ context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error) {
	r.getInstallationCalls++
	r.requestedTeamIDs = append(r.requestedTeamIDs, slackTeamID)
	if r.installationErr != nil {
		return slackrepository.SlackWorkspaceRecord{}, r.installationErr
	}
	return r.installation, nil
}

func (r *eventRepositoryStub) GetSlackWorkspace(_ context.Context, _ uuid.UUID) (slackrepository.SlackWorkspaceRecord, error) {
	r.getInstallationCalls++
	return r.installation, nil
}

func (r *eventRepositoryStub) FindWorkspaceByID(_ context.Context, _ uuid.UUID) (slackrepository.WorkspaceRecord, error) {
	r.findWorkspaceCalls++
	return r.workspace, nil
}

func (r *eventRepositoryStub) FindLinkedUserIDBySlackUser(_ context.Context, _ uuid.UUID, _, _ string) (*uuid.UUID, error) {
	r.findLinkedUserCalls++
	return r.linkedUserID, nil
}

func (r *eventRepositoryStub) UpgradeSlackCredential(_ context.Context, _ uuid.UUID, encrypted string, version int) error {
	r.credentialUpgrades++
	r.installation.BotAccessToken = encrypted
	r.installation.CredentialVersion = version
	if len(r.legacyCredentials) > 0 {
		r.legacyCredentials = r.legacyCredentials[1:]
	}
	return nil
}

func (r *eventRepositoryStub) ListLegacySlackCredentials(_ context.Context, limit int) ([]slackrepository.LegacySlackCredentialRecord, error) {
	if limit > len(r.legacyCredentials) {
		limit = len(r.legacyCredentials)
	}
	return append([]slackrepository.LegacySlackCredentialRecord(nil), r.legacyCredentials[:limit]...), nil
}

func (r *eventRepositoryStub) ScrubVersionedLegacySlackCredentials(_ context.Context, limit int) (int, error) {
	if limit > r.versionedLegacyCredentials {
		limit = r.versionedLegacyCredentials
	}
	r.versionedLegacyCredentials -= limit
	return limit, nil
}

func (r *eventRepositoryStub) DeactivateSlackWorkspaceByTeamID(_ context.Context, slackTeamID string, generation uuid.UUID) error {
	r.deactivatedTeamIDs = append(r.deactivatedTeamIDs, slackTeamID)
	r.deactivatedGenerations = append(r.deactivatedGenerations, generation)
	return nil
}

func (r *eventRepositoryStub) ClaimRecoverableSlackUninstalls(_ context.Context, limit int) ([]slackrepository.SlackUninstallRecord, error) {
	if limit > len(r.recoverableUninstalls) {
		limit = len(r.recoverableUninstalls)
	}
	return append([]slackrepository.SlackUninstallRecord(nil), r.recoverableUninstalls[:limit]...), nil
}

func (r *eventRepositoryStub) CompleteSlackUninstall(_ context.Context, id uuid.UUID, _ string) error {
	r.completedUninstalls = append(r.completedUninstalls, id)
	return nil
}

func (r *eventRepositoryStub) FailSlackUninstall(_ context.Context, id uuid.UUID, _ string, _ *time.Time) error {
	r.failedUninstalls = append(r.failedUninstalls, id)
	return nil
}

type eventStoryReaderStub struct {
	story        stories.CoreSingleStory
	workspaceIDs []uuid.UUID
	references   []string
}

type eventRequestReaderStub struct {
	request      integrationrequests.CoreIntegrationRequest
	workspaceIDs []uuid.UUID
	requestIDs   []uuid.UUID
	userIDs      []uuid.UUID
}

func (s *eventRequestReaderStub) GetForUser(_ context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	s.workspaceIDs = append(s.workspaceIDs, workspaceID)
	s.requestIDs = append(s.requestIDs, requestID)
	s.userIDs = append(s.userIDs, userID)
	return s.request, nil
}

func (s *eventStoryReaderStub) QueryByRef(_ context.Context, workspaceID uuid.UUID, reference string) (stories.CoreSingleStory, error) {
	s.workspaceIDs = append(s.workspaceIDs, workspaceID)
	s.references = append(s.references, reference)
	return s.story, nil
}

type inboundCompletion struct {
	id      uuid.UUID
	status  string
	message string
}

type appendedMessage struct {
	conversationID    uuid.UUID
	externalMessageID string
	role              string
	content           string
}

type completedDelivery struct {
	id                uuid.UUID
	externalMessageID string
}

type failedDelivery struct {
	id      uuid.UUID
	message string
}

type deliveryDestination struct {
	id        uuid.UUID
	channelID string
	threadID  string
}

type eventStoreStub struct {
	processInbound  bool
	processOutbound bool
	inboundErr      error
	outboundErr     error
	deliveryStatus  string
	history         []messagingrepository.MessageRecord
	conversation    messagingrepository.ConversationRecord
	conversationErr error

	startedEventIDs     []string
	completions         []inboundCompletion
	nonces              []messagingrepository.NonceInput
	conversations       []messagingrepository.ConversationInput
	conversationLookups []messagingrepository.ConversationInput
	appendedMessages    []appendedMessage
	outboundInputs      []messagingrepository.OutboundDeliveryInput
	setDeliveryContents []string
	destinationUpdates  []deliveryDestination
	completedDeliveries []completedDelivery
	failedDeliveries    []failedDelivery
	deliveryContent     *string
	deliveryMessageID   *string
	deliveryExpiresAt   *time.Time
	deliveryChannelID   string
	deliveryThreadID    *string
	recoverableEvents   []messagingrepository.InboundEventRecord
	releasedRecoveries  []messagingrepository.InboundEventRecord
	recoverableOutbound []messagingrepository.OutboundDeliveryRecord
	inboundRecords      map[string]messagingrepository.InboundEventRecord
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
		inboundRecords:    make(map[string]messagingrepository.InboundEventRecord),
		installGeneration: &generation,
	}
}

func (s *eventStoreStub) CreateNonce(_ context.Context, input messagingrepository.NonceInput) error {
	s.nonces = append(s.nonces, input)
	return nil
}

func (s *eventStoreStub) StartInboundEvent(_ context.Context, provider, externalWorkspaceID, externalEventID string) (messagingrepository.InboundEventRecord, bool, error) {
	if provider != "slack" {
		return messagingrepository.InboundEventRecord{}, false, errors.New("unexpected provider")
	}
	if externalWorkspaceID != "T1" {
		return messagingrepository.InboundEventRecord{}, false, errors.New("unexpected external workspace")
	}
	s.startedEventIDs = append(s.startedEventIDs, externalEventID)
	return messagingrepository.InboundEventRecord{
		ID:                  testInboundReceiptID,
		InstallGeneration:   s.installGeneration,
		ExternalWorkspaceID: externalWorkspaceID,
		ExternalEventID:     externalEventID,
		AttemptCount:        len(s.startedEventIDs),
	}, s.processInbound, s.inboundErr
}

func (s *eventStoreStub) GetInboundEvent(_ context.Context, provider, externalWorkspaceID, externalEventID string) (messagingrepository.InboundEventRecord, error) {
	if provider != "slack" {
		return messagingrepository.InboundEventRecord{}, errors.New("unexpected provider")
	}
	if externalWorkspaceID != "T1" {
		return messagingrepository.InboundEventRecord{}, errors.New("unexpected external workspace")
	}
	if s.getInboundErr != nil {
		return messagingrepository.InboundEventRecord{}, s.getInboundErr
	}
	record, ok := s.inboundRecords[externalEventID]
	if !ok {
		return messagingrepository.InboundEventRecord{}, messagingrepository.ErrNotFound
	}
	return record, nil
}

func (s *eventStoreStub) ClaimRecoverableInboundEvents(_ context.Context, _ string, limit int) ([]messagingrepository.InboundEventRecord, error) {
	if limit > len(s.recoverableEvents) {
		limit = len(s.recoverableEvents)
	}
	return append([]messagingrepository.InboundEventRecord(nil), s.recoverableEvents[:limit]...), nil
}

func (s *eventStoreStub) ReleaseInboundEventRecovery(_ context.Context, id uuid.UUID, generation int) error {
	s.releasedRecoveries = append(s.releasedRecoveries, messagingrepository.InboundEventRecord{
		ID:                 id,
		RecoveryGeneration: generation,
	})
	return nil
}

func (s *eventStoreStub) ListRecoverableOutboundDeliveries(_ context.Context, _ string, limit int) ([]messagingrepository.OutboundDeliveryRecord, error) {
	if limit > len(s.recoverableOutbound) {
		limit = len(s.recoverableOutbound)
	}
	return append([]messagingrepository.OutboundDeliveryRecord(nil), s.recoverableOutbound[:limit]...), nil
}

type eventQueueStub struct {
	payloads []tasks.SlackEventPayload
	errors   []error
}

func (q *eventQueueStub) EnqueueSlackEvent(_ context.Context, payload tasks.SlackEventPayload) error {
	q.payloads = append(q.payloads, payload)
	index := len(q.payloads) - 1
	if index < len(q.errors) {
		return q.errors[index]
	}
	return nil
}

func (s *eventStoreStub) CompleteInboundEvent(_ context.Context, id uuid.UUID, status, message string) error {
	s.completions = append(s.completions, inboundCompletion{id: id, status: status, message: message})
	return nil
}

func (s *eventStoreStub) UpsertConversation(_ context.Context, input messagingrepository.ConversationInput) (uuid.UUID, error) {
	s.conversations = append(s.conversations, input)
	return testConversationID, nil
}

func (s *eventStoreStub) FindConversation(_ context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error) {
	s.conversationLookups = append(s.conversationLookups, input)
	if s.conversationErr != nil {
		return messagingrepository.ConversationRecord{}, s.conversationErr
	}
	if s.conversation.ID == uuid.Nil {
		return messagingrepository.ConversationRecord{}, messagingrepository.ErrNotFound
	}
	return s.conversation, nil
}

func (s *eventStoreStub) FindChannelConversation(ctx context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error) {
	input.AudienceScope = messagingrepository.ConversationAudienceChannel
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

func (s *eventStoreStub) ListRecentMessages(_ context.Context, _ uuid.UUID, _ int) ([]messagingrepository.MessageRecord, error) {
	return s.history, nil
}

func (s *eventStoreStub) StartOutboundDelivery(_ context.Context, input messagingrepository.OutboundDeliveryInput) (messagingrepository.OutboundDeliveryRecord, bool, error) {
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
	return messagingrepository.OutboundDeliveryRecord{
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
	response  messaging.Response
	err       error
	requests  []messaging.Request
	onRespond func()
}

func (a *assistantStub) Respond(_ context.Context, request messaging.Request) (messaging.Response, error) {
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
	decision messagingbudget.AdmissionDecision
	err      error
	inputs   []messagingbudget.AdmissionInput
}

func (l *callLimiterStub) Admit(_ context.Context, input messagingbudget.AdmissionInput) (messagingbudget.AdmissionDecision, error) {
	l.inputs = append(l.inputs, input)
	return l.decision, l.err
}

type usageBudgetStub struct {
	checkSnapshot   messagingrepository.DailyUsageSnapshot
	checkErr        error
	recordSnapshot  messagingrepository.DailyUsageSnapshot
	recordErr       error
	checkLimits     []int64
	checkWorkspaces []uuid.UUID
	recordInputs    []messagingrepository.DailyUsageRecordInput
	recordLimits    []int64
}

type assistantContextCall struct {
	workspaceID    uuid.UUID
	userID         uuid.UUID
	allowedTeamIDs []uuid.UUID
	surface        messaging.RuntimeSurfaceContext
	now            time.Time
}

type assistantContextProviderStub struct {
	runtime *messaging.RuntimeContext
	err     error
	calls   []assistantContextCall
}

func (p *assistantContextProviderStub) Load(
	_ context.Context,
	workspaceID, userID uuid.UUID,
	allowedTeamIDs []uuid.UUID,
	surface messaging.RuntimeSurfaceContext,
	now time.Time,
) (*messaging.RuntimeContext, error) {
	p.calls = append(p.calls, assistantContextCall{
		workspaceID:    workspaceID,
		userID:         userID,
		allowedTeamIDs: append([]uuid.UUID(nil), allowedTeamIDs...),
		surface:        surface,
		now:            now,
	})
	return p.runtime, p.err
}

func (b *usageBudgetStub) Check(_ context.Context, workspaceID uuid.UUID, limit int64) (messagingrepository.DailyUsageSnapshot, error) {
	b.checkWorkspaces = append(b.checkWorkspaces, workspaceID)
	b.checkLimits = append(b.checkLimits, limit)
	return b.checkSnapshot, b.checkErr
}

func (b *usageBudgetStub) Record(_ context.Context, input messagingrepository.DailyUsageRecordInput, limit int64) (messagingrepository.DailyUsageSnapshot, error) {
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
		&callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}},
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
		SecretKey:                testSlackCredentialSecret,
		CallLimiter:              limiter,
		UsageBudget:              usage,
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
	})
	if err != nil {
		t.Fatalf("NewEventProcessor() error = %v", err)
	}
	processor.sender = sender
	processor.statusSetter = &assistantStatusSetterStub{}
	processor.random = bytes.NewReader(make([]byte, 64))
	return processor
}

func TestEventProcessorSetsNativeThinkingStatusBeforeAssistantResponse(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistantText := "That's **WEB-545**.\n\n- **Status:** Todo"
	assistant := &assistantStub{response: messaging.Response{Text: assistantText}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)
	assistant.onRespond = func() {
		if len(statusSetter.calls) != 1 || statusSetter.calls[0].status != slackAssistantThinkingStatus {
			t.Fatalf("status calls before Respond() = %+v", statusSetter.calls)
		}
	}

	if err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-thinking", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(statusSetter.calls) != 2 {
		t.Fatalf("status calls = %+v, want thinking then explicit clear", statusSetter.calls)
	}
	call := statusSetter.calls[0]
	if call.botToken != "xoxb-test-token" || call.channel != "D1" || call.threadTS != "10.1" || call.status != slackAssistantThinkingStatus {
		t.Fatalf("thinking status call = %+v", call)
	}
	if clearCall := statusSetter.calls[1]; clearCall.channel != call.channel || clearCall.threadTS != call.threadTS || clearCall.status != "" {
		t.Fatalf("clear status call = %+v, thinking call = %+v", clearCall, call)
	}
	if len(sender.messages) != 1 || !sender.messages[0].StandardMarkdown || sender.messages[0].Text != assistantText {
		t.Fatalf("assistant Markdown message = %+v", sender.messages)
	}
}

func TestEventProcessorRetriesNativeThinkingStatusClearOnce(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	processor := newTestEventProcessor(
		t,
		repo,
		newEventStoreStub(),
		&assistantStub{response: messaging.Response{Text: "Done."}},
		&accessCheckerStub{allowed: true},
		&messageSenderStub{externalMessageID: "10.2"},
	)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)
	statusSetter.errors = []error{nil, errors.New("temporary clear failure"), nil}

	if err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-thinking-clear-retry", "show my work"))); err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(statusSetter.calls) != 3 {
		t.Fatalf("status calls = %+v, want thinking, failed clear, deferred clear retry", statusSetter.calls)
	}
	if statusSetter.calls[0].status != slackAssistantThinkingStatus || statusSetter.calls[1].status != "" || statusSetter.calls[2].status != "" {
		t.Fatalf("status calls = %+v", statusSetter.calls)
	}
}

func TestEventProcessorClearsNativeThinkingStatusWhenAssistantFails(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	responseErr := errors.New("temporary assistant failure")
	assistant := &assistantStub{err: responseErr}
	processor := newTestEventProcessor(
		t,
		repo,
		store,
		assistant,
		&accessCheckerStub{allowed: true},
		&messageSenderStub{},
	)
	statusSetter := processor.statusSetter.(*assistantStatusSetterStub)

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-thinking-error", "show my work")))
	if !errors.Is(err, responseErr) {
		t.Fatalf("Process() error = %v, want %v", err, responseErr)
	}
	if len(statusSetter.calls) != 2 || statusSetter.calls[0].status != slackAssistantThinkingStatus || statusSetter.calls[1].status != "" {
		t.Fatalf("status calls = %+v, want thinking then explicit clear", statusSetter.calls)
	}
}

func TestEventProcessorDeduplicatesCompletedInboundEvent(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	store.processInbound = false
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-duplicate", "show my work")))
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

func TestEventProcessorBackfillsLegacyCredentials(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.legacyCredentials = []slackrepository.LegacySlackCredentialRecord{
		{SlackWorkspaceID: uuid.New(), Credential: "xoxb-legacy-one"},
		{SlackWorkspaceID: uuid.New(), Credential: "xoxb-legacy-two"},
	}
	processor := newTestEventProcessor(t, repo, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})

	upgraded, err := processor.BackfillLegacyCredentials(context.Background())
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if upgraded != 2 || repo.credentialUpgrades != 2 {
		t.Fatalf("credential upgrades = %d/%d, want 2/2", upgraded, repo.credentialUpgrades)
	}
	if len(repo.legacyCredentials) != 0 {
		t.Fatalf("legacy credentials remaining = %d, want 0", len(repo.legacyCredentials))
	}
	if !strings.HasPrefix(repo.installation.BotAccessToken, "v1.") {
		t.Fatalf("upgraded credential is not a versioned ciphertext: %q", repo.installation.BotAccessToken)
	}
}

func TestEventProcessorScrubsLegacyColumnsAfterVersionedCredentialRollout(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.versionedLegacyCredentials = 2
	processor := newTestEventProcessor(t, repo, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})

	updated, err := processor.BackfillLegacyCredentials(context.Background())
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if updated != 2 || repo.versionedLegacyCredentials != 0 {
		t.Fatalf("scrubbed/remaining = %d/%d, want 2/0", updated, repo.versionedLegacyCredentials)
	}
}

func TestEventProcessorRecoversInboxReceiptWithoutQueueingMessageContent(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	queue := &eventQueueStub{}
	processor.eventQueue = queue
	store.recoverableEvents = []messagingrepository.InboundEventRecord{
		{ExternalWorkspaceID: "T1", ExternalEventID: "Ev-recover", AttemptCount: 6, RecoveryGeneration: 7},
	}

	recovered, err := processor.RecoverPendingEvents(context.Background())
	if err != nil {
		t.Fatalf("RecoverPendingEvents() error = %v", err)
	}
	if recovered != 1 || len(queue.payloads) != 1 {
		t.Fatalf("recovered/payloads = %d/%d, want 1/1", recovered, len(queue.payloads))
	}
	if queue.payloads[0].ExternalWorkspaceID != "T1" || queue.payloads[0].EventID != "Ev-recover" {
		t.Fatalf("recovered payload = %+v, want event ID only", queue.payloads[0])
	}
	if queue.payloads[0].RecoveryAttempt != 7 {
		t.Fatalf("recovery attempt = %d, want 7", queue.payloads[0].RecoveryAttempt)
	}
}

func TestEventProcessorRecoveryContinuesPastQueueFailure(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	queueFailure := errors.New("queue unavailable")
	queue := &eventQueueStub{errors: []error{queueFailure}}
	processor.eventQueue = queue
	store.recoverableEvents = []messagingrepository.InboundEventRecord{
		{ID: uuid.New(), ExternalWorkspaceID: "T1", ExternalEventID: "Ev-first", RecoveryGeneration: 1},
		{ExternalWorkspaceID: "T1", ExternalEventID: "Ev-second", RecoveryGeneration: 1},
	}

	recovered, err := processor.RecoverPendingEvents(context.Background())

	if !errors.Is(err, queueFailure) {
		t.Fatalf("RecoverPendingEvents() error = %v, want queue failure", err)
	}
	if recovered != 1 || len(queue.payloads) != 2 || queue.payloads[1].EventID != "Ev-second" {
		t.Fatalf("recovered/payloads = %d/%+v, want second event enqueued", recovered, queue.payloads)
	}
	if len(store.releasedRecoveries) != 1 || store.releasedRecoveries[0].ID != store.recoverableEvents[0].ID || store.releasedRecoveries[0].RecoveryGeneration != 1 {
		t.Fatalf("released recoveries = %+v, want failed first claim released", store.releasedRecoveries)
	}
}

func TestEventProcessorRetriesTransientUninstallWithoutTargetingReconnect(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.installationErr = sql.ErrNoRows
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.eventQueue = &eventQueueStub{}
	processor.clientID = "client-id"
	processor.clientSecret = "client-secret"
	credentialPayload, credentialVersion, err := processor.codec.seal(slackCredential{AccessToken: "xoxb-disconnected"})
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
	store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{
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
	processor.eventQueue = &eventQueueStub{}

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
	store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{
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
	processor.eventQueue = &eventQueueStub{}

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
	store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{
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
	processor.eventQueue = &eventQueueStub{}

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
	store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{{
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
	processor.eventQueue = &eventQueueStub{}

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
	store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{{
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
	processor.eventQueue = &eventQueueStub{}

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
			store.recoverableOutbound = []messagingrepository.OutboundDeliveryRecord{
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
			processor.eventQueue = &eventQueueStub{}

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

func TestEventProcessorLoadsAndDecryptsCanonicalInboxPayload(t *testing.T) {
	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-canonical","event":{"type":"unsupported"}}`)
	encrypted, err := processor.codec.sealPayload(body)
	if err != nil {
		t.Fatalf("sealPayload() error = %v", err)
	}
	store.inboundRecords["Ev-canonical"] = messagingrepository.InboundEventRecord{
		ExternalWorkspaceID: "T1",
		ExternalEventID:     "Ev-canonical",
		Status:              "pending",
		PayloadEncrypted:    &encrypted,
	}

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
	storyReader := &eventStoryReaderStub{story: stories.CoreSingleStory{
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
		SecretKey:                testSlackCredentialSecret,
		CallLimiter:              &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}},
		UsageBudget:              &usageBudgetStub{},
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
		StoryReader:              storyReader,
	})
	require.NoError(t, err)

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
	encrypted, err := processor.codec.sealPayload(rawBody)
	require.NoError(t, err)
	store.inboundRecords[eventID] = messagingrepository.InboundEventRecord{
		ID:                  testInboundReceiptID,
		InstallGeneration:   uuidPointer(testInstallGeneration),
		ExternalWorkspaceID: "T1",
		ExternalEventID:     eventID,
		Status:              "pending",
		PayloadEncrypted:    &encrypted,
	}

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
	requestReader := &eventRequestReaderStub{request: integrationrequests.CoreIntegrationRequest{
		ID:              requestID,
		WorkspaceID:     testWorkspaceID,
		TeamID:          testAllowedTeamID,
		Title:           "Investigate mobile login",
		Status:          integrationrequests.StatusPending,
		Priority:        "High",
		CreatedByUserID: uuidPointer(testLinkedUserID),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}}
	processor, err := NewEventProcessor(nil, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, EventProcessorConfig{
		WebsiteURL:               "https://fortyone.app",
		SecretKey:                testSlackCredentialSecret,
		CallLimiter:              &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}},
		UsageBudget:              &usageBudgetStub{},
		ContextProvider:          &assistantContextProviderStub{},
		DailyWorkspaceTokenLimit: 1_000_000,
		RequestReader:            requestReader,
	})
	require.NoError(t, err)

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
	require.NoError(t, processor.Process(context.Background(), []byte(body)))

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
	store.inboundRecords["Ev-poison"] = messagingrepository.InboundEventRecord{
		ExternalWorkspaceID: "T1",
		ExternalEventID:     "Ev-poison",
		Status:              "pending",
		PayloadEncrypted:    &invalid,
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

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-busy", "show my work")))

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

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-delivered", "show my work")))

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

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-delivery-busy", "show my work")))

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

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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

	err := processor.Process(context.Background(), []byte(mentionEvent("Ev-link", "<@B1> show my work")))
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
	if !strings.Contains(message.Text, "https://app.fortyone.com/acme/settings/integrations/slack?slack_link_token=") {
		t.Fatalf("account-link message = %q", message.Text)
	}
	if strings.Contains(message.Text, testSlackBotAccessToken) {
		t.Fatal("account-link message leaked the bot token")
	}
	if len(assistant.requests) != 0 || len(access.workspaceIDs) != 0 {
		t.Fatalf("unlinked request reached access/assistant: access=%d assistant=%d", len(access.workspaceIDs), len(assistant.requests))
	}
}

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
			store.history = []messagingrepository.MessageRecord{
				{ExternalMessageID: &oldExternalID, Role: "user", Content: "Earlier question"},
				{Role: "assistant", Content: "Earlier answer"},
				// The repository returns the current message because the processor
				// persists it before loading history. It must not also be sent as a
				// conversation turn alongside Request.Prompt.
				{ExternalMessageID: &currentExternalID, Role: "user", Content: test.wantPrompt},
			}
			assistant := &assistantStub{response: messaging.Response{
				Text: "Two stories are due today.",
				Usage: messaging.Usage{
					InputTokens: 120, OutputTokens: 30, TotalTokens: 150,
				},
			}}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{externalMessageID: "10.2"}
			limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
			contextProvider := processor.contextProvider.(*assistantContextProviderStub)
			contextProvider.runtime = &messaging.RuntimeContext{
				Actor:     messaging.RuntimeActorContext{DisplayName: "Joseph Mukorivo", Username: "joseph"},
				LocalTime: time.Date(2026, time.August, 9, 7, 36, 0, 0, time.FixedZone("Africa/Harare", 2*60*60)),
			}

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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
			if test.name == "direct message" && contextCall.surface.Kind != messaging.RuntimeSurfaceDirect {
				t.Fatalf("direct message surface = %q", contextCall.surface.Kind)
			}
			if test.name != "direct message" && contextCall.surface.Kind != messaging.RuntimeSurfaceThread {
				t.Fatalf("channel surface = %q", contextCall.surface.Kind)
			}
			if test.name == "direct message" {
				if len(request.AllowedTeamIDs) != 1 || request.AllowedTeamIDs[0] != testAllowedTeamID || len(request.SharedTeamIDs) != 1 || request.SharedTeamIDs[0] != testAllowedTeamID || conversation.AudienceScope != messagingrepository.ConversationAudienceActor {
					t.Fatalf("direct assistant scope = allowed %v / shared %v / audience %q", request.AllowedTeamIDs, request.SharedTeamIDs, conversation.AudienceScope)
				}
			} else if len(request.AllowedTeamIDs) != 1 || request.AllowedTeamIDs[0] != testAllowedTeamID || len(request.SharedTeamIDs) != 1 || request.SharedTeamIDs[0] != testAllowedTeamID || conversation.AudienceScope != messagingrepository.ConversationAudienceChannel {
				t.Fatalf("channel assistant scope = allowed %v / shared %v / audience %q", request.AllowedTeamIDs, request.SharedTeamIDs, conversation.AudienceScope)
			}
			if len(request.Conversation) != test.wantConversationTurn || request.Conversation[0].Role != messaging.RoleUser || request.Conversation[1].Role != messaging.RoleAssistant {
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
	assistant := &assistantStub{response: messaging.Response{Text: "Your work is ready."}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processor.Process(context.Background(), []byte(mentionEvent("Ev-unmapped-personal-scope", "<@B1> show my work")))

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
	assistant := &assistantStub{response: messaging.Response{Text: "Team-wide answer"}}
	assistant.onRespond = func() {
		repo.sharedTeamIDs = []uuid.UUID{}
	}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processor.Process(context.Background(), []byte(mentionEvent("Ev-shared-scope-narrowed", "<@B1> what did everyone finish?")))

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
	assistant := &assistantStub{response: messaging.Response{Text: "Private team answer"}}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)

	err := processor.Process(context.Background(), []byte(mentionEvent("Ev-settings-race", "<@B1> show private work")))
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
	assistant := &assistantStub{response: messaging.Response{Text: "must not be used"}}
	sender := &messageSenderStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	contextProvider := processor.contextProvider.(*assistantContextProviderStub)
	contextProvider.err = errors.New("context database unavailable")

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-context-failed", "What time is it?")))

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
			store.conversation = messagingrepository.ConversationRecord{
				ID:        testConversationID,
				UpdatedAt: now,
			}
			previousQuestionID := "10.1"
			currentQuestionID := "10.2"
			store.history = []messagingrepository.MessageRecord{
				{ExternalMessageID: &previousQuestionID, Role: "user", Content: "What work is pending?"},
				{Role: "assistant", Content: "You have four open stories."},
				{ExternalMessageID: &currentQuestionID, Role: "user", Content: "Only show urgent items"},
			}
			assistant := &assistantStub{response: messaging.Response{Text: "ENG-3 is urgent."}}
			access := &accessCheckerStub{allowed: true}
			sender := &messageSenderStub{externalMessageID: "10.3"}
			processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
			processor.clock = fixedClock{now: now}

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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
	store.conversation = messagingrepository.ConversationRecord{
		ID:        testConversationID,
		UpdatedAt: now,
	}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	processor := newTestEventProcessor(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: false}, sender)
	processor.clock = fixedClock{now: now}

	if err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-channel-access", "U1", "show my work"))); err != nil {
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
	store.conversation = messagingrepository.ConversationRecord{
		ID:        testConversationID,
		UpdatedAt: now,
	}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{LimitedScope: "user"}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, &assistantStub{}, &accessCheckerStub{allowed: true}, sender, limiter, &usageBudgetStub{})
	processor.clock = fixedClock{now: now}

	if err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-channel-rate", "U1", "show my work"))); err != nil {
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
		conversation    messagingrepository.ConversationRecord
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
			conversation: messagingrepository.ConversationRecord{
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
			conversation: messagingrepository.ConversationRecord{
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
			conversation: messagingrepository.ConversationRecord{
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
			limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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

func TestEventProcessorDeniesAssistantWithoutEntitlement(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	assistant := &assistantStub{}
	access := &accessCheckerStub{allowed: false}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)

	if err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-access", "show my work"))); err != nil {
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
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
	text := "<@B1> " + strings.Repeat("x", messaging.MaximumMessageBytes+1)

	if err := processor.Process(context.Background(), []byte(mentionEvent("Ev-oversized", text))); err != nil {
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
			limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{LimitedScope: test.scope}}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

			if err := processor.Process(context.Background(), []byte(test.raw)); err != nil {
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
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{checkErr: &messagingrepository.DailyTokenLimitError{
		WorkspaceID: testWorkspaceID,
		Used:        1_000_000,
		Limit:       1_000_000,
	}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	if err := processor.Process(context.Background(), []byte(mentionEvent("Ev-daily-limit", "<@B1> show my work"))); err != nil {
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
			limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}, err: test.limiterErr}
			usage := &usageBudgetStub{checkErr: test.usageErr}
			processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

			err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-budget-infra", "show my work")))

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
	assistant := &assistantStub{err: errors.Join(errors.New("missing API key"), messaging.ErrAssistantNotConfigured)}
	access := &accessCheckerStub{allowed: true}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	processor := newTestEventProcessor(t, repo, store, assistant, access, sender)
	var logs bytes.Buffer
	processor.log = logger.NewWithJSON(&logs, slog.LevelError, "test")

	if err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-unavailable", "show my work"))); err != nil {
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
	require.Contains(t, logs.String(), "missing API key")
	require.Contains(t, logs.String(), `"slack_event_id":"Ev-unavailable"`)
}

func TestEventProcessorClassifiesOpenAIErrorsAfterRecordingPartialUsage(t *testing.T) {
	tests := []struct {
		name      string
		apiError  *messaging.APIError
		permanent bool
	}{
		{
			name:      "deterministic bad request becomes durable safe reply",
			apiError:  &messaging.APIError{StatusCode: http.StatusBadRequest, Code: "invalid_request_error", Message: "invalid model input", RequestID: "req_test_permanent"},
			permanent: true,
		},
		{
			name:     "ordinary rate limit remains retryable",
			apiError: &messaging.APIError{StatusCode: http.StatusTooManyRequests, Code: "rate_limit_exceeded", Message: "slow down"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newEventRepositoryStub()
			repo.linkedUserID = uuidPointer(testLinkedUserID)
			store := newEventStoreStub()
			partialUsage := messaging.Usage{InputTokens: 80, OutputTokens: 20, TotalTokens: 100}
			assistant := &assistantStub{response: messaging.Response{Usage: partialUsage}, err: test.apiError}
			sender := &messageSenderStub{externalMessageID: "10.2"}
			usage := &usageBudgetStub{}
			processor := newTestEventProcessorWithBudgets(
				t,
				repo,
				store,
				assistant,
				&accessCheckerStub{allowed: true},
				sender,
				&callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}},
				usage,
			)
			var logs bytes.Buffer
			processor.log = logger.NewWithJSON(&logs, slog.LevelError, "test")

			err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-openai-error", "show my work")))
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
	assistant := &assistantStub{response: messaging.Response{
		Text:  "This response must not be sent.",
		Usage: messaging.Usage{InputTokens: 40, OutputTokens: 10, TotalTokens: 50},
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
		&callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}},
		usage,
	)

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-usage-write", "show my work")))

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

func TestEventProcessorPersistedReplyBypassesNewBudgetDenials(t *testing.T) {
	repo := newEventRepositoryStub()
	repo.linkedUserID = uuidPointer(testLinkedUserID)
	store := newEventStoreStub()
	persistedReply := "The already-accounted answer."
	store.deliveryContent = &persistedReply
	assistant := &assistantStub{}
	sender := &messageSenderStub{externalMessageID: "10.2"}
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{LimitedScope: "workspace"}}
	usage := &usageBudgetStub{checkErr: &messagingrepository.DailyTokenLimitError{
		WorkspaceID: testWorkspaceID,
		Used:        1_000_000,
		Limit:       1_000_000,
	}}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-persisted-budget", "show my work")))

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
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	if err := processor.Process(context.Background(), []byte(mentionEvent("Ev-persisted-private-budget", "<@B1> show my work"))); err != nil {
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
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender, limiter, usage)

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-delivered-budget", "show my work")))

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

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-expired-answer", "show my work")))

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

	err := processor.Process(context.Background(), []byte(directMessageEvent("Ev-expired-access", "show my work")))

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
	assistant := &assistantStub{response: messaging.Response{Text: "A persisted answer."}}
	access := &accessCheckerStub{allowed: true}
	rateLimit := &RateLimitError{Method: "chat.postMessage", RetryAfter: 7 * time.Second}
	sender := &messageSenderStub{
		errors:            []error{rateLimit, nil},
		externalMessageID: "10.2",
	}
	limiter := &callLimiterStub{decision: messagingbudget.AdmissionDecision{Allowed: true}}
	usage := &usageBudgetStub{}
	processor := newTestEventProcessorWithBudgets(t, repo, store, assistant, access, sender, limiter, usage)
	raw := []byte(directMessageEvent("Ev-rate-limit", "What is due?"))

	err := processor.Process(context.Background(), raw)
	if !errors.Is(err, rateLimit) {
		t.Fatalf("first Process() error = %v, want %v", err, rateLimit)
	}
	if retryAfter, ok := SlackRetryAfter(err); !ok || retryAfter != 7*time.Second {
		t.Fatalf("SlackRetryAfter() = %s, %v; want 7s, true", retryAfter, ok)
	}
	if len(store.failedDeliveries) != 1 || !strings.Contains(store.failedDeliveries[0].message, "rate limited") {
		t.Fatalf("failed deliveries = %+v", store.failedDeliveries)
	}
	if len(store.completions) != 1 || store.completions[0].status != "failed" || !strings.Contains(store.completions[0].message, "rate limited") {
		t.Fatalf("first inbound completion = %+v", store.completions)
	}

	if err := processor.Process(context.Background(), raw); err != nil {
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

type threadSyncStub struct {
	input     integrationrequests.CoreInboundProviderCommentInput
	bindInput integrationrequests.CoreBindProviderThreadInput
}

func (s *threadSyncStub) IngestInboundProviderComment(_ context.Context, input integrationrequests.CoreInboundProviderCommentInput) (bool, error) {
	s.input = input
	return true, nil
}

func (s *threadSyncStub) BindProviderThread(_ context.Context, input integrationrequests.CoreBindProviderThreadInput) (integrationrequests.CoreProviderThread, error) {
	s.bindInput = input
	return integrationrequests.CoreProviderThread{ID: uuid.New()}, nil
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

	err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-unlinked-request-reply", "U-external", "Customer confirmed")))

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
	store.conversation = messagingrepository.ConversationRecord{ID: testConversationID, UpdatedAt: now}
	assistant := &assistantStub{response: messaging.Response{Text: "The customer confirmation is now part of this thread."}}
	sender := &messageSenderStub{externalMessageID: "10.3"}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, sender)
	processor.clock = fixedClock{now: now}
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-composed-request-reply", "U1", "Customer confirmed")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "completed")
	require.Equal(t, &testLinkedUserID, syncer.input.AuthorUserID)
	require.Len(t, store.conversationLookups, 1)
	lookup := store.conversationLookups[0]
	require.Equal(t, testLinkedUserID, lookup.UserID)
	require.Equal(t, "C1", lookup.ExternalChannelID)
	require.Equal(t, "10.1", lookup.ExternalThreadID)
	require.Equal(t, messagingrepository.ConversationAudienceChannel, lookup.AudienceScope)
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

	err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-request-comment-only", "U1", "Customer confirmed")))

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
	store.conversation = messagingrepository.ConversationRecord{ID: testConversationID, UpdatedAt: now}
	assistant := &assistantStub{}
	processor := newTestEventProcessor(t, repo, store, assistant, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.clock = fixedClock{now: now}
	syncer := &threadSyncStub{}
	processor.threadSync = syncer

	err := processor.Process(context.Background(), []byte(channelThreadEvent("Ev-request-mention-copy", "U1", "<@B1> summarize this")))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "ignored")
	require.Equal(t, "10.2", syncer.input.ExternalMessageID)
	require.Empty(t, store.conversationLookups)
	require.Empty(t, assistant.requests)
	require.Empty(t, store.outboundInputs)
}
