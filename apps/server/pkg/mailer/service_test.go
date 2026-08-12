package mailer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSenderForProfileUsesConfiguredMayaIdentity(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress:     "notifications@fortyone.app",
		FromName:        "FortyOne",
		MayaFromAddress: "maya@fortyone.app",
		MayaFromName:    "Maya",
	}}

	address, name := mailerService.senderForProfile(SenderProfileMaya)

	require.Equal(t, "maya@fortyone.app", address)
	require.Equal(t, "Maya", name)
}

func TestSenderForProfileFallsBackToDefaultAddressForMaya(t *testing.T) {
	mailerService := &service{config: Config{
		FromAddress: "notifications@fortyone.app",
		FromName:    "FortyOne",
	}}

	address, name := mailerService.senderForProfile(SenderProfileMaya)

	require.Equal(t, "notifications@fortyone.app", address)
	require.Equal(t, "Maya", name)
}
