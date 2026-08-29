package workerbootstrap

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadsSlackUninstallCredentials(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("SLACK_CLIENT_ID", "worker-client-id")
	t.Setenv("SLACK_CLIENT_SECRET", "worker-client-secret")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, "worker-client-id", cfg.Slack.ClientID)
	require.Equal(t, "worker-client-secret", cfg.Slack.ClientSecret)
}

func TestLoadConfigReadsCredentialVaultMaintenanceMode(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("APP_CREDENTIAL_VAULT_REWRAP_ON_STARTUP", "true")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.True(t, cfg.CredentialVault.RewrapOnStartup)
}

func TestLoadConfigReadsGitHubWebhookPayloadSecret(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("APP_GITHUB_WEBHOOK_PAYLOAD_SECRET", "github-webhook-payload-secret-with-32-bytes")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, "github-webhook-payload-secret-with-32-bytes", cfg.GitHub.WebhookPayloadSecret)
}

func TestLoadConfigReadsSlackWebhookPayloadSecret(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("APP_SLACK_WEBHOOK_PAYLOAD_SECRET", "slack-webhook-payload-secret-with-32-bytes")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, "slack-webhook-payload-secret-with-32-bytes", cfg.Slack.WebhookPayloadSecret)
}

func TestLoadConfigReadsFigmaWorkerConfiguration(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("FIGMA_CLIENT_ID", "figma-client-id")
	t.Setenv("FIGMA_CLIENT_SECRET", "figma-client-secret")
	t.Setenv("FIGMA_REDIRECT_URL", "https://api.example.com/integrations/figma/callback")
	t.Setenv("FIGMA_WEBHOOK_URL", "https://api.example.com/webhooks/figma")
	t.Setenv("APP_FIGMA_WEBHOOK_PAYLOAD_SECRET", "figma-webhook-payload-secret-with-32-bytes")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, "figma-client-id", cfg.Figma.ClientID)
	require.Equal(t, "figma-client-secret", cfg.Figma.ClientSecret)
	require.Equal(t, "https://api.example.com/integrations/figma/callback", cfg.Figma.RedirectURL)
	require.Equal(t, "https://api.example.com/webhooks/figma", cfg.Figma.WebhookURL)
	require.Equal(t, "figma-webhook-payload-secret-with-32-bytes", cfg.Figma.WebhookPayloadSecret)
}

func TestEnvironmentDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{})
	environment, found := configType.FieldByName("Environment")
	require.True(t, found)
	require.Equal(t, "development", environment.Tag.Get("default"))
	require.Equal(t, "APP_ENVIRONMENT", environment.Tag.Get("env"))
}

func TestFeedbackSecurityConfigurationIsSharedWithAPI(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Feedback)
	securityKey, found := configType.FieldByName("SecurityKey")
	require.True(t, found)
	require.Equal(t, "APP_FEEDBACK_SECURITY_KEY", securityKey.Tag.Get("env"))
	require.Equal(t, "development-only-feedback-security-key", securityKey.Tag.Get("default"))
}

func TestEmailAndMessagingSecurityConfigurationIsSharedWithAPI(t *testing.T) {
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

func TestFigmaWebhookPayloadConfigurationIsSharedWithAPI(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Figma)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-figma-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestGitHubWebhookPayloadConfigurationIsSharedWithAPI(t *testing.T) {
	configType := reflect.TypeOf(Config{}.GitHub)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-github-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestSlackWebhookPayloadConfigurationIsSharedWithAPI(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Slack)
	payloadSecret, found := configType.FieldByName("WebhookPayloadSecret")
	require.True(t, found)
	require.Equal(t, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET", payloadSecret.Tag.Get("env"))
	require.Equal(t, "development-only-slack-webhook-payload-secret", payloadSecret.Tag.Get("default"))
}

func TestValidateRuntimeConfigRejectsUnsafeProductionConfiguration(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "secret"
	cfg.DB.SSLMode = "disable"
	cfg.Redis.DisableTLS = true

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_AUTH_SECRET_KEY")
	require.ErrorContains(t, err, "APP_EMAIL_REPLY_SECURITY_KEY")
	require.ErrorContains(t, err, "APP_MESSAGING_MUTATION_HMAC_KEY")
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET")
	require.ErrorContains(t, err, "APP_DB_SSL_MODE")
	require.ErrorContains(t, err, "APP_REDIS_DISABLE_TLS")
}

func TestValidateRuntimeConfigAcceptsSecureProductionConfiguration(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)

	mode, err := validateRuntimeConfig(cfg)
	require.NoError(t, err)
	require.Equal(t, "production", mode.String())
}

func TestValidateRuntimeConfigRejectsStaticProductionAWSCredentials(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.AWS.AccessKeyID = "long-lived-access-key"
	cfg.AWS.SecretAccessKey = "long-lived-secret-key"

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "ECS task role")
	require.NotContains(t, err.Error(), cfg.AWS.AccessKeyID)
	require.NotContains(t, err.Error(), cfg.AWS.SecretAccessKey)
}

