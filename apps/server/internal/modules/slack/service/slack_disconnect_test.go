package slack

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDisconnectWorkspaceDeletesSlackWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret"})
	installGeneration := uuid.New()
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: "T123", InstallGeneration: installGeneration}, slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: installGeneration,
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/apps.uninstall", req.URL.String())
		require.NoError(t, req.ParseForm())
		require.Equal(t, "client", req.Form.Get("client_id"))
		require.Equal(t, "secret", req.Form.Get("client_secret"))
		require.Equal(t, "xoxb-token", req.Form.Get("token"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID, uuid.New())

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Equal(t, slackrepository.SlackWorkspaceRecord{}, repo.slackWorkspace)
	require.Len(t, repo.completedUninstalls, 1)
	require.Empty(t, repo.uninstalls[repo.completedUninstalls[0]].CredentialPayload)
}

func TestDisconnectWorkspaceRevokesLocalInstallWhenSlackUninstallFails(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret"})
	installGeneration := uuid.New()
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: "T123", InstallGeneration: installGeneration}, slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: installGeneration,
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error":"invalid_auth"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID, uuid.New())

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Equal(t, slackrepository.SlackWorkspaceRecord{}, repo.slackWorkspace)
	require.Len(t, repo.completedUninstalls, 1)
}

func TestDisconnectWorkspaceRetainsEncryptedCredentialForTransientUninstallRetry(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{ClientID: "client", ClientSecret: "secret"})
	installGeneration := uuid.New()
	credentialPayload, credentialVersion, err := service.credentials.seal(slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: "T123", InstallGeneration: installGeneration}, slackCredential{AccessToken: "xoxb-token"})
	require.NoError(t, err)
	repo.slackWorkspace = slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: installGeneration,
		IsActive:          true,
	}
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error":"internal_error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err = service.DisconnectWorkspace(context.Background(), workspaceID, uuid.New())

	require.NoError(t, err)
	require.True(t, repo.disconnected)
	require.Len(t, repo.failedUninstalls, 1)
	record := repo.uninstalls[repo.failedUninstalls[0]]
	require.Equal(t, slackdomain.UninstallFailed, record.Status)
	require.Equal(t, credentialPayload, record.CredentialPayload)
	require.NotNil(t, record.NextAttemptAt)
}
