package api

import (
	"context"
	"errors"
	"net/http"

	activitieshttp "github.com/complexus-tech/projects-api/internal/modules/activities/http"
	adminhttp "github.com/complexus-tech/projects-api/internal/modules/admin/http"
	agentreadinesshttp "github.com/complexus-tech/projects-api/internal/modules/agentreadiness/http"
	apiv1http "github.com/complexus-tech/projects-api/internal/modules/apiv1/http"
	calendarhttp "github.com/complexus-tech/projects-api/internal/modules/calendar/http"
	chatsessionshttp "github.com/complexus-tech/projects-api/internal/modules/chatsessions/http"
	commentshttp "github.com/complexus-tech/projects-api/internal/modules/comments/http"
	developercredentialshttp "github.com/complexus-tech/projects-api/internal/modules/developercredentials/http"
	developeroauthhttp "github.com/complexus-tech/projects-api/internal/modules/developeroauth/http"
	documentshttp "github.com/complexus-tech/projects-api/internal/modules/documents/http"
	emailreplyhttp "github.com/complexus-tech/projects-api/internal/modules/emailreply/http"
	epicshttp "github.com/complexus-tech/projects-api/internal/modules/epics/http"
	feedbackhttp "github.com/complexus-tech/projects-api/internal/modules/feedback/http"
	figmahttp "github.com/complexus-tech/projects-api/internal/modules/figma/http"
	githubhttp "github.com/complexus-tech/projects-api/internal/modules/github/http"
	googledrivehttp "github.com/complexus-tech/projects-api/internal/modules/googledrive/http"
	healthhttp "github.com/complexus-tech/projects-api/internal/modules/health/http"
	integrationrequestshttp "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/http"
	invitationshttp "github.com/complexus-tech/projects-api/internal/modules/invitations/http"
	keyresultshttp "github.com/complexus-tech/projects-api/internal/modules/keyresults/http"
	labelshttp "github.com/complexus-tech/projects-api/internal/modules/labels/http"
	linkshttp "github.com/complexus-tech/projects-api/internal/modules/links/http"
	mayahttp "github.com/complexus-tech/projects-api/internal/modules/maya/http"
	notificationshttp "github.com/complexus-tech/projects-api/internal/modules/notifications/http"
	objectiveshttp "github.com/complexus-tech/projects-api/internal/modules/objectives/http"
	objectivestatushttp "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/http"
	outboundwebhookshttp "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/http"
	reportshttp "github.com/complexus-tech/projects-api/internal/modules/reports/http"
	searchhttp "github.com/complexus-tech/projects-api/internal/modules/search/http"
	slackhttp "github.com/complexus-tech/projects-api/internal/modules/slack/http"
	sprintshttp "github.com/complexus-tech/projects-api/internal/modules/sprints/http"
	stateshttp "github.com/complexus-tech/projects-api/internal/modules/states/http"
	storieshttp "github.com/complexus-tech/projects-api/internal/modules/stories/http"
	subscriptionshttp "github.com/complexus-tech/projects-api/internal/modules/subscriptions/http"
	teamshttp "github.com/complexus-tech/projects-api/internal/modules/teams/http"
	teamsettingshttp "github.com/complexus-tech/projects-api/internal/modules/teamsettings/http"
	usershttp "github.com/complexus-tech/projects-api/internal/modules/users/http"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaceshttp "github.com/complexus-tech/projects-api/internal/modules/workspaces/http"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	ssehttp "github.com/complexus-tech/projects-api/internal/sse/http"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// userLookupAdapter wraps *users.Service to satisfy githubhttp.UserLookup.
type userLookupAdapter struct {
	svc *users.Service
}

func (a *userLookupAdapter) GetUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	u, err := a.svc.GetUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if u.FullName != "" {
		return u.FullName, nil
	}
	return u.Username, nil
}

type routes struct {
	services          services
	workspaceResolver mid.WorkspaceResolver
}

func NewWithServices(svcs services) routes {
	return routes{
		services:          svcs,
		workspaceResolver: workspaceResolver{service: svcs.workspaces},
	}
}

type workspaceResolver struct {
	service *workspaces.Service
}

