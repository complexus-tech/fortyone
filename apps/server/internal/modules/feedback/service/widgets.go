package feedback

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *Service) GetWidgetSettings(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSettings, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CoreWidgetSettings{}, invalidInput("workspace and portal ids are required")
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreWidgetSettings{}, err
		}
		return s.scopedNextRepo.GetWidgetSettingsScoped(ctx, scope, portalID)
	}
	return s.nextRepo.GetWidgetSettings(ctx, workspaceID, portalID)
}

func (s *Service) GetPublicWidgetSettings(ctx context.Context, portalSlug string) (CoreWidgetSettings, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreWidgetSettings{}, err
	}
	settings, err := s.nextRepo.GetPublicWidgetSettings(ctx, portal.ID)
	if errors.Is(err, ErrNotFound) {
		return CoreWidgetSettings{PortalID: portal.ID, AllowedOrigins: []string{ // GetPublicWidgetSettings returns only the non-secret configuration that an
			// embed needs before it renders. Portals without settings are safely disabled.
		}}, nil
	}
	return settings, err
}
func (s *Service) UpdateWidgetSettings(ctx context.Context, input CoreWidgetSettingsInput) (CoreWidgetSettings, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil {
		return CoreWidgetSettings{}, invalidInput("workspace and portal ids are required")
	}
	origins, err := normalizeAllowedOrigins(input.AllowedOrigins)
	if err != nil {
		return CoreWidgetSettings{}, err
	}
	input.AllowedOrigins = origins
	if input.Enabled {
		if len(origins) == 0 {
			return CoreWidgetSettings{}, invalidInput("at least one allowed origin is required before enabling the widget")
		}
	}
	if s.scopedNextRepo != nil {
		scope, err := accessScopeFromContext(ctx, input.WorkspaceID, uuid.Nil)
		if err != nil {
			return CoreWidgetSettings{}, err
		}
		input.Access = scope
		return s.scopedNextRepo.UpsertWidgetSettingsScoped(ctx, scope, input)
	}
	return s.nextRepo.UpsertWidgetSettings(ctx, input)
}
func (s *Service) CreateWidgetSigningSecret(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSecretResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreWidgetSecretResult{}, err
	}
	secret, _, err := s.randomToken(32)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	keyID := uuid.New()
	encrypted, err := s.sealWidgetSecret(portalID, 1, secret)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	var settings CoreWidgetSettings
	if s.scopedNextRepo != nil {
		scope, scopeErr := accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if scopeErr != nil {
			return CoreWidgetSecretResult{}, scopeErr
		}
		settings, err = s.scopedNextRepo.SetInitialWidgetSecretScoped(ctx, scope, portalID, keyID, encrypted, 1)
	} else {
		settings, err = s.nextRepo.SetInitialWidgetSecret(ctx, workspaceID, portalID, keyID, encrypted, 1)
	}
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	return CoreWidgetSecretResult{Settings: settings, SigningSecret: secret}, nil
}
func (s *Service) RotateWidgetSigningSecret(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSecretResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreWidgetSecretResult{}, err
	}
	var scope CoreAccessScope
	var current CoreWidgetSettings
	var err error
	if s.scopedNextRepo != nil {
		scope, err = accessScopeFromContext(ctx, workspaceID, uuid.Nil)
		if err != nil {
			return CoreWidgetSecretResult{}, err
		}
		current, err = s.scopedNextRepo.GetWidgetSettingsScoped(ctx, scope, portalID)
	} else {
		current, err = s.nextRepo.GetWidgetSettings(ctx, workspaceID, portalID)
	}
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	if strings.TrimSpace(current.SigningSecretEncrypted) == "" {
		return CoreWidgetSecretResult{}, invalidInput("create a widget signing secret before rotating it")
	}
	secret, _, err := s.randomToken(32)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	version := current.SigningSecretVersion + 1
	encrypted, err := s.sealWidgetSecret(portalID, version, secret)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	var settings CoreWidgetSettings
	if s.scopedNextRepo != nil {
		settings, err = s.scopedNextRepo.RotateWidgetSecretScoped(ctx, scope, portalID, encrypted, version, s.now().UTC().Add(widgetRotationGrace))
	} else {
		settings, err = s.nextRepo.RotateWidgetSecret(ctx, workspaceID, portalID, encrypted, version, s.now().UTC().Add(widgetRotationGrace))
	}
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	return CoreWidgetSecretResult{Settings: settings, SigningSecret: secret}, nil
}
func (s *Service) CreateWidgetContributorSession(ctx context.Context, input CoreWidgetSessionInput) (CoreContributorSessionResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreContributorSessionResult{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	settings, err := s.nextRepo.GetPublicWidgetSettings(ctx, portal.ID)
	if err != nil || !settings.Enabled {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	parentOrigin, err := normalizeOrigin(input.ParentOrigin)
	if err != nil || !slices.Contains(settings.AllowedOrigins, parentOrigin) {
		return CoreContributorSessionResult{}, ErrWidgetOriginNotAllowed
	}
	assertion, signedPart, signature, err := parseWidgetAssertion(input.Assertion)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	now := s.now().UTC()
	issuedAt := time.Unix(assertion.IssuedAt, 0).UTC()
	expiresAt := time.Unix(assertion.ExpiresAt, 0).UTC()
	assertionKeyID, keyErr := uuid.Parse(assertion.KeyID)
	if keyErr != nil || assertion.Version < 1 || assertionKeyID != settings.WidgetKeyID || assertion.Origin != parentOrigin || assertion.ExternalID == "" || assertion.Nonce == "" {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	if issuedAt.After(now.Add(widgetClockSkew)) || !expiresAt.After(now) || !expiresAt.After(issuedAt) || expiresAt.Sub(issuedAt) > widgetAssertionMaxLifetime {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	normalizedEmail, err := validate.Email(assertion.Email)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	assertion.ExternalID = strings.TrimSpace(assertion.ExternalID)
	assertion.DisplayName = strings.TrimSpace(assertion.DisplayName)
	if assertion.ExternalID == "" || assertion.DisplayName == "" || utf8.RuneCountInString(assertion.DisplayName) > maxDisplayNameRunes || len(assertion.ExternalID) > 255 || len(assertion.Nonce) > 255 {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	var avatarURL *string
	if strings.TrimSpace(assertion.AvatarURL) != "" {
		parsed, parseErr := url.ParseRequestURI(strings.TrimSpace(assertion.AvatarURL))
		if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
		}
		value := parsed.String()
		avatarURL = &value
	}
	encryptedSecret, err := s.nextRepo.GetWidgetSigningSecret(ctx, portal.ID, assertionKeyID, assertion.Version)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	secret, err := s.openWidgetSecret(portal.ID, assertion.Version, encryptedSecret)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signedPart))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	token, tokenHash, err := s.randomToken(32)
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	if err := s.nextRepo.ConsumeWidgetAssertionNonce(ctx, portal.ID, assertionKeyID, assertion.Version, assertion.Nonce, parentOrigin, expiresAt); err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.CreateExternalContributorSession(ctx, portal.ID, assertion.ExternalID, normalizedEmail, assertion.DisplayName, avatarURL, tokenHash, now.Add(widgetSessionLifetime))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session, Token: token}, nil
}
func (s *Service) sealWidgetSecret(portalID uuid.UUID, version int, plaintext string) (string, error) {
	nonce := make([]byte, s.security.widgetAEAD.NonceSize())
	if _, err := io.ReadFull(s.random, nonce); err != nil {
		return "", fmt.Errorf("generate feedback widget encryption nonce: %w", err)
	}
	aad := s.widgetSecretAAD(portalID, version)
	ciphertext := s.security.widgetAEAD.Seal(nil, nonce, []byte(plaintext), aad)
	return "v1." + base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}
func (s *Service) openWidgetSecret(portalID uuid.UUID, version int, envelope string) (string, error) {
	if !strings.HasPrefix(envelope, "v1.") {
		return "", errors.New("unsupported feedback widget signing envelope")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(envelope, "v1."))
	if err != nil || len(raw) <= s.security.widgetAEAD.NonceSize() {
		return "", errors.New("invalid feedback widget signing envelope")
	}
	nonceSize := s.security.widgetAEAD.NonceSize()
	plaintext, err := s.security.widgetAEAD.Open(nil, raw[:nonceSize], raw[nonceSize:], s.widgetSecretAAD(portalID, version))
	if err != nil {
		return "", fmt.Errorf("decrypt feedback widget signing secret: %w", err)
	}
	return string(plaintext), nil
}
func (s *Service) widgetSecretAAD(portalID uuid.UUID, version int) []byte {
	return []byte(string(s.security.widgetAADBase) + "\n" + portalID.String() + "\n" + strconv.Itoa(version))
}
func normalizeAllowedOrigins(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		origin, err := normalizeOrigin(value)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	slices.Sort(result)
	return result, nil
}
func normalizeOrigin(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", invalidInput("widget origins must be exact origins without a path, query, or fragment")
	}
	hostname := strings.ToLower(parsed.Hostname())
	local := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && local) {
		return "", invalidInput("widget origins must use HTTPS except for localhost development")
	}
	if strings.Contains(parsed.Host, "*") {
		return "", invalidInput("widget origin wildcards are not supported")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = ""
	return parsed.String(), nil
}
func parseWidgetAssertion(value string) (CoreWidgetIdentityAssertion, string, []byte, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) > 8<<10 {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(signature) != sha256.Size {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	var assertion CoreWidgetIdentityAssertion
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&assertion); err != nil {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	return assertion, parts[0], signature, nil
}
func encodeWidgetAssertionForTest(assertion CoreWidgetIdentityAssertion, secret string) (string, error) {
	payload, err := json.Marshal(assertion)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
