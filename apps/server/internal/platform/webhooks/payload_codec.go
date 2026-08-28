package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/secretbox"
	"github.com/google/uuid"
)

const boundPayloadVersion = 1

// BoundPayloadCodec encrypts exact webhook bytes together with the immutable
// durable identity that authenticated them. Open rejects a valid ciphertext
// copied to a different provider, delivery, workspace, or installation grant.
type BoundPayloadCodec struct {
	provider integrations.ProviderKey
	prefix   string
	box      *secretbox.Box
}

type boundPayloadEnvelope struct {
	Version                int       `json:"version"`
	Provider               string    `json:"provider"`
	DeliveryID             string    `json:"deliveryId"`
	WorkspaceID            uuid.UUID `json:"workspaceId"`
	InstallationID         uuid.UUID `json:"installationId"`
	InstallationGeneration uuid.UUID `json:"installationGeneration"`
	Body                   []byte    `json:"body"`
}

func NewBoundPayloadCodec(provider integrations.ProviderKey, prefix, secret string) (*BoundPayloadCodec, error) {
	provider = integrations.ProviderKey(strings.TrimSpace(string(provider)))
	prefix = strings.TrimSpace(prefix)
	if provider == "" || prefix == "" || strings.ContainsAny(prefix, "\r\n\t ") {
		return nil, fmt.Errorf("configure bound webhook payload codec: %w", ErrNotConfigured)
	}
	box, err := secretbox.New(secret)
	if err != nil {
		return nil, fmt.Errorf("configure bound webhook payload encryption: %w", err)
	}
	return &BoundPayloadCodec{provider: provider, prefix: prefix, box: box}, nil
}

func (codec *BoundPayloadCodec) Seal(ctx context.Context, binding PayloadBinding, payload []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if codec == nil || codec.box == nil || binding.Provider != codec.provider ||
		strings.TrimSpace(binding.DeliveryID) == "" || binding.WorkspaceID == uuid.Nil ||
		binding.InstallationID == uuid.Nil || binding.InstallationGeneration == uuid.Nil || len(payload) == 0 {
		return "", ErrInvalidDelivery
	}
	encoded, err := json.Marshal(boundPayloadEnvelope{
		Version:                boundPayloadVersion,
		Provider:               string(binding.Provider),
		DeliveryID:             binding.DeliveryID,
		WorkspaceID:            binding.WorkspaceID,
		InstallationID:         binding.InstallationID,
		InstallationGeneration: binding.InstallationGeneration,
		Body:                   payload,
	})
	if err != nil {
		return "", fmt.Errorf("encode bound webhook payload: %w", err)
	}
	defer clear(encoded)
	sealed, err := codec.box.Seal(encoded)
	if err != nil {
		return "", fmt.Errorf("encrypt bound webhook payload: %w", err)
	}
	return codec.prefix + sealed, nil
}

func (codec *BoundPayloadCodec) Open(record Record, value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if codec == nil || codec.box == nil || record.Provider != codec.provider || !strings.HasPrefix(value, codec.prefix) {
		return nil, errors.New("bound webhook payload envelope is invalid")
	}
	opened, err := codec.box.Open(strings.TrimPrefix(value, codec.prefix))
	if err != nil {
		return nil, fmt.Errorf("decrypt bound webhook payload: %w", err)
	}
	defer clear(opened.Plaintext)
	decoder := json.NewDecoder(bytes.NewReader(opened.Plaintext))
	decoder.DisallowUnknownFields()
	var envelope boundPayloadEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, errors.New("decode bound webhook payload: invalid envelope")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("decode bound webhook payload: trailing value")
	}
	if envelope.Version != boundPayloadVersion || envelope.Provider != string(record.Provider) ||
		envelope.DeliveryID != record.DeliveryID || envelope.WorkspaceID != record.WorkspaceID ||
		envelope.InstallationID != record.InstallationID || envelope.InstallationGeneration != record.InstallationGeneration ||
		len(envelope.Body) == 0 {
		return nil, errors.New("bound webhook payload does not match its durable identity")
	}
	return append([]byte(nil), envelope.Body...), nil
}

var _ PayloadProtector = (*BoundPayloadCodec)(nil)
