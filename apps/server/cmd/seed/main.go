package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	workspacebootstrap "github.com/complexus-tech/projects-api/internal/bootstrap/workspaces"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	workspaceuow "github.com/complexus-tech/projects-api/internal/modules/workspaces/uow"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/seeding"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/josemukorivo/config"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82/client"
)

// Replicate Config from main.go
type Config struct {
	DB struct {
		Host         string `default:"localhost"`
		Port         string `default:"5432" env:"APP_DB_PORT"`
		User         string `default:"postgres"`
		Password     string `default:"password"`
		Name         string `default:"complexus"`
		MaxOpenConns int    `default:"25" env:"APP_DB_MAX_OPEN_CONNS"`
		SSLMode      string `env:"APP_DB_SSL_MODE"`
		SSLRootCert  string `env:"APP_DB_SSL_ROOT_CERT"`
		DisableTLS   bool   `default:"true" env:"APP_DB_DISABLE_TLS"`
	}
	Cache struct {
		Host       string `default:"localhost" env:"APP_REDIS_HOST"`
		Port       string `default:"6379" env:"APP_REDIS_PORT"`
		Password   string `default:"" env:"APP_REDIS_PASSWORD"`
		Name       int    `default:"0" env:"APP_REDIS_DB"`
		DisableTLS bool   `default:"false" env:"APP_REDIS_DISABLE_TLS"`
	}
	Stripe struct {
		SecretKey     string `env:"STRIPE_SECRET_KEY"`
		WebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
	}
}

func main() {
	// CLI Flags
	name := flag.String("name", "Development", "Name of the workspace")
	slug := flag.String("slug", "dev", "Slug of the workspace")
	email := flag.String("email", "admin@example.com", "Email of the admin user")
	fullName := flag.String("fullname", "Admin User", "Full name of the admin user")

	ctx := context.Background()
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, "seeder")

	// Parse config
	var cfg Config
	if err := config.Parse("app", &cfg); err != nil {
		log.Error(ctx, "failed to parse config", "error", err)
		os.Exit(1)
	}
	disableTLS := flag.Bool(
		"disable-tls",
		cfg.DB.DisableTLS,
		"deprecated TLS fallback used only when APP_DB_SSL_MODE is empty",
	)
	flag.Parse()
	disableTLSWasSet := false
	flag.Visit(func(parsedFlag *flag.Flag) {
		if parsedFlag.Name == "disable-tls" {
			disableTLSWasSet = true
		}
	})
	if disableTLSWasSet && strings.TrimSpace(cfg.DB.SSLMode) != "" {
		log.Error(ctx, "--disable-tls cannot be combined with APP_DB_SSL_MODE")
		os.Exit(1)
	}

	databaseConfig := platformdatabase.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		SSLMode:      cfg.DB.SSLMode,
		SSLRootCert:  cfg.DB.SSLRootCert,
		DisableTLS:   *disableTLS,
	}
	effectiveSSLMode, err := platformdatabase.EffectiveSSLMode(databaseConfig)
	if err != nil {
		log.Error(ctx, "invalid database TLS configuration", "error", err)
		os.Exit(1)
	}

	log.Info(ctx, "database config",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"name", cfg.DB.Name,
		"user", cfg.DB.User,
		"ssl_mode", effectiveSSLMode,
	)

	// Connect to DB
	connections, err := platformdatabase.Open(ctx, databaseConfig)
	if err != nil {
		log.Error(ctx, "failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer connections.Close()

	// Connect to Redis (required for tasks and publisher)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Cache.Host, cfg.Cache.Port),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.Name,
	})
	defer rdb.Close()

	// Initialize Minimal Services
	tasksService, _ := tasks.New(rdb, log)
	publisher := publisher.New(rdb, log)

	// Dependency Tree
	usersRepo := usersrepository.New(connections.Pool)
	usersService := users.New(log, usersRepo, tasksService)

	teamsRepo := teamsrepository.New(connections.Pool)

	statesRepo := statesrepository.New(connections.Pool)
	statesService := states.New(statesRepo)

	mayaRepository := mayarepository.New(connections.Pool)
	storiesRepo := storiesrepository.New(log, connections.Pool)
	storiesService := stories.New(log, storiesRepo, publisher, tasksService)
	storiesService.ConfigureAutoSchedulingEligibility(mayaRepository.WorkspaceCanUseMaya)

	objectivesRepo := objectivesrepository.New(connections.Pool)
	_ = objectives.New(log, objectivesRepo)

	notificationRepo := notificationsrepository.New(connections.Pool)
	_ = notifications.New(log, notificationRepo, rdb, tasksService)

	// Initialize Stripe client (required by subscriptions service)
	stripeClient := &client.API{}
	if cfg.Stripe.SecretKey != "" {
		stripeClient = client.New(cfg.Stripe.SecretKey, nil)
	}

	subRepo := subscriptionsrepository.New(connections.Pool)
	subService := subscriptions.New(log, subRepo, stripeClient, cfg.Stripe.WebhookSecret, tasksService)

	workspaceRepo := workspacesrepository.New(connections.Pool)
	workspaceUnitOfWork, err := workspaceuow.New(connections.Pool, workspaceRepo, teamsRepo, usersRepo)
	if err != nil {
		log.Error(ctx, "failed to initialize workspace unit of work", "error", err)
		os.Exit(1)
	}
	workspacesService := workspaces.New(
		log,
		workspaceRepo,
		workspaceUnitOfWork,
		workspaces.Dependencies{
			SeedContent:   workspacebootstrap.NewSeedContentCreator(statesService, storiesService),
			Users:         workspacebootstrap.NewUserDirectory(usersService),
			Subscriptions: workspacebootstrap.NewSubscriptionManager(subService),
			Publisher:     publisher,
			Trials:        workspacebootstrap.NewTrialScheduler(tasksService),
		},
	)

	// Create Seeder and Run
	seeder := seeding.NewSeeder(usersService, workspacesService)
	err = seeder.Run(ctx, seeding.SeedData{
		UserEmail:     *email,
		UserFullName:  *fullName,
		WorkspaceName: *name,
		WorkspaceSlug: *slug,
	})

	if err != nil {
		fmt.Printf("Error during seeding: %v\n", err)
		os.Exit(1)
	}
}
