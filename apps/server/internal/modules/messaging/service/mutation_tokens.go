package messaging

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

func (m *storyMutationExecutor) signClaims(claims storyMutationClaims) (string, error) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(claims); err != nil {
		return "", fmt.Errorf("encode story mutation confirmation: %w", err)
	}
	payload := bytes.TrimSuffix(encoded.Bytes(), []byte("\n"))
	signature := hmac.New(sha256.New, m.key)
	_, _ = signature.Write(payload)
	token := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))
	if len(token) > maximumStoryMutationTokenSize {
		return "", fmt.Errorf("story mutation confirmation exceeds %d characters", maximumStoryMutationTokenSize)
	}
	return token, nil
}

func (m *storyMutationExecutor) newBatchToken(confirmationID uuid.UUID) (string, error) {
	payload := make([]byte, batchStoryTokenBytes)
	copy(payload, confirmationID[:])
	if _, err := io.ReadFull(m.random, payload[len(confirmationID):]); err != nil {
		return "", fmt.Errorf("generate opaque batch confirmation token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func batchConfirmationID(token string) (uuid.UUID, bool, error) {
	token = strings.TrimSpace(token)
	if strings.Contains(token, ".") {
		return uuid.Nil, false, nil
	}
	if token == "" || len(token) > maximumStoryMutationTokenSize {
		return uuid.Nil, true, fmt.Errorf("%w: token is missing or too large", ErrInvalidConfirmation)
	}
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(payload) != batchStoryTokenBytes || base64.RawURLEncoding.EncodeToString(payload) != token {
		return uuid.Nil, true, fmt.Errorf("%w: opaque token format", ErrInvalidConfirmation)
	}
	confirmationID, err := uuid.FromBytes(payload[:16])
	if err != nil || confirmationID == uuid.Nil {
		return uuid.Nil, true, fmt.Errorf("%w: opaque token identifier", ErrInvalidConfirmation)
	}
	return confirmationID, true, nil
}

func (m *storyMutationExecutor) verifyClaims(token string) (storyMutationClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > maximumStoryMutationTokenSize {
		return storyMutationClaims{}, fmt.Errorf("%w: token is missing or too large", ErrInvalidConfirmation)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return storyMutationClaims{}, fmt.Errorf("%w: token format", ErrInvalidConfirmation)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || base64.RawURLEncoding.EncodeToString(payload) != parts[0] {
		return storyMutationClaims{}, fmt.Errorf("%w: token payload", ErrInvalidConfirmation)
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || base64.RawURLEncoding.EncodeToString(providedSignature) != parts[1] {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}
	expectedSignature := hmac.New(sha256.New, m.key)
	_, _ = expectedSignature.Write(payload)
	if !hmac.Equal(providedSignature, expectedSignature.Sum(nil)) {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}

	var claims storyMutationClaims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return storyMutationClaims{}, fmt.Errorf("%w: token claims", ErrInvalidConfirmation)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return storyMutationClaims{}, fmt.Errorf("%w: trailing token claims", ErrInvalidConfirmation)
	}
	if claims.Version != storyMutationTokenVersion || claims.ConfirmationID == uuid.Nil || claims.WorkspaceID == uuid.Nil || claims.UserID == uuid.Nil || claims.TeamID == uuid.Nil || claims.ExpiresAt.IsZero() {
		return storyMutationClaims{}, fmt.Errorf("%w: incomplete token claims", ErrInvalidConfirmation)
	}
	normalizeLegacyUpdateTimeActions(&claims)
	if err := validateStoryMutationClaims(claims); err != nil {
		return storyMutationClaims{}, err
	}
	return claims, nil
}

func storyMutationConfirmationBinding(claims storyMutationClaims, token string) StoryMutationConfirmationBinding {
	return StoryMutationConfirmationBinding{
		ConfirmationID: claims.ConfirmationID,
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		TokenHash:      storyMutationTokenHash(token),
	}
}

func storyMutationTokenHash(token string) []byte {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return append([]byte(nil), digest[:]...)
}
