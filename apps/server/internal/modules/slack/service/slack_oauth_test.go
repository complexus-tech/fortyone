package slack

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateInstallSessionStoresOpaqueBoundStateWithCoreScopes(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
	})

	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	require.Equal(t, slackBotOAuthScopeValue(), installURL.Query().Get("scope"))
	state := installURL.Query().Get("state")
	require.NotEmpty(t, state)
	require.NotContains(t, state, ".")

	nonce, err := base64.RawURLEncoding.DecodeString(state)
	require.NoError(t, err)
	require.Len(t, nonce, slackOpaqueNonceSize)
	digest := sha256.Sum256(nonce)
	store := service.nonces.(*mockNonceStore)
	record, ok := store.records[nonceStoreKey(slackProviderMessaging, slackNoncePurposeOAuth, digest[:])]
	require.True(t, ok)
	require.Equal(t, workspaceID, record.WorkspaceID)
	require.NotNil(t, record.UserID)
	require.Equal(t, userID, *record.UserID)
	require.True(t, record.ExpiresAt.Equal(service.clock.Now().Add(slackOAuthNonceTTL)))
	require.JSONEq(t, `{"workspace_slug":"acme"}`, string(record.Payload))
}

func TestCreateAccountLinkSessionBindsDashboardUserAndConnectedSlackTeam(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		SlackTeamID: "T123",
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
	})

	session, err := service.CreateAccountLinkSession(
		context.Background(), workspaceID, userID, "acme", "https://acme.fortyone.app/teams/team-1/requests/request-1",
	)
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	require.False(t, session.Linked)
	require.True(t, session.CanLink)
	require.Equal(t, "T123", installURL.Query().Get("team"))
	require.Equal(t, slackBotOAuthScopeValue(), installURL.Query().Get("scope"))

	nonce, err := base64.RawURLEncoding.DecodeString(installURL.Query().Get("state"))
	require.NoError(t, err)
	digest := sha256.Sum256(nonce)
	record := service.nonces.(*mockNonceStore).records[nonceStoreKey(slackProviderMessaging, slackNoncePurposeAccount, digest[:])]
	require.Equal(t, workspaceID, record.WorkspaceID)
	require.Equal(t, "T123", valueOrEmpty(record.ExternalWorkspaceID))
	require.Equal(t, userID, *record.UserID)
	require.JSONEq(t, `{"workspace_slug":"acme","return_url":"https://acme.fortyone.app/teams/team-1/requests/request-1"}`, string(record.Payload))
}

