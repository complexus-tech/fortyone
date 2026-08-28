package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/secretbox"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const (
	legacySlackWebhookPayloadEnvelopePrefix = "slack-webhook.v1."
	legacySlackWebhookPayloadVersion        = 1
)

// LegacyCutover is the only Slack component allowed to receive the former
// APP_AUTH_SECRET_KEY. It exists solely for bounded compare-and-swap migration
// of credentials and durable webhook receipts. Normal OAuth, API calls, and
// webhook processing neither hold nor consult this decoder.
type LegacyCutover struct {
	box *secretbox.Box
}

func NewLegacyCutover(secret string) (*LegacyCutover, error) {
	box, err := secretbox.New(secret)
	if err != nil {
		return nil, fmt.Errorf("configure Slack legacy cutover: %w", err)
	}
	return &LegacyCutover{box: box}, nil
}

func (cutover *LegacyCutover) openCredential(value string) (slackCredential, int, error) {
	if cutover == nil || cutover.box == nil {
		return slackCredential{}, 0, errors.New("slack legacy credential cutover is not configured")
	}
	opened, err := cutover.box.Open(value)
	if err != nil {
		return slackCredential{}, 0, fmt.Errorf("decrypt legacy Slack credential: %w", err)
	}
	defer clear(opened.Plaintext)
	if opened.Version == 0 {
		token := strings.TrimSpace(string(opened.Plaintext))
		if token == "" {
			return slackCredential{}, 0, errors.New("slack access token is empty")
		}
		return slackCredential{AccessToken: token}, 0, nil
	}
	credential, err := decodeSlackCredential(opened.Plaintext)
	if err != nil {
		return slackCredential{}, 0, err
	}
	return credential, opened.Version, nil
}

// legacySlackWebhookPayloadEnvelope is the retired auth-secret-bound format.
// It remains here only so an explicit cutover can authenticate old durable
// identity before resealing the exact body with the dedicated v2 key.
type legacySlackWebhookPayloadEnvelope struct {
	Version                int       `json:"version"`
	Provider               string    `json:"provider"`
	DeliveryID             string    `json:"deliveryId"`
	WorkspaceID            uuid.UUID `json:"workspaceId"`
	InstallationID         uuid.UUID `json:"installationId"`
	InstallationGeneration uuid.UUID `json:"installationGeneration"`
	Body                   []byte    `json:"body"`
}

func (cutover *LegacyCutover) openWebhookPayload(record webhooks.Record, value string) ([]byte, error) {
	if cutover == nil || cutover.box == nil {
		return nil, errors.New("slack legacy webhook cutover is not configured")
	}
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, slackWebhookPayloadEnvelopePrefix) {
		return nil, errors.New("slack webhook payload is not a legacy cutover candidate")
	}

	if strings.HasPrefix(value, legacySlackWebhookPayloadEnvelopePrefix) {
		return cutover.openBoundLegacyWebhookPayload(record, strings.TrimPrefix(value, legacySlackWebhookPayloadEnvelopePrefix))
	}
	opened, err := cutover.box.Open(value)
	if err != nil {
		return nil, fmt.Errorf("decrypt legacy Slack webhook payload: %w", err)
	}
	defer clear(opened.Plaintext)
	if opened.Version == 0 {
		return nil, errors.New("unencrypted Slack webhook payload is not allowed")
	}
	body := append([]byte(nil), opened.Plaintext...)
	if err := validateLegacySlackWebhookBody(record, body); err != nil {
		clear(body)
		return nil, err
	}
	return body, nil
}

func (cutover *LegacyCutover) openBoundLegacyWebhookPayload(record webhooks.Record, encrypted string) ([]byte, error) {
	opened, err := cutover.box.Open(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt bound legacy Slack webhook payload: %w", err)
	}
	defer clear(opened.Plaintext)
	if opened.Version == 0 {
		return nil, errors.New("unencrypted Slack webhook payload is not allowed")
	}

	decoder := json.NewDecoder(bytes.NewReader(opened.Plaintext))
	decoder.DisallowUnknownFields()
	var envelope legacySlackWebhookPayloadEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, errors.New("decode legacy Slack webhook payload: invalid envelope")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("decode legacy Slack webhook payload: trailing value")
	}
	if envelope.Version != legacySlackWebhookPayloadVersion ||
		envelope.Provider != string(record.Provider) ||
		envelope.DeliveryID != record.DeliveryID ||
		envelope.WorkspaceID != record.WorkspaceID ||
		envelope.InstallationID != record.InstallationID ||
		envelope.InstallationGeneration != record.InstallationGeneration ||
		len(envelope.Body) == 0 {
		return nil, errors.New("legacy Slack webhook payload does not match its durable identity")
	}
	if err := validateLegacySlackWebhookBody(record, envelope.Body); err != nil {
		return nil, err
	}
	return append([]byte(nil), envelope.Body...), nil
}

func validateLegacySlackWebhookBody(record webhooks.Record, body []byte) error {
	envelope, err := decodeSlackEvent(body)
	if err != nil {
		return errors.New("decode legacy Slack webhook payload: invalid event")
	}
	if envelope.EventID != record.DeliveryID || envelope.TeamID != record.ExternalAccountID {
		return errors.New("legacy Slack webhook payload does not match its durable provider identity")
	}
	return nil
}
