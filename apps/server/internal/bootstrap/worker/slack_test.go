package workerbootstrap

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildSlackEventProcessorRequiresStoryMutationSideEffects(t *testing.T) {
	tests := []struct {
		name         string
		dependencies slackEventProcessorDependencies
		errorMessage string
	}{
		{
			name: "event publisher",
			dependencies: slackEventProcessorDependencies{
				Tasks:       new(tasks.Service),
				MayaActorID: uuid.New(),
			},
			errorMessage: "event publisher is required",
		},
		{
			name: "tasks service",
			dependencies: slackEventProcessorDependencies{
				EventPublisher: new(publisher.Publisher),
				MayaActorID:    uuid.New(),
			},
			errorMessage: "tasks service is required",
		},
		{
			name: "Maya actor",
			dependencies: slackEventProcessorDependencies{
				EventPublisher: new(publisher.Publisher),
				Tasks:          new(tasks.Service),
			},
			errorMessage: "Maya actor ID is required",
		},
		{
			name: "Maya workspace access",
			dependencies: slackEventProcessorDependencies{
				EventPublisher:  new(publisher.Publisher),
				Tasks:           new(tasks.Service),
				MayaActorID:     uuid.New(),
				CredentialVault: new(credentialvault.Vault),
			},
			errorMessage: "Maya workspace access is required",
		},
		{
			name: "credential vault",
			dependencies: slackEventProcessorDependencies{
				EventPublisher: new(publisher.Publisher),
				Tasks:          new(tasks.Service),
				MayaActorID:    uuid.New(),
				MayaAccess:     &mayaWorkspaceAccessStub{allowed: true},
			},
			errorMessage: "credential vault is required",
		},
		{
			name: "webhook payload secret",
			dependencies: slackEventProcessorDependencies{
				EventPublisher:  new(publisher.Publisher),
				Tasks:           new(tasks.Service),
				MayaActorID:     uuid.New(),
				MayaAccess:      &mayaWorkspaceAccessStub{allowed: true},
				CredentialVault: new(credentialvault.Vault),
			},
			errorMessage: "webhook payload secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := buildSlackEventProcessor(nil, nil, nil, Config{}, tt.dependencies)
			require.Nil(t, processor)
			require.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestGitHubCompatibilityUsesMayaRepositoryCapability(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	access := &mayaWorkspaceAccessStub{allowed: true}
	compatibility := buildGitHubCompatibilityDependencies(nil, access)

	require.NoError(t, compatibility.validate())
	allowed, err := compatibility.autoSchedulingEligibility(context.Background(), workspaceID)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, []uuid.UUID{workspaceID}, access.workspaceIDs)
}

func TestGitHubCompatibilityRejectsMissingMayaRepositoryCapability(t *testing.T) {
	t.Parallel()

	compatibility := buildGitHubCompatibilityDependencies(nil, nil)

	require.EqualError(
		t,
		compatibility.validate(),
		"GitHub auto-scheduling eligibility checker is required",
	)
}

func TestSlackEventProcessorConfigIncludesUninstallCredentials(t *testing.T) {
	var cfg Config
	cfg.Website.URL = "https://app.fortyone.test"
	cfg.Auth.SecretKey = "encryption-key"
	cfg.Slack.ClientID = "worker-client-id"
	cfg.Slack.ClientSecret = "worker-client-secret"
	webhookPayloadSecret := "worker-slack-webhook-payload-secret"

	processorConfig := slackEventProcessorConfig(cfg, nil, webhookPayloadSecret)

	require.Equal(t, cfg.Website.URL, processorConfig.WebsiteURL)
	require.Equal(t, webhookPayloadSecret, processorConfig.WebhookPayloadSecret)
	require.Equal(t, cfg.Slack.ClientID, processorConfig.ClientID)
	require.Equal(t, cfg.Slack.ClientSecret, processorConfig.ClientSecret)
	require.Equal(t, cfg.MessagingAssistant.WorkspaceTokensPerDay, processorConfig.DailyWorkspaceTokenLimit)
}

func TestMessagingAssistantCallLimiterConfigUsesPerMinuteBudgets(t *testing.T) {
	var cfg Config
	cfg.MessagingAssistant.UserCallsPerMinute = 18
	cfg.MessagingAssistant.WorkspaceCallsPerMinute = 240

	limiterConfig := messagingAssistantCallLimiterConfig(cfg)

	require.Equal(t, int64(18), limiterConfig.UserLimit)
	require.Equal(t, int64(240), limiterConfig.WorkspaceLimit)
	require.Equal(t, time.Minute, limiterConfig.Window)
}
