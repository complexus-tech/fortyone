package slackdomain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseUninstallKindRejectsUnknownValues(t *testing.T) {
	t.Parallel()

	for _, value := range []UninstallKind{
		UninstallDisconnect,
		UninstallWorkspaceDelete,
		UninstallOrphanedOAuth,
	} {
		parsed, err := ParseUninstallKind("  " + string(value) + "  ")
		require.NoError(t, err)
		require.Equal(t, value, parsed)
	}

	_, err := ParseUninstallKind("force_delete")
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestParseUninstallStatusRejectsUnknownValues(t *testing.T) {
	t.Parallel()

	for _, value := range []UninstallStatus{
		UninstallPending,
		UninstallProcessing,
		UninstallCompleted,
		UninstallFailed,
		UninstallRevocationRequired,
	} {
		parsed, err := ParseUninstallStatus(string(value))
		require.NoError(t, err)
		require.Equal(t, value, parsed)
	}

	_, err := ParseUninstallStatus("unknown")
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestEnqueueUninstallValidateRequiresCompleteIdentityAndCredential(t *testing.T) {
	t.Parallel()

	valid := EnqueueUninstall{
		SlackWorkspaceID:     uuid.New(),
		WorkspaceID:          uuid.New(),
		InstallGeneration:    uuid.New(),
		SlackTeamID:          "T123",
		UninstallKind:        UninstallDisconnect,
		CredentialPayload:    "vault.v2.payload",
		CredentialKeyVersion: 2,
	}
	require.NoError(t, valid.Validate())

	tests := []struct {
		name   string
		mutate func(*EnqueueUninstall)
	}{
		{name: "installation id", mutate: func(input *EnqueueUninstall) { input.SlackWorkspaceID = uuid.Nil }},
		{name: "workspace id", mutate: func(input *EnqueueUninstall) { input.WorkspaceID = uuid.Nil }},
		{name: "generation", mutate: func(input *EnqueueUninstall) { input.InstallGeneration = uuid.Nil }},
		{name: "team id", mutate: func(input *EnqueueUninstall) { input.SlackTeamID = " " }},
		{name: "kind", mutate: func(input *EnqueueUninstall) { input.UninstallKind = "force_delete" }},
		{name: "credential", mutate: func(input *EnqueueUninstall) { input.CredentialPayload = "" }},
		{name: "credential version", mutate: func(input *EnqueueUninstall) { input.CredentialKeyVersion = 0 }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := valid
			test.mutate(&input)
			require.ErrorIs(t, input.Validate(), ErrInvalidInput)
		})
	}
}
