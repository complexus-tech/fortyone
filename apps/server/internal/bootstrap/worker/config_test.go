package workerbootstrap

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadsSlackUninstallCredentials(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("SLACK_CLIENT_ID", "worker-client-id")
	t.Setenv("SLACK_CLIENT_SECRET", "worker-client-secret")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, "worker-client-id", cfg.Slack.ClientID)
	require.Equal(t, "worker-client-secret", cfg.Slack.ClientSecret)
}

func TestLoadConfigReadsMessagingAssistantBudgets(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("OPENAI_ASSISTANT_USER_CALLS_PER_MINUTE", "18")
	t.Setenv("OPENAI_ASSISTANT_WORKSPACE_CALLS_PER_MINUTE", "240")
	t.Setenv("OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY", "1500000")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, int64(18), cfg.MessagingAssistant.UserCallsPerMinute)
	require.Equal(t, int64(240), cfg.MessagingAssistant.WorkspaceCallsPerMinute)
	require.Equal(t, int64(1_500_000), cfg.MessagingAssistant.WorkspaceTokensPerDay)
}

func TestMessagingAssistantBudgetDefaults(t *testing.T) {
	configType := reflect.TypeOf(Config{}.MessagingAssistant)
	userLimit, found := configType.FieldByName("UserCallsPerMinute")
	require.True(t, found)
	workspaceLimit, found := configType.FieldByName("WorkspaceCallsPerMinute")
	require.True(t, found)
	dailyTokenLimit, found := configType.FieldByName("WorkspaceTokensPerDay")
	require.True(t, found)

	require.Equal(t, "12", userLimit.Tag.Get("default"))
	require.Equal(t, "120", workspaceLimit.Tag.Get("default"))
	require.Equal(t, "1000000", dailyTokenLimit.Tag.Get("default"))
}

func TestMessagingAssistantModelDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{})
	model, found := configType.FieldByName("AIModel")
	require.True(t, found)
	require.Equal(t, "gpt-5.6-luna", model.Tag.Get("default"))
}
