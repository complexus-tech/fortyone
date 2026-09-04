package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/bootstrap/githubadapter"
	"github.com/complexus-tech/projects-api/internal/bootstrap/integrationrequestsadapter"
	bootstrapproviders "github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	"github.com/complexus-tech/projects-api/internal/bootstrap/slackadapter"
	workspacebootstrap "github.com/complexus-tech/projects-api/internal/bootstrap/workspaces"
	activitiesrepository "github.com/complexus-tech/projects-api/internal/modules/activities/repository"
	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	adminrepository "github.com/complexus-tech/projects-api/internal/modules/admin/repository"
	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	chatsessionsrepository "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository"
	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/modules/developeraccess"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	documentsrepository "github.com/complexus-tech/projects-api/internal/modules/documents/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	feedbackrepository "github.com/complexus-tech/projects-api/internal/modules/feedback/repository"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	feedbackstory "github.com/complexus-tech/projects-api/internal/modules/feedback/storyadapter"
	figmarepository "github.com/complexus-tech/projects-api/internal/modules/figma/repository"
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	googledriverepository "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository"
	googledrive "github.com/complexus-tech/projects-api/internal/modules/googledrive/service"
	integrationrequestsrepository "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository"
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	invitationsrepository "github.com/complexus-tech/projects-api/internal/modules/invitations/repository"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	keyresultsrepository "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labelsrepository "github.com/complexus-tech/projects-api/internal/modules/labels/repository"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	linksrepository "github.com/complexus-tech/projects-api/internal/modules/links/repository"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatusrepository "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/repository"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	okractivitiesrepository "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	sprintsrepository "github.com/complexus-tech/projects-api/internal/modules/sprints/repository"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	teamsettingsrepository "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository"
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	workspaceuow "github.com/complexus-tech/projects-api/internal/modules/workspaces/uow"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	actorsrepository "github.com/complexus-tech/projects-api/internal/platform/actors/repository"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	platformidempotency "github.com/complexus-tech/projects-api/internal/platform/idempotency"
)

var _ messaging.StoryMutationService = (*stories.Service)(nil)

type services struct {
	activities           *activities.Service
	admin                *admin.Service
	attachments          *attachments.Service
	calendar             *calendar.Service
	chatSessions         *chatsessions.Service
	comments             *comments.Service
	documents            *documents.Service
	developerCredentials *developercredentials.Service
	developerOAuth       *developeroauth.Platform
	developerOAuthApps   *developeroauth.ApplicationManager
	developerAccess      *developeraccess.Resolver
	idempotency          *platformidempotency.Service
	outboundWebhooks     *outboundwebhooksservice.Manager
	emailReply           *emailreply.Service
	epics                *epics.Service
	feedback             *feedback.Service
	figma                *figma.Service
	googleDrive          *googledrive.Service
	github               *github.Service
	slack                *slack.Service
	integrationRequests  *integrationrequests.Service
	invitations          *invitations.Service
	keyResults           *keyresults.Service
	labels               *labels.Service
	links                *links.Service
	maya                 *maya.Service
	notifications        *notifications.Service
	objectives           *objectives.Service
	objectiveStats       *objectivestatus.Service
	okrActivities        *okractivities.Service
	reports              *reports.Service
	search               *search.Service
	sprints              *sprints.Service
	states               *states.Service
	stories              *stories.Service
	subscriptions        *subscriptions.Service
	teams                *teams.Service
	teamSettings         *teamsettings.Service
	users                *users.Service
	workspaces           *workspaces.Service
}

