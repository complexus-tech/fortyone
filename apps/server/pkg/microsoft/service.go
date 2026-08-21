package microsoft

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidToken  = errors.New("invalid microsoft token")
	ErrNotConfigured = errors.New("microsoft auth is not configured")
)

const (
	defaultTenant      = "common"
	defaultAuthority   = "https://login.microsoftonline.com"
	defaultGraphURL    = "https://graph.microsoft.com/v1.0/me"
	defaultJWKSURL     = "https://login.microsoftonline.com/common/discovery/v2.0/keys"
	defaultHTTPTimeout = 10 * time.Second
	defaultJWKSMaxAge  = time.Hour
	personalTenantID   = "9188040d-6c67-4c5b-b112-36a304b66dad"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Tenant       string
	RedirectURL  string
	HTTPClient   *http.Client
	AuthURL      string
	TokenURL     string
	ProfileURL   string
	JWKSURL      string
}

type Identity struct {
	Issuer            string
	TenantID          string
	ObjectID          string
	Subject           string
	Email             string
	FirstName         string
	LastName          string
	FullName          string
	PreferredUsername string
}

type Service struct {
	clientID   string
	oauth      *oauth2.Config
	httpClient *http.Client
	profileURL string
	jwksURL    string

	keysMu       sync.RWMutex
	keys         map[string]*rsa.PublicKey
	keysExpireAt time.Time
	now          func() time.Time
}

