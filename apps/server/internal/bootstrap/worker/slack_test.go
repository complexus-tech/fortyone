package workerbootstrap

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
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
				Tasks: new(tasks.Service),
			},
			errorMessage: "event publisher is required",
		},
		{
			name: "tasks service",
			dependencies: slackEventProcessorDependencies{
				EventPublisher: new(publisher.Publisher),
			},
			errorMessage: "tasks service is required",
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

func TestSlackEventProcessorConfigIncludesUninstallCredentials(t *testing.T) {
	var cfg Config
	cfg.Website.URL = "https://app.fortyone.test"
	cfg.Auth.SecretKey = "encryption-key"
	cfg.Slack.ClientID = "worker-client-id"
	cfg.Slack.ClientSecret = "worker-client-secret"

	processorConfig := slackEventProcessorConfig(cfg, nil)

	require.Equal(t, cfg.Website.URL, processorConfig.WebsiteURL)
	require.Equal(t, cfg.Auth.SecretKey, processorConfig.SecretKey)
	require.Equal(t, cfg.Slack.ClientID, processorConfig.ClientID)
	require.Equal(t, cfg.Slack.ClientSecret, processorConfig.ClientSecret)
	require.Equal(t, cfg.MessagingAssistant.WorkspaceTokensPerDay, processorConfig.DailyWorkspaceTokenLimit)
	require.Nil(t, processorConfig.EventQueue)
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