func TestValidateHTTPConfigRejectsEnabledMonitorWithoutCredentials(t *testing.T) {
	t.Parallel()

	err := validateHTTPConfig(validWorkerHTTPConfig(), MonitorConfig{Enabled: true})
	require.ErrorContains(t, err, "APP_WORKER_MONITOR_USERNAME")
	require.ErrorContains(t, err, "APP_WORKER_MONITOR_PASSWORD")
}

func TestValidateRuntimeConfigRejectsWeakProductionMonitorPassword(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.Monitor = MonitorConfig{
		Enabled:  true,
		Username: "operator",
		Password: "short",
	}

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_WORKER_MONITOR_PASSWORD")
}

func TestWorkerHTTPConfigDefaults(t *testing.T) {
	t.Parallel()

	configType := reflect.TypeOf(HTTPConfig{})
	expectedDefaults := map[string]string{
		"Host":              "0.0.0.0:8080",
		"ReadHeaderTimeout": "5s",
		"ReadTimeout":       "10s",
		"WriteTimeout":      "15s",
		"IdleTimeout":       "60s",
		"ShutdownTimeout":   "15s",
	}
	for fieldName, expected := range expectedDefaults {
		field, found := configType.FieldByName(fieldName)
		require.True(t, found, fieldName)
		require.Equal(t, expected, field.Tag.Get("default"), fieldName)
	}
}

func TestWorkerDatabasePoolConfigMatchesAPIRuntimeControls(t *testing.T) {
	t.Parallel()

	configType := reflect.TypeOf(Config{}.DB)
	expectedDefaults := map[string]string{
		"MinConns":          "0",
		"ConnectTimeout":    "10s",
		"MaxConnIdleTime":   "30m",
		"MaxConnLifetime":   "1h",
		"HealthCheckPeriod": "1m",
	}
	for fieldName, expected := range expectedDefaults {
		field, found := configType.FieldByName(fieldName)
		require.True(t, found, fieldName)
		require.Equal(t, expected, field.Tag.Get("default"), fieldName)
		require.NotEmpty(t, field.Tag.Get("env"), fieldName)
	}
}

func TestLoadConfigReadsMessagingAssistantBudgets(t *testing.T) {
	t.Setenv("APP_GITHUB_APP_ID", "0")
	t.Setenv("OPENAI_ASSISTANT_USER_CALLS_PER_MINUTE", "18")
	t.Setenv("OPENAI_ASSISTANT_WORKSPACE_CALLS_PER_MINUTE", "240")
	t.Setenv("OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY", "1500000")

	cfg, err := loadConfig()
	require.NoError(t, err)
	require.Equal(t, int64(18), cfg.MessagingAssistant.UserCallsPerMinute)
	require.Equal(t, int64(240), cfg.MessagingAssistant.WorkspaceCallsPerMinute)
	require.Equal(t, int64(1_500_000), cfg.MessagingAssistant.WorkspaceTokensPerDay)
}

func TestMessagingAssistantBudgetDefaults(t *testing.T) {
	configType := reflect.TypeOf(Config{}.MessagingAssistant)
	userLimit, found := configType.FieldByName("UserCallsPerMinute")
	require.True(t, found)
	workspaceLimit, found := configType.FieldByName("WorkspaceCallsPerMinute")
	require.True(t, found)
	dailyTokenLimit, found := configType.FieldByName("WorkspaceTokensPerDay")
	require.True(t, found)

	require.Equal(t, "12", userLimit.Tag.Get("default"))
	require.Equal(t, "120", workspaceLimit.Tag.Get("default"))
	require.Equal(t, "1000000", dailyTokenLimit.Tag.Get("default"))
}

func TestMessagingAssistantModelDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{})
	model, found := configType.FieldByName("AIModel")
	require.True(t, found)
	require.Equal(t, "gpt-5.6-luna", model.Tag.Get("default"))
}

func TestMayaSenderAddressDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaAddress, found := configType.FieldByName("MayaAddress")
	require.True(t, found)
	require.Equal(t, "maya@fortyone.app", mayaAddress.Tag.Get("default"))
}

func TestMayaSenderNameDefault(t *testing.T) {
	configType := reflect.TypeOf(Config{}.Email)
	mayaName, found := configType.FieldByName("MayaName")
	require.True(t, found)
	require.Equal(t, "Maya, AI Agent", mayaName.Tag.Get("default"))
}

func validWorkerHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Host:              "127.0.0.1:0",
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
		ShutdownTimeout:   time.Second,
	}
}

