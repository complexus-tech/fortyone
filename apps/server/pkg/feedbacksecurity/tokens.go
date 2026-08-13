// Package feedbacksecurity contains deterministic, domain-separated token
// derivation shared by the API and worker processes.
package feedbacksecurity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/google/uuid"
)

const unsubscribeDomain = "fortyone:feedback-unsubscribe:v1"

// DeriveUnsubscribeToken returns an opaque token and the only representation
// persisted in Postgres: its SHA-256 hash. Because derivation is stable for a
// delivery, the outbox sweeper can safely reconstruct a stranded task payload
// without storing the raw token.
func DeriveUnsubscribeToken(authSecret string, deliveryID uuid.UUID) (string, []byte, error) {
	key, err := DeriveUnsubscribeKey(authSecret)
	if err != nil {
		return "", nil, err
	}
	return DeriveUnsubscribeTokenWithKey(key[:], deliveryID)
}

// DeriveUnsubscribeKey derives the domain-specific key once for callers that
// create many deliveries without retaining the root authentication secret.
func DeriveUnsubscribeKey(authSecret string) ([sha256.Size]byte, error) {
	authSecret = strings.TrimSpace(authSecret)
	if authSecret == "" {
		return [sha256.Size]byte{}, errors.New("feedback unsubscribe key requires an auth secret")
	}
	return sha256.Sum256([]byte(unsubscribeDomain + "\x00" + authSecret)), nil
}

// DeriveUnsubscribeTokenWithKey is the keyed form used by the feedback service.
func DeriveUnsubscribeTokenWithKey(key []byte, deliveryID uuid.UUID) (string, []byte, error) {
	if len(key) != sha256.Size || deliveryID == uuid.Nil {
		return "", nil, errors.New("feedback unsubscribe token requires an auth secret and delivery id")
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(unsubscribeDomain + "\n" + deliveryID.String()))
	token := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	hash := sha256.Sum256([]byte(token))
	return token, hash[:], nil
}
