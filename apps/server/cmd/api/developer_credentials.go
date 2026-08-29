package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
)

type namedSecuritySecret struct {
	name   string
	secret string
}

func developerCredentialTokenConfig(cfg Config) (developercredentials.TokenKeyringConfig, error) {
	return developercredentials.ParseEncodedTokenKeyring(
		cfg.DeveloperCredentials.ActiveKeyID,
		cfg.DeveloperCredentials.ActiveKeyVersion.Uint32(),
		cfg.DeveloperCredentials.Keys,
	)
}

func newDeveloperCredentialTokenManager(cfg Config) (*developercredentials.TokenManager, error) {
	keyring, err := developerCredentialTokenConfig(cfg)
	if err != nil {
		return nil, err
	}
	return developercredentials.NewTokenManager(keyring)
}

func validateDeveloperCredentialConfig(mode deployment.Mode, cfg Config) error {
	keyring, err := developerCredentialTokenConfig(cfg)
	if err != nil {
		return err
	}
	if _, err := developercredentials.NewTokenManager(keyring); err != nil {
		return err
	}
	if mode.IsProduction() && developercredentials.ContainsDevelopmentDigestKey(keyring) {
		return errors.New("APP_API_CREDENTIAL_HMAC_KEYS must not contain the development key in production")
	}
	return nil
}

func validateDeveloperCredentialKeySeparation(cfg Config, applicationKeys []namedSecuritySecret) error {
	keyring, err := developerCredentialTokenConfig(cfg)
	if err != nil {
		return fmt.Errorf("validate APP_API_CREDENTIAL_HMAC_KEYS key separation: %w", err)
	}
	for _, applicationKey := range applicationKeys {
		if strings.TrimSpace(applicationKey.secret) == "" {
			continue
		}
		if developercredentials.DigestKeyringReusesSecret(keyring, applicationKey.secret) {
			return fmt.Errorf("APP_API_CREDENTIAL_HMAC_KEYS must not reuse %s", applicationKey.name)
		}
	}
	for _, digestKey := range keyring.Keys {
		encodedMaterial := base64.StdEncoding.EncodeToString(digestKey.Material)
		reusesVaultKey, err := credentialvault.ReusesSecretMaterial(
			cfg.CredentialVault.ActiveKeyID,
			cfg.CredentialVault.ActiveKeyVersion.Uint32(),
			cfg.CredentialVault.Keys,
			encodedMaterial,
		)
		if err != nil {
			return fmt.Errorf("validate credential keyring separation: %w", err)
		}
		if reusesVaultKey {
			return errors.New("APP_API_CREDENTIAL_HMAC_KEYS must not reuse APP_CREDENTIAL_VAULT_KEYS")
		}
	}
	return nil
}