func setValidWorkerCredentialVault(cfg *Config) {
	cfg.CredentialVault.ActiveKeyID = "production"
	cfg.CredentialVault.ActiveKeyVersion = 1
	cfg.CredentialVault.Keys = `{"production@1":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="}`
	cfg.EmailReply.SecurityKey = "worker-email-reply-key-with-at-least-32-bytes"
	cfg.Messaging.MutationHMACKey = "worker-messaging-mutation-key-with-at-least-32-bytes"
	cfg.InvitationTokens.HMACKeyID = "2026-08-v1"
	cfg.InvitationTokens.HMACKey = "worker-invitation-key-with-at-least-32-bytes"
	cfg.GitHub.WebhookPayloadSecret = "worker-github-payload-key-with-at-least-32-bytes"
	cfg.Slack.WebhookPayloadSecret = "worker-slack-payload-key-with-at-least-32-bytes"
	cfg.Figma.WebhookPayloadSecret = "worker-figma-payload-key-with-at-least-32-bytes"
}

func TestValidateRuntimeConfigRejectsPurposeKeyReuseInProduction(t *testing.T) {
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
			var cfg Config
			cfg.Environment = "production"
			cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
			cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
			cfg.DB.SSLMode = "verify-full"
			cfg.HTTP = validWorkerHTTPConfig()
			setValidWorkerCredentialVault(&cfg)
			test.configure(&cfg)

			_, err := validateRuntimeConfig(cfg)
			require.ErrorContains(t, err, test.name+" must not reuse APP_AUTH_SECRET_KEY")
			require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
		})
	}
}

func TestValidateRuntimeConfigRejectsInvitationKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.InvitationTokens.HMACKey = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "must not reuse APP_AUTH_SECRET_KEY")
}

func TestValidateRuntimeConfigRejectsFeedbackSecurityKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = cfg.Auth.SecretKey
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_FEEDBACK_SECURITY_KEY must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsFigmaWebhookPayloadKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.Figma.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_FIGMA_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsGitHubWebhookPayloadKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.GitHub.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsSlackWebhookPayloadKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-shared-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.Slack.WebhookPayloadSecret = cfg.Auth.SecretKey

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_SLACK_WEBHOOK_PAYLOAD_SECRET must not reuse APP_AUTH_SECRET_KEY")
	require.NotContains(t, err.Error(), cfg.Auth.SecretKey)
}

func TestValidateRuntimeConfigRejectsGitHubSigningAndPayloadKeyReuseInProduction(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.GitHub.WebhookSecret = cfg.GitHub.WebhookPayloadSecret

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_GITHUB_WEBHOOK_PAYLOAD_SECRET must not reuse GITHUB_WEBHOOK_SECRET")
	require.NotContains(t, err.Error(), cfg.GitHub.WebhookPayloadSecret)
}

func TestValidateRuntimeConfigRejectsVaultReuseOfGitHubWebhookPayloadKey(t *testing.T) {
	t.Parallel()

	const reusedSecret = "0123456789abcdef0123456789abcdef" // gitleaks:allow -- deterministic key-reuse test vector
	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.GitHub.WebhookPayloadSecret = reusedSecret
	cfg.CredentialVault.Keys = `{"production@1":"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}`

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_CREDENTIAL_VAULT_KEYS must not reuse APP_GITHUB_WEBHOOK_PAYLOAD_SECRET")
	require.NotContains(t, err.Error(), reusedSecret)
}

func TestValidateRuntimeConfigRejectsVaultReuseOfFigmaWebhookPayloadKey(t *testing.T) {
	t.Parallel()

	const reusedSecret = "0123456789abcdef0123456789abcdef" // gitleaks:allow -- deterministic key-reuse test vector
	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.Figma.WebhookPayloadSecret = reusedSecret
	cfg.CredentialVault.Keys = `{"production@1":"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}`

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_CREDENTIAL_VAULT_KEYS must not reuse APP_FIGMA_WEBHOOK_PAYLOAD_SECRET")
	require.NotContains(t, err.Error(), reusedSecret)
}

func TestValidateRuntimeConfigRejectsReformattedDevelopmentVaultKey(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "a-unique-production-secret-with-32-bytes"
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.CredentialVault.ActiveKeyID = "development"
	cfg.CredentialVault.Keys = `{ "development@1" : "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" }`

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "must not contain the public development key")
}

func TestValidateRuntimeConfigRejectsVaultKEKReuse(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.Environment = "production"
	cfg.Auth.SecretKey = "0123456789abcdef0123456789abcdef" // gitleaks:allow -- deterministic key-reuse test vector
	cfg.Feedback.SecurityKey = "a-unique-feedback-security-key-with-32-bytes"
	cfg.DB.SSLMode = "verify-full"
	cfg.HTTP = validWorkerHTTPConfig()
	setValidWorkerCredentialVault(&cfg)
	cfg.CredentialVault.Keys = `{"production@1":"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}`

	_, err := validateRuntimeConfig(cfg)
	require.ErrorContains(t, err, "APP_CREDENTIAL_VAULT_KEYS must not reuse APP_AUTH_SECRET_KEY")
}
