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

func TestHandleMessageActionAcknowledgesBeforeWorkAndSurvivesRequestCancellation(t *testing.T) {
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

	interaction := map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]any{"id": "T123", "domain": "acme"},
		"channel":      map[string]any{"id": "C123", "name": "general"},
		"user":         map[string]any{"id": "U123", "username": "joseph"},
		"message":      map[string]any{"user": "U456", "text": "Ship the Slack integration", "ts": "171234.000100"},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	type interactionResult struct {
		response InteractionResponse
		err      error
	}
	result := make(chan interactionResult, 1)
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	go func() {
		response, err := service.HandleInteractivity(requestCtx, []byte(form.Encode()))
		result <- interactionResult{response: response, err: err}
	}()
	select {
	case interactionResponse := <-result:
		require.NoError(t, interactionResponse.err)
		require.Equal(t, http.StatusOK, interactionResponse.response.StatusCode)
	case <-time.After(250 * time.Millisecond):
		close(repo.release)
		t.Fatal("message action was not acknowledged before workspace lookup completed")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for asynchronous message action")
	}
	cancelRequest()
	close(repo.release)

	select {
	case openErr := <-modalOpened:
		require.NoError(t, openErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message-action modal after request cancellation")
	}
}

func TestOpenCreateTaskModalConsumesTriggerBeforeTeamLookups(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &blockingTeamListRepo{
		mockRepo: &mockRepo{
			teams: []slackrepository.TeamRecord{{ID: teamID, Code: "WEB", Name: "Web"}},
			statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
				teamID: {{ID: uuid.New(), Name: "Todo", Category: "unstarted"}},
			},
			teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		},
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	modalOpened := make(chan struct{}, 1)
	modalUpdated := make(chan struct{}, 1)
	requestErrors := make(chan error, 2)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var payload map[string]any
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			requestErrors <- err
		}
		switch req.URL.String() {
		case "https://slack.com/api/views.open":
			view, _ := payload["view"].(map[string]any)
			if _, hasSubmit := view["submit"]; hasSubmit {
				requestErrors <- errors.New("loading modal must not expose submit before hydration")
			}
			modalOpened <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V-loading"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			if payload["view_id"] != "V-loading" {
				requestErrors <- fmt.Errorf("updated view id = %#v", payload["view_id"])
			}
			view, _ := payload["view"].(map[string]any)
			if _, hasSubmit := view["submit"]; !hasSubmit {
				requestErrors <- errors.New("hydrated modal is missing submit")
			}
			modalUpdated <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
	})}

	result := make(chan error, 1)
	go func() {
		result <- service.openCreateTaskModal(
			context.Background(),
			"trigger",
			"Ship it",
			"Created from Slack",
			requestSourceContext{SlackTeamID: "T123", SlackUserID: "U123"},
			workspaceID,
			actorID,
			"xoxb-token",
		)
	}()

	select {
	case <-modalOpened:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("loading modal was not opened before team lookup")
	}
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("team lookup did not start after opening the loading modal")
	}
	close(repo.release)
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out hydrating create story modal")
	}
	select {
	case <-modalUpdated:
	default:
		t.Fatal("create story modal was not hydrated")
	}
	close(requestErrors)
	for err := range requestErrors {
		require.NoError(t, err)
	}
}

func TestHandleMessageActionPostsPrivateFailureFeedback(t *testing.T) {
	repo := &mockRepo{err: errors.New("database unavailable")}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	feedback := make(chan CommandResponse, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://hooks.slack.com/actions/T123/message" {
			return nil, fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		}
		var response CommandResponse
		if err := json.NewDecoder(req.Body).Decode(&response); err != nil {
			return nil, err
		}
		feedback <- response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]any{"id": "T123"},
		"channel":      map[string]any{"id": "C123"},
		"user":         map[string]any{"id": "U123"},
		"message":      map[string]any{"user": "U456", "text": "Create a task", "ts": "171234.000100"},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	select {
	case failure := <-feedback:
		require.Equal(t, "ephemeral", failure.ResponseType)
		require.Contains(t, failure.Text, "could not update")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for private message-action failure feedback")
	}
}

func TestHandleBlockActionsRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	apiCalls := 0
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		apiCalls++
		return nil, errors.New("views.update must not be called")
	})}

	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source:         requestSourceContext{SlackTeamID: "T123", SlackUserID: "U123"},
		SelectedTeamID: allowedTeamID.String(),
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"type": "block_actions",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": string(metadata),
			"state": map[string]any{"values": map[string]any{
				"team":  map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
				"title": map[string]any{"value": map[string]any{"value": "Unauthorized task"}},
			}},
		},
		"actions": []map[string]any{{
			"block_id":        modalBlockTeam,
			"action_id":       modalActionTeamSelect,
			"selected_option": map[string]any{"value": blockedTeamID.String()},
		}},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))
	resp, err := service.handleBlockActions(context.Background(), payload)
	require.ErrorIs(t, err, ErrSlackTeamNotAvailable)
	require.Zero(t, resp.StatusCode)
	require.Zero(t, apiCalls)
}
