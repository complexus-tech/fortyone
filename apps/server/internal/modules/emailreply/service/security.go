package emailreply

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/secretbox"
)

const (
	// WebhookTokenHeader is configured as a custom header on the Brevo inbound
	// webhook. The value is derived from APP_EMAIL_REPLY_SECURITY_KEY, so the
	// security key itself is never disclosed to Brevo.
	WebhookTokenHeader = "X-FortyOne-Webhook-Token" // #nosec G101 -- HTTP header name, not a token value.

	webhookTokenPurpose = "fortyone:brevo:inbound-email-webhook:v1" // #nosec G101 -- HMAC domain separator, not a credential.
	payloadKeyPurpose   = "fortyone:brevo:inbound-email-payload:v1"
)

// DeriveWebhookToken returns the bearer capability configured on Brevo's
// inbound webhook. It is deterministic to make configuration reproducible and
// domain-separated from payload encryption under the same dedicated ingress
// key.
func DeriveWebhookToken(secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("email reply webhook secret is required")
	}
	digest := derivePurposeKey(secret, webhookTokenPurpose)
	return base64.RawURLEncoding.EncodeToString(digest), nil
}

func VerifyWebhookToken(secret, provided string) bool {
	expected, err := DeriveWebhookToken(secret)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(provided)))
}

// PayloadCodec encrypts inbound email payloads before they enter the durable
// messaging inbox. The derived key prevents ciphertext from being opened by a
// codec serving another capability even if a caller accidentally crosses the
// two email-reply key purposes.
type PayloadCodec struct {
	box *secretbox.Box
}

func NewPayloadCodec(secret string) (*PayloadCodec, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, errors.New("email reply payload secret is required")
	}
	key := base64.RawURLEncoding.EncodeToString(derivePurposeKey(secret, payloadKeyPurpose))
	box, err := secretbox.New(key)
	if err != nil {
		return nil, fmt.Errorf("configure email reply payload encryption: %w", err)
	}
	return &PayloadCodec{box: box}, nil
}

func (c *PayloadCodec) Seal(payload []byte) (string, error) {
	if c == nil || c.box == nil {
		return "", errors.New("email reply payload encryption is not configured")
	}
	if len(payload) == 0 {
		return "", errors.New("email reply payload is empty")
	}
	value, err := c.box.Seal(payload)
	if err != nil {
		return "", fmt.Errorf("encrypt email reply payload: %w", err)
	}
	return value, nil
}

func (c *PayloadCodec) Open(value string) ([]byte, error) {
	if c == nil || c.box == nil {
		return nil, errors.New("email reply payload encryption is not configured")
	}
	opened, err := c.box.Open(value)
	if err != nil {
		return nil, fmt.Errorf("decrypt email reply payload: %w", err)
	}
	if opened.Version == 0 {
		return nil, errors.New("unencrypted email reply payload is not allowed")
	}
	return opened.Plaintext, nil
}

func derivePurposeKey(secret, purpose string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(purpose))
	return mac.Sum(nil)
}
