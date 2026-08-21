package workerbootstrap

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	feedbackrepository "github.com/complexus-tech/projects-api/internal/modules/feedback/repository"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type App struct {
	log       *logger.Logger
	db        *sqlx.DB
	server    *asynq.Server
	scheduler *asynq.Scheduler
	taskMux   *asynq.ServeMux
	redisOpt  asynq.RedisClientOpt
	redis     *redis.Client
	tasks     *tasks.Service
}

func New(ctx context.Context, log *logger.Logger) (App, error) {
	log.Info(ctx, "Starting worker run function")

	cfg, err := loadConfig()
	if err != nil {
		return App{}, fmt.Errorf("error parsing worker configuration: %w", err)
	}
	if err := emailreply.ValidateRuntimeSecret(cfg.Auth.SecretKey, cfg.Email.Environment); err != nil {
		return App{}, fmt.Errorf("validate email reply security: %w", err)
	}

	db, err := openDB(cfg)
	if err != nil {
		return App{}, err
	}
	log.Info(ctx, "database connection established")

	redisOpt := redisClientOpt(cfg)
	redisClient := openRedis(cfg)

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return App{}, fmt.Errorf("error pinging redis: %w", err)
	}

	tasksService, err := tasks.New(redisClient, log)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return App{}, fmt.Errorf("initialize worker tasks service: %w", err)
	}
	resourcesTransferred := false
	defer func() {
		if resourcesTransferred {
			return
		}
		_ = tasksService.Close()
		_ = redisClient.Close()
	}()
	notificationsService := notifications.New(
		log,
		notificationsrepository.New(log, db),
		redisClient,
		tasksService,
	)
	messagingRepo := messagingrepository.New(db)
	emailThreads, err := emailthread.New(messagingRepo)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize Maya email threading: %w", err)
	}
	emailReplyIngress, err := emailreply.New(cfg.Auth.SecretKey, messagingRepo, tasksService)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize Brevo email reply recovery: %w", err)
	}

	cacheService := cache.New(redisClient, log)
	actorResolver := actors.NewResolver(log, db, cacheService)

	systemUserID, err := actorResolver.Resolve(ctx, actors.KeySystem)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("resolve system actor: %w", err)
	}

	githubUserID, err := actorResolver.Resolve(ctx, actors.KeyGitHub)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("resolve github actor: %w", err)
	}

	scheduler := asynq.NewScheduler(redisOpt, nil)
	if err := registerSchedules(scheduler); err != nil {
		_ = db.Close()
		return App{}, err
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency:    10,
			Queues:         cfg.Queues,
			RetryDelayFunc: integrationRetryDelay,
		},
	)

	brevoService, err := brevo.NewService(brevo.Config{
		APIKey: cfg.Brevo.APIKey,
	}, log)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("error initializing brevo service: %w", err)
	}

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
		_ = db.Close()
		return App{}, fmt.Errorf("error initializing mailer service: %w", err)
	}

	githubService, err := github.New(log, githubrepository.New(log, db), nil, nil, nil, github.Config{
		AppID:            cfg.GitHub.AppID,
		AppSlug:          cfg.GitHub.AppSlug,
		PrivateKeyBase64: cfg.GitHub.PrivateKeyBase64,
		RedirectURL:      cfg.GitHub.RedirectURL,
		WebhookSecret:    cfg.GitHub.WebhookSecret,
		WebsiteURL:       cfg.Website.URL,
		SecretKey:        cfg.Auth.SecretKey,
		GitHubUserID:     githubUserID,
	})
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("error initializing github service: %w", err)
	}
	if err := githubService.ValidateWorkerConfiguration(); err != nil {
		_ = db.Close()
		return App{}, err
	}

	storageConfig := storage.Config{
		Provider:          cfg.Storage.Provider,
		ProfilesBucket:    cfg.Storage.ProfilesBucket,
		LogosBucket:       cfg.Storage.LogosBucket,
		AttachmentsBucket: cfg.Storage.AttachmentsBucket,
		Azure: azure.Config{
			ConnectionString:   cfg.Azure.StorageConnectionString,
			StorageAccountName: cfg.Azure.StorageAccountName,
			AccountKey:         cfg.Azure.StorageAccountKey,
		},
		AWS: aws.Config{
			AccessKeyID:     cfg.AWS.AccessKeyID,
			SecretAccessKey: cfg.AWS.SecretAccessKey,
			Region:          cfg.AWS.Region,
			Endpoint:        cfg.AWS.Endpoint,
			PublicURL:       cfg.AWS.PublicURL,
			ForcePathStyle:  cfg.AWS.ForcePathStyle,
			Bucket:          cfg.AWS.Bucket,
		},
	}
	storageService, err := storage.NewStorageService(storageConfig, log)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize worker storage service: %w", err)
	}
	attachmentsService := attachments.New(
		log,
		attachmentsrepository.New(log, db),
		storageService,
		storageConfig,
		nil,
	)

	googleService, err := google.NewService(google.Config{
		ClientIDs:           parseWorkerClientIDs(cfg.Auth.GoogleClientIDs, cfg.GoogleClientID),
		ClientSecret:        cfg.Auth.GoogleClientSecret,
		CalendarRedirectURL: cfg.Auth.GoogleCalendarRedirectURL,
	})
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize worker Google service: %w", err)
	}
	microsoftService := microsoft.NewService(microsoft.Config{
		ClientID:            cfg.Auth.MicrosoftClientID,
		ClientSecret:        cfg.Auth.MicrosoftClientSecret,
		Tenant:              cfg.Auth.MicrosoftTenant,
		CalendarRedirectURL: cfg.Auth.MicrosoftCalendarRedirectURL,
	})
	eventPublisher := publisher.New(redisClient, log)
	storyScheduleOutbox := stories.NewScheduleTransitionOutboxDispatcher(
		log,
		storiesrepository.New(log, db),
		eventPublisher,
	)
	feedbackOutbox := feedback.New(
		feedbackrepository.New(log, db),
		nil,
		feedback.WithEventPublisher(log, eventPublisher),
		feedback.WithContributorFeatures(cfg.Auth.SecretKey, cfg.Website.URL, tasksService),
		feedback.WithGuestNotificationActor(systemUserID),
	)
	calendarService := calendar.New(log, calendarrepository.New(log, db), calendar.Config{
		SecretKey:  cfg.Auth.SecretKey,
		WebsiteURL: cfg.Website.URL,
		WebhookURL: cfg.Auth.GoogleCalendarWebhookURL,
		WebhookURLs: map[calendar.Provider]string{
			calendar.ProviderGoogle:    cfg.Auth.GoogleCalendarWebhookURL,
			calendar.ProviderMicrosoft: cfg.Auth.MicrosoftCalendarWebhookURL,
		},
		Providers: map[calendar.Provider]calendar.CalendarProvider{
			calendar.ProviderGoogle:    calendar.NewGoogleProvider(googleService),
			calendar.ProviderMicrosoft: calendar.NewMicrosoftProvider(microsoftService),
		},
		Tasks:   tasksService,
		Updates: eventPublisher,
	})
	mayaService := buildMayaService(log, db, cfg, calendarService, systemUserID, eventPublisher)
	emailCopyClient, err := emailcopy.New(emailcopy.Config{
		APIKey: cfg.AIAPIKey,
	})
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize email copy generator: %w", err)
	}
	var emailCopyGenerator emailcopy.Generator
	if emailCopyClient.Enabled() {
		emailCopyGenerator = emailCopyClient
	}
	emailReplyProcessor, err := buildEmailReplyProcessor(
		log,
		db,
		redisClient,
		cfg,
		tasksService,
		mailerService,
		messagingRepo,
		emailReplyIngress,
		emailThreads,
	)
	if err != nil {
		_ = db.Close()
		return App{}, err
	}
	slackEvents, err := buildSlackEventProcessor(log, db, redisClient, cfg, slackEventProcessorDependencies{
		EventPublisher: eventPublisher,
		Tasks:          tasksService,
	})
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("initialize Slack event processor: %w", err)
	}
	upgradedCredentials, err := slackEvents.BackfillLegacyCredentials(ctx)
	if err != nil {
		_ = db.Close()
		return App{}, fmt.Errorf("encrypt legacy Slack credentials: %w", err)
	}
	if upgradedCredentials > 0 {
		log.Info(ctx, "Encrypted legacy Slack credentials", "count", upgradedCredentials)
	}
	taskMux := buildTaskMux(log, db, brevoService, mailerService, githubService, mayaService, attachmentsService, emailCopyGenerator, emailThreads, notificationsService, slackEvents, emailReplyProcessor, emailReplyIngress, calendarService, systemUserID, tasksService, feedbackOutbox, storyScheduleOutbox, cfg.Auth.SecretKey)
	resourcesTransferred = true

	return App{
		log:       log,
		db:        db,
		server:    server,
		scheduler: scheduler,
		taskMux:   taskMux,
		redisOpt:  redisOpt,
		redis:     redisClient,
		tasks:     tasksService,
	}, nil
}

