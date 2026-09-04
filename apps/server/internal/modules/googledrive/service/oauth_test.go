package googledrive

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationURLUsesOnlyDriveFileScopes(t *testing.T) {
	t.Parallel()

	client := newGoogleClient(http.DefaultClient, Config{
		ClientID: "client-id", RedirectURL: "https://api.fortyone.app/integrations/google-drive/callback",
	}, oauthScopes)
	authorizationURL, err := client.AuthorizationURL("state", "verifier")
	require.NoError(t, err)
	parsed, err := url.Parse(authorizationURL)
	require.NoError(t, err)
	require.Equal(t, "openid email profile https://www.googleapis.com/auth/drive.file", parsed.Query().Get("scope"))
	require.Empty(t, parsed.Query().Get("include_granted_scopes"))
	require.Equal(t, "S256", parsed.Query().Get("code_challenge_method"))
	require.Equal(t, "offline", parsed.Query().Get("access_type"))
}

func TestExchangePreservesExplicitPartialConsentScopes(t *testing.T) {
	t.Parallel()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.NoError(t, request.ParseForm())
		require.Equal(t, "authorization_code", request.Form.Get("grant_type"))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"access_token":"access-token","refresh_token":"refresh-token","expires_in":3600,"scope":"openid email profile"}`))
	}))
	t.Cleanup(tokenServer.Close)

	client := newGoogleClient(tokenServer.Client(), Config{
		ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://api.fortyone.app/integrations/google-drive/callback",
	}, oauthScopes)
	client.tokenURL = tokenServer.URL

	_, scopes, err := client.Exchange(t.Context(), "code", "verifier")

	require.NoError(t, err)
	require.Equal(t, []string{"openid", "email", "profile"}, scopes)
	require.False(t, hasOAuthScope(scopes, googleDriveFileScope))
}

func TestExchangeDoesNotInferScopesWhenProviderOmitsScope(t *testing.T) {
	t.Parallel()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.NoError(t, request.ParseForm())
		require.Equal(t, "authorization_code", request.Form.Get("grant_type"))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"access_token":"access-token","refresh_token":"refresh-token","expires_in":3600}`))
	}))
	t.Cleanup(tokenServer.Close)

	client := newGoogleClient(tokenServer.Client(), Config{
		ClientID: "client-id", ClientSecret: "client-secret", RedirectURL: "https://api.fortyone.app/integrations/google-drive/callback",
	}, oauthScopes)
	client.tokenURL = tokenServer.URL

	_, scopes, err := client.Exchange(t.Context(), "code", "verifier")

	require.NoError(t, err)
	require.Empty(t, scopes)
	require.False(t, hasOAuthScope(scopes, googleDriveFileScope))
}

func TestCompleteOAuthRejectsDifferentAuthenticatedActorBeforeProviderExchange(t *testing.T) {
	t.Parallel()

	initiatingUserID := uuid.New()
	repo := &oauthRepositoryStub{state: domain.OAuthState{
		WorkspaceID: uuid.New(), UserID: initiatingUserID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
	}}
	client := &providerClientStub{}
	service := &Service{
		repo: repo, client: client, config: Config{WebsiteURL: "https://fortyone.app"}, now: time.Now,
	}

	redirectURL, err := service.CompleteOAuth(t.Context(), uuid.New(), "code", "opaque-state", "")

	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Equal(t, 1, repo.consumeCalls)
	require.Zero(t, client.exchangeCalls)
	require.Equal(t, "https://acme.fortyone.app/settings/account/google-drive?google_drive_error=connection_failed", redirectURL)
}

func TestCompleteOAuthConsumesStateThenFailsClosedWhenDriveIsNotConfigured(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &oauthRepositoryStub{state: domain.OAuthState{
		WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
	}}
	client := &providerClientStub{}
	service := &Service{
		repo: repo, client: client, config: Config{WebsiteURL: "https://fortyone.app"}, now: time.Now,
	}

	redirectURL, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

	require.ErrorIs(t, err, domain.ErrNotConfigured)
	require.Equal(t, 1, repo.consumeCalls)
	require.Zero(t, client.exchangeCalls)
	require.Equal(t, "https://acme.fortyone.app/settings/account/google-drive?google_drive_error=connection_failed", redirectURL)
}