func TestHandleSetupLinksDashboardOAuthUserAndReturnsToRequest(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		SlackTeamID: "T123",
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
	})
	session, err := service.CreateAccountLinkSession(
		context.Background(), workspaceID, userID, "acme", "https://acme.fortyone.app/teams/team-1/requests/request-1",
	)
	require.NoError(t, err)
	state, err := url.Parse(session.InstallURL)
	require.NoError(t, err)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/oauth.v2.access", request.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-token","team":{"id":"T123","name":"Acme","domain":"acme"},"authed_user":{"id":"U123"}}`))
	}))
	defer provider.Close()
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	redirectURL, err := service.HandleSetup(context.Background(), "oauth-code", state.Query().Get("state"), "")
	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/teams/team-1/requests/request-1?slack_link_status=success", redirectURL)
	require.Equal(t, userID, repo.slackUserLinks["T123:U123"])
}

func TestHandleSetupConsumesStateAndStoresEncryptedInstallationMetadata(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{ID: uuid.New(), WorkspaceID: workspaceID}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://fortyone.app",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	state := installURL.Query().Get("state")

	requests := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/oauth.v2.access":
			require.NoError(t, req.ParseForm())
			require.Equal(t, "oauth-code", req.Form.Get("code"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"access_token": "xoxb-secret",
				"refresh_token": "xoxe-refresh",
				"expires_in": 3600,
				"bot_user_id": "UBOT",
				"app_id": "A123",
				"scope": "app_mentions:read,channels:history,channels:read,chat:write,chat:write.public,commands,groups:history,groups:read,im:history,links:read,links:write",
				"team": {"id": "T123", "name": "Acme", "domain": "acme"},
				"enterprise": {"id": "E123"},
				"authed_user": {"id": "UADMIN"}
			}`))
		case "/conversations.list":
			require.Equal(t, "Bearer xoxb-secret", req.Header.Get("Authorization"))
			require.Equal(t, "200", req.URL.Query().Get("limit"))
			require.Equal(t, "public_channel,private_channel", req.URL.Query().Get("types"))
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [{
					"id": "C123",
					"name": "general",
					"is_private": false,
					"is_archived": false,
					"is_member": true
				}],
				"response_metadata": {"next_cursor": ""}
			}`))
		default:
			t.Fatalf("unexpected Slack API path %q", req.URL.Path)
		}
	}))
	defer testServer.Close()
	service.client = testServer.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = testServer.URL

	redirectURL, err := service.HandleSetup(context.Background(), "oauth-code", state, "")
	require.NoError(t, err)
	require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", redirectURL)
	require.Equal(t, userID, repo.lastOAuthUserID)
	require.NotEqual(t, "xoxb-secret", repo.lastOAuthInstall.BotAccessToken)
	require.NotEqual(t, uuid.Nil, repo.lastOAuthInstall.InstallGeneration)
	require.Positive(t, repo.lastOAuthInstall.CredentialVersion)
	require.Equal(t, "A123", valueOrEmpty(repo.lastOAuthInstall.SlackAppID))
	require.Equal(t, "E123", valueOrEmpty(repo.lastOAuthInstall.EnterpriseID))
	require.Equal(t, "UADMIN", valueOrEmpty(repo.lastOAuthInstall.AuthedUserID))
	credential, version, err := service.credentials.open(slackCredentialBinding{
		WorkspaceID:       workspaceID,
		SlackTeamID:       repo.lastOAuthInstall.SlackTeamID,
		InstallGeneration: repo.lastOAuthInstall.InstallGeneration,
	}, repo.lastOAuthInstall.BotAccessToken)
	require.NoError(t, err)
	require.Equal(t, repo.lastOAuthInstall.CredentialVersion, version)
	require.Equal(t, "xoxb-secret", credential.AccessToken)
	require.Equal(t, "xoxe-refresh", credential.RefreshToken)
	require.NotNil(t, credential.ExpiresAt)
	require.True(t, credential.ExpiresAt.Equal(service.clock.Now().Add(time.Hour)))
	require.Equal(t, 1, repo.upsertChannels)
	require.Equal(t, workspaceID, repo.lastChannelWorkspaceID)
	require.Equal(t, repo.slackWorkspace.ID, repo.lastChannelInstallID)
	require.Equal(t, []slackrepository.SlackChannelPayload{{
		SlackChannelID: "C123",
		Name:           "general",
		IsMember:       true,
	}}, repo.lastChannels)
	require.Equal(t, 2, requests)

	_, err = service.HandleSetup(context.Background(), "oauth-code", state, "")
	require.Error(t, err)
	require.Equal(t, 2, requests)
}

func TestHandleSetupRejectsExpiredStateBeforeOAuthExchange(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	service.clock = fixedClock{now: service.clock.Now().Add(slackOAuthNonceTTL + time.Second)}
	apiCalls := 0
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		apiCalls++
		return nil, errors.New("OAuth exchange must not run")
	})}

	_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

	require.ErrorContains(t, err, "invalid or expired")
	require.Zero(t, apiCalls)
}

func TestHandleSetupCleansUpOnlyConclusiveUnownedWorkspaceConflict(t *testing.T) {
	tests := []struct {
		name               string
		upsertErr          error
		currentSlackTeamID string
		teamLookupErr      error
		wantUninstall      bool
	}{
		{
			name:               "workspace connected to another team and selected team unowned",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
			currentSlackTeamID: "T-OLD",
			teamLookupErr:      slackdomain.ErrNotFound,
			wantUninstall:      true,
		},
		{
			name:               "selected team is legitimately owned elsewhere",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrSlackTeamAlreadyConnected),
			currentSlackTeamID: "T-NEW",
			wantUninstall:      false,
		},
		{
			name:               "uncertain commit is visible on reread",
			upsertErr:          fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
			currentSlackTeamID: "T-NEW",
			wantUninstall:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspaceID := uuid.New()
			userID := uuid.New()
			repo := &mockRepo{
				upsertSlackWorkspaceErr: test.upsertErr,
				getSlackTeamErr:         test.teamLookupErr,
				slackWorkspace: slackrepository.SlackWorkspaceRecord{
					ID:                uuid.New(),
					WorkspaceID:       workspaceID,
					SlackTeamID:       test.currentSlackTeamID,
					InstallGeneration: uuid.New(),
					IsActive:          true,
				},
			}
			service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "https://api.example.com/integrations/slack/setup",
				WebsiteURL:   "https://app.example.com",
			})
			session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
			require.NoError(t, err)
			installURL, err := url.Parse(session.InstallURL)
			require.NoError(t, err)
			uninstallCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/oauth.v2.access":
					_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-new","bot_user_id":"B1","team":{"id":"T-NEW","name":"New"},"authed_user":{"id":"U1"}}`))
				case "/apps.uninstall":
					uninstallCalls++
					require.NoError(t, req.ParseForm())
					require.Equal(t, "xoxb-new", req.Form.Get("token"))
					_, _ = w.Write([]byte(`{"ok":true}`))
				default:
					http.NotFound(w, req)
				}
			}))
			defer server.Close()
			service.client = server.Client()
			service.webClient = newSlackWebClient(service.client)
			service.webClient.baseURL = server.URL

			_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

			require.Error(t, err)
			require.Equal(t, test.wantUninstall, uninstallCalls == 1)
			if test.wantUninstall {
				require.Len(t, repo.uninstallInputs, 1)
				require.Equal(t, slackdomain.UninstallOrphanedOAuth, repo.uninstallInputs[0].UninstallKind)
				require.NotEqual(t, "xoxb-new", repo.uninstallInputs[0].CredentialPayload)
			} else {
				require.Empty(t, repo.uninstallInputs)
			}
		})
	}
}

