package slackhttp

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
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