func TestCompleteOAuthRejectsPartialConsentBeforeCredentialPersistence(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &oauthRepositoryStub{state: domain.OAuthState{
		WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
	}}
	client := &providerClientStub{
		exchangeToken:  domain.OAuthToken{AccessToken: "access-token", RefreshToken: "refresh-token"},
		exchangeScopes: []string{"openid", "email", "profile"},
		userInfoResult: ProviderUser{Subject: "google-subject", Email: "owner@example.com"},
	}
	config := configuredOAuthTestConfig(t)
	service := &Service{
		repo: repo, client: client, config: config, now: time.Now,
	}

	redirectURL, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

	require.ErrorContains(t, err, "required file access")
	require.Equal(t, 1, client.exchangeCalls)
	require.Zero(t, client.revokeCalls)
	require.Equal(t, 2, client.userInfoCalls)
	require.Zero(t, repo.upsertCalls)
	require.Equal(t, 1, repo.enqueueCalls)
	require.Nil(t, repo.enqueuedRevocation.SourceAccountID)
	require.Equal(t, userID, repo.enqueuedRevocation.UserID)
	require.Equal(t, "google-subject", repo.enqueuedRevocation.GoogleSubject)
	require.NotEmpty(t, repo.enqueuedRevocation.CredentialPayload)
	require.Equal(t, "https://acme.fortyone.app/settings/account/google-drive?google_drive_error=connection_failed", redirectURL)
}

func TestCompleteOAuthDoesNotRevokeGrantOwnedByAnotherFortyOneUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &oauthRepositoryStub{
		state: domain.OAuthState{
			WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
		},
		activeAccount:      domain.Account{UserID: uuid.New(), GoogleSubject: "shared-google-subject"},
		activeAccountFound: true,
	}
	client := &providerClientStub{
		exchangeToken: domain.OAuthToken{
			AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
		},
		exchangeScopes: oauthScopes,
		userInfoResult: ProviderUser{Subject: "shared-google-subject", Email: "owner@example.com"},
	}
	config := configuredOAuthTestConfig(t)
	service := &Service{
		repo: repo, client: client, config: config, now: time.Now,
	}

	_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

	require.ErrorIs(t, err, domain.ErrAccountOwned)
	require.Zero(t, client.revokeCalls)
	require.Zero(t, repo.upsertCalls)
	require.Zero(t, repo.enqueueCalls)
}

func TestCompleteOAuthHoldsProviderGatesAcrossExchangeAndPersistence(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	events := make([]string, 0, 10)
	repo := &oauthRepositoryStub{
		state: domain.OAuthState{
			WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
		},
		events: &events,
	}
	client := &providerClientStub{
		exchangeToken: domain.OAuthToken{
			AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
		},
		exchangeScopes: oauthScopes,
		userInfoResult: ProviderUser{Subject: "google-subject", Email: "owner@example.com"},
		events:         &events,
	}
	config := configuredOAuthTestConfig(t)
	service := &Service{
		repo: repo, client: client, config: config, now: time.Now,
	}

	_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

	require.NoError(t, err)
	require.Equal(t, []string{
		"user_lock_begin", "exchange", "user_info", "subject_lock_begin", "user_info",
		"get_subject", "get_connection", "upsert", "subject_lock_end", "user_lock_end",
	}, events)
}

func TestCompleteOAuthSecondIdentityFailureCleansUpOnlyAfterOwnershipProof(t *testing.T) {
	t.Parallel()

	identityErr := errors.New("Google identity unavailable")
	for _, test := range []struct {
		name                 string
		configureRepository  func(*oauthRepositoryStub)
		wantEnqueuedCleanup  int
		wantOwnershipFailure bool
	}{
		{name: "no owner stages encrypted cleanup", wantEnqueuedCleanup: 1},
		{
			name: "active owner prevents cleanup",
			configureRepository: func(repo *oauthRepositoryStub) {
				repo.activeAccountFound = true
				repo.activeAccount = domain.Account{UserID: uuid.New(), GoogleSubject: "google-subject"}
			},
		},
		{
			name: "indeterminate database result prevents cleanup",
			configureRepository: func(repo *oauthRepositoryStub) {
				repo.activeAccountErr = errors.New("database unavailable")
			},
			wantOwnershipFailure: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New()
			repo := &oauthRepositoryStub{state: domain.OAuthState{
				WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
			}}
			if test.configureRepository != nil {
				test.configureRepository(repo)
			}
			client := &providerClientStub{
				exchangeToken: domain.OAuthToken{
					AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
				},
				exchangeScopes: oauthScopes,
				userInfoResults: []ProviderUser{
					{Subject: "google-subject", Email: "owner@example.com"},
				},
				userInfoErrors: []error{nil, identityErr},
			}
			service := &Service{
				repo: repo, client: client, config: configuredOAuthTestConfig(t), now: time.Now,
			}

			_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

			require.ErrorIs(t, err, identityErr)
			if test.wantOwnershipFailure {
				require.ErrorContains(t, err, "verify Google account ownership before cleanup")
			}
			require.Equal(t, 2, client.userInfoCalls)
			require.Equal(t, test.wantEnqueuedCleanup, repo.enqueueCalls)
			require.Zero(t, client.revokeCalls)
			require.Zero(t, repo.upsertCalls)
		})
	}
}

