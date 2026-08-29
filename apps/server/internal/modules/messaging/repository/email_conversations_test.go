package messagingrepository

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEmailConversationStoreContractIsImplemented(t *testing.T) {
	t.Parallel()

	var _ messaging.EmailConversationStore = (*Repository)(nil)
}

func TestValidateEmailThreadInputRequiresOpaqueTokenHash(t *testing.T) {
	t.Parallel()

	input := validEmailThreadInput()
	input.ReplyTokenHash = []byte("raw-reply-token")

	require.ErrorIs(t, validateEmailThreadInput(input), messaging.ErrInvalidEmailReplyToken)
}

func TestNormalizeEmailThreadInputPreservesGroundingContext(t *testing.T) {
	t.Parallel()

	input := validEmailThreadInput()
	input.Provider = " brevo_email "
	input.RecipientEmail = " Joseph@Example.COM "
	input.Context = json.RawMessage(`{"objective":{"id":"abc","name":"Improve onboarding"}}`)

	normalized := normalizeEmailThreadInput(input)

	require.Equal(t, "brevo_email", normalized.Provider)
	require.Equal(t, "joseph@example.com", normalized.RecipientEmail)
	require.JSONEq(t, string(input.Context), string(normalized.Context))
	require.NoError(t, validateEmailThreadInput(normalized))
}

func TestValidateEmailMessageInputRequiresObjectMetadataAndKnownRoles(t *testing.T) {
	t.Parallel()

	input := validEmailMessageInput()
	input.ProviderMetadata = json.RawMessage(`[]`)
	require.ErrorIs(t, validateEmailMessageInput(input), messaging.ErrInvalidEmailConversation)

	input = validEmailMessageInput()
	input.Role = "tool"
	require.ErrorIs(t, validateEmailMessageInput(input), messaging.ErrInvalidEmailConversation)
}

func TestEmailMessageIdempotencyMatchIncludesContentAndRFCThreading(t *testing.T) {
	t.Parallel()

	input := validEmailMessageInput()
	providerID := input.ProviderMessageID
	internetID := input.InternetMessageID
	inReplyTo := input.InReplyToMessageID
	record := messaging.EmailMessageRecord{
		ThreadID:           input.ThreadID,
		IdempotencyKey:     input.IdempotencyKey,
		Direction:          input.Direction,
		Role:               input.Role,
		Kind:               input.Kind,
		ProviderMessageID:  &providerID,
		InternetMessageID:  &internetID,
		InReplyToMessageID: &inReplyTo,
		Subject:            input.Subject,
		Content:            input.Content,
		Context:            input.Context,
		ProviderMetadata:   input.ProviderMetadata,
	}

	require.True(t, emailMessageMatchesInput(record, input))
	input.Content = "Different reply"
	require.False(t, emailMessageMatchesInput(record, input))
}

func TestEmailMessageIdempotencyMatchIncludesInboundEvent(t *testing.T) {
	t.Parallel()

	input := validEmailMessageInput()
	inboundEventID := uuid.New()
	input.InboundEventID = &inboundEventID
	record := messaging.EmailMessageRecord{
		ThreadID:         input.ThreadID,
		InboundEventID:   input.InboundEventID,
		IdempotencyKey:   input.IdempotencyKey,
		Direction:        input.Direction,
		Role:             input.Role,
		Kind:             input.Kind,
		Subject:          input.Subject,
		Content:          input.Content,
		Context:          input.Context,
		ProviderMetadata: input.ProviderMetadata,
	}
	providerID := input.ProviderMessageID
	internetID := input.InternetMessageID
	inReplyTo := input.InReplyToMessageID
	record.ProviderMessageID = &providerID
	record.InternetMessageID = &internetID
	record.InReplyToMessageID = &inReplyTo

	require.True(t, emailMessageMatchesInput(record, input))
	differentEventID := uuid.New()
	input.InboundEventID = &differentEventID
	require.False(t, emailMessageMatchesInput(record, input))
}

func TestValidateEmailProposalLifecycleInputs(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	input := messaging.EmailActionProposalInput{
		ThreadID:              uuid.New(),
		WorkspaceID:           uuid.New(),
		UserID:                uuid.New(),
		SourceMessageID:       uuid.New(),
		IdempotencyKey:        "proposal:event-1",
		ActionKind:            "update_key_result",
		EntityType:            "key_result",
		EntityID:              uuid.New(),
		ExpectedEntityVersion: "2026-08-12T08:00:00Z",
		ProposedDiff:          json.RawMessage(`{"currentValue":{"from":32,"to":38}}`),
		ExpiresAt:             now.Add(24 * time.Hour),
		Now:                   now,
	}
	require.NoError(t, validateEmailActionProposalInput(input))

	input.ExpiresAt = now
	require.ErrorIs(t, validateEmailActionProposalInput(input), messaging.ErrInvalidEmailProposal)
}

func TestValidateEmailProposalDecisionAcceptsOnlyConsentOutcomes(t *testing.T) {
	t.Parallel()

	decision := messaging.EmailActionProposalDecision{
		ProposalID:     uuid.New(),
		ThreadID:       uuid.New(),
		WorkspaceID:    uuid.New(),
		UserID:         uuid.New(),
		ReplyTokenHash: bytes.Repeat([]byte{1}, sha256DigestSize),
		Decision:       messaging.EmailActionProposalConfirmed,
		Now:            time.Now().UTC(),
	}
	require.NoError(t, validateEmailActionProposalDecision(decision))

	decision.Decision = messaging.EmailActionProposalApplied
	require.ErrorIs(t, validateEmailActionProposalDecision(decision), messaging.ErrInvalidEmailProposal)
}

