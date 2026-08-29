package slack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPostSlackCreationAckPersistsSlackTeamBinding(t *testing.T) {
	store := newEventStoreStub()
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	service := newTestService(&mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		IsActive:          true,
	}}, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://slack.com/api/chat.postMessage" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	messageTS := service.postSlackCreationAck(
		context.Background(),
		workspaceID,
		installGeneration,
		"slack:view:V1:confirmation",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-token",
		"Task created in FortyOne.",
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "171.200", messageTS)
	require.Equal(t, "T1", store.outboundInputs[0].ExternalWorkspaceID)
	require.Equal(t, installGeneration, *store.outboundInputs[0].InstallGeneration)
	require.Len(t, store.completedDeliveries, 1)
}

func TestPostSlackCreationAckRoutesToSourceConversation(t *testing.T) {
	tests := []struct {
		name         string
		source       requestSourceContext
		wantChannel  string
		wantThreadTS string
	}{
		{
			name: "source channel root",
			source: requestSourceContext{
				SlackTeamID:    "T1",
				SlackChannelID: "C1",
				SlackMessageTS: "171.100",
				SlackUserID:    "U1",
			},
			wantChannel: "C1",
		},
		{
			name: "existing source thread",
			source: requestSourceContext{
				SlackTeamID:    "T1",
				SlackChannelID: "C1",
				SlackMessageTS: "171.200",
				SlackThreadTS:  "171.100",
				SlackUserID:    "U1",
			},
			wantChannel:  "C1",
			wantThreadTS: "171.100",
		},
		{
			name: "user fallback without source channel",
			source: requestSourceContext{
				SlackTeamID:   "T1",
				SlackThreadTS: "171.100",
				SlackUserID:   "U1",
			},
			wantChannel: "U1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newEventStoreStub()
			workspaceID := uuid.New()
			installGeneration := uuid.New()
			service := newTestService(&mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
				WorkspaceID:       workspaceID,
				SlackTeamID:       "T1",
				InstallGeneration: installGeneration,
				IsActive:          true,
			}}, &mockRequestStore{}, &mockStoryService{}, Config{})
			service.outbound = store
			service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.300"}`)),
					Header:     make(http.Header),
				}, nil
			})}

			service.postSlackCreationAck(
				context.Background(),
				workspaceID,
				installGeneration,
				"slack:view:V1:confirmation:"+test.name,
				test.source,
				"xoxb-token",
				"Joseph Mukorivo created WEB-123",
			)

			require.Len(t, store.outboundInputs, 1)
			require.Equal(t, test.wantChannel, store.outboundInputs[0].ExternalChannelID)
			require.Equal(t, test.wantThreadTS, store.outboundInputs[0].ExternalThreadID)
		})
	}
}

func TestBuildSlackStoryCreatedText(t *testing.T) {
	storyCode := buildStoryCode(" web ", 123)
	require.Equal(t, "WEB-123", storyCode)
	require.Equal(
		t,
		"Joseph Mukorivo created <https://acme.fortyone.app/work/WEB-123|WEB-123>",
		buildSlackStoryCreatedText("Joseph Mukorivo", storyCode, "https://acme.fortyone.app/work/WEB-123"),
	)
	require.Equal(
		t,
		"Joseph &lt;@U123&gt; &amp; Team created WEB-123",
		buildSlackStoryCreatedText("Joseph <@U123> & Team", storyCode, ""),
	)
	fallback := buildSlackStoryCreatedText(" ", "", "https://acme.fortyone.app/work/story-id")
	require.Equal(t, "A team member created <https://acme.fortyone.app/work/story-id|a story>", fallback)
	require.NotContains(t, fallback, "✅")
	require.NotContains(t, fallback, "in FortyOne")
}

func TestBuildSlackRequestLifecycleText(t *testing.T) {
	require.Equal(
		t,
		"Joseph Mukorivo <https://acme.fortyone.app/teams/6ba7b812-9dad-11d1-80b4-00c04fd430c8/requests/6ba7b813-9dad-11d1-80b4-00c04fd430c8|opened a request>",
		buildSlackRequestOpenedText("Joseph Mukorivo", "https://acme.fortyone.app/teams/6ba7b812-9dad-11d1-80b4-00c04fd430c8/requests/6ba7b813-9dad-11d1-80b4-00c04fd430c8"),
	)
	require.Equal(
		t,
		"Joseph &lt;@U123&gt; &amp; Team opened a request",
		buildSlackRequestOpenedText("Joseph <@U123> & Team", ""),
	)
	require.Equal(
		t,
		"Joseph Mukorivo linked a request to <https://acme.fortyone.app/work/WEB-123|WEB-123>",
		buildSlackStoryLinkedRequestText("Joseph Mukorivo", "WEB-123", "https://acme.fortyone.app/work/WEB-123"),
	)
}
