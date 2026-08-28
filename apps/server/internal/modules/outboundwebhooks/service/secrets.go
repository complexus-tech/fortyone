package outboundwebhooksservice

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

const (
	webhookSecretPrefix = "whsec_"
	webhookSecretBytes  = 32
	vaultProvider       = "fortyone_outbound_webhooks"
	vaultCredentialType = "endpoint_signing_secret"
)

type SecretVault interface {
	Seal(binding credentialvault.Context, plaintext []byte) (string, error)
	Open(binding credentialvault.Context, encoded string) (credentialvault.Secret, error)
}

type SecretManager struct {
	vault  SecretVault
	random io.Reader
}

func NewSecretManager(vault SecretVault) (*SecretManager, error) {
	return newSecretManager(vault, rand.Reader)
}

func newSecretManager(vault SecretVault, random io.Reader) (*SecretManager, error) {
	if vault == nil || random == nil {
		return nil, fmt.Errorf("outbound webhook secret manager requires a vault and random source")
	}
	return &SecretManager{vault: vault, random: random}, nil
}

func (manager *SecretManager) Generate(workspaceID, endpointID uuid.UUID, generation int) (outboundwebhooksdomain.SigningSecret, string, error) {
	if manager == nil || workspaceID == uuid.Nil || endpointID == uuid.Nil || generation <= 0 {
		return outboundwebhooksdomain.SigningSecret{}, "", outboundwebhooksdomain.ErrInvalidEndpoint
	}
	material := make([]byte, webhookSecretBytes)
	if _, err := io.ReadFull(manager.random, material); err != nil {
		return outboundwebhooksdomain.SigningSecret{}, "", fmt.Errorf("generate outbound webhook signing secret: %w", err)
	}
	encodedLength := base64.StdEncoding.EncodedLen(len(material))
	plaintext := make([]byte, len(webhookSecretPrefix)+encodedLength)
	copy(plaintext, webhookSecretPrefix)
	base64.StdEncoding.Encode(plaintext[len(webhookSecretPrefix):], material)
	clear(material)
	envelope, err := manager.vault.Seal(secretBinding(workspaceID, endpointID, generation), plaintext)
	if err != nil {
		clear(plaintext)
		return outboundwebhooksdomain.SigningSecret{}, "", fmt.Errorf("seal outbound webhook signing secret: %w", err)
	}
	secret := outboundwebhooksdomain.NewSigningSecret(string(plaintext))
	clear(plaintext)
	return secret, envelope, nil
}

func (manager *SecretManager) Open(workspaceID, endpointID uuid.UUID, generation int, envelope string) ([]byte, error) {
	if manager == nil || workspaceID == uuid.Nil || endpointID == uuid.Nil || generation <= 0 || envelope == "" {
		return nil, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	secret, err := manager.vault.Open(secretBinding(workspaceID, endpointID, generation), envelope)
	if err != nil {
		return nil, fmt.Errorf("open outbound webhook signing secret: %w", err)
	}
	defer secret.Destroy()
	plaintext := secret.Reveal()
	if _, err := decodeSigningSecret(plaintext); err != nil {
		clear(plaintext)
		return nil, err
	}
	return plaintext, nil
}

func secretBinding(workspaceID, endpointID uuid.UUID, generation int) credentialvault.Context {
	return credentialvault.Context{
		Provider:       vaultProvider,
		TenantID:       workspaceID.String(),
		SubjectID:      endpointID.String(),
		CredentialType: vaultCredentialType,
		Generation:     strconv.Itoa(generation),
	}
}

func decodeSigningSecret(plaintext []byte) ([]byte, error) {
	if len(plaintext) <= len(webhookSecretPrefix) || string(plaintext[:len(webhookSecretPrefix)]) != webhookSecretPrefix {
		return nil, fmt.Errorf("outbound webhook signing secret has an invalid format")
	}
	material, err := base64.StdEncoding.Strict().DecodeString(string(plaintext[len(webhookSecretPrefix):]))
	if err != nil || len(material) != webhookSecretBytes {
		clear(material)
		return nil, fmt.Errorf("outbound webhook signing secret has an invalid format")
	}
	return material, nil
}
