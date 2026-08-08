package workerbootstrap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
