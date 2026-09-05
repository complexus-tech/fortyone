package workerbootstrap

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"strings"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/internal/platform/appkeys"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/josemukorivo/config"
)

var defaultQueues = map[string]int{
	"critical":      6,
	"default":       3,
	"integrations":  2,
	"low":           1,
	"onboarding":    5,
	"cleanup":       2,
	"notifications": 4,
	"automation":    3,
}

type HTTPConfig struct {
	Host              string        `default:"0.0.0.0:8080" env:"APP_WORKER_HTTP_HOST"`
	ReadHeaderTimeout time.Duration `default:"5s" env:"APP_WORKER_HTTP_READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `default:"10s" env:"APP_WORKER_HTTP_READ_TIMEOUT"`
	WriteTimeout      time.Duration `default:"15s" env:"APP_WORKER_HTTP_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `default:"60s" env:"APP_WORKER_HTTP_IDLE_TIMEOUT"`
	ShutdownTimeout   time.Duration `default:"15s" env:"APP_WORKER_HTTP_SHUTDOWN_TIMEOUT"`
}

type MonitorConfig struct {
	Enabled  bool   `default:"false" env:"APP_WORKER_MONITOR_ENABLED"`
	Username string `env:"APP_WORKER_MONITOR_USERNAME"`
	Password string `env:"APP_WORKER_MONITOR_PASSWORD"`
}

