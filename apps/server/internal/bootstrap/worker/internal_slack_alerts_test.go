package workerbootstrap

import (
	"os"
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/stretchr/testify/require"
)

func TestInternalSlackAlertsDisabledRegisterNoWork(t *testing.T) {
	scheduler := &scheduleCapture{}
	require.NoError(t, registerInternalSlackAlerts(nil, scheduler, slack.InternalAlertConfig{}, nil, nil, nil))
	require.Empty(t, scheduler.entries)
}

func TestInternalSlackConfigEnabledByDefaultRequiresDestinationAndAllowsOptOut(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("APP_INTERNAL_SLACK_ENABLED", "")
	require.NoError(t, os.Unsetenv("APP_INTERNAL_SLACK_ENABLED"))
	t.Setenv("APP_INTERNAL_SLACK_TEAM_ID", "")
	t.Setenv("APP_INTERNAL_SLACK_CHANNEL_ID", "C0000000001")
	_, err := loadConfig()
	require.Error(t, err)
	t.Setenv("APP_INTERNAL_SLACK_TEAM_ID", "T0000000001")
	cfg, err := loadConfig()
	require.NoError(t, err)
	require.True(t, cfg.InternalSlack.Enabled)
	require.Equal(t, "T0000000001", cfg.InternalSlack.TeamID)
	require.Equal(t, "C0000000001", cfg.InternalSlack.ChannelID)

	t.Setenv("APP_INTERNAL_SLACK_ENABLED", "false")
	t.Setenv("APP_INTERNAL_SLACK_TEAM_ID", "")
	t.Setenv("APP_INTERNAL_SLACK_CHANNEL_ID", "")
	cfg, err = loadConfig()
	require.NoError(t, err)
	require.False(t, cfg.InternalSlack.Enabled)
}
