package emailreply

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInboundEmailSupportsBrevoMailboxShapes(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"Uuid":["receipt-1"],
		"MessageId":"<reply@example.com>",
		"From":{"Name":"Joseph","Address":"joseph@example.com"},
		"To":[{"Name":"Maya","Address":"maya+abcdefghijklmnop@reply.fortyone.app"}],
		"Recipients":["maya+abcdefghijklmnop@reply.fortyone.app"],
		"Cc":[],
		"Subject":"Re: Weekly strategy check-in",
		"ExtractedMarkdownMessage":"We moved this objective to on track."
	}`)

	email, err := decodeInboundEmail(raw)
	require.NoError(t, err)
	require.Equal(t, "joseph@example.com", email.From.Address)
	require.Equal(t, "maya+abcdefghijklmnop@reply.fortyone.app", email.Recipients[0].Address)
	token, err := extractReplyToken(email)
	require.NoError(t, err)
	require.Equal(t, "abcdefghijklmnop", token)
	require.Equal(t, "<reply@example.com>", externalEventID(email, raw))
}

func TestExtractReplyTokenRejectsAmbiguousRecipients(t *testing.T) {
	t.Parallel()

	_, err := extractReplyToken(InboundEmail{
		To: Mailboxes{
			{Address: "maya+abcdefghijklmnop@reply.fortyone.app"},
			{Address: "maya+qrstuvwxyzabcdef@reply.fortyone.app"},
		},
	})

	require.ErrorIs(t, err, ErrInvalidReplyToken)
}

func TestExternalEventIDFallsBackToPayloadHash(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{"From":{"Address":"joseph@example.com"}}`)
	first := externalEventID(InboundEmail{}, raw)
	second := externalEventID(InboundEmail{}, raw)
	reformatted := externalEventID(InboundEmail{}, json.RawMessage("{\n  \"From\": { \"Address\": \"joseph@example.com\" }\n}"))

	require.Equal(t, first, second)
	require.Equal(t, first, reformatted)
	require.Regexp(t, `^sha256:[a-f0-9]{64}$`, first)
	require.NotEqual(t, first, externalEventID(InboundEmail{}, json.RawMessage(`{"From":{"Address":"other@example.com"}}`)))
}

func TestDecodeInboundWebhookRejectsEmptyAndOversizedBatches(t *testing.T) {
	t.Parallel()

	_, err := decodeInboundWebhook([]byte(`{"items":[]}`))
	require.ErrorIs(t, err, ErrInvalidPayload)

	items := make([]json.RawMessage, maxInboundItemsPerBatch+1)
	for index := range items {
		items[index] = json.RawMessage(`{}`)
	}
	raw, err := json.Marshal(InboundWebhook{Items: items})
	require.NoError(t, err)
	_, err = decodeInboundWebhook(raw)
	require.ErrorIs(t, err, ErrInvalidPayload)
}
