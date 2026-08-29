package slack

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLinkSlackAccountCreatesManualMapping(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{
			ID:   workspaceID,
			Slug: "acme",
			Name: "Acme",
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		WebsiteURL: "https://fortyone.app",
	})

	link, err := service.buildSlackUserLinkURL(context.Background(), workspaceID, "T123", "U999")
	require.NoError(t, err)
	parsedLink, err := url.Parse(link)
	require.NoError(t, err)
	require.Equal(t, "acme.fortyone.app", parsedLink.Host)
	require.Equal(t, "/settings/integrations/slack", parsedLink.Path)
	token := parsedLink.Query().Get("slack_link_token")
	require.NotEmpty(t, token)
	require.NotContains(t, token, ".")
	nonce, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	digest := sha256.Sum256(nonce)
	storeKey := nonceStoreKey(slackProviderMessaging, slackNoncePurposeAccount, digest[:])
	store := service.nonces.(*mockNonceStore)

	result, err := service.LinkSlackAccount(context.Background(), workspaceID, userID, token)
	require.NoError(t, err)
	require.False(t, result.AlreadyLinked)

	require.NotNil(t, repo.slackUserLinks)
	require.Equal(t, userID, repo.slackUserLinks["T123:U999"])
	require.NotNil(t, store.records[storeKey].UserID)
	require.Equal(t, userID, *store.records[storeKey].UserID)
	result, err = service.LinkSlackAccount(context.Background(), workspaceID, userID, token)
	require.NoError(t, err)
	require.True(t, result.AlreadyLinked)
	require.Equal(t, "U999", result.SlackUserID)
}

func TestAutoLinkWorkspaceMembersMatchesExactNormalizedEmail(t *testing.T) {
	workspaceID := uuid.New()
	matchedUserID := uuid.New()
	unmatchedUserID := uuid.New()
	scopes := slackBotOAuthScopeValue()
	repo := &mockRepo{
		workspaceMembers: []slackrepository.WorkspaceMemberRecord{
			{UserID: matchedUserID, Email: " Person@Example.com "},
			{UserID: unmatchedUserID, Email: "other@example.com"},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-current",
			Scope:          &scopes,
			IsActive:       true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/users.list", request.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok": true,
			"members": [
				{"id":"U-MATCH","name":"person","real_name":"Different Display Name","profile":{"email":"person@example.com"}},
				{"id":"U-NO-MATCH","name":"other-person","real_name":"Other Person","profile":{"email":"different@example.com"}},
				{"id":"U-BOT","is_bot":true,"profile":{"email":"other@example.com"}}
			],
			"response_metadata":{"next_cursor":""}
		}`))
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	require.NoError(t, service.autoLinkWorkspaceMembers(context.Background(), repo.slackWorkspace))
	require.Equal(t, matchedUserID, repo.slackUserLinks["T123:U-MATCH"])
	require.NotContains(t, repo.slackUserLinks, "T123:U-NO-MATCH")
	require.NotContains(t, repo.slackUserLinks, "T123:U-BOT")
}

func TestSlackDisconnectCommandRemovesOnlyCallingUsersAccountLink(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
			IsActive:    true,
		},
		slackUserLinks: map[string]uuid.UUID{
			"T123:U123": userID,
			"T123:U999": uuid.New(),
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	body := []byte(url.Values{
		"team_id":    {"T123"},
		"user_id":    {"U123"},
		"trigger_id": {"trigger-1"},
		"text":       {"disconnect"},
	}.Encode())

	response, err := service.HandleCommand(context.Background(), body)
	require.NoError(t, err)
	require.Equal(t, "ephemeral", response.ResponseType)
	require.Equal(t, "Your Slack account has been disconnected from FortyOne.", response.Text)
	require.NotContains(t, repo.slackUserLinks, "T123:U123")
	require.Contains(t, repo.slackUserLinks, "T123:U999")
}