func buildServices(cfg mux.Config, dependencies Dependencies) services {
	developerCredentialService := buildDeveloperCredentialService(dependencies)
	developerAccessResolver, err := developeraccess.NewResolver(
		developerCredentialService,
		dependencies.DeveloperAPIOAuth,
	)
	if err != nil {
		panic("failed to initialize public API developer access: " + err.Error())
	}
	idempotencyService, err := platformidempotency.New(
		dependencies.DatabasePool,
		platformidempotency.DefaultConfig(),
	)
	if err != nil {
		panic("failed to initialize API idempotency service: " + err.Error())
	}
	outboundWebhookManager, err := buildOutboundWebhookManager(dependencies, developerCredentialService)
	if err != nil {
		panic("failed to initialize outbound webhook service: " + err.Error())
	}
	attachmentsService := attachments.New(
		cfg.Log,
		attachmentsrepository.New(dependencies.DatabasePool),
		cfg.StorageService,
		cfg.StorageConfig,
		cfg.TasksService,
	)

	usersRepository := usersrepository.New(dependencies.DatabasePool)
	usersService := users.New(
		cfg.Log,
		usersRepository,
		cfg.TasksService,
		users.WithVerificationTokens(
			dependencies.VerificationTokens,
			usersRepository,
		),
	)
	teamsRepository := teamsrepository.New(dependencies.DatabasePool)
	teamsService := teams.New(cfg.Log, teamsRepository)
	statesService := states.New(statesrepository.New(dependencies.DatabasePool))
	objectiveStatusService := objectivestatus.New(objectivestatusrepository.New(dependencies.DatabasePool))
	subscriptionsService := subscriptions.New(
		cfg.Log,
		subscriptionsrepository.New(dependencies.DatabasePool),
		cfg.StripeClient,
		cfg.WebhookSecret,
		cfg.TasksService,
		subscriptions.WithRedirectOrigin(cfg.WebsiteURL),
	)
	commentsService := comments.New(commentsrepository.New(cfg.Log, dependencies.DatabasePool))
	mayaRepository := mayarepository.New(dependencies.DatabasePool)
	storiesService := stories.New(
		cfg.Log,
		storiesrepository.New(
			cfg.Log,
			dependencies.DatabasePool,
			storiesrepository.WithAttachmentObjectStorage(
				cfg.StorageConfig.Provider,
				cfg.StorageConfig.AttachmentsBucket,
			),
		),
		cfg.Publisher,
		cfg.TasksService,
	)
	storyCommentCreator, err := bootstrapproviders.NewStoryCommentCreator(commentsService)
	if err != nil {
		panic("failed to initialize story comment adapter: " + err.Error())
	}
	storiesService.ConfigureCommentCreator(storyCommentCreator)
	storiesService.ConfigureAutoSchedulingEligibility(mayaRepository.WorkspaceCanUseMaya)
	integrationRequestsRepo := integrationrequestsrepository.New(dependencies.DatabasePool)
	messagingRepo := messagingrepository.New(dependencies.DatabasePool)
	emailReplyService, err := emailreply.New(cfg.EmailReplySecurityKey, messagingRepo, cfg.TasksService)
	if err != nil {
		panic("failed to initialize Brevo email reply ingress: " + err.Error())
	}
	linksService := links.New(cfg.Log, linksrepository.New(cfg.Log, dependencies.DatabasePool))
	workspaceRepository := workspacesrepository.New(dependencies.DatabasePool)
	workspaceUnitOfWork, err := workspaceuow.New(
		dependencies.DatabasePool,
		workspaceRepository,
		teamsRepository,
		usersRepository,
	)
	if err != nil {
		panic("failed to initialize workspace unit of work: " + err.Error())
	}
	workspacesService := workspaces.New(
		cfg.Log,
		workspaceRepository,
		workspaceUnitOfWork,
		workspaces.Dependencies{
			SeedContent:   workspacebootstrap.NewSeedContentCreator(statesService, storiesService),
			Users:         workspacebootstrap.NewUserDirectory(usersService),
			Subscriptions: workspacebootstrap.NewSubscriptionManager(subscriptionsService),
			Publisher:     cfg.Publisher,
			Trials:        workspacebootstrap.NewTrialScheduler(cfg.TasksService),
		},
	)

	invitationsService := invitations.New(
		invitationsrepository.New(dependencies.DatabasePool),
		dependencies.InvitationTokens,
		usersService,
		workspacesService,
	)

	okrActivitiesService := okractivities.New(okractivitiesrepository.New(dependencies.DatabasePool))
	keyResultsService := keyresults.New(
		cfg.Log,
		keyresultsrepository.New(dependencies.DatabasePool),
		keyresults.WithPublisher(cfg.Publisher),
	)
	objectivesService := objectives.New(cfg.Log, objectivesrepository.New(dependencies.DatabasePool), objectives.WithPublisher(cfg.Publisher))
	sprintsService := sprints.New(cfg.Log, sprintsrepository.New(dependencies.DatabasePool))
	searchService := search.New(cfg.Log, searchrepository.New(dependencies.DatabasePool))
	mutationConfirmer, err := messaging.NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		searchService,
		objectivesService,
		messaging.WithStoryMutations(cfg.MessagingMutationHMACKey),
		messaging.WithStoryMutationConfirmationStore(messagingRepo),
	)
	if err != nil {
		panic("failed to initialize Slack mutation confirmer: " + err.Error())
	}
	githubRepository := githubrepository.New(dependencies.DatabasePool)
	githubConfig := github.Config{
		AppID:                cfg.GitHubAppID,
		AppSlug:              cfg.GitHubAppSlug,
		ClientID:             cfg.GitHubClientID,
		ClientSecret:         cfg.GitHubClientSecret,
		PrivateKeyBase64:     cfg.GitHubKeyBase64,
		RedirectURL:          cfg.GitHubRedirect,
		WebhookSecret:        cfg.GitHubWebhook,
		WebsiteURL:           cfg.WebsiteURL,
		WebhookPayloadSecret: cfg.GitHubWebhookPayloadSecret,
		GitHubUserID:         cfg.GitHubUserID,
		CredentialVault:      dependencies.CredentialVault,
		OAuthStateStore:      cfg.Cache,
	}
	githubWebhookGateway, githubWebhookInbox, githubWebhookPayloads, err := buildGitHubWebhookGateway(
		dependencies.DatabasePool,
		githubRepository,
		cfg.TasksService,
		githubConfig,
	)
	if err != nil {
		panic("failed to initialize GitHub webhook gateway: " + err.Error())
	}
	githubConfig.WebhookGateway = githubWebhookGateway
	githubConfig.WebhookInbox = githubWebhookInbox
	githubConfig.WebhookPayloads = githubWebhookPayloads
	githubService, err := github.New(
		cfg.Log,
		githubRepository,
		githubadapter.NewStoryService(storiesService),
		githubadapter.NewRequestStore(integrationRequestsRepo),
		attachmentsService,
		githubConfig,
	)
	if err != nil {
		panic("failed to initialize github service: " + err.Error())
	}
	slackRepository := slackrepository.New(dependencies.DatabasePool)
	slackConfig := slack.Config{
		SigningSecret:        cfg.SlackSigningSecret,
		ClientID:             cfg.SlackClientID,
		ClientSecret:         cfg.SlackClientSecret,
		RedirectURL:          cfg.SlackRedirectURL,
		WebsiteURL:           cfg.WebsiteURL,
		WebhookPayloadSecret: cfg.SlackWebhookPayloadSecret,
		CredentialVault:      dependencies.CredentialVault,
	}
	slackWebhookGateway, err := buildSlackWebhookGateway(
		dependencies.DatabasePool,
		slackRepository,
		cfg.TasksService,
		slackConfig,
	)
	if err != nil {
		panic("failed to initialize Slack webhook gateway: " + err.Error())
	}
	slackRequestStore := slackadapter.NewRequestStore(integrationRequestsRepo)
	slackStoryService := slackadapter.NewStoryService(storiesService)
	slackMessagingStore := slackadapter.NewMessagingStore(messagingRepo)
	slackService := slack.New(
		cfg.Log,
		slackRepository,
		slackRequestStore,
		slackStoryService,
		slackConfig,
		slack.WithEventRuntime(slackWebhookGateway, slackMessagingStore),
		slack.WithNonceStore(slackMessagingStore),
		slack.WithMutationConfirmer(slackadapter.NewMutationConfirmer(mutationConfirmer)),
		slack.WithObjectiveReader(slackadapter.NewObjectiveReader(objectivesService)),
		slack.WithSprintReader(slackadapter.NewSprintReader(sprintsService)),
	)
	calendarService := calendar.New(
		cfg.Log,
		calendarrepository.New(dependencies.DatabasePool),
		calendar.Config{
			SecretKey:  cfg.SecretKey,
			WebsiteURL: cfg.WebsiteURL,
			WebhookURL: cfg.GoogleCalendarWebhookURL,
			WebhookURLs: map[calendar.Provider]string{
				calendar.ProviderGoogle:    cfg.GoogleCalendarWebhookURL,
				calendar.ProviderMicrosoft: cfg.MicrosoftCalendarWebhookURL,
			},
			RequireWebhook: true,
			Tasks:          cfg.TasksService,
			Updates:        cfg.Publisher,
			Providers: map[calendar.Provider]calendar.CalendarProvider{
				calendar.ProviderGoogle:    calendar.NewGoogleProvider(cfg.GoogleService),
				calendar.ProviderMicrosoft: calendar.NewMicrosoftProvider(cfg.MicrosoftService),
			},
		},
	)
	actorResolver := actors.NewResolver(actorsrepository.New(dependencies.DatabasePool))
	mayaActorID, err := actorResolver.Resolve(context.Background(), actors.KeySystem)
	if err != nil {
		panic("failed to resolve maya actor: " + err.Error())
	}
	reportsService := reports.New(cfg.Log, reportsrepository.New(cfg.Log, dependencies.DatabasePool))
	teamSettingsService := teamsettings.New(
		cfg.Log,
		teamsettingsrepository.New(dependencies.DatabasePool),
		newTeamSettingsAutomationScheduler(cfg.TasksService),
	)
	mayaPlanner := maya.NewPlanner()
	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		aiClient := maya.NewOpenAICompatibleClient(maya.OpenAICompatibleConfig{
			APIKey: strings.TrimSpace(cfg.AIAPIKey),
		})
		mayaPlanner = maya.NewPlannerWithAdvisor(maya.NewOpenAIAdvisor(aiClient))
	}
	mayaService := maya.New(maya.Dependencies{
		Repository:        mayaRepository,
		Realtime:          mayaRepository,
		Stories:           storiesService,
		Reports:           reportsService,
		Calendar:          calendarService,
		Users:             usersService,
		WorkspaceSettings: workspacesService,
		Planner:           mayaPlanner,
		MayaActorID:       mayaActorID,
	})
	storiesService.ConfigureMayaAssignment(mayaActorID, func(ctx context.Context, input stories.MayaAssignmentInput) error {
		if err := ensureBackgroundMayaEnabled(ctx, mayaRepository, input.Story.Workspace); err != nil {
			return err
		}
		return nil
	})
	slackProviderCallbacks := slackadapter.NewProviderCallbacks(slackService)
	integrationRequestsService := integrationrequests.New(
		cfg.Log,
		integrationRequestsRepo,
		integrationrequestsadapter.NewStoryCreator(storiesService),
		map[string]integrationrequests.ProviderAccepter{
			integrationrequests.ProviderGitHub: githubadapter.ProviderAccepter{Service: githubService},
			integrationrequests.ProviderSlack:  slackProviderCallbacks,
		},
		integrationrequests.WithProviderCommenter(integrationrequests.ProviderSlack, slackProviderCallbacks),
	)
	feedbackService := feedback.New(
		feedbackrepository.New(cfg.Log, dependencies.DatabasePool),
		feedbackstory.New(storiesService),
		feedback.WithEventPublisher(cfg.Log, cfg.Publisher),
		feedback.WithContributorFeatures(cfg.FeedbackSecurityKey, cfg.WebsiteURL, cfg.TasksService),
		feedback.WithGuestNotificationActor(mayaActorID),
	)
	figmaRepository := figmarepository.New(dependencies.DatabasePool)
	figmaConfig := figma.Config{
		ClientID:             cfg.FigmaClientID,
		ClientSecret:         cfg.FigmaClientSecret,
		RedirectURL:          cfg.FigmaRedirectURL,
		WebhookURL:           cfg.FigmaWebhookURL,
		WebsiteURL:           cfg.WebsiteURL,
		Credentials:          dependencies.CredentialVault,
		WebhookPayloadSecret: cfg.FigmaWebhookPayloadSecret,
	}
	figmaWebhookGateway, figmaWebhookInbox, figmaWebhookPayloads, err := buildFigmaWebhookGateway(
		dependencies.DatabasePool,
		figmaRepository,
		cfg.TasksService,
		figmaConfig,
	)
	if err != nil {
		panic("failed to initialize Figma webhook gateway: " + err.Error())
	}
	figmaConfig.WebhookGateway = figmaWebhookGateway
	figmaConfig.WebhookInbox = figmaWebhookInbox
	figmaConfig.WebhookPayloads = figmaWebhookPayloads
	figmaStories, err := bootstrapproviders.NewFigmaStoryAdapter(storiesService)
	if err != nil {
		panic("failed to initialize Figma story adapter: " + err.Error())
	}
	figmaService := figma.New(
		cfg.Log,
		figmaRepository,
		figmaStories,
		figmaConfig,
	)
	googleDriveService := googledrive.New(
		cfg.Log,
		googledriverepository.New(dependencies.DatabasePool),
		googledrive.Config{
			ClientID: cfg.GoogleDriveClientID, ClientSecret: cfg.GoogleDriveClientSecret,
			RedirectURL: cfg.GoogleDriveRedirectURL, PickerAPIKey: cfg.GoogleDrivePickerAPIKey,
			AppID: cfg.GoogleDriveAppID, WebsiteURL: cfg.WebsiteURL,
			Credentials: dependencies.CredentialVault,
		},
	)

	return services{
		activities: activities.New(activitiesrepository.New(dependencies.DatabasePool)),
		admin: admin.New(
			adminrepository.New(dependencies.DatabasePool),
			admin.WithAssetResolver(attachmentsService),
			admin.WithSubscriptionSyncer(subscriptionsService),
		),
		attachments:          attachmentsService,
		calendar:             calendarService,
		chatSessions:         chatsessions.New(cfg.Log, chatsessionsrepository.New(cfg.Log, dependencies.DatabasePool)),
		comments:             commentsService,
		documents:            documents.New(documentsrepository.New(dependencies.DatabasePool)),
		developerCredentials: developerCredentialService,
		developerOAuth:       dependencies.DeveloperOAuthPlatform,
		developerOAuthApps:   dependencies.DeveloperOAuthApplications,
		developerAccess:      developerAccessResolver,
		idempotency:          idempotencyService,
		outboundWebhooks:     outboundWebhookManager,
		emailReply:           emailReplyService,
		epics:                epics.New(),
		feedback:             feedbackService,
		figma:                figmaService,
		googleDrive:          googleDriveService,
		github:               githubService,
		slack:                slackService,
		integrationRequests:  integrationRequestsService,
		invitations:          invitationsService,
		keyResults:           keyResultsService,
		labels:               labels.New(labelsrepository.New(dependencies.DatabasePool)),
		links:                linksService,
		maya:                 mayaService,
		notifications:        notifications.New(cfg.Log, notificationsrepository.New(dependencies.DatabasePool), cfg.Redis, cfg.TasksService),
		objectives:           objectivesService,
		objectiveStats:       objectiveStatusService,
		okrActivities:        okrActivitiesService,
		reports:              reportsService,
		search:               searchService,
		sprints:              sprintsService,
		states:               statesService,
		stories:              storiesService,
		subscriptions:        subscriptionsService,
		teams:                teamsService,
		teamSettings:         teamSettingsService,
		users:                usersService,
		workspaces:           workspacesService,
	}
}

