package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/handlers"
	"github.com/complexus-tech/projects-api/internal/mux"
	"github.com/complexus-tech/projects-api/internal/repo/notificationsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/consumer"
	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/go-playground/validator/v10"
	"github.com/josemukorivo/config"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82/client"
)

var (
	service = "projects-api"
	version = "0.0.1"
	environ = "development"
)

type Config struct {
	Auth struct {
		SecretKey string `default:"secret" env:"APP_AUTH_SECRET_KEY"`
	}
	Web struct {
		APIHost         string        `default:"localhost:8000" env:"APP_API_HOST"`
		ReadTimeout     time.Duration `default:"120s" env:"APP_API_READ_TIMEOUT"`
		WriteTimeout    time.Duration `default:"10s" env:"APP_API_WRITE_TIMEOUT"`
		IdleTimeout     time.Duration `default:"10s" env:"APP_API_IDLE_TIMEOUT"`
		ShutdownTimeout time.Duration `default:"30s" env:"APP_API_SHUTDOWN_TIMEOUT"`
		DebugHost       string        `default:"localhost:9000" env:"APP_API_DEBUG_HOST"`
	}
	DB struct {
		Host         string `default:"localhost"`
		Port         string `default:"5432" env:"APP_DB_PORT"`
		User         string `default:"postgres"`
		Password     string `default:"password"`
		Name         string `default:"complexus"`
		MaxIdleConns int    `default:"25" env:"APP_DB_MAX_IDLE_CONNS"`
		MaxOpenConns int    `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
		DisableTLS   bool   `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
	Cache struct {
		Host       string `default:"localhost" env:"APP_REDIS_HOST"`
		Port       string `default:"6379" env:"APP_REDIS_PORT"`
		Password   string `default:"" env:"APP_REDIS_PASSWORD"`
		Name       int    `default:"0" env:"APP_REDIS_DB"`
		DisableTLS bool   `default:"false" env:"APP_REDIS_DISABLE_TLS"`
	}
	Email struct {
		Host        string `default:"smtp.gmail.com" env:"APP_EMAIL_HOST"`
		Port        int    `default:"587" env:"APP_EMAIL_PORT"`
		Username    string `env:"APP_EMAIL_USERNAME"`
		Password    string `env:"APP_EMAIL_PASSWORD"`
		FromAddress string `env:"APP_EMAIL_FROM_ADDRESS"`
		FromName    string `default:"Complexus" env:"APP_EMAIL_FROM_NAME"`
		Environment string `default:"development" env:"APP_EMAIL_ENVIRONMENT"`
		BaseDir     string `default:"." env:"APP_EMAIL_BASE_DIR"`
	}
	Tracing struct {
		Host string `default:"localhost:4318" env:"APP_TRACING_HOST"`
	}
	Google struct {
		ClientID string `conf:"required,env:GOOGLE_CLIENT_ID"`
	}
	Website struct {
		URL string `default:"http://qa.localhost:3000" env:"APP_WEBSITE_URL"`
	}
	Azure struct {
		StorageConnectionString string `env:"APP_AZURE_STORAGE_CONNECTION_STRING"`
		StorageAccountName      string `env:"APP_AZURE_STORAGE_ACCOUNT_NAME"`
		StorageAccountKey       string `env:"APP_AZURE_STORAGE_ACCOUNT_KEY"`
		ProfileImagesContainer  string `default:"profile-images" env:"APP_AZURE_CONTAINER_PROFILE_IMAGES"`
		WorkspaceLogosContainer string `default:"workspace-logos" env:"APP_AZURE_CONTAINER_WORKSPACE_LOGOS"`
		AttachmentsContainer    string `default:"attachments" env:"APP_AZURE_CONTAINER_ATTACHMENTS"`
	}
	Stripe struct {
		SecretKey     string `env:"STRIPE_SECRET_KEY"`
		WebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
	}
}

func main() {

	var logLevel slog.Level

	switch environ {
	case "development":
		logLevel = slog.LevelDebug
	case "production":
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelInfo
	}

	log := logger.NewWithText(os.Stdout, logLevel, service)
	ctx := context.Background()

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
	// Connect to postgres database
	db, err := database.Open(database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})

	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	log.Info(ctx, fmt.Sprintf("connected to database `%s`", cfg.DB.Name))

	defer func() {
		log.Info(ctx, "closing the database connection")
		db.Close()
	}()

	// Connect to redis client
	var tlsConfig *tls.Config
	if !cfg.Cache.DisableTLS {
		tlsConfig = &tls.Config{}
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:      net.JoinHostPort(cfg.Cache.Host, cfg.Cache.Port),
		Password:  cfg.Cache.Password,
		DB:        cfg.Cache.Name,
		TLSConfig: tlsConfig,
	})

	// Close the redis connection when the main function returns
	defer rdb.Close()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("error pinging redis: %w", err)
	}
	log.Info(ctx, fmt.Sprintf("connected to redis database `%d`", cfg.Cache.Name))

	// Initialize cache service
	cacheService := cache.New(rdb, log)
	log.Info(ctx, "initialized cache service")

	// Initialize validator
	validate := validator.New()

	// Initialize Azure configuration
	azureConfig := azure.Config{
		ConnectionString:        cfg.Azure.StorageConnectionString,
		StorageAccountName:      cfg.Azure.StorageAccountName,
		AccountKey:              cfg.Azure.StorageAccountKey,
		ProfileImagesContainer:  cfg.Azure.ProfileImagesContainer,
		WorkspaceLogosContainer: cfg.Azure.WorkspaceLogosContainer,
		AttachmentsContainer:    cfg.Azure.AttachmentsContainer,
	}

	// Initialize email service
	emailService, err := email.NewService(email.Config{
		Host:        cfg.Email.Host,
		Port:        cfg.Email.Port,
		Username:    cfg.Email.Username,
		Password:    cfg.Email.Password,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
		Environment: cfg.Email.Environment,
		BaseDir:     cfg.Email.BaseDir,
	}, log)
	if err != nil {
		return fmt.Errorf("error initializing email service: %w", err)
	}
	log.Info(ctx, "email service initialized")

	// Create publisher
	publisher := publisher.New(rdb, log)

	// Create services
	notificationService := notifications.New(log, notificationsrepo.New(log, db))
	storiesService := stories.New(log, storiesrepo.New(log, db), publisher)
	objectivesService := objectives.New(log, objectivesrepo.New(log, db))
	usersService := users.New(log, usersrepo.New(log, db))
	statusesService := states.New(log, statesrepo.New(log, db))

	// Create consumer using Redis Streams
	consumer := consumer.New(rdb, db, log, cfg.Website.URL, notificationService, emailService, storiesService, objectivesService, usersService, statusesService)

	// Start consumer in a goroutine
	go func() {
		log.Info(ctx, "starting redis stream consumer")
		if err := consumer.Start(ctx); err != nil {
			log.Error(ctx, "failed to start consumer", "error", err)
		}
	}()

	// Initialize job scheduler
	jobScheduler, err := jobs.NewScheduler(db, log)
	if err != nil {
		return fmt.Errorf("error initializing job scheduler: %w", err)
	}

	// Start scheduler in a goroutine
	go func() {
		log.Info(ctx, "starting background job scheduler")
		jobScheduler.Start()
	}()

	// Register for shutdown in defer cleanup
	defer func() {
		log.Info(ctx, "shutting down job scheduler")
		if err := jobScheduler.Shutdown(); err != nil {
			log.Error(ctx, "error shutting down job scheduler", "error", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)

	// Start Tracing
	t := tracing.New(service, version, environ, cfg.Tracing.Host)
	tp, err := t.StartTracing()
	if err != nil {
		return fmt.Errorf("error starting tracing: %w", err)
	}
	log.Info(ctx, fmt.Sprintf("started open telemetry tracing on %s", cfg.Tracing.Host))

	// Graceful shutdown of tracing if server is stopped
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Error(ctx, "error shutting down tracer provider", "error", err)
		}
	}()

	tracer := tp.Tracer(service)

	// Initialize Google service
	googleService, err := google.NewService(cfg.Google.ClientID)
	if err != nil {
		return fmt.Errorf("error initializing google service: %w", err)
	}
	log.Info(ctx, "google auth service initialized")

	// Initialize Stripe client
	stripeClient := client.New(cfg.Stripe.SecretKey, nil)

	// Initialize SSE Hub
	sseHub := sse.NewHub(ctx, log, rdb)
	go sseHub.Run()
	log.Info(ctx, "SSE Hub initialized and started")

	// Update mux configuration
	muxConfig := mux.Config{
		DB:            db,
		Redis:         rdb,
		Publisher:     publisher,
		Shutdown:      shutdown,
		Log:           log,
		Tracer:        tracer,
		SecretKey:     cfg.Auth.SecretKey,
		EmailService:  emailService,
		GoogleService: googleService,
		Validate:      validate,
		AzureConfig:   azureConfig,
		Cache:         cacheService,
		StripeClient:  stripeClient,
		WebhookSecret: cfg.Stripe.WebhookSecret,
		SSEHub:        sseHub,
		CorsOrigin:    "*",
	}

	// Create the mux
	routeAdder := handlers.New()
	handler := mux.New(muxConfig, routeAdder)

	server := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handler,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info(ctx, "started http server on :8000")
		serverErrors <- server.ListenAndServe()
	}()

	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error: listening and serving: %s", err)

	case <-shutdown:
		log.Info(ctx, "starting shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Info(ctx, "error: graceful shutdown: %s", err)
			if err := server.Close(); err != nil {
				log.Info(ctx, "error: could not stop http server: %s", err)
			}
		}
	}
	return nil
}
