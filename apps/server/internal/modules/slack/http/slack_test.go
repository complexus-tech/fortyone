package slackhttp

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInvalidSlackRequestDoesNotConsumeDiagnosticWorker(t *testing.T) {
	service := slack.New(nil, nil, nil, nil, slack.Config{SigningSecret: "signing-secret"})
	handler := New(nil, service)
	request := httptest.NewRequest("POST", "/integrations/slack/events", strings.NewReader(`{"type":"event_callback"}`))
	request.Header.Set("X-Slack-Request-Timestamp", "1")
	request.Header.Set("X-Slack-Signature", "v0=invalid")
	recorder := httptest.NewRecorder()

	err := handler.HandleEvents(context.Background(), recorder, request)

	require.NoError(t, err)
	require.Equal(t, 0, len(handler.requestLogSlots))
}

func TestOversizedSlackRequestDoesNotConsumeDiagnosticWorker(t *testing.T) {
	service := slack.New(nil, nil, nil, nil, slack.Config{SigningSecret: "signing-secret"})
	handler := New(nil, service)
	request := httptest.NewRequest(
		"POST",
		"/integrations/slack/events",
		io.LimitReader(strings.NewReader(strings.Repeat("x", maxSlackRequestBodyBytes+1)), maxSlackRequestBodyBytes+1),
	)
	recorder := httptest.NewRecorder()

	err := handler.HandleEvents(context.Background(), recorder, request)

	require.NoError(t, err)
	require.Equal(t, 0, len(handler.requestLogSlots))
}

func TestAssistantConfiguredOrDefaultPreservesPreviousPutSemantics(t *testing.T) {
	t.Parallel()

	configured := true
	unconfigured := false

	require.True(t, assistantConfiguredOrDefault(nil))
	require.True(t, assistantConfiguredOrDefault(&configured))
	require.False(t, assistantConfiguredOrDefault(&unconfigured))
}

func TestToAppChannelAudiencesIncludesExplicitConfigurationState(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	output := toAppChannelAudiences([]slack.CoreSlackChannelAudience{
		{
			Channel:      slack.CoreSlackChannel{SlackChannelID: "C1", Name: "general"},
			IsConfigured: true,
			TeamIDs:      []uuid.UUID{teamID},
		},
	})

	require.Len(t, output, 1)
	require.Equal(t, "C1", output[0].Channel.SlackChannelID)
	require.Equal(t, "general", output[0].Channel.Name)
	require.True(t, output[0].IsConfigured)
	require.Equal(t, []string{teamID.String()}, output[0].TeamIDs)
}
