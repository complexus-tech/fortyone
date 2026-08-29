package appkeys

import (
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/stretchr/testify/require"
)

func TestIntegrationKeysAreStableAndPurposeSeparated(t *testing.T) {
	t.Parallel()

	first, err := NewIntegrationKeys("test-application-root-secret-with-at-least-32-bytes")
	require.NoError(t, err)
	second, err := NewIntegrationKeys("test-application-root-secret-with-at-least-32-bytes")
	require.NoError(t, err)

	require.Equal(t, first.GitHubWebhookPayloadSecret, second.GitHubWebhookPayloadSecret)
	require.Equal(t, first.SlackWebhookPayloadSecret, second.SlackWebhookPayloadSecret)
	require.Equal(t, first.FigmaWebhookPayloadSecret, second.FigmaWebhookPayloadSecret)
	require.NotEqual(t, first.GitHubWebhookPayloadSecret, first.SlackWebhookPayloadSecret)
	require.NotEqual(t, first.GitHubWebhookPayloadSecret, first.FigmaWebhookPayloadSecret)
	require.NotEqual(t, first.SlackWebhookPayloadSecret, first.FigmaWebhookPayloadSecret)

	active, err := first.CredentialVault.ActiveKeyRef()
	require.NoError(t, err)
	require.Equal(t, derivedKeyID, active.ID)
	require.Equal(t, derivedVersion, active.Version)

	binding := credentialvault.Context{
		Provider:       "slack",
		TenantID:       "workspace-1",
		SubjectID:      "team-1",
		CredentialType: "bot-oauth",
		Generation:     "generation-1",
	}
	envelope, err := first.CredentialVault.Seal(binding, []byte("provider-token"))
	require.NoError(t, err)
	opened, err := second.CredentialVault.Open(binding, envelope)
	require.NoError(t, err)
	defer opened.Destroy()
	require.Equal(t, []byte("provider-token"), opened.Reveal())
}

func TestIntegrationKeysChangeWithRootSecret(t *testing.T) {
	t.Parallel()

	first, err := NewIntegrationKeys("first-test-application-root-secret-with-32-bytes")
	require.NoError(t, err)
	second, err := NewIntegrationKeys("second-test-application-root-secret-with-32-bytes")
	require.NoError(t, err)

	require.NotEqual(t, first.GitHubWebhookPayloadSecret, second.GitHubWebhookPayloadSecret)
	require.NotEqual(t, first.SlackWebhookPayloadSecret, second.SlackWebhookPayloadSecret)
	require.NotEqual(t, first.FigmaWebhookPayloadSecret, second.FigmaWebhookPayloadSecret)

	binding := credentialvault.Context{
		Provider:       "github",
		TenantID:       "workspace-1",
		SubjectID:      "user-1",
		CredentialType: "user-oauth-access-token",
		Generation:     "generation-1",
	}
	envelope, err := first.CredentialVault.Seal(binding, []byte("provider-token"))
	require.NoError(t, err)
	_, err = second.CredentialVault.Open(binding, envelope)
	require.Error(t, err)
	require.True(t, errors.Is(err, credentialvault.ErrAuthentication))
}

func TestIntegrationKeysRequireRootSecret(t *testing.T) {
	t.Parallel()

	_, err := NewIntegrationKeys("  ")
	require.ErrorContains(t, err, "application root secret is required")
}
