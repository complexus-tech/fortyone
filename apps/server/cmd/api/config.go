package main

import (
	"errors"
	"strings"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/configvalue"
)

type databaseConfig struct {
	Host                      string        `default:"localhost" env:"APP_DB_HOST"`
	Port                      string        `default:"5432" env:"APP_DB_PORT"`
	User                      string        `default:"postgres" env:"APP_DB_USER"`
	Password                  string        `default:"password" env:"APP_DB_PASSWORD"`
	Name                      string        `default:"complexus" env:"APP_DB_NAME"`
	MigrationMaxIdleConns     int           `default:"25" env:"APP_DB_MAX_IDLE_CONNS"`
	MigrationStatementTimeout time.Duration `default:"30m" env:"APP_DB_MIGRATION_STATEMENT_TIMEOUT"`
	MaxOpenConns              int           `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
	MinConns                  int           `default:"0" env:"APP_DB_MIN_CONNS"`
	ConnectTimeout            time.Duration `default:"10s" env:"APP_DB_CONNECT_TIMEOUT"`
	MaxConnIdleTime           time.Duration `default:"30m" env:"APP_DB_MAX_CONN_IDLE_TIME"`
	MaxConnLifetime           time.Duration `default:"1h" env:"APP_DB_MAX_CONN_LIFETIME"`
	HealthCheckPeriod         time.Duration `default:"1m" env:"APP_DB_HEALTH_CHECK_PERIOD"`
	SSLMode                   string        `env:"APP_DB_SSL_MODE"`
	SSLRootCert               string        `env:"APP_DB_SSL_ROOT_CERT"`
	DisableTLS                bool          `default:"true" env:"APP_DB_DISABLE_TLS"`
}

type migrationProcessConfig struct {
	Environment string `default:"development" env:"APP_ENVIRONMENT"`
	DB          databaseConfig
}

// Config is the API process configuration schema. Keep environment bindings in
// this capability-named file so process composition remains easy to navigate.
type Config struct {
	Environment string `default:"development" env:"APP_ENVIRONMENT"`
	Auth        struct {
		SecretKey                    string `default:"secret" env:"APP_AUTH_SECRET_KEY"`
		CookieDomain                 string `env:"APP_AUTH_COOKIE_DOMAIN"`
		GoogleClientIDs              string `env:"APP_AUTH_GOOGLE_CLIENT_IDS"`
		GoogleClientSecret           string `env:"APP_AUTH_GOOGLE_CLIENT_SECRET"`
		GoogleRedirectURL            string `env:"APP_AUTH_GOOGLE_REDIRECT_URL"`
		GoogleCalendarRedirectURL    string `env:"APP_AUTH_GOOGLE_CALENDAR_REDIRECT_URL"`
		GoogleCalendarWebhookURL     string `env:"APP_AUTH_GOOGLE_CALENDAR_WEBHOOK_URL"`
		MicrosoftClientID            string `env:"APP_AUTH_MICROSOFT_CLIENT_ID"`
		MicrosoftClientSecret        string `env:"APP_AUTH_MICROSOFT_CLIENT_SECRET"`
		MicrosoftTenant              string `default:"common" env:"APP_AUTH_MICROSOFT_TENANT"`
		MicrosoftRedirectURL         string `env:"APP_AUTH_MICROSOFT_REDIRECT_URL"`
		MicrosoftCalendarRedirectURL string `env:"APP_AUTH_MICROSOFT_CALENDAR_REDIRECT_URL"`
		MicrosoftCalendarWebhookURL  string `env:"APP_AUTH_MICROSOFT_CALENDAR_WEBHOOK_URL"`
	}
	Feedback struct {
		IngressSecret string `env:"FEEDBACK_INGRESS_SECRET"`
		SecurityKey   string `default:"development-only-feedback-security-key" env:"APP_FEEDBACK_SECURITY_KEY"`
	}
	EmailReply struct {
		SecurityKey string `default:"development-only-email-reply-security-key" env:"APP_EMAIL_REPLY_SECURITY_KEY"`
	}
	Messaging struct {
		MutationHMACKey string `default:"development-only-messaging-mutation-hmac-key" env:"APP_MESSAGING_MUTATION_HMAC_KEY"`
	}
	VerificationTokens struct {
		HMACKey   string `default:"development-only-verification-hmac-key" env:"APP_VERIFICATION_TOKEN_HMAC_KEY"`
		HMACKeyID string `default:"v1" env:"APP_VERIFICATION_TOKEN_HMAC_KEY_ID"`
	}
	InvitationTokens struct {
		HMACKey      string `default:"development-only-invitation-hmac-key" env:"APP_INVITATION_TOKEN_HMAC_KEY"`
		HMACKeyID    string `default:"v1" env:"APP_INVITATION_TOKEN_HMAC_KEY_ID"`
		PreviousKeys string `env:"APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS"`
	}
	CredentialVault struct {
		ActiveKeyID      string                 `default:"development" env:"APP_CREDENTIAL_VAULT_ACTIVE_KEY_ID"`
		ActiveKeyVersion configvalue.KeyVersion `default:"1" env:"APP_CREDENTIAL_VAULT_ACTIVE_KEY_VERSION"`
		Keys             string                 `default:"{\"development@1\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}" env:"APP_CREDENTIAL_VAULT_KEYS"`
	}
	DeveloperCredentials struct {
		ActiveKeyID      string                 `default:"development" env:"APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_ID"`
		ActiveKeyVersion configvalue.KeyVersion `default:"1" env:"APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_VERSION"`
		Keys             string                 `default:"{\"development@1\":\"ZGV2ZWxvcGVyLWNyZWRlbnRpYWwtZGV2LWtleS0wMDE=\"}" env:"APP_API_CREDENTIAL_HMAC_KEYS"`
	}
	DeveloperOAuth struct {
		AccessTokenSigningKey string        `default:"oauth-development-access-sign-01" env:"APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY"`
		ActiveDigestKeyID     string        `default:"development" env:"APP_OAUTH_TOKEN_HMAC_ACTIVE_KEY_ID"`
		DigestKeys            string        `default:"{\"development\":\"b2F1dGgtZGV2ZWxvcG1lbnQtZGlnZXN0LWtleS0wMDE=\"}" env:"APP_OAUTH_TOKEN_HMAC_KEYS"`
		DynamicClientTTL      time.Duration `default:"720h" env:"APP_OAUTH_DYNAMIC_CLIENT_TTL"`
	}
	Web struct {
		APIHost                  string        `default:"localhost:8000" env:"APP_API_HOST"`
		PublicURL                string        `default:"http://localhost:8000" env:"APP_API_PUBLIC_URL"`
		CORSAllowedOrigins       string        `default:"http://localhost:3000" env:"APP_API_CORS_ALLOWED_ORIGINS"`
		ReadHeaderTimeout        time.Duration `default:"10s" env:"APP_API_READ_HEADER_TIMEOUT"`
		ReadTimeout              time.Duration `default:"5m" env:"APP_API_READ_TIMEOUT"`
		WriteTimeout             time.Duration `default:"5m" env:"APP_API_WRITE_TIMEOUT"`
		IdleTimeout              time.Duration `default:"60s" env:"APP_API_IDLE_TIMEOUT"`
		ShutdownTimeout          time.Duration `default:"30s" env:"APP_API_SHUTDOWN_TIMEOUT"`
		ReadinessCheckTimeout    time.Duration `default:"2s" env:"APP_API_READINESS_CHECK_TIMEOUT"`
		TelemetryShutdownTimeout time.Duration `default:"5s" env:"APP_API_TELEMETRY_SHUTDOWN_TIMEOUT"`
		DebugHost                string        `default:"localhost:9000" env:"APP_API_DEBUG_HOST"`
	}
	MCP struct {
		LoginURL string `default:"http://localhost:3000" env:"APP_MCP_LOGIN_URL"`
	}
	DB    databaseConfig
	Cache struct {
		Host         string        `default:"localhost" env:"APP_REDIS_HOST"`
		Port         string        `default:"6379" env:"APP_REDIS_PORT"`
		Password     string        `default:"" env:"APP_REDIS_PASSWORD"`
		Name         int           `default:"0" env:"APP_REDIS_DB"`
		DisableTLS   bool          `default:"false" env:"APP_REDIS_DISABLE_TLS"`
		DialTimeout  time.Duration `default:"10s" env:"APP_REDIS_DIAL_TIMEOUT"`
		ReadTimeout  time.Duration `default:"30s" env:"APP_REDIS_READ_TIMEOUT"`
		WriteTimeout time.Duration `default:"30s" env:"APP_REDIS_WRITE_TIMEOUT"`
		PoolSize     int           `default:"100" env:"APP_REDIS_POOL_SIZE"`
	}
	Email struct {
		Host        string `default:"smtp.gmail.com" env:"APP_EMAIL_HOST"`
		Port        int    `default:"587" env:"APP_EMAIL_PORT"`
		Username    string `env:"APP_EMAIL_USERNAME"`
		Password    string `env:"APP_EMAIL_PASSWORD"`
		FromAddress string `env:"APP_EMAIL_FROM_ADDRESS"`
		FromName    string `default:"FortyOne" env:"APP_EMAIL_FROM_NAME"`
		MayaAddress string `default:"maya@fortyone.app" env:"APP_EMAIL_MAYA_FROM_ADDRESS"`
		MayaName    string `default:"Maya, AI Agent" env:"APP_EMAIL_MAYA_FROM_NAME"`
		Environment string `default:"development" env:"APP_EMAIL_ENVIRONMENT"`
		BaseDir     string `default:"." env:"APP_EMAIL_BASE_DIR"`
	}
	Brevo struct {
		APIKey string `env:"APP_BREVO_API_KEY"`
	}
	Tracing struct {
		Endpoint string            `default:"localhost:4318" env:"APP_TRACING_ENDPOINT"`
		Headers  map[string]string `default:"{}" env:"APP_TRACING_HEADERS"`
	}
	Google struct {
		ClientID string `env:"GOOGLE_CLIENT_ID"`
	}
	Website struct {
		URL string `default:"http://localhost:3000" env:"APP_WEBSITE_URL"`
	}
	Storage struct {
		Provider          string `env:"APP_STORAGE_PROVIDER" default:"aws"`
		ProfilesBucket    string `env:"STORAGE_PROFILE_IMAGES_NAME" default:"profiles"`
		LogosBucket       string `env:"STORAGE_WORKSPACE_LOGOS_NAME" default:"logos"`
		AttachmentsBucket string `env:"STORAGE_ATTACHMENTS_NAME" default:"attachments"`
	}
	Azure struct {
		StorageConnectionString string `env:"APP_AZURE_STORAGE_CONNECTION_STRING"`
		StorageAccountName      string `env:"APP_AZURE_STORAGE_ACCOUNT_NAME"`
		StorageAccountKey       string `env:"APP_AZURE_STORAGE_ACCOUNT_KEY"`
	}
	AWS struct {
		AccessKeyID     string `env:"APP_AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `env:"APP_AWS_SECRET_ACCESS_KEY"`
		Region          string `env:"APP_AWS_REGION" default:"us-east-1"`
		Endpoint        string `env:"APP_AWS_ENDPOINT"`
		PublicURL       string `env:"APP_AWS_PUBLIC_URL"`
		ForcePathStyle  bool   `default:"false" env:"APP_AWS_FORCE_PATH_STYLE"`
		Bucket          string `env:"APP_AWS_BUCKET" default:"fortyone"`
	}
	Stripe struct {
		SecretKey     string `env:"STRIPE_SECRET_KEY"`
		WebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
	}
	GitHub struct {
		AppID                int64  `env:"APP_GITHUB_APP_ID"`
		AppSlug              string `env:"GITHUB_APP_SLUG"`
		ClientID             string `env:"GITHUB_CLIENT_ID"`
		ClientSecret         string `env:"GITHUB_CLIENT_SECRET"`
		PrivateKeyBase64     string `env:"GITHUB_PRIVATE_KEY_BASE64"`
		RedirectURL          string `env:"GITHUB_REDIRECT_URL"`
		WebhookURL           string `env:"GITHUB_WEBHOOK_URL"`
		WebhookSecret        string `env:"GITHUB_WEBHOOK_SECRET"`
		WebhookPayloadSecret string `default:"development-only-github-webhook-payload-secret" env:"APP_GITHUB_WEBHOOK_PAYLOAD_SECRET"`
	}
	Slack struct {
		ClientID             string `env:"SLACK_CLIENT_ID"`
		ClientSecret         string `env:"SLACK_CLIENT_SECRET"`
		SigningSecret        string `env:"SLACK_SIGNING_SECRET"`
		RedirectURL          string `default:"https://api.fortyone.app/integrations/slack/setup" env:"SLACK_REDIRECT_URL"`
		WebhookPayloadSecret string `default:"development-only-slack-webhook-payload-secret" env:"APP_SLACK_WEBHOOK_PAYLOAD_SECRET"`
	}
	Figma struct {
		ClientID             string `env:"FIGMA_CLIENT_ID"`
		ClientSecret         string `env:"FIGMA_CLIENT_SECRET"`
		RedirectURL          string `default:"https://api.fortyone.app/integrations/figma/callback" env:"FIGMA_REDIRECT_URL"`
		WebhookURL           string `default:"https://api.fortyone.app/webhooks/figma" env:"FIGMA_WEBHOOK_URL"`
		WebhookPayloadSecret string `default:"development-only-figma-webhook-payload-secret" env:"APP_FIGMA_WEBHOOK_PAYLOAD_SECRET"`
	}
}

