// Package appkeys derives purpose-specific application keys from the single
// stable application root secret. The derived keys are cryptographically
// isolated even though operators only manage APP_AUTH_SECRET_KEY.
package appkeys

import (
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
)

const (
	derivedKeySize = 32
	derivedKeyID   = "application-root"
	derivedVersion = uint32(1)

	derivationSalt = "fortyone/application-keys/v1"

	providerCredentialPurpose = "provider-credentials/v1"
	githubWebhookPurpose      = "github-webhook-payload/v1"
	slackWebhookPurpose       = "slack-webhook-payload/v1"
	figmaWebhookPurpose       = "figma-webhook-payload/v1"
)

// IntegrationKeys contains the independent keys shared by the API and worker
// integration runtimes. Values are derived once during bootstrap and are never
// accepted as separate environment variables.
type IntegrationKeys struct {
	CredentialVault            *credentialvault.Vault
	GitHubWebhookPayloadSecret string
	SlackWebhookPayloadSecret  string
	FigmaWebhookPayloadSecret  string
}

// NewIntegrationKeys derives the integration key domains from rootSecret.
// Fixed, versioned purpose labels prevent one derived key from being used in a
// different protocol. Changing a label is a key rotation and must be treated as
// a migration.
func NewIntegrationKeys(rootSecret string) (IntegrationKeys, error) {
	if strings.TrimSpace(rootSecret) == "" {
		return IntegrationKeys{}, errors.New("derive integration keys: application root secret is required")
	}

	vaultMaterial, err := derive(rootSecret, providerCredentialPurpose)
	if err != nil {
		return IntegrationKeys{}, err
	}
	defer clear(vaultMaterial)

	vault, err := credentialvault.New(credentialvault.Config{
		Active: credentialvault.KeyRef{ID: derivedKeyID, Version: derivedVersion},
		Keys: []credentialvault.Key{{
			Ref:      credentialvault.KeyRef{ID: derivedKeyID, Version: derivedVersion},
			Material: vaultMaterial,
		}},
	})
	if err != nil {
		return IntegrationKeys{}, err
	}

	githubSecret, err := deriveEncoded(rootSecret, githubWebhookPurpose)
	if err != nil {
		return IntegrationKeys{}, err
	}
	slackSecret, err := deriveEncoded(rootSecret, slackWebhookPurpose)
	if err != nil {
		return IntegrationKeys{}, err
	}
	figmaSecret, err := deriveEncoded(rootSecret, figmaWebhookPurpose)
	if err != nil {
		return IntegrationKeys{}, err
	}

	return IntegrationKeys{
		CredentialVault:            vault,
		GitHubWebhookPayloadSecret: githubSecret,
		SlackWebhookPayloadSecret:  slackSecret,
		FigmaWebhookPayloadSecret:  figmaSecret,
	}, nil
}

func deriveEncoded(rootSecret, purpose string) (string, error) {
	material, err := derive(rootSecret, purpose)
	if err != nil {
		return "", err
	}
	defer clear(material)
	return base64.RawStdEncoding.EncodeToString(material), nil
}

func derive(rootSecret, purpose string) ([]byte, error) {
	material, err := hkdf.Key(
		sha256.New,
		[]byte(rootSecret),
		[]byte(derivationSalt),
		purpose,
		derivedKeySize,
	)
	if err != nil {
		return nil, errors.New("derive integration key material")
	}
	return material, nil
}
