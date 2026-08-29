package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/appkeys"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/redis/go-redis/v9"
)

func redisOptions(cfg Config) *redis.Options {
	var tlsConfig *tls.Config
	if !cfg.Cache.DisableTLS {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: strings.TrimSpace(cfg.Cache.Host),
		}
	}

	return &redis.Options{
		Addr:         net.JoinHostPort(cfg.Cache.Host, cfg.Cache.Port),
		Password:     cfg.Cache.Password,
		DB:           cfg.Cache.Name,
		TLSConfig:    tlsConfig,
		DialTimeout:  cfg.Cache.DialTimeout,
		ReadTimeout:  cfg.Cache.ReadTimeout,
		WriteTimeout: cfg.Cache.WriteTimeout,
		PoolSize:     cfg.Cache.PoolSize,
	}
}

func validateRuntimeConfig(cfg Config) (deployment.Mode, error) {
	mode, err := deployment.Parse(cfg.Environment)
	if err != nil {
		return "", fmt.Errorf("validate runtime configuration: %w", err)
	}

	sslMode, databaseErr := platformdatabase.EffectiveSSLMode(platformdatabase.Config{
		SSLMode:     cfg.DB.SSLMode,
		SSLRootCert: cfg.DB.SSLRootCert,
		DisableTLS:  cfg.DB.DisableTLS,
	})
	secretErr := deployment.ValidateProductionSecrets(mode,
		deployment.SecretRequirement{
			Name:            "APP_AUTH_SECRET_KEY",
			Value:           cfg.Auth.SecretKey,
			ForbiddenValues: []string{"secret"},
		},
		deployment.SecretRequirement{
			Name:  "FEEDBACK_INGRESS_SECRET",
			Value: cfg.Feedback.IngressSecret,
		},
		deployment.SecretRequirement{
			Name:            "APP_FEEDBACK_SECURITY_KEY",
			Value:           cfg.Feedback.SecurityKey,
			ForbiddenValues: []string{"development-only-feedback-security-key"},
		},
		deployment.SecretRequirement{
			Name:            "APP_EMAIL_REPLY_SECURITY_KEY",
			Value:           cfg.EmailReply.SecurityKey,
			ForbiddenValues: []string{"development-only-email-reply-security-key"},
		},
		deployment.SecretRequirement{
			Name:            "APP_MESSAGING_MUTATION_HMAC_KEY",
			Value:           cfg.Messaging.MutationHMACKey,
			ForbiddenValues: []string{"development-only-messaging-mutation-hmac-key"},
		},
		deployment.SecretRequirement{
			Name:            "APP_VERIFICATION_TOKEN_HMAC_KEY",
			Value:           cfg.VerificationTokens.HMACKey,
			ForbiddenValues: []string{"development-only-verification-hmac-key"},
		},
		deployment.SecretRequirement{
			Name:            "APP_INVITATION_TOKEN_HMAC_KEY",
			Value:           cfg.InvitationTokens.HMACKey,
			ForbiddenValues: []string{"development-only-invitation-hmac-key"},
		},
		deployment.SecretRequirement{
			Name:  "APP_API_CREDENTIAL_HMAC_KEYS",
			Value: cfg.DeveloperCredentials.Keys,
		},
		deployment.SecretRequirement{
			Name:            "APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY",
			Value:           cfg.DeveloperOAuth.AccessTokenSigningKey,
			ForbiddenValues: []string{"oauth-development-access-sign-01"},
		},
		deployment.SecretRequirement{
			Name:  "APP_OAUTH_TOKEN_HMAC_KEYS",
			Value: cfg.DeveloperOAuth.DigestKeys,
		},
	)
	invitationConfig, invitationConfigErr := invitationTokenConfig(cfg)
	securityKeySeparationErr := validateSecurityKeySeparation(mode, cfg, invitationConfig)
	developerCredentialConfigErr := validateDeveloperCredentialConfig(mode, cfg)
	developerOAuthConfigErr := validateDeveloperOAuthConfig(mode, cfg)
	_, verificationTokenErr := users.NewVerificationTokenManager(verificationTokenConfig(cfg))
	_, integrationKeyErr := appkeys.NewIntegrationKeys(cfg.Auth.SecretKey)
	var invitationTokenErr error
	if invitationConfigErr == nil {
		_, invitationTokenErr = invitations.NewInvitationTokenManager(invitationConfig)
	}
	transportErr := deployment.ValidateProductionTransports(mode, deployment.TransportSecurity{
		PostgreSQLSSLMode: sslMode,
		RedisTLSDisabled:  cfg.Cache.DisableTLS,
	})
	awsCredentialErr := deployment.ValidateAWSCredentialSource(
		mode,
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
	)
	originPolicy, originPolicyErr := web.NewOriginPolicy(cfg.Web.CORSAllowedOrigins)
	if originPolicyErr != nil {
		originPolicyErr = fmt.Errorf("APP_API_CORS_ALLOWED_ORIGINS: %w", originPolicyErr)
	} else if mode.IsProduction() {
		originPolicyErr = originPolicy.ValidateHTTPS()
		if originPolicyErr != nil {
			originPolicyErr = fmt.Errorf("APP_API_CORS_ALLOWED_ORIGINS: %w", originPolicyErr)
		}
	}

	if err := errors.Join(
		databaseErr,
		secretErr,
		invitationConfigErr,
		securityKeySeparationErr,
		verificationTokenErr,
		invitationTokenErr,
		integrationKeyErr,
		developerCredentialConfigErr,
		developerOAuthConfigErr,
		transportErr,
		awsCredentialErr,
		originPolicyErr,
	); err != nil {
		return "", fmt.Errorf("validate runtime configuration: %w", err)
	}
	return mode, nil
}