func verificationTokenConfig(cfg Config) users.VerificationTokenConfig {
	return users.VerificationTokenConfig{
		Current: users.VerificationTokenKey{
			ID:     cfg.VerificationTokens.HMACKeyID,
			Secret: cfg.VerificationTokens.HMACKey,
		},
	}
}

func invitationTokenConfig(cfg Config) (invitations.InvitationTokenConfig, error) {
	previous, err := parseInvitationTokenKeys(cfg.InvitationTokens.PreviousKeys)
	if err != nil {
		return invitations.InvitationTokenConfig{}, err
	}
	return invitations.InvitationTokenConfig{
		Current: invitations.InvitationTokenKey{
			ID:     cfg.InvitationTokens.HMACKeyID,
			Secret: cfg.InvitationTokens.HMACKey,
		},
		Previous: previous,
	}, nil
}

func parseInvitationTokenKeys(input string) ([]invitations.InvitationTokenKey, error) {
	entries := parseCommaSeparated(input)
	keys := make([]invitations.InvitationTokenKey, 0, len(entries))
	for _, entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, errors.New("APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS must contain comma-separated key-id=secret entries")
		}
		keys = append(keys, invitations.InvitationTokenKey{
			ID:     strings.TrimSpace(parts[0]),
			Secret: strings.TrimSpace(parts[1]),
		})
	}
	return keys, nil
}

func parseCommaSeparated(input string) []string {
	parts := strings.Split(input, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values
}
