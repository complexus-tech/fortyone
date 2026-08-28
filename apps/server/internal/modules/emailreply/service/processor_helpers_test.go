package emailreply

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type summarizerStub struct {
	requests []SummaryRequest
	err      error
}

func (stub *summarizerStub) Summarize(_ context.Context, request SummaryRequest) (SummaryGeneration, error) {
	stub.requests = append(stub.requests, request)
	if stub.err != nil {
		return SummaryGeneration{}, stub.err
	}
	return SummaryGeneration{Summary: "durable summary"}, nil
}

type summaryStoreStub struct {
	ProcessorStore
	updates []ThreadSummaryUpdate
}

func (stub *summaryStoreStub) UpdateThreadSummary(
	_ context.Context,
	input ThreadSummaryUpdate,
) (Thread, error) {
	stub.updates = append(stub.updates, input)
	return Thread{
		ID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		Summary: input.Summary, SummaryThroughSequence: input.ThroughSequence,
		NextMessageSequence: input.ThroughSequence + 20,
	}, nil
}

func TestPlainInboundReplyUsesCurrentExtractedBodyAndRemovesFormatting(t *testing.T) {
	t.Parallel()

	raw := "This is fallback\n\nOn Monday, Maya wrote:\nOld message"
	reply := plainInboundReply(InboundEmail{
		ExtractedMarkdownMessage: "## Update\n\n- Set **activation** to [On Track](https://example.com)\n\n> old quoted message",
		RawTextBody:              &raw,
	})

	require.Equal(t, "Update\n\nSet activation to On Track", reply)
	require.NotContains(t, reply, "https://")
	require.NotContains(t, reply, "old quoted")
}

func TestRefreshSummaryStartsAtFirstNewlyOmittedTurn(t *testing.T) {
	t.Parallel()

	thread := Thread{
		ID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New(),
		Summary: "earlier summary", SummaryThroughSequence: 10, NextMessageSequence: 25,
	}
	messages := make([]Message, 0, 14)
	for sequence := int64(11); sequence <= 24; sequence++ {
		role := MessageRoleUser
		if sequence%2 == 0 {
			role = MessageRoleAssistant
		}
		messages = append(messages, Message{
			ID: uuid.New(), Sequence: sequence, Role: role, Content: "turn",
		})
	}
	store := &summaryStoreStub{}
	summarizer := &summarizerStub{}
	processor := &Processor{store: store, summarizer: summarizer}

	updated, err := processor.refreshSummary(context.Background(), thread, messages)

	require.NoError(t, err)
	require.Len(t, summarizer.requests, 1)
	require.Len(t, summarizer.requests[0].OmittedTurns, 2)
	require.Equal(t, int64(12), updated.SummaryThroughSequence)
	require.Equal(t, int64(12), store.updates[0].ThroughSequence)
}

func TestControlReplySkipsUnavailableSummarizer(t *testing.T) {
	t.Parallel()

	thread := Thread{
		ID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New(),
		SummaryThroughSequence: 1, NextMessageSequence: 20,
	}
	messages := make([]Message, 0, 18)
	for sequence := int64(2); sequence < 20; sequence++ {
		messages = append(messages, Message{
			ID: uuid.New(), Sequence: sequence, Role: MessageRoleUser, Content: "turn",
		})
	}
	summarizer := &summarizerStub{err: errors.New("provider unavailable")}
	processor := &Processor{store: &summaryStoreStub{}, summarizer: summarizer}

	updated, err := processor.refreshSummaryForReply(context.Background(), " CONFIRM ", thread, messages)

	require.NoError(t, err)
	require.Equal(t, thread, updated)
	require.Empty(t, summarizer.requests)
}

func TestConversationHistoryExcludesSummarizedAndCurrentInbound(t *testing.T) {
	t.Parallel()

	currentID := uuid.New()
	now := time.Now()
	history := conversationHistory([]Message{
		{ID: uuid.New(), Sequence: 4, Role: MessageRoleAssistant, Content: "summarized", CreatedAt: now},
		{ID: uuid.New(), Sequence: 5, Role: MessageRoleAssistant, Content: "latest Maya", CreatedAt: now},
		{ID: currentID, Sequence: 6, Role: MessageRoleUser, Content: "current reply", CreatedAt: now},
	}, 4, currentID)

	require.Equal(t, []HistoryTurn{{Role: ConversationRoleAssistant, Text: "latest Maya", SentAt: now}}, history)
}

func TestDeterministicReplyThreadHeadersUseCurrentInboundMessage(t *testing.T) {
	t.Parallel()

	thread := Thread{
		RootInternetMessageID:   "<root@fortyone.app>",
		LatestInternetMessageID: "<current-user@example.com>",
	}
	require.Equal(t,
		[]string{"<root@fortyone.app>", "<current-user@example.com>"},
		emailReplyReferences(thread),
	)
}

func TestBoundedProviderMessageIDHashesOversizedValues(t *testing.T) {
	t.Parallel()

	raw := strings.Repeat("m", 1_200)
	first := boundedProviderMessageID(raw, "fallback")
	second := boundedProviderMessageID(raw, "different-fallback")

	require.Equal(t, first, second)
	require.LessOrEqual(t, len(first), 998)
	require.True(t, strings.HasPrefix(first, "sha256:"))
}

