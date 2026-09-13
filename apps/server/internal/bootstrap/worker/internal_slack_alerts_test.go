package workerbootstrap

import (
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/stretchr/testify/require"
)

func TestInternalSlackAlertsDisabledRegisterNoWork(t *testing.T) {
	scheduler := &scheduleCapture{}
	require.NoError(t, registerInternalSlackAlerts(nil, scheduler, slack.InternalAlertConfig{}, nil, nil, nil))
	require.Empty(t, scheduler.entries)
}

func TestInternalSlackConfigRequiresExplicitDestination(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("APP_INTERNAL_SLACK_ENABLED", "true")
	t.Setenv("APP_INTERNAL_SLACK_TEAM_ID", "")
	t.Setenv("APP_INTERNAL_SLACK_CHANNEL_ID", "C014XSVSRF1")
	_, err := loadConfig()
	require.Error(t, err)
	t.Setenv("APP_INTERNAL_SLACK_TEAM_ID", "T015B85FC6R")
	cfg, err := loadConfig()
	require.NoError(t, err)
	require.True(t, cfg.InternalSlack.Enabled)
	require.Equal(t, "T015B85FC6R", cfg.InternalSlack.TeamID)
	require.Equal(t, "C014XSVSRF1", cfg.InternalSlack.ChannelID)
}
