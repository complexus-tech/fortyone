package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleSlackStoryCompactEditActionOpensFocusedModal(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	var opened map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/views.open", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&opened))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	fixture.service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions",
		"trigger_id":"trigger-compact",
		"team":{"id":"T123","domain":"acme"},
		"user":{"id":"U123","username":"joseph"},
		"channel":{"id":"C123","name":"general"},
		"container":{"message_ts":"1754700000.123","channel_id":"C123","app_unfurl_url":"https://acme.fortyone.app/work/web-123"},
		"app_unfurl":{"app_unfurl_url":"https://acme.fortyone.app/work/web-123"},
		"actions":[{"action_id":"fortyone_edit_story_priority","value":"WEB-123"}]
	}`), &payload))

	response, err := fixture.service.handleSlackStoryCompactEditAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)
	view := opened["view"].(map[string]any)
	require.Equal(t, slackWorkObjectCompactEditCallbackID, view["callback_id"])
	blocks := view["blocks"].([]any)
	element := blocks[0].(map[string]any)["element"].(map[string]any)
	require.Equal(t, "static_select", element["type"])
	require.Equal(t, "High", element["initial_option"].(map[string]any)["value"])
}

func TestHandleSlackStoryCompactEditSubmissionUpdatesPriorityAndRefreshesUnfurl(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	var unfurl SlackChatUnfurlRequest
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.unfurl", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&unfurl))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	fixture.service.webClient.baseURL = provider.URL

	metadata, err := json.Marshal(slackWorkObjectCompactEditMetadata{
		Source: requestSourceContext{
			SlackTeamID:    "T123",
			SlackChannelID: "C123",
			SlackMessageTS: "1754700000.123",
			SlackUserID:    "U123",
		},
		StoryURL:       "https://acme.fortyone.app/work/web-123",
		StoryReference: "WEB-123",
		Field:          slackWorkObjectCompactPriorityField,
	})
	require.NoError(t, err)
	payload := fixture.payload
	payload.Type = "view_submission"
	payload.View.Type = "modal"
	payload.View.CallbackID = slackWorkObjectCompactEditCallbackID
	payload.View.PrivateMetadata = string(metadata)
	payload.View.State.Values = interactionViewStateValues{
		"priority": workObjectSelectState("priority", "Urgent"),
	}

	response, err := fixture.service.handleSlackStoryCompactEditSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)
	require.Equal(t, 1, fixture.stories.updateCalls)
	require.Equal(t, "Urgent", fixture.stories.updates["priority"])
	require.Equal(t, "C123", unfurl.Channel)
	require.Equal(t, "1754700000.123", unfurl.TS)
	require.Equal(t, "Urgent", unfurl.Metadata.Entities[0].EntityPayload.Fields["priority"].Value)
}

func TestHandleSlackStoryCompactEditSubmissionUpdatesStatusAndRefreshesUnfurl(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	newStatus := fixture.repo.statuses[1]
	var unfurl SlackChatUnfurlRequest
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.unfurl", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&unfurl))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	fixture.service.webClient.baseURL = provider.URL

	metadata, err := json.Marshal(slackWorkObjectCompactEditMetadata{
		Source: requestSourceContext{
			SlackTeamID:    "T123",
			SlackChannelID: "C123",
			SlackMessageTS: "1754700000.123",
			SlackUserID:    "U123",
		},
		StoryURL:       "https://acme.fortyone.app/work/web-123",
		StoryReference: "WEB-123",
		Field:          slackWorkObjectCompactStatusField,
	})
	require.NoError(t, err)
	payload := fixture.payload
	payload.Type = "view_submission"
	payload.View.Type = "modal"
	payload.View.CallbackID = slackWorkObjectCompactEditCallbackID
	payload.View.PrivateMetadata = string(metadata)
	payload.View.State.Values = interactionViewStateValues{
		"status": workObjectSelectState("status", newStatus.ID.String()),
	}

	response, err := fixture.service.handleSlackStoryCompactEditSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, 200, response.StatusCode)
	require.Equal(t, 1, fixture.stories.updateCalls)
	require.Equal(t, newStatus.ID, fixture.stories.updates["status_id"])
	require.Equal(t, "C123", unfurl.Channel)
	require.Equal(t, "1754700000.123", unfurl.TS)
	require.Equal(t, newStatus.Name, unfurl.Metadata.Entities[0].EntityPayload.Fields["status"].Value)
}