func TestCompleteOAuthStagesUnpersistedGrantForRecoverableCallbackFailures(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		configure  func(*oauthRepositoryStub, *providerClientStub)
		wantError  string
		wantUpsert int
	}{
		{
			name: "workspace is bound to another Google account",
			configure: func(repo *oauthRepositoryStub, _ *providerClientStub) {
				repo.connectionFound = true
				repo.connection = domain.Connection{Account: domain.Account{GoogleSubject: "different-subject"}}
			},
			wantError: "disconnect the current Google Drive account",
		},
		{
			name: "Google omits offline refresh token",
			configure: func(_ *oauthRepositoryStub, client *providerClientStub) {
				client.exchangeToken.RefreshToken = ""
			},
			wantError: "offline refresh token",
		},
		{
			name: "connection persistence fails",
			configure: func(repo *oauthRepositoryStub, _ *providerClientStub) {
				repo.upsertErr = errors.New("database unavailable")
			},
			wantError:  "database unavailable",
			wantUpsert: 1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New()
			repo := &oauthRepositoryStub{state: domain.OAuthState{
				WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
			}}
			client := &providerClientStub{
				exchangeToken: domain.OAuthToken{
					AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
				},
				exchangeScopes: oauthScopes,
				userInfoResult: ProviderUser{Subject: "google-subject", Email: "owner@example.com"},
			}
			test.configure(repo, client)
			service := &Service{
				repo: repo, client: client, config: configuredOAuthTestConfig(t), now: time.Now,
			}

			_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

			require.ErrorContains(t, err, test.wantError)
			require.Equal(t, 1, repo.enqueueCalls)
			require.Nil(t, repo.enqueuedRevocation.SourceAccountID)
			require.Equal(t, test.wantUpsert, repo.upsertCalls)
			require.Zero(t, client.revokeCalls)
		})
	}
}

func TestCompleteOAuthFallsBackToImmediateRevokeOnlyAfterProvenOwnershipAbsence(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		configure func(*oauthRepositoryStub, *Config)
	}{
		{
			name: "vault seal failure",
			configure: func(_ *oauthRepositoryStub, config *Config) {
				config.Credentials = failingSealVault{
					CredentialVault: config.Credentials,
					err:             errors.New("vault unavailable"),
				}
			},
		},
		{
			name: "outbox persistence failure",
			configure: func(repo *oauthRepositoryStub, _ *Config) {
				repo.enqueueErr = errors.New("outbox unavailable")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New()
			repo := &oauthRepositoryStub{state: domain.OAuthState{
				WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
			}}
			client := &providerClientStub{
				exchangeToken: domain.OAuthToken{
					AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
				},
				exchangeScopes: []string{"openid", "email", "profile"},
				userInfoResult: ProviderUser{Subject: "google-subject", Email: "owner@example.com"},
			}
			config := configuredOAuthTestConfig(t)
			test.configure(repo, &config)
			service := &Service{repo: repo, client: client, config: config, now: time.Now}

			_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

			require.ErrorContains(t, err, "required file access")
			require.Equal(t, 1, client.revokeCalls)
			require.Equal(t, "refresh-token", client.revokedToken)
		})
	}
}

func TestUnpersistedGrantCleanupNeverRevokesWithoutProvenOwnershipAbsence(t *testing.T) {
	t.Parallel()

	repo := &oauthRepositoryStub{enqueueErr: errors.New("ownership query unavailable")}
	client := &providerClientStub{}
	service := &Service{
		repo: repo, client: client, config: configuredOAuthTestConfig(t), now: time.Now,
	}

	err := service.stageUnpersistedGrantRevocation(
		t.Context(),
		domain.OAuthState{WorkspaceID: uuid.New(), UserID: uuid.New()},
		ProviderUser{Subject: "google-subject"},
		domain.OAuthToken{AccessToken: "access-token", RefreshToken: "refresh-token"},
		false,
	)

	require.ErrorContains(t, err, "ownership query unavailable")
	require.Equal(t, 1, repo.enqueueCalls)
	require.Zero(t, client.revokeCalls)
}

func TestCompleteOAuthAmbiguousUpsertAndCleanupFailureNeverRevokes(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repo := &oauthRepositoryStub{
		state: domain.OAuthState{
			WorkspaceID: uuid.New(), UserID: userID, WorkspaceSlug: "acme", CodeVerifier: "verifier",
		},
		upsertErr:  errors.New("commit outcome unknown"),
		enqueueErr: errors.New("replacement connection unavailable"),
	}
	client := &providerClientStub{
		exchangeToken: domain.OAuthToken{
			AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
		},
		exchangeScopes: oauthScopes,
		userInfoResult: ProviderUser{Subject: "google-subject", Email: "owner@example.com"},
	}
	service := &Service{
		repo: repo, client: client, config: configuredOAuthTestConfig(t), now: time.Now,
	}

	_, err := service.CompleteOAuth(t.Context(), userID, "code", "opaque-state", "")

	require.ErrorContains(t, err, "commit outcome unknown")
	require.ErrorContains(t, err, "replacement connection unavailable")
	require.Equal(t, 1, repo.upsertCalls)
	require.Equal(t, 1, repo.enqueueCalls)
	require.Zero(t, client.revokeCalls)
}