type Config struct {
	Environment string `default:"development" env:"APP_ENVIRONMENT"`
	HTTP        HTTPConfig
	Monitor     MonitorConfig
	DB          struct {
		Host         string `default:"localhost" env:"APP_DB_HOST"`
		Port         string `default:"5432" env:"APP_DB_PORT"`
		User         string `default:"postgres" env:"APP_DB_USER"`
		Password     string `default:"password" env:"APP_DB_PASSWORD"`
		Name         string `default:"complexus" env:"APP_DB_NAME"`
		MaxOpenConns int    `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
		MinConns     int    `default:"0" env:"APP_DB_MIN_CONNS"`

		ConnectTimeout    time.Duration `default:"10s" env:"APP_DB_CONNECT_TIMEOUT"`
		MaxConnIdleTime   time.Duration `default:"30m" env:"APP_DB_MAX_CONN_IDLE_TIME"`
		MaxConnLifetime   time.Duration `default:"1h" env:"APP_DB_MAX_CONN_LIFETIME"`
		HealthCheckPeriod time.Duration `default:"1m" env:"APP_DB_HEALTH_CHECK_PERIOD"`
		SSLMode           string        `env:"APP_DB_SSL_MODE"`
		SSLRootCert       string        `env:"APP_DB_SSL_ROOT_CERT"`
		DisableTLS        bool          `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
	Redis struct {
		Host         string        `default:"localhost" env:"APP_REDIS_HOST"`
		Port         string        `default:"6379" env:"APP_REDIS_PORT"`
		Password     string        `default:"" env:"APP_REDIS_PASSWORD"`
		Name         int           `default:"0" env:"APP_REDIS_DB"`
		DisableTLS   bool          `default:"false" env:"APP_REDIS_DISABLE_TLS"`
		DialTimeout  time.Duration `default:"10s" env:"APP_REDIS_DIAL_TIMEOUT"`
		ReadTimeout  time.Duration `default:"30s" env:"APP_REDIS_READ_TIMEOUT"`
		WriteTimeout time.Duration `default:"30s" env:"APP_REDIS_WRITE_TIMEOUT"`
		PoolSize     int           `default:"30" env:"APP_REDIS_POOL_SIZE"`
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
	Auth struct {
		SecretKey                    string `default:"secret" env:"APP_AUTH_SECRET_KEY"`
		GoogleClientIDs              string `env:"APP_AUTH_GOOGLE_CLIENT_IDS"`
		GoogleClientSecret           string `env:"APP_AUTH_GOOGLE_CLIENT_SECRET"`
		GoogleCalendarRedirectURL    string `env:"APP_AUTH_GOOGLE_CALENDAR_REDIRECT_URL"`
		GoogleCalendarWebhookURL     string `env:"APP_AUTH_GOOGLE_CALENDAR_WEBHOOK_URL"`
		MicrosoftClientID            string `env:"APP_AUTH_MICROSOFT_CLIENT_ID"`
		MicrosoftClientSecret        string `env:"APP_AUTH_MICROSOFT_CLIENT_SECRET"`
		MicrosoftTenant              string `default:"common" env:"APP_AUTH_MICROSOFT_TENANT"`
		MicrosoftCalendarRedirectURL string `env:"APP_AUTH_MICROSOFT_CALENDAR_REDIRECT_URL"`
		MicrosoftCalendarWebhookURL  string `env:"APP_AUTH_MICROSOFT_CALENDAR_WEBHOOK_URL"`
	}
	Feedback struct {
		SecurityKey string `default:"development-only-feedback-security-key" env:"APP_FEEDBACK_SECURITY_KEY"`
	}
	EmailReply struct {
		SecurityKey string `default:"development-only-email-reply-security-key" env:"APP_EMAIL_REPLY_SECURITY_KEY"`
	}
	Messaging struct {
		MutationHMACKey string `default:"development-only-messaging-mutation-hmac-key" env:"APP_MESSAGING_MUTATION_HMAC_KEY"`
	}
	InvitationTokens struct {
		HMACKey      string `default:"development-only-invitation-hmac-key" env:"APP_INVITATION_TOKEN_HMAC_KEY"`
		HMACKeyID    string `default:"v1" env:"APP_INVITATION_TOKEN_HMAC_KEY_ID"`
		PreviousKeys string `env:"APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS"`
	}
	APIPublicURL string `default:"http://localhost:8000" env:"APP_API_PUBLIC_URL"`
	Website      struct {
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
	AIAPIKey           string `env:"OPENAI_API_KEY"`
	AIModel            string `default:"gpt-5.6-luna" env:"OPENAI_MODEL"`
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	MessagingAssistant struct {
		UserCallsPerMinute      int64 `default:"12" env:"OPENAI_ASSISTANT_USER_CALLS_PER_MINUTE"`
		WorkspaceCallsPerMinute int64 `default:"120" env:"OPENAI_ASSISTANT_WORKSPACE_CALLS_PER_MINUTE"`
		WorkspaceTokensPerDay   int64 `default:"1000000" env:"OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY"`
	}
	Slack struct {
		ClientID      string `env:"SLACK_CLIENT_ID"`
		ClientSecret  string `env:"SLACK_CLIENT_SECRET"`
		SigningSecret string `env:"SLACK_SIGNING_SECRET"`
	}
	GitHub struct {
		AppID            int64  `env:"APP_GITHUB_APP_ID"`
		AppSlug          string `env:"GITHUB_APP_SLUG"`
		PrivateKeyBase64 string `env:"GITHUB_PRIVATE_KEY_BASE64"`
		RedirectURL      string `env:"GITHUB_REDIRECT_URL"`
		WebhookSecret    string `env:"GITHUB_WEBHOOK_SECRET"`
	}
	Figma struct {
		ClientID     string `env:"FIGMA_CLIENT_ID"`
		ClientSecret string `env:"FIGMA_CLIENT_SECRET"`
		RedirectURL  string `default:"https://api.fortyone.app/integrations/figma/callback" env:"FIGMA_REDIRECT_URL"`
		WebhookURL   string `default:"https://api.fortyone.app/webhooks/figma" env:"FIGMA_WEBHOOK_URL"`
	}
	Queues map[string]int `default:"{\"critical\":6,\"default\":3,\"integrations\":2,\"low\":1,\"onboarding\":5,\"cleanup\":2,\"notifications\":4,\"automation\":3}"`
}

func loadConfig() (Config, error) {
	var cfg Config
	if err := config.Parse("app", &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Queues == nil {
		cfg.Queues = cloneQueueConfig(defaultQueues)
	}
	cfg.Feedback.SecurityKey = strings.TrimSpace(cfg.Feedback.SecurityKey)
	return cfg, nil
}

func cloneQueueConfig(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	maps.Copy(dst, src)
	return dst
}

func validateRuntimeConfig(cfg Config) (deployment.Mode, error) {
	mode, err := deployment.Parse(cfg.Environment)
	if err != nil {
		return "", fmt.Errorf("validate worker configuration: %w", err)
	}

	sslMode, databaseErr := platformdatabase.EffectiveSSLMode(platformdatabase.Config{
		SSLMode:     cfg.DB.SSLMode,
		SSLRootCert: cfg.DB.SSLRootCert,
		DisableTLS:  cfg.DB.DisableTLS,
	})
	secretRequirements := []deployment.SecretRequirement{{
		Name:            "APP_AUTH_SECRET_KEY",
		Value:           cfg.Auth.SecretKey,
		ForbiddenValues: []string{"secret"},
	}, {
		Name:            "APP_FEEDBACK_SECURITY_KEY",
		Value:           cfg.Feedback.SecurityKey,
		ForbiddenValues: []string{"development-only-feedback-security-key"},
	}, {
		Name:            "APP_EMAIL_REPLY_SECURITY_KEY",
		Value:           cfg.EmailReply.SecurityKey,
		ForbiddenValues: []string{"development-only-email-reply-security-key"},
	}, {
		Name:            "APP_MESSAGING_MUTATION_HMAC_KEY",
		Value:           cfg.Messaging.MutationHMACKey,
		ForbiddenValues: []string{"development-only-messaging-mutation-hmac-key"},
	}, {
		Name:            "APP_INVITATION_TOKEN_HMAC_KEY",
		Value:           cfg.InvitationTokens.HMACKey,
		ForbiddenValues: []string{"development-only-invitation-hmac-key"},
	}}
	if cfg.Monitor.Enabled {
		secretRequirements = append(secretRequirements, deployment.SecretRequirement{
			Name:  "APP_WORKER_MONITOR_PASSWORD",
			Value: cfg.Monitor.Password,
		})
	}
	secretErr := deployment.ValidateProductionSecrets(mode, secretRequirements...)
	transportErr := deployment.ValidateProductionTransports(mode, deployment.TransportSecurity{
		PostgreSQLSSLMode: sslMode,
		RedisTLSDisabled:  cfg.Redis.DisableTLS,
	})
	awsCredentialErr := deployment.ValidateAWSCredentialSource(
		mode,
		cfg.AWS.AccessKeyID,
		cfg.AWS.SecretAccessKey,
	)
	httpErr := validateHTTPConfig(cfg.HTTP, cfg.Monitor)
	_, integrationKeyErr := appkeys.NewIntegrationKeys(cfg.Auth.SecretKey)
	invitationConfig, invitationConfigErr := workerInvitationTokenConfig(cfg)
	var invitationTokenErr error
	if invitationConfigErr == nil {
		_, invitationTokenErr = invitations.NewInvitationTokenManager(invitationConfig)
	}
	keySeparationErr := validateWorkerSecurityKeySeparation(mode, cfg, invitationConfig)

	if err := errors.Join(
		databaseErr,
		secretErr,
		transportErr,
		awsCredentialErr,
		httpErr,
		integrationKeyErr,
		invitationConfigErr,
		invitationTokenErr,
		keySeparationErr,
	); err != nil {
		return "", fmt.Errorf("validate worker configuration: %w", err)
	}
	return mode, nil
}

func workerInvitationTokenConfig(cfg Config) (invitations.InvitationTokenConfig, error) {
	entries := strings.Split(cfg.InvitationTokens.PreviousKeys, ",")
	previous := make([]invitations.InvitationTokenKey, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return invitations.InvitationTokenConfig{}, errors.New("APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS must contain comma-separated key-id=secret entries")
		}
		previous = append(previous, invitations.InvitationTokenKey{
			ID:     strings.TrimSpace(parts[0]),
			Secret: strings.TrimSpace(parts[1]),
		})
	}
	return invitations.InvitationTokenConfig{
		Current: invitations.InvitationTokenKey{
			ID:     cfg.InvitationTokens.HMACKeyID,
			Secret: cfg.InvitationTokens.HMACKey,
		},
		Previous: previous,
	}, nil
}

func validateWorkerSecurityKeySeparation(
	mode deployment.Mode,
	cfg Config,
	invitationConfig invitations.InvitationTokenConfig,
) error {
	if !mode.IsProduction() {
		return nil
	}
	keys := append([]invitations.InvitationTokenKey{invitationConfig.Current}, invitationConfig.Previous...)
	seen := make(map[string]string) // #nosec G101 -- Runtime secret equality set; map values are configuration-key names.
	for _, candidate := range []struct {
		name  string
		value string
	}{
		{name: "APP_AUTH_SECRET_KEY", value: cfg.Auth.SecretKey},
		{name: "APP_FEEDBACK_SECURITY_KEY", value: cfg.Feedback.SecurityKey},
		{name: "APP_EMAIL_REPLY_SECURITY_KEY", value: cfg.EmailReply.SecurityKey},
		{name: "APP_MESSAGING_MUTATION_HMAC_KEY", value: cfg.Messaging.MutationHMACKey},
		{name: "GITHUB_PRIVATE_KEY_BASE64", value: cfg.GitHub.PrivateKeyBase64},
		{name: "GITHUB_WEBHOOK_SECRET", value: cfg.GitHub.WebhookSecret},
		{name: "SLACK_CLIENT_SECRET", value: cfg.Slack.ClientSecret},
		{name: "SLACK_SIGNING_SECRET", value: cfg.Slack.SigningSecret},
		{name: "FIGMA_CLIENT_SECRET", value: cfg.Figma.ClientSecret},
	} {
		secret := strings.TrimSpace(candidate.value)
		if secret == "" {
			continue
		}
		if existing, exists := seen[secret]; exists {
			return fmt.Errorf("%s must not reuse %s", candidate.name, existing)
		}
		seen[secret] = candidate.name
	}
	for _, key := range keys {
		secret := strings.TrimSpace(key.Secret)
		if secret == "" {
			continue
		}
		name := "APP_INVITATION_TOKEN_HMAC_KEY"
		if key.ID != invitationConfig.Current.ID {
			name = "APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS[" + key.ID + "]"
		}
		if existing, exists := seen[secret]; exists {
			return fmt.Errorf("%s must not reuse %s", name, existing)
		}
		seen[secret] = name
	}
	return nil
}

func validateHTTPConfig(httpConfig HTTPConfig, monitorConfig MonitorConfig) error {
	var validationErrors []error
	host := strings.TrimSpace(httpConfig.Host)
	if host == "" {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_HOST is required"))
	} else if _, _, err := net.SplitHostPort(host); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("APP_WORKER_HTTP_HOST must include a valid host and port: %w", err))
	}
	if httpConfig.ReadHeaderTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_READ_HEADER_TIMEOUT must be positive"))
	}
	if httpConfig.ReadTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_READ_TIMEOUT must be positive"))
	}
	if httpConfig.WriteTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_WRITE_TIMEOUT must be positive"))
	}
	if httpConfig.IdleTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_IDLE_TIMEOUT must be positive"))
	}
	if httpConfig.ShutdownTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("APP_WORKER_HTTP_SHUTDOWN_TIMEOUT must be positive"))
	}
	if monitorConfig.Enabled {
		if strings.TrimSpace(monitorConfig.Username) == "" {
			validationErrors = append(validationErrors, errors.New("APP_WORKER_MONITOR_USERNAME is required when queue monitoring is enabled"))
		}
		if strings.TrimSpace(monitorConfig.Password) == "" {
			validationErrors = append(validationErrors, errors.New("APP_WORKER_MONITOR_PASSWORD is required when queue monitoring is enabled"))
		}
	}
	return errors.Join(validationErrors...)
}
