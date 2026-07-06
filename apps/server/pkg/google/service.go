package google

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	googoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidToken  = errors.New("invalid google token")
	ErrNotConfigured = errors.New("google auth is not configured")
)

const (
	scopeOpenID                 = "openid"
	scopeEmail                  = "email"
	scopeProfile                = "profile"
	scopeCalendarFreeBusy       = "https://www.googleapis.com/auth/calendar.freebusy"
	scopeCalendarEventsFreeBusy = "https://www.googleapis.com/auth/calendar.events.freebusy"
	scopeCalendarEventsReadonly = "https://www.googleapis.com/auth/calendar.events.readonly"
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
	ClientIDs           []string
	ClientSecret        string
	RedirectURL         string
	CalendarRedirectURL string
}

// Service handles Google OAuth operations.
type Service struct {
	clientIDs     []string
	validator     *idtoken.Validator
	oauth         *oauth2.Config
	calendarOAuth *oauth2.Config
}

type CalendarToken struct {
	Token    *oauth2.Token
	Identity Identity
	Scopes   []string
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
	calendarRedirectURL := strings.TrimSpace(cfg.CalendarRedirectURL)
	if len(normalizedClientIDs) > 0 && clientSecret != "" && redirectURL != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     normalizedClientIDs[0],
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{scopeOpenID, scopeEmail, scopeProfile},
			Endpoint:     googoauth.Endpoint,
		}
	}
	var calendarOAuthConfig *oauth2.Config
	if len(normalizedClientIDs) > 0 && clientSecret != "" && calendarRedirectURL != "" {
		calendarOAuthConfig = &oauth2.Config{
			ClientID:     normalizedClientIDs[0],
			ClientSecret: clientSecret,
			RedirectURL:  calendarRedirectURL,
			Scopes: []string{
				scopeOpenID,
				scopeEmail,
				scopeProfile,
				scopeCalendarFreeBusy,
				scopeCalendarEventsFreeBusy,
				scopeCalendarEventsReadonly,
			},
			Endpoint: googoauth.Endpoint,
		}
	}

	return &Service{
		clientIDs:     normalizedClientIDs,
		validator:     validator,
		oauth:         oauthConfig,
		calendarOAuth: calendarOAuthConfig,
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

// CalendarAuthCodeURL builds an offline Google OAuth URL for Calendar free/busy access.
func (s *Service) CalendarAuthCodeURL(state string) (string, error) {
	state = strings.TrimSpace(state)
	if state == "" {
		return "", ErrInvalidToken
	}
	if s.calendarOAuth == nil {
		return "", ErrNotConfigured
	}

	return s.calendarOAuth.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	), nil
}

// ExchangeCalendarCode exchanges an OAuth code for Calendar access and verifies the returned identity token.
func (s *Service) ExchangeCalendarCode(ctx context.Context, code string) (CalendarToken, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return CalendarToken{}, ErrInvalidToken
	}
	if s.calendarOAuth == nil {
		return CalendarToken{}, ErrNotConfigured
	}

	token, err := s.calendarOAuth.Exchange(ctx, code)
	if err != nil {
		return CalendarToken{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	identity, err := s.calendarIdentityFromToken(ctx, token)
	if err != nil {
		return CalendarToken{}, err
	}
	scopes := calendarTokenScopes(token, s.calendarOAuth.Scopes)
	if !hasAnyScope(scopes, scopeCalendarFreeBusy, scopeCalendarEventsFreeBusy, scopeCalendarEventsReadonly) {
		return CalendarToken{}, fmt.Errorf(
			"%w: google did not grant a calendar availability scope",
			ErrInvalidToken,
		)
	}
	return CalendarToken{
		Token:    token,
		Identity: identity,
		Scopes:   scopes,
	}, nil
}

func (s *Service) calendarIdentityFromToken(ctx context.Context, token *oauth2.Token) (Identity, error) {
	if token == nil {
		return Identity{}, ErrInvalidToken
	}
	idToken, _ := token.Extra("id_token").(string)
	if strings.TrimSpace(idToken) == "" {
		return Identity{}, nil
	}
	return s.VerifyToken(ctx, idToken)
}

// CalendarHTTPClient returns an OAuth HTTP client for Google Calendar calls.
func (s *Service) CalendarHTTPClient(ctx context.Context, token *oauth2.Token) (*http.Client, error) {
	if token == nil {
		return nil, ErrInvalidToken
	}
	if s.calendarOAuth == nil {
		return nil, ErrNotConfigured
	}
	return s.calendarOAuth.Client(ctx, token), nil
}

func calendarTokenScopes(token *oauth2.Token, fallback []string) []string {
	if token == nil {
		return append([]string{}, fallback...)
	}
	rawScopes, _ := token.Extra("scope").(string)
	if strings.TrimSpace(rawScopes) == "" {
		return append([]string{}, fallback...)
	}
	parts := strings.Fields(rawScopes)
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			scopes = append(scopes, part)
		}
	}
	return scopes
}

func hasAnyScope(scopes []string, required ...string) bool {
	if len(scopes) == 0 || len(required) == 0 {
		return false
	}
	granted := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		granted[strings.TrimSpace(scope)] = struct{}{}
	}
	for _, scope := range required {
		if _, ok := granted[strings.TrimSpace(scope)]; ok {
			return true
		}
	}
	return false
}
