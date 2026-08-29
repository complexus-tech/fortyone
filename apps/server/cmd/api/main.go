package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	bootstrapapi "github.com/complexus-tech/projects-api/internal/bootstrap/api"
	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	"github.com/complexus-tech/projects-api/internal/migrations"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	actorsrepository "github.com/complexus-tech/projects-api/internal/platform/actors/repository"
	"github.com/complexus-tech/projects-api/internal/platform/appkeys"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/josemukorivo/config"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82/client"
)

var (
	service = "projects-api"
	version = "development"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "Run database migrations and exit")
	flag.Parse()

	log := logger.NewWithJSON(os.Stdout, configuredLogLevel(), service)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *migrateOnly {
		if err := runMigrations(ctx, log); err != nil {
			log.Error(ctx, fmt.Sprintf("migrations failed: %s", err))
			os.Exit(1)
		}
		log.Info(ctx, "migrations completed")
		return
	}

	if err := run(ctx, log); err != nil {
		log.Error(ctx, fmt.Sprintf("error shutting down: %s", err))
		os.Exit(1)
	}
}

// run starts the HTTP server, tracing and listens for OS signals to gracefully shutdown.
func run(ctx context.Context, log *logger.Logger) error {
	var cfg Config

	err := config.Parse("app", &cfg)
	if err != nil {
		return fmt.Errorf("error parsing config: %s", err)
	}
	cfg.Feedback.IngressSecret = strings.TrimSpace(cfg.Feedback.IngressSecret)
	cfg.Feedback.SecurityKey = strings.TrimSpace(cfg.Feedback.SecurityKey)
	if len(cfg.Feedback.IngressSecret) < 32 {
		return fmt.Errorf("FEEDBACK_INGRESS_SECRET must contain at least 32 characters")
	}
	mode, err := validateRuntimeConfig(cfg)
	if err != nil {
		return err
	}
	if _, err := providers.BuiltInRegistry(); err != nil {
		return fmt.Errorf("validate built-in integration providers: %w", err)
	}
	if err := validateAPIHTTPConfig(cfg); err != nil {
		return fmt.Errorf("validate API HTTP configuration: %w", err)
	}
	originPolicy, err := web.NewOriginPolicy(cfg.Web.CORSAllowedOrigins)
	if err != nil {
		return fmt.Errorf("initialize CORS origin policy: %w", err)
	}
	log.Info(ctx, "starting API process", "version", version, "environment", mode.String())
	verificationTokens, err := users.NewVerificationTokenManager(verificationTokenConfig(cfg))
	if err != nil {
		return fmt.Errorf("initialize verification token security: %w", err)
	}
	invitationTokenConfig, err := invitationTokenConfig(cfg)
	if err != nil {
		return fmt.Errorf("initialize invitation token security: %w", err)
	}
	invitationTokens, err := invitations.NewInvitationTokenManager(invitationTokenConfig)
	if err != nil {
		return fmt.Errorf("initialize invitation token security: %w", err)
	}
	integrationKeys, err := appkeys.NewIntegrationKeys(cfg.Auth.SecretKey)
	if err != nil {
		return fmt.Errorf("initialize integration security: %w", err)
	}
	credentialVault := integrationKeys.CredentialVault
	developerCredentialTokens, err := newDeveloperCredentialTokenManager(cfg)
	if err != nil {
		return fmt.Errorf("initialize developer credential security: %w", err)
	}
	lifecycleOwnsResources := false

	// Connect to postgres database
	connections, err := platformdatabase.Open(ctx, platformdatabase.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MinConns:     cfg.DB.MinConns,

		ConnectTimeout:    cfg.DB.ConnectTimeout,
		MaxConnIdleTime:   cfg.DB.MaxConnIdleTime,
		MaxConnLifetime:   cfg.DB.MaxConnLifetime,
		HealthCheckPeriod: cfg.DB.HealthCheckPeriod,
		SSLMode:           cfg.DB.SSLMode,
		SSLRootCert:       cfg.DB.SSLRootCert,
		DisableTLS:        cfg.DB.DisableTLS,
	})

	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	defer func() {
		if lifecycleOwnsResources {
			return
		}
		log.Info(ctx, "closing the database connection")
		if err := connections.Close(); err != nil {
			log.Error(ctx, "error closing the database connection", "error", err)
		}
	}()

	developerOAuthServices, err := newDeveloperOAuthServices(cfg, connections.Pool)
	if err != nil {
		return fmt.Errorf("initialize developer OAuth security: %w", err)
	}

	log.Info(ctx, fmt.Sprintf("connected to database `%s`", cfg.DB.Name))

	// Connect to redis client
	rdb := redis.NewClient(redisOptions(cfg))

	defer func() {
		if lifecycleOwnsResources {
			return
		}
		if err := rdb.Close(); err != nil {
			log.Error(ctx, "error closing Redis client", "error", err)
		}
	}()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("error pinging redis: %w", err)
	}
	log.Info(ctx, fmt.Sprintf("connected to redis database `%d`", cfg.Cache.Name))

	// Initialize cache service
	cacheService := cache.New(rdb, log)
	log.Info(ctx, "initialized cache service")

	// Initialize Azure configuration
	azureConfig := azure.Config{
		ConnectionString:   cfg.Azure.StorageConnectionString,
		StorageAccountName: cfg.Azure.StorageAccountName,
		AccountKey:         cfg.Azure.StorageAccountKey,
	}

	awsConfig := aws.Config{
		AccessKeyID:     cfg.AWS.AccessKeyID,
		SecretAccessKey: cfg.AWS.SecretAccessKey,
		Region:          cfg.AWS.Region,
		Endpoint:        cfg.AWS.Endpoint,
		PublicURL:       cfg.AWS.PublicURL,
		ForcePathStyle:  cfg.AWS.ForcePathStyle,
		Bucket:          cfg.AWS.Bucket,
	}

	storageConfig := storage.Config{
		Provider:          cfg.Storage.Provider,
		ProfilesBucket:    cfg.Storage.ProfilesBucket,
		LogosBucket:       cfg.Storage.LogosBucket,
		AttachmentsBucket: cfg.Storage.AttachmentsBucket,
		Azure:             azureConfig,
		AWS:               awsConfig,
	}

	storageService, err := storage.NewStorageService(storageConfig, log)
	if err != nil {
		return fmt.Errorf("error initializing storage service: %w", err)
	}

	// Initialize mailer service
	mailerService, err := mailer.NewService(mailer.Config{
		Host:            cfg.Email.Host,
		Port:            cfg.Email.Port,
		Username:        cfg.Email.Username,
		Password:        cfg.Email.Password,
		FromAddress:     cfg.Email.FromAddress,
		FromName:        cfg.Email.FromName,
		MayaFromAddress: cfg.Email.MayaAddress,
		MayaFromName:    cfg.Email.MayaName,
		Environment:     cfg.Email.Environment,
		BaseDir:         cfg.Email.BaseDir,
	}, log)
	if err != nil {
		return fmt.Errorf("error initializing mailer service: %w", err)
	}
	log.Info(ctx, "mailer service initialized")

	// Initialize Brevo service
	brevoService, err := brevo.NewService(brevo.Config{
		APIKey: cfg.Brevo.APIKey,
	}, log)
	if err != nil {
		return fmt.Errorf("error initializing brevo service: %w", err)
	}

	// Create publisher
	publisher := publisher.New(rdb, log)

	// Initialize tasks service for Asynq
	tasksService, err := tasks.New(rdb, log)
	if err != nil {
		log.Error(ctx, "failed to initialize tasks service", "error", err)
		return fmt.Errorf("error initializing tasks service: %w", err)
	}
	log.Info(ctx, "tasks service initialized")

	actorResolver := actors.NewResolver(actorsrepository.New(connections.Pool))

	legacyShutdown := make(chan os.Signal, 1)

	// Start Tracing
	t := tracing.New(service, version, mode.String(), cfg.Tracing.Endpoint, cfg.Tracing.Headers)
	tp, err := t.StartTracing()
	if err != nil {
		return fmt.Errorf("error starting tracing: %w", err)
	}
	log.Info(ctx, fmt.Sprintf("started open telemetry tracing on %s", cfg.Tracing.Endpoint))

	// Graceful shutdown of tracing if server is stopped
	defer func() {
		if lifecycleOwnsResources {
			return
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Web.TelemetryShutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Error(ctx, "error shutting down tracer provider", "error", err)
		}
	}()

	tracer := tp.Tracer(service)

	googleClientIDs := parseCommaSeparated(cfg.Auth.GoogleClientIDs)
	if len(googleClientIDs) == 0 {
		if legacyClientID := strings.TrimSpace(cfg.Google.ClientID); legacyClientID != "" {
			googleClientIDs = []string{legacyClientID}
		}
	}

	// Initialize Google service
	googleService, err := google.NewService(google.Config{
		ClientIDs:           googleClientIDs,
		ClientSecret:        cfg.Auth.GoogleClientSecret,
		RedirectURL:         cfg.Auth.GoogleRedirectURL,
		CalendarRedirectURL: cfg.Auth.GoogleCalendarRedirectURL,
	})
	if err != nil {
		return fmt.Errorf("error initializing google service: %w", err)
	}
	log.Info(ctx, "google auth service initialized")
	microsoftService := microsoft.NewService(microsoft.Config{
		ClientID:            cfg.Auth.MicrosoftClientID,
		ClientSecret:        cfg.Auth.MicrosoftClientSecret,
		Tenant:              cfg.Auth.MicrosoftTenant,
		RedirectURL:         cfg.Auth.MicrosoftRedirectURL,
		CalendarRedirectURL: cfg.Auth.MicrosoftCalendarRedirectURL,
	})
	log.Info(ctx, "microsoft auth service initialized")

	// Initialize Stripe client
	stripeClient := client.New(cfg.Stripe.SecretKey, nil)

	githubUserID, err := actorResolver.Resolve(ctx, actors.KeyGitHub)
	if err != nil {
		return fmt.Errorf("resolve github actor: %w", err)
	}

	readiness, err := platformhealth.NewReadiness(
		cfg.Web.ReadinessCheckTimeout,
		platformhealth.Dependency{
			Name:  "postgres",
			Check: connections.Pool.Ping,
		},
		platformhealth.Dependency{
			Name: "redis",
			Check: func(checkCtx context.Context) error {
				return rdb.Ping(checkCtx).Err()
			},
		},
	)
	if err != nil {
		return fmt.Errorf("initialize API readiness: %w", err)
	}

	sseHub := sse.NewHub(log, rdb)

	// Update mux configuration
	muxConfig := mux.Config{
		Redis:                       rdb,
		Publisher:                   publisher,
		Shutdown:                    legacyShutdown,
		Log:                         log,
		Tracer:                      tracer,
		DeploymentMode:              mode,
		Readiness:                   readiness,
		SecretKey:                   cfg.Auth.SecretKey,
		FeedbackIngressSecret:       cfg.Feedback.IngressSecret,
		FeedbackSecurityKey:         cfg.Feedback.SecurityKey,
		EmailReplySecurityKey:       cfg.EmailReply.SecurityKey,
		MessagingMutationHMACKey:    cfg.Messaging.MutationHMACKey,
		CookieDomain:                cfg.Auth.CookieDomain,
		EmailService:                mailerService,
		BrevoService:                brevoService,
		GoogleService:               googleService,
		MicrosoftService:            microsoftService,
		GoogleCalendarWebhookURL:    cfg.Auth.GoogleCalendarWebhookURL,
		MicrosoftCalendarWebhookURL: cfg.Auth.MicrosoftCalendarWebhookURL,
		StorageConfig:               storageConfig,
		StorageService:              storageService,
		Cache:                       cacheService,
		TasksService:                tasksService,
		StripeClient:                stripeClient,
		WebhookSecret:               cfg.Stripe.WebhookSecret,
		WebsiteURL:                  cfg.Website.URL,
		APIPublicURL:                cfg.Web.PublicURL,
		MCPLoginURL:                 cfg.MCP.LoginURL,
		GitHubAppID:                 cfg.GitHub.AppID,
		GitHubAppSlug:               cfg.GitHub.AppSlug,
		GitHubClientID:              cfg.GitHub.ClientID,
		GitHubClientSecret:          cfg.GitHub.ClientSecret,
		GitHubUserID:                githubUserID,
		GitHubKeyBase64:             cfg.GitHub.PrivateKeyBase64,
		GitHubRedirect:              cfg.GitHub.RedirectURL,
		GitHubWebhook:               cfg.GitHub.WebhookSecret,
		GitHubWebhookPayloadSecret:  integrationKeys.GitHubWebhookPayloadSecret,
		SlackSigningSecret:          cfg.Slack.SigningSecret,
		SlackClientID:               cfg.Slack.ClientID,
		SlackClientSecret:           cfg.Slack.ClientSecret,
		SlackRedirectURL:            cfg.Slack.RedirectURL,
		SlackWebhookPayloadSecret:   integrationKeys.SlackWebhookPayloadSecret,
		FigmaClientID:               cfg.Figma.ClientID,
		FigmaClientSecret:           cfg.Figma.ClientSecret,
		FigmaRedirectURL:            cfg.Figma.RedirectURL,
		FigmaWebhookURL:             cfg.Figma.WebhookURL,
		FigmaWebhookPayloadSecret:   integrationKeys.FigmaWebhookPayloadSecret,
		AIAPIKey:                    strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		SSEHub:                      sseHub,
		AllowedOrigins:              originPolicy,
	}

	runtime, err := bootstrapapi.BuildRuntime(
		muxConfig,
		bootstrapapi.Dependencies{
			DatabasePool:               connections.Pool,
			VerificationTokens:         verificationTokens,
			InvitationTokens:           invitationTokens,
			CredentialVault:            credentialVault,
			DeveloperCredentialTokens:  developerCredentialTokens,
			DeveloperOAuthPlatform:     developerOAuthServices.Platform,
			DeveloperAPIOAuth:          developerOAuthServices.PublicAPI,
			DeveloperOAuthApplications: developerOAuthServices.Applications,
		},
		cfg.Website.URL,
		mailerService,
	)
	if err != nil {
		return fmt.Errorf("error building bootstrap runtime: %w", err)
	}

	handler := mux.New(muxConfig, runtime.RouteAdder)
	process, err := bootstrapapi.NewProcess(bootstrapapi.ProcessOptions{
		Config: bootstrapapi.ProcessConfig{
			Address:                  cfg.Web.APIHost,
			ReadHeaderTimeout:        cfg.Web.ReadHeaderTimeout,
			ReadTimeout:              cfg.Web.ReadTimeout,
			WriteTimeout:             cfg.Web.WriteTimeout,
			IdleTimeout:              cfg.Web.IdleTimeout,
			ShutdownTimeout:          cfg.Web.ShutdownTimeout,
			TelemetryShutdownTimeout: cfg.Web.TelemetryShutdownTimeout,
		},
		Handler:   handler,
		Log:       log,
		Readiness: readiness,
		Telemetry: tp,
		Components: []bootstrapapi.Component{
			{Name: "redis_stream_consumer", Runtime: runtime.Consumer},
			{Name: "sse_hub", Runtime: sseHub},
		},
		Resources: []bootstrapapi.Resource{
			{Name: "redis", Close: rdb.Close},
			{Name: "database", Close: connections.Close},
		},
	})
	if err != nil {
		return fmt.Errorf("initialize API process lifecycle: %w", err)
	}

	lifecycleOwnsResources = true
	return process.Run(ctx)
}

