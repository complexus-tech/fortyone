package agentreadinesshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOAuthClientCredentialsRequiresBasicAndReturnsNoRefreshToken(t *testing.T) {
	t.Parallel()

	oauth := newStubOAuthPlatform()
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})
	installationID := uuid.New()
	form := url.Values{
		"grant_type":      {"client_credentials"},
		"installation_id": {installationID.String()},
		"resource":        {oauth.apiResource},
		"scope":           {"stories:write"},
	}
	request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(oauth.application.ClientID, "f41_ocs_v1_secret")
	recorder := httptest.NewRecorder()
	ctx := web.SetValues(context.Background(), &web.Values{RequestID: "request-client-credentials"})

	require.NoError(t, handler.Token(ctx, recorder, request))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"access_token":"valid-application-access-token"`)
	require.NotContains(t, recorder.Body.String(), "refresh_token")
	require.Equal(t, developeroauth.ClientCredentialsExchange{
		ClientID: oauth.application.ClientID, ClientSecret: "f41_ocs_v1_secret",
		InstallationID: installationID, Resource: oauth.apiResource, Scopes: []string{"stories:write"},
		RequestID: "request-client-credentials",
	}, oauth.clientCredentialsRequest())
}

func TestOAuthClientCredentialsNeverReturnsTokenWhenAuditPersistenceFails(t *testing.T) {
	t.Parallel()

	oauth := newStubOAuthPlatform()
	oauth.exchangeClientCredentialsErr = errors.New("immutable audit unavailable")
	handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})
	request := httptest.NewRequest(
		http.MethodPost,
		"/oauth/token",
		strings.NewReader(clientCredentialsForm(uuid.New()).Encode()),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(oauth.application.ClientID, "f41_ocs_v1_secret")
	recorder := httptest.NewRecorder()

	err := handler.Token(context.Background(), recorder, request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable audit unavailable")
	require.NotContains(t, recorder.Body.String(), oauth.applicationToken.AccessToken.Reveal())
}

func TestOAuthClientCredentialsRejectsEveryNonBasicSecretChannel(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		target  string
		form    url.Values
		headers func(*http.Request)
	}{
		{
			name: "missing authorization",
			form: clientCredentialsForm(uuid.New()),
		},
		{
			name:    "bearer authorization",
			form:    clientCredentialsForm(uuid.New()),
			headers: func(request *http.Request) { request.Header.Set("Authorization", "Bearer secret") },
		},
		{
			name: "body client secret",
			form: func() url.Values {
				values := clientCredentialsForm(uuid.New())
				values.Set("client_secret", "body-secret")
				return values
			}(),
			headers: func(request *http.Request) { request.SetBasicAuth("client", "basic-secret") },
		},
		{
			name: "body client id",
			form: func() url.Values {
				values := clientCredentialsForm(uuid.New())
				values.Set("client_id", "body-client")
				return values
			}(),
			headers: func(request *http.Request) { request.SetBasicAuth("client", "basic-secret") },
		},
		{
			name:    "query client id",
			target:  "/oauth/token?client_id=query-client",
			form:    clientCredentialsForm(uuid.New()),
			headers: func(request *http.Request) { request.SetBasicAuth("client", "basic-secret") },
		},
		{
			name:    "query client secret",
			target:  "/oauth/token?client_secret=query-secret",
			form:    clientCredentialsForm(uuid.New()),
			headers: func(request *http.Request) { request.SetBasicAuth("client", "basic-secret") },
		},
		{
			name: "duplicate authorization",
			form: clientCredentialsForm(uuid.New()),
			headers: func(request *http.Request) {
				request.SetBasicAuth("client", "basic-secret")
				request.Header.Add("Authorization", "Basic Y2xpZW50Om90aGVy")
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			oauth := newStubOAuthPlatform()
			handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})
			target := test.target
			if target == "" {
				target = "/oauth/token"
			}
			request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(test.form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if test.headers != nil {
				test.headers(request)
			}
			recorder := httptest.NewRecorder()

			require.NoError(t, handler.Token(context.Background(), recorder, request))
			require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"error":"invalid_client"`)
			require.Contains(t, recorder.Header().Get("WWW-Authenticate"), "invalid_client")
			require.Equal(t, developeroauth.ClientCredentialsExchange{}, oauth.clientCredentialsRequest())
			require.NotContains(t, recorder.Body.String(), "secret")
		})
	}
}

func TestOAuthClientCredentialsRejectsAmbiguousOrOversizedFormFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(url.Values)
	}{
		{name: "repeated grant type", mutate: func(values url.Values) { values.Add("grant_type", "client_credentials") }},
		{name: "repeated installation", mutate: func(values url.Values) { values.Add("installation_id", uuid.NewString()) }},
		{name: "repeated resource", mutate: func(values url.Values) { values.Add("resource", "https://other.example/api") }},
		{name: "repeated scope", mutate: func(values url.Values) { values.Add("scope", "stories:read") }},
		{name: "missing scope", mutate: func(values url.Values) { values.Del("scope") }},
		{name: "oversized resource", mutate: func(values url.Values) {
			values.Set("resource", strings.Repeat("x", maximumOAuthResourceBytes+1))
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			oauth := newStubOAuthPlatform()
			handler := New(Config{APIPublicURL: "https://api.fortyone.app", OAuth: oauth})
			form := clientCredentialsForm(uuid.New())
			test.mutate(form)
			request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.SetBasicAuth(oauth.application.ClientID, "f41_ocs_v1_secret")
			recorder := httptest.NewRecorder()

			require.NoError(t, handler.Token(context.Background(), recorder, request))
			require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"error":"invalid_request"`)
			require.Equal(t, developeroauth.ClientCredentialsExchange{}, oauth.clientCredentialsRequest())
		})
	}
}

func clientCredentialsForm(installationID uuid.UUID) url.Values {
	return url.Values{
		"grant_type":      {"client_credentials"},
		"installation_id": {installationID.String()},
		"resource":        {"https://api.fortyone.app/api/v1"},
		"scope":           {"stories:write"},
	}
}
