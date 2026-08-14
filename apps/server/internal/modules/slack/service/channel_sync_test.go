package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSyncChannelsRefreshesEntirePaginatedSnapshot(t *testing.T) {
	workspaceID := uuid.New()
	installationID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:             installationID,
		WorkspaceID:    workspaceID,
		SlackTeamID:    "T123",
		BotAccessToken: "xoxb-current",
		IsActive:       true,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	requestNumber := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/conversations.list", request.URL.Path)
		require.Equal(t, "Bearer xoxb-current", request.Header.Get("Authorization"))
		require.Equal(t, "200", request.URL.Query().Get("limit"))
		require.Equal(t, "public_channel,private_channel", request.URL.Query().Get("types"))
		w.Header().Set("Content-Type", "application/json")

		switch requestNumber {
		case 0:
			require.Empty(t, request.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [{
					"id": "C-GENERAL",
					"name": "general",
					"is_private": false,
					"is_archived": false,
					"is_member": true
				}],
				"response_metadata": {"next_cursor": "page-2"}
			}`))
		case 1:
			require.Equal(t, "page-2", request.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [{
					"id": "G-PRIVATE",
					"name": "leadership",
					"is_private": true,
					"is_archived": false,
					"is_member": true
				}],
				"response_metadata": {"next_cursor": ""}
			}`))
		case 2:
			require.Empty(t, request.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [
					{
						"id": "C-GENERAL",
						"name": "company-wide",
						"is_private": false,
						"is_archived": false,
						"is_member": true
					},
					{
						"id": "C-PRODUCT",
						"name": "product",
						"is_private": false,
						"is_archived": false,
						"is_member": false
					}
				],
				"response_metadata": {"next_cursor": ""}
			}`))
		default:
			t.Fatalf("unexpected conversations.list request %d", requestNumber+1)
		}
		requestNumber++
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	require.NoError(t, service.SyncChannels(context.Background(), workspaceID))
	require.Equal(t, 1, repo.upsertChannels)
	require.Equal(t, workspaceID, repo.lastChannelWorkspaceID)
	require.Equal(t, installationID, repo.lastChannelInstallID)
	require.Equal(t, []slackrepository.SlackChannelPayload{
		{
			SlackChannelID: "C-GENERAL",
			Name:           "general",
			IsMember:       true,
		},
		{
			SlackChannelID: "G-PRIVATE",
			Name:           "leadership",
			IsPrivate:      true,
			IsMember:       true,
		},
	}, repo.lastChannels)

	require.NoError(t, service.SyncChannels(context.Background(), workspaceID))
	require.Equal(t, 2, repo.upsertChannels)
	require.Equal(t, []slackrepository.SlackChannelPayload{
		{
			SlackChannelID: "C-GENERAL",
			Name:           "company-wide",
			IsMember:       true,
		},
		{
			SlackChannelID: "C-PRODUCT",
			Name:           "product",
		},
	}, repo.lastChannels)
	require.Equal(t, 3, requestNumber)
}

func TestSyncChannelsDoesNotPersistAPartialSnapshot(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		SlackTeamID:    "T123",
		BotAccessToken: "xoxb-current",
		IsActive:       true,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	requestNumber := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if requestNumber == 0 {
			require.Empty(t, request.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [{"id":"C-GENERAL","name":"general"}],
				"response_metadata": {"next_cursor": "page-2"}
			}`))
		} else {
			require.Equal(t, "page-2", request.URL.Query().Get("cursor"))
			_, _ = w.Write([]byte(`{"ok":false,"error":"internal_error"}`))
		}
		requestNumber++
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	err := service.SyncChannels(context.Background(), workspaceID)

	require.Error(t, err)
	require.Zero(t, repo.upsertChannels)
	require.Equal(t, 2, requestNumber)
}

func TestSyncChannelsRejectsRepeatedPaginationCursor(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		SlackTeamID:    "T123",
		BotAccessToken: "xoxb-current",
		IsActive:       true,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	requestNumber := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if requestNumber == 0 {
			require.Empty(t, request.URL.Query().Get("cursor"))
		} else {
			require.Equal(t, "page-2", request.URL.Query().Get("cursor"))
		}
		requestNumber++
		_, _ = w.Write([]byte(`{
			"ok": true,
			"channels": [{"id":"C-GENERAL","name":"general"}],
			"response_metadata": {"next_cursor": "page-2"}
		}`))
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	err := service.SyncChannels(context.Background(), workspaceID)

	require.ErrorContains(t, err, "repeated a cursor")
	require.Zero(t, repo.upsertChannels)
	require.Equal(t, 2, requestNumber)
}

func TestHandleSetupKeepsInstallationWhenInitialChannelSyncFails(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
		SecretKey:    "encryption-secret",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/oauth.v2.access":
			_, _ = w.Write([]byte(`{
				"ok": true,
				"access_token": "xoxb-secret",
				"team": {"id": "T123", "name": "Acme", "domain": "acme"},
				"authed_user": {"id": "UADMIN"}
			}`))
		case "/conversations.list":
			_, _ = w.Write([]byte(`{"ok":false,"error":"missing_scope"}`))
		default:
			t.Fatalf("unexpected Slack API path %q", request.URL.Path)
		}
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	redirectURL, err := service.HandleSetup(
		context.Background(),
		"oauth-code",
		installURL.Query().Get("state"),
		"",
	)

	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", redirectURL)
	require.Equal(t, userID, repo.lastOAuthUserID)
	require.Zero(t, repo.upsertChannels)
}
