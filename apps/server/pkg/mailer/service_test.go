package mailer

import (
	"bytes"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSenderForProfileUsesConfiguredMayaIdentity(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress:     "notifications@fortyone.app",
		FromName:        "FortyOne",
		MayaFromAddress: "maya@updates.fortyone.app",
		MayaFromName:    "Maya from FortyOne",
	}}

	address, name := mailerService.senderForProfile(SenderProfileMaya)

	require.Equal(t, "maya@updates.fortyone.app", address)
	require.Equal(t, "Maya from FortyOne", name)
}

func TestBuildMessageCreatesMultipartThreadedEmail(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress: "notifications@fortyone.app",
		FromName:    "FortyOne",
	}}

	message, err := mailerService.buildMessage(Email{
		To:            []string{"joseph@example.com"},
		Subject:       "A strategy update",
		Body:          "<p>Your objective is back on track.</p>",
		PlainTextBody: "Your objective is back on track.",
		IsHTML:        true,
		Sender:        SenderProfileMaya,
		ReplyTo:       "maya+opaque-token@reply.fortyone.app",
		MessageID:     "<maya-outbound-2@fortyone.app>",
		InReplyTo:     "<recipient-reply-1@example.com>",
		References: []string{
			"<maya-outbound-1@fortyone.app>",
			"<recipient-reply-1@example.com>",
		},
	})
	require.NoError(t, err)

	var raw bytes.Buffer
	_, err = message.WriteTo(&raw)
	require.NoError(t, err)
	email := raw.String()
	require.Contains(t, email, "Reply-To: maya+opaque-token@reply.fortyone.app")
	require.Contains(t, email, "Message-ID: <maya-outbound-2@fortyone.app>")
	require.Contains(t, email, "In-Reply-To: <recipient-reply-1@example.com>")
	require.Contains(t, email, "References: <maya-outbound-1@fortyone.app> <recipient-reply-1@example.com>")
	require.Contains(t, email, "multipart/alternative")
	require.Contains(t, email, "Content-Type: text/plain")
	require.Contains(t, email, "Content-Type: text/html")
}

func TestBuildMessageFormatsCommaDisplayNameAsValidAddress(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress:     "notifications@fortyone.app",
		FromName:        "FortyOne",
		MayaFromAddress: "maya@fortyone.app",
		MayaFromName:    "Maya, AI Agent",
	}}

	message, err := mailerService.buildMessage(Email{
		To:      []string{"joseph@example.com"},
		Subject: "A strategy update",
		Body:    "Your objective is back on track.",
		Sender:  SenderProfileMaya,
	})
	require.NoError(t, err)

	fromHeaders := message.GetHeader("From")
	require.Len(t, fromHeaders, 1)

	from, err := mail.ParseAddress(fromHeaders[0])
	require.NoError(t, err)
	require.Equal(t, "maya@fortyone.app", from.Address)
	require.Equal(t, "Maya, AI Agent", from.Name)
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress: "notifications@fortyone.app",
		FromName:    "FortyOne",
	}}

	_, err := mailerService.buildMessage(Email{
		To:        []string{"joseph@example.com"},
		Subject:   "A strategy update",
		Body:      "Safe body",
		ReplyTo:   "maya@fortyone.app\r\nBcc: attacker@example.com",
		MessageID: "<maya-outbound-2@fortyone.app>",
	})
	require.ErrorContains(t, err, "Reply-To header contains a line break")
	require.False(t, strings.Contains(err.Error(), "attacker@example.com"))
}

func TestSenderForProfileUsesBuiltInMayaAddress(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress: "notifications@fortyone.app",
		FromName:    "FortyOne",
	}}

	address, name := mailerService.senderForProfile(SenderProfileMaya)

	require.Equal(t, "maya@fortyone.app", address)
	require.Equal(t, "Maya, AI Agent", name)
}