func validateSecurityKeySeparation(
	mode deployment.Mode,
	cfg Config,
	invitationConfig invitations.InvitationTokenConfig,
) error {
	if !mode.IsProduction() {
		return nil
	}

	keys := []namedSecuritySecret{
		{name: "APP_AUTH_SECRET_KEY", secret: cfg.Auth.SecretKey},
		{name: "FEEDBACK_INGRESS_SECRET", secret: cfg.Feedback.IngressSecret},
		{name: "APP_FEEDBACK_SECURITY_KEY", secret: cfg.Feedback.SecurityKey},
		{name: "APP_EMAIL_REPLY_SECURITY_KEY", secret: cfg.EmailReply.SecurityKey},
		{name: "APP_MESSAGING_MUTATION_HMAC_KEY", secret: cfg.Messaging.MutationHMACKey},
		{name: "APP_VERIFICATION_TOKEN_HMAC_KEY", secret: cfg.VerificationTokens.HMACKey},
		{name: "APP_INVITATION_TOKEN_HMAC_KEY", secret: invitationConfig.Current.Secret},
		{name: "APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY", secret: cfg.DeveloperOAuth.AccessTokenSigningKey},
		{name: "GITHUB_CLIENT_SECRET", secret: cfg.GitHub.ClientSecret},
		{name: "GITHUB_PRIVATE_KEY_BASE64", secret: cfg.GitHub.PrivateKeyBase64},
		{name: "GITHUB_WEBHOOK_SECRET", secret: cfg.GitHub.WebhookSecret},
		{name: "SLACK_CLIENT_SECRET", secret: cfg.Slack.ClientSecret},
		{name: "SLACK_SIGNING_SECRET", secret: cfg.Slack.SigningSecret},
		{name: "FIGMA_CLIENT_SECRET", secret: cfg.Figma.ClientSecret},
	}
	for _, previous := range invitationConfig.Previous {
		keys = append(keys, namedSecuritySecret{
			name:   "APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS[" + previous.ID + "]",
			secret: previous.Secret,
		})
	}

	for index, key := range keys {
		secret := strings.TrimSpace(key.secret)
		if secret == "" {
			continue
		}
		for otherIndex := index + 1; otherIndex < len(keys); otherIndex++ {
			if secret == strings.TrimSpace(keys[otherIndex].secret) {
				return fmt.Errorf("%s must not reuse %s", keys[otherIndex].name, key.name)
			}
		}
	}
	if err := validateDeveloperCredentialKeySeparation(cfg, keys); err != nil {
		return err
	}
	if err := validateDeveloperOAuthKeySeparation(cfg, keys); err != nil {
		return err
	}
	return nil
}
