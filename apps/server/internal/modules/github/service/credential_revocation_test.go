package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRevokeUserOAuthTokenUsesBasicAuthAndKeepsSecretsOutOfURL(t *testing.T) {
	t.Parallel()
	const (
		clientID     = "oauth-client-id"
		clientSecret = "oauth-client-secret"
		accessToken  = "gho_access-token"
	)
	service := &Service{
		cfg: Config{ClientID: clientID, ClientSecret: clientSecret},
		httpClient: &http.Client{Transport: githubOAuthRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodDelete || request.URL.Path != "/applications/"+clientID+"/token" {
				t.Fatalf("request = %s %s", request.Method, request.URL.Path)
			}
			if strings.Contains(request.URL.String(), accessToken) || strings.Contains(request.URL.String(), clientSecret) {
				t.Fatalf("credential appeared in URL: %s", request.URL)
			}
			username, password, ok := request.BasicAuth()
			if !ok || username != clientID || password != clientSecret {
				t.Fatal("OAuth application Basic authentication was not applied")
			}
			body, err := io.ReadAll(request.Body)
			if err != nil || !strings.Contains(string(body), accessToken) {
				t.Fatalf("request body = %q, error = %v", body, err)
			}
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    request,
			}, nil
		})},
	}

	if err := service.revokeUserOAuthToken(context.Background(), accessToken); err != nil {
		t.Fatalf("revokeUserOAuthToken() error = %v", err)
	}
}

func TestRevokeUserOAuthTokenIsIdempotentAndReturnsSafeErrors(t *testing.T) {
	t.Parallel()
	const secret = "must-not-leak"
	for _, test := range []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{name: "already revoked", statusCode: http.StatusNotFound},
		{name: "provider unavailable", statusCode: http.StatusInternalServerError, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &Service{
				cfg: Config{ClientID: "client-id", ClientSecret: secret},
				httpClient: &http.Client{Transport: githubOAuthRoundTripper(func(request *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: test.statusCode,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(`{"message":"provider detail"}`)),
						Request:    request,
					}, nil
				})},
			}
			err := service.revokeUserOAuthToken(context.Background(), "access-"+secret)
			if test.wantError && !errors.Is(err, ErrGitHubOAuthRevocationUnavailable) {
				t.Fatalf("revokeUserOAuthToken() error = %v", err)
			}
			if !test.wantError && err != nil {
				t.Fatalf("revokeUserOAuthToken() error = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), secret) {
				t.Fatalf("error disclosed credential: %v", err)
			}
		})
	}
}
