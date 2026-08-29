package main

import (
	"crypto/tls"
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSlackRedirectURLDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Slack)
	redirectURL, found := configType.FieldByName("RedirectURL")
	require.True(t, found)
	require.Equal(t, "https://api.fortyone.app/integrations/slack/setup", redirectURL.Tag.Get("default"))
}

func TestEnvironmentDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{})
	environment, found := configType.FieldByName("Environment")
	require.True(t, found)
	require.Equal(t, "development", environment.Tag.Get("default"))
	require.Equal(t, "APP_ENVIRONMENT", environment.Tag.Get("env"))
}

func TestValidateRuntimeConfigRejectsUnsafeProductionConfiguration(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "secret"
	cfg.VerificationTokens.HMACKey = "development-only-verification-hmac-key"
	cfg.VerificationTokens.HMACKeyID = "v1"
	cfg.DB.SSLMode = "disable"
	cfg.Cache.DisableTLS = true

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_AUTH_SECRET_KEY")
	require.ErrorContains(t, err, "FEEDBACK_INGRESS_SECRET")
	require.ErrorContains(t, err, "APP_EMAIL_REPLY_SECURITY_KEY")
	require.ErrorContains(t, err, "APP_MESSAGING_MUTATION_HMAC_KEY")
	require.ErrorContains(t, err, "APP_VERIFICATION_TOKEN_HMAC_KEY")
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_DB_SSL_MODE")
	require.ErrorContains(t, err, "APP_REDIS_DISABLE_TLS")
}

func TestValidateRuntimeConfigAcceptsSecureProductionConfiguration(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()

	mode, err := validateRuntimeConfig(cfg)
	require.NoError(t, err)
	require.Equal(t, "production", mode.String())
}

func TestValidateRuntimeConfigRejectsStaticProductionAWSCredentials(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.AWS.AccessKeyID = "long-lived-access-key"
	cfg.AWS.SecretAccessKey = "long-lived-secret-key"

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "ECS task role")
	require.NotContains(t, err.Error(), cfg.AWS.AccessKeyID)
	require.NotContains(t, err.Error(), cfg.AWS.SecretAccessKey)
}

func TestValidateRuntimeConfigRejectsUnsafeCORSOrigins(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"wildcard":        "*",
		"insecure origin": "http://app.fortyone.app",
		"sibling pattern": "https://*.fortyone.app",
	}
	for name, origins := range tests {
		name, origins := name, origins
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := secureProductionConfig()
			cfg.Web.CORSAllowedOrigins = origins
			_, err := validateRuntimeConfig(cfg)
			require.ErrorContains(t, err, "APP_API_CORS_ALLOWED_ORIGINS")
		})
	}
}

