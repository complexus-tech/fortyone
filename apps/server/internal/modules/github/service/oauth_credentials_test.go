package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type githubOAuthRoundTripper func(*http.Request) (*http.Response, error)

func (fn githubOAuthRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestExchangeOAuthCodeRejectsUnsafeProviderResponses(t *testing.T) {
	t.Parallel()
	const (
		clientSecret = "github-client-secret"
		oauthCode    = "github-one-time-code"
	)
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       error
	}{
		{name: "redirect", statusCode: http.StatusFound, body: `{"access_token":"must-not-be-used"}`, want: ErrGitHubOAuthCodeRejected},
		{name: "provider error", statusCode: http.StatusOK, body: `{"error":"bad_verification_code","error_description":"secret provider detail"}`, want: ErrGitHubOAuthCodeRejected},
		{name: "empty token", statusCode: http.StatusOK, body: `{"access_token":""}`, want: ErrGitHubOAuthExchangeUnavailable},
		{name: "multiple values", statusCode: http.StatusOK, body: `{"access_token":"first"}{"access_token":"second"}`, want: ErrGitHubOAuthExchangeUnavailable},
		{name: "oversized body", statusCode: http.StatusOK, body: strings.Repeat("x", maxGitHubOAuthResponseBytes+1), want: ErrGitHubOAuthExchangeUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &Service{
				cfg: Config{ClientID: "github-client-id", ClientSecret: clientSecret},
				httpClient: &http.Client{Transport: githubOAuthRoundTripper(func(request *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: test.statusCode,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(test.body)),
						Request:    request,
					}, nil
				})},
			}

			token, err := service.exchangeOAuthCode(context.Background(), oauthCode)
			if token != "" || !errors.Is(err, test.want) {
				t.Fatalf("exchangeOAuthCode() = (%q, %v), want safe exchange error", token, err)
			}
			for _, secret := range []string{clientSecret, oauthCode, "secret provider detail"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("error disclosed secret %q: %v", secret, err)
				}
			}
		})
	}
}

func TestExchangeOAuthCodeRequiresBoundedCredentials(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		service *Service
		code    string
		want    error
	}{
		{service: &Service{cfg: Config{ClientSecret: "secret"}, httpClient: http.DefaultClient}, code: "code", want: ErrGitHubOAuthNotConfigured},
		{service: &Service{cfg: Config{ClientID: "client"}, httpClient: http.DefaultClient}, code: "code", want: ErrGitHubOAuthNotConfigured},
		{service: &Service{cfg: Config{ClientID: "client", ClientSecret: "secret"}, httpClient: http.DefaultClient}, want: ErrGitHubOAuthCodeInvalid},
	} {
		if _, err := test.service.exchangeOAuthCode(context.Background(), test.code); !errors.Is(err, test.want) {
			t.Fatalf("exchangeOAuthCode() error = %v", err)
		}
	}
}

func TestExchangeOAuthCodeKeepsCredentialsOutOfURL(t *testing.T) {
	t.Parallel()
	const (
		clientSecret = "github-client-secret"
		oauthCode    = "github-one-time-code"
	)
	service := &Service{
		cfg: Config{ClientID: "github-client-id", ClientSecret: clientSecret},
		httpClient: &http.Client{Transport: githubOAuthRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.URL.RawQuery != "" || strings.Contains(request.URL.String(), clientSecret) || strings.Contains(request.URL.String(), oauthCode) {
				t.Fatalf("OAuth credentials appeared in request URL: %q", request.URL.String())
			}
			if request.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Fatalf("Content-Type = %q", request.Header.Get("Content-Type"))
			}
			if err := request.ParseForm(); err != nil {
				t.Fatalf("parse OAuth form: %v", err)
			}
			if request.PostForm.Get("client_secret") != clientSecret || request.PostForm.Get("code") != oauthCode {
				t.Fatal("OAuth credentials were not sent in the form body")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"gho_returned"}`)),
				Request:    request,
			}, nil
		})},
	}

	token, err := service.exchangeOAuthCode(context.Background(), oauthCode)
	if err != nil {
		t.Fatalf("exchangeOAuthCode() error = %v", err)
	}
	if token != "gho_returned" {
		t.Fatal("exchangeOAuthCode() returned a different token")
	}
}
