package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/db/seeding"
	"github.com/complexus-tech/projects-api/internal/repo/mentionsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/notificationsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivestatusrepo"
	"github.com/complexus-tech/projects-api/internal/repo/okractivitiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/subscriptionsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/teamsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/repo/workspacesrepo"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
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
	System struct {
		UserID string `default:"00000000-0000-0000-0000-000000000001" env:"APP_SYSTEM_USER_ID"`
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
	disableTLS := flag.Bool("disable-tls", true, "Disable TLS for database connection")
	flag.Parse()

	ctx := context.Background()
	log := logger.NewWithJSON(os.Stdout, slog.LevelDebug, "seeder")

	// Parse config
	var cfg Config
	if err := config.Parse("app", &cfg); err != nil {
		log.Error(ctx, "failed to parse config", "error", err)
		os.Exit(1)
	}

	log.Info(ctx, "database config",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"name", cfg.DB.Name,
		"user", cfg.DB.User,
		"disable_tls", cfg.DB.DisableTLS,
	)

	// Connect to DB
	db, err := database.Open(database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   *disableTLS, // Use flag which defaults to true
	})
	if err != nil {
		log.Error(ctx, "failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Connect to Redis (required for tasks and publisher)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Cache.Host, cfg.Cache.Port),
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.Name,
	})
	defer rdb.Close()

	// Initialize Minimal Services
	cacheService := cache.New(rdb, log)
	tasksService, _ := tasks.New(rdb, log)
	publisher := publisher.New(rdb, log)

	systemUserID, _ := uuid.Parse(cfg.System.UserID)

	// Dependency Tree
	usersRepo := usersrepo.New(log, db)
	usersService := users.New(log, usersRepo, tasksService)

	teamsRepo := teamsrepo.New(log, db)
	teamsService := teams.New(log, teamsRepo)

	statesRepo := statesrepo.New(log, db)
	statesService := states.New(log, statesRepo)

	mentionsRepo := mentionsrepo.New(log, db)
	storiesRepo := storiesrepo.New(log, db)
	storiesService := stories.New(log, storiesRepo, mentionsRepo, publisher)

	okrActivitiesRepo := okractivitiesrepo.New(log, db)
	okrActivitiesService := okractivities.New(log, okrActivitiesRepo)
	objectivesRepo := objectivesrepo.New(log, db)
	_ = objectives.New(log, objectivesRepo, okrActivitiesService)

	notificationRepo := notificationsrepo.New(log, db)
	_ = notifications.New(log, notificationRepo, rdb, tasksService)

	objStatusRepo := objectivestatusrepo.New(log, db)
	objStatusService := objectivestatus.New(log, objStatusRepo)

	// Initialize Stripe client (required by subscriptions service)
	stripeClient := &client.API{}
	if cfg.Stripe.SecretKey != "" {
		stripeClient = client.New(cfg.Stripe.SecretKey, nil)
	}

	subRepo := subscriptionsrepo.New(log, db)
	subService := subscriptions.New(log, subRepo, stripeClient, cfg.Stripe.WebhookSecret, tasksService)

	workspaceRepo := workspacesrepo.New(log, db)
	workspacesService := workspaces.New(
		log,
		workspaceRepo,
		db,
		teamsService,
		storiesService,
		statesService,
		usersService,
		objStatusService,
		subService,
		nil, // attachments (not needed for seed)
		cacheService,
		systemUserID,
		publisher,
		tasksService,
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
