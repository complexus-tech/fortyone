package calendar

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/workspaceurl"
	"github.com/google/uuid"
)

func (s *Service) provider(provider Provider) (CalendarProvider, error) {
	if s.cfg.Providers == nil {
		return nil, ErrCalendarNotConfigured
	}
	configured := s.cfg.Providers[provider]
	if configured == nil {
		return nil, ErrCalendarNotConfigured
	}
	return configured, nil
}

func (s *Service) signState(claims stateClaims) (string, error) {
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return "", ErrCalendarNotConfigured
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (s *Service) verifyState(value string) (stateClaims, error) {
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return stateClaims{}, ErrCalendarNotConfigured
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return stateClaims{}, ErrInvalidCalendarState
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return stateClaims{}, ErrInvalidCalendarState
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return stateClaims{}, err
	}
	var claims stateClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return stateClaims{}, err
	}
	return claims, nil
}

func (s *Service) verifyActiveState(value string) (stateClaims, error) {
	claims, err := s.verifyState(strings.TrimSpace(value))
	if err != nil {
		return stateClaims{}, err
	}
	if claims.WorkspaceID == uuid.Nil ||
		claims.UserID == uuid.Nil ||
		(claims.Provider != ProviderGoogle && claims.Provider != ProviderMicrosoft) ||
		strings.TrimSpace(claims.WorkspaceSlug) == "" {
		return stateClaims{}, ErrInvalidCalendarState
	}
	if !s.now().Before(time.Unix(claims.ExpiresAt, 0)) {
		return stateClaims{}, fmt.Errorf("%w: expired", ErrInvalidCalendarState)
	}
	return claims, nil
}

func (s *Service) encryptTokenPayload(token ProviderToken) (string, error) {
	key := s.encryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(randReader{s.randBytes}, nonce); err != nil {
		return "", err
	}
	plaintext, err := json.Marshal(token) // #nosec G117 -- immediately AEAD-sealed below.
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func (s *Service) decryptTokenPayload(value string) (ProviderToken, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return ProviderToken{}, err
	}
	key := s.encryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return ProviderToken{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ProviderToken{}, err
	}
	if len(raw) < gcm.NonceSize() {
		return ProviderToken{}, errors.New("calendar token payload is too short")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ProviderToken{}, err
	}
	var token ProviderToken
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return ProviderToken{}, err
	}
	return token, nil
}

func (s *Service) tokenForConnection(ctx context.Context, connection CoreConnection, provider CalendarProvider) (ProviderToken, error) {
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return ProviderToken{}, err
	}
	refresher, ok := provider.(CalendarTokenRefresher)
	if !ok {
		return token, nil
	}
	refreshed, err := refresher.RefreshToken(ctx, token)
	if err != nil {
		return ProviderToken{}, err
	}
	if refreshed.AccessToken == token.AccessToken &&
		refreshed.RefreshToken == token.RefreshToken &&
		refreshed.TokenType == token.TokenType &&
		refreshed.Expiry.Equal(token.Expiry) {
		return refreshed, nil
	}
	payload, err := s.encryptTokenPayload(refreshed)
	if err != nil {
		return ProviderToken{}, err
	}
	if err := s.repo.UpdateConnectionToken(ctx, connection, payload); err != nil {
		return ProviderToken{}, err
	}
	return refreshed, nil
}

func (s *Service) encryptionKey() [32]byte {
	return sha256.Sum256([]byte(s.cfg.SecretKey))
}

func (s *Service) workspaceCalendarURL(workspaceSlug, query string) string {
	path := workspaceurl.Build(
		s.cfg.WebsiteURL,
		workspaceSlug,
		"settings",
		"account",
		"calendar",
	)
	if path == "" {
		return "/"
	}
	if strings.TrimSpace(query) == "" {
		return path
	}
	return path + "?" + query
}

func fallbackTimezone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UTC"
	}
	return value
}