func TestHandleSetupDoesNotUninstallAfterUncertainCleanupPersistenceFailure(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &mockRepo{
		upsertSlackWorkspaceErr: fmt.Errorf("%w: %w", slackrepository.ErrActiveInstallationConflict, slackrepository.ErrWorkspaceAlreadyConnected),
		getSlackTeamErr:         slackdomain.ErrNotFound,
		enqueueUninstallErr:     errors.New("database commit outcome is unknown"),
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T-OLD",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://api.example.com/integrations/slack/setup",
		WebsiteURL:   "https://app.example.com",
	})
	session, err := service.CreateInstallSession(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	installURL, err := url.Parse(session.InstallURL)
	require.NoError(t, err)
	uninstallCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/oauth.v2.access":
			_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-new","bot_user_id":"B1","team":{"id":"T-NEW","name":"New"},"authed_user":{"id":"U1"}}`))
		case "/apps.uninstall":
			uninstallCalls++
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()
	service.client = server.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = server.URL

	_, err = service.HandleSetup(context.Background(), "oauth-code", installURL.Query().Get("state"), "")

	require.Error(t, err)
	require.Zero(t, uninstallCalls, "an uncertain enqueue must never trigger an unguarded apps.uninstall call")
	require.Len(t, repo.uninstallInputs, 1)
}

func TestRecordRequestLogPersistsOnlyStructuredNonSensitiveFields(t *testing.T) {
	workspaceID := uuid.New()
	repo := &mockRepo{workspace: slackrepository.WorkspaceRecord{ID: workspaceID}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	body := []byte(url.Values{
		"team_id":    {"T123"},
		"user_id":    {"U123"},
		"channel_id": {"C123"},
		"command":    {"/fortyone"},
		"trigger_id": {"sensitive-trigger"},
		"text":       {"customer-sensitive content"},
	}.Encode())

	service.RecordRequestLog(context.Background(), CoreRequestLogInput{
		RequestType: "commands",
		Endpoint:    "/integrations/slack/commands",
		RawBody:     body,
		Headers: map[string]string{
			"X-Slack-Signature":    "sensitive-signature",
			"X-Slack-Retry-Num":    "2",
			"X-Slack-Retry-Reason": "http_timeout",
		},
		Outcome: "processed",
	})

	require.Equal(t, &workspaceID, repo.lastRequestLog.WorkspaceID)
	require.Equal(t, "T123", valueOrEmpty(repo.lastRequestLog.SlackTeamID))
	require.Equal(t, "U123", valueOrEmpty(repo.lastRequestLog.SlackUserID))
	require.Equal(t, "/fortyone", valueOrEmpty(repo.lastRequestLog.Command))
	require.Nil(t, repo.lastRequestLog.TriggerID)
	require.Nil(t, repo.lastRequestLog.RequestBody)
	require.JSONEq(t, `{"X-Slack-Retry-Num":"2","X-Slack-Retry-Reason":"http_timeout"}`, string(repo.lastRequestLog.Headers))
}
