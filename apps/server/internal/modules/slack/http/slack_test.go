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

func TestRequestLogLimitIsBoundedAndUnambiguous(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		want    int
		wantErr bool
	}{
		{name: "default", want: 50},
		{name: "custom", query: "?limit=125", want: 125},
		{name: "zero", query: "?limit=0", wantErr: true},
		{name: "above maximum", query: "?limit=201", wantErr: true},
		{name: "malformed", query: "?limit=many", wantErr: true},
		{name: "repeated", query: "?limit=20&limit=30", wantErr: true},
		{name: "overflow", query: "?limit=999999999999999999999999", wantErr: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/request-logs"+test.query, nil)
			got, err := requestLogLimit(request)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
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