func (s services) validate() error {
	if s.activities == nil {
		return fmt.Errorf("missing service: activities")
	}
	if s.admin == nil {
		return fmt.Errorf("missing service: admin")
	}
	if s.attachments == nil {
		return fmt.Errorf("missing service: attachments")
	}
	if s.calendar == nil {
		return fmt.Errorf("missing service: calendar")
	}
	if s.chatSessions == nil {
		return fmt.Errorf("missing service: chatSessions")
	}
	if s.comments == nil {
		return fmt.Errorf("missing service: comments")
	}
	if s.documents == nil {
		return fmt.Errorf("missing service: documents")
	}
	if s.developerCredentials == nil {
		return fmt.Errorf("missing service: developerCredentials")
	}
	if s.developerOAuth == nil {
		return fmt.Errorf("missing service: developerOAuth")
	}
	if s.developerOAuthApps == nil {
		return fmt.Errorf("missing service: developerOAuthApps")
	}
	if s.developerAccess == nil {
		return fmt.Errorf("missing service: developerAccess")
	}
	if s.idempotency == nil {
		return fmt.Errorf("missing service: idempotency")
	}
	if s.outboundWebhooks == nil {
		return fmt.Errorf("missing service: outboundWebhooks")
	}
	if s.emailReply == nil {
		return fmt.Errorf("missing service: emailReply")
	}
	if s.epics == nil {
		return fmt.Errorf("missing service: epics")
	}
	if s.feedback == nil {
		return fmt.Errorf("missing service: feedback")
	}
	if s.figma == nil {
		return fmt.Errorf("missing service: figma")
	}
	if s.googleDrive == nil {
		return fmt.Errorf("missing service: googleDrive")
	}
	if s.github == nil {
		return fmt.Errorf("missing service: github")
	}
	if s.slack == nil {
		return fmt.Errorf("missing service: slack")
	}
	if s.integrationRequests == nil {
		return fmt.Errorf("missing service: integrationRequests")
	}
	if s.invitations == nil {
		return fmt.Errorf("missing service: invitations")
	}
	if s.keyResults == nil {
		return fmt.Errorf("missing service: keyResults")
	}
	if s.labels == nil {
		return fmt.Errorf("missing service: labels")
	}
	if s.links == nil {
		return fmt.Errorf("missing service: links")
	}
	if s.maya == nil {
		return fmt.Errorf("missing service: maya")
	}
	if s.notifications == nil {
		return fmt.Errorf("missing service: notifications")
	}
	if s.objectives == nil {
		return fmt.Errorf("missing service: objectives")
	}
	if s.objectiveStats == nil {
		return fmt.Errorf("missing service: objectiveStats")
	}
	if s.okrActivities == nil {
		return fmt.Errorf("missing service: okrActivities")
	}
	if s.reports == nil {
		return fmt.Errorf("missing service: reports")
	}
	if s.search == nil {
		return fmt.Errorf("missing service: search")
	}
	if s.sprints == nil {
		return fmt.Errorf("missing service: sprints")
	}
	if s.states == nil {
		return fmt.Errorf("missing service: states")
	}
	if s.stories == nil {
		return fmt.Errorf("missing service: stories")
	}
	if s.subscriptions == nil {
		return fmt.Errorf("missing service: subscriptions")
	}
	if s.teams == nil {
		return fmt.Errorf("missing service: teams")
	}
	if s.teamSettings == nil {
		return fmt.Errorf("missing service: teamSettings")
	}
	if s.users == nil {
		return fmt.Errorf("missing service: users")
	}
	if s.workspaces == nil {
		return fmt.Errorf("missing service: workspaces")
	}

	return nil
}