func parseWorkerClientIDs(value, fallback string) []string {
	parts := strings.Split(value, ",")
	clientIDs := make([]string, 0, len(parts))
	for _, part := range parts {
		if clientID := strings.TrimSpace(part); clientID != "" {
			clientIDs = append(clientIDs, clientID)
		}
	}
	if len(clientIDs) == 0 {
		if clientID := strings.TrimSpace(fallback); clientID != "" {
			clientIDs = append(clientIDs, clientID)
		}
	}
	return clientIDs
}

func (a App) Run(ctx context.Context) error {
	defer func() {
		a.log.Info(ctx, "closing database connection")
		a.db.Close()
	}()
	defer func() {
		if a.tasks != nil {
			_ = a.tasks.Close()
		}
		if a.redis != nil {
			_ = a.redis.Close()
		}
	}()

	monitor := asynqmon.New(asynqmon.Options{
		RootPath:     "/",
		RedisConnOpt: a.redisOpt,
	})
	http.Handle(monitor.RootPath()+"/", monitor)

	go func() {
		a.log.Info(ctx, "Starting Asynqmon monitoring server...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			a.log.Error(ctx, "Failed to start HTTP server", "error", err)
		}
	}()

	go func() {
		a.log.Info(ctx, "Starting cleanup scheduler...")
		if err := a.scheduler.Run(); err != nil {
			a.log.Error(ctx, "Failed to start scheduler", "error", err)
		}
	}()

	defer func() {
		a.log.Info(ctx, "shutting down scheduler")
		a.scheduler.Shutdown()
	}()

	a.log.Info(ctx, "Starting Asynq worker server...")
	if err := a.server.Run(a.taskMux); err != nil {
		a.log.Error(ctx, "Asynq server Run() failed", "error", err)
		return fmt.Errorf("asynq server run error: %w", err)
	}

	return nil
}
