package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandleCommandRespondsEvenWhenOpeningModalFails(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{
			"T123:U123": actorID,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	failureResponse := make(chan CommandResponse, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == "https://slack.com/api/views.open" {
			return nil, errors.New("slack api unavailable")
		}
		if req.URL.String() != "https://hooks.slack.com/actions/T123/response" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		var response CommandResponse
		if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
			return nil, err
		}
		failureResponse <- response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("team_domain", "acme")
	form.Set("channel_id", "C123")
	form.Set("channel_name", "general")
	form.Set("user_id", "U123")
	form.Set("user_name", "joseph")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/response")

	resp, err := service.HandleCommand(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Empty(t, resp.ResponseType)
	require.Empty(t, resp.Text)
	select {
	case failure := <-failureResponse:
		require.Equal(t, "ephemeral", failure.ResponseType)
		require.Contains(t, failure.Text, "Unable to open")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous slash-command failure feedback")
	}
}

func TestHandleCommandAcknowledgesBeforeWorkAndSurvivesRequestCancellation(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	baseRepo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	repo := &blockingSlackWorkspaceRepo{
		mockRepo: baseRepo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	modalOpened := make(chan error, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://slack.com/api/views.open":
			modalOpened <- req.Context().Err()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V123"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			modalOpened <- fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("channel_id", "C123")
	form.Set("user_id", "U123")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/response")

	type commandResult struct {
		response CommandResponse
		err      error
	}
	result := make(chan commandResult, 1)
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	go func() {
		response, err := service.HandleCommand(requestCtx, []byte(form.Encode()))
		result <- commandResult{response: response, err: err}
	}()

	select {
	case command := <-result:
		require.NoError(t, command.err)
		require.Empty(t, command.response.ResponseType)
		require.Empty(t, command.response.Text)
	case <-time.After(250 * time.Millisecond):
		close(repo.release)
		t.Fatal("slash command was not acknowledged before workspace lookup completed")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for asynchronous slash-command work")
	}
	cancelRequest()
	close(repo.release)

	select {
	case openErr := <-modalOpened:
		require.NoError(t, openErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for modal open after request cancellation")
	}
}

func TestHandleCommandPromptsAccountLinkWhenSlackUserIsUnmapped(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		WebsiteURL: "https://fortyone.app",
	})
	type connectResult struct {
		response CommandResponse
		err      error
	}
	connectResponse := make(chan connectResult, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var response CommandResponse
		var resultErr error
		if req.URL.String() != "https://hooks.slack.com/actions/T123/connect" {
			resultErr = fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		} else {
			resultErr = json.NewDecoder(req.Body).Decode(&response)
		}
		connectResponse <- connectResult{response: response, err: resultErr}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("team_domain", "acme")
	form.Set("channel_id", "C123")
	form.Set("channel_name", "general")
	form.Set("user_id", "U456")
	form.Set("user_name", "joseph")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship it")
	form.Set("response_url", "https://hooks.slack.com/actions/T123/connect")

	resp, err := service.HandleCommand(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Empty(t, resp.ResponseType)
	require.Empty(t, resp.Text)
	select {
	case connect := <-connectResponse:
		require.NoError(t, connect.err)
		require.Equal(t, "ephemeral", connect.response.ResponseType)
		require.Contains(t, connect.response.Text, "Connect FortyOne account")
		require.Contains(t, connect.response.Text, "slack_link_token=")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack account-link prompt")
	}
}