func TestConversationReplyCopyPreservesInboundThreadSubject(t *testing.T) {
	copy := EmailCopy{Subject: "A newly generated subject"}

	threaded := conversationReplyCopy(copy, "3 overdue tasks need a decision")

	require.Equal(t, "Re: 3 overdue tasks need a decision", threaded.Subject)
	require.Equal(t, "A newly generated subject", copy.Subject)
}

type processorStateStub struct{}

func (processorStateStub) OpenStoredInboundEmail(string) (StoredInboundEmail, error) {
	return StoredInboundEmail{}, errors.New("not used")
}

func (processorStateStub) SealProcessorState(payload []byte) (string, error) {
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func (processorStateStub) OpenProcessorState(sealed string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(sealed)
}

type deliveryStoreStub struct {
	ProcessorStore
	payload   []byte
	completed bool
}

func (stub *deliveryStoreStub) SetOutboundDeliveryContentAndProviderPayload(_ context.Context, _ uuid.UUID, _ string, payload []byte) error {
	stub.payload = bytes.Clone(payload)
	return nil
}

func (stub *deliveryStoreStub) CompleteOutboundDelivery(context.Context, uuid.UUID, string) error {
	stub.completed = true
	return nil
}

func (*deliveryStoreStub) FailOutboundDelivery(context.Context, uuid.UUID, string) error { return nil }

type replyThreadStub struct {
	newTokenCalls int
	input         ReplyPreparation
}

func (stub *replyThreadStub) NewReplyToken(context.Context, Thread) (string, error) {
	stub.newTokenCalls++
	return "abcdefghijklmnop", nil
}

func (stub *replyThreadStub) PrepareReply(_ context.Context, input ReplyPreparation) (string, error) {
	stub.input = input
	return "maya+abcdefghijklmnop@reply.fortyone.app", nil
}

type copyRendererStub struct{}

func (copyRendererStub) RenderHTML(EmailCopy) (string, error) {
	return "<p>Done.</p>", nil
}

type mailerStub struct {
	emails []mailer.Email
}

func (stub *mailerStub) Send(_ context.Context, email mailer.Email) error {
	stub.emails = append(stub.emails, email)
	return nil
}

func (*mailerStub) SendTemplated(context.Context, mailer.TemplatedEmail) error {
	return errors.New("unexpected templated send")
}

func TestDeliverReplyUsesCurrentInboundThreadHeadersAndFrozenToken(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	store := &deliveryStoreStub{}
	threads := &replyThreadStub{}
	mail := &mailerStub{}
	processor := &Processor{
		store: store, inbound: processorStateStub{}, renderer: copyRendererStub{}, threads: threads, mailer: mail,
		now: func() time.Time { return now },
	}
	thread := Thread{
		ID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New(), RecipientEmail: "joseph@example.com",
		RootInternetMessageID: "<root@fortyone.app>", LatestInternetMessageID: "<current-inbound@example.com>",
	}
	expiresAt := now.Add(time.Hour)
	delivery := OutboundDelivery{
		ID: uuid.New(), IdempotencyKey: "email-reply:event-1", ExpiresAt: &expiresAt,
	}
	teamID := uuid.New()

	err := processor.deliverReply(context.Background(), delivery, thread,
		deterministicReply("Re: Weekly check-in", "Done."), []uuid.UUID{teamID})

	require.NoError(t, err)
	require.True(t, store.completed)
	require.Equal(t, 1, threads.newTokenCalls)
	require.Equal(t, "<current-inbound@example.com>", threads.input.InReplyTo)
	require.Len(t, mail.emails, 1)
	require.Equal(t, "<current-inbound@example.com>", mail.emails[0].InReplyTo)
	require.Equal(t, []string{"<root@fortyone.app>", "<current-inbound@example.com>"}, mail.emails[0].References)
	require.Contains(t, mail.emails[0].Body, "FortyOne")
	require.NotContains(t, mail.emails[0].Body, "```")

	var envelope emailDeliveryEnvelope
	require.NoError(t, json.Unmarshal(store.payload, &envelope))
	require.NotEmpty(t, envelope.Sealed)
	opened, err := processor.decodeEmailDeliveryPayload(store.payload)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{teamID}, opened.AuthorizedTeamIDs)
}

func TestFrozenDeliveryAuthorizationRejectsLostTeam(t *testing.T) {
	t.Parallel()

	codec := processorStateStub{}
	requiredTeamID := uuid.New()
	payload, err := json.Marshal(emailDeliveryPayload{
		To: []string{"joseph@example.com"}, Subject: "Update", HTML: "<p>Update</p>", PlainText: "Update",
		ReplyToken: "abcdefghijklmnop", MessageID: "<message@fortyone.app>", Kind: MessageKindAnswer,
		HistoryIdempotencyKey: "outbound:event", AuthorizationVersion: 1, AuthorizedTeamIDs: []uuid.UUID{requiredTeamID},
	})
	require.NoError(t, err)
	sealed, err := codec.SealProcessorState(payload)
	require.NoError(t, err)
	envelope, err := json.Marshal(emailDeliveryEnvelope{Sealed: sealed})
	require.NoError(t, err)
	processor := &Processor{inbound: codec}

	require.ErrorIs(t, processor.authorizeFrozenDelivery(envelope, AuthorizedContext{AllowedTeamIDs: []uuid.UUID{uuid.New()}}), ErrActionUnauthorized)
	require.NoError(t, processor.authorizeFrozenDelivery(envelope, AuthorizedContext{AllowedTeamIDs: []uuid.UUID{requiredTeamID}}))
}