func TestEmailProposalConfirmationReplayIncludesApplyLifecycle(t *testing.T) {
	t.Parallel()

	for _, status := range []messaging.EmailActionProposalStatus{
		messaging.EmailActionProposalConfirmed,
		messaging.EmailActionProposalApplying,
		messaging.EmailActionProposalApplied,
		messaging.EmailActionProposalFailed,
	} {
		require.True(t, emailProposalDecisionAlreadyWon(status, messaging.EmailActionProposalConfirmed))
	}
	require.False(t, emailProposalDecisionAlreadyWon(
		messaging.EmailActionProposalCancelled,
		messaging.EmailActionProposalConfirmed,
	))
	require.False(t, emailProposalDecisionAlreadyWon(
		messaging.EmailActionProposalConfirmed,
		messaging.EmailActionProposalCancelled,
	))
}

func TestEmailConversationQueriesRetainGuestAccessAndCurrentMembershipGuards(t *testing.T) {
	t.Parallel()

	source := readEmailQuerySource(t, "email_conversations.sql")
	require.Contains(t, source, "INNER JOIN workspace_members AS member")
	require.Contains(t, source, "workspace.deleted_at IS NULL")
	require.Contains(t, source, "actor.is_active = true")
	require.Contains(t, source, "actor.is_system = false")
	require.Equal(t, 2, strings.Count(source, "member.role IN ('admin', 'member', 'guest')"))

	proposalSource := readEmailQuerySource(t, "email_action_proposals.sql")
	require.Contains(t, proposalSource, "member.role IN ('admin', 'member', 'guest')")
}

func TestEmailMigrationProtectsIdentityHistoryAndProposalLifecycle(t *testing.T) {
	t.Parallel()

	up := readEmailMigration(t, "000118_messaging_email_conversations.up.sql")
	require.Contains(t, up, "CREATE TABLE public.messaging_email_threads")
	require.Contains(t, up, "CREATE TABLE public.messaging_email_reply_tokens")
	require.Contains(t, up, "CREATE TABLE public.messaging_email_messages")
	require.Contains(t, up, "CREATE TABLE public.messaging_email_action_proposals")
	require.Contains(t, up, "UNIQUE (thread_id, sequence)")
	require.Contains(t, up, "UNIQUE (thread_id, id)")
	require.Contains(t, up, "WHERE status = 'pending'")
	require.Contains(t, up, "status IN ('pending', 'confirmed', 'applying', 'applied', 'failed', 'cancelled', 'expired', 'superseded')")
	require.Contains(t, up, "octet_length(token_hash) = 32")
	require.Contains(t, up, "root_internet_message_id")
	require.Contains(t, up, "latest_internet_message_id")
	require.Contains(t, up, "summary_through_sequence")
	require.Contains(t, up, "'onboarding',\n            'email_reply'")
	require.Contains(t, up, "expires_at > created_at")

	down := readEmailMigration(t, "000118_messaging_email_conversations.down.sql")
	require.Contains(t, down, "DROP TABLE IF EXISTS public.messaging_email_action_proposals")
	require.Contains(t, down, "DROP TABLE IF EXISTS public.messaging_email_threads")
	require.Contains(t, down, "WHERE purpose = 'email_reply'")
	require.Contains(t, down, "FROM public.messaging_email_threads")
	require.Contains(t, down, "FROM public.messaging_email_reply_tokens")
	require.Contains(t, down, "FROM public.messaging_email_messages")
	require.Contains(t, down, "FROM public.messaging_email_action_proposals")
	require.Contains(t, down, "'creation_confirmation',\n            'onboarding'")
	require.NotContains(t, down, "'onboarding',\n            'email_reply'")
}

func validEmailThreadInput() messaging.EmailThreadInput {
	return messaging.EmailThreadInput{
		Provider:            "brevo_email",
		WorkspaceID:         uuid.New(),
		UserID:              uuid.New(),
		RecipientEmail:      "joseph@example.com",
		ExternalThreadID:    uuid.NewString(),
		Context:             json.RawMessage(`{}`),
		ReplyTokenHash:      bytes.Repeat([]byte{1}, sha256DigestSize),
		ReplyTokenExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	}
}

func validEmailMessageInput() messaging.EmailMessageInput {
	return messaging.EmailMessageInput{
		ThreadID:           uuid.New(),
		WorkspaceID:        uuid.New(),
		UserID:             uuid.New(),
		IdempotencyKey:     "brevo:event-1",
		Direction:          messaging.EmailMessageDirectionInbound,
		Role:               messaging.EmailMessageRoleUser,
		Kind:               messaging.EmailMessageKindReply,
		ProviderMessageID:  "event-1",
		InternetMessageID:  "<reply-1@example.com>",
		InReplyToMessageID: "<maya-1@fortyone.app>",
		Subject:            "Re: Objective update",
		Content:            "Set progress to 38%.",
		Context:            json.RawMessage(`{"objective":"Improve onboarding"}`),
		ProviderMetadata:   json.RawMessage(`{"provider":"brevo"}`),
	}
}

func readEmailQuerySource(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("queries", name))
	require.NoError(t, err)
	return string(data)
}

func readEmailMigration(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", name))
	require.NoError(t, err)
	return string(data)
}
