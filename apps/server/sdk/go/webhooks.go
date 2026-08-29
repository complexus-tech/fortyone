package fortyone

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	webhookSecretPrefix     = "whsec_"
	webhookSignatureVersion = "v1"
	webhookSecretBytes      = 32
	maxWebhookBodyBytes     = 256 * 1024
	defaultWebhookTolerance = 5 * time.Minute
)

type WebhookVerificationErrorCode string

const (
	WebhookInvalidSecret    WebhookVerificationErrorCode = "invalid_secret"
	WebhookInvalidHeaders   WebhookVerificationErrorCode = "invalid_headers"
	WebhookInvalidTimestamp WebhookVerificationErrorCode = "invalid_timestamp"
	WebhookStaleTimestamp   WebhookVerificationErrorCode = "stale_timestamp"
	WebhookInvalidSignature WebhookVerificationErrorCode = "invalid_signature"
	WebhookInvalidBody      WebhookVerificationErrorCode = "invalid_body"
)

// WebhookVerificationError exposes only a stable classification and safe
// summary. It never retains the signing secret, signature, or request body.
type WebhookVerificationError struct {
	Code WebhookVerificationErrorCode
	msg  string
}

func (err *WebhookVerificationError) Error() string { return err.msg }

// VerifiedWebhook identifies an authenticated delivery. Callers must still
// durably deduplicate ID before applying business effects.
type VerifiedWebhook struct {
	ID        uuid.UUID
	Timestamp time.Time
}

// WebhookVerifier verifies FortyOne's exact-byte HMAC envelope before JSON is
// decoded. It is safe for concurrent use.
type WebhookVerifier struct {
	secret    []byte
	tolerance time.Duration
	now       func() time.Time
}

func (*WebhookVerifier) String() string            { return "FortyOneWebhookVerifier{secret:[REDACTED]}" }
func (verifier *WebhookVerifier) GoString() string { return verifier.String() }

// NewWebhookVerifier creates a verifier for a show-once whsec_ signing secret.
func NewWebhookVerifier(secret string, tolerance time.Duration) (*WebhookVerifier, error) {
	decoded, err := decodeSigningSecret(secret)
	if err != nil {
		return nil, err
	}
	if tolerance == 0 {
		tolerance = defaultWebhookTolerance
	}
	if tolerance < time.Second || tolerance > time.Hour {
		return nil, &WebhookVerificationError{Code: WebhookInvalidTimestamp, msg: "webhook timestamp tolerance must be from one second through one hour"}
	}
	return &WebhookVerifier{secret: decoded, tolerance: tolerance, now: time.Now}, nil
}

// Verify validates Webhook-Id, Webhook-Timestamp, and every v1 value in
// Webhook-Signature. The raw body must be passed unchanged.
func (verifier *WebhookVerifier) Verify(body []byte, header http.Header) (VerifiedWebhook, error) {
	if len(body) == 0 || len(body) > maxWebhookBodyBytes {
		return VerifiedWebhook{}, &WebhookVerificationError{Code: WebhookInvalidBody, msg: "webhook body is empty or exceeds the supported limit"}
	}
	rawDeliveryID, ok := singleHeader(header, "Webhook-Id")
	if !ok || !isCanonicalUUID(rawDeliveryID) {
		return VerifiedWebhook{}, &WebhookVerificationError{Code: WebhookInvalidHeaders, msg: "Webhook-Id is missing or malformed"}
	}
	deliveryID, _ := uuid.Parse(rawDeliveryID)
	rawTimestamp, ok := singleHeader(header, "Webhook-Timestamp")
	if !ok {
		return VerifiedWebhook{}, &WebhookVerificationError{Code: WebhookInvalidTimestamp, msg: "Webhook-Timestamp is missing or malformed"}
	}
	timestampSeconds, err := parseWebhookTimestamp(rawTimestamp)
	if err != nil {
		return VerifiedWebhook{}, err
	}
	timestamp := time.Unix(timestampSeconds, 0).UTC()
	if absoluteDuration(verifier.now().Sub(timestamp)) > verifier.tolerance {
		return VerifiedWebhook{}, &WebhookVerificationError{Code: WebhookStaleTimestamp, msg: "webhook timestamp is outside the accepted replay window"}
	}

	mac := hmac.New(sha256.New, verifier.secret)
	_, _ = fmt.Fprintf(mac, "%s.%s.", rawDeliveryID, rawTimestamp)
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	for _, headerValue := range header.Values("Webhook-Signature") {
		for _, candidate := range strings.Fields(headerValue) {
			parts := strings.Split(candidate, ",")
			if len(parts) != 2 || parts[0] != webhookSignatureVersion {
				continue
			}
			provided, decodeErr := base64.StdEncoding.Strict().DecodeString(parts[1])
			if decodeErr == nil && hmac.Equal(provided, expected) {
				return VerifiedWebhook{ID: deliveryID, Timestamp: timestamp}, nil
			}
		}
	}
	return VerifiedWebhook{}, &WebhookVerificationError{Code: WebhookInvalidSignature, msg: "webhook signature is invalid"}
}

func decodeSigningSecret(secret string) ([]byte, error) {
	if !strings.HasPrefix(secret, webhookSecretPrefix) {
		return nil, &WebhookVerificationError{Code: WebhookInvalidSecret, msg: "webhook signing secret is malformed"}
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(secret, webhookSecretPrefix))
	if err != nil || len(decoded) != webhookSecretBytes {
		return nil, &WebhookVerificationError{Code: WebhookInvalidSecret, msg: "webhook signing secret is malformed"}
	}
	return decoded, nil
}

func parseWebhookTimestamp(raw string) (int64, error) {
	if raw == "" || len(raw) > 12 {
		return 0, &WebhookVerificationError{Code: WebhookInvalidTimestamp, msg: "Webhook-Timestamp is missing or malformed"}
	}
	for _, character := range raw {
		if character < '0' || character > '9' {
			return 0, &WebhookVerificationError{Code: WebhookInvalidTimestamp, msg: "Webhook-Timestamp is missing or malformed"}
		}
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds <= 0 {
		return 0, &WebhookVerificationError{Code: WebhookInvalidTimestamp, msg: "Webhook-Timestamp is missing or malformed"}
	}
	return seconds, nil
}

func singleHeader(header http.Header, name string) (string, bool) {
	values := header.Values(name)
	returnValue := ""
	if len(values) == 1 {
		returnValue = values[0]
	}
	return returnValue, len(values) == 1 && returnValue != "" && returnValue == strings.TrimSpace(returnValue)
}

func isCanonicalUUID(raw string) bool {
	if len(raw) != 36 || raw[8] != '-' || raw[13] != '-' || raw[18] != '-' || raw[23] != '-' {
		return false
	}
	parsed, err := uuid.Parse(raw)
	return err == nil && strings.EqualFold(parsed.String(), raw)
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
