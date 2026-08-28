package slack

import (
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/stretchr/testify/require"
)

func TestNewEventProcessorDoesNotRequireLegacyAuthSecretForNormalRuntime(t *testing.T) {
	t.Parallel()
	repository := newEventRepositoryStub()
	store := newEventStoreStub()
	processor, err := NewEventProcessor(
		nil,
		repository,
		store,
		&assistantStub{},
		&accessCheckerStub{allowed: true},
		EventProcessorConfig{
			WebhookPayloadSecret:     testSlackWebhookPayloadSecret,
			CredentialVault:          newTestCredentialVault(t),
			CallLimiter:              &callLimiterStub{decision: AssistantAdmissionDecision{Allowed: true}},
			UsageBudget:              &usageBudgetStub{},
			ContextProvider:          &assistantContextProviderStub{},
			DailyWorkspaceTokenLimit: 1_000_000,
			WebhookInbox:             store,
			WebhookRecovery:          &webhookRecoveryStub{},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, processor)
}

func mustTestLegacyCutover(t testing.TB) *LegacyCutover {
	t.Helper()
	cutover, err := NewLegacyCutover(testSlackCredentialSecret)
	if err != nil {
		t.Fatalf("NewLegacyCutover() error = %v", err)
	}
	return cutover
}

func sealTestEventInstallation(t testing.TB, processor *EventProcessor, repository *eventRepositoryStub) {
	t.Helper()
	if repository == nil || processor == nil || processor.codec == nil ||
		credentialvault.IsEnvelope(repository.installation.BotAccessToken) {
		return
	}
	plaintext := strings.TrimSpace(repository.installation.BotAccessToken)
	if plaintext == "" {
		return
	}
	encrypted, version, err := processor.codec.seal(slackCredentialBinding{
		WorkspaceID:       repository.installation.WorkspaceID,
		SlackTeamID:       repository.installation.SlackTeamID,
		InstallGeneration: repository.installation.InstallGeneration,
	}, slackCredential{AccessToken: plaintext})
	if err != nil {
		t.Fatalf("seal test Slack installation credential: %v", err)
	}
	repository.installation.BotAccessToken = encrypted
	repository.installation.CredentialVersion = version
}