func TestValidateRuntimeConfigRejectsSemanticDevelopmentVaultKeys(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		activeKeyID string
		keys        string
	}{
		"formatted development key": {
			activeKeyID: "development",
			keys:        `{ "development@1" : "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" }`,
		},
		"renamed development material": {
			activeKeyID: "production",
			keys:        `{"production@1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := secureProductionConfig()
			cfg.CredentialVault.ActiveKeyID = tt.activeKeyID
			cfg.CredentialVault.Keys = tt.keys

			_, err := validateRuntimeConfig(cfg)
			require.ErrorContains(t, err, "APP_CREDENTIAL_VAULT_KEYS must not contain the development key in production")
		})
	}
}

func TestValidateRuntimeConfigRejectsVaultReuseOfApplicationSecrets(t *testing.T) {
	t.Parallel()

	const reusedSecret = "0123456789abcdef0123456789abcdef"
	tests := map[string]struct {
		configName string
		configure  func(*Config)
	}{
		"authentication": {
			configName: "APP_AUTH_SECRET_KEY",
			configure:  func(cfg *Config) { cfg.Auth.SecretKey = reusedSecret },
		},
		"verification": {
			configName: "APP_VERIFICATION_TOKEN_HMAC_KEY",
			configure:  func(cfg *Config) { cfg.VerificationTokens.HMACKey = reusedSecret },
		},
		"email reply": {
			configName: "APP_EMAIL_REPLY_SECURITY_KEY",
			configure:  func(cfg *Config) { cfg.EmailReply.SecurityKey = reusedSecret },
		},
		"messaging mutation": {
			configName: "APP_MESSAGING_MUTATION_HMAC_KEY",
			configure:  func(cfg *Config) { cfg.Messaging.MutationHMACKey = reusedSecret },
		},
		"current invitation": {
			configName: "APP_INVITATION_TOKEN_HMAC_KEY",
			configure:  func(cfg *Config) { cfg.InvitationTokens.HMACKey = reusedSecret },
		},
		"previous invitation": {
			configName: "APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS[old]",
			configure: func(cfg *Config) {
				cfg.InvitationTokens.PreviousKeys = "old=" + reusedSecret
			},
		},
		"Figma webhook payload": {
			configName: "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET",
			configure: func(cfg *Config) {
				cfg.Figma.WebhookPayloadSecret = reusedSecret
			},
		},
		"GitHub webhook payload": {
			configName: "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET",
			configure: func(cfg *Config) {
				cfg.GitHub.WebhookPayloadSecret = reusedSecret
			},
		},
		"Slack webhook payload": {
			configName: "APP_SLACK_WEBHOOK_PAYLOAD_SECRET",
			configure: func(cfg *Config) {
				cfg.Slack.WebhookPayloadSecret = reusedSecret
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := secureProductionConfig()
			cfg.CredentialVault.Keys = `{"production@1":"` + base64.StdEncoding.EncodeToString([]byte(reusedSecret)) + `"}`
			tt.configure(&cfg)

			_, err := validateRuntimeConfig(cfg)
			require.ErrorContains(t, err, "APP_CREDENTIAL_VAULT_KEYS must not reuse "+tt.configName)
			require.NotContains(t, err.Error(), reusedSecret)
		})
	}
}

func TestValidateRuntimeConfigRejectsReusedAuthAndVerificationKeys(t *testing.T) {
	t.Parallel()

	const sharedSecret = "shared-production-secret-with-at-least-32-bytes"
	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = sharedSecret
	cfg.Feedback.IngressSecret = "a-unique-feedback-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.VerificationTokens.HMACKey = sharedSecret
	cfg.VerificationTokens.HMACKeyID = "2026-08-v1"
	cfg.DB.SSLMode = "verify-full"

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_VERIFICATION_TOKEN_HMAC_KEY must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), sharedSecret)
}

func TestVerificationTokenConfigurationDefaults(t *testing.T) {
	configType := reflect.TypeOf(Config{}.VerificationTokens)
	hmacKey, found := configType.FieldByName("HMACKey")
	require.True(t, found)
	require.Equal(t, "APP_VERIFICATION_TOKEN_HMAC_KEY", hmacKey.Tag.Get("env"))

	keyID, found := configType.FieldByName("HMACKeyID")
	require.True(t, found)
	require.Equal(t, "v1", keyID.Tag.Get("default"))
	require.Equal(t, "APP_VERIFICATION_TOKEN_HMAC_KEY_ID", keyID.Tag.Get("env"))
}

func TestFeedbackSecurityConfigurationIsDedicated(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Feedback)
	securityKey, found := configType.FieldByName("SecurityKey")
	require.True(t, found)
	require.Equal(t, "APP_FEEDBACK_SECURITY_KEY", securityKey.Tag.Get("env"))
	require.Equal(t, "development-only-feedback-security-key", securityKey.Tag.Get("default"))
}

func TestEmailAndMessagingSecurityConfigurationIsDedicated(t *testing.T) {
	t.Parallel()

	emailKey, found := reflect.TypeOf(Config{}.EmailReply).FieldByName("SecurityKey")
	require.True(t, found)
	require.Equal(t, "APP_EMAIL_REPLY_SECURITY_KEY", emailKey.Tag.Get("env"))
	require.Equal(t, "development-only-email-reply-security-key", emailKey.Tag.Get("default"))

	mutationKey, found := reflect.TypeOf(Config{}.Messaging).FieldByName("MutationHMACKey")
	require.True(t, found)
	require.Equal(t, "APP_MESSAGING_MUTATION_HMAC_KEY", mutationKey.Tag.Get("env"))
	require.Equal(t, "development-only-messaging-mutation-hmac-key", mutationKey.Tag.Get("default"))
}

func TestValidateRuntimeConfigRejectsPurposeKeyReuse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name      string
		configure func(*Config)
	}{
		"email reply": {
			name:      "APP_EMAIL_REPLY_SECURITY_KEY",
			configure: func(cfg *Config) { cfg.EmailReply.SecurityKey = cfg.Auth.SecretKey },
		},
		"messaging mutation": {
			name:      "APP_MESSAGING_MUTATION_HMAC_KEY",
			configure: func(cfg *Config) { cfg.Messaging.MutationHMACKey = cfg.Auth.SecretKey },
		},
	}
	for testName, test := range tests {
		testName, test := testName, test
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			cfg := secureProductionConfig()
			test.configure(&cfg)
			_, err := validateRuntimeConfig(cfg)
			require.ErrorContains(t, err, test.name+" must not reuse APP_AUTH_SECRET_KEY")
			require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
		})
	}
}

func TestFigmaWebhookPayloadConfigurationIsDedicated(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Figma)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-figma-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestGitHubWebhookPayloadConfigurationIsDedicated(t *testing.T) {
	configType := reflect.TypeOf(Config{}.GitHub)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-github-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestSlackWebhookPayloadConfigurationIsDedicated(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Slack)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-slack-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestValidateRuntimeConfigRejectsFeedbackSecurityKeyReuse(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.Feedback.SecurityKey = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_FEEDBACK_SECURITY_KEY must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsFigmaWebhookPayloadKeyReuse(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.Figma.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsGitHubWebhookPayloadKeyReuse(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.GitHub.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsSlackWebhookPayloadKeyReuse(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.Slack.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsGitHubSigningAndPayloadKeyReuse(t *testing.T) {
	t.Parallel()

	cfg := secureProductionConfig()
	cfg.GitHub.WebhookSecret = cfg.GitHub.WebhookPayloadSecret

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET must not reuse GITHUB_WEBHOOK_SECRET")
	require.NotContains(t, err.Error(), cfg.GitHub.WebhookPayloadSecret)
}

func TestDeveloperCredentialConfigurationDefaults(t *testing.T) {
	configType := reflect.TypeOf(Config{}.DeveloperCredentials)
	activeKeyID, found := configType.FieldByName("ActiveKeyID")
	require.True(t, found)
	require.Equal(t, "APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_ID", activeKeyID.Tag.Get("env"))

	activeVersion, found := configType.FieldByName("ActiveKeyVersion")
	require.True(t, found)
	require.Equal(t, "APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_VERSION", activeVersion.Tag.Get("env"))

	keys, found := configType.FieldByName("Keys")
	require.True(t, found)
	require.Equal(t, "APP_API_CREDENTIAL_HMAC_KEYS", keys.Tag.Get("env"))
}

func TestDeveloperOAuthConfigurationDefaults(t *testing.T) {
	configType := reflect.TypeOf(Config{}.DeveloperOAuth)

	signingKey, found := configType.FieldByName("AccessTokenSigningKey")
	require.True(t, found)
	require.Equal(t, "APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY", signingKey.Tag.Get("env"))

	activeKeyID, found := configType.FieldByName("ActiveDigestKeyID")
	require.True(t, found)
	require.Equal(t, "APP_OAUTH_TOKEN_HMAC_ACTIVE_KEY_ID", activeKeyID.Tag.Get("env"))

	digestKeys, found := configType.FieldByName("DigestKeys")
	require.True(t, found)
	require.Equal(t, "APP_OAUTH_TOKEN_HMAC_KEYS", digestKeys.Tag.Get("env"))
}

func TestValidateRuntimeConfigRejectsDevelopmentDeveloperCredentialKey(t *testing.T) {
	t.Parallel()
	cfg := secureProductionConfig()
	cfg.DeveloperCredentials.ActiveKeyID = "renamed"
	cfg.DeveloperCredentials.Keys = `{"renamed@1":"ZGV2ZWxvcGVyLWNyZWRlbnRpYWwtZGV2LWtleS0wMDE="}`

	_, err := validateRuntimeConfig(cfg)

	require.ErrorContains(t, err, "APP_API_CREDENTIAL_HMAC_KEYS must not contain the development key in production")
}

func TestValidateRuntimeConfigRejectsDeveloperCredentialKeyReuse(t *testing.T) {
	t.Parallel()
	const reusedSecret = "auth-key-0123456789abcdef0123456"
	cfg := secureProductionConfig()
	cfg.Auth.SecretKey = reusedSecret
	cfg.DeveloperCredentials.Keys = `{"production@1":"` + base64.StdEncoding.EncodeToString([]byte(reusedSecret)) + `"}`

	_, err := validateRuntimeConfig(cfg)

	require.ErrorContains(t, err, "APP_API_CREDENTIAL_HMAC_KEYS must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), reusedSecret)
}

func TestRedisOptionsAlwaysVerifyTLSPeer(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Cache.Host = " redis.internal.example "
	cfg.Cache.Port = "6379"

	options := redisOptions(cfg)
	require.NotNil(t, options.TLSConfig)
	require.False(t, options.TLSConfig.InsecureSkipVerify)
	require.Equal(t, uint16(tls.VersionTLS12), options.TLSConfig.MinVersion)
	require.Equal(t, "redis.internal.example", options.TLSConfig.ServerName)

	cfg.Cache.DisableTLS = true
	require.Nil(t, redisOptions(cfg).TLSConfig)
}

func TestMayaSenderAddressDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaAddress, found := configType.FieldByName("MayaAddress")
	require.True(t, found)
	require.Equal(t, "maya@fortyone.app", mayaAddress.Tag.Get("default"))
}

func TestEmailSenderNameDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	fromName, found := configType.FieldByName("FromName")
	require.True(t, found)
	require.Equal(t, "FortyOne", fromName.Tag.Get("default"))
}

func TestMayaSenderNameDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaName, found := configType.FieldByName("MayaName")
	require.True(t, found)
	require.Equal(t, "Maya, AI Agent", mayaName.Tag.Get("default"))
}

func secureProductionConfig() Config {
	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.IngressSecret = "a-unique-feedback-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.EmailReply.SecurityKey = "a-separate-email-reply-key-with-32-bytes"
	cfg.Messaging.MutationHMACKey = "a-separate-messaging-mutation-key-with-32-bytes"
	cfg.VerificationTokens.HMACKey = "a-separate-verification-key-with-32-bytes"
	cfg.VerificationTokens.HMACKeyID = "2026-08-v1"
	cfg.InvitationTokens.HMACKey = "a-separate-invitation-key-with-32-bytes"
	cfg.InvitationTokens.HMACKeyID = "2026-08-v1"
	cfg.CredentialVault.ActiveKeyID = "production"
	cfg.CredentialVault.ActiveKeyVersion = 1
	cfg.CredentialVault.Keys = `{"production@1":"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}`
	cfg.DeveloperCredentials.ActiveKeyID = "production"
	cfg.DeveloperCredentials.ActiveKeyVersion = 1
	cfg.DeveloperCredentials.Keys = `{"production@1":"` + base64.StdEncoding.EncodeToString([]byte("abcdef0123456789abcdef0123456789")) + `"}`
	cfg.DeveloperOAuth.AccessTokenSigningKey = "oauth-access-signing-production-01"
	cfg.DeveloperOAuth.ActiveDigestKeyID = "production"
	cfg.DeveloperOAuth.DigestKeys = `{"production":"` + base64.StdEncoding.EncodeToString([]byte("oauth-digest-production-key-0001")) + `"}`
	cfg.DeveloperOAuth.DynamicClientTTL = 30 * 24 * time.Hour
	cfg.GitHub.WebhookPayloadSecret = "a-separate-github-payload-key-with-32-bytes"
	cfg.Slack.WebhookPayloadSecret = "a-separate-slack-payload-key-with-32-bytes"
	cfg.Figma.WebhookPayloadSecret = "a-separate-figma-payload-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.Web.CORSAllowedOrigins = "https://app.fortyone.app"
	return cfg
}