func (resolver workspaceResolver) ResolveCurrentWorkspace(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (mid.WorkspaceInfo, error) {
	membership, err := resolver.service.ResolveCurrentMembership(ctx, slug, userID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return mid.WorkspaceInfo{}, mid.ErrWorkspaceAccessDenied
		}
		return mid.WorkspaceInfo{}, err
	}
	return mid.WorkspaceInfo{
		ID:       membership.WorkspaceID,
		Name:     membership.Name,
		Slug:     membership.Slug,
		UserRole: membership.Role,
	}, nil
}

func (resolver workspaceResolver) RecordWorkspaceAccess(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	return resolver.service.RecordAccess(ctx, workspaceID, userID)
}

func (r routes) BuildAllRoutes(app *web.App, cfg mux.Config) {
	svcs := r.services
	if err := svcs.validate(); err != nil {
		panic("bootstrap service validation failed: " + err.Error())
	}
	browserSessions := mid.NewBrowserSessionResolver(cfg.Cache, svcs.users)

	agentreadinesshttp.Routes(agentreadinesshttp.Config{
		APIPublicURL:      cfg.APIPublicURL,
		Workspaces:        svcs.workspaces,
		Teams:             svcs.teams,
		States:            svcs.states,
		Stories:           svcs.stories,
		Sprints:           svcs.sprints,
		Objectives:        svcs.objectives,
		ObjectiveStatuses: svcs.objectiveStats,
		KeyResults:        svcs.keyResults,
		Reports:           svcs.reports,
		OAuth:             svcs.developerOAuth,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		LoginURL:          cfg.MCPLoginURL,
		Log:               cfg.Log,
	}, app)

	healthhttp.Routes(healthhttp.Config{
		Log:       cfg.Log,
		Readiness: cfg.Readiness,
	}, app)
	developercredentialshttp.Routes(developercredentialshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.developerCredentials,
	}, app)
	developeroauthhttp.Routes(developeroauthhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.developerOAuthApps,
	}, app)
	outboundwebhookshttp.Routes(outboundwebhookshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.outboundWebhooks,
	}, app)
	apiv1http.Routes(apiv1http.Config{
		Log:                  cfg.Log,
		SecretKey:            cfg.SecretKey,
		Cache:                cfg.Cache,
		DeveloperCredentials: svcs.developerAccess,
		Workspaces:           svcs.workspaces,
		Teams:                svcs.teams,
		Stories:              svcs.stories,
		StoryComments:        svcs.stories,
		Labels:               svcs.labels,
		States:               svcs.states,
		Sprints:              svcs.sprints,
		Objectives:           svcs.objectives,
		KeyResults:           svcs.keyResults,
		Idempotency:          svcs.idempotency,
		Webhooks:             svcs.outboundWebhooks,
	}, app)
	emailreplyhttp.Routes(emailreplyhttp.Config{
		Service: svcs.emailReply,
	}, app)

	adminhttp.Routes(adminhttp.Config{
		Log:             cfg.Log,
		SecretKey:       cfg.SecretKey,
		Cache:           cfg.Cache,
		BrowserSessions: browserSessions,
		Service:         svcs.admin,
	}, app)

	githubhttp.Routes(githubhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.github,
		Users:             &userLookupAdapter{svc: svcs.users},
	}, app)
	figmahttp.Routes(figmahttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.figma,
	}, app)
	googledrivehttp.Routes(googledrivehttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		BrowserSessions:   browserSessions,
		Service:           svcs.googleDrive,
	}, app)
	slackhttp.Routes(slackhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.slack,
	}, app)
	calendarhttp.Routes(calendarhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.calendar,
	}, app)
	mayahttp.Routes(mayahttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.maya,
		Workspaces:        svcs.workspaces,
		Stories:           svcs.stories,
		States:            svcs.states,
		Teams:             svcs.teams,
		Users:             svcs.users,
		Objectives:        svcs.objectives,
		KeyResults:        svcs.keyResults,
		Search:            svcs.search,
		Activities:        svcs.activities,
		Feedback:          svcs.feedback,
		Notifications:     svcs.notifications,
		Reports:           svcs.reports,
		Sprints:           svcs.sprints,
		AIAPIKey:          cfg.AIAPIKey,
	}, app)

	integrationrequestshttp.Routes(integrationrequestshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.integrationRequests,
	}, app)

	feedbackhttp.Routes(feedbackhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		IngressSecret:     cfg.FeedbackIngressSecret,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.feedback,
		Teams:             svcs.teams,
		Attachments:       svcs.attachments,
	}, app)

	storieshttp.Routes(storieshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Publisher:         cfg.Publisher,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Stories:           svcs.stories,
		Users:             svcs.users,
		Links:             svcs.links,
		Attachments:       svcs.attachments,
	}, app)

	objectiveshttp.Routes(objectiveshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Objectives:        svcs.objectives,
		KeyResults:        svcs.keyResults,
		OKRActivities:     svcs.okrActivities,
		Attachments:       svcs.attachments,
	}, app)

	objectivestatushttp.Routes(objectivestatushttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.objectiveStats,
	}, app)

	labelshttp.Routes(labelshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.labels,
	}, app)

	linkshttp.Routes(linkshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.links,
	}, app)

	sprintshttp.Routes(sprintshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Sprints:           svcs.sprints,
		Attachments:       svcs.attachments,
	}, app)

	epicshttp.Routes(epicshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.epics,
	}, app)

	documentshttp.Routes(documentshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.documents,
		Attachments:       svcs.attachments,
	}, app)

	stateshttp.Routes(stateshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.states,
	}, app)

	teamshttp.Routes(teamshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.teams,
	}, app)

	usershttp.Routes(usershttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		DeploymentMode:    cfg.DeploymentMode,
		SecretKey:         cfg.SecretKey,
		CookieDomain:      cfg.CookieDomain,
		WebsiteURL:        cfg.WebsiteURL,
		GoogleService:     cfg.GoogleService,
		MicrosoftService:  cfg.MicrosoftService,
		Publisher:         cfg.Publisher,
		TasksService:      cfg.TasksService,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Users:             svcs.users,
		Attachments:       svcs.attachments,
	}, app)

	workspaceshttp.Routes(workspaceshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Workspaces:        svcs.workspaces,
		Attachments:       svcs.attachments,
	}, app)

	commentshttp.Routes(commentshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.comments,
	}, app)

	activitieshttp.Routes(activitieshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Activities:        svcs.activities,
		Attachments:       svcs.attachments,
	}, app)

	reportshttp.Routes(reportshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		Reports:           svcs.reports,
		Attachments:       svcs.attachments,
	}, app)

	keyresultshttp.Routes(keyresultshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		StorageConfig:     cfg.StorageConfig,
		StorageService:    cfg.StorageService,
		KeyResults:        svcs.keyResults,
		OKRActivities:     svcs.okrActivities,
		Attachments:       svcs.attachments,
	}, app)

	notificationshttp.Routes(notificationshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Redis:             cfg.Redis,
		TasksService:      cfg.TasksService,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.notifications,
		Users:             svcs.users,
		Attachments:       svcs.attachments,
	}, app)

	invitationshttp.Routes(invitationshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Publisher:         cfg.Publisher,
		StripeClient:      cfg.StripeClient,
		StripeSecret:      cfg.WebhookSecret,
		TasksService:      cfg.TasksService,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Invitations:       svcs.invitations,
		UsersService:      svcs.users,
	}, app)

	searchhttp.Routes(searchhttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.search,
	}, app)

	subscriptionshttp.Routes(subscriptionshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Subscriptions:     svcs.subscriptions,
		Users:             svcs.users,
		Workspaces:        svcs.workspaces,
	}, app)

	ssehttp.Routes(ssehttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		SSEHub:            cfg.SSEHub,
		Origins:           cfg.AllowedOrigins,
		BrowserSessions:   browserSessions,
	}, app)

	teamsettingshttp.Routes(teamsettingshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.teamSettings,
	}, app)

	chatsessionshttp.Routes(chatsessionshttp.Config{
		WorkspaceResolver: r.workspaceResolver,
		Log:               cfg.Log,
		SecretKey:         cfg.SecretKey,
		Cache:             cfg.Cache,
		BrowserSessions:   browserSessions,
		Service:           svcs.chatSessions,
	}, app)

	app.NotFound(func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return web.RespondError(ctx, w, errors.New("API route not found"), http.StatusNotFound)
	})

}
