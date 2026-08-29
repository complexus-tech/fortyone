package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSlackWorkObjectPublisherUnfurlsTypedMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/chat.unfurl", req.URL.Path)
		require.Equal(t, "Bearer xoxb-test", req.Header.Get("Authorization"))
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"entity_type":"slack#/entities/task"`)
		require.NotContains(t, string(body), "xoxb-test")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	require.NoError(t, publisher.Unfurl(context.Background(), "xoxb-test", request))
}

func TestSlackWorkObjectPublisherUsesComposerDestination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "unfurl-123", payload["unfurl_id"])
		require.Equal(t, "composer", payload["source"])
		require.NotContains(t, payload, "channel")
		require.NotContains(t, payload, "ts")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryUnfurlRequest("COMPOSER", "draft-ts", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	request.Channel = ""
	request.TS = ""
	request.UnfurlID = "unfurl-123"
	request.Source = "composer"
	require.NoError(t, publisher.Unfurl(context.Background(), "xoxb-test", request))
}

func TestPublishSlackStoryUnfurlLogsProviderErrorCode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"error":"cannot_unfurl_url"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	var logs bytes.Buffer
	processor := &EventProcessor{
		log:         logger.NewWithJSON(&logs, slog.LevelInfo, "test"),
		workObjects: newSlackWorkObjectPublisher(client),
	}
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://complexus.fortyone.app/work/WEB-544",
		Title:         "Private story title that must not be logged",
	})
	require.NoError(t, err)

	err = processor.publishSlackStoryUnfurl(context.Background(), normalizedSlackEvent{
		EventID:   "Ev-preview",
		TeamID:    "T123",
		UserID:    "U123",
		ChannelID: "C123",
		MessageTS: "1754700000.123",
	}, uuid.MustParse("11111111-1111-4111-8111-111111111111"), request, 1, false, "xoxb-test")

	require.Error(t, err)
	require.Contains(t, logs.String(), `"msg":"Slack story preview publish failed"`)
	require.Contains(t, logs.String(), `"slack_error_code":"cannot_unfurl_url"`)
	require.Contains(t, logs.String(), `"event_id":"Ev-preview"`)
	require.NotContains(t, logs.String(), "Private story title that must not be logged")
}

func TestApplySlackUnfurlEventDestinationUsesComposerIdentity(t *testing.T) {
	t.Parallel()

	request := SlackChatUnfurlRequest{Channel: "COMPOSER", TS: "draft-ts"}
	applySlackUnfurlEventDestination(&request, normalizedSlackEvent{
		Source:   "composer",
		UnfurlID: "unfurl-123",
	})

	require.Empty(t, request.Channel)
	require.Empty(t, request.TS)
	require.Equal(t, "unfurl-123", request.UnfurlID)
	require.Equal(t, "composer", request.Source)
	require.NoError(t, validateSlackUnfurlRequestDestination(request))
}

func TestValidateSlackUnfurlRequestDestinationRejectsMixedPairs(t *testing.T) {
	t.Parallel()

	err := validateSlackUnfurlRequestDestination(SlackChatUnfurlRequest{
		Channel:  "C123",
		TS:       "1754700000.123",
		UnfurlID: "unfurl-123",
		Source:   "composer",
	})
	require.ErrorContains(t, err, "exactly one destination pair")
}

func TestSlackWorkObjectPublisherPresentsEntityDetailsWithoutAppUnfurlURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/entity.presentDetails", req.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "trigger-123", payload["trigger_id"])
		metadata, ok := payload["metadata"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, slackTaskEntityType, metadata["entity_type"])
		require.NotContains(t, metadata, "app_unfurl_url")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	request, err := BuildSlackStoryEntityDetailsRequest("trigger-123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	require.NoError(t, publisher.PresentDetails(context.Background(), "xoxb-test", request))
}

func TestSlackWorkObjectPublisherPostsRichCreationReceipt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/chat.postMessage", req.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(req.Body).Decode(&payload))
		require.Equal(t, "Joseph created <https://acme.fortyone.app/work/WEB-123|WEB-123>", payload["text"])
		require.Equal(t, false, payload["unfurl_links"])
		require.NotNil(t, payload["metadata"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"ts":"1754700000.456"}`))
	}))
	defer server.Close()

	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	publisher := newSlackWorkObjectPublisher(client)
	receipt, err := BuildSlackStoryCreationReceipt("Joseph", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Fix workspace login",
	})
	require.NoError(t, err)
	externalMessageID, err := publisher.PostCreationReceipt(context.Background(), "xoxb-test", "C123", "", "client-id", receipt)
	require.NoError(t, err)
	require.Equal(t, "1754700000.456", externalMessageID)
}

func TestSlackWorkObjectPublisherRejectsMixedAuthAndMetadata(t *testing.T) {
	t.Parallel()

	publisher := newSlackWorkObjectPublisher(newSlackWebClient(http.DefaultClient))
	request, err := BuildSlackStoryUnfurlRequest("C123", "1754700000.123", SlackStoryWorkObjectInput{
		AccessGranted: true,
		StoryURL:      "https://acme.fortyone.app/work/WEB-123",
		Title:         "Private title",
	})
	require.NoError(t, err)
	request.UserAuthRequired = true
	request.UserAuthURL = "https://acme.fortyone.app/settings/integrations/slack"
	err = publisher.Unfurl(context.Background(), "xoxb-test", request)
	require.ErrorContains(t, err, "cannot mix")
}

func TestSlackBotOAuthScopeValueIncludesChannelInventoryAndLinkScopes(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		"app_mentions:read,channels:history,channels:read,chat:write,chat:write.public,commands,groups:history,groups:read,im:history,links:read,links:write,users:read,users:read.email",
		slackBotOAuthScopeValue(),
	)
}
