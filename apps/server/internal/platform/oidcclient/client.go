package oidcclient

import (
	"context"
	"crypto/subtle"
	"errors"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var ErrProvider = errors.New("identity provider sign-in failed")

type Credential struct{ Issuer, ClientID, ClientSecret string }
type Attempt struct {
	Nonce, Verifier string
	CreatedAt       time.Time
}
type Identity struct {
	Subject, Email  string
	AuthenticatedAt time.Time
}

// Client returns verified identity facts only. Provider tokens and error bodies
// are never returned to the application, persisted, or written to logs.
type Client struct {
	http     *http.Client
	validate func(context.Context, string) error
}

func New() *Client {
	return &Client{http: publicHTTPClient(), validate: func(ctx context.Context, url string) error {
		_, err := safehttp.Resolve(ctx, net.DefaultResolver, url)
		return err
	}}
}
func (c *Client) discover(ctx context.Context, credential Credential, redirect string) (*oidc.Provider, *oauth2.Config, error) {
	if err := c.validate(ctx, credential.Issuer); err != nil {
		return nil, nil, ErrProvider
	}
	ctx = oidc.ClientContext(ctx, c.http)
	provider, err := oidc.NewProvider(ctx, credential.Issuer)
	if err != nil {
		return nil, nil, ErrProvider
	}
	endpoint := provider.Endpoint()
	for _, url := range []string{endpoint.AuthURL, endpoint.TokenURL} {
		if err := c.validate(ctx, url); err != nil {
			return nil, nil, ErrProvider
		}
	}
	return provider, &oauth2.Config{ClientID: credential.ClientID, ClientSecret: credential.ClientSecret, RedirectURL: redirect, Endpoint: endpoint, Scopes: []string{oidc.ScopeOpenID, "email", "profile"}}, nil
}
func (c *Client) Validate(ctx context.Context, credential Credential, redirect string) error {
	_, _, err := c.discover(ctx, credential, redirect)
	return err
}
func (c *Client) AuthorizationURL(ctx context.Context, credential Credential, attempt Attempt, state, redirect string) (string, error) {
	_, config, err := c.discover(ctx, credential, redirect)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL(state, oidc.Nonce(attempt.Nonce), oauth2.S256ChallengeOption(attempt.Verifier), oauth2.SetAuthURLParam("prompt", "login"), oauth2.SetAuthURLParam("max_age", "0")), nil
}
func (c *Client) Exchange(ctx context.Context, credential Credential, attempt Attempt, code, redirect string) (Identity, error) {
	provider, config, err := c.discover(ctx, credential, redirect)
	if err != nil {
		return Identity{}, err
	}
	ctx = oidc.ClientContext(ctx, c.http)
	token, err := config.Exchange(ctx, code, oauth2.VerifierOption(attempt.Verifier))
	if err != nil {
		return Identity{}, ErrProvider
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		return Identity{}, ErrProvider
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: credential.ClientID, SupportedSigningAlgs: []string{oidc.RS256, oidc.ES256}}).Verify(ctx, raw)
	if err != nil {
		return Identity{}, ErrProvider
	}
	if subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(attempt.Nonce)) != 1 {
		return Identity{}, ErrProvider
	}
	if idToken.AccessTokenHash != "" {
		if err := idToken.VerifyAccessToken(token.AccessToken); err != nil {
			return Identity{}, ErrProvider
		}
	}
	var claims struct {
		Email           string `json:"email"`
		EmailVerified   bool   `json:"email_verified"`
		AuthenticatedAt int64  `json:"auth_time"`
		IssuedAt        int64  `json:"iat"`
		AuthorizedParty string `json:"azp"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return Identity{}, ErrProvider
	}
	now := time.Now()
	authenticated := time.Unix(claims.AuthenticatedAt, 0).UTC()
	if !claims.EmailVerified || idToken.Subject == "" || len(idToken.Subject) > 512 || claims.AuthenticatedAt <= 0 || authenticated.Before(attempt.CreatedAt.Add(-time.Minute)) || authenticated.After(now.Add(30*time.Second)) || claims.IssuedAt <= 0 || time.Unix(claims.IssuedAt, 0).After(now.Add(30*time.Second)) || ((len(idToken.Audience) > 1 || claims.AuthorizedParty != "") && claims.AuthorizedParty != credential.ClientID) {
		return Identity{}, ErrProvider
	}
	email := strings.TrimSpace(claims.Email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 320 {
		return Identity{}, ErrProvider
	}
	return Identity{Subject: idToken.Subject, Email: email, AuthenticatedAt: authenticated}, nil
}
