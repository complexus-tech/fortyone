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

func TestHandleSetupCompletesSameWorkspaceSameTeamOAuthRefresh(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	refreshedGeneration := uuid.New()
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T-PRIMARY",
			InstallGeneration: refreshedGeneration,
			IsActive:          true,
		},
	}
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

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/oauth.v2.access", req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": true,
			"access_token": "xoxb-refreshed",
			"bot_user_id": "UBOT",
			"team": {"id": "T-PRIMARY", "name": "Primary", "domain": "primary"},
			"authed_user": {"id": "UADMIN"}
		}`))
	}))
	defer testServer.Close()
	service.client = testServer.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = testServer.URL

	redirectURL, err := service.HandleSetup(
		context.Background(),
		"oauth-code",
		installURL.Query().Get("state"),
		"",
	)

	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", redirectURL)
	require.Equal(t, workspaceID, repo.slackWorkspace.WorkspaceID)
	require.Equal(t, userID, repo.lastOAuthUserID)
	require.Equal(t, "T-PRIMARY", repo.lastOAuthInstall.SlackTeamID)
	require.Equal(t, "Primary", repo.lastOAuthInstall.SlackTeamName)
	require.NotEqual(t, "xoxb-refreshed", repo.lastOAuthInstall.BotAccessToken)
	require.Empty(t, repo.uninstallInputs)
}
