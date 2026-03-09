package google

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	googoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidToken  = errors.New("invalid google token")
	ErrNotConfigured = errors.New("google auth is not configured")
)

type Identity struct {
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	FullName      string
	Picture       string
}

type Config struct {
	ClientIDs    []string
	ClientSecret string
	RedirectURL  string
}

// Service handles Google OAuth operations.
type Service struct {
	clientIDs []string
	validator *idtoken.Validator
	oauth     *oauth2.Config
}

// NewService creates a new Google service instance.
func NewService(cfg Config) (*Service, error) {
	validator, err := idtoken.NewValidator(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	normalizedClientIDs := make([]string, 0, len(cfg.ClientIDs))
	for _, clientID := range cfg.ClientIDs {
		clientID = strings.TrimSpace(clientID)
		if clientID == "" {
			continue
		}
		normalizedClientIDs = append(normalizedClientIDs, clientID)
	}

	var oauthConfig *oauth2.Config
	clientSecret := strings.TrimSpace(cfg.ClientSecret)
	redirectURL := strings.TrimSpace(cfg.RedirectURL)
	if len(normalizedClientIDs) > 0 && clientSecret != "" && redirectURL != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     normalizedClientIDs[0],
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googoauth.Endpoint,
		}
	}

	return &Service{
		clientIDs: normalizedClientIDs,
		validator: validator,
		oauth:     oauthConfig,
	}, nil
}

// VerifyToken validates a Google ID token against configured Google client IDs.
func (s *Service) VerifyToken(ctx context.Context, token string) (Identity, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Identity{}, ErrInvalidToken
	}

	if len(s.clientIDs) == 0 {
		return Identity{}, ErrNotConfigured
	}

	var lastErr error
	for _, clientID := range s.clientIDs {
		payload, err := s.validator.Validate(ctx, token, clientID)
		if err != nil {
			lastErr = err
			continue
		}

		email, _ := payload.Claims["email"].(string)
		emailVerified, _ := payload.Claims["email_verified"].(bool)
		firstName, _ := payload.Claims["given_name"].(string)
		lastName, _ := payload.Claims["family_name"].(string)
		fullName, _ := payload.Claims["name"].(string)
		picture, _ := payload.Claims["picture"].(string)

		return Identity{
			Email:         strings.TrimSpace(email),
			EmailVerified: emailVerified,
			FirstName:     strings.TrimSpace(firstName),
			LastName:      strings.TrimSpace(lastName),
			FullName:      strings.TrimSpace(fullName),
			Picture:       strings.TrimSpace(picture),
		}, nil
	}

	return Identity{}, fmt.Errorf("%w: %v", ErrInvalidToken, lastErr)
}

// AuthCodeURL builds a Google OAuth authorization URL for backend-driven auth.
func (s *Service) AuthCodeURL(state string) (string, error) {
	state = strings.TrimSpace(state)
	if state == "" {
		return "", ErrInvalidToken
	}
	if s.oauth == nil {
		return "", ErrNotConfigured
	}

	return s.oauth.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	), nil
}

// ExchangeCode exchanges an OAuth code and verifies the returned id_token.
func (s *Service) ExchangeCode(ctx context.Context, code string) (Identity, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Identity{}, ErrInvalidToken
	}
	if s.oauth == nil {
		return Identity{}, ErrNotConfigured
	}

	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	idToken, _ := token.Extra("id_token").(string)
	if strings.TrimSpace(idToken) == "" {
		return Identity{}, ErrInvalidToken
	}

	return s.VerifyToken(ctx, idToken)
}