type idTokenClaims struct {
	jwt.RegisteredClaims
	Nonce             string `json:"nonce"`
	TenantID          string `json:"tid"`
	ObjectID          string `json:"oid"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

type graphProfile struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyType string `json:"kty"`
	KeyID   string `json:"kid"`
	Use     string `json:"use"`
	N       string `json:"n"`
	E       string `json:"e"`
}

func NewService(cfg Config) *Service {
	clientID := strings.TrimSpace(cfg.ClientID)
	clientSecret := strings.TrimSpace(cfg.ClientSecret)
	redirectURL := strings.TrimSpace(cfg.RedirectURL)
	tenant := strings.TrimSpace(cfg.Tenant)
	if tenant == "" {
		tenant = defaultTenant
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	authURL := strings.TrimSpace(cfg.AuthURL)
	if authURL == "" {
		authURL = fmt.Sprintf("%s/%s/oauth2/v2.0/authorize", defaultAuthority, url.PathEscape(tenant))
	}
	tokenURL := strings.TrimSpace(cfg.TokenURL)
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("%s/%s/oauth2/v2.0/token", defaultAuthority, url.PathEscape(tenant))
	}

	var oauthConfig *oauth2.Config
	if clientID != "" && clientSecret != "" && redirectURL != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "profile", "email", "User.Read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		}
	}

	profileURL := strings.TrimSpace(cfg.ProfileURL)
	if profileURL == "" {
		profileURL = defaultGraphURL
	}
	jwksURL := strings.TrimSpace(cfg.JWKSURL)
	if jwksURL == "" {
		jwksURL = defaultJWKSURL
	}

	return &Service{
		clientID:   clientID,
		oauth:      oauthConfig,
		httpClient: httpClient,
		profileURL: profileURL,
		jwksURL:    jwksURL,
		keys:       make(map[string]*rsa.PublicKey),
		now:        time.Now,
	}
}

func (s *Service) AuthCodeURL(state, nonce, verifier string) (string, error) {
	if s.oauth == nil {
		return "", ErrNotConfigured
	}
	if strings.TrimSpace(state) == "" || strings.TrimSpace(nonce) == "" || strings.TrimSpace(verifier) == "" {
		return "", ErrInvalidToken
	}

	return s.oauth.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
		oauth2.SetAuthURLParam("response_mode", "query"),
	), nil
}

func (s *Service) ExchangeCode(ctx context.Context, code, verifier, nonce string) (Identity, error) {
	if s.oauth == nil {
		return Identity{}, ErrNotConfigured
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(verifier) == "" || strings.TrimSpace(nonce) == "" {
		return Identity{}, ErrInvalidToken
	}

	oauthContext := context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)
	token, err := s.oauth.Exchange(oauthContext, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return Identity{}, fmt.Errorf("%w: exchange authorization code: %v", ErrInvalidToken, err)
	}
	idToken, _ := token.Extra("id_token").(string)
	identity, err := s.verifyIDToken(ctx, idToken, nonce)
	if err != nil {
		return Identity{}, err
	}

	profile, err := s.fetchProfile(ctx, token.AccessToken)
	if err != nil {
		return Identity{}, err
	}
	// Personal Microsoft accounts can expose a legacy Microsoft Graph profile
	// identifier that differs from the signed oid claim. Keep tid + oid as the
	// canonical identity for those accounts; Entra tenant profiles must match.
	if !graphProfileMatchesIdentity(profile.ID, identity) {
		return Identity{}, fmt.Errorf("%w: graph profile does not match token subject", ErrInvalidToken)
	}

	identity.Email = firstNonEmpty(identity.Email, profile.Mail, profile.UserPrincipalName)
	identity.FullName = firstNonEmpty(identity.FullName, profile.DisplayName)
	identity.FirstName = firstNonEmpty(identity.FirstName, profile.GivenName)
	identity.LastName = firstNonEmpty(identity.LastName, profile.Surname)
	return identity, nil
}

func (s *Service) verifyIDToken(ctx context.Context, rawToken, nonce string) (Identity, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Identity{}, fmt.Errorf("%w: missing id token", ErrInvalidToken)
	}

	claims := &idTokenClaims{}
	parsed, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			keyID, _ := token.Header["kid"].(string)
			if keyID == "" {
				return nil, fmt.Errorf("%w: missing signing key id", ErrInvalidToken)
			}
			return s.signingKey(ctx, keyID)
		},
		jwt.WithAudience(s.clientID),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil || !parsed.Valid {
		return Identity{}, fmt.Errorf("%w: verify id token: %v", ErrInvalidToken, err)
	}

	tenantID := strings.TrimSpace(claims.TenantID)
	objectID := strings.TrimSpace(claims.ObjectID)
	if _, err := uuid.Parse(tenantID); err != nil {
		return Identity{}, fmt.Errorf("%w: invalid tenant id", ErrInvalidToken)
	}
	if _, err := uuid.Parse(objectID); err != nil {
		return Identity{}, fmt.Errorf("%w: invalid object id", ErrInvalidToken)
	}
	if !strings.EqualFold(strings.TrimSpace(claims.Nonce), strings.TrimSpace(nonce)) {
		return Identity{}, fmt.Errorf("%w: nonce mismatch", ErrInvalidToken)
	}

	expectedIssuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID)
	if !strings.EqualFold(strings.TrimSuffix(claims.Issuer, "/"), expectedIssuer) {
		return Identity{}, fmt.Errorf("%w: issuer mismatch", ErrInvalidToken)
	}

	return Identity{
		Issuer:            expectedIssuer,
		TenantID:          tenantID,
		ObjectID:          objectID,
		Subject:           strings.TrimSpace(claims.Subject),
		Email:             strings.TrimSpace(claims.Email),
		FirstName:         strings.TrimSpace(claims.GivenName),
		LastName:          strings.TrimSpace(claims.FamilyName),
		FullName:          strings.TrimSpace(claims.Name),
		PreferredUsername: strings.TrimSpace(claims.PreferredUsername),
	}, nil
}

func (s *Service) fetchProfile(ctx context.Context, accessToken string) (graphProfile, error) {
	if strings.TrimSpace(accessToken) == "" {
		return graphProfile{}, fmt.Errorf("%w: missing access token", ErrInvalidToken)
	}

	profileURL, err := url.Parse(s.profileURL)
	if err != nil {
		return graphProfile{}, fmt.Errorf("invalid microsoft profile url: %w", err)
	}
	query := profileURL.Query()
	query.Set("$select", "id,displayName,givenName,surname,mail,userPrincipalName")
	profileURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL.String(), nil)
	if err != nil {
		return graphProfile{}, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return graphProfile{}, fmt.Errorf("fetch microsoft profile: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return graphProfile{}, fmt.Errorf("%w: microsoft profile returned status %d", ErrInvalidToken, response.StatusCode)
	}

	var profile graphProfile
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return graphProfile{}, fmt.Errorf("decode microsoft profile: %w", err)
	}
	if strings.TrimSpace(profile.ID) == "" {
		return graphProfile{}, fmt.Errorf("%w: microsoft profile is missing id", ErrInvalidToken)
	}
	return profile, nil
}

func (s *Service) signingKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	s.keysMu.RLock()
	key := s.keys[keyID]
	fresh := s.now().Before(s.keysExpireAt)
	s.keysMu.RUnlock()
	if key != nil && fresh {
		return key, nil
	}

	if err := s.refreshSigningKeys(ctx); err != nil {
		return nil, err
	}
	s.keysMu.RLock()
	defer s.keysMu.RUnlock()
	key = s.keys[keyID]
	if key == nil {
		return nil, fmt.Errorf("%w: unknown signing key", ErrInvalidToken)
	}
	return key, nil
}

func (s *Service) refreshSigningKeys(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.jwksURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("fetch microsoft signing keys: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch microsoft signing keys: status %d", response.StatusCode)
	}

	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return fmt.Errorf("decode microsoft signing keys: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, value := range document.Keys {
		if value.KeyType != "RSA" || value.KeyID == "" || value.N == "" || value.E == "" {
			continue
		}
		key, err := rsaKey(value.N, value.E)
		if err != nil {
			continue
		}
		keys[value.KeyID] = key
	}
	if len(keys) == 0 {
		return errors.New("microsoft signing keys response contained no usable keys")
	}

	maxAge := cacheMaxAge(response.Header.Get("Cache-Control"), defaultJWKSMaxAge)
	s.keysMu.Lock()
	s.keys = keys
	s.keysExpireAt = s.now().Add(maxAge)
	s.keysMu.Unlock()
	return nil
}

func rsaKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(modulus)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(exponent)
	if err != nil {
		return nil, err
	}
	if len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, errors.New("invalid rsa exponent")
	}
	e := 0
	for _, value := range eBytes {
		e = (e << 8) + int(value)
	}
	if e < 2 {
		return nil, errors.New("invalid rsa exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func cacheMaxAge(value string, fallback time.Duration) time.Duration {
	for _, directive := range strings.Split(value, ",") {
		key, raw, ok := strings.Cut(strings.TrimSpace(directive), "=")
		if !ok || !strings.EqualFold(key, "max-age") {
			continue
		}
		seconds, err := strconv.Atoi(strings.TrimSpace(raw))
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isPersonalTenant(tenantID string) bool {
	return strings.EqualFold(strings.TrimSpace(tenantID), personalTenantID)
}

func graphProfileMatchesIdentity(profileID string, identity Identity) bool {
	return isPersonalTenant(identity.TenantID) ||
		strings.EqualFold(strings.TrimSpace(profileID), strings.TrimSpace(identity.ObjectID))
}