func runMigrations(ctx context.Context, log *logger.Logger) error {
	cfg, err := parseMigrationProcessConfig()
	if err != nil {
		return err
	}
	mode, err := deployment.Parse(cfg.Environment)
	if err != nil {
		return fmt.Errorf("validate migration configuration: %w", err)
	}
	sslMode, err := platformdatabase.EffectiveSSLMode(platformdatabase.Config{
		SSLMode:     cfg.DB.SSLMode,
		SSLRootCert: cfg.DB.SSLRootCert,
		DisableTLS:  cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("validate migration database transport: %w", err)
	}
	if err := deployment.ValidateProductionTransports(mode, deployment.TransportSecurity{
		PostgreSQLSSLMode: sslMode,
	}); err != nil {
		return fmt.Errorf("validate migration database transport: %w", err)
	}

	log.Info(ctx, "running database migrations")
	return migrations.Run(ctx, platformdatabase.MigrationConfig{
		Config: platformdatabase.Config{
			Host:           cfg.DB.Host,
			Port:           cfg.DB.Port,
			User:           cfg.DB.User,
			Password:       cfg.DB.Password,
			Name:           cfg.DB.Name,
			MaxOpenConns:   cfg.DB.MaxOpenConns,
			ConnectTimeout: cfg.DB.ConnectTimeout,
			SSLMode:        cfg.DB.SSLMode,
			SSLRootCert:    cfg.DB.SSLRootCert,
			DisableTLS:     cfg.DB.DisableTLS,
		},
		MaxIdleConns:     cfg.DB.MigrationMaxIdleConns,
		StatementTimeout: cfg.DB.MigrationStatementTimeout,
	})
}

func parseMigrationProcessConfig() (migrationProcessConfig, error) {
	var cfg migrationProcessConfig
	if err := config.Parse("app", &cfg); err != nil {
		return migrationProcessConfig{}, fmt.Errorf("parse migration configuration: %w", err)
	}
	return cfg, nil
}

func configuredLogLevel() slog.Level {
	value := strings.TrimSpace(os.Getenv("APP_ENVIRONMENT"))
	if value == "" {
		value = deployment.Development.String()
	}
	mode, err := deployment.Parse(value)
	if err == nil && mode == deployment.Development {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
