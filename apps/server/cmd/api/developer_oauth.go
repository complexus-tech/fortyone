package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	developeroauthrepository "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/jackc/pgx/v5/pgxpool"
)

func developerOAuthTokenConfig(cfg Config) (developeroauth.TokenKeyringConfig, error) {
	return developeroauth.ParseEncodedTokenKeyring(
		cfg.DeveloperOAuth.ActiveDigestKeyID,
		cfg.DeveloperOAuth.DigestKeys,
	)
}

type developerOAuthServices struct {
	Platform     *developeroauth.Platform
	PublicAPI    *developeroauth.Service
	Applications *developeroauth.ApplicationManager
}

func newDeveloperOAuthServices(cfg Config, pool *pgxpool.Pool) (developerOAuthServices, error) {
	keyring, err := developerOAuthTokenConfig(cfg)
	if err != nil {
		return developerOAuthServices{}, err
	}
	tokens, err := developeroauth.NewTokenManager(keyring)
	if err != nil {
		return developerOAuthServices{}, err
	}
	repository, err := developeroauthrepository.New(pool)
	if err != nil {
		return developerOAuthServices{}, err
	}
	issuer := strings.TrimRight(cfg.Web.PublicURL, "/")
	mcp, err := developeroauth.New(
		repository,
		tokens,
		developeroauth.WallClock{},
		developeroauth.RandomIDGenerator{},
		developeroauth.Config{
			Issuer: issuer, Resource: issuer + "/mcp",
			AccessTokenSigningKey: cfg.DeveloperOAuth.AccessTokenSigningKey,
			DynamicClientTTL:      cfg.DeveloperOAuth.DynamicClientTTL,
		},
	)
	if err != nil {
		return developerOAuthServices{}, fmt.Errorf("initialize MCP OAuth audience: %w", err)
	}
	publicAPI, err := developeroauth.New(
		repository,
		tokens,
		developeroauth.WallClock{},
		developeroauth.RandomIDGenerator{},
		developeroauth.Config{
			Issuer: issuer, Resource: issuer + "/api/v1",
			ScopePolicy:            developeroauth.PublicAPIResourceScopePolicy(),
			AccessTokenSigningKey:  cfg.DeveloperOAuth.AccessTokenSigningKey,
			DynamicClientTTL:       cfg.DeveloperOAuth.DynamicClientTTL,
			ApplicationActors:      repository,
			ApplicationActorScopes: developeroauth.PublicAPIApplicationActorScopePolicy(),
		},
	)
	if err != nil {
		return developerOAuthServices{}, fmt.Errorf("initialize public API OAuth audience: %w", err)
	}
	platform, err := developeroauth.NewPlatform(mcp, publicAPI)
	if err != nil {
		return developerOAuthServices{}, fmt.Errorf("initialize developer OAuth platform: %w", err)
	}
	applications, err := developeroauth.NewApplicationManager(
		repository,
		tokens,
		developeroauth.WallClock{},
		developeroauth.RandomIDGenerator{},
		publicAPI.Resource(),
	)
	if err != nil {
		return developerOAuthServices{}, fmt.Errorf("initialize OAuth application management: %w", err)
	}
	return developerOAuthServices{
		Platform: platform, PublicAPI: publicAPI, Applications: applications,
	}, nil
}

func validateDeveloperOAuthConfig(mode deployment.Mode, cfg Config) error {
	keyring, err := developerOAuthTokenConfig(cfg)
	if err != nil {
		return err
	}
	if _, err := developeroauth.NewTokenManager(keyring); err != nil {
		return err
	}
	if len(cfg.DeveloperOAuth.AccessTokenSigningKey) < 32 {
		return errors.New("APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY must contain at least 32 bytes")
	}
	if cfg.DeveloperOAuth.DynamicClientTTL < time.Hour || cfg.DeveloperOAuth.DynamicClientTTL > 90*24*time.Hour {
		return errors.New("APP_OAUTH_DYNAMIC_CLIENT_TTL must be between 1h and 2160h")
	}
	if mode.IsProduction() && developeroauth.ContainsDevelopmentDigestKey(keyring) {
		return errors.New("APP_OAUTH_TOKEN_HMAC_KEYS must not contain the development key in production")
	}
	if developeroauth.DigestKeyringReusesSecret(keyring, cfg.DeveloperOAuth.AccessTokenSigningKey) {
		return errors.New("APP_OAUTH_TOKEN_HMAC_KEYS must not reuse APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY")
	}
	return nil
}

func validateDeveloperOAuthKeySeparation(cfg Config, applicationKeys []namedSecuritySecret) error {
	oauthKeyring, err := developerOAuthTokenConfig(cfg)
	if err != nil {
		return fmt.Errorf("validate APP_OAUTH_TOKEN_HMAC_KEYS key separation: %w", err)
	}
	developerKeyring, err := developerCredentialTokenConfig(cfg)
	if err != nil {
		return fmt.Errorf("validate APP_API_CREDENTIAL_HMAC_KEYS key separation: %w", err)
	}
	for _, applicationKey := range applicationKeys {
		if strings.TrimSpace(applicationKey.secret) == "" {
			continue
		}
		if developeroauth.DigestKeyringReusesSecret(oauthKeyring, applicationKey.secret) {
			return fmt.Errorf("APP_OAUTH_TOKEN_HMAC_KEYS must not reuse %s", applicationKey.name)
		}
	}
	for _, oauthKey := range oauthKeyring.Keys {
		encodedMaterial := base64.StdEncoding.EncodeToString(oauthKey.Material)
		reusesVaultKey, err := credentialvault.ReusesSecretMaterial(
			cfg.CredentialVault.ActiveKeyID,
			cfg.CredentialVault.ActiveKeyVersion.Uint32(),
			cfg.CredentialVault.Keys,
			encodedMaterial,
		)
		if err != nil {
			return fmt.Errorf("validate OAuth/vault key separation: %w", err)
		}
		if reusesVaultKey {
			return errors.New("APP_OAUTH_TOKEN_HMAC_KEYS must not reuse APP_CREDENTIAL_VAULT_KEYS")
		}
		if developercredentials.DigestKeyringReusesSecret(developerKeyring, encodedMaterial) {
			return errors.New("APP_OAUTH_TOKEN_HMAC_KEYS must not reuse APP_API_CREDENTIAL_HMAC_KEYS")
		}
	}
	return nil
}
