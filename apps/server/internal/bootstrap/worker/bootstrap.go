package workerbootstrap

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	feedbackrepository "github.com/complexus-tech/projects-api/internal/modules/feedback/repository"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	invitationsrepository "github.com/complexus-tech/projects-api/internal/modules/invitations/repository"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	actorsrepository "github.com/complexus-tech/projects-api/internal/platform/actors/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	platformidempotency "github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/aws"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/brevo"
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
	"github.com/redis/go-redis/v9"
)

type App struct {
	log           *logger.Logger
	database      *platformdatabase.Connections
	server        taskServer
	scheduler     taskScheduler
	taskMux       *asynq.ServeMux
	redisOpt      asynq.RedisClientOpt
	redis         *redis.Client
	pingRedis     redisPingFunc
	httpConfig    HTTPConfig
	monitorConfig MonitorConfig
	ready         *atomic.Bool
}

func New(ctx context.Context, log *logger.Logger) (App, error) {
	log.Info(ctx, "Starting worker run function")

	cfg, err := loadConfig()
	if err != nil {
		return App{}, fmt.Errorf("error parsing worker configuration: %w", err)
	}
	if _, err := validateRuntimeConfig(cfg); err != nil {
		return App{}, err
	}
	credentialVault, err := credentialvault.NewFromEncodedKeyring(
		cfg.CredentialVault.ActiveKeyID,
		cfg.CredentialVault.ActiveKeyVersion.Uint32(),
		cfg.CredentialVault.Keys,
	)
	if err != nil {
		return App{}, fmt.Errorf("initialize provider credential vault: %w", err)
	}
	invitationConfig, err := workerInvitationTokenConfig(cfg)
	if err != nil {
		return App{}, fmt.Errorf("initialize invitation token security: %w", err)
	}
	invitationTokens, err := invitations.NewInvitationTokenManager(invitationConfig)
	if err != nil {
		return App{}, fmt.Errorf("initialize invitation token security: %w", err)
	}

	connections, err := openDB(ctx, cfg)
	if err != nil {
		return App{}, err
	}
	databaseTransferred := false
	defer func() {
		if !databaseTransferred {
			_ = connections.Close()
		}
	}()
	log.Info(ctx, "database connection established")

	redisOpt := redisClientOpt(cfg)
	redisClient := openRedis(cfg)

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		_ = redisClient.Close()
		return App{}, fmt.Errorf("error pinging redis: %w", err)
	}

	tasksService, err := tasks.New(redisClient, log)
	if err != nil {
		_ = redisClient.Close()
		return App{}, fmt.Errorf("initialize worker tasks service: %w", err)
	}
	resourcesTransferred := false
	defer func() {
		if resourcesTransferred {
			return
		}
		_ = redisClient.Close()
	}()
	notificationsStore := notificationsrepository.New(connections.Pool)
	notificationsService := notifications.New(
		log,
		notificationsStore,
		redisClient,
		tasksService,
	)
	messagingRepo := messagingrepository.New(connections.Pool)
	emailThreads, err := emailthread.New(messagingRepo)
	if err != nil {
		return App{}, fmt.Errorf("initialize Maya email threading: %w", err)
	}
	emailReplyIngress, err := emailreply.New(cfg.EmailReply.SecurityKey, messagingRepo, tasksService)
	if err != nil {
		return App{}, fmt.Errorf("initialize Brevo email reply recovery: %w", err)
	}

	actorResolver := actors.NewResolver(actorsrepository.New(connections.Pool))

	systemUserID, err := actorResolver.Resolve(ctx, actors.KeySystem)
	if err != nil {
		return App{}, fmt.Errorf("resolve system actor: %w", err)
	}

	githubUserID, err := actorResolver.Resolve(ctx, actors.KeyGitHub)
	if err != nil {
		return App{}, fmt.Errorf("resolve github actor: %w", err)
	}

	scheduler := asynq.NewScheduler(redisOpt, nil)
	if err := registerSchedules(scheduler); err != nil {
		return App{}, err
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency:     10,
			Queues:          cfg.Queues,
			RetryDelayFunc:  integrationRetryDelay,
			ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
		},
	)

	brevoService, err := brevo.NewService(brevo.Config{
		APIKey: cfg.Brevo.APIKey,
	}, log)
	if err != nil {
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
		return App{}, fmt.Errorf("error initializing mailer service: %w", err)
	}
	invitationEmail, err := newInvitationEmailSender(mailerService, cfg.Website.URL)
	if err != nil {
		return App{}, fmt.Errorf("initialize invitation email delivery: %w", err)
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
		return App{}, fmt.Errorf("initialize worker storage service: %w", err)
	}
	attachmentsService := attachments.New(
		log,
		attachmentsrepository.New(connections.Pool),
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
		return App{}, fmt.Errorf("initialize worker Google service: %w", err)
	}
	microsoftService := microsoft.NewService(microsoft.Config{
		ClientID:            cfg.Auth.MicrosoftClientID,
		ClientSecret:        cfg.Auth.MicrosoftClientSecret,
		Tenant:              cfg.Auth.MicrosoftTenant,
		CalendarRedirectURL: cfg.Auth.MicrosoftCalendarRedirectURL,
	})
	eventPublisher := publisher.New(redisClient, log)
	mayaRepository := mayarepository.New(connections.Pool)
	githubService, err := buildGitHubWorkerService(
		log,
		connections.Pool,
		cfg,
		githubUserID,
		systemUserID,
		eventPublisher,
		tasksService,
		attachmentsService,
		credentialVault,
		buildGitHubCompatibilityDependencies(connections.Pool, mayaRepository),
	)
	if err != nil {
		return App{}, fmt.Errorf("initialize GitHub worker service: %w", err)
	}
	figmaRuntime, err := buildFigmaWorker(
		log,
		connections.Pool,
		cfg,
		eventPublisher,
		tasksService,
		credentialVault,
	)
	if err != nil {
		return App{}, fmt.Errorf("initialize Figma worker service: %w", err)
	}
	invitationOutbox := invitations.NewOutboxDispatcher(
		log,
		invitationsrepository.New(connections.Pool),
		invitationTokens,
		invitationEmail,
	)
	storyScheduleOutbox := stories.NewScheduleTransitionOutboxDispatcher(
		log,
		storiesrepository.New(log, connections.Pool),
		eventPublisher,
	)
	feedbackOutbox := feedback.New(
		feedbackrepository.New(log, connections.Pool),
		nil,
		feedback.WithEventPublisher(log, eventPublisher),
		feedback.WithContributorFeatures(cfg.Feedback.SecurityKey, cfg.Website.URL, tasksService),
		feedback.WithGuestNotificationActor(systemUserID),
	)
	calendarService := calendar.New(log, calendarrepository.New(connections.Pool), calendar.Config{
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
	mayaService := buildMayaService(
		log,
		connections.Pool,
		mayaRepository,
		cfg,
		calendarService,
		systemUserID,
		eventPublisher,
	)
	emailCopyClient, err := emailcopy.New(emailcopy.Config{
		APIKey: cfg.AIAPIKey,
	})
	if err != nil {
		return App{}, fmt.Errorf("initialize email copy generator: %w", err)
	}
	var emailCopyGenerator emailcopy.Generator
	if emailCopyClient.Enabled() {
		emailCopyGenerator = emailCopyClient
	}
	emailReplyProcessor, err := buildEmailReplyProcessor(
		log,
		connections.Pool,
		mayaRepository,
		redisClient,
		cfg,
		tasksService,
		mailerService,
		messagingRepo,
		emailReplyIngress,
		emailThreads,
		systemUserID,
	)
	if err != nil {
		return App{}, err
	}
	slackEvents, err := buildSlackEventProcessor(log, connections.Pool, redisClient, cfg, slackEventProcessorDependencies{
		EventPublisher:  eventPublisher,
		Tasks:           tasksService,
		MayaActorID:     systemUserID,
		MayaAccess:      mayaRepository,
		CredentialVault: credentialVault,
	})
	if err != nil {
		return App{}, fmt.Errorf("initialize Slack event processor: %w", err)
	}
	upgradedGitHubCredentials, err := githubService.BackfillLegacyUserCredentials(ctx)
	if err != nil {
		return App{}, fmt.Errorf("encrypt legacy GitHub user credentials: %w", err)
	}
	if upgradedGitHubCredentials > 0 {
		log.Info(ctx, "Encrypted legacy GitHub user credentials", "count", upgradedGitHubCredentials)
	}
	slackCutover, err := slack.NewLegacyCutover(cfg.Auth.SecretKey)
	if err != nil {
		return App{}, fmt.Errorf("initialize bounded Slack legacy cutover: %w", err)
	}
	slackBackfill, err := backfillLegacySlackData(ctx, slackEvents, slackCutover)
	slackCutover = nil
	if err != nil {
		return App{}, fmt.Errorf("complete legacy Slack security cutover: %w", err)
	}
	if slackBackfill.Credentials > 0 {
		log.Info(ctx, "Encrypted legacy Slack credentials", "count", slackBackfill.Credentials)
	}
	if slackBackfill.WebhookPayloads > 0 {
		log.Info(ctx, "Encrypted legacy Slack webhook payloads", "count", slackBackfill.WebhookPayloads)
	}
	figmaMigration, err := migrateLegacyFigmaCredentials(
		ctx,
		figmaRuntime,
		credentialVault,
		cfg.Auth.SecretKey,
	)
	if err != nil {
		return App{}, err
	}
	if figmaMigration.Migrated > 0 || figmaMigration.Stale > 0 {
		log.Info(
			ctx,
			"Migrated legacy Figma credentials",
			"scanned", figmaMigration.Scanned,
			"migrated", figmaMigration.Migrated,
			"stale", figmaMigration.Stale,
		)
	}
	if cfg.CredentialVault.RewrapOnStartup {
		githubRotation, err := githubService.RewrapUserCredentials(ctx)
		if err != nil {
			return App{}, fmt.Errorf("rewrap GitHub user credentials: %w", err)
		}
		log.Info(
			ctx,
			"Verified and rewrapped GitHub user credentials",
			"active_key_id", githubRotation.ActiveKey.ID,
			"active_key_version", githubRotation.ActiveKey.Version,
			"scanned", githubRotation.Scanned,
			"current", githubRotation.Current,
			"rewrapped", githubRotation.Rewrapped,
			"stale", githubRotation.Stale,
		)
		slackRotation, err := slackEvents.RewrapCredentials(ctx)
		if err != nil {
			return App{}, fmt.Errorf("rewrap Slack credentials: %w", err)
		}
		log.Info(
			ctx,
			"Verified and rewrapped Slack credentials",
			"active_key_id", slackRotation.ActiveKey.ID,
			"active_key_version", slackRotation.ActiveKey.Version,
			"scanned", slackRotation.Scanned,
			"current", slackRotation.Current,
			"rewrapped", slackRotation.Rewrapped,
			"stale", slackRotation.Stale,
		)
		figmaRotation, err := figmaRuntime.service.RewrapCredentials(ctx)
		if err != nil {
			return App{}, fmt.Errorf("rewrap Figma credentials: %w", err)
		}
		log.Info(
			ctx,
			"Verified and rewrapped Figma credentials",
			"active_key_id", figmaRotation.ActiveKey.ID,
			"active_key_version", figmaRotation.ActiveKey.Version,
			"scanned", figmaRotation.Scanned,
			"current", figmaRotation.Current,
			"rewrapped", figmaRotation.Rewrapped,
			"stale", figmaRotation.Stale,
		)
	}
	outboundWebhookDispatcher, err := buildOutboundWebhookDispatcher(connections.Pool, credentialVault)
	if err != nil {
		return App{}, fmt.Errorf("initialize outbound webhook worker: %w", err)
	}
	storyMutationEventDispatcher, err := buildStoryMutationEventDispatcher(log, connections.Pool)
	if err != nil {
		return App{}, fmt.Errorf("initialize story mutation event worker: %w", err)
	}
	idempotencyReceipts, err := platformidempotency.New(
		connections.Pool,
		platformidempotency.DefaultConfig(),
	)
	if err != nil {
		return App{}, fmt.Errorf("initialize API idempotency receipt cleanup: %w", err)
	}
	taskMux := buildTaskMux(taskMuxDependencies{
		Log: log, DatabasePool: connections.Pool,
		Brevo: brevoService, Mailer: mailerService,
		GitHub: githubService, Figma: figmaRuntime.service, Maya: mayaService,
		MayaRepository: mayaRepository,
		Attachments:    attachmentsService, EmailCopy: emailCopyGenerator,
		EmailThreads: emailThreads, Notifications: notificationsService,
		WeeklyDigest: notificationsStore,
		SlackEvents:  slackEvents, EmailReplies: emailReplyProcessor,
		EmailRecovery: emailReplyIngress, Calendar: calendarService,
		SystemUserID: systemUserID, FeedbackTasks: tasksService,
		FeedbackOutbox:      feedbackOutbox,
		StoryScheduleOutbox: storyScheduleOutbox,
		InvitationOutbox:    invitationOutbox,
		FeedbackSecurityKey: cfg.Feedback.SecurityKey,
		IdempotencyReceipts: idempotencyReceipts,
	})
	if err := registerOutboundWebhookTask(taskMux, log, storyMutationEventDispatcher, outboundWebhookDispatcher); err != nil {
		return App{}, fmt.Errorf("register outbound webhook worker: %w", err)
	}
	resourcesTransferred = true
	databaseTransferred = true

	return App{
		log:       log,
		database:  connections,
		server:    server,
		scheduler: scheduler,
		taskMux:   taskMux,
		redisOpt:  redisOpt,
		redis:     redisClient,
		pingRedis: func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		},
		httpConfig:    cfg.HTTP,
		monitorConfig: cfg.Monitor,
		ready:         &atomic.Bool{},
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
